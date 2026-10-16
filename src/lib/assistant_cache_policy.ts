/** Matches API paths under a deployment base path, never URL query text. */
export function isAssistantApiUrl(url: URL): boolean {
    return /\/api\/v1\/llm\/assistant(?:\/|$)/.test(url.pathname);
}

/** Remove responses an earlier worker's configurable map rule may have cached. */
export async function removeAssistantCachedResponses(storage: CacheStorage): Promise<void> {
    for (const name of await storage.keys()) {
        if (!name.startsWith('ezbookkeeping-')) {
            continue;
        }
        const cache = await storage.open(name);
        for (const request of await cache.keys()) {
            if (isAssistantApiUrl(new URL(request.url))) {
                await cache.delete(request);
            }
        }
    }
}
