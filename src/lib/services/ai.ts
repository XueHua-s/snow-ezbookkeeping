import {
    BASE_API_URL_PATH,
    DEFAULT_LLM_API_TIMEOUT
} from '@/consts/api.ts';
import type { ApiResponse } from '@/core/api.ts';
import { readAssistantStream, startAssistantDeadline } from '@/lib/assistant_stream.ts';
import {
    getBrowserTimezoneName,
    getCurrentUnixTime,
    getTimezoneOffsetMinutes
} from '@/lib/datetime.ts';
import {
    getAIAssistantRunTimeout
} from '@/lib/server_settings.ts';
import {
    getTimeZone
} from '@/lib/settings.ts';
import type { AIAssistantChatStreamCallbacks, AIAssistantErrorEnvelope } from '@/lib/types/assistant.ts';
import {
    clearCurrentTokenAndUserInfo,
    getCurrentToken
} from '@/lib/userstate.ts';
import { getBasePath } from '@/lib/web.ts';
import type {
AIAssistantChatRequest,
AIAssistantChatResponse,
CodingAssistantRequest,
CodingAssistantResponse,
RecognizedTransactionResponse
} from '@/models/large_language_model.ts';
import axios from 'axios';

import { type ApiRequestConfig, type ApiResponsePromise, cancelableStreamingRequests } from './transport.ts';

export const aiServices = {
    recognizeTransactionText: ({ text }: { text: string }): ApiResponsePromise<RecognizedTransactionResponse> => {
        return axios.post<ApiResponse<RecognizedTransactionResponse>>('v1/llm/transactions/recognize_text.json', {
            text: text
        }, {
            timeout: DEFAULT_LLM_API_TIMEOUT
        } as ApiRequestConfig);
    },
    recognizeReceiptImage: ({ imageFile, cancelableUuid }: { imageFile: File, cancelableUuid?: string }): ApiResponsePromise<RecognizedTransactionResponse> => {
        return axios.postForm<ApiResponse<RecognizedTransactionResponse>>('v1/llm/transactions/recognize_receipt_image.json', {
            image: imageFile
        }, {
            timeout: DEFAULT_LLM_API_TIMEOUT,
            cancelableUuid: cancelableUuid
        } as ApiRequestConfig);
    },
    generateCode: (req: CodingAssistantRequest, cancelableUuid?: string): ApiResponsePromise<CodingAssistantResponse> => {
        return axios.post<ApiResponse<CodingAssistantResponse>>('v1/llm/code/generate.json', req, {
            timeout: DEFAULT_LLM_API_TIMEOUT,
            cancelableUuid: cancelableUuid
        } as ApiRequestConfig);
    },
    chatWithAIAssistant: async ({ req, cancelableUuid }: { req: AIAssistantChatRequest, cancelableUuid?: string }): ApiResponsePromise<AIAssistantChatResponse> => {
        const controller = new AbortController();
        const stopDeadline = startAssistantDeadline(controller, (getAIAssistantRunTimeout() ?? DEFAULT_LLM_API_TIMEOUT) + 10000);
        if (cancelableUuid) {
            cancelableStreamingRequests[cancelableUuid] = controller;
        }
        try {
            return await axios.post<ApiResponse<AIAssistantChatResponse>>('v1/llm/assistant/chat.json', req, {
                timeout: 0, // The shared deadline supports durations beyond browser timer limits.
                signal: controller.signal
            });
        } catch (error: unknown) {
            if (controller.signal.aborted) {
                if (controller.signal.reason?.name === 'TimeoutError') {
                    throw { timedOut: true, incomplete: true, message: 'Unable to get AI assistant response' };
                }
                throw { canceled: true };
            }
            throw error;
        } finally {
            stopDeadline();
            if (cancelableUuid && cancelableStreamingRequests[cancelableUuid] === controller) {
                delete cancelableStreamingRequests[cancelableUuid];
            }
        }
    },
    chatWithAIAssistantStream: async ({ req, cancelableUuid, callbacks }: {
        req: AIAssistantChatRequest;
        cancelableUuid?: string;
        callbacks?: AIAssistantChatStreamCallbacks;
    }): Promise<void> => {
        const token = getCurrentToken();
        const abortController = new AbortController();
        const stopDeadline = startAssistantDeadline(abortController, (getAIAssistantRunTimeout() ?? DEFAULT_LLM_API_TIMEOUT) + 10000);

        if (cancelableUuid) {
            cancelableStreamingRequests[cancelableUuid] = abortController;
        }

        let timezoneName = getTimeZone();

        if (!timezoneName || timezoneName.trim().length < 1) {
            timezoneName = getBrowserTimezoneName();
        }

        try {
            const requestHeaders: Record<string, string> = {
                'Content-Type': 'application/json',
                'Accept-Language': axios.defaults.headers.common['Accept-Language'] as string || '',
                'X-Timezone-Offset': String(getTimezoneOffsetMinutes(getCurrentUnixTime())),
                'X-Timezone-Name': timezoneName
            };

            if (token) {
                requestHeaders['Authorization'] = `Bearer ${token}`;
            }

            const response = await fetch(`${getBasePath()}${BASE_API_URL_PATH}/v1/llm/assistant/chat/stream.json`, {
                method: 'POST',
                headers: requestHeaders,
                body: JSON.stringify(req),
                signal: abortController.signal
            });

            if (!response.ok) {
                let errorPayload: AIAssistantErrorEnvelope | null = null;

                try {
                    errorPayload = await response.json() as AIAssistantErrorEnvelope;
                } catch {
                    // A non-JSON HTTP error still fails without fabricating completion.
                }

                if (errorPayload && errorPayload.errorCode && (
                    errorPayload.errorCode === 202001
                    || errorPayload.errorCode === 202002
                    || errorPayload.errorCode === 202003
                    || errorPayload.errorCode === 202004
                    || errorPayload.errorCode === 202005
                    || errorPayload.errorCode === 202006
                    || errorPayload.errorCode === 202012
                )) {
                    clearCurrentTokenAndUserInfo(false);
                    location.reload();
                    throw { processed: true };
                }

                if (errorPayload) {
                    throw { response: { data: errorPayload } };
                }

                throw { message: 'Unable to get AI assistant response' };
            }

            if (!response.body) {
                throw { message: 'Unable to get AI assistant response' };
            }

            await readAssistantStream(response.body, callbacks, abortController.signal);
        } catch (error: unknown) {
            const typedError = error as {
                name?: string;
                canceled?: boolean;
            };

            if (abortController.signal.aborted || typedError.name === 'AbortError' || typedError.canceled) {
                if (abortController.signal.reason?.name === 'TimeoutError') {
                    throw { timedOut: true, incomplete: true, message: 'Unable to get AI assistant response' };
                }
                throw { canceled: true };
            }

            throw error;
        } finally {
            stopDeadline();
            abortController.abort();
            if (cancelableUuid && cancelableStreamingRequests[cancelableUuid] === abortController) {
                delete cancelableStreamingRequests[cancelableUuid];
            }
        }
    }
};
