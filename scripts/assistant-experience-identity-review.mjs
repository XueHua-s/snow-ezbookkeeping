import { createRequire } from 'node:module';
import { writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { environment } from './assistant-experience-e2e-environment.mjs';
import { identityFlows } from './assistant-experience-e2e-identity.mjs';
const require = createRequire(import.meta.resolve('@vitest/browser-playwright'));
const { chromium } = require('playwright');
const production = process.env.EBK_E2E_PRODUCTION === '1';
const env = await environment({ production });
let browser;
const results = [];
try {
    browser = await chromium.launch({ headless: true });
    for (const variant of (process.env.EBK_E2E_VARIANT ? [process.env.EBK_E2E_VARIANT] : ['desktop', 'mobile'])) await identityFlows({ env, browser, variant, production, results });
    console.log(JSON.stringify(results, null, 2));
    await writeFile(resolve(env.artifacts, 'identity-results.json'), JSON.stringify({ platform: process.env.EBK_E2E_DOCKER_IMAGE ? `Docker image ${process.env.EBK_E2E_DOCKER_IMAGE}` : 'host', frontend: production ? 'dist' : 'Vite', results }, null, 2));
} catch (error) {
    await writeFile(resolve(env.artifacts, 'identity-results.json'), JSON.stringify({ status: 'FAIL', results, error: String(error) }, null, 2));
    throw error;
} finally { try { await browser?.close(); } finally { await env.close(); } }
