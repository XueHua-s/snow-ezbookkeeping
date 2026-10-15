import type { ApiResponse } from '@/core/api.ts';
import type {
TransactionAmountsRequestParams,
TransactionAmountsResponse,
TransactionDailyAmountsRequest,
TransactionDailyAmountsResponseItem,
TransactionStatisticAssetTrendsRequest,
TransactionStatisticAssetTrendsResponseItem,
TransactionStatisticRequest,
TransactionStatisticResponse,
TransactionStatisticTrendsRequest,
TransactionStatisticTrendsResponseItem
} from '@/models/transaction.ts';
import {
    TransactionAmountsRequest
} from '@/models/transaction.ts';
import axios from 'axios';

import { type ApiResponsePromise } from './transport.ts';

export const statisticsServices = {
    getTransactionStatistics: (req: TransactionStatisticRequest): ApiResponsePromise<TransactionStatisticResponse> => {
        const queryParams: string[] = [];

        if (req.startTime) {
            queryParams.push(`start_time=${req.startTime}`);
        }

        if (req.endTime) {
            queryParams.push(`end_time=${req.endTime}`);
        }

        if (req.tagFilter) {
            queryParams.push(`tag_filter=${encodeURIComponent(req.tagFilter)}`);
        }

        if (req.keyword) {
            queryParams.push(`keyword=${encodeURIComponent(req.keyword)}`);
        }

        if (req.matchMode) {
            queryParams.push(`match_mode=${req.matchMode}`);
        }

        return axios.get<ApiResponse<TransactionStatisticResponse>>(`v1/transactions/statistics.json?use_transaction_timezone=${req.useTransactionTimezone}` + (queryParams.length ? '&' + queryParams.join('&') : ''));
    },
    getTransactionStatisticsTrends: (req: TransactionStatisticTrendsRequest): ApiResponsePromise<TransactionStatisticTrendsResponseItem[]> => {
        const queryParams: string[] = [];

        if (req.startYearMonth) {
            queryParams.push(`start_year_month=${req.startYearMonth}`);
        }

        if (req.endYearMonth) {
            queryParams.push(`end_year_month=${req.endYearMonth}`);
        }

        if (req.tagFilter) {
            queryParams.push(`tag_filter=${encodeURIComponent(req.tagFilter)}`);
        }

        if (req.keyword) {
            queryParams.push(`keyword=${encodeURIComponent(req.keyword)}`);
        }

        if (req.matchMode) {
            queryParams.push(`match_mode=${req.matchMode}`);
        }

        return axios.get<ApiResponse<TransactionStatisticTrendsResponseItem[]>>(`v1/transactions/statistics/trends.json?use_transaction_timezone=${req.useTransactionTimezone}` + (queryParams.length ? '&' + queryParams.join('&') : ''));
    },
    getTransactionStatisticsAssetTrends: (req: TransactionStatisticAssetTrendsRequest): ApiResponsePromise<TransactionStatisticAssetTrendsResponseItem[]> => {
        const queryParams: string[] = [];

        if (req.startTime) {
            queryParams.push(`start_time=${req.startTime}`);
        }

        if (req.endTime) {
            queryParams.push(`end_time=${req.endTime}`);
        }

        return axios.get<ApiResponse<TransactionStatisticAssetTrendsResponseItem[]>>('v1/transactions/statistics/asset_trends.json' + (queryParams.length ? '?' + queryParams.join('&') : ''));
    },
    getTransactionAmounts: (params: TransactionAmountsRequestParams, excludeAccountIds: string[], excludeCategoryIds: string[], tagFilter: string): ApiResponsePromise<TransactionAmountsResponse> => {
        const req = TransactionAmountsRequest.of(params);
        let queryParams = req.buildQuery();

        if (excludeAccountIds && excludeAccountIds.length) {
            queryParams = queryParams + `&exclude_account_ids=${excludeAccountIds.join(',')}`;
        }

        if (excludeCategoryIds && excludeCategoryIds.length) {
            queryParams = queryParams + `&exclude_category_ids=${excludeCategoryIds.join(',')}`;
        }

        if (tagFilter) {
            queryParams = queryParams + `&tag_filter=${encodeURIComponent(tagFilter)}`;
        }

        return axios.get<ApiResponse<TransactionAmountsResponse>>(`v1/transactions/amounts.json?${queryParams}`);
    },
    getTransactionDailyAmounts: (req: TransactionDailyAmountsRequest): ApiResponsePromise<TransactionDailyAmountsResponseItem[]> => {
        const queryParams: string[] = [
            `start_time=${req.startTime}`,
            `end_time=${req.endTime}`,
            `use_transaction_timezone=${req.useTransactionTimezone}`
        ];

        if (req.excludeAccountIds.length) {
            queryParams.push(`exclude_account_ids=${req.excludeAccountIds.join(',')}`);
        }

        if (req.excludeCategoryIds.length) {
            queryParams.push(`exclude_category_ids=${req.excludeCategoryIds.join(',')}`);
        }

        if (req.tagFilter) {
            queryParams.push(`tag_filter=${encodeURIComponent(req.tagFilter)}`);
        }

        return axios.get<ApiResponse<TransactionDailyAmountsResponseItem[]>>(`v1/transactions/amounts/daily.json?${queryParams.join('&')}`);
    }
};
