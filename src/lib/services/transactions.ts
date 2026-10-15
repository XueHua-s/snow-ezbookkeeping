import {
    DEFAULT_BATCH_UPDATE_TRANSACTIONS_API_TIMEOUT
} from '@/consts/api.ts';
import type { ApiResponse } from '@/core/api.ts';
import {
    isDefined
} from '@/lib/common.ts';
import type {
TransactionAllListRequest,
TransactionBatchAddTagsRequest,
TransactionBatchClearTagsRequest,
TransactionBatchDeleteRequest,
TransactionBatchRemoveTagsRequest,
TransactionBatchUpdateAccountRequest,
TransactionBatchUpdateCategoryRequest,
TransactionCreateRequest,
TransactionDeleteRequest,
TransactionInfoPageWrapperResponse,
TransactionInfoPageWrapperResponse2,
TransactionInfoResponse,
TransactionListByMaxTimeRequest,
TransactionListInMonthByPageRequest,
TransactionModifyRequest,
TransactionMoveBetweenAccountsRequest,
TransactionReconciliationStatementRequest,
TransactionReconciliationStatementResponse,
TransactionUnreconciledCountRequest,
TransactionUnreconciledCountResponse
} from '@/models/transaction.ts';
import axios from 'axios';

import { type ApiRequestConfig, type ApiResponsePromise } from './transport.ts';

export const transactionsServices = {
    getTransactions: (req: TransactionListByMaxTimeRequest): ApiResponsePromise<TransactionInfoPageWrapperResponse> => {
        const tagFilter = encodeURIComponent(req.tagFilter);
        const amountFilter = encodeURIComponent(req.amountFilter);
        const keyword = encodeURIComponent(req.keyword);
        return axios.get<ApiResponse<TransactionInfoPageWrapperResponse>>(`v1/transactions/list.json?max_time=${req.maxTime}&min_time=${req.minTime}&type=${req.type}&category_ids=${req.categoryIds}&account_ids=${req.accountIds}&tag_filter=${tagFilter}&amount_filter=${amountFilter}&keyword=${keyword}&match_mode=${req.matchMode}&must_have_pictures=${!!req.mustHavePictures}&count=${req.count}&page=${req.page}&with_count=${req.withCount}&with_pictures=${!!req.withPictures}&trim_account=true&trim_category=true&trim_tag=true`);
    },
    getAllTransactionsByMonth: (req: TransactionListInMonthByPageRequest): ApiResponsePromise<TransactionInfoPageWrapperResponse2> => {
        const tagFilter = encodeURIComponent(req.tagFilter);
        const amountFilter = encodeURIComponent(req.amountFilter);
        const keyword = encodeURIComponent(req.keyword);
        return axios.get<ApiResponse<TransactionInfoPageWrapperResponse2>>(`v1/transactions/list/by_month.json?year=${req.year}&month=${req.month}&type=${req.type}&category_ids=${req.categoryIds}&account_ids=${req.accountIds}&tag_filter=${tagFilter}&amount_filter=${amountFilter}&keyword=${keyword}&match_mode=${req.matchMode}&must_have_pictures=${!!req.mustHavePictures}&with_pictures=${!!req.withPictures}&trim_account=true&trim_category=true&trim_tag=true`);
    },
    getAllTransactions: (req: TransactionAllListRequest): ApiResponsePromise<TransactionInfoResponse[]> => {
        return axios.get<ApiResponse<TransactionInfoResponse[]>>(`v1/transactions/list/all.json?trim_account=true&with_pictures=${!!req.withPictures}&trim_category=true&trim_tag=true&start_time=${req.startTime}&end_time=${req.endTime}`);
    },
    getReconciliationStatements: (req: TransactionReconciliationStatementRequest): ApiResponsePromise<TransactionReconciliationStatementResponse> => {
        return axios.get<ApiResponse<TransactionReconciliationStatementResponse>>(`v1/transactions/reconciliation_statements.json?account_id=${req.accountId}&start_time=${req.startTime}&end_time=${req.endTime}`);
    },
    getUnreconciledTransactionCounts: (req: TransactionUnreconciledCountRequest): ApiResponsePromise<TransactionUnreconciledCountResponse> => {
        return axios.get<ApiResponse<TransactionUnreconciledCountResponse>>(`v1/transactions/unreconciled_transaction_counts.json?account_ids=${req.accountIds.join(',')}`);
    },
    getTransaction: ({ id, withPictures }: { id: string, withPictures: boolean | undefined }): ApiResponsePromise<TransactionInfoResponse> => {
        if (!isDefined(withPictures)) {
            withPictures = true;
        }

        return axios.get<ApiResponse<TransactionInfoResponse>>(`v1/transactions/get.json?id=${id}&with_pictures=${withPictures}&trim_account=true&trim_category=true&trim_tag=true`);
    },
    addTransaction: (req: TransactionCreateRequest): ApiResponsePromise<TransactionInfoResponse> => {
        return axios.post<ApiResponse<TransactionInfoResponse>>('v1/transactions/add.json', req);
    },
    modifyTransaction: (req: TransactionModifyRequest): ApiResponsePromise<TransactionInfoResponse> => {
        return axios.post<ApiResponse<TransactionInfoResponse>>('v1/transactions/modify.json', req);
    },
    batchUpdateTransactionCategories: (req: TransactionBatchUpdateCategoryRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transactions/batch_update/category.json', req, {
            timeout: DEFAULT_BATCH_UPDATE_TRANSACTIONS_API_TIMEOUT
        } as ApiRequestConfig);
    },
    batchUpdateTransactionAccounts: (req: TransactionBatchUpdateAccountRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transactions/batch_update/account.json', req, {
            timeout: DEFAULT_BATCH_UPDATE_TRANSACTIONS_API_TIMEOUT
        } as ApiRequestConfig);
    },
    batchAddTagsToTransaction: (req: TransactionBatchAddTagsRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transactions/batch_update/tag/add.json', req, {
            timeout: DEFAULT_BATCH_UPDATE_TRANSACTIONS_API_TIMEOUT
        } as ApiRequestConfig);
    },
    batchRemoveTagsFromTransaction: (req: TransactionBatchRemoveTagsRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transactions/batch_update/tag/remove.json', req, {
            timeout: DEFAULT_BATCH_UPDATE_TRANSACTIONS_API_TIMEOUT
        } as ApiRequestConfig);
    },
    batchClearAllTagsFromTransaction: (req: TransactionBatchClearTagsRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transactions/batch_update/tag/clear.json', req, {
            timeout: DEFAULT_BATCH_UPDATE_TRANSACTIONS_API_TIMEOUT
        } as ApiRequestConfig);
    },
    moveAllTransactionsBetweenAccounts: (req: TransactionMoveBetweenAccountsRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transactions/move/all.json', req);
    },
    deleteTransaction: (req: TransactionDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transactions/delete.json', req);
    },
    batchDeleteTransaction: (req: TransactionBatchDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transactions/batch_delete.json', req, {
            timeout: DEFAULT_BATCH_UPDATE_TRANSACTIONS_API_TIMEOUT
        } as ApiRequestConfig);
    }
};
