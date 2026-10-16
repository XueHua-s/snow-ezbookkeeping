import assert from 'node:assert/strict';

/** Production-only: verifies the installed worker, HTTP headers and real IndexedDB, not source strings. */
export async function cacheFlows({ page, context, identity, variant, results }) {
    await page.evaluate(async () => { await navigator.serviceWorker.ready; });
    await page.waitForFunction(() => !!navigator.serviceWorker.controller, { timeout: 30000 });
    const policy = await page.evaluate(async () => {
        const id = new URLSearchParams(location.hash.split('?')[1]).get('conversationId');
        const response = await fetch(`/api/v1/llm/assistant/conversations/get.json?id=${id}`, { headers: { Authorization: `Bearer ${localStorage.getItem('ebk_user_token')}` } });
        const urls = [];
        for (const name of await caches.keys()) for (const request of await (await caches.open(name)).keys()) if (/\/api\/v1\/llm\/assistant(?:\/|$)/.test(new URL(request.url).pathname)) urls.push(request.url);
        return { cacheControl: response.headers.get('cache-control'), pragma: response.headers.get('pragma'), cachedAssistantResponses: urls.length, controlled: !!navigator.serviceWorker.controller };
    });
    assert.match(policy.cacheControl, /no-store/); assert.equal(policy.cachedAssistantResponses, 0); assert.equal(policy.controlled, true);
    const local = await page.evaluate(() => new Promise((resolve, reject) => {
        const request = indexedDB.open('ezbookkeeping-assistant-local-v1', 1);
        request.onerror = () => reject(request.error);
        request.onsuccess = () => { const db = request.result; const read = db.transaction('owners').objectStore('owners').getAll(); read.onsuccess = () => { resolve(read.result); db.close(); }; read.onerror = () => reject(read.error); };
    }));
    assert.ok(local.length > 0); assert.ok(local.every(owner => owner.pending.length === 0));
    assert.ok(!JSON.stringify(local).includes(identity.token)); assert.ok(!JSON.stringify(local).includes('Verified isolated ledger'));
    await context.setOffline(true);
    const networkOnly = await page.evaluate(async () => {
        try { await fetch('/api/v1/llm/assistant/conversations/list.json', { headers: { Authorization: `Bearer ${localStorage.getItem('ebk_user_token')}` } }); return false; } catch { return true; }
    });
    assert.equal(networkOnly, true); await context.setOffline(false);
    results.push({ variant, scenario: 'production service worker controls page; HTTP no-store; no assistant CacheStorage response; IndexedDB stores drafts not answer/token; offline API cannot replay cached history', status: 'PASS', policy });
}
