<script setup lang="ts">
import { ASSISTANT_STATE_LABELS, ASSISTANT_PROGRESS_LABELS, ASSISTANT_TOOL_LABELS, ASSISTANT_MAX_REFERENCES, ASSISTANT_TRANSACTION_TYPE_LABELS, ASSISTANT_TRANSFER_SIDE_LABELS } from '@/consts/assistant.ts';
import { computed, type DeepReadonly } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import { TransactionType } from '@/core/transaction.ts';
import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';
import type { AIAssistantConversationMessage } from '@/views/base/assistant/types.ts';
import type { AIAssistantQueryScope } from '@/models/large_language_model.ts';

const props = defineProps<{ message: AIAssistantConversationMessage }>();
const emit = defineEmits<{ openReference: [id: string] }>();
const { tt, formatAmountToLocalizedNumeralsWithCurrency } = useI18n();
const references = computed(() => props.message.references?.filter(reference => /^[1-9]\d*$/.test(reference.id)).slice(0, ASSISTANT_MAX_REFERENCES).map(reference => ({
    ...reference,
    sourceAmountText: reference.hideAmount ? DISPLAY_HIDDEN_AMOUNT : !reference.sourceAmountMinor && !Number.isSafeInteger(reference.sourceAmount) ? tt('Amount cannot be displayed precisely') : formatAmountToLocalizedNumeralsWithCurrency(parseBigDecimal(reference.sourceAmountMinor ?? (Number.isSafeInteger(reference.sourceAmount) ? reference.sourceAmount : 0)), reference.currency || false),
    destinationAmountText: reference.hideAmount ? DISPLAY_HIDDEN_AMOUNT : !reference.destinationAmountMinor && !Number.isSafeInteger(reference.destinationAmount ?? 0) ? tt('Amount cannot be displayed precisely') : formatAmountToLocalizedNumeralsWithCurrency(parseBigDecimal(reference.destinationAmountMinor ?? (Number.isSafeInteger(reference.destinationAmount ?? 0) ? reference.destinationAmount ?? 0 : 0)), reference.destinationCurrency || false)
})) ?? []);
const queries = computed(() => props.message.queries?.map(query => ({ ...query, scopeText: scopeDetails(query.scope).join(' · ') })) ?? []);

// Render the validated business projection, never the original arguments or continuation token.
function scopeDetails(scope: DeepReadonly<AIAssistantQueryScope>): string[] {
    const details = [scope.includesHidden ? tt('Hidden records included') : tt('Hidden records excluded')];
    if (scope.excludesDeleted) details.push(tt('Deleted records excluded'));
    if (scope.accountIds?.length) details.push(tt('Selected accounts: {count}', { count: scope.accountIds.length }));
    else details.push(tt('All valid accounts'));
    if (scope.categoryIds?.length) details.push(tt('Selected categories: {count}', { count: scope.categoryIds.length }));
    if (scope.tagIds?.length) details.push(tt(scope.tagMatch === 'all' ? 'All selected tags: {count}' : 'Any selected tags: {count}', { count: scope.tagIds.length }));
    if (scope.noTags) details.push(tt('Without Tags'));
    if (scope.currency) details.push(tt('Currency') + ': ' + scope.currency);
    if (scope.types?.length) details.push(scope.types.map(type => tt(ASSISTANT_TRANSACTION_TYPE_LABELS[type] ?? 'Unknown')).join(', '));
    if (scope.transferSide) details.push(tt(ASSISTANT_TRANSFER_SIDE_LABELS[scope.transferSide]));
    if (scope.amountMin !== undefined) details.push(tt('Minimum amount') + ': ' + scope.amountMin);
    if (scope.amountMax !== undefined) details.push(tt('Maximum amount') + ': ' + scope.amountMax);
    if (scope.keyword) details.push(tt('Keyword') + ': ' + scope.keyword);
    return details;
}
</script>

