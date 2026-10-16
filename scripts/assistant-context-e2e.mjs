import { createRequire } from 'node:module';
import assert from 'node:assert/strict';
import { mkdir, writeFile } from 'node:fs/promises';
import { environment } from './assistant-experience-e2e-environment.mjs';
const require = createRequire(import.meta.resolve('@vitest/browser-playwright'));
const { chromium } = require('playwright');
const artifacts = process.env.EBK_E2E_ARTIFACTS || '.vitest/assistant-context';
const currentYear = new Date().getUTCFullYear();
const periods = [
    { name: 'baseline', start_time: `${currentYear - 1}-01-01T00:00:00Z`, end_time: `${currentYear}-01-01T00:00:00Z` },
    { name: 'current', start_time: `${currentYear}-01-01T00:00:00Z`, end_time: `${currentYear + 1}-01-01T00:00:00Z` }
];
function expand(value) {
    if (Array.isArray(value)) return value.map(expand);
    if (!value || typeof value !== 'object') return value;
    if (value.format === 'column_rows') return value.rows.map(row => Object.fromEntries(value.columns.map((key, i) => [key, expand(row[i])])));
    return Object.fromEntries(Object.entries(value).map(([key, item]) => [key, expand(item)]));
}
let checked = 0;
const env = await environment({ respondToProvider: async ({ request, res }) => {
    const latest = request.messages.findLastIndex(message => message.role === 'user');
    const outputs = request.messages.slice(latest + 1).filter(message => message.role === 'tool');
    res.setHeader('Content-Type', 'application/json');
    if (!outputs.length) {
        const calls = [
            { id: 'distribution', type: 'function', function: { name: 'aggregate_transactions', arguments: JSON.stringify({ filter: { types: ['expense'] }, periods, group_by: ['category', 'month'], comparison: { baseline: 'baseline', current: 'current' }, page_size: 200 }) } },
            { id: 'totals', type: 'function', function: { name: 'aggregate_transactions', arguments: JSON.stringify({ filter: { types: ['expense'] }, periods, comparison: { baseline: 'baseline', current: 'current' } }) } }
        ];
        res.end(JSON.stringify({ choices: [{ index: 0, finish_reason: 'tool_calls', message: { role: 'assistant', tool_calls: calls } }] }));
        return true;
    }
    assert.equal(outputs.length, 2, 'Both calls in a large-result batch must finish');
    for (const output of outputs) {
        const result = expand(JSON.parse(output.content));
        assert.equal(result.ok, true, 'A later tool must not be starved by the first result');
        assert.equal(result.totals_coverage, 'all_matching');
        const sums = Object.fromEntries(result.totals.map(row => [row.period, row.sum]));
        assert.equal(sums.baseline, '5000.00');
        assert.equal(sums.current, '10123.45');
        if (output.tool_call_id === 'distribution') {
            assert.equal(result.has_more, true);
            assert.equal(result.coverage, 'page');
            assert(result.next_cursor, 'Reduced result pages must retain continuation');
            assert(result.data.length > 0);
        }
    }
    checked++;
    assert.equal(request.stream, true);
    res.setHeader('Content-Type', 'text/event-stream');
    const frame = (content, finish_reason = null) => `data: ${JSON.stringify({ choices: [{ index: 0, delta: { content }, finish_reason }] })}\n\n`;
    res.write(frame('已核对完整范围：去年支出 **CNY 5000.00**，今年支出 **CNY 10123.45**。'));
    res.write(frame('分类明细当前为部分分组，不能据此断言全部分类排名。', 'stop'));
    res.end('data: [DONE]\n\n');
    return true;
} });
let browser;
const results = [];
try {
    browser = await chromium.launch();
    await mkdir(artifacts, { recursive: true });
    for (const variant of ['desktop', 'mobile']) {
        const identity = await env.user(`context_${variant}`);
        const account = await env.api('v1/accounts/add', { name: 'Synthetic historical cash', category: 1, type: 1, icon: '1', color: '333333', currency: 'CNY', balance: '10000000', balanceTime: Date.UTC(currentYear - 2, 0, 1) / 1000 }, identity.token);
        const parent = await env.api('v1/transaction/categories/add', { name: 'Synthetic comparison', type: 2, icon: '1', color: '333333' }, identity.token);
        for (let i = 0; i < 50; i++) {
            const category = await env.api('v1/transaction/categories/add', { name: `合成支出分类 ${i}`, type: 2, parentId: parent.id, icon: '1', color: '333333' }, identity.token);
            for (const [year, amount] of [[currentYear - 1, 10000], [currentYear, 20000]]) {
                await env.api('v1/transactions/add', { type: 3, categoryId: category.id, time: Date.UTC(year, i % 9, 10) / 1000 + i, sourceAccountId: account.id, sourceAmount: amount, comment: 'Synthetic year comparison only' }, identity.token);
            }
        }
        const context = await browser.newContext({ viewport: variant === 'desktop' ? { width: 1280, height: 900 } : { width: 390, height: 844 }, timezoneId: 'UTC' });
        await context.addInitScript(login => {
            localStorage.setItem('ebk_user_token', login.token);
            localStorage.setItem('ebk_user_info', JSON.stringify(login.user));
        }, identity);
        const page = await context.newPage();
        const errors = [];
        page.on('pageerror', error => errors.push(error.message));
        await page.goto(`${env.url}/${variant}.html${variant === 'mobile' ? '#!' : '#'}/assistant`);
        const input = page.getByRole('textbox', { name: 'Ask your personal finance question' });
        await input.waitFor({ timeout: 60000 });
        await input.fill(`对比${currentYear}年和${currentYear - 1}年全年的支出，分析消费变化。`);
        await page.getByRole('button', { name: 'Send', exact: true }).click();
        const answer = page.locator('.answer').last();
        try {
            await answer.getByText('Answer completed', { exact: true }).waitFor({ timeout: 60000 });
        } catch (error) {
            throw env.providerError() || error;
        }
        const text = await answer.innerText();
        assert(text.includes('CNY 5000.00') && text.includes('CNY 10123.45'));
        assert.equal(await answer.locator('.run-error').count(), 0);
        assert.deepEqual(errors, []);
        await page.screenshot({ path: `${artifacts}/${variant}.png` });
        results.push({ variant, scenario: 'large category/month comparison and second aggregate in one turn complete with exact whole-range totals', status: 'PASS' });
        await context.close();
    }
    assert.equal(checked, 2);
    await writeFile(`${artifacts}/results.json`, JSON.stringify({ results, provider: 'deterministic HTTP fixture; real SQLite tools and browser, no real-model quality claim' }, null, 2));
    console.log(JSON.stringify(results));
} finally {
    const cleanup = await Promise.allSettled([browser?.close(), env.close()]);
    const failures = cleanup.filter(result => result.status === 'rejected');
    if (failures.length) throw new AggregateError(failures.map(result => result.reason), 'Context E2E cleanup failed');
}
