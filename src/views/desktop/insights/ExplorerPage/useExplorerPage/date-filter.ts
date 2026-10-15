import { DateRange, DateRangeScene } from '@/core/datetime.ts';
import {
    getDateRangeByDateType,
    getDateTypeByDateRange,
    getShiftedDateRangeAndDateType
} from '@/lib/datetime.ts';
import type { useExplorerPageState } from './state.ts';

interface Context {
    showCustomDateRangeDialog: ReturnType<typeof useExplorerPageState>['showCustomDateRangeDialog'];
    currentFilter: ReturnType<typeof useExplorerPageState>['currentFilter'];
    firstDayOfWeek: ReturnType<typeof useExplorerPageState>['firstDayOfWeek'];
    fiscalYearStart: ReturnType<typeof useExplorerPageState>['fiscalYearStart'];
    explorersStore: ReturnType<typeof useExplorerPageState>['explorersStore'];
    loading: ReturnType<typeof useExplorerPageState>['loading'];
    router: ReturnType<typeof useExplorerPageState>['router'];
    getFilterLinkUrl: ReturnType<typeof useExplorerPageState>['getFilterLinkUrl'];
}

export function useExplorerPageDateFilter(context: Context) {
    const { showCustomDateRangeDialog, currentFilter, firstDayOfWeek, fiscalYearStart, explorersStore, loading, router, getFilterLinkUrl } = context;

    function setDateFilter(dateType: number): void {
        if (dateType === DateRange.Custom.type) { // Custom
            showCustomDateRangeDialog.value = true;
            return;
        } else if (currentFilter.value.dateRangeType === dateType) {
            return;
        }

        const dateRange = getDateRangeByDateType(dateType, firstDayOfWeek.value, fiscalYearStart.value);

        if (!dateRange) {
            return;
        }

        const changed = explorersStore.updateTransactionExplorerFilter({
            dateRangeType: dateRange.dateType,
            startTime: dateRange.minTime,
            endTime: dateRange.maxTime
        });

        if (changed) {
            loading.value = true;
            explorersStore.updateTransactionExplorerInvalidState(true);
            router.push(getFilterLinkUrl());
        }
    }

    function setCustomDateFilter(startTime: number, endTime: number): void {
        if (!startTime || !endTime) {
            return;
        }

        const chartDateType = getDateTypeByDateRange(startTime, endTime, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.InsightsExplorer);

        const changed = explorersStore.updateTransactionExplorerFilter({
            dateRangeType: chartDateType,
            startTime: startTime,
            endTime: endTime
        });

        showCustomDateRangeDialog.value = false;

        if (changed) {
            loading.value = true;
            explorersStore.updateTransactionExplorerInvalidState(true);
            router.push(getFilterLinkUrl());
        }
    }

    function shiftDateRange(scale: number): void {
        if (currentFilter.value.dateRangeType === DateRange.All.type) {
            return;
        }

        const newDateRange = getShiftedDateRangeAndDateType(currentFilter.value.startTime, currentFilter.value.endTime, scale, firstDayOfWeek.value, fiscalYearStart.value, DateRangeScene.Normal);

        const changed = explorersStore.updateTransactionExplorerFilter({
            dateRangeType: newDateRange.dateType,
            startTime: newDateRange.minTime,
            endTime: newDateRange.maxTime
        });

        if (changed) {
            loading.value = true;
            explorersStore.updateTransactionExplorerInvalidState(true);
            router.push(getFilterLinkUrl());
        }
    }

    return { setDateFilter, setCustomDateFilter, shiftDateRange };
}
