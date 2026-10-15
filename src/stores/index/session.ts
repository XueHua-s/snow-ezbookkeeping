import services from '@/lib/services.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';
import { useExplorersStore } from '@/stores/explorer.ts';
import { useOverviewStore } from '@/stores/overview.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useStatisticsStore } from '@/stores/statistics.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useTransactionTemplatesStore } from '@/stores/transactionTemplate.ts';
import { useUserStore } from '@/stores/user.ts';
import { useUserCustomIconsStore } from '@/stores/userCustomIcon.ts';
import type { Ref, UnwrapRef } from 'vue';
import { ref } from 'vue';
interface IndexState {
    settingsStore: ReturnType<typeof useSettingsStore>;
    userStore: ReturnType<typeof useUserStore>;
    accountsStore: ReturnType<typeof useAccountsStore>;
    transactionCategoriesStore: ReturnType<typeof useTransactionCategoriesStore>;
    transactionTagsStore: ReturnType<typeof useTransactionTagsStore>;
    overviewStore: ReturnType<typeof useOverviewStore>;
    statisticsStore: ReturnType<typeof useStatisticsStore>;
    exchangeRatesStore: ReturnType<typeof useExchangeRatesStore>;
    userCustomIconsStore: ReturnType<typeof useUserCustomIconsStore>;
    currentNotification: Ref<UnwrapRef<string | null>>;
    resetAllStates: (resetUserInfoAndSettings: boolean) => void;
    setNotificationContent: (content: string | null) => void;
    generateOAuth2LoginUrl: (platform: 'mobile' | 'desktop', clientSessionId: string) => string;
    generateOAuth2LinkUrl: (platform: 'mobile' | 'desktop', clientSessionId: string) => string;
}



// Owns session for the index store.
export function useIndexSession(): IndexState  {
    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const accountsStore = useAccountsStore();

    const transactionCategoriesStore = useTransactionCategoriesStore();

    const transactionTagsStore = useTransactionTagsStore();

    const transactionTemplatesStore = useTransactionTemplatesStore();

    const transactionsStore = useTransactionsStore();

    const overviewStore = useOverviewStore();

    const statisticsStore = useStatisticsStore();

    const explorersStore = useExplorersStore();

    const exchangeRatesStore = useExchangeRatesStore();

    const userCustomIconsStore = useUserCustomIconsStore();

    const currentNotification = ref<string | null>(null);

    function resetAllStates(resetUserInfoAndSettings: boolean): void {
        if (resetUserInfoAndSettings) {
            exchangeRatesStore.resetLatestExchangeRates();
        }

        setNotificationContent(null);

        userCustomIconsStore.resetCustomIcons();
        explorersStore.resetTransactionExplorers();
        statisticsStore.resetTransactionStatistics();
        overviewStore.resetTransactionOverview();
        transactionsStore.resetTransactions();
        transactionTagsStore.resetTransactionTags();
        transactionCategoriesStore.resetTransactionCategories();
        transactionTemplatesStore.resetTransactionTemplates();
        accountsStore.resetAccounts();

        if (resetUserInfoAndSettings) {
            userStore.resetUserBasicInfo();
        }
    }

    function setNotificationContent(content: string | null): void {
        currentNotification.value = content;
    }

    function generateOAuth2LoginUrl(platform: 'mobile' | 'desktop', clientSessionId: string): string {
        return services.generateOAuth2LoginUrl(platform, clientSessionId);
    }

    function generateOAuth2LinkUrl(platform: 'mobile' | 'desktop', clientSessionId: string): string {
        return services.generateOAuth2LinkUrl(platform, clientSessionId);
    }

    return {
        settingsStore, userStore, accountsStore, transactionCategoriesStore, transactionTagsStore, overviewStore, statisticsStore,
        exchangeRatesStore, userCustomIconsStore, currentNotification, resetAllStates, setNotificationContent,
        generateOAuth2LoginUrl, generateOAuth2LinkUrl,
    };
}
