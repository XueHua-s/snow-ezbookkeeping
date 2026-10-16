import { afterEach, expect, it, vi } from 'vitest';
import { mkdtemp, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';

const state = vi.hoisted(() => ({ inspectFailure: 'unavailable' }));
vi.mock('node:child_process', () => ({
    execFile: vi.fn((_command, args, _options, callback) => {
        const error = new Error(args[0] === 'create' ? 'create timed out; outcome unknown' : 'inspect failed');
        error.stderr = state.inspectFailure === 'missing' && args[0] === 'inspect'
            ? `Error: No such object: ${args[1]}` : 'Cannot connect to the Docker daemon';
        callback(error);
    }),
    execFileSync: vi.fn()
}));
import { dockerRuntime } from './assistant-experience-docker-runtime.mjs';

let runtime;
let directory;
afterEach(async () => {
    state.inspectFailure = 'missing';
    await runtime?.close();
    if (directory) await rm(directory, { recursive: true, force: true });
});

it('reports uncertain cleanup and retains exit recovery until absence is positively established', async () => {
    directory = await mkdtemp(resolve(tmpdir(), 'ebk-docker-cleanup-test-'));
    const listeners = process.listenerCount('exit');
    runtime = dockerRuntime({ image: 'synthetic', port: 12345, artifacts: directory });
    state.inspectFailure = 'unavailable';
    await expect(runtime.start('/synthetic/test.ini')).rejects.toThrow('outcome unknown');
    await expect(runtime.close()).rejects.toThrow('inspect failed');
    expect(process.listenerCount('exit')).toBe(listeners + 1);
    state.inspectFailure = 'missing';
    await runtime.close();
    expect(process.listenerCount('exit')).toBe(listeners);
    await runtime.close();
    expect(process.listenerCount('exit')).toBe(listeners);
});
