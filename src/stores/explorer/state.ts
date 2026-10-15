import { ALL_CURRENCIES } from '@/consts/currency.ts';
import { keys } from '@/core/base.ts';
import type { DateTime } from '@/core/datetime.ts';
import {
    DEFAULT_TRANSACTION_EXPLORER_DATE_RANGE
} from '@/core/explorer.ts';
import { NormalizedText } from '@/core/text.ts';
import { TimezoneTypeForStatistics } from '@/core/timezone.ts';
import {
    isDefined
} from '@/lib/common.ts';
import {
    parseDateTimeFromUnixTime,
    parseDateTimeFromUnixTimeWithTimezoneOffset
} from '@/lib/datetime.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import {
    type InsightsExplorerMatchContext,
    InsightsExplorer,
    type InsightsExplorerBasicInfo
} from '@/models/explorer.ts';
import type { TransactionInfoResponse, TransactionInsightDataItem } from '@/models/transaction.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useUserStore } from '@/stores/user.ts';
import type { Ref, UnwrapRef } from 'vue';
import { ref } from 'vue';
import type { TransactionExplorerFilter } from './types.ts';
interface ExplorerState {
    settingsStore: ReturnType<typeof useSettingsStore>;
    userStore: ReturnType<typeof useUserStore>;
    accountsStore: ReturnType<typeof useAccountsStore>;
    transactionCategoriesStore: ReturnType<typeof useTransactionCategoriesStore>;
    transactionTagsStore: ReturnType<typeof useTransactionTagsStore>;
    exchangeRatesStore: ReturnType<typeof useExchangeRatesStore>;
    currencyDisplayOrders: Record<string, number>;
    buildInsightsExplorerMatchContext: (exploration: InsightsExplorer, transaction: TransactionInsightDataItem) => InsightsExplorerMatchContext;
    transactionExplorerFilter: Ref<UnwrapRef<TransactionExplorerFilter>>;
    transactionExplorerAllData: Ref<UnwrapRef<TransactionInfoResponse[]>>;
    transactionExplorerStateInvalid: Ref<UnwrapRef<boolean>>;
    allExplorationBasicInfos: Ref<UnwrapRef<InsightsExplorerBasicInfo[]>>;
    allExplorationBasicInfosMap: Ref<UnwrapRef<Record<string, InsightsExplorerBasicInfo>>>;
    currentExploration: Ref<UnwrapRef<InsightsExplorer>>;
    insightsExplorerListStateInvalid: Ref<UnwrapRef<boolean>>;
}



// Owns state for the explorer store.
export function useExplorerState(): ExplorerState  {
    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const exchangeRatesStore = useExchangeRatesStore();

    const currencyDisplayOrders: Record<string, number> = (() => {
        const result: Record<string, number> = {};
        let index: number = 0;

        for (const currency of keys(ALL_CURRENCIES)) {
            result[currency] = ++index;
        }

        return result;
    })();

    function buildInsightsExplorerMatchContext(exploration: InsightsExplorer, transaction: TransactionInsightDataItem): InsightsExplorerMatchContext {
        let cachedTransactionDateTime: DateTime | undefined = undefined;
        let cachedNormalizedDescription: NormalizedText | undefined = undefined;

        return {
            getTransactionDateTime(): DateTime {
                if (!cachedTransactionDateTime) {
                    let transactionTimeUtfOffset: number | undefined = undefined;

                    if (exploration.timezoneUsedForDateRange === TimezoneTypeForStatistics.TransactionTimezone.type) {
                        transactionTimeUtfOffset = transaction.utcOffset;
                    }

                    cachedTransactionDateTime = isDefined(transactionTimeUtfOffset) ? parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.time, transactionTimeUtfOffset) : parseDateTimeFromUnixTime(transaction.time);
                }

                return cachedTransactionDateTime;
            },
            getNormalizedDescription(): NormalizedText {
                if (!cachedNormalizedDescription) {
                    cachedNormalizedDescription = NormalizedText.of(transaction.comment ? transaction.comment : '');
                }

                return cachedNormalizedDescription;
            }
        };
    }

    const transactionExplorerFilter = ref<TransactionExplorerFilter>({
        dateRangeType: DEFAULT_TRANSACTION_EXPLORER_DATE_RANGE.type,
        startTime: 0,
        endTime: 0
    });

    const transactionExplorerAllData = ref<TransactionInfoResponse[]>([]);

    const transactionExplorerStateInvalid = ref<boolean>(true);

    const allExplorationBasicInfos = ref<InsightsExplorerBasicInfo[]>([]);

    const allExplorationBasicInfosMap = ref<Record<string, InsightsExplorerBasicInfo>>({});

    const currentExploration = ref<InsightsExplorer>(InsightsExplorer.createNewExplorer(generateRandomUUID()));

    const insightsExplorerListStateInvalid = ref<boolean>(true);

    return {
        settingsStore, userStore, accountsStore, transactionCategoriesStore, transactionTagsStore, exchangeRatesStore,
        currencyDisplayOrders, buildInsightsExplorerMatchContext, transactionExplorerFilter, transactionExplorerAllData,
        transactionExplorerStateInvalid, allExplorationBasicInfos, allExplorationBasicInfosMap, currentExploration,
        insightsExplorerListStateInvalid,
    };
}
