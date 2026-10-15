import { type BeforeResolveFunction, itemAndIndex } from '@/core/base.ts';
import logger from '@/lib/logger.ts';
import services,{ type ApiResponsePromise } from '@/lib/services.ts';
import {
    type InsightsExplorerInfoResponse,
    type InsightsExplorerNewDisplayOrderRequest,
    InsightsExplorer,
    InsightsExplorerBasicInfo
} from '@/models/explorer.ts';
import type { useExplorerFilter } from './filter.ts';
import type { useExplorerLoading } from './loading.ts';
import type { useExplorerMutation } from './mutation.ts';
import type { useExplorerState } from './state.ts';

interface Context {
    addExplorationToInsightsExplorerList: ReturnType<typeof useExplorerMutation>['addExplorationToInsightsExplorerList'];
    updateExplorationInInsightsExplorerList: ReturnType<typeof useExplorerMutation>['updateExplorationInInsightsExplorerList'];
    allExplorationBasicInfos: ReturnType<typeof useExplorerState>['allExplorationBasicInfos'];
    insightsExplorerListStateInvalid: ReturnType<typeof useExplorerState>['insightsExplorerListStateInvalid'];
    updateInsightsExplorerListInvalidState: ReturnType<typeof useExplorerFilter>['updateInsightsExplorerListInvalidState'];
    updateExplorerDisplayOrderInInsightsExplorerList: ReturnType<typeof useExplorerMutation>['updateExplorerDisplayOrderInInsightsExplorerList'];
    loadAllExplorationBasicInfos: ReturnType<typeof useExplorerLoading>['loadAllExplorationBasicInfos'];
    updateExplorationVisibilityInInsightsExplorerList: ReturnType<typeof useExplorerMutation>['updateExplorationVisibilityInInsightsExplorerList'];
    removeExplorationFromInsightsExplorerList: ReturnType<typeof useExplorerMutation>['removeExplorationFromInsightsExplorerList'];
}

// Owns management for the explorer store.
export function useExplorerManagement(context: Context) {
    const {
        addExplorationToInsightsExplorerList, updateExplorationInInsightsExplorerList, allExplorationBasicInfos,
        insightsExplorerListStateInvalid, updateInsightsExplorerListInvalidState,
        updateExplorerDisplayOrderInInsightsExplorerList, loadAllExplorationBasicInfos,
        updateExplorationVisibilityInInsightsExplorerList, removeExplorationFromInsightsExplorerList,
    } = context;

    function saveExploration({ exploration, saveAs, clientSessionId }: { exploration: InsightsExplorer, saveAs?: boolean, clientSessionId: string }): Promise<InsightsExplorer> {
        return new Promise((resolve, reject) => {
            let promise: ApiResponsePromise<InsightsExplorerInfoResponse>;

            if (!exploration.id || saveAs) {
                promise = services.addExploration(exploration.toCreateRequest(clientSessionId));
            } else {
                promise = services.modifyExploration(exploration.toModifyRequest());
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (!exploration.id) {
                        reject({ message: 'Unable to add exploration' });
                    } else {
                        reject({ message: 'Unable to save exploration' });
                    }
                    return;
                }

                const explorationBasicInfo = InsightsExplorerBasicInfo.of(data.result);

                if (!exploration.id || saveAs) {
                    addExplorationToInsightsExplorerList(explorationBasicInfo);
                } else {
                    updateExplorationInInsightsExplorerList(explorationBasicInfo);
                }

                resolve(InsightsExplorer.of(data.result));
            }).catch(error => {
                logger.error('failed to save exploration', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!exploration.id) {
                        reject({ message: 'Unable to add exploration' });
                    } else {
                        reject({ message: 'Unable to save exploration' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function changeExplorationDisplayOrder({ explorationId, from, to }: { explorationId: string, from: number, to: number }): Promise<void> {
        return new Promise((resolve, reject) => {
            let currentExploration: InsightsExplorerBasicInfo | null = null;

            for (const exploration of allExplorationBasicInfos.value) {
                if (exploration.id === explorationId) {
                    currentExploration = exploration;
                    break;
                }
            }

            if (!currentExploration || !allExplorationBasicInfos.value[to]) {
                reject({ message: 'Unable to move exploration' });
                return;
            }

            if (!insightsExplorerListStateInvalid.value) {
                updateInsightsExplorerListInvalidState(true);
            }

            updateExplorerDisplayOrderInInsightsExplorerList({ from, to });

            resolve();
        });
    }

    function updateExplorationDisplayOrders(): Promise<boolean> {
        const newDisplayOrders: InsightsExplorerNewDisplayOrderRequest[] = [];

        for (const [exploration, index] of itemAndIndex(allExplorationBasicInfos.value)) {
            newDisplayOrders.push({
                id: exploration.id,
                displayOrder: index + 1
            });
        }

        return new Promise((resolve, reject) => {
            services.moveExploration({
                newDisplayOrders: newDisplayOrders
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to move exploration' });
                    return;
                }

                loadAllExplorationBasicInfos({ force: false }).finally(() => {
                    if (insightsExplorerListStateInvalid.value) {
                        updateInsightsExplorerListInvalidState(false);
                    }

                    resolve(data.result);
                });
            }).catch(error => {
                logger.error('failed to save explorations display order', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to move exploration' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function hideExploration({ exploration, hidden }: { exploration: InsightsExplorer | InsightsExplorerBasicInfo, hidden: boolean }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.hideExploration({
                id: exploration.id,
                hidden: hidden
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this exploration' });
                    } else {
                        reject({ message: 'Unable to unhide this exploration' });
                    }
                    return;
                }

                exploration.hidden = hidden;
                updateExplorationVisibilityInInsightsExplorerList({ explorationId: exploration.id, hidden });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to change exploration visibility', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this exploration' });
                    } else {
                        reject({ message: 'Unable to unhide this exploration' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function deleteExploration({ exploration, beforeResolve }: { exploration: InsightsExplorer, beforeResolve?: BeforeResolveFunction }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteExploration({
                id: exploration.id
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this exploration' });
                    return;
                }

                if (beforeResolve) {
                    beforeResolve(() => {
                        removeExplorationFromInsightsExplorerList(exploration);
                    });
                } else {
                    removeExplorationFromInsightsExplorerList(exploration);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete exploration', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this exploration' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return { saveExploration, changeExplorationDisplayOrder, updateExplorationDisplayOrders, hideExploration, deleteExploration };
}
