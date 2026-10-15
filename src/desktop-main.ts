import { createDesktopVuetify } from '@/lib/desktop_vuetify.ts';
import { createPinia } from 'pinia';
import { createApp } from 'vue';
import { type I18n, createI18n } from 'vue-i18n';


import 'vuetify/styles';

import {
    BarChart,
    BoxplotChart,
    CandlestickChart,
    HeatmapChart,
    LineChart,
    PieChart,
    RadarChart,
    SankeyChart,
    ScatterChart,
    SunburstChart,
    TreemapChart
} from 'echarts/charts';
import {
    CalendarComponent,
    GridComponent,
    LegendComponent,
    TooltipComponent,
    VisualMapComponent
} from 'echarts/components';
import * as echarts from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import VChart from 'vue-echarts';

import 'line-awesome/dist/line-awesome/css/line-awesome.css';

import { VueDatePicker } from '@vuepic/vue-datepicker';
import '@vuepic/vue-datepicker/dist/main.css';

import draggable from 'vuedraggable';

import router from '@/router/desktop.ts';

import { getI18nOptions } from '@/locales/helpers.ts';

import DateTimePicker from '@/components/common/DateTimePicker.vue';
import MapView from '@/components/common/MapView.vue';
import MonthPicker from '@/components/common/MonthPicker.vue';
import PinCodeInput from '@/components/common/PinCodeInput.vue';
import TransactionCalendar from '@/components/common/TransactionCalendar.vue';

import AccountAndCategorySankeyChart from '@/components/desktop/AccountAndCategorySankeyChart.vue';
import AccountBalanceTrendsChart from '@/components/desktop/AccountBalanceTrendsChart.vue';
import AmountFilterInput from '@/components/desktop/AmountFilterInput.vue';
import AmountInput from '@/components/desktop/AmountInput.vue';
import AmountInputDialog from '@/components/desktop/AmountInputDialog.vue';
import AuthIllustration from '@/components/desktop/AuthIllustration.vue';
import AxisChart from '@/components/desktop/AxisChart.vue';
import BtnVerticalGroup from '@/components/desktop/BtnVerticalGroup.vue';
import CalendarHeatMapChart from '@/components/desktop/CalendarHeatMapChart.vue';
import CodeEditor from '@/components/desktop/CodeEditor.vue';
import ColorSelect from '@/components/desktop/ColorSelect.vue';
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import CurrencySelect from '@/components/desktop/CurrencySelect.vue';
import CustomChart from '@/components/desktop/CustomChart.vue';
import DataExportDialog from '@/components/desktop/DataExportDialog.vue';
import DateRangeCalendarHeatMapChart from '@/components/desktop/DateRangeCalendarHeatMapChart.vue';
import DateRangeSelectionDialog from '@/components/desktop/DateRangeSelectionDialog.vue';
import DateSelect from '@/components/desktop/DateSelect.vue';
import DateTimeSelect from '@/components/desktop/DateTimeSelect.vue';
import FiscalYearStartSelect from '@/components/desktop/FiscalYearStartSelect.vue';
import HeatMapChart from '@/components/desktop/HeatMapChart.vue';
import HierarchyChart from '@/components/desktop/HierarchyChart.vue';
import IconSelect from '@/components/desktop/IconSelect.vue';
import ItemIcon from '@/components/desktop/ItemIcon.vue';
import JsonExportDialog from '@/components/desktop/JsonExportDialog.vue';
import JsonImportDialog from '@/components/desktop/JsonImportDialog.vue';
import LanguageSelect from '@/components/desktop/LanguageSelect.vue';
import LanguageSelectButton from '@/components/desktop/LanguageSelectButton.vue';
import MainPageLayout from '@/components/desktop/MainPageLayout.vue';
import MonthlyIncomeAndExpenseChart from '@/components/desktop/MonthlyIncomeAndExpenseChart.vue';
import MonthRangeSelectionDialog from '@/components/desktop/MonthRangeSelectionDialog.vue';
import MonthSelectionDialog from '@/components/desktop/MonthSelectionDialog.vue';
import NumberInput from '@/components/desktop/NumberInput.vue';
import OneColumnDialogLayout from '@/components/desktop/OneColumnDialogLayout.vue';
import PieChartComponent from '@/components/desktop/PieChart.vue';
import RadarChartComponent from '@/components/desktop/RadarChart.vue';
import RenameDialog from '@/components/desktop/RenameDialog.vue';
import ScheduleFrequencySelect from '@/components/desktop/ScheduleFrequencySelect.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import StepsBar from '@/components/desktop/StepsBar.vue';
import SwitchToMobileDialog from '@/components/desktop/SwitchToMobileDialog.vue';
import ToggleButton from '@/components/desktop/ToggleButton.vue';
import TransactionTagAutoComplete from '@/components/desktop/TransactionTagAutoComplete.vue';
import TrendsChart from '@/components/desktop/TrendsChart.vue';
import TwoColumnDialogLayout from '@/components/desktop/TwoColumnDialogLayout.vue';
import TwoColumnSelect from '@/components/desktop/TwoColumnSelect.vue';

