import SnackBar from '@/components/desktop/SnackBar.vue';

export type SnackBarType = InstanceType<typeof SnackBar>;

export type SandboxRequest = {
    source: string;
    transactions: CustomChartTransaction[];
    code: string;
};

export type SandboxResponse = {
    result?: string;
    knownError?: string;
    error?: string;
};

export enum CustomChartAIState {
    Idle = 'idle',
    Prompting = 'prompting',
    Generating = 'generating'
}

export interface CustomChartTransaction {
    id: string;
    unixTime: number;
    year: number;
    month: number;
    dayOfMonth: number;
    dayOfWeek: number;
    hour: number;
    minute: number;
    second: number;
    utcOffset: number;
    type: string;
    primaryCategoryName: string;
    secondaryCategoryName: string;
    sourceAccountName: string;
    destinationAccountName?: string;
    sourceCurrency: string;
    destinationCurrency?: string;
    sourceAmount: number;
    destinationAmount: number;
    sourceAmountInDefaultCurrency?: number;
    destinationAmountInDefaultCurrency?: number;
    tagNames: string[];
    geoLocation?: {
        latitude: number;
        longitude: number
    };
    comment: string;
    inQueries: string[];
}
