import { createRequire } from 'node:module';
import { mkdir, writeFile } from 'node:fs/promises';
import assert from 'node:assert/strict';
import { environment } from './assistant-experience-e2e-environment.mjs';
const require = createRequire(import.meta.resolve('@vitest/browser-playwright'));
const { chromium } = require('playwright');
const artifacts = process.env.EBK_E2E_ARTIFACTS || 'docs/ai-assistant-ui-review/gate10';
const production = process.env.EBK_E2E_PRODUCTION === '1';
const longWord = 'synthetic-long-reference-'.repeat(18);
const answerContent = `## Monthly review\n\nA deliberately long synthetic ledger reference: ${longWord}\n\n| Account | Currency | Amount | Reference |\n| --- | --- | ---: | --- |\n| Cash | CNY | 123.45 | ${longWord} |\n\n` + 'Financial reading remains calm and legible. '.repeat(35);
const env = await environment({ answerContent, production });
let browser;
const results = [];
try {
    browser = await chromium.launch();
    await mkdir(artifacts, { recursive: true });
    for (const variant of ['desktop', 'mobile']) {
        const identity = await env.user(`design_${variant}`);
        const context = await browser.newContext({ viewport: variant === 'desktop' ? { width: 720, height: 900 } : { width: 390, height: 844 } });
        await context.addInitScript(login => {
            localStorage.setItem('ebk_user_token', login.token);
            localStorage.setItem('ebk_user_info', JSON.stringify(login.user));
        }, identity);
        const page = await context.newPage();
        const errors = [];
        page.on('pageerror', error => errors.push(error.message));
        await page.goto(`${env.url}/${variant}${production ? '' : '.html'}${variant === 'mobile' ? '#!' : '#'}/assistant?conversationId=`);
        const input = page.getByRole('textbox', { name: 'Ask your personal finance question' });
        await input.waitFor({ timeout: 60000 });
        await input.fill('Review long synthetic references and a financial table');
        await page.getByRole('button', { name: 'Send', exact: true }).click();
        await page.getByRole('heading', { name: 'Monthly review' }).waitFor({ timeout: 30000 });
        await page.waitForFunction(() => document.querySelector('.experience-messages')?.clientHeight > 300);
        await page.locator('.experience-messages').evaluate(panel => { panel.scrollTop = 0; });
        await page.getByRole('button', { name: 'Conversation history', exact: true }).focus();
        await page.keyboard.press('Tab');
        const metrics = await page.locator('.cloud-toolbar').evaluate(toolbar => ({
            titleWidth: toolbar.querySelector('.toolbar-title').getBoundingClientRect().width,
            controls: [...toolbar.querySelectorAll(':scope > button, :scope > details > summary')].map(el => ({ name: el.getAttribute('aria-label'), width: el.getBoundingClientRect().width, height: el.getBoundingClientRect().height })),
            viewport: innerWidth,
            documentWidth: document.documentElement.scrollWidth,
            table: (() => { const el = document.querySelector('.assistant-markdown table'); return { width: el.clientWidth, scrollWidth: el.scrollWidth, overflow: getComputedStyle(el).overflowX, numerals: getComputedStyle(el).fontVariantNumeric }; })()
        }));
        assert(metrics.titleWidth >= 150);
        assert(metrics.controls.every(control => control.name && control.width >= 44 && control.height >= 44));
        assert(metrics.documentWidth <= metrics.viewport);
        assert.equal(metrics.table.overflow, 'auto');
        assert.equal(metrics.table.numerals, 'tabular-nums');
        assert.equal(errors.length, 0);
        await page.screenshot({ path: `${artifacts}/${variant}-long-focus.png` });
        results.push({ variant, status: 'PASS', metrics, errors });
        await context.close();
    }
    await writeFile(`${artifacts}/long-review.json`, JSON.stringify(results, null, 2));
} finally {
    try { await browser?.close(); } finally { await env.close(); }
}
