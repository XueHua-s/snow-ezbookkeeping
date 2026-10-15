#!/usr/bin/env python3
"""Opt-in native full-database restore and old-binary rollback on synthetic data.
Uses task-owned Docker containers, creates unique databases, never drops existing DBs.
Evidence (including synthetic rows) remains in the printed temporary directory.
"""
import argparse, configparser, hashlib, json, os, pathlib, subprocess, tempfile, time

ROOT = pathlib.Path(__file__).resolve().parents[1]
p = argparse.ArgumentParser()
p.add_argument('--mysql-container', required=True)
p.add_argument('--postgres-container', required=True)
p.add_argument('--mysql-port', required=True)
p.add_argument('--postgres-port', required=True)
p.add_argument('--old-ref', default='1a21bc9bfd2b155a3ed9b8ac57caff7c76bbb293')
a = p.parse_args()
password = os.environ.get('EBK_MIGRATION_TEST_PASSWORD')
if not password:
    raise SystemExit('Set EBK_MIGRATION_TEST_PASSWORD to the task-only test password')
if not all(name.startswith('ebk-experience-verify-') for name in [a.mysql_container, a.postgres_container]):
    raise SystemExit('Refusing containers outside the dedicated verification namespace')
work = pathlib.Path(tempfile.mkdtemp(prefix='ebk_experience_migration_'))
print(work, flush=True)
(work/'storage').mkdir()
env = dict(os.environ, BUILD_PIPELINE='1', EBK_WORK_DIR=str(ROOT))
log = open(work/'commands.log', 'w')

def run(cmd, *, data=None, cwd=ROOT, output=None):
    safe = [v.replace(password, '<test-password>') for v in map(str, cmd)]
    log.write(json.dumps(safe)+'\n'); log.flush()
    r = subprocess.run(list(map(str,cmd)), input=data, stdout=subprocess.PIPE, stderr=subprocess.PIPE, cwd=cwd, env=env)
    log.write('exit='+str(r.returncode)+'\n'+r.stderr.decode(errors='replace').replace(password,'<test-password>')+'\n');log.flush()
    if output: pathlib.Path(output).write_bytes(r.stdout)
    if r.returncode: raise RuntimeError('command failed; see '+str(work/'commands.log')+' '+r.stderr.decode(errors='replace')[-1500:])
    return r.stdout

def config(kind, name):
    path = work/(name+'.ini')
    port = a.mysql_port if kind=='mysql' else a.postgres_port
    user = 'root' if kind=='mysql' else 'postgres'
    template=configparser.ConfigParser(interpolation=None, strict=False)
    template.read(ROOT/'conf/ezbookkeeping.ini')
    for key,value in {'type':kind,'host':'127.0.0.1:'+port,'name':name,'user':user,'passwd':password,'ssl_mode':'disable','db_path':str(work/(name+'.db')),'log_query':'false'}.items(): template.set('database',key,value)
    template.set('security','secret_key','migration-synthetic-signing-secret')
    template.set('log','mode','console'); template.set('log','level','error')
    template.set('storage','local_filesystem_path',str(work/'storage'))
    with path.open('w') as out: template.write(out)
    path.chmod(0o600)
    return path

def mysql(sql):
    return run(['docker','exec','-i','-e','MYSQL_PWD='+password,a.mysql_container,'mysql','-uroot'],data=sql.encode())
def pg(sql):
    return run(['docker','exec','-i',a.postgres_container,'psql','-U','postgres','-v','ON_ERROR_STOP=1'],data=sql.encode())
def manifest(cfg, name, legacy=False):
    raw = run([work/'helper',cfg,'legacy-manifest' if legacy else 'manifest'])
    # Bootstrap logging may precede JSON; the canonical payload starts on its own line.
    start=raw.find(b'{\n')
    value=json.loads(raw[start:]); (work/(name+'.json')).write_text(json.dumps(value,indent=2,sort_keys=True))
    return value

old = work/'old-source'; old.mkdir()
archive = run(['git','archive',a.old_ref])
run(['tar','-xf','-','-C',old],data=archive)
run(['go','build','-o',work/'old-ebk','./ezbookkeeping.go'],cwd=old)
run(['go','build','-o',work/'new-ebk','./ezbookkeeping.go'])
run(['go','build','-o',work/'helper','./scripts/assistant-experience-migration-check'])
results={}
for kind in ['sqlite3','mysql','postgres']:
    suffix=str(time.time_ns())
    src='ebk_experience_migration_'+kind+'_'+suffix
    dst=src+'_restored'
    if kind=='mysql': mysql('CREATE DATABASE `'+src+'`; CREATE DATABASE `'+dst+'`;')
    if kind=='postgres': pg('CREATE DATABASE "'+src+'"; CREATE DATABASE "'+dst+'";')
    c=config(kind,src); restored=config(kind,dst)
    run([work/'old-ebk','--conf-path',c,'database','update'],output=work/(kind+'-old-init.log'))
    run([work/'helper',c,'legacy'])
    legacy=manifest(c,kind+'-legacy-before',True)
    run([work/'new-ebk','--conf-path',c,'database','update'],output=work/(kind+'-upgrade.log'))
    run([work/'new-ebk','--conf-path',c,'database','update'],output=work/(kind+'-upgrade-repeat.log'))
    assert manifest(c,kind+'-legacy-upgraded',True)==legacy, 'legacy rows changed during migration'
    run([work/'helper',c,'seed'])
    before=manifest(c,kind+'-before')
    backup=work/(kind+'.backup')
    if kind=='sqlite3':
        run(['sqlite3',work/(src+'.db'),'.backup '+str(backup)])
        run(['sqlite3',work/(dst+'.db'),'.restore '+str(backup)])
    elif kind=='mysql':
        raw=run(['docker','exec','-e','MYSQL_PWD='+password,a.mysql_container,'mysqldump','-uroot','--single-transaction','--skip-lock-tables','--no-tablespaces',src],output=backup)
        run(['docker','exec','-i','-e','MYSQL_PWD='+password,a.mysql_container,'mysql','-uroot',dst],data=raw)
    else:
        raw=run(['docker','exec',a.postgres_container,'pg_dump','-U','postgres','-Fc',src],output=backup)
        run(['docker','exec','-i',a.postgres_container,'pg_restore','-U','postgres','--exit-on-error','-d',dst],data=raw)
    assert manifest(restored,kind+'-restored')==before, 'restored data differs'
    run([work/'old-ebk','--conf-path',restored,'database','update'],output=work/(kind+'-rollback-old.log'))
    assert manifest(restored,kind+'-after-old-rollback')==before, 'old binary modified new records'
    run([work/'new-ebk','--conf-path',restored,'database','update'],output=work/(kind+'-reupgrade.log'))
    assert manifest(restored,kind+'-reupgraded')==before, 're-upgrade changed data'
    run([work/'helper',restored,'recover'],output=work/(kind+'-recovery.log'))
    after=manifest(restored,kind+'-recovered')
    for key in ['ledger','embedding','memory','receipt']:
        assert before[key]==after[key], 'recovery changed unrelated '+key
    results[kind]={'status':'PASS','backup_sha256':hashlib.sha256(backup.read_bytes()).hexdigest(),'source':src,'restored':dst}
    (work/'results.json').write_text(json.dumps(results,indent=2))
    print(kind+' PASS',flush=True)
print('Evidence retained at '+str(work))
