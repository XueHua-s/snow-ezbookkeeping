import { execFile, execFileSync } from 'node:child_process';
import { promisify } from 'node:util';
import { randomUUID } from 'node:crypto';
import { writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';

const execute = promisify(execFile);
export async function dockerCommand(args) {
    return (await execute('docker', args, { timeout: 30000, maxBuffer: 4 * 1024 * 1024 })).stdout.trim();
}

// Only a named, task-owned container is mutated. No host directories are mounted.
export function dockerRuntime({ image, port, artifacts }) {
    const name = `ebk-assistant-e2e-${randomUUID()}`;
    let created = false;
    let attempted = false;
    const exitCleanup = () => {
        if (attempted && !closed) {
            try { execFileSync('docker', ['rm', '-f', name], { timeout: 10000, stdio: 'ignore' }); } catch { /* Normal close reports failures. */ }
        }
    };
    let closed = false;
    const events = [];
    async function record(action) {
        events.push({ action, at: new Date().toISOString() });
        await writeFile(resolve(artifacts, 'docker-lifecycle.json'), JSON.stringify({ name, events }, null, 2));
    }
    return {
        async start(configFile) {
            if (!created) {
                attempted = true;
                process.once('exit', exitCleanup);
                await dockerCommand(['create', '--pull', 'never', '--name', name, '--label', 'ebk.validation=assistant-experience',
                    '--user', '1000:1000', '-p', `127.0.0.1:${port}:8080`,
                    '-e', 'EBK_CONF_PATH=/ezbookkeeping/test.ini', image]);
                created = true;
                const inspected = JSON.parse(await dockerCommand(['inspect', name]))[0];
                await writeFile(resolve(artifacts, 'docker-runtime.json'), JSON.stringify({
                    name, imageId: inspected.Image, user: inspected.Config.User,
                    entrypoint: inspected.Config.Entrypoint, mounts: inspected.Mounts,
                    ports: inspected.HostConfig.PortBindings,
                    provider: 'Host loopback fixture via Colima host.lima.internal; no provider port published; default Docker bridge has outbound access'
                }, null, 2));
            }
            await dockerCommand(['cp', configFile, `${name}:/ezbookkeeping/test.ini`]);
            // The caller makes this synthetic config readable before cp; directory remains private.
            await dockerCommand(['start', name]);
            await record('start');
        },
        async stop() {
            if (created && !closed) { await dockerCommand(['stop', '-t', '3', name]); await record('stop'); }
        },
        async running() {
            return await dockerCommand(['inspect', '--format', '{{.State.Running}}', name]) === 'true';
        },
        async identity() { return await dockerCommand(['exec', name, 'id']); },
        async close() {
            if (!attempted || closed) return;
            if (!created) {
                try { await dockerCommand(['inspect', name]); created = true; } catch (error) {
                    if (!new RegExp(`No such (object|container): ${name}`).test(error.stderr ?? '')) throw error;
                    closed = true; process.off('exit', exitCleanup); return;
                }
            }
            const errors = [];
            try {
                const logs = await execute('docker', ['logs', name], { timeout: 30000, maxBuffer: 16 * 1024 * 1024 });
                await writeFile(resolve(artifacts, 'application.log'), logs.stdout + logs.stderr);
            } catch (error) { errors.push(error); }
            try { await dockerCommand(['rm', '-f', name]); closed = true; process.off('exit', exitCleanup); await record('remove'); } catch (error) { errors.push(error); }
            if (errors.length) throw new AggregateError(errors, 'Docker cleanup failed');
        }
    };
}
