<script setup lang="ts">
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import RenameDialog from '@/components/desktop/RenameDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { DEFAULT_TAG_GROUP_ID } from '@/consts/tag.ts';
import {
    mdiCheck,
    mdiClose,
    mdiDeleteOutline,
    mdiDotsVertical,
    mdiDrag,
    mdiEyeOffOutline,
    mdiEyeOutline,
    mdiFolderMoveOutline,
    mdiMenu,
    mdiMenuDown,
    mdiMenuUp,
    mdiPencilOutline,
    mdiPlus,
    mdiPound,
    mdiRefresh,
    mdiSort,
    mdiSortAlphabeticalAscending,
    mdiSortAlphabeticalDescending
} from '@mdi/js';
import TagGroupChangeDisplayOrderDialog from './dialog/TagGroupChangeDisplayOrderDialog.vue';
import { useListPage } from './ListPage/useListPage.ts';

const {
    tt, activeTagGroupId, newTag, editingTag, loading, showHidden, displayOrderModified, allTagGroupsWithDefault, tags,
    noAvailableTag, hasEditingTag, isTagModified, switchTagGroup, edit, updating, activeTab, alwaysShowNav, showNav,
    hoveredTagId, tagUpdating, tagHiding, tagMoving, tagRemoving, currentMovingTag, showTagMoveToDialog,
    totalAvailableTagsCount, displayTotalAvailableTagsCount, availableTagCount, reload, add, addTagGroup, renameTagGroup,
    showChangeGroupDisplayOrderDialog, removeTagGroup, showMoveTagDialog, moveTagToGroup, save, cancelSave, sortByName,
    saveSortResult, hide, remove, onMove,
} = useListPage();

</script>

