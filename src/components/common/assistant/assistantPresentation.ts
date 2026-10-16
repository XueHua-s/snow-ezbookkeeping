import type { AssistantReasonCode, AssistantRunState } from '@/models/ai_assistant_conversation.ts';
export const assistantStateLabels: Record<AssistantRunState, string> = {
    queued: 'Queued', preparing: 'Organizing conversation context', running: 'Generating answer', cancel_requested: 'Stopping generation', completed: 'Answer completed', failed: 'Assistant request failed', incomplete: 'Answer incomplete', interrupted: 'Generation interrupted', canceled: 'Assistant request canceled'
};
export const assistantReasonLabels: Record<AssistantReasonCode, string> = {
    queue_timeout: 'The request waited too long in the queue. You can retry.',
    run_timeout: 'Generation timed out. The saved partial answer is kept; you can retry.',
    execution_limit: 'The answer reached its execution limit. You can refine your question and retry.',
    generation_unavailable: 'Generation is currently unavailable. Your saved conversations remain accessible.',
    input_too_long: 'Input exceeds the limit. Your text has been kept.',
    conversation_conflict: 'This conversation changed. Review the latest messages before sending again.',
    data_epoch_conflict: 'Assistant data was cleared. Review the current history before sending again.',
    memory_conflict: 'Memory changed on another device. Review the latest version.',
    context_changed: 'Related memory or sources changed. You can regenerate this answer.',
    already_running: 'This conversation already has an active answer. Stop it before sending again.',
    quota_exceeded: 'Assistant capacity is full. Export or manage your data to continue.',
    queue_full: 'The assistant is busy. Try again later; your question is kept.',
    subscription_limit: 'Live updates are busy. Checking the saved answer instead.',
    cursor_expired: 'This query expired. Run the query again to see current records.',
    context_limit: 'This conversation is too large to process. Continue in a new conversation.',
    output_limit: 'The answer reached its limit. The partial text is kept.',
    empty_response: 'The model returned no answer. You can retry your question.',
    storage_unavailable: 'The answer could not be saved. Check the connection and retry.',
    provider_unavailable: 'The model service is unavailable. You can retry later.',
    interrupted: 'Generation was interrupted. You can retry your question.',
    not_accessible: 'This item does not exist or is not accessible.'
};
