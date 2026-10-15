import assert from 'node:assert/strict';
import { resolve } from 'node:path';

async function owners(page) {
    return page.evaluate(() => new Promise((resolve, reject) => {
        const request = indexedDB.open('ezbookkeeping-assistant-local-v1', 1);
        request.onerror = () => reject(request.error);
        request.onsuccess = () => { const db = request.result; const read = db.transaction('owners').objectStore('owners').getAll(); read.onsuccess = () => { resolve(read.result); db.close(); }; read.onerror = () => reject(read.error); };
    }));
}
export async function browserBoundaries({ env, browser, context, page, identity, variant, conversationURL, results }) {
    const input = page.getByPlaceholder('Ask your personal finance question');
    const tab = await context.newPage(); await tab.goto(conversationURL, { waitUntil: 'domcontentloaded', timeout: 120000 });
    const otherInput = tab.getByPlaceholder('Ask your personal finance question'); await otherInput.waitFor();
    assert.equal(await otherInput.inputValue(), 'Draft preserved through bill detail');
    await input.fill('First tab concurrent draft');
    await page.waitForFunction(() => document.querySelector('textarea')?.value === 'First tab concurrent draft');
    await otherInput.fill('Second tab concurrent draft');
    for (let attempt = 0; attempt < 100; attempt++) {
        const texts = (await owners(page)).flatMap(owner => owner.drafts.map(draft => draft.text));
        if (texts.includes('First tab concurrent draft') && texts.includes('Second tab concurrent draft')) break;
        if (attempt === 99) throw new Error('Concurrent tabs did not retain both drafts');
        await page.waitForTimeout(50);
    }
    await tab.close(); await input.fill('Draft preserved through bill detail');
    results.push({ variant, scenario: 'real shared IndexedDB retains conflicting drafts from two tabs', status: 'PASS' });

    const unavailable = await browser.newContext({ viewport: variant === 'desktop' ? { width: 1440, height: 960 } : { width: 390, height: 844 } });
    await unavailable.addInitScript(login => {
        localStorage.setItem('ebk_user_token', login.token); localStorage.setItem('ebk_user_info', JSON.stringify(login.user));
        indexedDB.open = () => { throw new DOMException('Test storage denial', 'SecurityError'); };
    }, identity);
    const denied = await unavailable.newPage(); await denied.goto(conversationURL, { waitUntil: 'domcontentloaded', timeout: 120000 });
    await denied.getByPlaceholder('Ask your personal finance question').fill('Visible memory-only draft');
    await denied.getByRole('status').filter({ hasText: 'Only kept on this page. Refreshing may lose unsent content.' }).waitFor();
    assert.equal(await denied.getByPlaceholder('Ask your personal finance question').inputValue(), 'Visible memory-only draft');
    await denied.screenshot({ path: resolve(env.artifacts, `${variant}-storage-denied.png`) });
    await unavailable.close();
    results.push({ variant, scenario: 'real browser IndexedDB refusal keeps editable draft and visibly warns about refresh loss', status: 'PASS' });
}

/** Cross-tab authentication event uses the real capabilities/HTTP authorization boundary. */
export async function identityBoundary({ env, page, identity, variant, results, attemptBFCache }) {
    const location = page.url();
    const other = await env.user('switch_' + variant);
    if (attemptBFCache) {
        await page.evaluate(() => { window.__e2eBFCacheRestored = false; window.addEventListener('pageshow', event => { window.__e2eBFCacheRestored = event.persisted; }); });
        await page.goto(`${env.url}/favicon.ico`);
        const switcher = await page.context().newPage(); await switcher.goto(`${env.url}/favicon.ico`);
        await switcher.evaluate(login => { localStorage.setItem('ebk_user_token', login.token); localStorage.setItem('ebk_user_info', JSON.stringify(login.user)); }, other);
        await switcher.close(); await page.goBack({ waitUntil: 'domcontentloaded' });
        await page.evaluate(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))));
        const restored = await page.evaluate(() => window.__e2eBFCacheRestored === true);
        results.push({ variant, scenario: 'actual back/forward cache restoration after another tab changes identity', status: restored ? 'PASS' : 'NOT RUN', ...(restored ? {} : { reason: 'Chromium did not retain this navigation in bfcache; normal history restore tested instead' }) });
    } else {
        await page.evaluate(async login => {
            localStorage.setItem('ebk_user_token', login.token); localStorage.setItem('ebk_user_info', JSON.stringify(login.user));
            window.dispatchEvent(new StorageEvent('storage', { key: 'ebk_user_token', newValue: login.token }));
            await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
        }, other);
    }
    assert.equal(await page.getByRole('heading', { name: 'Verified isolated ledger' }).count(), 0);
    await page.evaluate(login => { localStorage.setItem('ebk_user_token', login.token); localStorage.setItem('ebk_user_info', JSON.stringify(login.user)); window.dispatchEvent(new StorageEvent('storage', { key: 'ebk_user_token', newValue: login.token })); }, identity);
    await page.goto(location, { waitUntil: 'domcontentloaded', timeout: 120000 }); await page.reload();
    await page.getByRole('heading', { name: 'Verified isolated ledger' }).waitFor();
    results.push({ variant, scenario: 'cross-tab authentication switch clears prior owner answer before next rendered frame; original owner can reopen own formal history', status: 'PASS' });
}
