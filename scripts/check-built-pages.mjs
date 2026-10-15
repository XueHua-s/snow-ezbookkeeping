import { createRequire } from 'node:module';

// Use the same installed Playwright/browser as the Vitest browser provider.
const require = createRequire(import.meta.resolve('@vitest/browser-playwright'));
const { chromium } = require('playwright');
const baseUrl = process.env.EBK_SMOKE_BASE_URL;
if (!baseUrl) throw new Error('Set EBK_SMOKE_BASE_URL to the running production-build application.');

const browser = await chromium.launch({ headless: true });
try {
    for (const variant of ['desktop', 'mobile']) {
        const context = await browser.newContext({
            ignoreHTTPSErrors: process.env.EBK_SMOKE_IGNORE_TLS_ERRORS === 'true',
            viewport: variant === 'desktop' ? { width: 1440, height: 960 } : { width: 390, height: 844 }
        });
        try {
            const page = await context.newPage();
            const errors = [];
            page.on('pageerror', error => errors.push(error.stack ?? error.message));
            page.on('requestfailed', request => errors.push(`${request.url()}: ${request.failure()?.errorText}`));
            const response = await page.goto(new URL(variant, baseUrl.endsWith('/') ? baseUrl : `${baseUrl}/`).href, {
                waitUntil: 'networkidle', timeout: 45000
            });
            if (response?.status() !== 200) throw new Error(`${variant}: document did not return 200`);
            try {
                await page.locator('input[type="password"]').first().waitFor({ state: 'visible', timeout: 15000 });
            } catch (error) {
                throw new Error(`${variant}: login page did not render; ${errors.join('; ')}`, { cause: error });
            }
            if (errors.length) throw new Error(`${variant}: ${errors.join('; ')}`);
            console.log(`${variant}: production bundle rendered login page without runtime or network errors`);
        } finally {
            await context.close();
        }
    }
} finally {
    await browser.close();
}
