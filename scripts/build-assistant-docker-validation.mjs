import { cp, mkdir, mkdtemp, readFile, readdir, writeFile } from 'node:fs/promises';
import { homedir } from 'node:os';
import { resolve, relative } from 'node:path';
import { createHash } from 'node:crypto';
import { boundedCommand } from './assistant-bounded-command.mjs';
import { buildResources } from './assistant-build-resources.mjs';
const capture = boundedCommand;

const artifacts = resolve(process.env.EBK_E2E_ARTIFACTS || 'docs/ai-assistant-experience-evidence/audit-docker');
const image = process.env.EBK_E2E_DOCKER_IMAGE || 'ebk-assistant-validation:20261016';
await mkdir(artifacts, { recursive: true });
await mkdir(resolve(homedir(), 'Library/Caches'), { recursive: true });
const directory = await mkdtemp(resolve(homedir(), 'Library/Caches/ebk-assistant-build-'));
const commands = [];
const resources = buildResources(directory);
let primaryError;
async function run(command, args, timeout = 30000) {
    const startedAt = new Date().toISOString();
    const result = await boundedCommand(command, args, { timeout, capture: false });
    commands.push({ command, args, startedAt, exit: result.exitCode });
}
async function digest(paths, root = process.cwd()) {
    const files = [];
    async function collect(path) {
        const entries = await readdir(resolve(root, path), { withFileTypes: true });
        for (const entry of entries) {
            const next = `${path}/${entry.name}`;
            if (entry.isDirectory()) await collect(next); else if (entry.isFile()) files.push(next);
        }
    }
    for (const path of paths) {
        if (path.includes('.')) files.push(path); else await collect(path);
    }
    const hash = createHash('sha256');
    for (const path of files.sort()) { hash.update(path); hash.update('\0'); hash.update(await readFile(resolve(root, path))); }
    return { sha256: hash.digest('hex'), files: files.length };
}
const source = ['go.mod', 'go.sum', 'ezbookkeeping.go', 'pkg', 'cmd', 'templates', 'docker/docker-entrypoint.sh', 'conf/ezbookkeeping.ini'];
try {
    await run('npm', ['run', 'build'], 600000);
    for (const path of [...source, 'LICENSE', 'dist']) {
        await mkdir(resolve(directory, relative('.', path), '..'), { recursive: true });
        await cp(resolve(path), resolve(directory, path), { recursive: true });
    }
    const hashes = { source: await digest(source), frontend: await digest(['dist']) };
    if (JSON.stringify(hashes) !== JSON.stringify({ source: await digest(source, directory), frontend: await digest(['dist'], directory) })) throw new Error('Copied sources differ');
    await run('docker', ['run', '--rm', '--name', resources.container('builder'),
        '-e', 'BUILD_PIPELINE=1', '-e', 'GOMODCACHE=/gomod',
        '-v', `${homedir()}/go/pkg/mod:/gomod:ro`, '-v', `${directory}:/work`, '-w', '/work',
        'golang:1.27.1-alpine3.24', 'sh', '-c',
        'sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories && apk add --no-cache gcc musl-dev git && go version && CGO_ENABLED=1 go build -o /work/ezbookkeeping ./ezbookkeeping.go'], 600000);
    await writeFile(resolve(directory, 'Dockerfile.validation'), `FROM golang:1.27.1-alpine3.24
RUN sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories && apk add --no-cache ca-certificates tzdata && addgroup -S -g 1000 ezbookkeeping && adduser -S -G ezbookkeeping -u 1000 ezbookkeeping
WORKDIR /ezbookkeeping
COPY --chown=1000:1000 ezbookkeeping ./ezbookkeeping
COPY --chown=1000:1000 dist ./public
COPY --chown=1000:1000 templates ./templates
COPY --chown=1000:1000 conf ./conf
COPY --chown=1000:1000 LICENSE ./LICENSE
COPY docker/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh && mkdir data storage log && chown 1000:1000 data storage log
USER 1000:1000
EXPOSE 8080
ENTRYPOINT ["/docker-entrypoint.sh"]
`);
    await run('docker', ['build', '-f', resolve(directory, 'Dockerfile.validation'), '-t', image, directory], 600000);
    await run('docker', ['image', 'inspect', '--format', '{{.Id}} {{.Architecture}} {{.Os}}', image]);
    const imageInfo = JSON.parse((await capture('docker', ['image', 'inspect', image])).stdout)[0];
    const baseInfo = JSON.parse((await capture('docker', ['image', 'inspect', 'golang:1.27.1-alpine3.24'])).stdout)[0];
    const binaryBuildInfo = (await capture('go', ['version', '-m', resolve(directory, 'ezbookkeeping')])).stdout;
    const alpineVersion = (await capture('docker', ['run', '--rm', '--name', resources.container('probe'), image, 'cat', '/etc/alpine-release'])).stdout.trim();
    const frontendTools = { node: process.version, npm: (await capture('npm', ['--version'])).stdout.trim() };
    await writeFile(resolve(artifacts, 'build-provenance.json'), JSON.stringify({
        at: new Date().toISOString(), image, imageId: imageInfo.Id, os: imageInfo.Os, architecture: imageInfo.Architecture, frontendTools, hashes, commands, baseImageId: baseInfo.Id, alpineVersion, binaryBuildInfo,
        binarySha256: createHash('sha256').update(await readFile(resolve(directory, 'ezbookkeeping'))).digest('hex'),
        frontendBuild: 'Host npm run build executed by this script; not stock Node 26 Docker frontend stage',
        runtime: 'Cached Go builder base with Alpine 3.24.2 and extra Go toolchain, actual project entrypoint, UID:GID 1000:1000; no synthetic credentials in image; NOT stock runtime image'
    }, null, 2));
} catch (error) {
    primaryError = error; throw error;
} finally {
    try { await resources.close(); } catch (cleanupError) {
        throw primaryError ? new AggregateError([primaryError, cleanupError], 'Build and cleanup failed') : cleanupError;
    }
}
