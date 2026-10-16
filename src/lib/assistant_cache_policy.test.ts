import { describe, expect, it, vi } from 'vitest';
import { isAssistantApiUrl, removeAssistantCachedResponses } from './assistant_cache_policy.ts';

describe('assistant offline cache boundary', () => {
    it('recognizes all assistant reads beneath deployment prefixes without matching query text', () => {
        for (const path of ['/api/v1/llm/assistant/runs/stream.json', '/books/api/v1/llm/assistant/conversations/export.json']) {
            expect(isAssistantApiUrl(new URL(path, 'https://ledger.test'))).toBe(true);
        }
        for (const path of ['/desktop/?next=/api/v1/llm/assistant/', '/api/v1/llm/assistant-other/file', '/img/desktop/logo.png']) {
            expect(isAssistantApiUrl(new URL(path, 'https://ledger.test'))).toBe(false);
        }
    });

    it('removes existing private responses while preserving map tiles, assets and other applications', async () => {
        const privateRequest = new Request('https://ledger.test/base/api/v1/llm/assistant/conversations/get.json?id=1');
        const tile = new Request('https://tiles.test/1/2/3.png');
        const remove = vi.fn().mockResolvedValue(true);
        const open = vi.fn().mockResolvedValue({ keys: async () => [privateRequest, tile], delete: remove });
        const storage = { keys: async () => ['ezbookkeeping-map-cache', 'another-app'], open } as unknown as CacheStorage;
        await removeAssistantCachedResponses(storage);
        expect(open).toHaveBeenCalledExactlyOnceWith('ezbookkeeping-map-cache');
        expect(remove).toHaveBeenCalledExactlyOnceWith(privateRequest);
    });
});
