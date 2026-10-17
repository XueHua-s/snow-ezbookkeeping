import { describe, expect, it, vi } from 'vitest';
import { mocks, response, conversation, snapshot, deferred, store } from './testFixture.ts';

const oldAnswer = snapshot({ state: 'completed', assistantMessageId: 'old', content: 'Old answer' });
const answer = snapshot({ state: 'completed', content: 'Current answer' });
const detail = (current = answer) => ({ ...conversation('1', { nextTurnSequence: 2 }), messages: [{
    id: '2', turnSequence: 1, selectedAnswer: current, selectedAssistantMessageId: current.assistantMessageId,
    versions: [oldAnswer, answer].map(item => ({ id: item.assistantMessageId, attemptNumber: item.attemptNumber, state: item.state })), hasMoreVersions: false
}] });
async function prepare() {
    await store.initialize();
    mocks.api['get']!.mockResolvedValueOnce(response(detail()));
    await store.open('1');
    mocks.api['message']!.mockResolvedValueOnce(response(oldAnswer));
    await store.viewVersion('old');
    mocks.api['message']!.mockResolvedValueOnce(response(snapshot({ userMessageId: 'historical', assistantMessageId: 'historical-answer' })));
    await store.viewVersion('historical-answer');
}

describe('regeneration follows its new answer', () => {
    it.each(['retry', 'regenerate'])('exits browsing only the %s question and shows its stream through completion', async intent => {
        await prepare();
        // Browsing back to the current answer must also release its frozen snapshot.
        if (intent === 'regenerate') {
            mocks.api['message']!.mockResolvedValueOnce(response(answer));
            await store.viewVersion(answer.assistantMessageId);
        }
        const running = snapshot({ runId: 'new-run', assistantMessageId: 'new-answer', attemptNumber: 3, revision: 1, content: 'Starting' });
        mocks.api['retry']!.mockResolvedValueOnce(response({ status: 'accepted', conversationId: '1' }));
        mocks.api['get']!.mockResolvedValueOnce(response({ ...detail(running), activeRun: running }));
        await store.retry('2', intent);
        expect(store.viewedVersions['2']).toBeUndefined();
        expect(store.viewedVersions['historical']?.assistantMessageId).toBe('historical-answer');
        store.mergeSnapshot({ ...running, revision: 2, content: 'Streaming new answer' });
        expect(store.detail!.messages[0]!.selectedAnswer?.content).toBe('Streaming new answer');
        store.mergeSnapshot({ ...running, revision: 3, state: 'completed', content: 'Finished new answer' });
        expect(store.detail!.messages[0]!.selectedAnswer?.content).toBe('Finished new answer');
        expect(store.viewedVersions['2']).toBeUndefined();
    });
    it('keeps the selected history when the command fails', async () => {
        await prepare();
        mocks.api['retry']!.mockRejectedValueOnce(new Error('connection lost'));
        await expect(store.retry('2')).rejects.toThrow('connection lost');
        expect(store.viewedVersions['2']?.assistantMessageId).toBe('old');
    });
    it('does not revive an outstanding history read after regeneration is accepted', async () => {
        await prepare();
        const pendingRead = deferred<unknown>();
        mocks.api['message']!.mockReturnValueOnce(pendingRead.promise);
        const reading = store.viewVersion('old');
        mocks.api['retry']!.mockResolvedValueOnce(response({ status: 'accepted', conversationId: '1' }));
        mocks.api['get']!.mockResolvedValueOnce(response(detail()));
        await store.retry('2');
        pendingRead.resolve(response(oldAnswer));
        await reading;
        expect(store.viewedVersions['2']).toBeUndefined();
    });
    it('does not clear browsing established after navigation and reopening the same conversation', async () => {
        await prepare();
        const pendingRetry = deferred<unknown>();
        mocks.api['retry']!.mockReturnValueOnce(pendingRetry.promise);
        const retry = store.retry('2');
        await vi.waitFor(() => expect(mocks.api['retry']).toHaveBeenCalledTimes(1));
        await store.open('other');
        mocks.api['get']!.mockResolvedValueOnce(response(detail()));
        await store.open('1');
        mocks.api['message']!.mockResolvedValueOnce(response(oldAnswer));
        await store.viewVersion('old');
        pendingRetry.resolve(response({ status: 'accepted', conversationId: '1' }));
        await retry;
        expect(store.viewedVersions['2']?.assistantMessageId).toBe('old');
    });
});
