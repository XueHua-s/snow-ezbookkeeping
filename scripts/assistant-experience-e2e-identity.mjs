import assert from 'node:assert/strict';
import { resolve } from 'node:path';

const prefix = 'v1/llm/assistant/';
const question = page => page.getByPlaceholder('Ask your personal finance question');
const history = page => page.getByRole('navigation', { name: 'Conversation history', exact: true });
async function requestDelete(page) {
    if (!await history(page).isVisible()) await page.getByRole('button', { name: 'Conversation history', exact: true }).click();
    await history(page).getByRole('button', { name: 'Delete', exact: true }).first().click();
    await page.getByRole('dialog', { name: 'Confirm', exact: true }).getByRole('button', { name: 'Confirm', exact: true }).waitFor();
}
async function command(env, login, path, payload = {}) {
    const caps = await env.api(prefix + 'capabilities', undefined, login.token);
    return env.api(prefix + path, { clientRequestId: crypto.randomUUID(), requestIssuedAt: new Date().toISOString(),
        expectedDeletionEpoch: caps.deletionEpoch, ...payload }, login.token);
}

/** Observe the first RAF scheduled by the native receiving-tab storage event, not two later frames. */
export async function nativeIdentitySwitch(page, switcher, login) {
    await page.evaluate(() => {
        window.__identityFrame = undefined;
        const receive = event => {
            if (event.key !== 'ebk_user_token') return;
            window.removeEventListener('storage', receive);
            requestAnimationFrame(() => {
                window.__identityFrame = { trusted: event.isTrusted, text: document.querySelector('.cloud-workspace')?.textContent ?? '',
                    inputs: [...document.querySelectorAll('.cloud-workspace input, .cloud-workspace textarea')].map(input => input.value),
                    dialogs: document.querySelectorAll('.cloud-workspace dialog[open]').length };
            });
        };
        window.addEventListener('storage', receive);
    });
    await switcher.evaluate(login => {
        localStorage.setItem('ebk_user_info', JSON.stringify(login.user));
        localStorage.setItem('ebk_user_token', login.token);
    }, login);
    await page.waitForFunction(() => window.__identityFrame, undefined, { timeout: 20000 });
    const frame = await page.evaluate(() => window.__identityFrame);
    assert.equal(frame.trusted, true, 'must receive a browser-generated storage event');
    return frame;
}

