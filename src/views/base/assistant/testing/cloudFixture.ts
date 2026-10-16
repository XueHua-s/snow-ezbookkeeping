import type { Mock } from 'vitest';
import { fixtureConversation, fixtureMessage } from '@/components/common/assistant/experienceFixture.ts';
import type { AssistantConversationDetail, AssistantSendRequest } from '@/models/ai_assistant_conversation.ts';

/** Protocol fixture only: components, Pinia state and local persistence remain real. */
export function setupCloudAPI(api: Record<string, Mock>) {
    const response = (result: unknown) => ({ data: { success: true, result } });
    let detail: AssistantConversationDetail = { conversation: fixtureConversation, messages: [fixtureMessage], hasMore: false };
    let generationAvailable = true;
    let ownerScope: string = crypto.randomUUID();
    api['capabilities']!.mockImplementation(async () => response({ schemaVersion: 1, ownerScope, serverTime: new Date().toISOString(), deletionEpoch: 1,
        generationAvailable, historyAvailable: true, memoryAvailable: true, settingsVersion: 1, memoryUseEpoch: 1, memoryEnabled: true,
        limits: { inputCharacters: 4000, inputBytes: 16384, titleCharacters: 100, conversationBytes: 2097152, memoryBytes: 1024 }, usage: {} }));
    api['list']!.mockImplementation(async () => response({ items: [detail.conversation] }));
    api['get']!.mockImplementation(async () => response(detail));
    api['state']!.mockImplementation(async () => response({ id: '1', accessible: true, contextVersion: detail.conversation.contextVersion, metadataVersion: 1, activeRunRevision: 0 }));
    api['command']!.mockResolvedValue(response({ status: 'unknown' }));
    api['memories']!.mockResolvedValue(response({ items: [], enabled: true, settingsVersion: 1, memoryUseEpoch: 1 }));
    api['send']!.mockImplementation(async (req: AssistantSendRequest) => {
        detail = { ...detail, conversation: { ...detail.conversation, contextVersion: detail.conversation.contextVersion + 1 }, messages: [...detail.messages,
            { ...fixtureMessage, id: '20', content: req.content, selectedAnswer: { ...fixtureMessage.selectedAnswer!, content: 'Verified cloud answer', userMessageId: '20', assistantMessageId: '21', runId: '21' } }] };
        return response({ status: 'accepted', conversationId: '1', clientRequestId: req.clientRequestId });
    });
    api['subscribe']!.mockResolvedValue(undefined);
    return { setOwner(value: string) { ownerScope = value; }, setGeneration(value: boolean) { generationAvailable = value; }, getDetail: () => detail };
}
