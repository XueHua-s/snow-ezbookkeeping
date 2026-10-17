import assert from 'node:assert/strict';
import { resolve } from 'node:path';
import { nativeIdentitySwitch } from './assistant-experience-e2e-identity.mjs';

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

/** Native cross-tab authentication and separately labelled bfcache probe. */
export async function identityBoundary({ env, page, identity, variant, results, attemptBFCache }) {
    const location = page.url();
    const other = await env.user('switch_' + variant);
    if (attemptBFCache) {
        await page.evaluate(() => {
            window.__e2eBFCacheRestored = false;
            window.__e2eBFCacheFrame = undefined;
            window.addEventListener('pageshow', event => {
                window.__e2eBFCacheRestored = event.persisted;
                if (event.persisted) requestAnimationFrame(() => {
                    window.__e2eBFCacheFrame = { text: document.body.innerText, draft: document.querySelector('textarea')?.value ?? '' };
                });
            });
        });
        await page.goto(`${env.url}/favicon.ico`);
        const switcher = await page.context().newPage(); await switcher.goto(`${env.url}/favicon.ico`);
        await switcher.evaluate(login => { localStorage.setItem('ebk_user_token', login.token); localStorage.setItem('ebk_user_info', JSON.stringify(login.user)); }, other);
        await switcher.close(); await page.goBack({ waitUntil: 'domcontentloaded' });
        const restored = await page.evaluate(() => window.__e2eBFCacheRestored === true);
        let restoreEvidence;
        if (restored) {
            await page.waitForFunction(() => window.__e2eBFCacheFrame !== undefined);
            const frame = await page.evaluate(() => window.__e2eBFCacheFrame);
            assert.ok(!frame.text.includes('Verified isolated ledger'), 'Old owner answer in first restored frame');
            assert.ok(!frame.text.includes('Show the isolated expense'), 'Old owner title in first restored frame');
            assert.ok(!frame.draft.includes('Draft preserved through bill detail'), 'Old owner draft in first restored frame');
            restoreEvidence = { firstReceivingFrameHasOldAnswer: false, firstReceivingFrameHasOldDraft: false };
        } else {
            restoreEvidence = await page.evaluate(() => {
                const entry = performance.getEntriesByType('navigation')[0];
                return { notRestoredReasons: entry?.notRestoredReasons?.toJSON?.() ?? null, navigationType: entry?.type };
            });
        }
        results.push({ variant, scenario: 'actual back/forward cache restoration after another tab changes identity', status: restored ? 'PASS' : 'NOT RUN', evidence: restoreEvidence, ...(restored ? {} : { reason: 'Chromium did not retain this navigation in bfcache; normal history restore tested instead' }) });
    } else {
        const switcher = await page.context().newPage();
        await switcher.goto(location, { waitUntil: 'domcontentloaded' });
        await switcher.getByPlaceholder('Ask your personal finance question').waitFor();
        const frame = await nativeIdentitySwitch(page, switcher, other);
        assert.ok(!frame.text.includes('Verified isolated ledger'));
        await switcher.close();
    }
    assert.equal(await page.getByRole('heading', { name: 'Verified isolated ledger' }).count(), 0);
    const restorer = await page.context().newPage();
    await restorer.goto(`${env.url}/favicon.ico`);
    await nativeIdentitySwitch(page, restorer, identity);
    await restorer.close();
    await page.goto(location, { waitUntil: 'domcontentloaded', timeout: 120000 }); await page.reload();
    await page.getByRole('heading', { name: 'Verified isolated ledger' }).last().waitFor();
    results.push({ variant, scenario: attemptBFCache ? 'history navigation after native identity switch hides prior owner; original history remains readable' : 'native storage event first receiving RAF hides prior answer; original owner reopens formal history', status: 'PASS' });
}
