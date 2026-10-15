<script setup lang="ts">
import { useI18n } from '@/locales/helpers.ts';
import { TransactionListPageType } from '@/views/base/transactions/TransactionListPageBase.ts';
import type { useListPage } from '@/views/desktop/transactions/ListPage/useListPage.ts';
import type { ShallowUnwrapRef } from 'vue';

type Scope = ShallowUnwrapRef<ReturnType<typeof useListPage>>;

defineProps<{
    loading: Scope['loading'];
    pageType: Scope['pageType'];
    allPageCounts: Scope['allPageCounts'];
    recentMonthDateRanges: Scope['recentMonthDateRanges'];
    recentDateRangeIndex: Scope['recentDateRangeIndex'];
}>();

const queryPageType = defineModel<Scope['queryPageType']>('queryPageType', { required: true });
const queryType = defineModel<Scope['queryType']>('queryType', { required: true });
const countPerPage = defineModel<Scope['countPerPage']>('countPerPage', { required: true });

const emit = defineEmits<{
    changeDateFilter: Parameters<Scope['changeDateFilter']>;
}>();

const changeDateFilter = (...args: Parameters<Scope['changeDateFilter']>) => emit('changeDateFilter', ...args);

const { tt } = useI18n();

</script>

<template>
    <div class="mb-2">
                <btn-vertical-group :disabled="loading" :buttons="TransactionListPageType.values().map(item => {
        return {
            name: tt(item.name),
            value: item.type
        }
    })" v-model="queryPageType" />
            </div><v-divider class="my-2" /><div class="my-2">
                <span class="mx-3 text-body-medium">{{ tt('Transaction Type') }}</span>
                <v-select
                    item-title="displayName"
                    item-value="type"
                    class="mt-1"
                    density="compact"
                    :disabled="loading"
                    :items="[
                        { displayName: tt('All Types'), type: 0 },
                        { displayName: tt('Modify Balance'), type: 1 },
                        { displayName: tt('Income'), type: 2 },
                        { displayName: tt('Expense'), type: 3 },
                        { displayName: tt('Transfer'), type: 4 }
                    ]"
                    v-model="queryType"
                />
            </div><div class="my-2" v-if="pageType === TransactionListPageType.List.type || pageType === TransactionListPageType.Gallery.type">
                <span class="mx-3 text-body-medium">{{ tt('Transactions Per Page') }}</span>
                <v-select class="mt-1" density="compact"
                          item-title="name"
                          item-value="value"
                          :disabled="loading"
                          :items="allPageCounts"
                          v-model="countPerPage"
                />
            </div><li class="nav-link" :key="idx" v-for="(recentDateRange, idx) in recentMonthDateRanges">
                <a class="d-flex align-center cursor-pointer my-1"
                   :class="{ 'router-link-active router-link-exact-active': recentDateRangeIndex === idx, 'disabled': loading }"
                   @click="changeDateFilter(recentDateRange)">
                    <span class="nav-item-title text-truncate">{{ recentDateRange.displayName }}</span>
                </a>
            </li>
</template>
