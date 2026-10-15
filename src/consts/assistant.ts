import type { AIAssistantTool } from '@/models/large_language_model.ts';

export const ASSISTANT_MAX_FRAME_LENGTH = 1024 * 1024;
export const ASSISTANT_MAX_HISTORY_CHARACTERS = 2048;
export const ASSISTANT_MAX_HISTORY_MESSAGES = 12;
export const ASSISTANT_MAX_REFERENCES = 8;
export const ASSISTANT_PAGEABLE_TOOLS: readonly string[] = ['list_accounts', 'list_categories', 'list_tags', 'query_transactions', 'aggregate_transactions', 'get_account_balances'];

export const ASSISTANT_STATE_LABELS = {
    requesting: 'Working on your question', completed: 'Answer completed', incomplete: 'Answer incomplete',
    failed: 'Assistant request failed', canceled: 'Assistant request canceled'
} as const;
export const ASSISTANT_PROGRESS_LABELS = { querying: 'Reading your ledger', aggregating: 'Calculating matching transactions', answering: 'Preparing your answer' } as const;
export const ASSISTANT_TOOL_LABELS: Record<AIAssistantTool, string> = {
    list_accounts: 'Accounts', list_categories: 'Category', list_tags: 'Tags', query_transactions: 'Transactions',
    get_transaction: 'Transaction Details', aggregate_transactions: 'Statistics', get_account_balances: 'Account Balances'
};

export const ASSISTANT_TRANSACTION_TYPE_LABELS: Readonly<Record<string, string>> = {
    income: 'Income', expense: 'Expense', transfer: 'Transfer', balance_modification: 'Modify Balance'
};
export const ASSISTANT_TRANSFER_SIDE_LABELS = {
    source: 'Transfer source side', destination: 'Transfer destination side', either: 'Either transfer side'
} as const;
