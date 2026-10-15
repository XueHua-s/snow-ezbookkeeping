import type { SetupProps } from '@/lib/ui/props.ts';
import {
    type CommonTrendsChartProps,
    type TrendsBarChartClickEvent,
    type TrendsChartDateType,
    useTrendsChartBase
} from '@/components/base/TrendsChartBase.ts';
import { itemAndIndex } from '@/core/base.ts';
import type { ColorStyleValue, ColorValue } from '@/core/color.ts';
import {
    type UnixTimeRange,
    DateRangeScene
} from '@/core/datetime.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import {
    ChartDataAggregationType,
    ChartDateAggregationType
} from '@/core/statistics.ts';
import {
    getDisplayColor
} from '@/lib/color.ts';
import {
    getDateTypeByDateRange,
    getFiscalYearFromUnixTime,
    getYearMonthFirstUnixTime,
    getYearMonthLastUnixTime,
    parseDateTimeFromUnixTime
} from '@/lib/datetime.ts';
import {
    BIG_DECIMAL_ZERO,
    isBigDecimal
} from '@/lib/numeral.ts';
import {
    sortStatisticsItems
} from '@/lib/statistics.ts';
import { useI18n } from '@/locales/helpers.ts';
import type { SortableTransactionStatisticDataItem } from '@/models/transaction.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useUserStore } from '@/stores/user.ts';
import { computed, ref, watch } from 'vue';

export interface TrendsBarChartLegend {
    readonly id: string;
    readonly name: string;
    readonly color: ColorStyleValue;
    readonly displayOrders: number[];
}

export interface TrendsBarChartDataAmount extends SortableTransactionStatisticDataItem, TrendsBarChartLegend {
    value: BigDecimal;
}

export interface TrendsBarChartDataItem {
    index: number;
    dateRange: UnixTimeRange;
    displayDateRange: string;
    items: TrendsBarChartDataAmount[];
    totalAmount: BigDecimal;
    totalPositiveAmount: BigDecimal;
    maxAmount: BigDecimal;
    percent: number;
}

export interface TrendsBarChartVirtualListData {
    items: TrendsBarChartDataItem[],
    topPosition: number
}

export interface TrendsBarChartData {
    readonly data: TrendsBarChartDataItem[];
    readonly legends: TrendsBarChartLegend[];
}

export interface MobileTrendsChartProps<T extends TrendsChartDateType> extends CommonTrendsChartProps<T> {
    loading?: boolean;
}

