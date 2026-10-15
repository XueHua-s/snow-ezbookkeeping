import { createRequire } from 'node:module';
import assert from 'node:assert/strict';
import { writeFile, mkdir, open } from 'node:fs/promises';
import { spawn } from 'node:child_process';
import { resolve, dirname } from 'node:path';
import { browserBoundaries } from './assistant-experience-e2e-boundaries.mjs';
import { cacheFlows } from './assistant-experience-e2e-cache.mjs';
import { extendedFlows } from './assistant-experience-e2e-flows.mjs';
import { environment } from './assistant-experience-e2e-environment.mjs';
const require = createRequire(import.meta.resolve('@vitest/browser-playwright'));
const { chromium } = require('playwright');
const production = process.env.EBK_E2E_PRODUCTION === '1';
const env = await environment({ production });
const browser = await chromium.launch({ headless: true, ignoreDefaultArgs: ['--disable-back-forward-cache'] });
const results = [];
await writeFile(resolve(env.artifacts, 'results.json'), JSON.stringify({ status: 'RUNNING', results }));
try {
    if (production) {
        const gatePath = process.env.EBK_E2E_PRODUCTION_GATE_LOG || resolve(env.artifacts, 'production-smoke.log');
        await mkdir(dirname(gatePath), { recursive: true }); const gateLog = await open(gatePath, 'w');
        try { await new Promise((resolve, reject) => { const check = spawn('npm', ['run', 'test:production'], { env: { ...process.env, EBK_SMOKE_BASE_URL: env.base }, stdio: ['ignore', gateLog.fd, gateLog.fd] }); check.on('exit', code => code === 0 ? resolve() : reject(new Error(`npm run test:production failed (${code}); see ${gatePath}`))); }); }
        finally { await gateLog.close(); }
        results.push({ scenario: 'npm run test:production against isolated current Go application serving production dist', status: 'PASS' });
    }
    for (const variant of ['desktop', 'mobile']) {
        const identity = await env.user(variant);
        const context = await browser.newContext({ permissions: ['clipboard-read', 'clipboard-write'], viewport: variant === 'desktop' ? { width: 1440, height: 960 } : { width: 390, height: 844 } });
        await context.addInitScript(login => { if (!localStorage.getItem('ebk_user_token')) { localStorage.setItem('ebk_user_token', login.token); localStorage.setItem('ebk_user_info', JSON.stringify(login.user)); } }, identity);
        const page = await context.newPage(); const errors = []; page.on('pageerror', error => errors.push(error.message));
        await page.goto(`${env.url}/${variant}${production ? '' : '.html'}${variant === 'mobile' ? '#!' : '#'} /assistant?conversationId=`.replace('# /', '#/').replace('#! /', '#!/'), { waitUntil: 'domcontentloaded', timeout: 120000 });
        const input = page.getByPlaceholder('Ask your personal finance question'); await input.waitFor({ timeout: 60000 });
        await page.screenshot({ path: resolve(env.artifacts, `${variant}-empty.png`) });
        await input.fill('Show the isolated expense'); await page.getByRole('button', { name: 'Send', exact: true }).click();
        await page.getByRole('heading', { name: 'Verified isolated ledger' }).waitFor({ timeout: 30000 });
        const conversationURL = page.url(); assert.match(conversationURL, /conversationId=\d+/);
        await input.fill('Draft preserved through bill detail');
        await page.screenshot({ path: resolve(env.artifacts, `${variant}-answer.png`) });
        await page.getByRole('button').filter({ hasText: 'View Details' }).first().scrollIntoViewIfNeeded();
        const readingOffset = await page.locator('.experience-messages').evaluate(async panel => {
            await new Promise(resolve => requestAnimationFrame(resolve)); return panel.scrollTop;
        });
        await page.getByRole('button').filter({ hasText: 'View Details' }).first().click();
        if (variant === 'mobile') { await page.waitForURL(/transaction\/detail/); await page.goBack(); }
        else { const dialog = page.getByRole('dialog'); await dialog.waitFor(); await page.keyboard.press('Escape'); }
        await input.waitFor(); assert.equal(await input.inputValue(), 'Draft preserved through bill detail');
        await page.waitForFunction(expected => Math.abs((document.querySelector('.experience-messages')?.scrollTop ?? -999) - expected) < 4, readingOffset);
        await page.reload(); await input.waitFor(); assert.equal(await input.inputValue(), 'Draft preserved through bill detail');
        await page.getByRole('heading', { name: 'Verified isolated ledger' }).waitFor();
        assert.equal(errors.length, 0, errors.join('\n'));
        if (production) await cacheFlows({ page, context, identity, variant, results });
        results.push({ variant, scenario: 'actual route send/detail/back restores reading offset and draft; refresh restores cloud answer', status: 'PASS' });
        await browserBoundaries({ env, browser, context, page, identity, variant, conversationURL, results });
        await extendedFlows({ env, browser, context, page, identity, variant, conversationURL, results, production });
        assert.equal(errors.length, 0, errors.join('\n'));
        await context.close();
    }
    await writeFile(resolve(env.artifacts, 'results.json'), JSON.stringify({ executedAt: new Date().toISOString(), frontend: production ? 'production dist with service worker' : 'Vite actual app routes', browser: browser.version(), status: results.some(result => result.status !== 'PASS') ? 'PARTIAL' : 'PASS', results, backend: process.env.EBK_E2E_BINARY ? 'explicit prebuilt development snapshot; rebuild required for final acceptance' : 'built from current worktree', providerCalls: env.calls(), provider: 'deterministic protocol fixture; not model certification' }, null, 2));
    console.log(JSON.stringify(results));
} catch (error) { for (const context of browser.contexts()) for (const page of context.pages()) await page.screenshot({ path: resolve(env.artifacts, 'failure.png') }).catch(() => {}); await writeFile(resolve(env.artifacts, 'results.json'), JSON.stringify({ status: 'FAIL', results, error: String(error), providerCalls: env.calls() }, null, 2)); throw error; } finally { await browser.close(); await env.close(); }
