import { expect, it } from 'vitest';
import { mkdtemp, readFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';
import { spawn } from 'node:child_process';
import { boundedCommand } from './assistant-bounded-command.mjs';

async function waitFor(read) {
    for (let i = 0; i < 100; i++) {
        try { const value = await read(); if (value) return value; } catch { /* Wait for fixture. */ }
        await new Promise(resolve => setTimeout(resolve, 20));
    }
    throw new Error('Fixture did not reach expected state');
}
function gone(pid) {
    try { process.kill(pid, 0); return false; } catch (error) { return error.code === 'ESRCH'; }
}
const descendant = path => `const {spawn}=require('node:child_process'); const fs=require('node:fs');
const child=spawn(process.execPath,['-e', 'process.on("SIGTERM",()=>{}); setInterval(()=>{},1000)'],{stdio:'ignore'});
fs.writeFileSync(${JSON.stringify(path)},String(child.pid)); process.on('SIGTERM',()=>{}); setInterval(()=>{},1000);`;

it('times out and kills a real resistant subprocess and its grandchild', async () => {
    const directory = await mkdtemp(resolve(tmpdir(), 'ebk-bounded-test-'));
    const pidFile = resolve(directory, 'pid');
    const abort = new AbortController();
    const watchdog = setTimeout(() => abort.abort(), 2000);
    let pid;
    try {
        const command = boundedCommand(process.execPath, ['-e', descendant(pidFile)], { timeout: 200, signal: abort.signal });
        const assertion = expect(command).rejects.toThrow('timed out after 200ms');
        pid = Number(await waitFor(() => readFile(pidFile, 'utf8')));
        await assertion;
        await waitFor(() => gone(pid));
    } finally {
        clearTimeout(watchdog); if (pid && !gone(pid)) process.kill(pid, 'SIGKILL');
        await rm(directory, { recursive: true, force: true });
    }
});

it('reports spawn failures and releases process handlers', async () => {
    const exits = process.listenerCount('exit'); const signals = process.listenerCount('SIGTERM');
    await expect(boundedCommand('/no-such-ebk-command', [])).rejects.toThrow('ENOENT');
    expect(process.listenerCount('exit')).toBe(exits); expect(process.listenerCount('SIGTERM')).toBe(signals);
});

it.each(['SIGINT', 'SIGTERM'])('interrupts real command descendants on %s and reaches caller finally', async signal => {
    const directory = await mkdtemp(resolve(tmpdir(), 'ebk-interrupt-test-'));
    const pidFile = resolve(directory, 'pid'); const finished = resolve(directory, 'finished');
    const source = `import {boundedCommand} from ${JSON.stringify(new URL('./assistant-bounded-command.mjs', import.meta.url).href)};
import {writeFileSync} from 'node:fs';
try { await boundedCommand(process.execPath,['-e',${JSON.stringify(descendant(pidFile))}],{timeout:30000}); }
catch {} finally {writeFileSync(${JSON.stringify(finished)},'closed');}`;
    const runner = spawn(process.execPath, ['--input-type=module', '-e', source], { stdio: 'ignore' });
    const ended = new Promise(resolve => runner.once('exit', (code, signal) => resolve({ code, signal })));
    let pid;
    try {
        pid = Number(await waitFor(() => readFile(pidFile, 'utf8')));
        runner.kill(signal);
        expect(await ended).toEqual({ code: signal === 'SIGINT' ? 130 : 143, signal: null });
        expect(await readFile(finished, 'utf8')).toBe('closed');
        await waitFor(() => gone(pid));
    } finally {
        runner.kill('SIGKILL'); if (pid && !gone(pid)) process.kill(pid, 'SIGKILL');
        await rm(directory, { recursive: true, force: true });
    }
});

it('caps captured output even when a resistant child keeps flooding both pipes', async () => {
    const command = boundedCommand(process.execPath, ['-e', `process.on('SIGTERM',()=>{});
const block='x'.repeat(1024*1024); setInterval(()=>{process.stdout.write(block);process.stderr.write(block);},1);`], { timeout: 3000 });
    const error = await command.catch(error => error);
    expect(error.message).toContain('output limit exceeded');
    expect(Buffer.byteLength(error.stdout) + Buffer.byteLength(error.stderr)).toBe(16 * 1024 * 1024);
});
