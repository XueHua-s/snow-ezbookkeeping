import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import { writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { environment } from './assistant-experience-e2e-environment.mjs';
const require = createRequire(import.meta.resolve('@vitest/browser-playwright'));
const { chromium } = require('playwright');
assert.ok(process.env.EBK_E2E_BINARY, 'An explicitly selected old-main binary is required');
const env = await environment({ production: true });
const browser = await chromium.launch({ headless: true });
const results = [];
try {
 for (const variant of ['desktop', 'mobile']) {
  const identity = await env.user('compat_' + variant);
  const context = await browser.newContext({ viewport: variant === 'desktop' ? {width:1440,height:960} : {width:390,height:844} });
  await context.addInitScript(login => {localStorage.setItem('ebk_user_token', login.token);localStorage.setItem('ebk_user_info',JSON.stringify(login.user));}, identity);
  const page = await context.newPage(); const responses=[]; let navigations=0;
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
} finally {await browser.close();await env.close();}
