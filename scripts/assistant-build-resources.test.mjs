import { expect, it } from 'vitest';
import { mkdtemp, mkdir, writeFile, readFile, access, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';
import { spawn } from 'node:child_process';
import { boundedCommand } from './assistant-bounded-command.mjs';
import { buildResources } from './assistant-build-resources.mjs';

async function fixture() {
    const root = await mkdtemp(resolve(tmpdir(), 'ebk-build-cleanup-'));
    const directory = resolve(root, 'source'); await mkdir(directory);
    const docker = resolve(root, 'docker');
    await writeFile(docker, `#!${process.execPath}
const fs=require('node:fs'); const path=require('node:path'); const root=${JSON.stringify(root)};
const args=process.argv.slice(2), action=args[0], name=args.at(-1);
fs.appendFileSync(path.join(root,'calls'),JSON.stringify(args)+'\\n');
if(action==='rm') {
 if(fs.existsSync(path.join(root,'unavailable'))) {console.error('Cannot connect to Docker daemon');process.exit(1);}
 const container=path.join(root,name);
 if(!fs.existsSync(container)){console.error('Error: No such container: '+name);process.exit(1);}
 fs.unlinkSync(container);process.exit(0);
}
fs.writeFileSync(path.join(root,name),'created');
if(action==='fail'){console.error('CLI lost connection after create');process.exit(7);}
const child=require('node:child_process').spawn(process.execPath,['-e','process.on("SIGTERM",()=>{});setInterval(()=>{},1000)'],{stdio:'ignore'});
fs.writeFileSync(path.join(root,'descendant'),String(child.pid));
process.on('SIGTERM',()=>{});setInterval(()=>{},1000);
`, { mode: 0o700 });
    return { root, directory, docker };
}
async function missing(path) { try { await access(path); return false; } catch (error) { return error.code === 'ENOENT'; } }
async function eventually(read) {
    for (let i = 0; i < 100; i++) { try { if (await read()) return; } catch {} await new Promise(resolve => setTimeout(resolve, 20)); }
    throw new Error('Fixture deadline');
}

it('cleans the exact attempted container after a real CLI failure, including already absent names', async () => {
    const f = await fixture(); const resources = buildResources(f.directory, f.docker);
    try {
        const name = resources.container('builder');
        await expect(boundedCommand(f.docker, ['fail', name])).rejects.toThrow('exited 7');
        resources.container('probe'); // An attempted create that never reached the daemon.
        await resources.close();
        expect(await missing(resolve(f.root, name))).toBe(true);
        expect(await missing(f.directory)).toBe(true);
        const calls = (await readFile(resolve(f.root, 'calls'), 'utf8')).trim().split('\n').map(JSON.parse);
        expect(calls.filter(args => args[0] === 'rm')).toHaveLength(2);
        expect(calls.filter(args => args[0] === 'rm').every(args => /^ebk-assistant-(builder|probe)-[a-f0-9-]{36}$/.test(args[2]))).toBe(true);
    } finally { await resources.close(); await rm(f.root, { recursive: true, force: true }); }
});

it('reports unknown daemon cleanup errors, still removes source, and keeps recovery registered', async () => {
    const f = await fixture(); const before = process.listenerCount('exit');
    const resources = buildResources(f.directory, f.docker);
    try {
        const name = resources.container('builder');
        await expect(boundedCommand(f.docker, ['fail', name])).rejects.toThrow('exited 7');
        await writeFile(resolve(f.root, 'unavailable'), 'yes');
        await expect(resources.close()).rejects.toThrow('Build resource cleanup failed');
        expect(await missing(f.directory)).toBe(true);
        expect(await missing(resolve(f.root, name))).toBe(false);
        expect(process.listenerCount('exit')).toBe(before + 1);
        await rm(resolve(f.root, 'unavailable')); await resources.close();
        expect(process.listenerCount('exit')).toBe(before);
    } finally { await rm(resolve(f.root, 'unavailable'), { force: true }); await resources.close(); await rm(f.root, { recursive: true, force: true }); }
});

it.each(['SIGINT', 'SIGTERM'])('exit recovery on %s kills descendants and removes uncertain daemon resources', async signal => {
    const f = await fixture();
    const source = `import {buildResources} from ${JSON.stringify(new URL('./assistant-build-resources.mjs', import.meta.url).href)};
import {boundedCommand} from ${JSON.stringify(new URL('./assistant-bounded-command.mjs', import.meta.url).href)};
const resources=buildResources(${JSON.stringify(f.directory)},${JSON.stringify(f.docker)});
const name=resources.container('builder');
await boundedCommand(${JSON.stringify(f.docker)},['hang',name],{timeout:30000});`;
    const runner = spawn(process.execPath, ['--input-type=module', '-e', source], { stdio: 'ignore' });
    const ended = new Promise(resolve => runner.once('exit', (code, signal) => resolve({ code, signal })));
    let pid;
    try {
        await eventually(async () => { pid = Number(await readFile(resolve(f.root, 'descendant'), 'utf8')); return pid; });
        runner.kill(signal);
        expect(await ended).toEqual({ code: signal === 'SIGINT' ? 130 : 143, signal: null });
        expect(await missing(f.directory)).toBe(true);
        const calls = (await readFile(resolve(f.root, 'calls'), 'utf8')).trim().split('\n').map(JSON.parse);
        expect(calls[1]).toEqual(['rm', '-f', calls[0][1]]);
        expect(await missing(resolve(f.root, calls[0][1]))).toBe(true);
        await eventually(() => { try { process.kill(pid, 0); return false; } catch (error) { return error.code === 'ESRCH'; } });
    } finally {
        runner.kill('SIGKILL'); if (pid) { try { process.kill(pid, 'SIGKILL'); } catch {} }
        await rm(f.root, { recursive: true, force: true });
    }
});

it('makes exit-fallback daemon failures visible while still removing temporary source', async () => {
    const f = await fixture();
    await writeFile(resolve(f.root, 'unavailable'), 'yes');
    const source = `import {buildResources} from ${JSON.stringify(new URL('./assistant-build-resources.mjs', import.meta.url).href)};
import {writeFileSync} from 'node:fs';
const resources=buildResources(${JSON.stringify(f.directory)},${JSON.stringify(f.docker)});
const name=resources.container('probe');writeFileSync(${JSON.stringify(f.root)}+'/'+name,'created');process.exit(2);`;
    try {
        const error = await boundedCommand(process.execPath, ['--input-type=module', '-e', source]).catch(error => error);
        expect(error.exitCode).toBe(2); expect(error.stderr).toContain('Build cleanup failed for task containers');
        expect(error.stderr).toContain('Cannot connect to Docker daemon');
        expect(await missing(f.directory)).toBe(true);
    } finally { await rm(f.root, { recursive: true, force: true }); }
});
