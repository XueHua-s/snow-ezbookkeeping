import { KnownErrorCode } from '@/consts/api.ts';
import { generateRandomUUID } from '@/lib/misc.ts';
import { InsightsExplorer } from '@/models/explorer.ts';
import type { useExplorerPageState } from './state.ts';

interface Context {
    currentExploration: ReturnType<typeof useExplorerPageState>['currentExploration'];
    explorersStore: ReturnType<typeof useExplorerPageState>['explorersStore'];
    initExploration: ReturnType<typeof useExplorerPageState>['initExploration'];
    router: ReturnType<typeof useExplorerPageState>['router'];
    getFilterLinkUrl: ReturnType<typeof useExplorerPageState>['getFilterLinkUrl'];
    loading: ReturnType<typeof useExplorerPageState>['loading'];
    snackbar: ReturnType<typeof useExplorerPageState>['snackbar'];
    explorerChangeDisplayOrderDialog: ReturnType<typeof useExplorerPageState>['explorerChangeDisplayOrderDialog'];
    renameDialog: ReturnType<typeof useExplorerPageState>['renameDialog'];
    tt: ReturnType<typeof useExplorerPageState>['tt'];
    updating: ReturnType<typeof useExplorerPageState>['updating'];
    clientSessionId: ReturnType<typeof useExplorerPageState>['clientSessionId'];
    confirmDialog: ReturnType<typeof useExplorerPageState>['confirmDialog'];
}

export function useExplorerPageManagement(context: Context) {
    const {
        currentExploration, explorersStore, initExploration, router, getFilterLinkUrl, loading, snackbar,
        explorerChangeDisplayOrderDialog, renameDialog, tt, updating, clientSessionId, confirmDialog,
    } = context;

    function createNewExploration(): void {
        if (!currentExploration.value.id) {
            return;
        }

        explorersStore.updateCurrentExploration(InsightsExplorer.createNewExplorer(generateRandomUUID()));
        initExploration.value = InsightsExplorer.of(explorersStore.currentExploration);
        router.push(getFilterLinkUrl());
    }

    function loadExploration(explorationId: string, force?: boolean, init?: boolean): Promise<unknown> | null {
        if (!force && currentExploration.value && currentExploration.value.id === explorationId) {
            return null;
        }

        if (!init) {
            loading.value = true;
        }

        return explorersStore.getExploration({
            explorationId: explorationId
        }).then(exploration => {
            explorersStore.updateCurrentExploration(exploration);
            initExploration.value = InsightsExplorer.of(exploration);

            if (!init) {
                loading.value = false;
            }

            router.push(getFilterLinkUrl());
        }).catch(error => {
            if (!init) {
                loading.value = false;
            }

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function showChangeExplorerDisplayOrderDialog(): void {
        explorerChangeDisplayOrderDialog.value?.open().then(() => {
            for (const exploration of explorersStore.allExplorationBasicInfos) {
                if (exploration.id === currentExploration.value.id) {
                    if (initExploration.value) {
                        initExploration.value.hidden = exploration.hidden;
                    }

                    currentExploration.value.hidden = exploration.hidden;
                }
            }

            if (explorersStore.insightsExplorerListStateInvalid) {
                loading.value = true;

                explorersStore.loadAllExplorationBasicInfos({
                    force: false
                }).then(() => {
                    loading.value = false;
                }).catch(() => {
                    loading.value = false;
                });
            }
        });
    }

    function saveExploration(saveAs?: boolean): void {
        if (saveAs || !currentExploration.value.name) {
            renameDialog.value?.open(currentExploration.value.name || '', tt('Set Exploration Name')).then((newName: string) => {
                currentExploration.value.name = newName;
                doSaveExploration(saveAs);
            })
        } else {
            doSaveExploration(saveAs);
        }
    }

    function doSaveExploration(saveAs?: boolean): Promise<unknown> {
        const oldExplorationId = currentExploration.value.id;

        updating.value = true;

        return explorersStore.saveExploration({
            exploration: currentExploration.value,
            saveAs: saveAs,
            clientSessionId: clientSessionId.value
        }).then(newExploration => {
            updating.value = false;
            clientSessionId.value = generateRandomUUID();
            explorersStore.updateCurrentExploration(newExploration);
            initExploration.value = InsightsExplorer.of(newExploration);

            if (oldExplorationId !== newExploration.id) {
                router.push(getFilterLinkUrl());
            }
        }).catch(error => {
            updating.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);

                if (error.error && error.error.errorCode === KnownErrorCode.NothingWillBeUpdated) {
                    initExploration.value = InsightsExplorer.of(currentExploration.value);
                }
            }
        });
    }

    function restoreExploration(): void {
        if (!currentExploration.value.id) {
            return;
        }

        confirmDialog.value?.open('Are you sure you want to restore to last saved state? All unsaved changes will be lost.').then(() => {
            loadExploration(currentExploration.value.id, true);
        });
    }

    function setExplorationName(): void {
        renameDialog.value?.open(currentExploration.value.name || '').then((newName: string) => {
            currentExploration.value.name = newName;
        });
    }

    function hideExploration(hidden: boolean): void {
        updating.value = true;

        explorersStore.hideExploration({
            exploration: currentExploration.value,
            hidden: hidden
        }).then(() => {
            updating.value = false;

            if (initExploration.value) {
                initExploration.value.hidden = hidden;
            }

            currentExploration.value.hidden = hidden;
        }).catch(error => {
            updating.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function removeExploration(): void {
        if (!currentExploration.value.id) {
            return;
        }

        confirmDialog.value?.open('Are you sure you want to delete this exploration?').then(() => {
            updating.value = true;

            explorersStore.deleteExploration({
                exploration: currentExploration.value
            }).then(() => {
                updating.value = false;
                createNewExploration();
            }).catch(error => {
                updating.value = false;

                if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        });
    }

    return {
        createNewExploration, loadExploration, showChangeExplorerDisplayOrderDialog, saveExploration, restoreExploration,
        setExplorationName, hideExploration, removeExploration,
    };
}