export async function identityFlows({ env, browser, variant, production = false, results }) {
    const first = await env.user(`boundary_a_${variant}`);
    const second = await env.user(`boundary_b_${variant}`);
    const context = await browser.newContext({ viewport: variant === 'desktop' ? { width: 1440, height: 960 } : { width: 390, height: 844 } });
    context.setDefaultTimeout(20000);
    context.setDefaultNavigationTimeout(120000);
    const errors = [];
    context.on('page', page => {
        page.on('pageerror', error => { errors.push(error.stack || error.message); console.error('PAGE_ERROR', page.url(), error.message); });
    });
    await context.addInitScript(login => {
        if (!location.protocol.startsWith('http')) return;
        if (!localStorage.getItem('ebk_user_token')) {
            localStorage.setItem('ebk_user_token', login.token);
            localStorage.setItem('ebk_user_info', JSON.stringify(login.user));
        }
    }, first);
    const root = `${env.url}/${variant}${production ? '' : '.html'}${variant === 'mobile' ? '#!' : '#'}/assistant?conversationId=`;
    const page = await context.newPage();
    const switcher = await context.newPage();
    const note = scenario => { results.push({ variant, scenario, status: 'PASS' }); console.log(`${variant}: ${scenario}`); };
    async function open(login, url = root) {
        if (await page.evaluate(() => localStorage.getItem('ebk_user_token')) !== login.token) {
            await nativeIdentitySwitch(page, switcher, login);
            await question(page).waitFor();
            await page.waitForFunction(() => !new URLSearchParams(location.hash.split('?')[1]).get('conversationId') && !document.querySelector('.cloud-history'));
        }
        const id = new URLSearchParams(url.split('?')[1]).get('conversationId');
        if (id) {
            const detail = await env.api(prefix + `conversations/get?id=${id}`, undefined, login.token);
            if (!await history(page).isVisible()) await page.getByRole('button', { name: 'Conversation history', exact: true }).click();
            await history(page).getByTitle(detail.conversation.title, { exact: true }).click();
            await page.waitForFunction(id => new URLSearchParams(location.hash.split('?')[1]).get('conversationId') === id, id);
        } else {
            await page.locator('.cloud-toolbar').getByRole('button', { name: 'New conversation', exact: true }).click();
        }
        await question(page).waitFor();
        await page.waitForFunction(() => !document.querySelector('.cloud-history'));
    }
    async function seed(login, text) {
        await open(login);
        await question(page).fill(text);
        await page.getByRole('button', { name: 'Send', exact: true }).click();
        await page.getByRole('heading', { name: 'Verified isolated ledger' }).waitFor();
        return { url: page.url(), id: new URLSearchParams(page.url().split('?')[1]).get('conversationId') };
    }
    async function snapshot(login, id) {
        const detail = await env.api(prefix + `conversations/get?id=${id}`, undefined, login.token);
        const memories = await env.api(prefix + 'memories/list', undefined, login.token);
        return { conversation: detail.conversation, messages: detail.messages, memories: memories.items };
    }
    try {
        // Both real application documents start with an empty shared profile and the production initializer.
        await Promise.all([page.goto(root), switcher.goto(root)]);
        await Promise.all([question(page).waitFor(), question(switcher).waitFor()]);
        await Promise.all([question(page).fill('Concurrent production draft A'), question(switcher).fill('Concurrent production draft B')]);
        for (const tab of [page, switcher]) {
            await tab.getByText('Local drafts', { exact: true }).click();
            await tab.getByText('Concurrent production draft A', { exact: true }).waitFor();
            await tab.getByText('Concurrent production draft B', { exact: true }).waitFor();
            assert.ok(await tab.getByText('Conflicting draft — both copies have been kept', { exact: true }).count() >= 2);
        }
        await page.evaluate(() => localStorage.removeItem('assistant-device-id'));
        await page.reload(); await question(page).waitFor();
        await page.getByText('Local drafts', { exact: true }).click();
        await page.getByText('Concurrent production draft A', { exact: true }).waitFor();
        await page.getByText('Concurrent production draft B', { exact: true }).waitFor();
        note('native two app tabs: default production initializer retains both conflicting drafts and reload after missing legacy key');
        // Change identity before seeding so the first switch is guaranteed to produce an event.
        await open(second);
        const a = await seed(first, 'Boundary owner A question');
        const b = await seed(second, 'Boundary owner B question');
        const caps = await env.api(prefix + 'capabilities', undefined, first.token);
        await command(env, first, 'memories/save', { expectedVersion: 0, expectedMemoryUseEpoch: caps.memoryUseEpoch,
            text: 'Owner A saved preference', key: crypto.randomUUID(), type: 'preference', intent: 'add' });
        const beforeA = await snapshot(first, a.id);
        const beforeB = await snapshot(second, b.id);
        const writes = [];
        page.on('request', request => {
            if (request.method() === 'POST' && /assistant\/(conversations\/(clear|delete|update)|memories\/(save|delete))\.json/.test(request.url())) writes.push(request.url());
        });
        for (const editor of ['confirmation', 'history']) {
            await open(first, a.url);
            await page.getByRole('heading', { name: 'Verified isolated ledger' }).waitFor();
            if (editor === 'confirmation') {
                await requestDelete(page);
            } else {
                await page.getByRole('button', { name: 'Conversation history', exact: true }).click();
                await history(page).getByRole('button', { name: 'Rename', exact: true }).click();
                await history(page).getByLabel('Conversation title', { exact: true }).fill('Unsubmitted old owner title');
            }
            const frame = await nativeIdentitySwitch(page, switcher, second);
            await question(page).waitFor();
            assert.equal(frame.dialogs, 0);
            assert.ok(!frame.text.includes('Verified isolated ledger'));
            assert.ok(!frame.text.includes('Boundary owner A question'));
            assert.ok(!frame.text.includes('Owner A saved preference'));
            assert.ok(!frame.inputs.some(text => text.includes('Unsubmitted old owner')));
            assert.equal(await page.locator('.memories form, .history-item form, .confirmation-dialog[open]').count(), 0);
            await open(second, b.url);
            if (editor === 'history') {
                await page.getByRole('button', { name: 'Conversation history', exact: true }).click();
                await history(page).getByRole('button', { name: 'Rename', exact: true }).click();
                assert.equal(await history(page).getByLabel('Conversation title', { exact: true }).inputValue(), beforeB.conversation.title);
            }
            assert.deepEqual(await snapshot(first, a.id), beforeA);
            assert.deepEqual(await snapshot(second, b.id), beforeB);
            assert.equal(writes.length, 0);
            note(`native storage switch with ${editor} open: first receiving RAF clears old UI; both owners' server objects unchanged; no write request`);
        }
        // Removed controls must stay unreachable across identities and route changes.
        for (const boundary of ['owner', 'route']) {
            await open(first, a.url);
            assert.equal(await page.locator('.cloud-toolbar summary, .memories, .export-dialog').count(), 0);
            assert.equal(await page.getByRole('button', { name: /Export|Long-term memory|Clear all assistant data/ }).count(), 0);
            const downloads = [];
            const onDownload = download => downloads.push(download.suggestedFilename());
            page.on('download', onDownload);
            await requestDelete(page);
            const staleConfirm = await page.getByRole('dialog', { name: 'Confirm', exact: true }).getByRole('button', { name: 'Confirm', exact: true }).elementHandle();
            if (boundary === 'owner') {
                await nativeIdentitySwitch(page, switcher, second);
            } else {
                const previousURL = page.url();
                await page.goBack({ waitUntil: 'domcontentloaded' });
                assert.notEqual(page.url(), previousURL);
            }
            await page.waitForFunction(() => !document.querySelector('.confirmation-dialog[open]'));
            await staleConfirm.evaluate(button => button.click());
            await page.evaluate(() => new Promise(resolve => requestAnimationFrame(resolve)));
            assert.deepEqual(downloads, []);
            page.off('download', onDownload);
            assert.deepEqual(await snapshot(first, a.id), beforeA);
            assert.deepEqual(await snapshot(second, b.id), beforeB);
            assert.equal(writes.length, 0);
            note(`pending single deletion then ${boundary} change: stale control cannot mutate either owner; removed export/memory controls absent and no download`);
            if (boundary === 'route' && variant === 'mobile' && !await question(page).isVisible()) {
                await page.getByRole('link').filter({ hasText: 'Private assistant for personal bills and bookkeeping suggestions' }).click();
                await question(page).waitFor();
            }
            await open(second, b.url);
        }
        // Same-owner deletion epoch invalidates a pending confirmation in another live app tab.
        await open(first, a.url);
        await switcher.goto(a.url); await switcher.reload({ waitUntil: 'domcontentloaded' }); await question(switcher).waitFor();
        await requestDelete(page);
        const staleConfirm = await page.getByRole('dialog', { name: 'Confirm', exact: true }).getByRole('button', { name: 'Confirm', exact: true }).elementHandle();
        await command(env, first, 'conversations/clear');
        // Exercise the retained server epoch boundary without restoring the removed bulk-delete UI.
        await page.evaluate(() => window.dispatchEvent(new Event('online')));
        await page.waitForFunction(() => !document.querySelector('.confirmation-dialog[open]'));
        const freshCaps = await env.api(prefix + 'capabilities', undefined, first.token);
        assert.ok(freshCaps.deletionEpoch > caps.deletionEpoch);
        await command(env, first, 'memories/save', { expectedVersion: 0, expectedMemoryUseEpoch: freshCaps.memoryUseEpoch,
            text: 'New epoch retained memory', key: crypto.randomUUID(), type: 'preference', intent: 'add' });
        // Invoke the retained old DOM control after invalidation; it must not revive its old permission.
        await staleConfirm.evaluate(button => button.click());
        await page.evaluate(() => new Promise(resolve => requestAnimationFrame(resolve)));
        const retained = await env.api(prefix + 'memories/list', undefined, first.token);
        assert.deepEqual(retained.items.map(item => item.text), ['New epoch retained memory']);
        assert.equal((await env.api(prefix + 'capabilities', undefined, first.token)).deletionEpoch, freshCaps.deletionEpoch);
        assert.deepEqual(await snapshot(second, b.id), beforeB);
        assert.equal(writes.length, 0);
        note('server API clear and capability refresh advance epoch: old confirmation declines; retained old control cannot delete new epoch memory; other owner unchanged');
        assert.equal(errors.length, 0, errors.join('\n'));
    } catch (error) {
        await page.screenshot({ path: resolve(env.artifacts, `${variant}-identity-failure.png`) });
        console.error({ variant, pages: await page.locator('.assistant-cloud-page').evaluateAll(pages => pages.map(page => ({ classes: page.className, rect: page.getBoundingClientRect().toJSON(), transform: getComputedStyle(page).transform }))), url: page.url(), body: (await page.locator('body').innerText()).slice(0, 2000), errors });
        throw error;
    } finally { await context.close(); }
}
