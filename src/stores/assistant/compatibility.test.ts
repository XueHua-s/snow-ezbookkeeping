import { describe, expect, it } from 'vitest';
import { capability, deferred, mocks, response, store } from './testFixture.ts';

const missingCapabilities = { response: { status: 404 } };

describe('assistant capability compatibility', () => {
    it('shows the upgrade requirement when foreground synchronization supersedes initialization', async () => {
        const initial = deferred<never>();
        mocks.api['capabilities']!.mockReturnValueOnce(initial.promise).mockRejectedValueOnce(missingCapabilities);
        const initialize = store.initialize();
        await store.synchronize();
        initial.reject(missingCapabilities);
        await initialize;
        expect(store.legacyBackend).toBe(true);
        expect(store.authenticated).toBe(false);
        expect(mocks.api['send']).not.toHaveBeenCalled();
    });

    it('clears the upgrade requirement when the backend gains capabilities', async () => {
        mocks.api['capabilities']!.mockRejectedValueOnce(missingCapabilities);
        await store.synchronize();
        expect(store.legacyBackend).toBe(true);
        await store.synchronize();
        expect(store.legacyBackend).toBe(false);
        expect(store.authenticated).toBe(true);
    });

    it('does not let a superseded capability 404 overwrite a supported backend', async () => {
        const initial = deferred<never>();
        mocks.api['capabilities']!.mockReturnValueOnce(initial.promise).mockResolvedValueOnce(response(capability()));
        const initialize = store.initialize();
        await store.synchronize();
        initial.reject(missingCapabilities);
        await initialize;
        expect(store.legacyBackend).toBe(false);
        expect(store.authenticated).toBe(true);
    });

    it('does not mistake a history endpoint 404 for missing capabilities', async () => {
        mocks.api['list']!.mockRejectedValueOnce(missingCapabilities);
        await store.initialize();
        expect(store.legacyBackend).toBe(false);
        expect(store.authenticated).toBe(true);
        expect(store.error).toEqual(missingCapabilities);
    });
});
