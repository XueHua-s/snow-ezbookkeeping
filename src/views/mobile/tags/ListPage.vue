<script setup lang="ts">
import { DEFAULT_TAG_GROUP_ID } from '@/consts/tag.ts';
import { TextDirection } from '@/core/text.ts';
import type { Router } from 'framework7/types';
import { useListPage } from './ListPage/useListPage.ts';

const props = defineProps<{
    f7router: Router.Router;
}>();

const {
    tt, activeTagGroupId, newTag, editingTag, loading, showHidden, displayOrderModified, allTagGroupsWithDefault, tags,
    noAvailableTag, hasEditingTag, isTagModified, switchTagGroup, edit, sortable, moveToTagGroupId, tagToMove, tagToDelete,
    showMoveTagPopup, showMoreActionSheet, showDeleteActionSheet, displayOrderSaving, textDirection, firstShowingId,
    lastShowingId, displayTagGroupName, getTagDomId, reload, add, save, cancelSave, hide, moveTagToGroup,
    updateTagGroupSelected, remove, sortByName, setSortable, saveSortResult, cancelSort, addTagGroup, renameTagGroup,
    removeTagGroup, changeTagGroupDisplayOrder, onSort, scrollPopoverToSelectedItem, onPageAfterIn,
} = useListPage(props);

</script>

