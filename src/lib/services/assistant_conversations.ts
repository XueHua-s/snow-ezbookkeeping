import axios, { type AxiosResponse } from 'axios';
import type { ApiResponse } from '@/core/api.ts';
import type * as DTO from '@/models/ai_assistant_conversation.ts';
import { BASE_API_URL_PATH } from '@/consts/api.ts';
import { getCurrentToken } from '@/lib/userstate.ts';
import { getBasePath } from '@/lib/web.ts';
import { getBrowserTimezoneName, getCurrentUnixTime, getTimezoneOffsetMinutes } from '@/lib/datetime.ts';
import { getTimeZone } from '@/lib/settings.ts';
import { AssistantSubscriptionError, readAssistantConversationStream, type AssistantSnapshotCallbacks } from '@/lib/assistant_conversation_stream.ts';
import './transport.ts';
import type { ApiResponsePromise, ApiRequestConfig } from './transport.ts';

const prefix = 'v1/llm/assistant/';
function get<T>(path: string, params: unknown, signal?: AbortSignal): ApiResponsePromise<T> {
    return axios.get<ApiResponse<T>>(prefix + path + '.json', { params, signal, ignoreError: true } as ApiRequestConfig);
}
function post<T>(path: string, request: unknown, signal?: AbortSignal): ApiResponsePromise<T> {
    // No retry interceptor: ambiguous write outcomes must be resolved using commands/get.
    return axios.post<ApiResponse<T>>(prefix + path + '.json', request, { signal, ignoreError: true } as ApiRequestConfig);
}

export const assistantConversationServices = {
    capabilities: (signal?: AbortSignal) => get<DTO.AssistantCapabilities>('capabilities', {}, signal),
    command: (clientRequestId: string, signal?: AbortSignal) => get<DTO.AssistantCommandResult>('commands/get', { clientRequestId }, signal),
    list: (request: DTO.AssistantListRequest, signal?: AbortSignal) => get<DTO.AssistantConversationList>('conversations/list', request, signal),
    search: (request: DTO.AssistantSearchRequest, signal?: AbortSignal) => post<DTO.AssistantSearchResult>('conversations/search', request, signal),
    get: (request: DTO.AssistantGetConversationRequest, signal?: AbortSignal) => get<DTO.AssistantConversationDetail>('conversations/get', request, signal),
    state: (id: string, signal?: AbortSignal) => get<DTO.AssistantConversationState>('conversations/state', { id }, signal),
    update: (request: DTO.AssistantUpdateConversationRequest) => post<DTO.AssistantCommandResult>('conversations/update', request),
    delete: (request: DTO.AssistantDeleteConversationRequest) => post<DTO.AssistantCommandResult>('conversations/delete', request),
    clear: (request: DTO.AssistantClearRequest) => post<DTO.AssistantCommandResult>('conversations/clear', request),
    branch: (request: DTO.AssistantBranchRequest) => post<DTO.AssistantCommandResult>('conversations/branch', request),
    export: (request: DTO.AssistantExportRequest, signal?: AbortSignal): Promise<AxiosResponse<Blob>> =>
        axios.post(prefix + 'conversations/export.json', request, { responseType: 'blob', signal, timeout: 0, ignoreError: true } as ApiRequestConfig),
    send: (request: DTO.AssistantSendRequest) => post<DTO.AssistantCommandResult>('messages/send', request),
    retry: (request: DTO.AssistantRetryRequest) => post<DTO.AssistantCommandResult>('messages/retry', request),
    message: (assistantMessageId: string, signal?: AbortSignal) => get<DTO.AssistantRunSnapshot>('messages/get', { assistantMessageId }, signal),
    versions: (request: DTO.AssistantVersionsRequest, signal?: AbortSignal) => get<DTO.AssistantVersionsResult>('messages/versions', request, signal),
    selectVersion: (request: DTO.AssistantSelectVersionRequest) => post<DTO.AssistantCommandResult>('messages/select_version', request),
    run: (request: DTO.AssistantRunRequest, signal?: AbortSignal) => get<DTO.AssistantRunSnapshot>('runs/get', request, signal),
    cancel: (request: DTO.AssistantCancelRequest) => post<DTO.AssistantCommandResult>('runs/cancel', request),
    memories: (signal?: AbortSignal) => get<DTO.AssistantMemoryList>('memories/list', {}, signal),
    saveMemory: (request: DTO.AssistantSaveMemoryRequest) => post<DTO.AssistantCommandResult>('memories/save', request),
    deleteMemory: (request: DTO.AssistantDeleteMemoryRequest) => post<DTO.AssistantCommandResult>('memories/delete', request),
    memorySettings: (request: DTO.AssistantMemorySettingsRequest) => post<DTO.AssistantCommandResult>('memories/settings', request),
    subscribe: subscribeAssistantRun
};

/** A subscription is only a reader. Aborting it never calls runs/cancel or sends a prompt. */
export async function subscribeAssistantRun(runId: string, afterRevision: number, callbacks: AssistantSnapshotCallbacks,
    signal?: AbortSignal): Promise<void> {
    const token = getCurrentToken();
    const headers: Record<string, string> = {
        Accept: 'text/event-stream',
        'Accept-Language': String(axios.defaults.headers.common['Accept-Language'] || ''),
        'X-Timezone-Offset': String(getTimezoneOffsetMinutes(getCurrentUnixTime())),
        'X-Timezone-Name': getTimeZone()?.trim() || getBrowserTimezoneName()
    };
    if (token) { headers['Authorization'] = `Bearer ${token}`; }
    const params = new URLSearchParams({ runId, afterRevision: String(afterRevision) });
    const response = await fetch(`${getBasePath()}${BASE_API_URL_PATH}/${prefix}runs/stream.json?${params}`, {
        headers, signal, cache: 'no-store'
    });
    if (!response.ok) {
        const retry = response.headers.get('Retry-After');
        const retryAfterMs = retry ? (/^\d+$/.test(retry) ? Number(retry) * 1000 : Math.max(0, Date.parse(retry) - Date.now())) : undefined;
        let envelope: unknown;
        try { envelope = await response.json(); } catch { /* Status alone still identifies authentication/rate failures. */ }
        throw new AssistantSubscriptionError(response.status === 401 ? 'unauthorized' : response.status === 429 ? 'limited' : 'http',
            response.status, Number.isFinite(retryAfterMs) ? retryAfterMs : undefined, envelope);
    }
    if (!response.body || !response.headers.get('Content-Type')?.toLowerCase().startsWith('text/event-stream')) {
        throw new AssistantSubscriptionError('protocol');
    }
    await readAssistantConversationStream(response.body, runId, callbacks, signal, afterRevision);
}
