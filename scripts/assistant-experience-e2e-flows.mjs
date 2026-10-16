import assert from 'node:assert/strict';
import { identityBoundary } from './assistant-experience-e2e-boundaries.mjs';
import { resolve } from 'node:path';

const assistant = 'v1/llm/assistant/';
async function command(env, identity, path, payload) {
    const caps = await env.api(assistant + 'capabilities', undefined, identity.token);
    return env.api(assistant + path, { clientRequestId: crypto.randomUUID(), requestIssuedAt: new Date().toISOString(), expectedDeletionEpoch: caps.deletionEpoch, ...payload }, identity.token);
}
export async function extendedFlows({ env, browser, context, page, identity, variant, conversationURL, results, production }) {
    const note = scenario => results.push({ variant, scenario, status: 'PASS' });
    const input = page.getByPlaceholder('Ask your personal finance question');
    const id = new URLSearchParams(conversationURL.split('?')[1]).get('conversationId');
    const second = await browser.newContext({ viewport: variant === 'desktop' ? { width: 1440, height: 960 } : { width: 390, height: 844 } });
    await second.addInitScript(login => { localStorage.setItem('ebk_user_token', login.token); localStorage.setItem('ebk_user_info', JSON.stringify(login.user)); }, identity);
    const other = await second.newPage(); await other.goto(conversationURL, { waitUntil: 'domcontentloaded', timeout: 120000 });
    await other.getByRole('heading', { name: 'Verified isolated ledger' }).waitFor({ timeout: 30000 });
    assert.equal(await other.getByPlaceholder('Ask your personal finance question').inputValue(), '');
    const state = await page.evaluate(async id => (await (await fetch(`/api/v1/llm/assistant/conversations/get.json?id=${id}`, { headers: { Authorization: `Bearer ${localStorage.getItem('ebk_user_token')}` } })).json()).result, id);
    await command(env, identity, 'conversations/update', { id, expectedMetadataVersion: state.conversation.metadataVersion, title: 'Changed from second device' });
    await page.getByText('Changed from second device', { exact: true }).first().waitFor({ timeout: 22000 });
    assert.equal(await input.inputValue(), 'Draft preserved through bill detail'); note('second device reads same history; idle foreground synchronizes rename without replacing local draft');
    await second.close();

    await page.locator('.experience-actions').filter({ has: page.getByRole('button', { name: 'Regenerate', exact: true }) }).getByRole('button', { name: 'Copy Markdown', exact: true }).click();
    assert.equal(await page.evaluate(() => navigator.clipboard.readText()), state.messages[0].selectedAnswer.content);
    note('real browser clipboard receives exact selected original Markdown');

    const beforeOffline = env.calls(); await context.setOffline(true);
    await input.fill('Offline draft is never auto sent');
    await page.getByText('Offline. Drafts are saved on this device.', { exact: true }).first().waitFor();
    await context.setOffline(false); await page.waitForFunction(() => navigator.onLine);
    await page.waitForTimeout(1200); assert.equal(env.calls(), beforeOffline); assert.equal(await input.inputValue(), 'Offline draft is never auto sent');
    note('offline edits and online recovery do not send a model request');

    const beforeDisabled = env.calls();
    await env.setGeneration(false);
    await page.evaluate(() => window.dispatchEvent(new Event('online')));
    await page.locator('.composer-notes [role="status"]').waitFor({ timeout: 22000 });
    await page.getByRole('heading', { name: 'Verified isolated ledger' }).scrollIntoViewIfNeeded();
    assert.ok(await page.getByRole('heading', { name: 'Verified isolated ledger' }).isVisible());
    assert.equal(await input.inputValue(), 'Offline draft is never auto sent');
    assert.equal(await page.getByRole('button', { name: 'Send', exact: true }).isDisabled(), true);
    assert.equal(await page.locator('.cloud-toolbar summary').count(), 0);
    assert.equal(await page.getByRole('button', { name: /Export|Long-term memory|Clear all assistant data/ }).count(), 0);
    assert.equal(env.calls(), beforeDisabled);
    await env.setGeneration(true); await page.evaluate(() => window.dispatchEvent(new Event('online')));
    await page.waitForFunction(() => !document.querySelector('.composer-notes [role="status"]'));
    note('server generation disabled across restart retains history/draft and does not call provider');


    note('memory management, export and bulk deletion are absent from the assistant UI');

    await page.getByRole('button', { name: 'Conversation history', exact: true }).click();
    const history = page.getByRole('navigation', { name: 'Conversation history', exact: true });
    await history.getByLabel('Search history', { exact: true }).fill('isolated expense');
    await history.locator('.result').first().waitFor(); await page.screenshot({ path: resolve(env.artifacts, `${variant}-search.png`) });
    await history.getByLabel('Search history', { exact: true }).fill('');
    await history.getByRole('button', { name: 'Archive', exact: true }).click();
    await history.getByText('No conversations yet', { exact: true }).waitFor();
    await history.getByLabel('Include archived', { exact: true }).check();
    await history.getByRole('button', { name: 'Unarchive', exact: true }).click();
    await history.getByRole('button', { name: 'Archive', exact: true }).waitFor();
    assert.equal(await history.locator('.history-item').count(), 1);
    await history.getByLabel('Include archived', { exact: true }).uncheck();
    await history.locator('.history-item').first().waitFor();
    await history.getByRole('button', { name: 'Close', exact: true }).click();
    // A leaving native modal keeps the composer inert until its transition finishes.
    await page.waitForFunction(() => !document.querySelector('.cloud-history'));
    note('search and archive/unarchive; Include archived retains active conversation after unarchive');

    const beforeUnknown = env.calls();
    await input.fill('E2E_SLOW uncertain response');
    let dropped;
    const accepted = new Promise(resolve => { dropped = resolve; });
    let interceptedSend;
    await page.route('**/api/v1/llm/assistant/messages/send.json', async route => { interceptedSend = route.request().postDataJSON(); await route.fetch(); await route.abort('failed'); dropped(); });
    await page.getByRole('button', { name: 'Send', exact: true }).click(); await accepted;
    await page.reload(); await page.unroute('**/api/v1/llm/assistant/messages/send.json');
    await page.getByText('E2E_SLOW uncertain response', { exact: true }).waitFor().catch(async error => {
        const server = await env.api(`${assistant}conversations/get?id=${id}`, undefined, identity.token);
        console.error('Accepted send recovery mismatch', { variant, interceptedSend, serverMessages: server.messages.map(message => ({ id: message.id, content: message.content })), url: page.url(), body: await page.locator('.cloud-workspace').innerText(), draft: await input.inputValue() });
        throw error;
    });
    await page.getByRole('heading', { name: 'Verified isolated ledger' }).nth(1).waitFor({ timeout: 30000 });
    const recovered = await page.evaluate(async id => (await (await fetch(`/api/v1/llm/assistant/conversations/get.json?id=${id}`, { headers: { Authorization: `Bearer ${localStorage.getItem('ebk_user_token')}` } })).json()).result, id);
    assert.equal(recovered.messages.filter(message => message.content === 'E2E_SLOW uncertain response').length, 1);
    assert.equal(env.calls() - beforeUnknown, 2);
    note('lost accepted send response and refresh reconcile the original command without duplicate messages');

    await input.fill(`E2E_FAIL_ONCE ${variant} retry keeps the question`);
    await page.getByRole('button', { name: 'Send', exact: true }).click();
    await page.getByText('Assistant request failed', { exact: true }).first().waitFor({ timeout: 30000 });
    await page.screenshot({ path: resolve(env.artifacts, `${variant}-failure.png`) });
    await page.locator('.answer').last().getByRole('button', { name: 'Retry', exact: true }).click();
    await page.getByRole('heading', { name: 'Verified isolated ledger' }).nth(2).waitFor({ timeout: 30000 });
    const retryDetail = await env.api(`${assistant}conversations/get?id=${id}`, undefined, identity.token);
    assert.equal(retryDetail.messages.length, 3); assert.equal(retryDetail.messages.at(-1).versions.length, 2);
    note('provider failure stays visible and explicit retry creates another answer attempt without duplicating question');

    await page.locator('.question').first().getByRole('button', { name: 'Edit and create branch', exact: true }).click();
    await input.fill('Edited branch expense question');
    const beforeBranch = env.calls();
    let branchAccepted;
    const receivedBranch = new Promise(resolve => { branchAccepted = resolve; });
    await page.route('**/api/v1/llm/assistant/commands/get.json*', route => route.abort('failed'));
    await page.route('**/api/v1/llm/assistant/conversations/branch.json', async route => { await route.fetch(); await route.abort('failed'); branchAccepted(); });
    await page.locator('.experience-composer').getByRole('button', { name: 'Edit and create branch', exact: true }).click();
    await receivedBranch;
    await page.unroute('**/api/v1/llm/assistant/commands/get.json*');
    await page.unroute('**/api/v1/llm/assistant/conversations/branch.json');
    await page.reload(); await input.waitFor();
    await page.waitForFunction(old => new URLSearchParams(location.hash.split('?')[1]).get('conversationId') !== old, id);
    assert.equal(await input.inputValue(), 'Edited branch expense question'); assert.equal(env.calls(), beforeBranch);
    await page.getByText('New branch created. Review your question below and send to start generating.', { exact: true }).waitFor();
    await page.screenshot({ path: resolve(env.artifacts, `${variant}-branch.png`) });
    await page.getByRole('button', { name: 'Send', exact: true }).click();
    await page.getByRole('heading', { name: 'Verified isolated ledger' }).waitFor();
    const branches = await env.api(assistant + 'conversations/list', undefined, identity.token);
    assert.equal(branches.items.length, 2);
    note('lost branch acceptance and refresh recover one target and edited draft; only explicit send calls provider');

    await identityBoundary({ env, page, identity, variant, results, attemptBFCache: production });
    await page.getByRole('button', { name: 'Conversation history', exact: true }).click();
    const deleteButton = history.getByRole('button', { name: 'Delete', exact: true }).first();
    await deleteButton.click();
    const confirmation = page.getByRole('dialog', { name: 'Confirm', exact: true });
    await confirmation.getByRole('button', { name: 'Cancel', exact: true }).click();
    assert.equal(await deleteButton.evaluate(el => el === document.activeElement), true);
    assert.equal((await env.api(assistant + 'conversations/list', undefined, identity.token)).items.length, 2);
    for (let remaining = 2; remaining > 0; remaining--) {
        if (!await history.isVisible()) await page.getByRole('button', { name: 'Conversation history', exact: true }).click();
        await history.getByRole('button', { name: 'Delete', exact: true }).first().click();
        await confirmation.getByRole('button', { name: 'Confirm', exact: true }).click();
        await page.waitForFunction(count => !document.querySelector('.cloud-history') || document.querySelectorAll('.history-item').length === count, remaining - 1).catch(async error => {
            const server = await env.api(assistant + 'conversations/list', undefined, identity.token);
            console.error('Conversation deletion did not reconcile', { variant, expectedRemaining: remaining - 1, serverIds: server.items.map(item => item.id), url: page.url(), historyText: await history.innerText(), alerts: await page.getByRole('alert').allTextContents() });
            throw error;
        });
        if (!await history.isVisible()) await page.getByRole('button', { name: 'Conversation history', exact: true }).click();
        await page.waitForFunction(count => document.querySelectorAll('.history-item').length === count, remaining - 1);
        await page.waitForFunction(() => !document.querySelector('.confirmation-dialog[open]'));
        if (remaining === 2) await page.waitForFunction(() => !new URLSearchParams(location.hash.split('?')[1]).get('conversationId'));
    }
    await history.getByRole('button', { name: 'Close', exact: true }).click();
    await page.getByText('How can I help with your finances?', { exact: true }).waitFor();
    const listing = await env.api(assistant + 'conversations/list', undefined, identity.token); assert.equal(listing.items.length, 0);
    const ledger = await env.api('v1/accounts/list', undefined, identity.token); assert.ok(JSON.stringify(ledger).includes(identity.fixtureIds.accountId));
    const bill = await env.api(`v1/transactions/get?id=${identity.fixtureIds.transactionId}`, undefined, identity.token); assert.equal(bill.sourceAmount, 12345);
    note('single-conversation deletion restores focus on cancellation and removes only confirmed conversations while retaining the ledger');
}
