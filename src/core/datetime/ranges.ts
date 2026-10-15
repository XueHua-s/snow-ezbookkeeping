import type { TypeAndName } from '@/core/base.ts';


export enum DateRangeScene {
    Normal = 0,
    TrendAnalysis = 1,
    AssetTrends = 2,
    InsightsExplorer = 3
}

export class DateRange implements TypeAndName {
    private static readonly allInstances: DateRange[] = [];
    private static readonly allInstancesByType: Record<number, DateRange> = {};

    // All date range
    public static readonly All = new DateRange(0, 'All', false, false, false, DateRangeScene.Normal, DateRangeScene.TrendAnalysis, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);

    // Date ranges for normal scene only
    public static readonly Today = new DateRange(1, 'Today', false, false, false, DateRangeScene.Normal, DateRangeScene.InsightsExplorer);
    public static readonly Yesterday = new DateRange(2, 'Yesterday', false, false, false, DateRangeScene.Normal, DateRangeScene.InsightsExplorer);
    public static readonly LastSevenDays = new DateRange(3, 'Recent 7 days', false, false, false, DateRangeScene.Normal, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);
    public static readonly LastThirtyDays = new DateRange(4, 'Recent 30 days', false, false, false, DateRangeScene.Normal, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);
    public static readonly ThisWeek = new DateRange(5, 'This week', false, false, false, DateRangeScene.Normal, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);
    public static readonly LastWeek = new DateRange(6, 'Last week', false, false, false, DateRangeScene.Normal, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);
    public static readonly ThisMonth = new DateRange(7, 'This month', false, false, false, DateRangeScene.Normal, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);
    public static readonly LastMonth = new DateRange(8, 'Last month', false, false, false, DateRangeScene.Normal, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);

    // Date ranges for normal and trend analysis scene
    public static readonly ThisYear = new DateRange(9, 'This year', false, false, false, DateRangeScene.Normal, DateRangeScene.TrendAnalysis, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);
    public static readonly LastYear = new DateRange(10, 'Last year', false, false, false, DateRangeScene.Normal, DateRangeScene.TrendAnalysis, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);
    public static readonly ThisFiscalYear = new DateRange(11, 'This fiscal year', false, false, true, DateRangeScene.Normal, DateRangeScene.TrendAnalysis, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);
    public static readonly LastFiscalYear = new DateRange(12, 'Last fiscal year', false, false, true, DateRangeScene.Normal, DateRangeScene.TrendAnalysis, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);

    // Billing cycle date ranges for normal scene only
    public static readonly CurrentBillingCycle = new DateRange(51, 'Current Billing Cycle', true, false, true, DateRangeScene.Normal);
    public static readonly PreviousBillingCycle = new DateRange(52, 'Previous Billing Cycle', true, false, true, DateRangeScene.Normal);

    // Last reconciled time ranges for normal scene only
    public static readonly SinceLastReconciledTime = new DateRange(71, 'Since Last Reconciled Time', false, true, true, DateRangeScene.Normal)

    // Date ranges for trend analysis scene only
    public static readonly RecentTwelveMonths = new DateRange(101, 'Recent 12 months', false, false, false, DateRangeScene.TrendAnalysis, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);
    public static readonly RecentTwentyFourMonths = new DateRange(102, 'Recent 24 months', false, false, false, DateRangeScene.TrendAnalysis, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);
    public static readonly RecentThirtySixMonths = new DateRange(103, 'Recent 36 months', false, false, false, DateRangeScene.TrendAnalysis, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);
    public static readonly RecentTwoYears = new DateRange(104, 'Recent 2 years', false, false, false, DateRangeScene.TrendAnalysis, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);
    public static readonly RecentThreeYears = new DateRange(105, 'Recent 3 years', false, false, false, DateRangeScene.TrendAnalysis, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);
    public static readonly RecentFiveYears = new DateRange(106, 'Recent 5 years', false, false, false, DateRangeScene.TrendAnalysis, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);

    // Custom date range
    public static readonly Custom = new DateRange(255, 'Custom Date', false, false, true, DateRangeScene.Normal, DateRangeScene.TrendAnalysis, DateRangeScene.AssetTrends, DateRangeScene.InsightsExplorer);

    public readonly type: number;
    public readonly name: string;
    public readonly isBillingCycle: boolean;
    public readonly isLastReconciledTimeRange: boolean;
    public readonly isUserCustomRange: boolean;
    private readonly availableScenes: Record<number, boolean>;

    private constructor(type: number, name: string, isBillingCycle: boolean, isLastReconciledTimeRange: boolean, isUserCustomRange: boolean, ...availableScenes: DateRangeScene[]) {
        this.type = type;
        this.name = name;
        this.isBillingCycle = isBillingCycle;
        this.isLastReconciledTimeRange = isLastReconciledTimeRange;
        this.isUserCustomRange = isUserCustomRange;
        this.availableScenes = {};

        if (availableScenes) {
            for (const scene of availableScenes) {
                this.availableScenes[scene] = true;
            }
        }

        DateRange.allInstances.push(this);
        DateRange.allInstancesByType[type] = this;
    }

    public isAvailableForScene(scene: DateRangeScene): boolean {
        return this.availableScenes[scene] || false;
    }

    public static values(): DateRange[] {
        return DateRange.allInstances;
    }

    public static valueOf(type: number): DateRange | undefined {
        return DateRange.allInstancesByType[type];
    }

    public static isAvailableForScene(type: number, scene: DateRangeScene): boolean {
        const dateRange = DateRange.allInstancesByType[type];
        return dateRange?.isAvailableForScene(scene) || false;
    }

    public static isBillingCycle(type: number): boolean {
        const dateRange = DateRange.allInstancesByType[type];
        return dateRange?.isBillingCycle || false;
    }

    public static isLastReconciledTimeRange(type: number): boolean {
        const dateRange = DateRange.allInstancesByType[type];
        return dateRange?.isLastReconciledTimeRange || false;
    }
}
