import { itemAndIndex } from '@/core/base.ts';
import type { InsightsExplorer, InsightsExplorerBasicInfo } from '@/models/explorer.ts';
import type { useExplorerState } from './state.ts';

interface Context {
    allExplorationBasicInfos: ReturnType<typeof useExplorerState>['allExplorationBasicInfos'];
    allExplorationBasicInfosMap: ReturnType<typeof useExplorerState>['allExplorationBasicInfosMap'];
}

// Owns mutation for the explorer store.
export function useExplorerMutation(context: Context) {
    const { allExplorationBasicInfos, allExplorationBasicInfosMap } = context;

    function loadInsightsExplorerList(explorations: InsightsExplorerBasicInfo[]): void {
        allExplorationBasicInfos.value = explorations;
        allExplorationBasicInfosMap.value = {};

        for (const exploration of explorations) {
            allExplorationBasicInfosMap.value[exploration.id] = exploration;
        }
    }

    function addExplorationToInsightsExplorerList(exploration: InsightsExplorerBasicInfo): void {
        allExplorationBasicInfos.value.push(exploration);
        allExplorationBasicInfosMap.value[exploration.id] = exploration;
    }

    function updateExplorationInInsightsExplorerList(currentExploration: InsightsExplorerBasicInfo): void {
        for (const [explorer, index] of itemAndIndex(allExplorationBasicInfos.value)) {
            if (explorer.id === currentExploration.id) {
                allExplorationBasicInfos.value.splice(index, 1, currentExploration);
                break;
            }
        }

        allExplorationBasicInfosMap.value[currentExploration.id] = currentExploration;
    }

    function updateExplorerDisplayOrderInInsightsExplorerList({ from, to }: { from: number, to: number }): void {
        allExplorationBasicInfos.value.splice(to, 0, allExplorationBasicInfos.value.splice(from, 1)[0] as InsightsExplorer);
    }

    function updateExplorationVisibilityInInsightsExplorerList({ explorationId, hidden }: { explorationId: string, hidden: boolean }): void {
        if (allExplorationBasicInfosMap.value[explorationId]) {
            allExplorationBasicInfosMap.value[explorationId]!.hidden = hidden;
        }
    }

    function removeExplorationFromInsightsExplorerList(currentExploration: InsightsExplorerBasicInfo): void {
        for (const [insightsExplorer, index] of itemAndIndex(allExplorationBasicInfos.value)) {
            if (insightsExplorer.id === currentExploration.id) {
                allExplorationBasicInfos.value.splice(index, 1);
                break;
            }
        }

        if (allExplorationBasicInfosMap.value[currentExploration.id]) {
            delete allExplorationBasicInfosMap.value[currentExploration.id];
        }
    }

    return {
        loadInsightsExplorerList, addExplorationToInsightsExplorerList, updateExplorationInInsightsExplorerList,
        updateExplorerDisplayOrderInInsightsExplorerList, updateExplorationVisibilityInInsightsExplorerList,
        removeExplorationFromInsightsExplorerList,
    };
}
