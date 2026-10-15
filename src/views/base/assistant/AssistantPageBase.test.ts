import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { PendingAssistantRequest } from './types/testing.ts';
import { effectScope, reactive, type EffectScope } from 'vue';
import type { AIAssistantChatStreamChunk } from '@/models/large_language_model.ts';

vi.mock('@/lib/services.ts', () => ({ default: { chatWithAIAssistantStream: vi.fn(), cancelRequest: vi.fn() } }));
vi.mock('@/lib/server_settings.ts', () => ({ isAIAssistantEnabled: () => true }));
vi.mock('@/lib/misc.ts', () => ({ generateRandomUUID: () => crypto.randomUUID() }));
vi.mock('@/stores/user.ts', () => ({ useUserStore: vi.fn() }));
import services from '@/lib/services.ts';
import { useUserStore } from '@/stores/user.ts';
import { buildAssistantHistory, useAssistantPageBase } from './AssistantPageBase.ts';
import type { AIAssistantConversationMessage } from './types.ts';

let requests: PendingAssistantRequest[];
let scope: EffectScope;
let page: ReturnType<typeof useAssistantPageBase>;
let user: { currentUserBasicInfo: { username: string } | null };
const candidate = { tool: 'query_transactions', cursor: 'opaque' } as const;

function finish(index: number, chunk: Partial<AIAssistantChatStreamChunk> = {}): void {
    requests[index]!.callbacks?.onDone?.({ type: 'done', reply: 'answer', ...chunk });
    requests[index]!.resolve();
}
function send(text: string): Promise<void> {
    page.messageInput.value = text;
    return page.sendMessage();
}

beforeEach(() => {
    requests = [];
    user = reactive({ currentUserBasicInfo: { username: 'alice' } });
    vi.mocked(useUserStore).mockReturnValue(user as ReturnType<typeof useUserStore>);
    vi.mocked(services.chatWithAIAssistantStream).mockImplementation(args => new Promise<void>((resolve, reject) => requests.push({ ...args, resolve, reject })));
    scope = effectScope();
    page = scope.run(() => useAssistantPageBase())!;
});
afterEach(() => scope.stop());

