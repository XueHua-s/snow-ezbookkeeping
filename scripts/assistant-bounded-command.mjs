import { spawn } from 'node:child_process';

const activeGroups = new Set();
function killGroup(pid, signal) {
    try { process.kill(-pid, signal); } catch (error) { if (error.code !== 'ESRCH') throw error; }
}

// Task exit recovery calls this before deleting daemon-side containers.
export function stopActiveCommands() {
    for (const pid of activeGroups) killGroup(pid, 'SIGKILL');
}

// Commands own a POSIX process group: npm/CLI grandchildren cannot outlive the
// deadline or keep captured pipes open after their direct parent exits.
export function boundedCommand(command, args, { timeout = 30000, capture = true, env = process.env, signal, stdio } = {}) {
    return new Promise((resolve, reject) => {
        if (signal?.aborted) { reject(new Error(`${command}: aborted`)); return; }
        const child = spawn(command, args, { env, detached: true, stdio: stdio ?? (capture ? ['ignore', 'pipe', 'pipe'] : 'inherit') });
        let stdout = ''; let stderr = ''; let capturedBytes = 0; let failure; let deadline; let escalation; let settled = false;
        if (child.pid) activeGroups.add(child.pid);
        const exitRecovery = () => { if (child.pid) killGroup(child.pid, 'SIGKILL'); };
        const interrupt = name => {
            process.exitCode = name === 'SIGINT' ? 130 : 143;
            terminate(new Error(`${command}: interrupted by ${name}`));
        };
        const onInt = () => interrupt('SIGINT');
        const onTerm = () => interrupt('SIGTERM');
        const onAbort = () => terminate(new Error(`${command}: aborted`));
        function finish(code) {
            if (settled) return;
            settled = true;
            clearTimeout(deadline); clearTimeout(escalation);
            process.off('exit', exitRecovery); process.off('SIGINT', onInt); process.off('SIGTERM', onTerm);
            signal?.removeEventListener('abort', onAbort);
            try { exitRecovery(); } catch (error) { failure ??= error; }
            activeGroups.delete(child.pid);
            if (failure || code !== 0) {
                const error = failure ?? new Error(`${command} exited ${code}`);
                Object.assign(error, { stdout, stderr, exitCode: code }); reject(error);
            } else resolve({ stdout, stderr, exitCode: code });
        }
        function terminate(error) {
            if (failure || settled) return;
            failure = error;
            try { if (child.pid) killGroup(child.pid, 'SIGTERM'); } catch (killError) { failure = new AggregateError([error, killError]); }
            escalation = setTimeout(() => {
                try { exitRecovery(); } catch (killError) { failure = new AggregateError([failure, killError]); }
                // Do not await a broken child-process close notification forever.
                finish(null);
            }, 500);
        }
        process.once('exit', exitRecovery); process.on('SIGINT', onInt); process.on('SIGTERM', onTerm);
        signal?.addEventListener('abort', onAbort, { once: true });
        child.once('error', error => { failure = error; finish(null); });
        child.once('close', finish);
        for (const [stream, append] of [[child.stdout, chunk => { stdout += chunk; }], [child.stderr, chunk => { stderr += chunk; }]]) {
            stream?.on('data', chunk => {
                if (failure || settled) return;
                const remaining = 16 * 1024 * 1024 - capturedBytes;
                const accepted = chunk.subarray(0, remaining);
                capturedBytes += accepted.length; append(accepted);
                if (chunk.length > remaining) terminate(new Error(`${command}: output limit exceeded`));
            });
        }
        deadline = setTimeout(() => terminate(new Error(`${command}: timed out after ${timeout}ms`)), timeout);
    });
}