import TextFieldAutoWidth from '@/directives/desktop/textfieldAutoWidth.ts';

import '@/styles/desktop/amount-color.scss';
import '@/styles/desktop/index.scss';

import App from './DesktopApp.vue';

const app = createApp(App);
const pinia = createPinia();
const i18n = createI18n(getI18nOptions()) as I18n<Record<string, unknown>, Record<string, unknown>, Record<string, unknown>, string, false>;
const vuetify = createDesktopVuetify(i18n.global);

echarts.use([
    CanvasRenderer,
    LineChart,
    BarChart,
    PieChart,
    ScatterChart,
    BoxplotChart,
    CandlestickChart,
    RadarChart,
    TreemapChart,
    SunburstChart,
    HeatmapChart,
    SankeyChart,
    GridComponent,
    CalendarComponent,
    TooltipComponent,
    LegendComponent,
    VisualMapComponent
]);

app.use(pinia);
app.use(i18n);
app.use(vuetify);
app.use(router);

app.component('VChart', VChart);
app.component('VueDatePicker', VueDatePicker);
app.component('DraggableList', draggable);

app.component('PinCodeInput', PinCodeInput);
app.component('MapView', MapView);
app.component('DateTimePicker', DateTimePicker);
app.component('MonthPicker', MonthPicker);
app.component('TransactionCalendar', TransactionCalendar);

app.component('MainPageLayout', MainPageLayout);
app.component('OneColumnDialogLayout', OneColumnDialogLayout);
app.component('TwoColumnDialogLayout', TwoColumnDialogLayout);
app.component('AuthIllustration', AuthIllustration);
app.component('ItemIcon', ItemIcon);
app.component('ToggleButton', ToggleButton);
app.component('BtnVerticalGroup', BtnVerticalGroup);
app.component('NumberInput', NumberInput);
app.component('AmountInput', AmountInput);
app.component('AmountInputDialog', AmountInputDialog);
app.component('AmountFilterInput', AmountFilterInput);
app.component('LanguageSelect', LanguageSelect);
app.component('LanguageSelectButton', LanguageSelectButton);
app.component('CurrencySelect', CurrencySelect);
app.component('DateTimeSelect', DateTimeSelect);
app.component('DateSelect', DateSelect);
app.component('FiscalYearStartSelect', FiscalYearStartSelect);
app.component('ColorSelect', ColorSelect);
app.component('IconSelect', IconSelect);
app.component('TwoColumnSelect', TwoColumnSelect);
app.component('TransactionTagAutoComplete', TransactionTagAutoComplete);
app.component('ScheduleFrequencySelect', ScheduleFrequencySelect);
app.component('StepsBar', StepsBar);
app.component('ConfirmDialog', ConfirmDialog);
app.component('CodeEditor', CodeEditor);
app.component('DataExportDialog', DataExportDialog);
app.component('JsonImportDialog', JsonImportDialog);
app.component('JsonExportDialog', JsonExportDialog);
app.component('SnackBar', SnackBar);
app.component('PieChart', PieChartComponent);
app.component('RadarChart', RadarChartComponent);
app.component('AxisChart', AxisChart);
app.component('TrendsChart', TrendsChart);
app.component('HierarchyChart', HierarchyChart);
app.component('HeatMapChart', HeatMapChart);
app.component('CalendarHeatMapChart', CalendarHeatMapChart);
app.component('DateRangeCalendarHeatMapChart', DateRangeCalendarHeatMapChart);
app.component('CustomChart', CustomChart);
app.component('RenameDialog', RenameDialog);
app.component('DateRangeSelectionDialog', DateRangeSelectionDialog);
app.component('MonthSelectionDialog', MonthSelectionDialog);
app.component('MonthRangeSelectionDialog', MonthRangeSelectionDialog);
app.component('MonthlyIncomeAndExpenseChart', MonthlyIncomeAndExpenseChart);
app.component('AccountBalanceTrendsChart', AccountBalanceTrendsChart);
app.component('AccountAndCategorySankeyChart', AccountAndCategorySankeyChart);
app.component('SwitchToMobileDialog', SwitchToMobileDialog);

app.directive('TextFieldAutoWidth', TextFieldAutoWidth);

app.mount('#app');
