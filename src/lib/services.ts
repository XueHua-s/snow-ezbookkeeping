import { transportServices } from './services/transport.ts';
import { authorizationServices } from './services/authorization.ts';
import { usersServices } from './services/users.ts';
import { accountsServices } from './services/accounts.ts';
import { transactionsServices } from './services/transactions.ts';
import { statisticsServices } from './services/statistics.ts';
import { importServices } from './services/import.ts';
import { classificationServices } from './services/classification.ts';
import { explorersServices } from './services/explorers.ts';
import { aiServices } from './services/ai.ts';
import { iconsServices } from './services/icons.ts';
import { exchangeServices } from './services/exchange.ts';
import { urlsServices } from './services/urls.ts';

export type { ApiResponsePromise } from './services/transport.ts';
export type { AIAssistantChatStreamCallbacks } from './types/assistant.ts';

// Stable client facade; each module owns one endpoint family's wire contract.
export default {
    ...transportServices,
    ...authorizationServices,
    ...usersServices,
    ...accountsServices,
    ...transactionsServices,
    ...statisticsServices,
    ...importServices,
    ...classificationServices,
    ...explorersServices,
    ...aiServices,
    ...iconsServices,
    ...exchangeServices,
    ...urlsServices,
};
