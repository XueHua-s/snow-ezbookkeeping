import assert from 'node:assert/strict';
import { randomUUID } from 'node:crypto';
import { writeFile, readFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { boundedCommand } from './assistant-bounded-command.mjs';
import { dockerCommand } from './assistant-experience-docker-runtime.mjs';
import { environment } from './assistant-experience-e2e-environment.mjs';

assert.ok(process.env.EBK_E2E_DOCKER_IMAGE, 'Select the current validation image with EBK_E2E_DOCKER_IMAGE');
const responseClosures = [];
const env = await environment({ production: true, onProviderResponseClose: event => responseClosures.push(event) });
const results = [];
const note = (scenario, details = {}) => results.push({ scenario, status: 'PASS', ...details });
const prefix = 'v1/llm/assistant/';
try {
    assert.match(await env.runtime.identity(), /uid=1000.*gid=1000/);
    note('Actual project entrypoint serves fresh database as UID:GID 1000:1000');
    const anonymous = await fetch(`${env.base}/api/${prefix}capabilities.json`);
    assert.equal(anonymous.status, 400);
    const rejection = await anonymous.json();
    assert.equal(rejection.success, false); assert.equal(rejection.errorCode, 202012);
    note('Anonymous assistant access rejected', { httpStatus: anonymous.status });
    const identity = await env.user('docker_smoke');
    note('Synthetic registration, authorization, account/category/transaction creation');
    const read = path => env.api(prefix + path, undefined, identity.token);
    async function command(path, input) {
        const caps = await read('capabilities');
        return env.api(prefix + path, { clientRequestId: randomUUID(), requestIssuedAt: new Date().toISOString(),
            expectedDeletionEpoch: caps.deletionEpoch, ...input }, identity.token);
    }
    const send = content => command('messages/send', { expectedContextVersion: 0, mode: 'chat', content, locale: 'en', timezone: 'UTC' });
    async function finished(runId, expected) {
        for (let i = 0; i < 200; i++) {
            const run = await read(`runs/get?runId=${runId}`);
            if (['completed', 'failed', 'canceled', 'interrupted', 'incomplete'].includes(run.state)) {
                assert.equal(run.state, expected); return run;
            }
            await new Promise(resolve => setTimeout(resolve, 100));
        }
        throw new Error(`Run never reached ${expected}`);
    }
    const sent = await send('Read the isolated expense');
    const completed = await finished(sent.runId, 'completed');
    assert.match(completed.content, /123.45/);
    assert.ok(completed.references.some(reference => reference.id === identity.fixtureIds.transactionId));
    assert.ok(completed.queries?.length > 0, 'Completed query must retain real tool evidence');
    const query = completed.queries.find(query => query.tool === 'query_transactions');
    assert.ok(query); assert.equal(query.scope.timezone, 'UTC');
    assert.deepEqual(query.scope.types, ['expense']); assert.equal(query.scope.excludesDeleted, true);
    assert.ok(query.coverage); assert.ok(query.queriedAt);
    assert.equal(completed.references[0].sourceAmountMinor, '12345');
    const original = await read(`conversations/get?id=${sent.conversationId}`);
    assert.equal(original.messages[0].selectedAnswer.state, 'completed');
    note('Persistent send executes real query and reads back answer/reference/evidence', {
        referenceCount: completed.references.length, queryCount: completed.queries.length
    });
    const failing = await send('E2E_FAIL Docker explicit provider failure');
    await finished(failing.runId, 'failed');
    note('Provider 503 yields persistent failed terminal state');
    const callsBefore = env.calls();
    const slow = await send('E2E_SLOW Docker cancel');
    for (let i = 0; i < 100 && env.calls() === callsBefore; i++) await new Promise(resolve => setTimeout(resolve, 50));
    assert.ok(env.calls() > callsBefore, 'Cancel only after provider actually receives the slow request');
    const slowRequestNumber = env.calls();
    const cancelStarted = performance.now();
    await command('runs/cancel', { runId: slow.runId });
    await finished(slow.runId, 'canceled');
    for (let i = 0; i < 60 && !responseClosures.some(event => event.requestNumber === slowRequestNumber); i++) await new Promise(resolve => setTimeout(resolve, 50));
    const closure = responseClosures.find(event => event.requestNumber === slowRequestNumber);
    assert.ok(closure, 'Canceled upstream response must close before the five-second provider delay');
    assert.equal(closure.completed, false, 'Cancellation must close an unfinished response, not a normal response end');
    note('Cancel after provider request closes unfinished upstream response and persists canceled', {
        upstreamResponseCompleted: closure.completed, closeObservedWithinMs: Math.round(performance.now() - cancelStarted)
    });
    await env.setGeneration(false);
    assert.equal((await read('capabilities')).generationAvailable, false);
    const disabled = await read(`conversations/get?id=${sent.conversationId}`);
    assert.equal(disabled.messages[0].selectedAnswer.content, completed.content);
    const disabledCalls = env.calls();
    await assert.rejects(send('Disabled generation must reject'), /provider_unavailable/);
    assert.equal(env.calls(), disabledCalls);
    note('Same-container restart disables generation; authenticated history remains readable and sends rejected');
    await env.setGeneration(true);
    assert.equal((await read('capabilities')).generationAvailable, true);
    const restored = await read(`conversations/get?id=${sent.conversationId}`);
    assert.equal(restored.messages[0].selectedAnswer.content, completed.content);
    const reopened = await send('Generation reopened after container restart');
    await finished(reopened.runId, 'completed');
    const transaction = await env.api(`v1/transactions/get?id=${identity.fixtureIds.transactionId}`, undefined, identity.token);
    assert.equal(transaction.sourceAmount, 12345);
    note('Same token/history/ledger persist across two real container restarts; generation reopens');
    for (const variant of ['desktop', 'mobile']) assert.equal((await fetch(`${env.base}/${variant}`)).status, 200);
    await boundedCommand('npm', ['run', 'test:production'], { timeout: 180000, capture: false, env: { ...process.env, EBK_SMOKE_BASE_URL: env.base } });
    note('npm run test:production: desktop/mobile production login renders without network/runtime errors');
    await env.runtime.stop();
    const runtime = JSON.parse(await readFile(resolve(env.artifacts, 'docker-runtime.json')));
    const database = resolve(env.directory, 'inspection.db');
    await dockerCommand(['cp', `${runtime.name}:/ezbookkeeping/data/fixture.db`, database]);
    const tables = (await boundedCommand('sqlite3', [database, "SELECT name FROM sqlite_master WHERE type='table' AND (name LIKE 'assistant_%' OR name LIKE '%embedding%') ORDER BY name"])).stdout.trim().split('\n');
    for (const table of ['assistant_conversation', 'assistant_message', 'assistant_memory', 'assistant_user_state', 'assistant_command_receipt', 'a_i_assistant_embedding']) assert.ok(tables.includes(table), table);
    note('Stopped-container SQLite inspection confirms five additive tables and retained legacy embedding table', { tables });
    await writeFile(resolve(env.artifacts, 'smoke-results.json'), JSON.stringify({
        status: 'PASS', at: new Date().toISOString(), results, providerCalls: env.calls(),
        provider: 'Deterministic host-loopback protocol fixture; not model quality evidence'
    }, null, 2));
    console.log(JSON.stringify(results, null, 2));
} catch (error) {
    await writeFile(resolve(env.artifacts, 'smoke-results.json'), JSON.stringify({ status: 'FAIL', results, error: String(error) }, null, 2));
    throw error;
} finally { await env.close(); }
