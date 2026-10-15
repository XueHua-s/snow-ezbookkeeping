import {
    TransactionExplorerChartType,
    type TransactionExplorerChartTypeValue,
    type TransactionExplorerConditionRelation,
    TransactionExplorerCustomChartDisplayLayout,
    TransactionExplorerDataDimension,
    type TransactionExplorerDataDimensionType,
    TransactionExplorerValueMetric,
    type TransactionExplorerValueMetricType
} from '@/core/explorer.ts';
import { ChartSortingType } from '@/core/statistics.ts';
import { TimezoneTypeForStatistics } from '@/core/timezone.ts';
import { TransactionExplorerQuery } from './query.ts';

export class InsightsExplorerBasicInfo implements InsightsExplorerInfoResponse {
    public id: string;
    public name: string;
    public displayOrder: number;
    public hidden: boolean;
    public data: Record<string, string | number | object[]> = {};

    private constructor(id: string, name: string, displayOrder: number, hidden: boolean) {
        this.id = id;
        this.name = name;
        this.displayOrder = displayOrder;
        this.hidden = hidden;
    }

    public static of(explorerResponse: InsightsExplorerInfoResponse): InsightsExplorerBasicInfo {
        return new InsightsExplorerBasicInfo(
            explorerResponse.id,
            explorerResponse.name,
            explorerResponse.displayOrder,
            explorerResponse.hidden
        );
    }

    public static ofMulti(explorerResponses: InsightsExplorerInfoResponse[]): InsightsExplorerBasicInfo[] {
        const explorers: InsightsExplorerBasicInfo[] = [];

        for (const explorerResponse of explorerResponses) {
            explorers.push(InsightsExplorerBasicInfo.of(explorerResponse));
        }

        return explorers;
    }
}

export class InsightsExplorer implements InsightsExplorerInfoResponse {
    public id: string;
    public name: string;
    public displayOrder: number;
    public hidden: boolean;
    public queries: TransactionExplorerQuery[];
    public timezoneUsedForDateRange: number;
    public datatableQuerySource: string;
    public countPerPage: number;
    public chartType: TransactionExplorerChartTypeValue;
    public categoryDimension: TransactionExplorerDataDimensionType;
    public seriesDimension: TransactionExplorerDataDimensionType;
    public amountRangeCount: number;
    public valueMetric: TransactionExplorerValueMetricType;
    public chartSortingType: number;
    public customChartDisplayLayout: number;
    public customChartScript: string;

    public static readonly Default: InsightsExplorer = new InsightsExplorer(
        '',
        '',
        0,
        false,
        [],
        TimezoneTypeForStatistics.Default.type,
        '',
        15,
        TransactionExplorerChartType.Default.value,
        TransactionExplorerDataDimension.CategoryDimensionDefault.value,
        TransactionExplorerDataDimension.SeriesDimensionDefault.value,
        5,
        TransactionExplorerValueMetric.Default.value,
        ChartSortingType.Default.type,
        TransactionExplorerCustomChartDisplayLayout.Default.type,
        ''
    );

    private constructor(id: string, name: string, displayOrder: number, hidden: boolean, queries: TransactionExplorerQuery[], timezoneUsedForDateRange: number, datatableQuerySource: string, countPerPage: number, chartType: TransactionExplorerChartTypeValue, categoryDimension: TransactionExplorerDataDimensionType, seriesDimension: TransactionExplorerDataDimensionType, amountRangeCount: number, valueMetric: TransactionExplorerValueMetricType, chartSortingType: number, customChartDisplayLayout: number, customChartScript: string) {
        this.id = id;
        this.name = name;
        this.displayOrder = displayOrder;
        this.hidden = hidden;
        this.queries = queries;
        this.timezoneUsedForDateRange = timezoneUsedForDateRange;
        this.datatableQuerySource = datatableQuerySource;
        this.countPerPage = countPerPage;
        this.chartType = chartType;
        this.categoryDimension = categoryDimension;
        this.seriesDimension = seriesDimension;
        this.amountRangeCount = amountRangeCount;
        this.valueMetric = valueMetric;
        this.chartSortingType = chartSortingType;
        this.customChartDisplayLayout = customChartDisplayLayout;
        this.customChartScript = customChartScript;
    }

    public get data(): Record<string, string | number | object[]> {
        return {
            queries: this.queries.map(q => q.toJsonObject()),
            timezoneUsedForDateRange: this.timezoneUsedForDateRange,
            datatableQuerySource: this.datatableQuerySource,
            countPerPage: this.countPerPage,
            chartType: this.chartType,
            categoryDimension: this.categoryDimension,
            seriesDimension: this.seriesDimension,
            amountRangeCount: this.amountRangeCount,
            valueMetric: this.valueMetric,
            chartSortingType: this.chartSortingType,
            customChartDisplayLayout: this.customChartDisplayLayout,
            customChartScript: this.customChartScript
        };
    }

    public getQueryiesPrettyJson(): string {
        return JSON.stringify(this.queries.map(q => q.toJsonObject()), null, 4);
    }

    public toCreateRequest(clientSessionId: string): InsightsExplorerCreateRequest {
        return {
            name: this.name,
            data: this.data,
            clientSessionId: clientSessionId
        };
    }

    public toModifyRequest(): InsightsExplorerModifyRequest {
        return {
            id: this.id,
            name: this.name,
            data: this.data,
            hidden: this.hidden
        };
    }

