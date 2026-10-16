import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import { writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { environment } from './assistant-experience-e2e-environment.mjs';
const require = createRequire(import.meta.resolve('@vitest/browser-playwright'));
const { chromium } = require('playwright');
assert.ok(process.env.EBK_E2E_BINARY, 'An explicitly selected old-main binary is required');
assert.ok(!process.env.EBK_E2E_DOCKER_IMAGE, 'Old-backend compatibility cannot use the current Docker image');
const env = await environment({ production: true });
let browser;
const results = [];
const diagnostics = [];
try {
 browser = await chromium.launch({ headless: true });
 for (const variant of ['desktop', 'mobile']) {
  const identity = await env.user('compat_' + variant);
  const probe = await fetch(`${env.base}/api/v1/llm/assistant/capabilities.json`, { headers: { Authorization: `Bearer ${identity.token}` } });
  diagnostics.push({ variant, capabilityHTTPStatus: probe.status, contentType: probe.headers.get('content-type') });
  assert.equal(probe.status, 404, 'The selected old backend must lack saved-assistant capability API');
  const context = await browser.newContext({ viewport: variant === 'desktop' ? {width:1440,height:960} : {width:390,height:844} });
  await context.addInitScript(login => {localStorage.setItem('ebk_user_token', login.token);localStorage.setItem('ebk_user_info',JSON.stringify(login.user));}, identity);
  const page = await context.newPage(); const responses=[]; let navigations=0;
  diagnostics.push({variant, responses});
  page.on('requestfailed', request => diagnostics.push({ path: new URL(request.url()).pathname, failure: request.failure()?.errorText }));
  page.on('pageerror', error => diagnostics.push({ pageError: String(error) }));
  page.on('response',r=>{if(r.url().includes('/llm/assistant/'))responses.push({path:new URL(r.url()).pathname,status:r.status()});});
  page.on('framenavigated', frame=>{if(frame===page.mainFrame())navigations++;});
  await page.goto(`${env.url}/${variant}${variant==='mobile'?'#!/':'#/'}assistant?conversationId=`,{waitUntil:'domcontentloaded'});
  await page.getByText('Upgrade the server to use saved assistant conversations.',{exact:true}).waitFor({timeout:30000});
  const settledNavigations = navigations;
  await page.waitForTimeout(5000);
  assert.equal(env.calls(),0); assert.equal(navigations,settledNavigations); assert.ok(responses.some(r=>r.path.includes('capabilities')&&r.status===404)); assert.ok(responses.length<=2);
  await page.screenshot({path:resolve(env.artifacts,`${variant}-old-backend.png`)});
  results.push({variant,status:'PASS',responses,navigations,providerCalls:env.calls(),explanation:'Explicit server upgrade message; no refresh loop or model request'});
  await context.close();
 }
 await writeFile(resolve(env.artifacts,'compat-results.json'),JSON.stringify({executedAt:new Date().toISOString(),frontend:'final production dist',backend:'old main baseline 1a21bc9bfd2b155a3ed9b8ac57caff7c76bbb293; explicit compatibility-only override',results},null,2));
 console.log(JSON.stringify(results));
} catch (error) {
 for (const context of browser?.contexts() ?? []) for (const page of context.pages()) {
  await page.screenshot({ path: resolve(env.artifacts, 'compat-failure.png') }).catch(() => {});
  await writeFile(resolve(env.artifacts, 'compat-failure.json'), JSON.stringify({ error: String(error), diagnostics, url: page.url(), body: await page.locator('body').innerText() }, null, 2));
 }
 throw error;
} finally {try { await browser?.close(); } finally { await env.close(); }}