<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <v-layout class="page-with-navigation-drawer">
                    <v-navigation-drawer class="scrollable-tabs-navigation-drawer" :permanent="alwaysShowNav" v-model="showNav">
                        <div class="mx-4 my-3">
                            <span class="text-body-medium">{{ tt('Total tags') }}</span>
                            <div class="text-body-large mt-1">
                                <span v-if="!loading || totalAvailableTagsCount > 0">{{ displayTotalAvailableTagsCount }}</span>
                                <span v-else-if="loading && totalAvailableTagsCount <= 0">
                                    <v-skeleton-loader class="skeleton-no-margin pt-2 pb-1" type="text" :loading="true"></v-skeleton-loader>
                                </span>
                            </div>
                        </div>
                        <v-divider />
                        <v-tabs show-arrows
                                class="scrollable-vertical-tabs"
                                style="max-height: calc(100% - 82px)"
                                direction="vertical"
                                :prev-icon="mdiMenuUp" :next-icon="mdiMenuDown"
                                :disabled="loading || updating" v-model="activeTagGroupId">
                            <v-tab class="tab-text-truncate" :disabled="loading || updating || displayOrderModified || hasEditingTag"
                                   :key="tagGroup.id" :value="tagGroup.id"
                                   v-for="tagGroup in allTagGroupsWithDefault"
                                   @click="switchTagGroup(tagGroup.id)">
                                <span class="text-truncate">{{ tagGroup.name }}</span>
                            </v-tab>
                            <template v-if="loading && (!allTagGroupsWithDefault || allTagGroupsWithDefault.length < 2)">
                                <v-skeleton-loader class="skeleton-no-margin mx-4 mt-h1 py-3" type="text"
                                                   :key="itemIdx" :loading="true" v-for="itemIdx in [ 1, 2, 3, 4, 5 ]"></v-skeleton-loader>
                            </template>
                        </v-tabs>
                    </v-navigation-drawer>
                    <v-main>
                        <v-window class="d-flex flex-grow-1 disable-tab-transition w-100-window-container" v-model="activeTab">
                            <v-window-item value="tagListPage">
                                <v-card variant="flat" min-height="780">
                                    <template #title>
                                        <div class="title-and-toolbar d-flex align-center">
                                            <v-btn class="me-3 d-lg-none" density="compact" color="default" variant="plain"
                                                   :aria-label="tt('Open Menu')" :ripple="false" :icon="true" @click="showNav = !showNav">
                                                <v-icon :icon="mdiMenu" size="24" />
                                            </v-btn>
                                            <span>{{ tt('Transaction Tags') }}</span>
                                            <v-btn class="ms-3" color="default" variant="outlined"
                                                   :disabled="loading || updating || hasEditingTag" @click="add">{{ tt('Add') }}</v-btn>
                                            <v-btn class="ms-3" color="primary" variant="tonal"
                                                   :disabled="loading || updating || hasEditingTag" @click="saveSortResult"
                                                   v-if="displayOrderModified">{{ tt('Save Display Order') }}</v-btn>
                                            <v-btn density="compact" color="default" variant="text" class="ms-2"
                                                   :aria-label="tt('Refresh')" :icon="true" :disabled="loading || updating || hasEditingTag"
                                                   :loading="loading" @click="reload">
                                                <template #loader>
                                                    <v-progress-circular indeterminate size="20"/>
                                                </template>
                                                <v-icon :icon="mdiRefresh" size="24" />
                                                <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                                            </v-btn>
                                            <v-spacer/>
                                            <v-btn density="comfortable" color="default" variant="text" class="ms-2"
                                                   :aria-label="tt('More')" :disabled="loading || updating || hasEditingTag" :icon="true">
                                                <v-icon :icon="mdiDotsVertical" />
                                                <v-menu activator="parent">
                                                    <v-list>
                                                        <v-list-item :prepend-icon="mdiPlus" @click="addTagGroup">
                                                            <v-list-item-title>{{ tt('Add Tag Group') }}</v-list-item-title>
                                                        </v-list-item>
                                                        <v-list-item :prepend-icon="mdiPencilOutline"
                                                                     @click="renameTagGroup"
                                                                     v-if="activeTagGroupId && activeTagGroupId !== DEFAULT_TAG_GROUP_ID">
                                                            <v-list-item-title>{{ tt('Rename Tag Group') }}</v-list-item-title>
                                                        </v-list-item>
                                                        <v-list-item :prepend-icon="mdiDeleteOutline"
                                                                     :disabled="tags && tags.length > 0"
                                                                     @click="removeTagGroup"
                                                                     v-if="activeTagGroupId && activeTagGroupId !== DEFAULT_TAG_GROUP_ID">
                                                            <v-list-item-title>{{ tt('Delete Tag Group') }}</v-list-item-title>
                                                        </v-list-item>
                                                        <v-divider class="my-2" v-if="allTagGroupsWithDefault.length >= 2"/>
                                                        <v-list-item :prepend-icon="mdiSort"
                                                                     :disabled="!allTagGroupsWithDefault || allTagGroupsWithDefault.length < 2"
                                                                     :title="tt('Change Group Display Order')"
                                                                     v-if="allTagGroupsWithDefault.length >= 2"
                                                                     @click="showChangeGroupDisplayOrderDialog"></v-list-item>
                                                        <v-divider class="my-2"/>
                                                        <v-list-item :prepend-icon="mdiSortAlphabeticalAscending"
                                                                     :disabled="!tags || tags.length < 2"
                                                                     :title="tt('Sort by Name (A to Z)')"
                                                                     @click="sortByName(false)"></v-list-item>
                                                        <v-list-item :prepend-icon="mdiSortAlphabeticalDescending"
                                                                     :disabled="!tags || tags.length < 2"
                                                                     :title="tt('Sort by Name (Z to A)')"
                                                                     @click="sortByName(true)"></v-list-item>
                                                        <v-divider class="my-2"/>
                                                        <v-list-item :prepend-icon="mdiEyeOutline"
                                                                     :title="tt('Show Hidden Transaction Tags')"
                                                                     v-if="!showHidden" @click="showHidden = true"></v-list-item>
                                                        <v-list-item :prepend-icon="mdiEyeOffOutline"
                                                                     :title="tt('Hide Hidden Transaction Tags')"
                                                                     v-if="showHidden" @click="showHidden = false"></v-list-item>
                                                    </v-list>
                                                </v-menu>
                                            </v-btn>
                                        </div>
                                    </template>

                                    <v-table class="transaction-tags-table table-striped"
                                             density="default" :hover="!loading">
                                        <thead>
                                        <tr>
                                            <th>
                                                <div class="d-flex align-center">
                                                    <span>{{ tt('Tag Title') }}</span>
                                                    <v-spacer/>
                                                    <span>{{ tt('Operation') }}</span>
                                                </div>
                                            </th>
                                        </tr>
                                        </thead>

                                        <tbody v-if="loading && noAvailableTag && !newTag">
                                        <tr :key="itemIdx" v-for="itemIdx in [ 1, 2, 3, 4, 5 ]">
                                            <td class="px-0">
                                                <v-skeleton-loader type="text" :loading="true"></v-skeleton-loader>
                                            </td>
                                        </tr>
                                        </tbody>

                                        <tbody v-if="!loading && noAvailableTag && !newTag">
                                        <tr>
                                            <td>{{ tt('No available tag') }}</td>
                                        </tr>
                                        </tbody>

                                        <draggable-list tag="tbody"
                                                        item-key="id"
                                                        handle=".drag-handle"
                                                        ghost-class="dragging-item"
                                                        :class="{ 'has-bottom-border': newTag }"
                                                        :disabled="noAvailableTag"
                                                        v-model="tags"
                                                        @change="onMove">
                                            <template #item="{ element }">
                                                <tr class="transaction-tags-table-row-tag" v-if="showHidden || !element.hidden"
                                                    @mouseenter="hoveredTagId = element.id" @mouseleave="hoveredTagId = ''">
                                                    <td>
                                                        <div class="d-flex align-center">
                                                            <div class="d-flex align-center" v-if="editingTag.id !== element.id">
                                                                <v-badge class="right-bottom-icon" color="secondary"
                                                                         location="bottom right" offset-x="8" :icon="mdiEyeOffOutline"
                                                                         v-if="element.hidden">
                                                                    <v-icon size="20" start :icon="mdiPound"/>
                                                                </v-badge>
                                                                <v-icon size="20" start :icon="mdiPound" v-else-if="!element.hidden"/>
                                                                <span class="transaction-tag-name">{{ element.name }}</span>
                                                            </div>

                                                            <v-text-field class="w-100 me-2" type="text" autocomplete="off"
                                                                          density="compact" variant="underlined"
                                                                          :disabled="loading || updating"
                                                                          :placeholder="tt('Tag Title')"
                                                                          v-model="editingTag.name"
                                                                          v-else-if="editingTag.id === element.id"
                                                                          @keyup.enter="save(editingTag)"
                                                            >
                                                                <template #prepend>
                                                                    <v-badge class="right-bottom-icon" color="secondary"
                                                                             location="bottom right" offset-x="8" :icon="mdiEyeOffOutline"
                                                                             v-if="element.hidden">
                                                                        <v-icon size="20" start :icon="mdiPound"/>
                                                                    </v-badge>
                                                                    <v-icon size="20" start :icon="mdiPound" v-else-if="!element.hidden"/>
                                                                </template>
                                                            </v-text-field>

                                                            <v-spacer/>

                                                            <template v-if="hoveredTagId === element.id && !loading">
                                                                <v-btn class="px-2 ms-2" color="default"
                                                                       density="comfortable" variant="text"
                                                                       :prepend-icon="element.hidden ? mdiEyeOutline : mdiEyeOffOutline"
                                                                       :loading="tagHiding[element.id]"
                                                                       :disabled="loading || updating"
                                                                       v-if="editingTag.id !== element.id"
                                                                       @click="hide(element, !element.hidden)">
                                                                    <template #loader>
                                                                        <v-progress-circular indeterminate size="20" width="2"/>
                                                                    </template>
                                                                    {{ element.hidden ? tt('Show') : tt('Hide') }}
                                                                </v-btn>
                                                                <v-btn class="px-2" color="default"
                                                                       density="comfortable" variant="text"
                                                                       :prepend-icon="mdiFolderMoveOutline"
                                                                       :loading="tagMoving[element.id]"
                                                                       :disabled="loading || updating || allTagGroupsWithDefault.length < 2"
                                                                       v-if="editingTag.id !== element.id"
                                                                       @click="showMoveTagDialog(element)">
                                                                    <template #loader>
                                                                        <v-progress-circular indeterminate size="20" width="2"/>
                                                                    </template>
                                                                    {{ tt('Move') }}
                                                                </v-btn>
                                                                <v-btn class="px-2" color="default"
                                                                       density="comfortable" variant="text"
                                                                       :prepend-icon="mdiPencilOutline"
                                                                       :loading="tagUpdating[element.id]"
                                                                       :disabled="loading || updating"
                                                                       v-if="editingTag.id !== element.id"
                                                                       @click="edit(element)">
                                                                    <template #loader>
                                                                        <v-progress-circular indeterminate size="20" width="2"/>
                                                                    </template>
                                                                    {{ tt('Edit') }}
                                                                </v-btn>
                                                                <v-btn class="px-2" color="default"
                                                                       density="comfortable" variant="text"
                                                                       :prepend-icon="mdiDeleteOutline"
                                                                       :loading="tagRemoving[element.id]"
                                                                       :disabled="loading || updating"
                                                                       v-if="editingTag.id !== element.id"
                                                                       @click="remove(element)">
                                                                    <template #loader>
                                                                        <v-progress-circular indeterminate size="20" width="2"/>
                                                                    </template>
                                                                    {{ tt('Delete') }}
                                                                </v-btn>
                                                            </template>

                                                            <template v-if="editingTag.id === element.id">
                                                                <v-btn class="px-2"
                                                                       density="comfortable" variant="text"
                                                                       :prepend-icon="mdiCheck"
                                                                       :loading="tagUpdating[element.id]"
                                                                       :disabled="loading || updating || !isTagModified(element)"
                                                                       @click="save(editingTag)">
                                                                    <template #loader>
                                                                        <v-progress-circular indeterminate size="20" width="2"/>
                                                                    </template>
                                                                    {{ tt('Save') }}
                                                                </v-btn>
                                                                <v-btn class="px-2" color="default"
                                                                       density="comfortable" variant="text"
                                                                       :prepend-icon="mdiClose"
                                                                       :disabled="loading || updating"
                                                                       @click="cancelSave(editingTag)">
                                                                    {{ tt('Cancel') }}
                                                                </v-btn>
                                                            </template>

                                                            <span class="ms-2">
                                                                <v-icon :class="!loading && !updating && !hasEditingTag && availableTagCount > 1 ? 'drag-handle' : 'disabled'"
                                                                        :aria-label="tt('Drag to Reorder')" :icon="mdiDrag"/>
                                                                <v-tooltip activator="parent" v-if="!loading && !updating && !hasEditingTag && availableTagCount > 1 && hoveredTagId === element.id">{{ tt('Drag to Reorder') }}</v-tooltip>
                                                            </span>
                                                        </div>
                                                    </td>
                                                </tr>
                                            </template>
                                        </draggable-list>

                                        <tbody ref="newTagRow" v-if="newTag">
                                        <tr :class="{ 'even-row': (availableTagCount & 1) === 1}">
                                            <td>
                                                <div class="d-flex align-center">
                                                    <v-text-field class="w-100 me-2" type="text" autocomplete="off" color="primary"
                                                                  density="compact" variant="underlined"
                                                                  :disabled="loading || updating" :placeholder="tt('Tag Title')"
                                                                  v-model="newTag.name" @keyup.enter="save(newTag)">
                                                        <template #prepend>
                                                            <v-icon size="20" start :icon="mdiPound"/>
                                                        </template>
                                                    </v-text-field>

                                                    <v-spacer/>

                                                    <v-btn class="px-2" density="comfortable" variant="text"
                                                           :prepend-icon="mdiCheck"
                                                           :loading="tagUpdating['']"
                                                           :disabled="loading || updating || !isTagModified(newTag)"
                                                           @click="save(newTag)">
                                                        <template #loader>
                                                            <v-progress-circular indeterminate size="20" width="2"/>
                                                        </template>
                                                        {{ tt('Save') }}
                                                    </v-btn>
                                                    <v-btn class="px-2" color="default"
                                                           density="comfortable" variant="text"
                                                           :prepend-icon="mdiClose"
                                                           :disabled="loading || updating"
                                                           @click="cancelSave(newTag)">
                                                        {{ tt('Cancel') }}
                                                    </v-btn>
                                                    <span class="ms-2">
                                                        <v-icon class="disabled" :icon="mdiDrag"/>
                                                    </span>
                                                </div>
                                            </td>
                                        </tr>
                                        </tbody>
                                    </v-table>
                                </v-card>
                            </v-window-item>
                        </v-window>
                    </v-main>
                </v-layout>
            </v-card>
        </v-col>
    </v-row>

    <v-dialog width="640" v-model="showTagMoveToDialog">
        <one-column-dialog-layout content-class="pa-0" :disabled="loading || updating"
                                  :title="tt('Move to...')" :cancel-button-title="tt('Close')"
                                  @cancel="showTagMoveToDialog = false">
            <template #content>
                <v-table hover density="comfortable" class="w-100 table-striped">
                    <tbody>
                    <tr class="cursor-pointer" :key="tagGroup.id" v-for="tagGroup in allTagGroupsWithDefault" v-show="activeTagGroupId !== tagGroup.id">
                        <td @click="moveTagToGroup(currentMovingTag, tagGroup.id)">
                            <span>{{ tagGroup.name }}</span>
                        </td>
                    </tr>
                    </tbody>
                </v-table>
            </template>
        </one-column-dialog-layout>
    </v-dialog>

    <tag-group-change-display-order-dialog ref="tagGroupChangeDisplayOrderDialog" />

    <rename-dialog ref="renameDialog"
                   :default-title="tt('Rename Tag Group')"
                   :label="tt('Tag Group Name')" :placeholder="tt('Tag Group Name')" />
    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
</template>



<style>
.transaction-tags-table {
    tr:not(:last-child) > td > div,
    .has-bottom-border tr:last-child > td > div {
        padding-bottom: 1px;
    }

    tr.transaction-tags-table-row-tag {
        .right-bottom-icon {
            .v-badge__badge {
                padding-bottom: 1px;
            }
        }
    }

    .v-text-field {
        .v-input__prepend {
            margin-inline-end: 0;

            > .v-icon {
                opacity: 1;
            }
        }

        &.v-input--plain-underlined .v-input__prepend {
            padding-top: 4px;
        }

        .v-field {
            font-size: 0.875rem;

            .v-field__input {
                padding-top: 0;
            }
        }
    }
}
</style>