    public static of(explorerResponse: InsightsExplorerInfoResponse): InsightsExplorer {
        const data = explorerResponse.data;
        const queries: TransactionExplorerQuery[] = [];
        let timezoneUsedForDateRange = InsightsExplorer.Default.timezoneUsedForDateRange;
        let datatableQuerySource = InsightsExplorer.Default.datatableQuerySource;
        let countPerPage = InsightsExplorer.Default.countPerPage;
        let chartType = InsightsExplorer.Default.chartType;
        let categoryDimension = InsightsExplorer.Default.categoryDimension;
        let seriesDimension = InsightsExplorer.Default.seriesDimension;
        let amountRangeCount = InsightsExplorer.Default.amountRangeCount;
        let valueMetric = InsightsExplorer.Default.valueMetric;
        let chartSortingType = InsightsExplorer.Default.chartSortingType;
        let customChartDisplayLayout = InsightsExplorer.Default.customChartDisplayLayout;
        let customChartScript = InsightsExplorer.Default.customChartScript;
        let hasDatatableQuerySource = false;

        if (data) {
            if (typeof data['timezoneUsedForDateRange'] === 'number') {
                timezoneUsedForDateRange = data['timezoneUsedForDateRange'] as number;
            }

            if (typeof data['datatableQuerySource'] === 'string') {
                datatableQuerySource = data['datatableQuerySource'] as string;
            }

            if (typeof data['countPerPage'] === 'number') {
                countPerPage = data['countPerPage'] as number;
            }

            if (typeof data['chartType'] === 'string') {
                chartType = data['chartType'] as TransactionExplorerChartTypeValue;
            }

            if (typeof data['categoryDimension'] === 'string') {
                categoryDimension = data['categoryDimension'] as TransactionExplorerDataDimensionType;
            }

            if (typeof data['seriesDimension'] === 'string') {
                seriesDimension = data['seriesDimension'] as TransactionExplorerDataDimensionType;
            }

            if (typeof data['amountRangeCount'] === 'number') {
                amountRangeCount = data['amountRangeCount'] as number;
            }

            if (typeof data['valueMetric'] === 'string') {
                valueMetric = data['valueMetric'] as TransactionExplorerValueMetricType;
            }

            if (typeof data['chartSortingType'] === 'number') {
                chartSortingType = data['chartSortingType'] as number;
            }

            if (typeof data['customChartDisplayLayout'] === 'number') {
                customChartDisplayLayout = data['customChartDisplayLayout'] as number;
            }

            if (typeof data['customChartScript'] === 'string') {
                customChartScript = data['customChartScript'] as string;
            }

            if (Array.isArray(data['queries'])) {
                const queryItems = data['queries'] as object[];

                for (const queryItem of queryItems) {
                    const query = TransactionExplorerQuery.parse(queryItem);

                    if (query) {
                        queries.push(query);
                    }

                    if (query && query.id === datatableQuerySource) {
                        hasDatatableQuerySource = true;
                    }
                }
            }

            if (!hasDatatableQuerySource) {
                datatableQuerySource = '';
            }
        }

        return new InsightsExplorer(
            explorerResponse.id,
            explorerResponse.name,
            explorerResponse.displayOrder,
            explorerResponse.hidden,
            queries,
            timezoneUsedForDateRange,
            datatableQuerySource,
            countPerPage,
            chartType,
            categoryDimension,
            seriesDimension,
            amountRangeCount,
            valueMetric,
            chartSortingType,
            customChartDisplayLayout,
            customChartScript
        );
    }

    public static createNewExplorer(newQueryId: string): InsightsExplorer {
        return new InsightsExplorer(
            '',
            '',
            0,
            false,
            [TransactionExplorerQuery.create(newQueryId)],
            InsightsExplorer.Default.timezoneUsedForDateRange,
            InsightsExplorer.Default.datatableQuerySource,
            InsightsExplorer.Default.countPerPage,
            InsightsExplorer.Default.chartType,
            InsightsExplorer.Default.categoryDimension,
            InsightsExplorer.Default.seriesDimension,
            InsightsExplorer.Default.amountRangeCount,
            InsightsExplorer.Default.valueMetric,
            InsightsExplorer.Default.chartSortingType,
            InsightsExplorer.Default.customChartDisplayLayout,
            InsightsExplorer.Default.customChartScript
        );
    }
}

export interface InsightsExplorerCreateRequest {
    readonly name: string;
    readonly data: Record<string, string | number | object[]>;
    readonly clientSessionId?: string;
}

export interface InsightsExplorerModifyRequest {
    readonly id: string;
    readonly name: string;
    readonly data: Record<string, string | number | object[]>;
    readonly hidden: boolean;
    readonly clientSessionId?: string;
}

export interface InsightsExplorerHideRequest {
    readonly id: string;
    readonly hidden: boolean;
}

export interface InsightsExplorerMoveRequest {
    readonly newDisplayOrders: InsightsExplorerNewDisplayOrderRequest[];
}

export interface InsightsExplorerNewDisplayOrderRequest {
    readonly id: string;
    readonly displayOrder: number;
}

export interface InsightsExplorerDeleteRequest {
    readonly id: string;
}

export interface InsightsExplorerInfoResponse {
    readonly id: string;
    readonly name: string;
    readonly displayOrder: number;
    readonly hidden: boolean;
    readonly data: Record<string, string | number | object[]>;
}

export interface ExpressionNode {
    textualExpression: string;
    operator?: TransactionExplorerConditionRelation;
}
