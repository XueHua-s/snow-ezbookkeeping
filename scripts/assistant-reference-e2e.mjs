import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import { writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { environment } from './assistant-experience-e2e-environment.mjs';
const require = createRequire(import.meta.resolve('@vitest/browser-playwright'));
const { chromium } = require('playwright');
const production = process.env.EBK_E2E_PRODUCTION === '1';
const env = await environment({ production });
let browser;
const results = [];
const messages = {
 changed: 'This bill changed after the original query. The saved answer has not been updated.',
 missing: 'This bill is no longer accessible. It may have been deleted or access changed.',
 network: 'Unable to check the current bill. Retry when the connection is available.'
};
try {
 browser = await chromium.launch({ headless: true });
 for (const variant of ['desktop','mobile']) {
  const identity = await env.user('reference_' + variant); const foreign = await env.user('foreign_' + variant);
  const context = await browser.newContext({ viewport:variant==='desktop'?{width:1440,height:960}:{width:390,height:844}, serviceWorkers: 'block' });
  await context.addInitScript(login=>{localStorage.setItem('ebk_user_token',login.token);localStorage.setItem('ebk_user_info',JSON.stringify(login.user));},identity);
  const page=await context.newPage(); const runtimeErrors=[];page.on('pageerror',e=>runtimeErrors.push(e.message));
  await page.goto(`${env.url}/${variant}${production?'':'.html'}${variant==='mobile'?'#!/':'#/'}assistant?conversationId=`,{waitUntil:'domcontentloaded',timeout:120000});
  const input=page.getByPlaceholder('Ask your personal finance question');await input.waitFor({timeout:60000});
  await input.fill('Show my isolated expense');await page.getByRole('button',{name:'Send',exact:true}).click();
  await page.getByRole('heading',{name:'Verified isolated ledger'}).waitFor({timeout:30000});
  await input.fill('Preserve this unsent reference-check draft');
  const reference=page.getByRole('button').filter({hasText:'View Details'}).first();
  await reference.click({ trial: true });
  const location=page.url();let calls=env.calls();
  async function check(kind, keepOpen = false) {
   await reference.click({ trial: true });
   await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));
   await reference.click({ trial: true });
   await page.waitForTimeout(100);
   const reading=await page.locator('.experience-messages').evaluate(el=>({offset:el.scrollTop,following:el.scrollHeight-el.clientHeight-el.scrollTop<=80}));
   await reference.click(); await page.getByText(messages[kind],{exact:true}).waitFor();
   assert.equal(await input.inputValue(),'Preserve this unsent reference-check draft');assert.equal(page.url(),location);
   await page.waitForFunction(expected=>{const el=document.querySelector('.experience-messages');return expected.following?el.scrollHeight-el.clientHeight-el.scrollTop<=80:Math.abs(el.scrollTop-expected.offset)<4;},reading,{timeout:5000}).catch(async error=>{console.error(JSON.stringify({variant,kind,expected:reading,actual:await page.locator('.experience-messages').evaluate(el=>({offset:el.scrollTop,gap:el.scrollHeight-el.clientHeight-el.scrollTop}))}));throw error;});
   assert.equal(env.calls(),calls);assert.equal(await page.getByRole('heading',{name:'Verified isolated ledger',includeHidden:true}).count(),1);
   await page.screenshot({path:resolve(env.artifacts,`${variant}-reference-${kind}.png`)});
   if(!keepOpen) await page.getByRole('dialog').getByRole('button',{name:'Close',exact:true}).click();
  }
  const original=await env.api(`v1/transactions/get?id=${identity.fixtureIds.transactionId}`,undefined,identity.token);
  await env.api('v1/transactions/modify',{...original,sourceAmount:23456,comment:'Changed after saved response',pictureIds:[]},identity.token);
  const modified=await env.api(`v1/transactions/get?id=${identity.fixtureIds.transactionId}`,undefined,identity.token);
  assert.ok(Number.isInteger(modified.lastModifiedUnixTime)&&modified.lastModifiedUnixTime>1000000000&&modified.lastModifiedUnixTime<10000000000);
  await check('changed',true);
  if(variant==='desktop') {await page.waitForTimeout(16000);await page.getByText(messages.changed,{exact:true}).waitFor();}
  await page.getByRole('button',{name:'View current bill',exact:true}).waitFor();
  await page.getByRole('button',{name:'Ask this question again',exact:true}).waitFor();
  await page.getByRole('button',{name:'View current bill',exact:true}).click();
  if(variant==='mobile'){await page.waitForURL(/transaction\/detail/);await page.waitForFunction(()=>Array.from(document.querySelectorAll('textarea,input')).some(el=>el.value==='Changed after saved response')||document.body.innerText.includes('Changed after saved response'));await page.screenshot({path:resolve(env.artifacts,`${variant}-reference-current-bill.png`)});await page.goBack();}else{await page.getByRole('dialog').waitFor();await page.waitForFunction(()=>Array.from(document.querySelectorAll('textarea,input')).some(el=>el.value==='Changed after saved response')||document.body.innerText.includes('Changed after saved response'));await page.screenshot({path:resolve(env.artifacts,`${variant}-reference-current-bill.png`)});await page.keyboard.press('Escape');}
  await input.waitFor();assert.equal(await input.inputValue(),'Preserve this unsent reference-check draft');
  assert.equal(page.url(),location);
  results.push({variant,scenario:'real ledger modification is explicitly marked; saved answer/draft/reading offset retained; explicit current-bill and ask-again controls',status:'PASS'});
  const detail='**/api/v1/transactions/get.json?*';
  await page.route(detail,route=>route.abort('failed'));await check('network');await page.unroute(detail);
  results.push({variant,scenario:'transport failure shows distinct retry explanation and preserves answer/draft/reading offset',status:'PASS'});
  await page.route(detail,route=>{const url=new URL(route.request().url());url.searchParams.set('id',foreign.fixtureIds.transactionId);return route.continue({url:url.toString()});});
  await check('missing');await page.unroute(detail);
  results.push({variant,scenario:'preflight ID redirected to isolated foreign owner bill; actual HTTP authorization rejection uses neutral inaccessible explanation and preserves draft/offset',status:'PASS'});
  await check('changed',true);
  const conversationId=new URLSearchParams(page.url().split('?')[1]).get('conversationId');
  const detailPath=`v1/llm/assistant/conversations/get?id=${conversationId}`;
  const before=await env.api(detailPath,undefined,identity.token);const oldVersion=before.messages[0].selectedAnswer;
  await page.getByRole('button',{name:'Ask this question again',exact:true}).click();
  let after;
  for(let attempt=0;attempt<100;attempt++){after=await env.api(detailPath,undefined,identity.token);if(after.messages[0].versions.length===2 && after.messages[0].selectedAnswer.attemptNumber===2 && after.messages[0].selectedAnswer.state==='completed')break;await page.waitForTimeout(200);}
  assert.equal(after.messages.length,1);assert.equal(after.messages[0].versions.length,2);
  const historical=await env.api(`v1/llm/assistant/messages/get?assistantMessageId=${oldVersion.assistantMessageId}`,undefined,identity.token);
  assert.deepEqual(historical,oldVersion);
  assert.equal(after.messages[0].selectedAnswer.references[0].sourceAmountMinor,'23456');
  assert.equal(env.calls()-calls,2);calls=env.calls();
  assert.equal(await input.inputValue(),'Preserve this unsent reference-check draft');
  results.push({variant,scenario:'explicit ask-again creates second answer version using current ledger; original answer and queriedAt unchanged; unsent draft retained',status:'PASS'});
  await env.api('v1/transactions/delete',{id:identity.fixtureIds.transactionId},identity.token);await check('missing');
  results.push({variant,scenario:'real deleted bill uses same neutral inaccessible explanation and preserves saved answer/draft/reading offset',status:'PASS'});
  assert.deepEqual(runtimeErrors,[]);await context.close();
 }
 await writeFile(resolve(env.artifacts,'reference-results.json'),JSON.stringify({executedAt:new Date().toISOString(),frontend:production?'production dist':'actual Vite routes',backend:process.env.EBK_E2E_DOCKER_IMAGE ? `Docker image ${process.env.EBK_E2E_DOCKER_IMAGE}; isolated SQLite; see docker-runtime.json` : 'freshly built current Go and isolated SQLite',provider:'deterministic protocol fixture',status:'PASS',results,providerCalls:env.calls()},null,2));
 console.log(JSON.stringify(results));
} catch(error) {for(const context of (browser?.contexts() ?? []))for(const page of context.pages())await page.screenshot({path:resolve(env.artifacts,'reference-failure.png')}).catch(()=>{});throw error;}
finally {try { await browser?.close(); } finally { await env.close(); }}
