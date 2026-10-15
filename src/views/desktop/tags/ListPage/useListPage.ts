import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import RenameDialog from '@/components/desktop/RenameDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { DEFAULT_TAG_GROUP_ID } from '@/consts/tag.ts';
import { getAvailableTagCount } from '@/lib/tag.ts';
import { useI18n } from '@/locales/helpers.ts';
import { TransactionTag } from '@/models/transaction_tag.ts';
import { TransactionTagGroup } from '@/models/transaction_tag_group.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useTagListPageBase } from '@/views/base/tags/TagListPageBase.ts';
import TagGroupChangeDisplayOrderDialog from '@/views/desktop/tags/dialog/TagGroupChangeDisplayOrderDialog.vue';
import { computed, nextTick, ref, useTemplateRef, watch } from 'vue';
import { useDisplay } from 'vuetify';

export type TagGroupChangeDisplayOrderDialogType = InstanceType<typeof TagGroupChangeDisplayOrderDialog>;

export type RenameDialogType = InstanceType<typeof RenameDialog>;

export type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;

export type SnackBarType = InstanceType<typeof SnackBar>;

// Coordinates ListPage state and effects; called once from the component setup.
export function useListPage() {
    const { lgAndUp } = useDisplay();

    const { tt, formatNumberToLocalizedNumerals } = useI18n();

    const {
        activeTagGroupId,
        newTag,
        editingTag,
        loading,
        showHidden,
        displayOrderModified,
        allTagGroupsWithDefault,
        tags,
        noAvailableTag,
        hasEditingTag,
        isTagModified,
        switchTagGroup,
        createNewTag,
        edit
    } = useTagListPageBase();

    const transactionTagsStore = useTransactionTagsStore();

    const newTagRow = useTemplateRef<HTMLElement>('newTagRow');

    const tagGroupChangeDisplayOrderDialog = useTemplateRef<TagGroupChangeDisplayOrderDialogType>('tagGroupChangeDisplayOrderDialog');

    const renameDialog = useTemplateRef<RenameDialogType>('renameDialog');

    const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');

    const snackbar = useTemplateRef<SnackBarType>('snackbar');

    const updating = ref<boolean>(false);

    const activeTab = ref<string>('tagListPage');

    const alwaysShowNav = ref<boolean>(lgAndUp.value);

    const showNav = ref<boolean>(lgAndUp.value);

    const hoveredTagId = ref<string>('');

    const tagUpdating = ref<Record<string, boolean>>({});

    const tagHiding = ref<Record<string, boolean>>({});

    const tagMoving = ref<Record<string, boolean>>({});

    const tagRemoving = ref<Record<string, boolean>>({});

    const currentMovingTag = ref<TransactionTag | null>(null);

    const currentMoveTargetGroupId = ref<string>(DEFAULT_TAG_GROUP_ID);

    const showTagMoveToDialog = ref<boolean>(false);

    const totalAvailableTagsCount = computed<number>(() => transactionTagsStore.allAvailableTagsCount);

    const displayTotalAvailableTagsCount = computed<string>(() => formatNumberToLocalizedNumerals(transactionTagsStore.allAvailableTagsCount));

    const availableTagCount = computed<number>(() => getAvailableTagCount(tags.value, showHidden.value));

    function reload(): void {
        if (hasEditingTag.value) {
            return;
        }

        loading.value = true;

        transactionTagsStore.loadAllTags({
            force: true
        }).then(() => {
            loading.value = false;
            displayOrderModified.value = false;

            snackbar.value?.showMessage('Tag list has been updated');
        }).catch(error => {
            loading.value = false;

            if (error && error.isUpToDate) {
                displayOrderModified.value = false;
            }

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function add(): void {
        createNewTag();

        nextTick(() => {
            newTagRow.value?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        });
    }

    function addTagGroup(): void {
        renameDialog.value?.open('', tt('New Tag Group Name')).then((newName: string) => {
            updating.value = true;

            transactionTagsStore.saveTagGroup({
                tagGroup: TransactionTagGroup.createNewTagGroup(newName)
            }).then(tagGroup => {
                updating.value = false;
                activeTagGroupId.value = tagGroup.id;
            }).catch(error => {
                updating.value = false;

                if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        });
    }

    function renameTagGroup(): void {
        const tagGroup = transactionTagsStore.allTransactionTagGroupsMap[activeTagGroupId.value];

        if (!tagGroup) {
            snackbar.value?.showMessage('Unable to rename this tag group');
            return;
        }

        renameDialog.value?.open(tagGroup.name || '').then((newName: string) => {
            updating.value = true;

            const newTagGroup = tagGroup.clone();
            newTagGroup.name = newName;

            transactionTagsStore.saveTagGroup({
                tagGroup: newTagGroup
            }).then(() => {
                updating.value = false;
            }).catch(error => {
                updating.value = false;

                if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        });
    }

    function showChangeGroupDisplayOrderDialog(): void {
        tagGroupChangeDisplayOrderDialog.value?.open().then(() => {
            if (transactionTagsStore.transactionTagGroupListStateInvalid) {
                loading.value = true;

                transactionTagsStore.loadAllTagGroups({
                    force: false
                }).then(() => {
                    loading.value = false;
                }).catch(() => {
                    loading.value = false;
                });
            }
        });
    }

    function removeTagGroup(): void {
        const tagGroup = transactionTagsStore.allTransactionTagGroupsMap[activeTagGroupId.value];

        if (!tagGroup) {
            snackbar.value?.showMessage('Unable to delete this tag group');
            return;
        }

        const currentTagGroupIndex = allTagGroupsWithDefault.value.findIndex(group => group.id === tagGroup.id);

        confirmDialog.value?.open('Are you sure you want to delete this tag group?').then(() => {
            updating.value = true;

            transactionTagsStore.deleteTagGroup({
                tagGroup: tagGroup
            }).then(() => {
                updating.value = false;

                if (allTagGroupsWithDefault.value[currentTagGroupIndex]) {
                    const newActiveTagGroup = allTagGroupsWithDefault.value[currentTagGroupIndex];
                    activeTagGroupId.value = newActiveTagGroup ? newActiveTagGroup.id : DEFAULT_TAG_GROUP_ID;
                } else if (allTagGroupsWithDefault.value[currentTagGroupIndex - 1]) {
                    const newActiveTagGroup = allTagGroupsWithDefault.value[currentTagGroupIndex - 1];
                    activeTagGroupId.value = newActiveTagGroup ? newActiveTagGroup.id : DEFAULT_TAG_GROUP_ID;
                } else {
                    activeTagGroupId.value = DEFAULT_TAG_GROUP_ID;
                }
            }).catch(error => {
                updating.value = false;

                if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        });
    }

    function showMoveTagDialog(tag: TransactionTag): void {
        currentMovingTag.value = tag;
        currentMoveTargetGroupId.value = tag.groupId || DEFAULT_TAG_GROUP_ID;
        showTagMoveToDialog.value = true;
    }

    function moveTagToGroup(tag: TransactionTag | null, targetTagGroupId: string): void {
        if (!tag) {
            snackbar.value?.showMessage('Unable to move tag');
            return;
        }

        updating.value = true;
        tagMoving.value[tag.id] = true;

        const newTag = tag.clone();
        newTag.groupId = targetTagGroupId;

        transactionTagsStore.saveTag({
            tag: newTag
        }).then(() => {
            updating.value = false;
            tagMoving.value[tag.id] = false;
        }).catch(error => {
            updating.value = false;
            tagMoving.value[tag.id] = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });

        showTagMoveToDialog.value = false;
    }

    function save(tag: TransactionTag): void {
        updating.value = true;
        tagUpdating.value[tag.id || ''] = true;

        transactionTagsStore.saveTag({
            tag: tag
        }).then(() => {
            updating.value = false;
            tagUpdating.value[tag.id || ''] = false;

            if (tag.id) {
                editingTag.value.id = '';
                editingTag.value.name = '';
            } else {
                newTag.value = null;
            }
        }).catch(error => {
            updating.value = false;
            tagUpdating.value[tag.id || ''] = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function cancelSave(tag: TransactionTag): void {
        if (tag.id) {
            editingTag.value.id = '';
            editingTag.value.name = '';
        } else {
            newTag.value = null;
        }
    }

    function sortByName(desc: boolean): void {
        const changed = transactionTagsStore.sortTagDisplayOrderByTagName(activeTagGroupId.value, desc);

        if (changed) {
            displayOrderModified.value = true;
        }
    }

    function saveSortResult(): void {
        if (!displayOrderModified.value) {
            return;
        }

        loading.value = true;

        transactionTagsStore.updateTagDisplayOrders(activeTagGroupId.value).then(() => {
            loading.value = false;
            displayOrderModified.value = false;
        }).catch(error => {
            loading.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function hide(tag: TransactionTag, hidden: boolean): void {
        updating.value = true;
        tagHiding.value[tag.id] = true;

        transactionTagsStore.hideTag({
            tag: tag,
            hidden: hidden
        }).then(() => {
            updating.value = false;
            tagHiding.value[tag.id] = false;
        }).catch(error => {
            updating.value = false;
            tagHiding.value[tag.id] = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    }

    function remove(tag: TransactionTag): void {
        confirmDialog.value?.open('Are you sure you want to delete this tag?').then(() => {
            updating.value = true;
            tagRemoving.value[tag.id] = true;

            transactionTagsStore.deleteTag({
                tag: tag
            }).then(() => {
                updating.value = false;
                tagRemoving.value[tag.id] = false;
            }).catch(error => {
                updating.value = false;
                tagRemoving.value[tag.id] = false;

                if (!error.processed) {
                    snackbar.value?.showError(error);
                }
            });
        });
    }

    function onMove(event: { moved: { element: { id: string }; oldIndex: number; newIndex: number } }): void {
        if (!event || !event.moved) {
            return;
        }

        const moveEvent = event.moved;

        if (!moveEvent.element || !moveEvent.element.id) {
            snackbar.value?.showMessage('Unable to move tag');
            return;
        }

        transactionTagsStore.changeTagDisplayOrder({
            tagId: moveEvent.element.id,
            from: moveEvent.oldIndex,
            to: moveEvent.newIndex
        }).then(() => {
            displayOrderModified.value = true;
        }).catch(error => {
            snackbar.value?.showError(error);
        });
    }

    transactionTagsStore.loadAllTags({
        force: false
    }).then(() => {
        loading.value = false;
    }).catch(error => {
        loading.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });

    watch(lgAndUp, (newValue) => {
        alwaysShowNav.value = newValue;

        if (!showNav.value) {
            showNav.value = newValue;
        }
    });

    return {
        tt, activeTagGroupId, newTag, editingTag, loading, showHidden, displayOrderModified, allTagGroupsWithDefault, tags,
        noAvailableTag, hasEditingTag, isTagModified, switchTagGroup, edit, newTagRow, tagGroupChangeDisplayOrderDialog,
        renameDialog, confirmDialog, snackbar, updating, activeTab, alwaysShowNav, showNav, hoveredTagId, tagUpdating, tagHiding,
        tagMoving, tagRemoving, currentMovingTag, showTagMoveToDialog, totalAvailableTagsCount, displayTotalAvailableTagsCount,
        availableTagCount, reload, add, addTagGroup, renameTagGroup, showChangeGroupDisplayOrderDialog, removeTagGroup,
        showMoveTagDialog, moveTagToGroup, save, cancelSave, sortByName, saveSortResult, hide, remove, onMove,
    };
}
