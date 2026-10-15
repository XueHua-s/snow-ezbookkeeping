<script setup lang="ts">
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import { CategoryType } from '@/core/category.ts';
import { getCategoryIconType } from '@/lib/icon.ts';
import {
    mdiDeleteOutline,
    mdiDotsVertical,
    mdiDrag,
    mdiEyeOffOutline,
    mdiEyeOutline,
    mdiMenu,
    mdiPencilOutline,
    mdiRefresh
} from '@mdi/js';
import { VNavigationDrawer } from 'vuetify/components/VNavigationDrawer';
import EditDialog from './list/dialogs/EditDialog.vue';
import PresetDialog from './list/dialogs/PresetDialog.vue';
import { useListPage } from './ListPage/useListPage.ts';

const {
    tt, loading, primaryCategoryId, activeCategoryType, activeTab, updating, hoveredCategoryId, categoryHiding,
    categoryRemoving, displayOrderModified, cardMinHeight, alwaysShowNav, showNav, showHidden, showPresetDialog,
    primaryCategories, hasSubCategories, categories, noAvailableCategory, noCategory, availableCategoryCount,
    isCategorySupportSwitch, switchAllPrimaryCategories, switchPrimaryCategory, reload, add, edit, hide, remove,
    saveSortResult, onMove, onPresetCategorySaved,
} = useListPage();

</script>

