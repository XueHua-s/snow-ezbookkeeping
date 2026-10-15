import { isArray, isObject, isTextualUUID } from '@/lib/common.ts';
import logger from '@/lib/logger.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import { InsightsExplorer, TransactionExplorerQuery } from '@/models/explorer.ts';
import type { useExplorerPageState } from './state.ts';

interface Context {
    snackbar: ReturnType<typeof useExplorerPageState>['snackbar'];
    explorersStore: ReturnType<typeof useExplorerPageState>['explorersStore'];
    initExploration: ReturnType<typeof useExplorerPageState>['initExploration'];
    currentExploration: ReturnType<typeof useExplorerPageState>['currentExploration'];
}

export function useExplorerPageImport(context: Context) {
    const { snackbar, explorersStore, initExploration, currentExploration } = context;

    function onImportQueries(data: string): boolean {
        try {
            const queryItems = JSON.parse(data);

            if (!isArray(queryItems)) {
                snackbar.value?.showError('Queries import failed. Please make sure the queries are valid and try again.');
                return false;
            }

            if (!queryItems || queryItems.length < 1) {
                snackbar.value?.showError('No valid queries found in the input. Please make sure the queries are valid and try again.');
                return false;
            }

            const queries: TransactionExplorerQuery[] = [];
            const queryIds: Record<string, boolean> = {};

            for (const queryItem of queryItems) {
                if (!isObject(queryItem)) {
                    snackbar.value?.showError('Queries import failed. Please make sure the queries are valid and try again.');
                    return false;
                }

                let originalId: string = '';

                if (!('id' in queryItem) || !queryItem['id']) {
                    (queryItem as Record<string, unknown>)['id'] = generateRandomUUID();
                } else {
                    const queryId = queryItem['id'];

                    if (!isTextualUUID(queryId)) {
                        snackbar.value?.showMessage('format.misc.queryIdInvalidTip', { id: queryId });
                        return false;
                    }

                    originalId = queryId;
                }

                const query = TransactionExplorerQuery.parse(queryItem);

                if (!query) {
                    snackbar.value?.showMessage('format.misc.queryInvalidTip', { id: originalId });
                    return false;
                }

                if (queryIds[query.id]) {
                    snackbar.value?.showMessage('format.misc.queryIdDuplicatedTip', { id: query.id });
                    return false;
                }

                queries.push(query);
                queryIds[query.id] = true;
            }

            if (!queries || queries.length < 1) {
                snackbar.value?.showError('No valid queries found in the input. Please make sure the queries are valid and try again.');
                return false;
            }

            explorersStore.currentExploration.queries.length = 0;
            explorersStore.currentExploration.queries.push(...queries);
            initExploration.value = InsightsExplorer.of(currentExploration.value);
            return true;
        } catch (error) {
            logger.error('Failed to import queries', error);
            snackbar.value?.showError('Queries import failed. Please make sure the queries are valid and try again.');
            return false;
        }
    }

    return { onImportQueries };
}
