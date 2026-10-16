import { EventEmitter } from 'node:events';
import { PassThrough } from 'node:stream';
import { mkdtemp, writeFile, readFile, access, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';
import { afterEach, expect, it, vi } from 'vitest';

const state = vi.hoisted(() => ({ servers: [], children: [], directories: [], failure: '', vite: undefined }));
vi.mock('node:fs/promises', async importOriginal => {
    const actual = await importOriginal();
    return { ...actual, mkdtemp: async (...args) => {
        const directory = await actual.mkdtemp(...args);
        if (args[0].includes('ebk-assistant-e2e-')) state.directories.push(directory);
        return directory;
    } };
});
vi.mock('node:http', async importOriginal => {
    const actual = await importOriginal();
    return { ...actual, createServer: (...args) => {
        const server = actual.createServer(...args);
        state.servers.push(server);
        return server;
    } };
});
vi.mock('node:child_process', () => ({ spawn: () => {
    const child = Object.assign(new EventEmitter(), {
        stdout: new PassThrough(), stderr: new PassThrough(), exitCode: null, signalCode: null,
        kill: vi.fn(signal => {
            if (state.failure === 'stubborn' && signal === 'SIGTERM') return;
            child.signalCode = signal;
            queueMicrotask(() => child.emit('close'));
        })
    });
    state.children.push(child);
    if (['spawn', 'build'].includes(state.failure)) queueMicrotask(() => child.emit('error', new Error('synthetic spawn error')));
    if (state.failure === 'ready') queueMicrotask(() => { child.exitCode = 1; child.emit('close'); });
    return child;
} }));
vi.mock('vite', () => ({ createServer: vi.fn(async () => {
    if (state.failure === 'vite-create') throw new Error('synthetic vite create');
    state.vite = { listen: vi.fn(async () => { if (state.failure === 'normal') return; throw new Error('synthetic vite listen'); }), httpServer: { address: () => ({ port: 12345 }) }, close: vi.fn(async () => { if (state.failure === 'cleanup') throw new Error('synthetic close failure'); }) };
    return state.vite;
}) }));
import { environment } from './assistant-experience-e2e-environment.mjs';

afterEach(async () => {
    // Also clean the deliberately failing implementation during the red run.
    for (const server of state.servers.splice(0)) {
        server.closeAllConnections();
        await new Promise(resolve => server.close(resolve));
    }
    for (const child of state.children.splice(0)) child.kill();
    for (const directory of state.directories.splice(0)) {
        try { await expect(access(directory)).rejects.toMatchObject({ code: 'ENOENT' }); }
        finally { await rm(directory, { recursive: true, force: true }); }
    }
    state.vite = undefined;
    vi.useRealTimers();
    vi.unstubAllEnvs();
    vi.unstubAllGlobals();
});

it.each(['spawn', 'build', 'ready', 'vite-create', 'vite-listen'])('cleans partial startup after %s failure', async failure => {
    state.failure = failure;
    vi.stubEnv('EBK_E2E_BINARY', failure === 'build' ? undefined : '/synthetic/fixture-binary');
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: failure !== 'spawn' && failure !== 'ready' })));
    const listeners = process.listenerCount('exit');
    await expect(environment()).rejects.toThrow();
    expect(state.servers.every(server => !server.listening)).toBe(true);
    expect(process.listenerCount('exit')).toBe(listeners);
    if (failure.startsWith('vite')) expect(state.children[0].kill).toHaveBeenCalledWith('SIGTERM');
    if (failure === 'vite-listen') expect(state.vite.close).toHaveBeenCalledOnce();
});

it('bounds readiness requests and escalates an unresponsive task-owned backend', async () => {
    state.failure = 'stubborn';
    vi.useFakeTimers();
    vi.stubEnv('EBK_E2E_BINARY', '/synthetic/fixture-binary');
    vi.stubGlobal('fetch', vi.fn(async (_url, options) => {
        expect(options.signal).toBeInstanceOf(AbortSignal);
        return { ok: true };
    }));
    const failed = expect(environment()).rejects.toThrow('synthetic vite listen');
    await vi.waitFor(() => expect(state.children[0]?.kill).toHaveBeenCalledWith('SIGTERM'));
    await vi.advanceTimersByTimeAsync(3000);
    await failed;
    expect(state.children[0].kill).toHaveBeenCalledWith('SIGKILL');
    expect(state.servers.every(server => !server.listening)).toBe(true);
});

it('reports cleanup rejection while still releasing the other resources', async () => {
    state.failure = 'cleanup';
    vi.stubEnv('EBK_E2E_BINARY', '/synthetic/fixture-binary');
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true })));
    await expect(environment()).rejects.toThrow('Isolated startup and cleanup failed');
    expect(state.children[0].kill).toHaveBeenCalledWith('SIGTERM');
    expect(state.servers.every(server => !server.listening)).toBe(true);
});

it('supplies a deadline that aborts an unresponsive readiness request', async () => {
    state.failure = 'request-timeout';
    vi.useFakeTimers();
    vi.stubEnv('EBK_E2E_BINARY', '/synthetic/fixture-binary');
    const controller = new AbortController();
    const timeout = vi.spyOn(AbortSignal, 'timeout').mockReturnValue(controller.signal);
    const fetcher = vi.fn((_url, { signal }) => new Promise((_resolve, reject) => {
        signal.addEventListener('abort', () => reject(new Error('synthetic deadline')));
    }));
    vi.stubGlobal('fetch', fetcher);
    const failed = expect(environment()).rejects.toThrow('Isolated application exited');
    await vi.waitFor(() => expect(fetcher).toHaveBeenCalledOnce());
    expect(timeout).toHaveBeenCalledWith(1000);
    state.children[0].exitCode = 1;
    controller.abort();
    await vi.advanceTimersByTimeAsync(100);
    await failed;
    expect(state.servers.every(server => !server.listening)).toBe(true);
});

it('normal host close removes its fixture directory but retains an externally selected binary', async () => {
    state.failure = 'normal';
    const external = await mkdtemp(resolve(tmpdir(), 'external-e2e-test-'));
    const binary = resolve(external, 'explicit-binary');
    try {
        await writeFile(binary, 'synthetic executable placeholder');
        vi.stubEnv('EBK_E2E_BINARY', binary);
        vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true })));
        const env = await environment();
        await access(resolve(env.directory, 'test.ini'));
        await env.close();
        await expect(access(env.directory)).rejects.toMatchObject({ code: 'ENOENT' });
        expect(await readFile(binary, 'utf8')).toBe('synthetic executable placeholder');
    } finally { await rm(external, { recursive: true, force: true }); }
});