<template>
    <f7-page :ptr="!sortable && !hasEditingTag" @ptr:refresh="reload" @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading }" :back-link="tt('Back')" v-if="!sortable"></f7-nav-left>
            <f7-nav-left v-else-if="sortable">
                <f7-link icon-f7="xmark" :class="{ 'disabled': displayOrderSaving }" :aria-label="tt('Cancel')" @click="cancelSort"></f7-link>
            </f7-nav-left>
            <f7-nav-title>
                <f7-link popover-open=".tag-group-popover-menu" :class="{ 'disabled': loading || sortable || displayOrderModified || hasEditingTag }">
                    <span style="color: var(--f7-text-color)">{{ displayTagGroupName }}</span>
                    <f7-icon class="page-title-bar-icon" color="gray" style="opacity: 0.5" f7="chevron_down_circle_fill"></f7-icon>
                </f7-link>
            </f7-nav-title>
            <f7-nav-right :class="{ 'navbar-compact-icons': true, 'disabled': loading }">
                <f7-link icon-f7="ellipsis" :class="{ 'disabled': hasEditingTag || sortable }" :aria-label="tt('More')" @click="showMoreActionSheet = true"></f7-link>
                <f7-link icon-f7="plus" :class="{ 'disabled': hasEditingTag }" :aria-label="tt('Add')" v-if="!sortable" @click="add"></f7-link>
                <f7-link icon-f7="checkmark_alt" :class="{ 'disabled': displayOrderSaving || !displayOrderModified || hasEditingTag }" :aria-label="tt('Save')" @click="saveSortResult" v-else-if="sortable"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-popover class="tag-group-popover-menu"
                    @popover:open="scrollPopoverToSelectedItem">
            <f7-list dividers>
                <f7-list-item link="#" no-chevron popover-close
                              :title="tagGroup.name"
                              :class="{ 'list-item-selected': activeTagGroupId === tagGroup.id }"
                              :key="tagGroup.id"
                              v-for="tagGroup in allTagGroupsWithDefault"
                              @click="switchTagGroup(tagGroup.id)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="activeTagGroupId === tagGroup.id"></f7-icon>
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-popover>

        <f7-list strong inset dividers class="tag-item-list margin-top-half skeleton-text" v-if="loading">
            <f7-list-item :key="itemIdx" v-for="itemIdx in [ 1, 2, 3 ]">
                <template #media>
                    <f7-icon class="transaction-tag-icon" f7="number"></f7-icon>
                </template>
                <template #title>
                    <div class="display-flex">
                        <div class="transaction-tag-list-item-content list-item-valign-middle padding-inline-start-half">Tag Name</div>
                    </div>
                </template>
            </f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="tag-item-list margin-top-half" v-if="!loading && noAvailableTag && !newTag">
            <f7-list-item :title="tt('No available tag')"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers sortable class="tag-item-list margin-top-half"
                 :sortable-enabled="sortable" @sortable:sort="onSort"
                 v-if="!loading">
            <f7-list-item swipeout
                          :class="{ 'actual-first-child': tag.id === firstShowingId, 'actual-last-child': tag.id === lastShowingId && !newTag, 'editing-list-item': editingTag.id === tag.id }"
                          :id="getTagDomId(tag)"
                          :key="tag.id"
                          v-for="tag in tags"
                          v-show="showHidden || !tag.hidden"
                          @taphold="setSortable()">
                <template #media>
                    <f7-icon class="transaction-tag-icon" f7="number">
                        <f7-badge color="gray" class="right-bottom-icon" v-if="tag.hidden">
                            <f7-icon f7="eye_slash_fill"></f7-icon>
                        </f7-badge>
                    </f7-icon>
                </template>
                <template #title>
                    <div class="display-flex">
                        <div class="transaction-tag-list-item-content list-item-valign-middle padding-inline-start-half"
                             v-if="editingTag.id !== tag.id">
                            {{ tag.name }}
                        </div>
                        <f7-input class="list-title-input padding-inline-start-half"
                                  type="text"
                                  :placeholder="tt('Tag Title')"
                                  v-else-if="editingTag.id === tag.id"
                                  v-model:value="editingTag.name"
                                  @keyup.enter="save(editingTag)">
                        </f7-input>
                    </div>
                </template>
                <template #after>
                    <f7-button raised fill icon-f7="checkmark_alt" color="blue"
                               :class="{ 'no-padding': true, 'disabled': !isTagModified(tag) }"
                               :aria-label="tt('Save')"
                               v-if="editingTag.id === tag.id"
                               @click="save(editingTag)">
                    </f7-button>
                    <f7-button raised fill icon-f7="xmark" color="gray"
                               class="no-padding margin-inline-start-half"
                               :aria-label="tt('Cancel')"
                               v-if="editingTag.id === tag.id"
                               @click="cancelSave(editingTag)">
                    </f7-button>
                </template>
                <f7-swipeout-actions :left="textDirection === TextDirection.LTR"
                                     :right="textDirection === TextDirection.RTL"
                                     v-if="sortable && editingTag.id !== tag.id">
                    <f7-swipeout-button class="padding-horizontal" overswipe close
                                        :aria-label="tag.hidden ? tt('Show') : tt('Hide')"
                                        :color="tag.hidden ? 'blue' : 'gray'"
                                        @click="hide(tag, !tag.hidden)">
                        <f7-icon :f7="tag.hidden ? 'eye' : 'eye_slash'"></f7-icon>
                    </f7-swipeout-button>
                </f7-swipeout-actions>
                <f7-swipeout-actions :left="textDirection === TextDirection.RTL"
                                     :right="textDirection === TextDirection.LTR"
                                     v-if="!sortable && editingTag.id !== tag.id">
                    <f7-swipeout-button color="primary" close :class="{ 'disabled': allTagGroupsWithDefault.length < 2 }" :text="tt('Move')" @click="moveTagToGroup(tag)"></f7-swipeout-button>
                    <f7-swipeout-button color="orange" close :text="tt('Edit')" @click="edit(tag)"></f7-swipeout-button>
                    <f7-swipeout-button color="red" class="padding-horizontal" :aria-label="tt('Delete')" @click="remove(tag, false)">
                        <f7-icon f7="trash"></f7-icon>
                    </f7-swipeout-button>
                </f7-swipeout-actions>
            </f7-list-item>

            <f7-list-item ref="newTagItem" class="editing-list-item" v-if="newTag">
                <template #media>
                    <f7-icon class="transaction-tag-icon" f7="number"></f7-icon>
                </template>
                <template #title>
                    <div class="display-flex">
                        <f7-input class="list-title-input padding-inline-start-half"
                                  type="text"
                                  :placeholder="tt('Tag Title')"
                                  v-model:value="newTag.name"
                                  @keyup.enter="save(newTag)">
                        </f7-input>
                    </div>
                </template>
                <template #after>
                    <f7-button raised fill icon-f7="checkmark_alt" color="blue"
                               :class="{ 'no-padding': true, 'disabled': !isTagModified(newTag) }"
                               :aria-label="tt('Save')"
                               @click="save(newTag)">
                    </f7-button>
                    <f7-button raised fill icon-f7="xmark" color="gray"
                               class="no-padding margin-inline-start-half"
                               :aria-label="tt('Cancel')"
                               @click="cancelSave(newTag)">
                    </f7-button>
                </template>
            </f7-list-item>
        </f7-list>

        <f7-popup push :close-on-escape="false" :opened="showMoveTagPopup"
                  @popup:closed="showMoveTagPopup = false">
            <f7-page>
                <f7-navbar>
                    <f7-nav-left>
                        <f7-link popup-close icon-f7="xmark" :aria-label="tt('Cancel')"></f7-link>
                    </f7-nav-left>
                    <f7-nav-title :title="tt('Move to...')"></f7-nav-title>
                    <f7-nav-right>
                        <f7-link icon-f7="checkmark_alt"
                                 :class="{ 'disabled': !tagToMove || !moveToTagGroupId }"
                                 :aria-label="tt('Move')"
                                 @click="moveTagToGroup(tagToMove, moveToTagGroupId)"></f7-link>
                    </f7-nav-right>
                </f7-navbar>

                <f7-list strong inset dividers class="margin-top" v-if="!loading && allTagGroupsWithDefault.length < 2">
                    <f7-list-item :title="tt('No available tag group')"></f7-list-item>
                </f7-list>

                <f7-list strong inset dividers class="margin-vertical" v-if="allTagGroupsWithDefault.length >= 2">
                    <template :key="tagGroup.id" v-for="tagGroup in allTagGroupsWithDefault">
                        <f7-list-item checkbox
                                      :title="tagGroup.name"
                                      :value="tagGroup.id"
                                      :checked="moveToTagGroupId === tagGroup.id"
                                      :key="tagGroup.id"
                                      @change="updateTagGroupSelected"
                                      v-if="tagToMove?.groupId !== tagGroup.id"></f7-list-item>
                    </template>
                </f7-list>
            </f7-page>
        </f7-popup>

        <f7-actions close-by-outside-click close-on-escape :opened="showMoreActionSheet" @actions:closed="showMoreActionSheet = false">
            <f7-actions-group>
                <f7-actions-button @click="addTagGroup">{{ tt('Add Tag Group') }}</f7-actions-button>
                <f7-actions-button @click="renameTagGroup"
                                   v-if="activeTagGroupId && activeTagGroupId !== DEFAULT_TAG_GROUP_ID">{{ tt('Rename Tag Group') }}</f7-actions-button>
                <f7-actions-button color="red" :class="{ 'disabled': tags && tags.length > 0 }"
                                   @click="removeTagGroup"
                                   v-if="activeTagGroupId && activeTagGroupId !== DEFAULT_TAG_GROUP_ID">{{ tt('Delete Tag Group') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group v-if="allTagGroupsWithDefault.length >= 2">
                <f7-actions-button @click="changeTagGroupDisplayOrder">{{ tt('Change Group Display Order') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button :class="{ 'disabled': !tags || tags.length < 2 }" @click="setSortable()">{{ tt('Sort') }}</f7-actions-button>
                <f7-actions-button :class="{ 'disabled': !tags || tags.length < 2 }" @click="sortByName(false)">{{ tt('Sort by Name (A to Z)') }}</f7-actions-button>
                <f7-actions-button :class="{ 'disabled': !tags || tags.length < 2 }" @click="sortByName(true)">{{ tt('Sort by Name (Z to A)') }}</f7-actions-button>
                <f7-actions-button v-if="!showHidden" @click="showHidden = true">{{ tt('Show Hidden Transaction Tags') }}</f7-actions-button>
                <f7-actions-button v-if="showHidden" @click="showHidden = false">{{ tt('Hide Hidden Transaction Tags') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <f7-actions close-by-outside-click close-on-escape :opened="showDeleteActionSheet" @actions:closed="showDeleteActionSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to delete this tag?') }}</f7-actions-label>
                <f7-actions-button color="red" @click="remove(tagToDelete, true)">{{ tt('Delete') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>



<style>
.tag-item-list.list .item-media + .item-inner {
    margin-inline-start: 5px;
}

.transaction-tag-list-item-content {
    overflow: hidden;
    text-overflow: ellipsis;
}

.tag-group-popover-menu .popover-inner {
    max-height: 440px;
    overflow-y: auto;
}
</style>
