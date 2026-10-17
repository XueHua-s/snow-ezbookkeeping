import { createRequire } from 'node:module';
import { mkdir, writeFile } from 'node:fs/promises';
import assert from 'node:assert/strict';
import { environment } from './assistant-experience-e2e-environment.mjs';
const require = createRequire(import.meta.resolve('@vitest/browser-playwright'));
const { chromium } = require('playwright');
const artifacts = process.env.EBK_E2E_ARTIFACTS || '.vitest/assistant-streaming';
let release;
let answerStarted = false;
let answerClosed = false;
const env = await environment({
    streamAnswer: true,
    onAnswerStarted: () => { answerStarted = true; },
    onProviderResponseClose: event => { if (answerStarted && !event.completed) answerClosed = true; },
    beforeAnswerCompleted: () => new Promise(resolve => { release = resolve; })
});
const browser = await chromium.launch();
const results = [];
try {
    await mkdir(artifacts, { recursive: true });
    for (const variant of ['desktop', 'mobile']) {
        const identity = await env.user(`stream_${variant}`);
        const context = await browser.newContext({ viewport: variant === 'desktop' ? { width: 1280, height: 900 } : { width: 390, height: 844 } });
        await context.addInitScript(login => {
            localStorage.setItem('ebk_user_token', login.token);
            localStorage.setItem('ebk_user_info', JSON.stringify(login.user));
        }, identity);
        const page = await context.newPage();
        const errors = [];
        page.on('pageerror', error => errors.push(error.message));
        const url = `${env.url}/${variant}.html${variant === 'mobile' ? '#!' : '#'}/assistant`;
        await page.goto(url);
        await page.getByRole('textbox', { name: 'Ask your personal finance question' }).waitFor({ timeout: 60000 });
        for (const stop of [false, true]) {
            answerStarted = false;
            answerClosed = false;
            const input = page.getByRole('textbox', { name: 'Ask your personal finance question' });
            await input.fill(stop ? 'Check my expense and let me stop midway' : 'Check my isolated expense and stream the answer');
            await page.getByRole('button', { name: 'Send', exact: true }).click();
            const answer = page.locator('.answer').last();
            await answer.locator('.assistant-markdown').getByText('Live verified answer: CNY 123.45.', { exact: true }).waitFor({ timeout: 30000 });
            assert(answerStarted && typeof release === 'function');
            assert(!answerClosed);
            assert.equal(await answer.locator('.answer-complete').count(), 0);
            assert.equal(await answer.locator('.generation-indicator').count(), 1);
            await page.screenshot({ path: `${artifacts}/${variant}-${stop ? 'cancel' : 'live'}.png` });
            if (stop) {
                await page.getByRole('button', { name: 'Stop generating', exact: true }).click();
                await answer.getByText('Assistant request canceled', { exact: true }).waitFor({ timeout: 15000 });
                assert((await answer.innerText()).includes('Live verified answer: CNY 123.45.'));
                for (let i = 0; i < 100 && !answerClosed; i++) await new Promise(resolve => setTimeout(resolve, 50));
                assert(answerClosed, 'Stop must close the upstream response before it completes');
                release();
            } else {
                // Reload while the upstream is still blocked: saved partial text must recover.
                await page.reload();
                const restored = page.locator('.answer').last();
                await restored.locator('.assistant-markdown').getByText('Live verified answer: CNY 123.45.', { exact: true }).waitFor({ timeout: 30000 });
                assert.equal(await restored.locator('.answer-complete').count(), 0);
                release();
                await restored.getByText('Answer completed', { exact: true }).waitFor({ timeout: 15000 });
                const text = await restored.locator('.assistant-markdown').innerText();
                assert.equal((text.match(/Live verified answer:/g) || []).length, 1);
                assert(text.includes('Completed after the held upstream response.'));
            }
            results.push({ variant, scenario: stop ? 'live partial -> stop -> upstream canceled' : 'live partial before finish -> reload -> complete without duplicate', status: 'PASS' });
        }
        assert.deepEqual(errors, []);
        await context.close();
    }
    await writeFile(`${artifacts}/results.json`, JSON.stringify({ results, provider: 'held synthetic HTTP SSE; real app, persistence, subscription and browser' }, null, 2));
    console.log(JSON.stringify(results));
} finally {
    release?.();
    await browser.close();
    await env.close();
}
