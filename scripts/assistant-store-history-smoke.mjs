import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';

export async function historySmoke({ env, page, state, conversationId }) {
    const fixture = JSON.parse(execFileSync('python3', ['scripts/assistant-store-history-fixture.py', env.directory + '/fixture.db', conversationId], { encoding: 'utf8' }));
    await page.goto(`${env.url}/desktop.html#/assistant?conversationId=${conversationId}`, { waitUntil: 'domcontentloaded', timeout: 120000 });
    await page.getByPlaceholder('Ask your personal finance question').waitFor({ timeout: 60000 });
    try { await page.waitForFunction(() => document.querySelectorAll('.experience-messages .turn').length === 30); } catch (error) {
        console.error(JSON.stringify({ url: page.url(), fixture, state: await state('return { id: store.conversationId, count: store.detail?.messages.length, error: String(store.error), pending: store.local.pending.map(item => ({ action: item.action, target: item.resolvedConversationId })) };'), text: await page.locator('body').innerText() }));
        await page.screenshot({ path: env.artifacts + '/history-failure.png' }); throw error;
    }
    const panel = page.locator('.experience-messages');
    const capture = () => panel.evaluate(element => {
        const top = element.getBoundingClientRect().top;
        const item = Array.from(element.querySelectorAll('[data-message-id]')).find(node => node.getBoundingClientRect().bottom > top);
        return { id: item?.dataset.messageId, offset: item?.getBoundingClientRect().top - top, count: element.querySelectorAll('.turn').length };
    });
    const anchors = [];
    for (let index = 0; index < 3; index++) {
        await panel.evaluate(element => { element.scrollTop = 0; }); await page.waitForTimeout(100);
        const before = await capture();
        await page.getByRole('button', { name: 'Load older messages', exact: true }).click();
        await page.waitForFunction(expected => document.querySelectorAll('.experience-messages .turn').length === expected, Math.min(90, (index + 2) * 30));
        await page.waitForTimeout(150);
        const after = await panel.evaluate((element, id) => {
            const item = Array.from(element.querySelectorAll('[data-message-id]')).find(node => node.dataset.messageId === id);
            return { id: item?.dataset.messageId, offset: item?.getBoundingClientRect().top - element.getBoundingClientRect().top, count: element.querySelectorAll('.turn').length };
        }, before.id);
        console.log(JSON.stringify({ index, before, after }));
        assert.equal(after.id, before.id, `prepending page ${index} changed message anchor`);
        assert(Math.abs(after.offset - before.offset) <= 2, `prepending page ${index} moved anchor by ${after.offset - before.offset}px`);
        anchors.push({ before, after });
    }
    const bounded = await state('return { count: store.detail.messages.length, first: store.detail.messages[0].turnSequence, last: store.detail.messages.at(-1).turnSequence };');
    assert.equal(bounded.count, 90); assert.equal(bounded.last, 90);
    assert.equal(await page.locator('.experience-messages .turn').count(), 90);
    await page.getByRole('button', { name: /Back to latest/ }).first().click();
    await page.waitForTimeout(250);
    const latest = await state('return store.detail.messages.at(-1).turnSequence;');
    assert.equal(latest, 120, 'Back to latest must load the real tail after bounded-window eviction');
    await panel.evaluate(element => { element.scrollTop = 1200; }); await page.waitForTimeout(150);
    const reflowBefore = await capture();
    await page.setViewportSize({ width: 1100, height: 760 }); await page.waitForTimeout(200);
    const reflowAfter = await panel.evaluate((element, id) => { const item = Array.from(element.querySelectorAll('[data-message-id]')).find(node => node.dataset.messageId === id); return { id: item?.dataset.messageId, offset: item?.getBoundingClientRect().top - element.getBoundingClientRect().top }; }, reflowBefore.id);
    assert(Math.abs(reflowAfter.offset - reflowBefore.offset) <= 2, 'viewport resize moved reading anchor');
    return { fixture, anchors, bounded, latest, viewportReflow: { before: reflowBefore, after: reflowAfter }, viewport: await page.viewportSize() };
}
