import { execFileSync } from 'node:child_process';
import { rm, rmSync } from 'node:fs';
import { promisify } from 'node:util';
import { randomUUID } from 'node:crypto';
import { boundedCommand, stopActiveCommands } from './assistant-bounded-command.mjs';

const removeDirectory = promisify(rm);
const absent = (error, name) => new RegExp(`No such (object|container): ${name}`).test(String(error.stderr ?? ''));

// Owns only this build's random container names and private source directory.
// Register a name BEFORE docker run: a timed-out CLI can still have created it.
export function buildResources(directory, docker = 'docker') {
    const names = new Set();
    let directoryRemoved = false;
    function detach() {
        process.off('exit', exitRecovery); process.off('SIGINT', onInt); process.off('SIGTERM', onTerm);
    }
    function exitRecovery() {
        const failures = [];
        try { stopActiveCommands(); } catch (error) { failures.push(error); }
        for (const name of names) {
            try { execFileSync(docker, ['rm', '-f', name], { timeout: 10000, killSignal: 'SIGKILL', stdio: ['ignore', 'pipe', 'pipe'] }); names.delete(name); }
            catch (error) { if (absent(error, name)) names.delete(name); else failures.push(error); }
        }
        try { if (!directoryRemoved) { rmSync(directory, { recursive: true, force: true }); directoryRemoved = true; } } catch (error) { failures.push(error); }
        if (failures.length) {
            process.stderr.write(`Build cleanup failed for task containers: ${[...names].join(', ')} (${failures.map(error => error.message).join('; ')})\n`);
            if (!process.exitCode) process.exitCode = 1;
        }
    }
    const onInt = () => process.exit(130);
    const onTerm = () => process.exit(143);
    process.once('exit', exitRecovery); process.on('SIGINT', onInt); process.on('SIGTERM', onTerm);
    return {
        container(role) {
            const name = `ebk-assistant-${role}-${randomUUID()}`;
            names.add(name); return name;
        },
        async close() {
            const failures = [];
            for (const name of names) {
                try { await boundedCommand(docker, ['rm', '-f', name]); names.delete(name); }
                catch (error) { if (absent(error, name)) names.delete(name); else failures.push(error); }
            }
            try { await removeDirectory(directory, { recursive: true, force: true }); directoryRemoved = true; } catch (error) { failures.push(error); }
            if (failures.length) throw new AggregateError(failures, 'Build resource cleanup failed');
            detach();
        }
    };
}