// Coordinates TrendsBarChart state and effects; called once from the component setup.
export function useTrendsBarChart(props: SetupProps<MobileTrendsChartProps<TrendsChartDateType>>, emit: {
    (e: 'click', value: TrendsBarChartClickEvent): void;
}) {
    const {
        tt,
        formatDateTimeToShortDate,
        formatDateTimeToGregorianLikeShortYear,
        formatDateTimeToGregorianLikeShortYearMonth,
        formatYearQuarterToGregorianLikeYearQuarter,
        formatDateTimeToGregorianLikeFiscalYear,
        formatAmountToLocalizedNumeralsWithCurrency
    } = useI18n();

    const { allDateRanges, getItemName } = useTrendsChartBase(props);

    const settingsStore = useSettingsStore();

    const userStore = useUserStore();

    const allDisplayDataItemsVersion = ref<number>(0);

    const unselectedLegends = ref<Record<string, boolean>>({});

    const virtualDataItems = ref<TrendsBarChartVirtualListData>({
        items: [],
        topPosition: 0
    });

    const chartColors = computed<ColorValue[]>(() => settingsStore.chartColorList);

    const useVirtualList = computed<boolean>(() => allDisplayDataItems.value.legends.length <= 1 || props.stacked);

    const allDisplayDataItems = computed<TrendsBarChartData>(() => {
        const allDateRangeItemsMap: Record<string, TrendsBarChartDataAmount[]> = {};
        const legends: TrendsBarChartLegend[] = [];

        for (const [item, index] of itemAndIndex(props.items)) {
            if (item.hidden) {
                continue;
            }

            const id = item.id ?? getItemName(item.name);

            const legend: TrendsBarChartLegend = {
                id: id,
                name: item.name ? getItemName(item.name) : id,
                color: getDisplayColor(props.useCustomColor && item.color ? item.color : chartColors.value[index % chartColors.value.length]),
                displayOrders: item.displayOrders ? item.displayOrders : [0]
            };

            legends.push(legend);

            if (unselectedLegends.value[id]) {
                continue;
            }

            const dateRangeItemMap: Record<string, TrendsBarChartDataAmount> = {};

            for (const dataItem of item.items) {
                let dateRangeKey = '';

                if (props.chartMode === 'daily' && 'month' in dataItem) {
                    if (props.dateAggregationType === ChartDateAggregationType.Year.type) {
                        dateRangeKey = dataItem.year.toString();
                    } else if (props.dateAggregationType === ChartDateAggregationType.FiscalYear.type) {
                        const fiscalYear = getFiscalYearFromUnixTime(
                            getYearMonthFirstUnixTime({ year: dataItem.year, month1base: dataItem.month }),
                            props.fiscalYearStart
                        );
                        dateRangeKey = fiscalYear.toString();
                    } else if (props.dateAggregationType === ChartDateAggregationType.Quarter.type) {
                        dateRangeKey = `${dataItem.year}-${Math.floor((dataItem.month - 1) / 3) + 1}`;
                    } else if (props.dateAggregationType === ChartDateAggregationType.Month.type) {
                        dateRangeKey = `${dataItem.year}-${dataItem.month}`;
                    } else { // if (props.dateAggregationType === ChartDateAggregationType.Day.type) {
                        dateRangeKey = `${dataItem.year}-${dataItem.month}-${dataItem.day}`;
                    }
                } else if (props.chartMode === 'monthly' && 'month1base' in dataItem) {
                    if (props.dateAggregationType === ChartDateAggregationType.Year.type) {
                        dateRangeKey = dataItem.year.toString();
                    } else if (props.dateAggregationType === ChartDateAggregationType.FiscalYear.type) {
                        const fiscalYear = getFiscalYearFromUnixTime(
                            getYearMonthFirstUnixTime({ year: dataItem.year, month1base: dataItem.month1base }),
                            props.fiscalYearStart
                        );
                        dateRangeKey = fiscalYear.toString();
                    } else if (props.dateAggregationType === ChartDateAggregationType.Quarter.type) {
                        dateRangeKey = `${dataItem.year}-${Math.floor((dataItem.month1base - 1) / 3) + 1}`;
                    } else { // if (props.dateAggregationType === ChartDateAggregationType.Month.type) {
                        dateRangeKey = `${dataItem.year}-${dataItem.month1base}`;
                    }
                }

                const value = dataItem.value;

                if (dateRangeItemMap[dateRangeKey]) {
                    if (isBigDecimal(value)) {
                        if (props.dataAggregationType === ChartDataAggregationType.Sum) {
                            dateRangeItemMap[dateRangeKey]!.value = dateRangeItemMap[dateRangeKey]!.value.add(value);
                        } else if (props.dataAggregationType === ChartDataAggregationType.Last) {
                            dateRangeItemMap[dateRangeKey]!.value = value;
                        }
                    }
                } else {
                    const allDataItems: TrendsBarChartDataAmount[] = allDateRangeItemsMap[dateRangeKey] || [];
                    const finalDataItem: TrendsBarChartDataAmount = Object.assign({}, legend, {
                        value: isBigDecimal(value) ? value : BIG_DECIMAL_ZERO
                    });

                    allDataItems.push(finalDataItem);
                    dateRangeItemMap[dateRangeKey] = finalDataItem;
                    allDateRangeItemsMap[dateRangeKey] = allDataItems;
                }
            }
        }

        const finalDataItems: TrendsBarChartDataItem[] = [];
        let maxTotalAmount: BigDecimal = BIG_DECIMAL_ZERO;

        for (const dateRange of allDateRanges.value) {
            let dateRangeKey = '';

            if (props.dateAggregationType === ChartDateAggregationType.Year.type) {
                dateRangeKey = dateRange.year.toString();
            } else if (props.dateAggregationType === ChartDateAggregationType.FiscalYear.type) {
                dateRangeKey = dateRange.year.toString();
            } else if (props.dateAggregationType === ChartDateAggregationType.Quarter.type && 'quarter' in dateRange) {
                dateRangeKey = `${dateRange.year}-${dateRange.quarter}`;
            } else if (props.dateAggregationType === ChartDateAggregationType.Month.type && 'month0base' in dateRange) {
                dateRangeKey = `${dateRange.year}-${dateRange.month0base + 1}`;
            } else if (props.dateAggregationType === ChartDateAggregationType.Day.type && 'day' in dateRange && props.chartMode === 'daily') {
                dateRangeKey = `${dateRange.year}-${dateRange.month}-${dateRange.day}`;
            }

            const minDateTime = parseDateTimeFromUnixTime(dateRange.minUnixTime);
            let displayDateRange = '';

            if (props.dateAggregationType === ChartDateAggregationType.Year.type) {
                displayDateRange = formatDateTimeToGregorianLikeShortYear(minDateTime);
            } else if (props.dateAggregationType === ChartDateAggregationType.FiscalYear.type) {
                displayDateRange = formatDateTimeToGregorianLikeFiscalYear(minDateTime);
            } else if (props.dateAggregationType === ChartDateAggregationType.Quarter.type && 'quarter' in dateRange) {
                displayDateRange = formatYearQuarterToGregorianLikeYearQuarter(dateRange.year, dateRange.quarter);
            } else if (props.dateAggregationType === ChartDateAggregationType.Month.type) {
                displayDateRange = formatDateTimeToGregorianLikeShortYearMonth(minDateTime);
            } else if (props.dateAggregationType === ChartDateAggregationType.Day.type && props.chartMode === 'daily') {
                displayDateRange = formatDateTimeToShortDate(minDateTime);
            }

            const dataItems = allDateRangeItemsMap[dateRangeKey] || [];
            let totalAmount: BigDecimal = BIG_DECIMAL_ZERO;
            let totalPositiveAmount = BIG_DECIMAL_ZERO;
            let maxAmount = BIG_DECIMAL_ZERO;

            sortStatisticsItems(dataItems, props.sortingType);

            for (const dataItem of dataItems) {
                if (dataItem.value.isPositive()) {
                    totalPositiveAmount = totalPositiveAmount.add(dataItem.value);
                }

                totalAmount = totalAmount.add(dataItem.value);

                if (dataItem.value.greaterThan(maxAmount)) {
                    maxAmount = dataItem.value;
                }
            }

            if (totalAmount.greaterThan(maxTotalAmount)) {
                maxTotalAmount = totalAmount;
            }

            const finalDataItem: TrendsBarChartDataItem = {
                index: finalDataItems.length,
                dateRange: dateRange,
                displayDateRange: displayDateRange,
                items: dataItems,
                totalAmount: totalAmount,
                totalPositiveAmount: totalPositiveAmount,
                maxAmount: maxAmount,
                percent: 0.0
            };

            finalDataItems.push(finalDataItem);
        }

        for (const finalDataItem of finalDataItems) {
            if (maxTotalAmount.isPositive() && finalDataItem.totalAmount.isPositive()) {
                finalDataItem.percent = finalDataItem.totalAmount.divide(maxTotalAmount).multiply(100).toDoubleNumber();
            } else {
                finalDataItem.percent = 0.0;
            }
        }

        return {
            data: finalDataItems,
            legends: legends
        };
    });

    function clickItem(item: TrendsBarChartDataItem): void {
        let itemId = '';

        for (const item of props.items) {
            if (item.hidden) {
                continue;
            }

            const id = item.id ?? getItemName(item.name);

            if (unselectedLegends.value[id]) {
                continue;
            }

            if (itemId.length) {
                itemId += ',';
            }

            itemId += id;
        }

        const dateRange = item.dateRange;
        let minUnixTime = dateRange.minUnixTime;
        let maxUnixTime = dateRange.maxUnixTime;

        if (props.chartMode === 'daily') {
            if (props.startTime) {
                if (props.startTime > minUnixTime) {
                    minUnixTime = props.startTime;
                }
            }

            if (props.endTime) {
                if (props.endTime < maxUnixTime) {
                    maxUnixTime = props.endTime;
                }
            }
        } else if (props.chartMode === 'monthly') {
            if (props.startYearMonth) {
                const startMinUnixTime = getYearMonthFirstUnixTime(props.startYearMonth);

                if (startMinUnixTime > minUnixTime) {
                    minUnixTime = startMinUnixTime;
                }
            }

            if (props.endYearMonth) {
                const endMaxUnixTime = getYearMonthLastUnixTime(props.endYearMonth);

                if (endMaxUnixTime < maxUnixTime) {
                    maxUnixTime = endMaxUnixTime;
                }
            }
        }

        const dateRangeType = getDateTypeByDateRange(minUnixTime, maxUnixTime, userStore.currentUserFirstDayOfWeek, userStore.currentUserFiscalYearStart, DateRangeScene.Normal);

        emit('click', {
            itemId: itemId,
            dateRange: {
                minTime: minUnixTime,
                maxTime: maxUnixTime,
                dateType: dateRangeType
            }
        });
    }

    function toggleLegend(legend: TrendsBarChartLegend): void {
        if (unselectedLegends.value[legend.id]) {
            delete unselectedLegends.value[legend.id];
        } else {
            unselectedLegends.value[legend.id] = true;
        }
    }

    function renderExternal(vl: unknown, vlData: TrendsBarChartVirtualListData): void {
        virtualDataItems.value = vlData;
    }

    watch(allDisplayDataItems, () => {
        allDisplayDataItemsVersion.value++;
    }, {
        deep: true
    });

    return {
        tt, formatAmountToLocalizedNumeralsWithCurrency, allDisplayDataItemsVersion, unselectedLegends, virtualDataItems,
        useVirtualList, allDisplayDataItems, clickItem, toggleLegend, renderExternal,
    };
}