<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <v-layout class="page-with-navigation-drawer">
                    <v-navigation-drawer ref="navbar" :permanent="alwaysShowNav" v-model="showNav">
                        <div class="mx-4 my-3">
                            <btn-vertical-group :disabled="loading" :buttons="[
                                { name: tt('Expense'), value: CategoryType.Expense },
                                { name: tt('Income'), value: CategoryType.Income },
                                { name: tt('Transfer'), value: CategoryType.Transfer }
                            ]" v-model="activeCategoryType" @update:model-value="switchAllPrimaryCategories" />
                        </div>
                        <v-divider />
                        <v-tabs show-arrows class="my-3" direction="vertical"
                                :disabled="loading" v-model="primaryCategoryId">
                            <v-tab class="tab-text-truncate" value="0" @click="switchAllPrimaryCategories">
                                <span class="text-truncate">{{ tt('Primary Categories') }}</span>
                            </v-tab>
                            <template :key="category.id" v-for="category in primaryCategories">
                                <v-tab class="tab-text-truncate" :value="category.id" v-if="!category.hidden"
                                       @click="switchPrimaryCategory(category)">
                                    <span class="text-truncate">{{ category.name }}</span>
                                </v-tab>
                            </template>
                            <template v-if="loading && (!primaryCategories || primaryCategories.length < 1)">
                                <v-skeleton-loader class="skeleton-no-margin mx-4 mt-h1 py-3" type="text"
                                                   :key="itemIdx" :loading="true" v-for="itemIdx in [ 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 ]"></v-skeleton-loader>
                            </template>
                        </v-tabs>
                    </v-navigation-drawer>
                    <v-main>
                        <v-window class="d-flex flex-grow-1 disable-tab-transition w-100-window-container" v-model="activeTab">
                            <v-window-item value="categoryPage">
                                <v-card variant="flat" :min-height="cardMinHeight">
                                    <template #title>
                                        <div class="title-and-toolbar d-flex align-center">
                                            <v-btn class="me-3 d-lg-none" density="compact" color="default" variant="plain"
                                                   :aria-label="tt('Open Menu')" :ripple="false" :icon="true" @click="showNav = !showNav">
                                                <v-icon :icon="mdiMenu" size="24" />
                                            </v-btn>
                                            <span>{{ tt('Transaction Categories') }}</span>
                                            <v-btn class="ms-3" color="default" variant="outlined"
                                                   :disabled="loading || updating" @click="add">{{ tt('Add') }}</v-btn>
                                            <v-btn class="ms-3" color="primary" variant="tonal"
                                                   :disabled="loading || updating" @click="saveSortResult"
                                                   v-if="displayOrderModified">{{ tt('Save Display Order') }}</v-btn>
                                            <v-btn density="compact" color="default" variant="text" class="ms-2"
                                                   :aria-label="tt('Refresh')" :icon="true" :loading="loading || updating" @click="reload(true)">
                                                <template #loader>
                                                    <v-progress-circular indeterminate size="20"/>
                                                </template>
                                                <v-icon :icon="mdiRefresh" size="24" />
                                                <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                                            </v-btn>
                                            <v-spacer/>
                                            <v-btn density="comfortable" color="default" variant="text" class="ms-2"
                                                   :aria-label="tt('More')" :disabled="loading || updating" :icon="true">
                                                <v-icon :icon="mdiDotsVertical" />
                                                <v-menu activator="parent">
                                                    <v-list>
                                                        <v-list-item :prepend-icon="mdiEyeOutline"
                                                                     :title="tt('Show Hidden Transaction Categories')"
                                                                     v-if="!showHidden" @click="showHidden = true"></v-list-item>
                                                        <v-list-item :prepend-icon="mdiEyeOffOutline"
                                                                     :title="tt('Hide Hidden Transaction Categories')"
                                                                     v-if="showHidden" @click="showHidden = false"></v-list-item>
                                                    </v-list>
                                                </v-menu>
                                            </v-btn>
                                        </div>
                                    </template>

                                    <v-table class="transaction-category-table table-striped"
                                             density="default" :hover="!loading">
                                        <thead>
                                        <tr>
                                            <th>
                                                <div class="d-flex align-center">
                                                    <span>{{ tt('Category Name') }}</span>
                                                    <v-spacer/>
                                                    <span>{{ tt('Operation') }}</span>
                                                </div>
                                            </th>
                                        </tr>
                                        </thead>

                                        <tbody v-if="loading && noAvailableCategory">
                                        <tr :key="itemIdx" v-for="itemIdx in [ 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 ]">
                                            <td class="px-0">
                                                <v-skeleton-loader type="text" :loading="true"></v-skeleton-loader>
                                            </td>
                                        </tr>
                                        </tbody>

                                        <tbody v-if="!loading && noAvailableCategory">
                                        <tr>
                                            <td>
                                                <div class="d-flex align-center">
                                                    <span>{{ tt('No available category') }}</span>
                                                    <v-btn class="ms-3" color="default" variant="outlined"
                                                           @click="showPresetDialog = true"
                                                           v-if="hasSubCategories && noCategory">
                                                        {{ tt('Add Default Categories') }}
                                                    </v-btn>
                                                </div>
                                            </td>
                                        </tr>
                                        </tbody>

                                        <draggable-list tag="tbody"
                                                        item-key="id"
                                                        handle=".drag-handle"
                                                        ghost-class="dragging-item"
                                                        :disabled="noAvailableCategory"
                                                        v-model="categories"
                                                        @change="onMove">
                                            <template #item="{ element }">
                                                <tr class="transaction-category-table-row" v-if="showHidden || !element.hidden"
                                                    @mouseenter="hoveredCategoryId = element.id" @mouseleave="hoveredCategoryId = ''">
                                                    <td>
                                                        <div class="d-flex align-center">
                                                            <div class="d-flex align-center" :class="{ 'cursor-pointer': isCategorySupportSwitch(element) }"
                                                                 @click="switchPrimaryCategory(element)">
                                                                <ItemIcon :icon-type="getCategoryIconType(element.iconType)"
                                                                          :icon-id="element.icon" :color="element.color"
                                                                          :hidden-status="element.hidden" />
                                                                <div class="d-flex flex-column py-2">
                                                                    <span class="ms-2">{{ element.name }}</span>
                                                                    <span class="transaction-category-comment ms-2">{{ element.comment }}</span>
                                                                </div>
                                                            </div>

                                                            <v-spacer/>

                                                            <template v-if="hoveredCategoryId === element.id && !loading">
                                                                <v-btn class="px-2 ms-2" color="default"
                                                                       density="comfortable" variant="text"
                                                                       :prepend-icon="element.hidden ? mdiEyeOutline : mdiEyeOffOutline"
                                                                       :loading="categoryHiding[element.id]"
                                                                       :disabled="loading || updating"
                                                                       @click="hide(element, !element.hidden)">
                                                                    <template #loader>
                                                                        <v-progress-circular indeterminate size="20" width="2"/>
                                                                    </template>
                                                                    {{ element.hidden ? tt('Show') : tt('Hide') }}
                                                                </v-btn>
                                                                <v-btn class="px-2" color="default"
                                                                       density="comfortable" variant="text"
                                                                       :prepend-icon="mdiPencilOutline"
                                                                       :disabled="loading || updating"
                                                                       @click="edit(element)">
                                                                    {{ tt('Edit') }}
                                                                </v-btn>
                                                                <v-btn class="px-2" color="default"
                                                                       density="comfortable" variant="text"
                                                                       :prepend-icon="mdiDeleteOutline"
                                                                       :loading="categoryRemoving[element.id]"
                                                                       :disabled="loading || updating"
                                                                       @click="remove(element)">
                                                                    <template #loader>
                                                                        <v-progress-circular indeterminate size="20" width="2"/>
                                                                    </template>
                                                                    {{ tt('Delete') }}
                                                                </v-btn>
                                                            </template>

                                                            <span class="ms-2">
                                                                <v-icon :class="!loading && !updating && availableCategoryCount > 1 ? 'drag-handle' : 'disabled'"
                                                                        :aria-label="tt('Drag to Reorder')" :icon="mdiDrag"/>
                                                                <v-tooltip activator="parent" v-if="!loading && !updating && availableCategoryCount > 1 && hoveredCategoryId === element.id">{{ tt('Drag to Reorder') }}</v-tooltip>
                                                            </span>
                                                        </div>
                                                    </td>
                                                </tr>
                                            </template>
                                        </draggable-list>
                                    </v-table>
                                </v-card>
                            </v-window-item>
                        </v-window>
                    </v-main>
                </v-layout>
            </v-card>
        </v-col>
    </v-row>

    <edit-dialog ref="editDialog" />
    <preset-dialog :category-type="activeCategoryType" v-model:show="showPresetDialog"
                            @category:saved="onPresetCategorySaved" />

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
</template>



<style>
.transaction-category-table .transaction-category-comment {
    font-size: 0.8rem;
    color: rgba(var(--v-theme-on-background), var(--v-medium-emphasis-opacity)) !important;
}
</style>