<template>
    <div class="assistant-evidence" v-if="message.role === 'assistant'">
        <div class="assistant-run-status" :class="`assistant-run-${message.state}`" role="status" aria-live="polite">
            <span class="assistant-status-dot" aria-hidden="true"></span>
            {{ tt(message.state === 'requesting' && message.progress ? ASSISTANT_PROGRESS_LABELS[message.progress] : ASSISTANT_STATE_LABELS[message.state]) }}
        </div>
        <p class="assistant-partial-notice" v-if="message.state === 'incomplete' || message.state === 'failed' || message.state === 'canceled'">
            {{ tt('This answer was not completed. Partial text may be unreliable.') }}
        </p>
        <div class="assistant-query-list" v-if="queries.length">
            <div class="assistant-evidence-title">{{ tt('Query Evidence') }}</div>
            <section class="assistant-query" :key="query.queryId" v-for="query in queries">
                <div class="assistant-query-heading">
                    <strong>{{ tt(ASSISTANT_TOOL_LABELS[query.tool]) }}</strong>
                    <span class="assistant-coverage">{{ tt(query.coverage === 'all_matching' ? 'All matching records' : 'Current page only') }}</span>
                </div>
                <div class="assistant-query-periods" v-if="query.scope.periods?.length">
                    <div :key="period.name" v-for="period in query.scope.periods">
                        {{ period.name }}: {{ period.startTime }} → {{ period.endTime }} ({{ tt('End time excluded') }})
                        <span v-if="period.matchedCount !== undefined"> · {{ tt('Matching transactions: {count}', { count: period.matchedCount }) }}</span>
                    </div>
                </div>
                <div v-else-if="query.tool === 'get_account_balances'">{{ tt('Current balances at query time') }}</div>
                <div v-else>
                    {{ query.scope.startTime || tt('No start limit') }} → {{ query.scope.endTime || tt('No end limit') }}
                    <span v-if="query.scope.endTime"> ({{ tt('End time excluded') }})</span>
                </div>
                <div>{{ tt('Timezone') }}: {{ query.scope.timezone }}</div>
                <div class="assistant-scope-details">{{ query.scopeText }}</div>
                <div class="assistant-query-time">{{ tt('Read at') }}: {{ query.queriedAt }}</div>
                <p class="assistant-query-warning" :key="warning" v-for="warning in query.warnings">{{ tt(warning) }}</p>
            </section>
        </div>
        <div class="assistant-message-references" v-if="references.length">
            <div class="assistant-evidence-title">{{ tt('Referenced Bills') }}</div>
            <p class="assistant-reference-notice">{{ tt('Bills read during this query; individual conclusions may use other query evidence.') }}</p>
            <button class="assistant-reference-item" type="button" :key="reference.id" v-for="reference in references"
                    @click="emit('openReference', reference.id)">
                <span>{{ reference.timeText || '-' }} · {{ reference.categoryName || tt('Uncategorized') }}</span>
                <span>{{ reference.sourceAccountName || tt('Account') }}<template v-if="reference.destinationAccountName"> → {{ reference.destinationAccountName }}</template></span>
                <span>{{ reference.sourceAmountText }}<template v-if="reference.type === TransactionType.Transfer"> → {{ reference.destinationAmountText }}</template></span>
                <span class="assistant-reference-open">{{ tt('View Details') }}</span>
            </button>
        </div>
    </div>
</template>

<style scoped>
.assistant-evidence { margin-top: 12px; font-size: 12px; line-height: 1.55; overflow-wrap: anywhere; }
.assistant-run-status { display: flex; align-items: center; gap: 7px; opacity: 0.8; }
.assistant-status-dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; flex-shrink: 0; }
.assistant-run-incomplete, .assistant-run-failed { color: #ad5700; }
.assistant-run-requesting .assistant-status-dot { animation: assistant-pulse 1.5s ease-in-out infinite; }
.assistant-partial-notice, .assistant-query-warning { margin: 6px 0 0; }
.assistant-query-list, .assistant-message-references { margin-top: 12px; padding-top: 10px; border-top: 1px solid rgba(127, 127, 127, 0.2); }
.assistant-evidence-title { font-weight: 600; margin-bottom: 6px; }
.assistant-query { padding: 9px 0; }
.assistant-query + .assistant-query { border-top: 1px solid rgba(127, 127, 127, 0.15); }
.assistant-query-heading { display: flex; flex-wrap: wrap; align-items: baseline; gap: 8px; margin-bottom: 4px; }
.assistant-coverage { padding: 1px 6px; border-radius: 4px; background: rgba(127, 127, 127, 0.1); }
.assistant-query-time, .assistant-reference-notice { opacity: 0.7; }
.assistant-reference-notice { margin: 0 0 7px; }
.assistant-reference-item { display: flex; width: 100%; flex-direction: column; align-items: flex-start; gap: 2px; text-align: start; font: inherit; color: inherit; cursor: pointer; background: var(--assistant-card, #fff); border: 1px solid var(--assistant-border, #e5ece4); border-radius: 16px; transition: background 180ms ease, transform 180ms ease; padding: 8px 10px; margin-top: 6px; }
.assistant-reference-item:hover { background: var(--assistant-tint, #e8f3e9); }
.assistant-reference-item:focus-visible { outline: 2px solid var(--assistant-action, #278c62); outline-offset: 2px; }
.assistant-reference-open { color: var(--assistant-action, #278c62); }
@keyframes assistant-pulse { 50% { opacity: 0.3; } }
@media (prefers-reduced-motion: reduce) { .assistant-reference-item { transition: none; } .assistant-run-requesting .assistant-status-dot { animation: none; } }
</style>
