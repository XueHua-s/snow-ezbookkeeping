<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import { ASSISTANT_TOOL_LABELS } from '@/consts/assistant.ts';
import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';
import type { AssistantRunSnapshot } from '@/models/ai_assistant_conversation.ts';
const props = defineProps<{ answer: AssistantRunSnapshot; canPage?: boolean; pageLabel?: string }>();
const emit = defineEmits<{ openReference: [id: string]; nextPage: [] }>();
const { tt, formatAmountToLocalizedNumeralsWithCurrency } = useI18n();
const references = computed(() =>
    props.answer.references
        .filter((item) => /^[1-9]\d*$/.test(item.id))
        .map((item) => ({
            ...item,
            amount: item.hideAmount
                ? DISPLAY_HIDDEN_AMOUNT
                : formatAmountToLocalizedNumeralsWithCurrency(
                      parseBigDecimal(item.sourceAmountMinor),
                      item.currency || false
                  ),
            destination: item.hideAmount
                ? DISPLAY_HIDDEN_AMOUNT
                : item.destinationAmountMinor === undefined
                  ? ''
                  : formatAmountToLocalizedNumeralsWithCurrency(
                        parseBigDecimal(item.destinationAmountMinor),
                        item.destinationCurrency || false
                    )
        }))
);
</script>
<template>
    <section class="evidence">
        <div v-for="query in answer.queries" :key="query.queryId">
            <p v-if="query.coverage !== 'all_matching'">{{ tt('Current page only') }}</p>
            <p v-for="warning in query.warnings" :key="warning" role="note">{{ tt(warning) }}</p>
        </div>
        <details v-if="answer.queries.length">
            <summary>{{ tt('Query Evidence') }} · {{ answer.queries.length }}</summary>
            <section v-for="query in answer.queries" :key="query.queryId" class="query">
                <strong
                    >{{ tt(ASSISTANT_TOOL_LABELS[query.tool]) }} ·
                    {{
                        tt(query.coverage === 'all_matching' ? 'All matching records' : 'Current page only')
                    }}</strong
                >
                <p>
                    {{ query.scope.startTime || tt('No start limit') }} →
                    {{ query.scope.endTime || tt('No end limit') }}
                </p>
                <p>
                    {{
                        tt(query.scope.includesHidden ? 'Hidden records included' : 'Hidden records excluded')
                    }}
                    · {{ tt('Deleted records excluded') }}
                </p>
                <p v-if="query.scope.accountIds?.length">
                    {{ tt('Selected accounts: {count}', { count: query.scope.accountIds.length }) }}
                </p>
                <p v-if="query.scope.categoryIds?.length">
                    {{ tt('Selected categories: {count}', { count: query.scope.categoryIds.length }) }}
                </p>
                <p v-if="query.scope.currency">{{ tt('Currency') }}: {{ query.scope.currency }}</p>
                <p v-if="query.scope.keyword">{{ tt('Keyword') }}: {{ query.scope.keyword }}</p>
                <p>
                    {{ tt('Timezone') }}: {{ query.scope.timezone }} · {{ tt('Read at') }}:
                    {{ query.queriedAt }}
                </p>
                <p v-for="period in query.scope.periods" :key="period.name">
                    {{ period.name }}: {{ period.startTime }} → {{ period.endTime }}
                </p>
            </section>
        </details>
        <div v-if="references.length" class="bills">
            <button
                v-for="item in references"
                :key="item.id"
                type="button"
                @click="emit('openReference', item.id)"
            >
                <span>{{ item.categoryName || tt('Uncategorized') }} · {{ item.timeText }}</span
                ><strong
                    >{{ item.amount
                    }}<template v-if="item.destination"> → {{ item.destination }}</template></strong
                ><span
                    >{{ item.sourceAccountName
                    }}<template v-if="item.destinationAccountName">
                        → {{ item.destinationAccountName }}</template
                    ></span
                ><span>{{ tt('View Details') }}</span>
            </button>
        </div>
        <button v-if="canPage && answer.pageableQueryId" type="button" @click="emit('nextPage')">
            {{ pageLabel || tt('View more bills') }}
        </button>
    </section>
</template>
<style scoped>
.evidence {
    margin-top: 12px;
    font-size: 0.85em;
    line-height: 1.65;
}
summary {
    min-height: 44px;
    padding: 10px 0;
    cursor: pointer;
    transition: background 180ms ease, transform 180ms ease;
}
.query {
    border-inline-start: 2px solid var(--assistant-border, #e5ece4);
    padding-inline-start: 12px;
}
.bills {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(220px, 100%), 1fr));
    gap: 8px;
}
button {
    min-height: 44px;
    padding: 12px;
    border: 1px solid var(--assistant-border, #e5ece4);
    border-radius: 17px;
    background: var(--assistant-card, #fff);
    color: inherit;
    font: inherit;
    text-align: start;
    cursor: pointer;
    transition: background 180ms ease, transform 180ms ease;
}
.bills button {
    display: flex;
    flex-direction: column;
    gap: 4px;
}
strong {
    font-variant-numeric: tabular-nums;
}
button:hover {
    background: var(--assistant-tint, #e8f3e9);
}
button:focus-visible,
summary:focus-visible {
    outline: 2px solid var(--assistant-action, #278c62);
    outline-offset: 2px;
}
button:active { transform: scale(.99); }
details[open] .query { animation: evidence-enter 180ms ease; }
@keyframes evidence-enter { from { opacity: 0; transform: translateY(-3px); } to { opacity: 1; transform: translateY(0); } }
@media (prefers-reduced-motion: reduce) { button, summary { transition: none; } details[open] .query { animation: none; } }
</style>
