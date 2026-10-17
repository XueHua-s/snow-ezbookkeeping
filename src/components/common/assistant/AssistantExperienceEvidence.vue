<script setup lang="ts">
import { computed } from 'vue';
import { mdiArrowRight } from '@mdi/js';
import { useI18n } from '@/locales/helpers.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';
import type { AssistantRunSnapshot } from '@/models/ai_assistant_conversation.ts';
const props = defineProps<{ answer: AssistantRunSnapshot; canPage?: boolean; pageLabel?: string; disabledReason?: string }>();
const emit = defineEmits<{ openReference: [id: string]; nextPage: [] }>();
const { tt, formatAmountToLocalizedNumeralsWithCurrency } = useI18n();
const warnings = computed(() => [...new Set(props.answer.queries.flatMap(query => query.warnings))]);
const hasPageDetails = computed(() => props.answer.queries.some(query => query.coverage === 'page'));
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
        <p v-for="warning in warnings" :key="warning" role="note">{{ tt(warning) }}</p>
        <p v-if="hasPageDetails">{{ tt('Displayed details cover this page only.') }}</p>
        <p v-if="answer.pageableQueryId">{{ tt('More matching records are available.') }}</p>
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
        <div v-if="canPage && answer.pageableQueryId" class="follow-up">
            <span class="follow-up-label">{{ tt('Keep asking') }}</span>
            <button class="more-bills" type="button" :disabled="!!disabledReason" :title="disabledReason ? tt(disabledReason) : undefined" @click="emit('nextPage')">
                <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="mdiArrowRight" /></svg>
                {{ pageLabel || tt('Continue finding matching bills') }}
            </button>
            <span class="follow-up-hint">{{ tt('The assistant will continue this query in another answer.') }}</span>
        </div>
    </section>
</template>
<style scoped>
.evidence {
    margin-top: 12px;
    font-size: 0.85em;
    line-height: 1.65;
}
.bills {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(220px, 100%), 1fr));
    gap: 8px;
}
.follow-up { display: flex; flex-direction: column; align-items: flex-start; gap: 6px; margin-block-start: 12px; }
.follow-up-label, .follow-up-hint { color: var(--assistant-muted, #6b8074); }
.more-bills { display: inline-flex; align-items: center; gap: 8px; }
.more-bills svg { width: 18px; height: 18px; flex-shrink: 0; fill: currentColor; }
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
button:disabled { opacity: .55; cursor: default; }
button:hover:not(:disabled) {
    background: var(--assistant-tint, #e8f3e9);
}
button:focus-visible {
    outline: 2px solid var(--assistant-action, #278c62);
    outline-offset: 2px;
}
button:active:not(:disabled) { transform: scale(.99); }
@media (prefers-reduced-motion: reduce) { button { transition: none; } }
</style>
