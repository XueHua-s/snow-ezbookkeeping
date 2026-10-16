<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';
import type { AssistantRunSnapshot } from '@/models/ai_assistant_conversation.ts';
const props = defineProps<{ answer: AssistantRunSnapshot; canPage?: boolean; pageLabel?: string }>();
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
button:focus-visible {
    outline: 2px solid var(--assistant-action, #278c62);
    outline-offset: 2px;
}
button:active { transform: scale(.99); }
@media (prefers-reduced-motion: reduce) { button { transition: none; } }
</style>
