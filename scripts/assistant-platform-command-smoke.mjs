// Isolated real HTTP verification of commands exposed by both platform shells.
import assert from 'node:assert/strict';
import { randomUUID } from 'node:crypto';
import { writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { environment } from './assistant-experience-e2e-environment.mjs';
const env = await environment({ production: process.env.EBK_E2E_PRODUCTION === '1', queryPageSize: 8, onProviderRequest: process.env.EBK_PLATFORM_TRACE ? metadata => console.log('provider_trace', JSON.stringify(metadata)) : undefined });
const checks = [];
try {
    const login = await env.user('platform');
    const prefix = 'v1/llm/assistant/';
    const token = login.token;
    for (let index = 0; index < 11; index++) await env.api('v1/transactions/add', { type: 3, categoryId: login.fixtureIds.categoryId, time: Math.floor(Date.now() / 1000) - 20 - index, sourceAccountId: login.fixtureIds.accountId, sourceAmount: 100, comment: 'Extra isolated pagination bill' }, token);
    async function read(path, params = {}) {
        const url = new URL(`${env.base}/api/${prefix}${path}.json`);
        Object.entries(params).forEach(([key, value]) => url.searchParams.set(key, String(value)));
        const response = await fetch(url, { headers: { Authorization: `Bearer ${token}` } });
        const data = await response.json(); assert.equal(data.success, true, JSON.stringify(data)); return data.result;
    }
    let caps = await read('capabilities');
    const command = () => ({ clientRequestId: randomUUID(), requestIssuedAt: new Date().toISOString(), expectedDeletionEpoch: caps.deletionEpoch });
    const write = (path, input) => env.api(prefix + path, { ...command(), ...input }, token);
    async function finished(runId) {
        for (let i = 0; i < 100; i++) {
            const run = await read('runs/get', { runId });
            if (['completed', 'failed', 'interrupted', 'incomplete', 'canceled'].includes(run.state)) { assert.equal(run.state, 'completed', JSON.stringify(run)); return run; }
            await new Promise(resolve => setTimeout(resolve, 100));
        }
        throw new Error('run did not finish');
    }
    const sent = await write('messages/send', { expectedContextVersion: 0, mode: 'chat', content: 'List my expenses', locale: 'en', timezone: 'UTC' });
    const first = await finished(sent.runId); assert.ok(first.content.includes('123.45')); checks.push('real send / tool query / completed reply');
    let detail = await read('conversations/get', { id: sent.conversationId });
    const retried = await write('messages/retry', { conversationId: sent.conversationId, userMessageId: sent.userMessageId, expectedContextVersion: detail.conversation.contextVersion, intent: 'regenerate' });
    const second = await finished(retried.runId); assert.notEqual(second.assistantMessageId, first.assistantMessageId); checks.push('regenerate preserves a second answer version');
    detail = await read('conversations/get', { id: sent.conversationId });
    await write('messages/select_version', { conversationId: sent.conversationId, userMessageId: sent.userMessageId, assistantMessageId: first.assistantMessageId, expectedContextVersion: detail.conversation.contextVersion });
    detail = await read('conversations/get', { id: sent.conversationId }); assert.equal(detail.messages[0].selectedAssistantMessageId, first.assistantMessageId); checks.push('select successful answer version');
    const queryBranch = await write('conversations/branch', { sourceConversationId: sent.conversationId, sourceMessageId: first.assistantMessageId, expectedContextVersion: detail.conversation.contextVersion, kind: 'continue' });
    const queryTarget = await read('conversations/get', { id: queryBranch.conversationId });
    assert.equal(queryTarget.conversation.pendingQuerySourceMessageId, first.assistantMessageId);
    const pageSend = await write('messages/send', { conversationId: queryBranch.conversationId, expectedContextVersion: queryTarget.conversation.contextVersion, mode: 'next_page', sourceMessageId: queryTarget.conversation.pendingQuerySourceMessageId, content: 'Show the next page' });
    const nextPage = await finished(pageSend.runId); assert.ok(nextPage.references.length > 0); assert.ok(nextPage.references.every(reference => !first.references.some(original => original.id === reference.id))); checks.push('history query branch restores authoritative source and pages explicitly');
    const restartBranch = await write('conversations/branch', { sourceConversationId: sent.conversationId, sourceMessageId: first.assistantMessageId, expectedContextVersion: detail.conversation.contextVersion, kind: 'continue' });
    const restartTarget = await read('conversations/get', { id: restartBranch.conversationId });
    const restartSend = await write('messages/send', { conversationId: restartBranch.conversationId, expectedContextVersion: restartTarget.conversation.contextVersion, mode: 'next_page', sourceMessageId: first.assistantMessageId, restartQuery: true, content: 'Query these bills again' });
    await finished(restartSend.runId); checks.push('explicit restart query uses authorized saved scope');
    const branch = await write('conversations/branch', { sourceConversationId: sent.conversationId, sourceMessageId: sent.userMessageId, expectedContextVersion: detail.conversation.contextVersion, kind: 'edit' });
    const target = await read('conversations/get', { id: branch.conversationId }); assert.equal(target.messages.length, 0); assert.equal(target.conversation.pendingContinuation, true); checks.push('branch first stage creates reviewable target without model call');
    const branchSend = await write('messages/send', { conversationId: branch.conversationId, expectedContextVersion: target.conversation.contextVersion, mode: 'chat', content: 'Edited expense question' });
    await finished(branchSend.runId); checks.push('branch second stage sends explicitly');
    let memories = await read('memories/list');
    const saved = await write('memories/save', { expectedVersion: 0, expectedMemoryUseEpoch: memories.memoryUseEpoch, text: 'Prefer concise replies', key: 'reply_style', type: 'preference', intent: 'add' });
    memories = await read('memories/list'); const memory = memories.items.find(item => item.id === saved.memoryId); assert.equal(memory.text, 'Prefer concise replies');
    await write('memories/settings', { enabled: false, expectedSettingsVersion: memories.settingsVersion });
    memories = await read('memories/list'); assert.equal(memories.enabled, false);
    await write('memories/delete', { id: memory.id, expectedVersion: memory.version, expectedMemoryUseEpoch: memories.memoryUseEpoch });
    memories = await read('memories/list'); assert.equal(memories.items.some(item => item.id === memory.id && item.status === 'active'), false); checks.push('add / disable / forget memory commands');
    detail = await read('conversations/get', { id: sent.conversationId });
    await write('conversations/update', { id: sent.conversationId, expectedMetadataVersion: detail.conversation.metadataVersion, title: 'Platform smoke', archived: true });
    detail = await read('conversations/get', { id: sent.conversationId }); assert.equal(detail.conversation.title, 'Platform smoke'); assert.equal(detail.conversation.archived, true); checks.push('rename and archive persist');
    const exported = await fetch(`${env.base}/api/${prefix}conversations/export.json`, { method: 'POST', headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }, body: JSON.stringify({ id: sent.conversationId, format: 'markdown', viewedVersionIds: [second.assistantMessageId] }) });
    assert.equal(exported.status, 200); assert.ok((await exported.text()).includes('123.45')); checks.push('current conversation Markdown export with viewed version');
    await write('conversations/delete', { id: sent.conversationId, expectedMetadataVersion: detail.conversation.metadataVersion });
    await write('conversations/clear', {}); caps = await read('capabilities'); const history = await read('conversations/list'); assert.equal(history.items.length, 0); checks.push('delete and clear commands');
    await writeFile(resolve(env.artifacts, 'platform-commands.json'), JSON.stringify({ runAt: new Date().toISOString(), passed: checks, realHTTP: true, realModel: false, backend: process.env.EBK_E2E_DOCKER_IMAGE ? `Docker image ${process.env.EBK_E2E_DOCKER_IMAGE}; see docker-runtime.json` : 'current host Go build', queryPageSize: 8 }, null, 2));
    console.log(JSON.stringify({ passed: checks, artifacts: env.artifacts }, null, 2));
} catch (error) {
    await writeFile(resolve(env.artifacts, 'platform-commands.json'), JSON.stringify({ runAt: new Date().toISOString(), passed: checks, failed: true, failure: String(error), realHTTP: true, realModel: false, backend: process.env.EBK_E2E_DOCKER_IMAGE ? `Docker image ${process.env.EBK_E2E_DOCKER_IMAGE}; see docker-runtime.json` : 'current host Go build', queryPageSize: 8 }, null, 2));
    throw error;
} finally { await env.close(); }
