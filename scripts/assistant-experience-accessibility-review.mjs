import { createRequire } from 'node:module';
import { mkdir, writeFile } from 'node:fs/promises';
import assert from 'node:assert/strict';
import { environment } from './assistant-experience-e2e-environment.mjs';
const require = createRequire(import.meta.resolve('@vitest/browser-playwright'));
const { chromium } = require('playwright');
const artifacts = process.env.EBK_E2E_ARTIFACTS || 'docs/ai-assistant-ui-review/gate11';
const production = process.env.EBK_E2E_PRODUCTION === '1';
const env = await environment({ production });
let browser;
const results = [];
function luminance(channels) {
    const linear = channels.map(value => value / 255).map(value => value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4);
    return linear[0] * 0.2126 + linear[1] * 0.7152 + linear[2] * 0.0722;
}
try {
    browser = await chromium.launch();
    await mkdir(artifacts, { recursive: true });
    for (const variant of ['desktop', 'mobile']) for (const dark of [false, true]) {
        const identity = await env.user(`a11y_${variant}_${dark}`);
        const context = await browser.newContext({ viewport: variant === 'desktop' ? { width: 720, height: 900 } : { width: 390, height: 844 }, hasTouch: variant === 'mobile', reducedMotion: 'reduce' });
        await context.addInitScript(({ identity, dark }) => {
            localStorage.setItem('ebk_user_token', identity.token);
            localStorage.setItem('ebk_user_info', JSON.stringify(identity.user));
            localStorage.setItem('ebk_app_settings', JSON.stringify({ theme: dark ? 'dark' : 'light' }));
        }, { identity, dark });
        const page = await context.newPage();
        const errors = [];
        page.on('pageerror', error => errors.push(error.message));
        await page.goto(`${env.url}/${variant}${production ? '' : '.html'}${variant === 'mobile' ? '#!' : '#'}/assistant?conversationId=`);
        const input = page.getByRole('textbox', { name: 'Ask your personal finance question' });
        await input.waitFor({ timeout: 60000 });
        await input.fill('Review synthetic ledger and accessible controls with a very long conversation title');
        const send = page.getByRole('button', { name: 'Send', exact: true });
        const colors = [];
        for (const state of ['rest', 'hover', 'focus', 'pressed']) {
            if (state === 'hover') await send.hover();
            if (state === 'focus') { await page.mouse.move(0, 0); await send.focus(); }
            if (state === 'pressed') { await send.hover(); await page.mouse.down(); }
            const value = await send.evaluate(el => {
                const style = getComputedStyle(el);
                const canvas = document.createElement('canvas');
                canvas.width = canvas.height = 1;
                const ctx = canvas.getContext('2d');
                const rgb = color => { ctx.clearRect(0, 0, 1, 1); ctx.fillStyle = color; ctx.fillRect(0, 0, 1, 1); return [...ctx.getImageData(0, 0, 1, 1).data].slice(0, 3); };
                return { foreground: style.color, background: style.backgroundColor, foregroundRGB: rgb(style.color), backgroundRGB: rgb(style.backgroundColor), disabled: el.disabled, filter: style.filter, outlineColor: style.outlineColor, colorScheme: style.colorScheme };
            });
            const a = luminance(value.foregroundRGB), b = luminance(value.backgroundRGB);
            value.ratio = (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);
            assert(!value.disabled && value.ratio >= 4.5, JSON.stringify(value));
            assert.equal(value.colorScheme, dark ? 'dark' : 'light');
            if (state === 'focus') assert.equal(value.outlineColor, dark ? 'rgb(255, 255, 255)' : 'rgb(0, 0, 0)');
            colors.push({ state, ...value });
            if (state === 'pressed') { await page.mouse.move(0, 0); await page.mouse.up(); }
        }
        await page.keyboard.press('Tab');
        await send.focus();
        await page.screenshot({ path: `${artifacts}/${variant}-${dark ? 'dark' : 'light'}-send-focus.png` });
        await send.click();
        await page.getByRole('heading', { name: 'Verified isolated ledger' }).waitFor({ timeout: 30000 });
        await input.fill('Next question');
        await page.getByLabel('More', { exact: true }).click();
        const title = page.locator('.full-title');
        assert.equal(await title.innerText(), await page.locator('.toolbar-title').innerText());
        assert(await title.isVisible());
        await page.screenshot({ path: `${artifacts}/${variant}-${dark ? 'dark' : 'light'}-title.png` });
        await page.keyboard.press('Escape');
        const history = page.getByRole('button', { name: 'Conversation history', exact: true });
        await history.click();
        await page.waitForFunction(() => document.querySelector('.cloud-history')?.matches(':modal'));
        const topLayer = await page.locator('.cloud-history').evaluate(dialog => {
            const outside = [...document.querySelectorAll('button, a[href], input')].find(el => !dialog.contains(el) && el.getClientRects().length);
            outside?.focus();
            return { outsideFound: !!outside, outsideFocusBlocked: document.activeElement !== outside, modal: dialog.matches(':modal') };
        });
        assert(topLayer.outsideFound && topLayer.outsideFocusBlocked && topLayer.modal);
        for (let i = 0; i < 18; i++) {
            await page.keyboard.press('Tab');
            assert(await page.locator('.cloud-history').evaluate(el => el.contains(document.activeElement)));
        }
        await page.screenshot({ path: `${artifacts}/${variant}-${dark ? 'dark' : 'light'}-history.png` });
        await page.locator('.cloud-history summary').first().click();
        await page.getByRole('button', { name: 'Rename', exact: true }).click();
        await page.getByRole('textbox', { name: 'Conversation title' }).focus();
        await page.keyboard.press('Escape');
        assert(await page.locator('.cloud-history').evaluate(el => el.matches(':modal')));
        assert.equal(await page.getByRole('textbox', { name: 'Conversation title' }).count(), 0);
        assert.equal(await page.evaluate(() => document.activeElement?.tagName), 'SUMMARY');
        await page.keyboard.press('Escape');
        await page.waitForFunction(() => !document.querySelector('.cloud-history'));
        assert.equal(await history.evaluate(el => el === document.activeElement), true);
        await history.click();
        await page.setViewportSize({ width: 1280, height: 900 });
        await page.waitForFunction(() => document.querySelector('.cloud-history')?.tagName === 'ASIDE');
        assert.equal(await page.locator('dialog:modal').count(), 0);
        await page.setViewportSize({ width: 390, height: 844 });
        await page.waitForFunction(() => document.querySelector('.cloud-history')?.matches(':modal'));
        await page.locator('.cloud-history').getByRole('button', { name: 'New conversation', exact: true }).click();
        await page.waitForFunction(() => !document.querySelector('.cloud-history'));
        assert.equal(errors.length, 0);
        results.push({ variant, dark, colors, topLayer, errors, titleAccessible: true, resizeAndNavigation: 'PASS' });
        await context.close();
    }
    await writeFile(`${artifacts}/review.json`, JSON.stringify(results, null, 2));
} finally {
    try { await browser?.close(); } finally { await env.close(); }
}