describe('assistant page session', () => {
    it('exposes progress and evidence, preserves current question and only sends continuation on explicit actions', async () => {
        const promise = send('  原文😀  ');
        expect(requests[0]!.req.message).toBe('  原文😀  ');
        expect(page.state.value).toBe('requesting');
        requests[0]!.callbacks?.onStatus?.('aggregating');
        requests[0]!.callbacks?.onReplyDelta?.('partial');
        expect(page.progress.value).toBe('aggregating');
        finish(0, { continuation: candidate });
        await promise;
        expect(page.state.value).toBe('completed');
        expect(page.progress.value).toBeUndefined();
        expect(page.canRequestNextPage.value).toBe(true);
        expect(requests).toHaveLength(1);
        const next = page.requestNextPage('下一页');
        expect(requests[1]!.req).toMatchObject({ message: '下一页', continuation: candidate });
        expect(requests[1]!.req.history).toEqual([{ role: 'user', content: '  原文😀  ' }, { role: 'assistant', content: 'answer' }]);
        finish(1);
        await next;
        expect(page.continuation.value).toBeUndefined();
    });

    it('keeps a completed summary exchange in followup history without changing summary request defaults', async () => {
        const first = send('只看支付宝');
        finish(0, { reply: '已选择支付宝' });
        await first;
        page.messageInput.value = '尚未发送的草稿';
        const summary = page.generateSummary('生成 AI 摘要');
        expect(requests[1]!.req).toEqual({ mode: 'summary', continuation: undefined, history: [
            { role: 'user', content: '只看支付宝' }, { role: 'assistant', content: '已选择支付宝' }
        ] });
        expect(page.messageInput.value).toBe('尚未发送的草稿');
        expect(page.messages.value.at(-2)).toMatchObject({ role: 'user', content: '生成 AI 摘要' });
        finish(1, { reply: '支付宝本月支出为 100 元' });
        await summary;
        const followup = send('那上个月呢');
        expect(requests[2]!.req.history).toEqual([
            { role: 'user', content: '只看支付宝' }, { role: 'assistant', content: '已选择支付宝' },
            { role: 'user', content: '生成 AI 摘要' }, { role: 'assistant', content: '支付宝本月支出为 100 元' }
        ]);
        finish(2);
        await followup;
    });

    it.each(['canceled', 'incomplete'] as const)('excludes an entire %s summary exchange from followup history', async state => {
        const summary = page.generateSummary('生成 AI 摘要');
        requests[0]!.callbacks?.onReplyDelta?.('未完成的摘要');
        if (state === 'canceled') {
            page.cancelCurrentRequest();
            requests[0]!.reject({ canceled: true });
            await summary;
        } else {
            requests[0]!.reject({ incomplete: true });
            await expect(summary).rejects.toMatchObject({ incomplete: true });
        }
        const followup = send('那上个月呢');
        expect(requests[1]!.req.history).toEqual([]);
        finish(1);
        await followup;
    });

    it('prevents consumers from changing the session or nested evidence outside its actions', async () => {
        const first = send('first');
        finish(0, { continuation: candidate, references: [{ id: '17', time: 1, type: 3, sourceAmount: 100 }] });
        await first;
        const warning = vi.spyOn(console, 'warn').mockImplementation(() => undefined);
        try {
            Reflect.set(page.messages.value[1]!.references![0]!, 'sourceAmount', 999);
            Reflect.set(page.messages.value, '0', { role: 'assistant', content: 'forged' });
            Reflect.set(page.continuation.value!, 'cursor', 'forged');
            Reflect.set(page.messages, 'value', []);
            expect(page.messages.value[0]!.content).toBe('first');
            expect(page.messages.value[1]!.references![0]!.sourceAmount).toBe(100);
            expect(page.continuation.value).toEqual(candidate);
        } finally {
            warning.mockRestore();
        }
        page.clearConversation();
        expect(page.messages.value).toEqual([]);
    });

    it.each(['completed', 'canceled', 'failed'] as const)('preserves an unsent draft when pagination is %s', async outcome => {
        const first = send('first');
        finish(0, { continuation: candidate });
        await first;
        page.messageInput.value = '尚未发送的草稿';
        const next = page.requestNextPage('下一页');
        expect(requests[1]!.req).toMatchObject({ message: '下一页', continuation: candidate });
        expect(page.messageInput.value).toBe('尚未发送的草稿');
        if (outcome === 'completed') {
            finish(1);
            await next;
        } else if (outcome === 'canceled') {
            page.cancelCurrentRequest();
            requests[1]!.reject({ canceled: true });
            await next;
        } else {
            requests[1]!.reject({ message: 'failed' });
            await expect(next).rejects.toMatchObject({ message: 'failed' });
        }
        expect(page.messageInput.value).toBe('尚未发送的草稿');
    });

    it('canceled callbacks and finally cannot overwrite a new request', async () => {
        const old = send('first');
        requests[0]!.callbacks?.onReplyDelta?.('partial');
        page.cancelCurrentRequest();
        expect(page.state.value).toBe('canceled');
        expect(services.cancelRequest).toHaveBeenCalledWith(requests[0]!.cancelableUuid);
        const current = send('second');
        expect(requests[1]!.req.history).toEqual([]);
        requests[0]!.callbacks?.onReplyDelta?.('stale');
        finish(0, { continuation: candidate });
        await old;
        expect(page.requesting.value).toBe(true);
        expect(page.messages.value[1]!.content).toBe('partial');
        expect(page.messages.value[3]!.content).toBe('');
        expect(page.continuation.value).toBeUndefined();
        finish(1);
        await current;
        expect(page.state.value).toBe('completed');
    });

    it.each(['clear', 'user', 'leave'] as const)('cleans context and aborts on %s, ignoring even late failures', async action => {
        const first = send('first');
        finish(0, { continuation: candidate });
        await first;
        const pending = send('second');
        page.messageInput.value = 'draft';
        if (action === 'clear') page.clearConversation();
        if (action === 'user') user.currentUserBasicInfo = { username: 'bob' };
        if (action === 'leave') scope.stop();
        expect(page.messages.value).toEqual([]);
        expect(page.messageInput.value).toBe('');
        expect(page.continuation.value).toBeUndefined();
        expect(page.requesting.value).toBe(false);
        expect(services.cancelRequest).toHaveBeenCalledWith(requests[1]!.cancelableUuid);
        requests[1]!.reject({ response: { data: { errorMessage: 'old' } } });
        await pending;
        expect(page.messages.value).toEqual([]);
    });

    it('preserves server error context, query scopes and continuation without adding partial answers to history', async () => {
        const error = { response: { data: { success: false, errorCode: 42, errorMessage: 'budget', path: '/assistant', context: {
            assistant: { state: 'incomplete', queries: [{ queryId: 'q1', scope: { includesHidden: true } }], continuation: candidate }
        } } } };
        const promise = send('first');
        requests[0]!.callbacks?.onReplyDelta?.('unverified');
        requests[0]!.reject(error);
        await expect(promise).rejects.toMatchObject(error);
        expect(page.messages.value[1]).toMatchObject({ state: 'incomplete', queries: error.response.data.context.assistant.queries, error });
        expect(page.continuation.value).toEqual(candidate);
        const next = send('new unrelated question');
        expect(requests[1]!.req).toMatchObject({ history: [], continuation: candidate, message: 'new unrelated question' });
        finish(1);
        await next;
    });

    it('exposes an explicit failed server terminal without relabeling it incomplete', async () => {
        const promise = send('first');
        const error = { response: { data: { success: false, errorMessage: 'unsupported', context: { assistant: { state: 'failed', queries: [] } } } } };
        requests[0]!.reject(error);
        await expect(promise).rejects.toMatchObject(error);
        expect(page.state.value).toBe('failed');
        expect(page.continuation.value).toBeUndefined();
    });

    it.each([{ timedOut: true }, { incomplete: true }, {}])('never completes transport success without a terminal, or transport failures %j', async error => {
        const pending = send('question');
        requests[0]!.reject(error);
        await expect(pending).rejects.toMatchObject({ message: 'Unable to get AI assistant response' });
        expect(page.state.value).toBe('incomplete');
        expect(page.requesting.value).toBe(false);
        const noTerminal = send('again');
        requests[1]!.resolve();
        await expect(noTerminal).rejects.toMatchObject({ incomplete: true });
        expect(page.state.value).toBe('incomplete');
    });
});

describe('assistant history', () => {
    it('keeps the latest six whole exchanges and truncates historical codepoints with an ellipsis', () => {
        const messages: AIAssistantConversationMessage[] = [];
        for (let i = 0; i < 8; i++) {
            messages.push({ id: `u${i}`, role: 'user', content: i === 7 ? '😀'.repeat(2050) : `question${i}`, createdAt: 0, state: 'completed' });
            messages.push({ id: `a${i}`, role: 'assistant', content: `answer${i}`, createdAt: 0, state: 'completed' });
        }
        messages.push({ id: 'partial', role: 'assistant', content: 'unverified', createdAt: 0, state: 'incomplete' });
        const history = buildAssistantHistory(messages);
        expect(history).toHaveLength(12);
        expect(history[0]!.content).toBe('question2');
        expect(history[10]!.content).toBe('😀'.repeat(2047) + '…');
        expect(Array.from(history[10]!.content)).toHaveLength(2048);
        expect(messages[14]!.content).toBe('😀'.repeat(2050));
        expect(history.at(-1)!.content).toBe('answer7');
    });
});
