import type { ApiResponse } from '@/core/api.ts';
import type {
TransactionCategoryCreateBatchRequest,
TransactionCategoryCreateRequest,
TransactionCategoryDeleteRequest,
TransactionCategoryHideRequest,
TransactionCategoryInfoResponse,
TransactionCategoryModifyRequest,
TransactionCategoryMoveRequest
} from '@/models/transaction_category.ts';
import type {
TransactionTagCreateBatchRequest,
TransactionTagCreateRequest,
TransactionTagDeleteRequest,
TransactionTagHideRequest,
TransactionTagInfoResponse,
TransactionTagModifyRequest,
TransactionTagMoveRequest
} from '@/models/transaction_tag.ts';
import type {
TransactionTagGroupCreateRequest,
TransactionTagGroupDeleteRequest,
TransactionTagGroupInfoResponse,
TransactionTagGroupModifyRequest,
TransactionTagGroupMoveRequest
} from '@/models/transaction_tag_group.ts';
import type {
TransactionTemplateCreateRequest,
TransactionTemplateDeleteRequest,
TransactionTemplateHideRequest,
TransactionTemplateInfoResponse,
TransactionTemplateModifyRequest,
TransactionTemplateMoveRequest
} from '@/models/transaction_template.ts';
import axios from 'axios';

import { type ApiResponsePromise } from './transport.ts';

export const classificationServices = {
    getAllTransactionCategories: (): ApiResponsePromise<Record<number, TransactionCategoryInfoResponse[]>> => {
        return axios.get<ApiResponse<Record<number, TransactionCategoryInfoResponse[]>>>('v1/transaction/categories/list.json');
    },
    getTransactionCategory: ({ id }: { id: string }): ApiResponsePromise<TransactionCategoryInfoResponse> => {
        return axios.get<ApiResponse<TransactionCategoryInfoResponse>>('v1/transaction/categories/get.json?id=' + id);
    },
    addTransactionCategory: (req: TransactionCategoryCreateRequest): ApiResponsePromise<TransactionCategoryInfoResponse> => {
        return axios.post<ApiResponse<TransactionCategoryInfoResponse>>('v1/transaction/categories/add.json', req);
    },
    addTransactionCategoryBatch: (req: TransactionCategoryCreateBatchRequest): ApiResponsePromise<Record<number, TransactionCategoryInfoResponse[]>> => {
        return axios.post<ApiResponse<Record<number, TransactionCategoryInfoResponse[]>>>('v1/transaction/categories/add_batch.json', req);
    },
    modifyTransactionCategory: (req: TransactionCategoryModifyRequest): ApiResponsePromise<TransactionCategoryInfoResponse> => {
        return axios.post<ApiResponse<TransactionCategoryInfoResponse>>('v1/transaction/categories/modify.json', req);
    },
    hideTransactionCategory: (req: TransactionCategoryHideRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transaction/categories/hide.json', req);
    },
    moveTransactionCategory: (req: TransactionCategoryMoveRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transaction/categories/move.json', req);
    },
    deleteTransactionCategory: (req: TransactionCategoryDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transaction/categories/delete.json', req);
    },
    getAllTransactionTagGroups: (): ApiResponsePromise<TransactionTagGroupInfoResponse[]> => {
        return axios.get<ApiResponse<TransactionTagInfoResponse[]>>('v1/transaction/tags/groups/list.json');
    },
    getTransactionTagGroup: ({ id }: { id: string }): ApiResponsePromise<TransactionTagGroupInfoResponse> => {
        return axios.get<ApiResponse<TransactionTagInfoResponse>>('v1/transaction/tags/groups/get.json?id=' + id);
    },
    addTransactionTagGroup: (req: TransactionTagGroupCreateRequest): ApiResponsePromise<TransactionTagGroupInfoResponse> => {
        return axios.post<ApiResponse<TransactionTagInfoResponse>>('v1/transaction/tags/groups/add.json', req);
    },
    modifyTransactionTagGroup: (req: TransactionTagGroupModifyRequest): ApiResponsePromise<TransactionTagGroupInfoResponse> => {
        return axios.post<ApiResponse<TransactionTagInfoResponse>>('v1/transaction/tags/groups/modify.json', req);
    },
    moveTransactionTagGroup: (req: TransactionTagGroupMoveRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transaction/tags/groups/move.json', req);
    },
    deleteTransactionTagGroup: (req: TransactionTagGroupDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transaction/tags/groups/delete.json', req);
    },
    getAllTransactionTags: (): ApiResponsePromise<TransactionTagInfoResponse[]> => {
        return axios.get<ApiResponse<TransactionTagInfoResponse[]>>('v1/transaction/tags/list.json');
    },
    getTransactionTag: ({ id }: { id: string }): ApiResponsePromise<TransactionTagInfoResponse> => {
        return axios.get<ApiResponse<TransactionTagInfoResponse>>('v1/transaction/tags/get.json?id=' + id);
    },
    addTransactionTag: (req: TransactionTagCreateRequest): ApiResponsePromise<TransactionTagInfoResponse> => {
        return axios.post<ApiResponse<TransactionTagInfoResponse>>('v1/transaction/tags/add.json', req);
    },
    addTransactionTagBatch: (req: TransactionTagCreateBatchRequest): ApiResponsePromise<TransactionTagInfoResponse[]> => {
        return axios.post<ApiResponse<TransactionTagInfoResponse[]>>('v1/transaction/tags/add_batch.json', req);
    },
    modifyTransactionTag: (req: TransactionTagModifyRequest): ApiResponsePromise<TransactionTagInfoResponse> => {
        return axios.post<ApiResponse<TransactionTagInfoResponse>>('v1/transaction/tags/modify.json', req);
    },
    hideTransactionTag: (req: TransactionTagHideRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transaction/tags/hide.json', req);
    },
    moveTransactionTag: (req: TransactionTagMoveRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transaction/tags/move.json', req);
    },
    deleteTransactionTag: (req: TransactionTagDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transaction/tags/delete.json', req);
    },
    getAllTransactionTemplates: ({ templateType }: { templateType: number }): ApiResponsePromise<TransactionTemplateInfoResponse[]> => {
        return axios.get<ApiResponse<TransactionTemplateInfoResponse[]>>('v1/transaction/templates/list.json?templateType=' + templateType);
    },
    getTransactionTemplate: ({ id }: { id: string }): ApiResponsePromise<TransactionTemplateInfoResponse> => {
        return axios.get<ApiResponse<TransactionTemplateInfoResponse>>('v1/transaction/templates/get.json?id=' + id);
    },
    addTransactionTemplate: (req: TransactionTemplateCreateRequest): ApiResponsePromise<TransactionTemplateInfoResponse> => {
        return axios.post<ApiResponse<TransactionTemplateInfoResponse>>('v1/transaction/templates/add.json', req);
    },
    modifyTransactionTemplate: (req: TransactionTemplateModifyRequest): ApiResponsePromise<TransactionTemplateInfoResponse> => {
        return axios.post<ApiResponse<TransactionTemplateInfoResponse>>('v1/transaction/templates/modify.json', req);
    },
    hideTransactionTemplate: (req: TransactionTemplateHideRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transaction/templates/hide.json', req);
    },
    moveTransactionTemplate: (req: TransactionTemplateMoveRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transaction/templates/move.json', req);
    },
    deleteTransactionTemplate: (req: TransactionTemplateDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transaction/templates/delete.json', req);
    }
};
