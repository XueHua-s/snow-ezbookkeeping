import { computed, ref, shallowRef, watch } from 'vue';
import { assistantConversationServices as api } from '@/lib/services/assistant_conversations.ts';
import { isAssistantRunTerminal as terminal } from '@/models/ai_assistant_conversation.ts';
import type * as DTO from '@/models/ai_assistant_conversation.ts';
import { result, statusCode, type ReadingPosition } from './types.ts';
import type { AssistantSession } from './session.ts';

function mergeVersions(previous: readonly DTO.AssistantVersionDTO[], incoming: readonly DTO.AssistantVersionDTO[]) {
    const versions = new Map(previous.map(version => [version.id, version]));
    for (const version of incoming) {
        const known = versions.get(version.id);
        // A page captured before streaming completed must not regress a known final state.
        if (!known || !terminal(known.state)) versions.set(version.id, version);
    }
    return [...versions.values()].sort((a, b) => b.attemptNumber - a.attemptNumber);
}

/** Owns the conversation window, read intent, and attached stream/poll resources. */
export function useAssistantConversation(session: AssistantSession, synchronize: () => Promise<void>) {
    const { conversationId, recentConversationId, error, capabilities, restoreDraft, applyLocal, fail } = session;
    const detail = shallowRef<DTO.AssistantConversationDetail>();
    const memories = shallowRef<DTO.AssistantMemoryList>();
    const viewedVersions = shallowRef<Record<string, DTO.AssistantRunSnapshot>>({});
    const loading = ref(false),
        updatesAvailable = ref(false),
        windowMissingLatest = ref(false);
    const online = ref(typeof navigator === 'undefined' || navigator.onLine);
    const busy = computed(() => !!detail.value?.activeRun && !terminal(detail.value.activeRun.state));
    const positions = new Map<string, ReadingPosition>();
    let navigation = 0;
    let olderWindow = false;
    let attached = 0;
    let timer: ReturnType<typeof setTimeout> | undefined;
    let subscription: AbortController | undefined;
    let subscriptionRun = '';
    let backoff = 1000,
        pollUntil = 0,
        streamFailures = 0;
    let memoryTicket = 0,
        olderTicket = 0;
    let answerTicket = 0;
    const acceptedAnswers = new Map<string, number>();
    let streamRun = '';
    const versionTickets = new Map<string, number>();
    function capture() {
        const valid = session.capture();
        const ticket = navigation;
        return () => valid() && ticket === navigation;
    }
    function stopSubscription() {
        subscription?.abort();
        subscription = undefined;
        subscriptionRun = '';
    }
    function reset() {
        navigation++;
        olderTicket++;
        memoryTicket++;
        answerTicket++;
        acceptedAnswers.clear();
        streamRun = '';
        versionTickets.clear();
        olderWindow = false;
        detail.value = undefined;
        memories.value = undefined;
        viewedVersions.value = {};
        positions.clear();
        loading.value = false;
        updatesAvailable.value = false;
        windowMissingLatest.value = false;
        backoff = 1000;
        pollUntil = 0;
        streamFailures = 0;
        stopSubscription();
        if (timer) clearTimeout(timer);
        timer = undefined;
    }
    watch(session.generation, reset, { flush: 'sync' });
    async function open(id = '', aroundMessageId?: string, versionId?: string) {
        if (session.disposed) return;
        const ticket = ++navigation;
        acceptedAnswers.clear();
        versionTickets.clear();
        const generation = session.generation.value;
        const valid = capture();
        stopSubscription();
        windowMissingLatest.value = false;
        olderWindow = false;
        updatesAvailable.value = false;
        conversationId.value = id;
        detail.value = undefined;
        viewedVersions.value = {};
        restoreDraft();
        loading.value = false;
        if (!id) return;
        loading.value = true;
        error.value = undefined;
        try {
            const next = await result(api.get({ id, aroundMessageId, limit: 30 }));
            if (ticket !== navigation || generation !== session.generation.value) return;
            detail.value = next;
            recentConversationId.value = id;
            try {
                localStorage.setItem('assistant-recent:' + capabilities.value?.ownerScope, id);
            } catch {
                /* Navigation still works without session.storage. */
            }
            if (versionId) await viewVersion(versionId);
            if (valid()) startSubscription();
        } catch (failure) {
            if (ticket !== navigation || generation !== session.generation.value) return;
            if (statusCode(failure) === 404 && session.storage) {
                const owner = session.storage;
                const saved = await owner.clearConversation(id);
                if (ticket !== navigation || generation !== session.generation.value || owner !== session.storage)
                    return;
                applyLocal(saved);
                conversationId.value = '';
                restoreDraft();
            }
            fail(failure);
        } finally {
            if (ticket === navigation) loading.value = false;
        }
    }
    async function loadOlder() {
        const current = detail.value;
        const valid = capture();
        const ticket = ++olderTicket;
        if (!current?.hasMore || session.disposed) return;
        const next = await result(
            api.get({ id: current.conversation.id, beforeTurnSequence: current.messages[0]?.turnSequence, limit: 30 }),
        ).catch((failure) => {
            if (valid() && ticket === olderTicket) throw failure;
        });
        if (!next || !valid() || ticket !== olderTicket || current !== detail.value) return;
        olderWindow = true;
        const combined = [...next.messages, ...current.messages].filter(
            (item, index, items) => items.findIndex((m) => m.id === item.id) === index,
        );
        windowMissingLatest.value ||= combined.length > 90;
        detail.value = {
            ...current,
            messages: combined.slice(0, 90),
            hasMore: next.hasMore,
            nextCursor: next.nextCursor,
        };
    }
    function applyDetail(next: DTO.AssistantConversationDetail) {
        const previous = detail.value;
        if (!previous || previous.conversation.id !== next.conversation.id) {
            detail.value = next;
            return;
        }
        if (
            next.conversation.contextVersion < previous.conversation.contextVersion ||
            next.conversation.metadataVersion < previous.conversation.metadataVersion
        )
            return;
        const retain = (fresh: DTO.AssistantRunSnapshot | undefined, old: DTO.AssistantRunSnapshot | undefined) =>
            fresh && old && fresh.runId === old.runId && (fresh.revision < old.revision || terminal(old.state))
                ? old
                : fresh;
        if (olderWindow) {
            updatesAvailable.value =
                updatesAvailable.value || next.conversation.contextVersion !== previous.conversation.contextVersion;
            next = {
                ...next,
                messages: previous.messages.map((old) => next.messages.find((fresh) => fresh.id === old.id) ?? old),
                hasMore: previous.hasMore,
                nextCursor: previous.nextCursor,
            };
        }
        detail.value = {
            ...next,
            activeRun: retain(next.activeRun, previous.activeRun),
            messages: next.messages.map((message) => {
                const old = previous.messages.find(item => item.id === message.id);
                const versions = mergeVersions(old?.versions ?? [], message.versions);
                // Attempt numbers are sequential: a complete retained history remains complete
                // even when the server's refreshed recent window advertises older answers.
                const complete = versions.length > 0 && versions.length === Math.max(...versions.map(item => item.attemptNumber));
                return {
                    ...message,
                    versions,
                    hasMoreVersions: !complete && (message.hasMoreVersions || !!old?.hasMoreVersions),
                    selectedAnswer: retain(message.selectedAnswer, old?.selectedAnswer),
                };
            }),
        };
    }
    function mergeSnapshot(snapshot: DTO.AssistantRunSnapshot) {
        const current = detail.value;
        if (!current || snapshot.conversationId !== conversationId.value) return;
        const existing =
            current.activeRun?.runId === snapshot.runId
                ? current.activeRun
                : current.messages.find((m) => m.selectedAnswer?.runId === snapshot.runId)?.selectedAnswer;
        if (existing && (existing.revision >= snapshot.revision || terminal(existing.state))) return;
        detail.value = {
            ...current,
            activeRun: snapshot,
            messages: current.messages.map((m) =>
                m.id === snapshot.userMessageId
                    ? {
                          ...m,
                          selectedAnswer: snapshot,
                          versions: [
                              ...m.versions.filter((v) => v.id !== snapshot.assistantMessageId),
                              {
                                  id: snapshot.assistantMessageId,
                                  attemptNumber: snapshot.attemptNumber,
                                  state: snapshot.state,
                                  createdAt: snapshot.savedAt,
                              },
                          ],
                      }
                    : m,
            ),
        };
        if (terminal(snapshot.state)) {
            stopSubscription();
            void synchronize();
        }
    }
    function startSubscription() {
        const run = detail.value?.activeRun;
        if (run && run.runId !== streamRun) {
            streamRun = run.runId;
            streamFailures = 0;
            backoff = 1000;
            pollUntil = 0;
        }
        if (
            session.disposed ||
            streamFailures >= 3 ||
            !attached ||
            !online.value ||
            document.hidden ||
            !run ||
            terminal(run.state) ||
            subscriptionRun === run.runId ||
            Date.now() < pollUntil
        )
            return;
        stopSubscription();
        const controller = new AbortController();
        subscription = controller;
        subscriptionRun = run.runId;
        const generation = session.generation.value;
        const ticket = navigation;
        void api
            .subscribe(
                run.runId,
                run.revision,
                {
                    onSnapshot: (snapshot) => {
                        if (
                            generation === session.generation.value &&
                            ticket === navigation &&
                            !controller.signal.aborted
                        ) {
                            backoff = 1000;
                            streamFailures = 0;
                            mergeSnapshot(snapshot);
                        }
                    },
                },
                controller.signal,
            )
            .catch((failure) => {
                if (controller.signal.aborted || generation !== session.generation.value || ticket !== navigation)
                    return;
                if (statusCode(failure) === 401) {
                    fail(failure);
                    return;
                }
                streamFailures++;
                const retry = (failure as { retryAfterMs?: number }).retryAfterMs ?? backoff;
                pollUntil = Date.now() + Math.max(1000, Math.min(retry, 60000));
                backoff = Math.min(backoff * 2, 15000);
            })
            .finally(() => {
                if (subscription === controller) {
                    subscription = undefined;
                    subscriptionRun = '';
                }
            });
    }
    function schedule() {
        if (timer) clearTimeout(timer);
        if (session.disposed || !attached || document.hidden) return;
        const valid = session.capture();
        timer = setTimeout(
            () => {
                void synchronize().finally(() => {
                    if (valid()) schedule();
                });
            },
            detail.value?.activeRun && !terminal(detail.value.activeRun.state) && !subscription
                ? Math.max(2000, backoff)
                : 14000 + Math.random() * 2000,
        );
    }
    function foreground() {
        online.value = navigator.onLine;
        if (document.hidden || !online.value) {
            stopSubscription();
            if (timer) clearTimeout(timer);
            return;
        }
        streamFailures = 0;
        const valid = session.capture();
        void synchronize().finally(() => {
            if (valid()) schedule();
        });
    }
    function attach() {
        if (session.disposed || ++attached > 1) return;
        window.addEventListener('online', foreground);
        window.addEventListener('offline', foreground);
        document.addEventListener('visibilitychange', foreground);
        foreground();
    }
    function detach() {
        attached = Math.max(0, attached - 1);
        if (attached) return;
        stopSubscription();
        if (timer) clearTimeout(timer);
        window.removeEventListener('online', foreground);
        window.removeEventListener('offline', foreground);
        document.removeEventListener('visibilitychange', foreground);
    }
    async function viewVersion(assistantMessageId: string) {
        const valid = capture();
        const ticket = ++answerTicket;
        const userMessageId = detail.value?.messages.find(
            (message) =>
                message.selectedAnswer?.assistantMessageId === assistantMessageId ||
                message.versions.some((version) => version.id === assistantMessageId),
        )?.id;
        if (userMessageId) acceptedAnswers.set(userMessageId, ticket);
        const latest = () => (userMessageId ? acceptedAnswers.get(userMessageId) === ticket : ticket === answerTicket);
        if (session.disposed) return;
        const answer = await result(api.message(assistantMessageId)).catch((failure) => {
            if (valid() && latest()) throw failure;
        });
        if (
            !answer ||
            !valid() ||
            answer.conversationId !== conversationId.value ||
            ticket < (acceptedAnswers.get(answer.userMessageId) ?? 0)
        )
            return;
        acceptedAnswers.set(answer.userMessageId, ticket);
        viewedVersions.value = Object.fromEntries([
            ...Object.entries(viewedVersions.value)
                .filter(([id]) => id !== answer.userMessageId)
                .slice(-19),
            [answer.userMessageId, answer],
        ]);
    }
    function followLatestAnswer(userMessageId: string) {
        // A confirmed regeneration supersedes even an outstanding historical read.
        acceptedAnswers.set(userMessageId, ++answerTicket);
        viewedVersions.value = Object.fromEntries(Object.entries(viewedVersions.value).filter(([id]) => id !== userMessageId));
    }
    function isLatestMessage(id: string) {
        return (
            detail.value?.messages.find((message) => message.id === id)?.turnSequence ===
            (detail.value?.conversation.nextTurnSequence ?? 0) - 1
        );
    }
    async function loadMemories() {
        const valid = session.capture();
        const ticket = ++memoryTicket;
        if (session.disposed) return;
        const next = await result(api.memories()).catch((failure) => {
            if (valid() && ticket === memoryTicket) throw failure;
        });
        if (next && valid() && ticket === memoryTicket) memories.value = next;
    }
    async function loadVersions(userMessageId: string, cursor?: string) {
        const valid = capture();
        const ticket = (versionTickets.get(userMessageId) ?? 0) + 1;
        versionTickets.set(userMessageId, ticket);
        if (session.disposed) return;
        const next = await result(api.versions({ userMessageId, cursor, limit: 20 })).catch((failure) => {
            if (valid() && ticket === versionTickets.get(userMessageId)) throw failure;
        });
        if (!next || !valid() || ticket !== versionTickets.get(userMessageId)) return undefined;
        const current = detail.value;
        if (current)
            detail.value = {
                ...current,
                messages: current.messages.map((message) =>
                    message.id === userMessageId
                        ? {
                              ...message,
                              versions: mergeVersions(message.versions, next.items),
                              hasMoreVersions: !!next.nextCursor,
                          }
                        : message,
                ),
            };
        return next;
    }
    function savePosition(position: ReadingPosition) {
        positions.delete(conversationId.value);
        positions.set(conversationId.value, position);
        if (positions.size > 20) positions.delete(positions.keys().next().value!);
    }
    async function loadLatest() {
        await open(conversationId.value);
    }
    function readingPosition() {
        return positions.get(conversationId.value);
    }
    function dispose() {
        attached = 1;
        detach();
        reset();
    }
    return {
        dispose,
        detail,
        memories,
        viewedVersions,
        loading,
        updatesAvailable,
        windowMissingLatest,
        online,
        busy,
        capture,
        open,
        loadOlder,
        applyDetail,
        mergeSnapshot,
        startSubscription,
        stopSubscription,
        schedule,
        foreground,
        attach,
        detach,
        viewVersion,
        followLatestAnswer,
        isLatestMessage,
        loadMemories,
        loadVersions,
        savePosition,
        loadLatest,
        readingPosition,
    };
}
export type AssistantConversation = ReturnType<typeof useAssistantConversation>;
