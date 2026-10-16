import { createRequire } from 'node:module';
const require = createRequire(import.meta.resolve('@vitest/browser-playwright'));
const { chromium } = require('playwright');
import { writeFile } from 'node:fs/promises';
import { strict as assert } from 'node:assert';
import { historySmoke } from './assistant-store-history-smoke.mjs';
import { environment } from './assistant-experience-e2e-environment.mjs';

const env = await environment();
let browser;
try {
    browser = await chromium.launch({ headless: true });
    const account = await env.user('store');
    const page = await browser.newPage();
    page.on('pageerror', error => console.error('Browser runtime:', error.message));
    await page.addInitScript(({ token, user }) => {
        localStorage.setItem('ebk_user_token', token); localStorage.setItem('ebk_user_info', JSON.stringify(user));
    }, account);
    await page.goto(env.url + '/desktop.html#/overview', { waitUntil: 'domcontentloaded', timeout: 120000 });
    await page.waitForFunction(() => !!document.querySelector('.v-application'), undefined, { timeout: 60000 });
    const state = async (code, input) => page.evaluate(async ({ code, input }) => {
        const { useAssistantStore } = await import('/stores/assistant.ts');
        return new Function('store', 'input', `return (async () => { ${code} })()`)(useAssistantStore(), input);
    }, { code, input });
    await state('await store.initialize(); await store.open(); await store.setDraft("Store actual API E2E_SLOW"); return store.authenticated;');
    let sends = 0;
    await page.route('**/api/v1/llm/assistant/messages/send.json', async route => {
        sends++; await route.fetch(); await route.abort('failed');
    });
    await page.route('**/api/v1/llm/assistant/commands/get.json*', route => route.abort('failed'));
    const original = await state('try { await store.send(); } catch {} return store.local.pending[0];');
    assert(original?.clientRequestId); assert.equal(original.text, 'Store actual API E2E_SLOW');
    await page.reload({ waitUntil: 'domcontentloaded', timeout: 120000 });
    await page.waitForFunction(() => !!document.querySelector('.v-application'), undefined, { timeout: 60000 });
    const restored = await state('await store.initialize(); return store.local.pending[0];');
    assert.equal(restored.clientRequestId, original.clientRequestId);
    assert.equal(restored.requestIssuedAt, original.requestIssuedAt);
    await page.unroute('**/api/v1/llm/assistant/commands/get.json*');
    await state('await store.synchronize();');
    const accepted = await state('return { id: store.conversationId, pending: store.local.pending.length };');
    assert(accepted.id); assert.equal(accepted.pending, 0); assert.equal(sends, 1);
    for (let attempt = 0; attempt < 90; attempt++) {
        const finished = await state('await store.synchronize(); return !store.busy;');
        if (finished) break;
        await page.waitForTimeout(300);
    }
    assert.equal(await state('return store.detail.activeRun?.state ?? store.detail.messages.at(-1)?.selectedAnswer?.state;'), 'completed');
    await page.unroute('**/api/v1/llm/assistant/messages/send.json');
    const branch = await state('const user = store.detail.messages.at(-1); const result = await store.branch(user.id, "edit", "E2E_FAIL edited question"); return { id: result.conversationId, draft: store.draft };');
    assert.notEqual(branch.id, accepted.id); assert.equal(branch.draft, 'E2E_FAIL edited question');
    await state('await store.send();');
    for (let attempt = 0; attempt < 60; attempt++) {
        const finished = await state('await store.synchronize(); return !store.busy;');
        if (finished) break;
        await page.waitForTimeout(300);
    }
    const failed = await state('return { id: store.conversationId, state: store.detail.activeRun?.state ?? store.detail.messages.at(-1)?.selectedAnswer?.state };');
    assert.equal(failed.id, branch.id); assert.equal(failed.state, 'failed');
    await state('const c = store.detail.conversation; await store.execute("delete", { id: c.id, expectedMetadataVersion: c.metadataVersion });');
    assert.equal(await state('return store.conversationId;'), '');
    const history = await historySmoke({ env, page, state, conversationId: accepted.id });
    const secondContext = await browser.newContext(); const second = await secondContext.newPage();
    await second.addInitScript(({ token, user }) => { localStorage.setItem('ebk_user_token', token); localStorage.setItem('ebk_user_info', JSON.stringify(user)); }, account);
    await second.goto(`${env.url}/desktop.html#/assistant?conversationId=${accepted.id}`, { waitUntil: 'domcontentloaded', timeout: 120000 });
    await second.waitForFunction(() => document.querySelectorAll('.experience-messages .turn').length === 30);
    await second.evaluate(async () => { const { useAssistantStore } = await import('/stores/assistant.ts'); await useAssistantStore().setDraft('second device local draft'); });
    await state('const c = store.detail.conversation; await store.execute("delete", { id: c.id, expectedMetadataVersion: c.metadataVersion });');
    const remoteDelete = await second.evaluate(async () => { const { useAssistantStore } = await import('/stores/assistant.ts'); const store = useAssistantStore(); await store.synchronize(); return { id: store.conversationId, count: store.detail?.messages.length ?? 0, draft: store.draft, localDrafts: store.local.drafts.length }; });
    assert.deepEqual(remoteDelete, { id: '', count: 0, draft: '', localDrafts: 0 });
    await second.waitForFunction(() => document.querySelectorAll('.experience-messages .turn').length === 0);
    await secondContext.close();
    const verification = { backend: 'current host Go build and isolated SQLite (not Docker)', frontend: 'Vite current source with real store module imports', browser: browser.version(), history, remoteDelete, passed: ['persist-before-HTTP', 'unknown-response-refresh-original-ID', 'confirm-without-resend', 'same-server-run-completed', 'two-stage-branch-failed-first-send-retains-target', 'delete-clears-current', 'independent-browser-context-remote-delete-clears-DOM-and-draft'], sends, providerCalls: env.calls(), artifacts: env.artifacts };
    await writeFile(env.artifacts + '/store-results.json', JSON.stringify(verification, null, 2));
    console.log(JSON.stringify(verification));
} finally { try { await browser?.close(); } finally { await env.close(); } }
