import type { SetupProps } from '@/lib/ui/props.ts';
import { DEFAULT_TAG_GROUP_ID } from '@/consts/tag.ts';
import { TextDirection } from '@/core/text.ts';
import { getFirstShowingId, getLastShowingId } from '@/lib/tag.ts';
import { scrollToSelectedItem } from '@/lib/ui/common.ts';
import { type Framework7Dom, hideLoading, onSwipeoutDeleted, showLoading, useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useI18n } from '@/locales/helpers.ts';
import { TransactionTag } from '@/models/transaction_tag.ts';
import { TransactionTagGroup } from '@/models/transaction_tag_group.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useTagListPageBase } from '@/views/base/tags/TagListPageBase.ts';
import type { Router } from 'framework7/types';
import { computed, nextTick, ref, useTemplateRef } from 'vue';



// Coordinates ListPage state and effects; called once from the component setup.
export function useListPage(props: SetupProps<{
    f7router: Router.Router;
}>) {
    const { tt, getCurrentLanguageTextDirection } = useI18n();

    const { showAlert, showConfirm, showPrompt, showToast, routeBackOnError } = useI18nUIComponents();

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

    const newTagItem = useTemplateRef<{ $el: HTMLElement }>('newTagItem');

    const loadingError = ref<unknown | null>(null);

    const sortable = ref<boolean>(false);

    const moveToTagGroupId = ref<string | undefined>(undefined);

    const tagToMove = ref<TransactionTag | null>(null);

    const tagToDelete = ref<TransactionTag | null>(null);

    const showMoveTagPopup = ref<boolean>(false);

    const showMoreActionSheet = ref<boolean>(false);

    const showDeleteActionSheet = ref<boolean>(false);

    const displayOrderSaving = ref<boolean>(false);

    const textDirection = computed<TextDirection>(() => getCurrentLanguageTextDirection());

    const firstShowingId = computed<string | null>(() => getFirstShowingId(tags.value, showHidden.value));

    const lastShowingId = computed<string | null>(() => getLastShowingId(tags.value, showHidden.value));

    const displayTagGroupName = computed<string>(() => {
        const tagGroup = transactionTagsStore.allTransactionTagGroupsMap[activeTagGroupId.value];
        return tagGroup ? tagGroup.name : tt('Default Group');
    });

    function getTagDomId(tag: TransactionTag): string {
        return 'tag_' + tag.id;
    }

    function parseTagIdFromDomId(domId: string): string | null {
        if (!domId || domId.indexOf('tag_') !== 0) {
            return null;
        }

        return domId.substring(4); // tag_
    }

    function init(): void {
        loading.value = true;

        transactionTagsStore.loadAllTags({
            force: false
        }).then(() => {
            loading.value = false;
        }).catch(error => {
            if (error.processed) {
                loading.value = false;
            } else {
                loadingError.value = error;
                showToast(error.message || error);
            }
        });
    }

    function reload(done?: () => void): void {
        if (sortable.value || hasEditingTag.value) {
            done?.();
            return;
        }

        const force = !!done;

        transactionTagsStore.loadAllTags({
            force: force
        }).then(() => {
            done?.();

            if (force) {
                showToast('Tag list has been updated');
            }
        }).catch(error => {
            done?.();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function add(): void {
        createNewTag();

        nextTick(() => {
            const el = newTagItem.value?.$el as HTMLElement | undefined;

            if (el) {
                el.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
            }
        });
    }

    function save(tag: TransactionTag): void {
        showLoading();

        transactionTagsStore.saveTag({
            tag: tag
        }).then(() => {
            hideLoading();

            if (tag.id) {
                editingTag.value.id = '';
                editingTag.value.name = '';
            } else {
                newTag.value = null;
            }
        }).catch(error => {
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
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

    function hide(tag: TransactionTag, hidden: boolean): void {
        showLoading();

        transactionTagsStore.hideTag({
            tag: tag,
            hidden: hidden
        }).then(() => {
            hideLoading();
        }).catch(error => {
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function moveTagToGroup(tag: TransactionTag | null, targetTagGroupId?: string): void {
        if (!tag) {
            showAlert('An error occurred');
            return;
        }

        if (!targetTagGroupId) {
            moveToTagGroupId.value = undefined;
            tagToMove.value = tag;
            showMoveTagPopup.value = true;
            return;
        }

        showMoveTagPopup.value = false;
        tagToMove.value = null;
        moveToTagGroupId.value = undefined;
        showLoading();

        const newTag = tag.clone();
        newTag.groupId = targetTagGroupId;

        transactionTagsStore.saveTag({
            tag: newTag,
            beforeResolve: (done) => {
                onSwipeoutDeleted(getTagDomId(tag), done);
            }
        }).then(() => {
            hideLoading();
        }).catch(error => {
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function updateTagGroupSelected(e: Event): void {
        const target = e.target as HTMLInputElement;

        if (target.checked) {
            moveToTagGroupId.value = target.value;
        } else {
            moveToTagGroupId.value = undefined;
        }
    }

    function remove(tag: TransactionTag | null, confirm: boolean): void {
        if (!tag) {
            showAlert('An error occurred');
            return;
        }

        if (!confirm) {
            tagToDelete.value = tag;
            showDeleteActionSheet.value = true;
            return;
        }

        showDeleteActionSheet.value = false;
        tagToDelete.value = null;
        showLoading();

        transactionTagsStore.deleteTag({
            tag: tag,
            beforeResolve: (done) => {
                onSwipeoutDeleted(getTagDomId(tag), done);
            }
        }).then(() => {
            hideLoading();
        }).catch(error => {
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function sortByName(desc: boolean): void {
        showHidden.value = true;
        sortable.value = true;

        const changed = transactionTagsStore.sortTagDisplayOrderByTagName(activeTagGroupId.value, desc);

        if (changed) {
            displayOrderModified.value = true;
        }
    }

    function setSortable(): void {
        if (sortable.value || hasEditingTag.value) {
            return;
        }

        showHidden.value = true;
        sortable.value = true;
        displayOrderModified.value = false;
    }

    function saveSortResult(): void {
        if (!displayOrderModified.value) {
            showHidden.value = false;
            sortable.value = false;
            return;
        }

        displayOrderSaving.value = true;
        showLoading();

        transactionTagsStore.updateTagDisplayOrders(activeTagGroupId.value).then(() => {
            displayOrderSaving.value = false;
            hideLoading();

            showHidden.value = false;
            sortable.value = false;
            displayOrderModified.value = false;
        }).catch(error => {
            displayOrderSaving.value = false;
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function cancelSort(): void {
        if (!displayOrderModified.value) {
            showHidden.value = false;
            sortable.value = false;
            return;
        }

        displayOrderSaving.value = true;
        showLoading();

        transactionTagsStore.loadAllTags({
            force: false
        }).then(() => {
            displayOrderSaving.value = false;
            hideLoading();

            showHidden.value = false;
            sortable.value = false;
            displayOrderModified.value = false;
        }).catch(error => {
            displayOrderSaving.value = false;
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    function addTagGroup(): void {
        showPrompt(tt('New Tag Group Name'), '', (value: string) => {
            showLoading();

            transactionTagsStore.saveTagGroup({
                tagGroup: TransactionTagGroup.createNewTagGroup(value)
            }).then(tagGroup => {
                hideLoading();
                activeTagGroupId.value = tagGroup.id;
            }).catch(error => {
                hideLoading();

                if (!error.processed) {
                    showToast(error.message || error);
                }
            });
        });
    }

    function renameTagGroup(): void {
        const tagGroup = transactionTagsStore.allTransactionTagGroupsMap[activeTagGroupId.value];

        if (!tagGroup) {
            showToast('Unable to rename this tag group');
            return;
        }

        showPrompt(tt('Rename Tag Group'), tagGroup.name || '', (value: string) => {
            showLoading();

            const newTagGroup = tagGroup.clone();
            newTagGroup.name = value;

            transactionTagsStore.saveTagGroup({
                tagGroup: newTagGroup
            }).then(() => {
                hideLoading();
            }).catch(error => {
                hideLoading();

                if (!error.processed) {
                    showToast(error.message || error);
                }
            });
        });
    }

    function removeTagGroup(): void {
        const tagGroup = transactionTagsStore.allTransactionTagGroupsMap[activeTagGroupId.value];

        if (!tagGroup) {
            showToast('Unable to delete this tag group');
            return;
        }

        const currentTagGroupIndex = allTagGroupsWithDefault.value.findIndex(group => group.id === tagGroup.id);

        showConfirm('Are you sure you want to delete this tag group?', () => {
            showLoading();

            transactionTagsStore.deleteTagGroup({
                tagGroup: tagGroup
            }).then(() => {
                hideLoading();

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
                hideLoading();

                if (!error.processed) {
                    showToast(error.message || error);
                }
            });
        });
    }

    function changeTagGroupDisplayOrder(): void {
        props.f7router.navigate('/tag/group/list');
    }

    function onSort(event: { el: { id: string }, from: number, to: number }): void {
        if (!event || !event.el || !event.el.id) {
            showToast('Unable to move tag');
            return;
        }

        const id = parseTagIdFromDomId(event.el.id);

        if (!id) {
            showToast('Unable to move tag');
            return;
        }

        transactionTagsStore.changeTagDisplayOrder({
            tagId: id,
            from: event.from,
            to: event.to
        }).then(() => {
            displayOrderModified.value = true;
        }).catch(error => {
            showToast(error.message || error);
        });
    }

    function scrollPopoverToSelectedItem(event: { $el: Framework7Dom }): void {
        scrollToSelectedItem(event.$el[0], '.popover-inner', '.popover-inner', 'li.list-item-selected');
    }

    function onPageAfterIn(): void {
        if (transactionTagsStore.transactionTagListStateInvalid && !loading.value) {
            reload();
        }

        routeBackOnError(props.f7router, loadingError);
    }

    init();

    return {
        tt, activeTagGroupId, newTag, editingTag, loading, showHidden, displayOrderModified, allTagGroupsWithDefault, tags,
        noAvailableTag, hasEditingTag, isTagModified, switchTagGroup, edit, newTagItem, sortable, moveToTagGroupId, tagToMove,
        tagToDelete, showMoveTagPopup, showMoreActionSheet, showDeleteActionSheet, displayOrderSaving, textDirection,
        firstShowingId, lastShowingId, displayTagGroupName, getTagDomId, reload, add, save, cancelSave, hide, moveTagToGroup,
        updateTagGroupSelected, remove, sortByName, setSortable, saveSortResult, cancelSort, addTagGroup, renameTagGroup,
        removeTagGroup, changeTagGroupDisplayOrder, onSort, scrollPopoverToSelectedItem, onPageAfterIn,
    };
}
