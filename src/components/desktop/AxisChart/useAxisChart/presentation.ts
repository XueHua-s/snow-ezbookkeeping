import type { SetupProps } from '@/lib/ui/props.ts';
import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';
import { type AxisChartSourceDataItem, ChartValueType } from '@/core/chart.ts';
import type { ColorValue } from '@/core/color.ts';
import type { BigDecimal } from '@/core/numeral.ts';
import { TextDirection } from '@/core/text.ts';
import { ThemeType } from '@/core/theme.ts';
import { useI18n } from '@/locales/helpers.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { computed, ref } from 'vue';
import { useTheme } from 'vuetify';
import { type AxisChartDisplayType } from './types.ts';


interface Context {
    props: SetupProps<{
    class?: string;
    skeleton?: boolean;
    noAnimation?: boolean;
    type: AxisChartDisplayType;
    stacked?: boolean;
    hideLegend?: boolean;
    legendPosition?: 'top' | 'bottom';
    hideXAxisLabels?: boolean;
    hideXAxisLine?: boolean;
    hideYAxisLabels?: boolean;
    hideHorizontalGridLines?: boolean;
    hideLineSymbols?: boolean;
    noMargin?: boolean;
    smoothCurve?: boolean;
    oneHundredPercentStacked?: boolean;
    sortingType: number;
    showValue?: boolean;
    showTotalAmountInTooltip?: boolean;
    totalNameInTooltip?: string;
    categoryTypeName: string;
    allCategoryNames: string[];
    items: AxisChartSourceDataItem[];
    valueType: ChartValueType;
    translateName?: boolean;
    defaultCurrency?: string;
    useCustomColor?: boolean;
    enableClickItem?: boolean;
    tooltipExtraColumnNames?: string[];
    tooltipExtraColumnTotalValues?: (categoryIndex: number, totalValue: BigDecimal, visibleSeriesIds: string[]) => string[];
    tooltipExtraColumnValues?: (seriesId: string, categoryIndex: number, currentValue: BigDecimal) => string[];
}>;
}

export function useAxisChartPresentation(context: Context) {
    const { props } = context;

    const theme = useTheme();

    const {
        tt,
        getCurrentLanguageTextDirection,
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping,
        formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping,
        formatAmountToLocalizedNumeralsWithCurrency,
        formatChartValueToLocalizedNumerals
    } = useI18n();

    const settingsStore = useSettingsStore();

    const selectedLegends = ref<Record<string, boolean>>({});

    const textDirection = computed<TextDirection>(() => getCurrentLanguageTextDirection());

    const isDarkMode = computed<boolean>(() => theme.global.name.value === ThemeType.Dark);

    const chartColors = computed<ColorValue[]>(() => settingsStore.chartColorList);

    const finalClass = computed<string>(() => {
        let finalClass = '';

        if (props.skeleton) {
            finalClass += 'transition-in';
        }

        if (props.class) {
            finalClass += ` ${props.class}`;
        } else {
            finalClass += ' axis-chart-container';
        }

        return finalClass;
    });

    const allItemsMap = computed<Record<string, AxisChartSourceDataItem>>(() => {
        const map: Record<string, AxisChartSourceDataItem> = {};

        for (const item of props.items) {
            map[item.id ?? getItemName(item.name)] = item;
        }

        return map;
    });

    function getItemName(name: string): string {
        return props.translateName ? tt(name) : name;
    }

    function getDisplayValue(value: BigDecimal): string {
        if (props.showValue === false && props.valueType === ChartValueType.Amount) {
            return formatAmountToLocalizedNumeralsWithCurrency(DISPLAY_HIDDEN_AMOUNT, props.defaultCurrency);
        }

        if (props.oneHundredPercentStacked) {
            return formatChartValueToLocalizedNumerals(value, ChartValueType.Percent, props.defaultCurrency);
        } else {
            return formatChartValueToLocalizedNumerals(value, props.valueType, props.defaultCurrency);
        }
    }

    return {
        formatAmountToWesternArabicNumeralsWithoutDigitGrouping, formatBigDecimalToWesternArabicNumeralsWithoutDigitGrouping,
        selectedLegends, textDirection, isDarkMode, chartColors, finalClass, allItemsMap, getItemName, getDisplayValue,
    };
}
