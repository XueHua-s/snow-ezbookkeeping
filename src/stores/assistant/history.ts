import { shallowRef } from 'vue';
import { assistantConversationServices as api } from '@/lib/services/assistant_conversations.ts';
import type { AssistantConversationList, AssistantSearchResult } from '@/models/ai_assistant_conversation.ts';
import { result } from './types.ts';

/**
 * Owns the bounded history/search windows, their pagination and stale-request fencing.
 * The store supplies its identity generation; callers never manage request tickets.
 * Visibility survives identity resets because the same sidebar may remain mounted.
 */
export function useAssistantHistory(identity: () => number) {
    const history = shallowRef<AssistantConversationList>({ items: [] });
    const searchResults = shallowRef<AssistantSearchResult>({ items: [] });
    let historyTicket = 0;
    let searchTicket = 0;
    let includeArchived = false;
    let visible = false;
    let lastRefresh = 0;
    let historyLoading = false;
    let disposed = false;

    function reset() {
        historyTicket++;
        historyLoading = false;
        searchTicket++;
        includeArchived = false;
        lastRefresh = 0;
        history.value = { items: [] };
        searchResults.value = { items: [] };
    }

    async function loadHistory(cursor?: string, archived = includeArchived) {
        if (disposed) return;
        const generation = identity();
        const ticket = ++historyTicket;
        includeArchived = archived;
        historyLoading = true;
        let next: AssistantConversationList;
        try {
            next = await result(api.list({ cursor, archived, limit: 30 }));
        } catch (failure) {
            if (generation !== identity() || ticket !== historyTicket) return;
            throw failure;
        } finally {
            if (ticket === historyTicket) historyLoading = false;
        }
        if (generation !== identity() || ticket !== historyTicket) return;
        // Explicit pagination keeps older pages reachable without retaining all records.
        const items = cursor ? [...history.value.items, ...next.items] : next.items;
        history.value = {
            ...next,
            items: items.filter((item, index) => items.findIndex((other) => other.id === item.id) === index).slice(-90),
        };
        lastRefresh = Date.now();
    }

    async function search(query: string, cursor?: string, includeVersions = false, archived = false) {
        if (disposed) return;
        const generation = identity();
        const ticket = ++searchTicket;
        let next: AssistantSearchResult;
        try {
            next = await result(api.search({ query, cursor, includeVersions, archived, limit: 30 }));
        } catch (failure) {
            if (generation !== identity() || ticket !== searchTicket) return;
            throw failure;
        }
        if (generation !== identity() || ticket !== searchTicket) return;
        const items = cursor ? [...searchResults.value.items, ...next.items] : next.items;
        searchResults.value = {
            ...next,
            items: items
                .filter(
                    (item, index) =>
                        items.findIndex(
                            (other) => other.messageId === item.messageId && other.versionId === item.versionId,
                        ) === index,
                )
                .slice(-90),
        };
    }

    async function setVisible(value: boolean) {
        visible = value;
        if (value) await loadHistory();
    }

    async function refreshVisible() {
        if (!historyLoading && visible && Date.now() - lastRefresh >= 30000) await loadHistory();
    }

    function dispose() {
        disposed = true;
        reset();
    }

    return { dispose, history, searchResults, loadHistory, search, reset, setVisible, refreshVisible };
}
