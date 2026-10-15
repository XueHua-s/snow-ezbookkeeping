import type { ApiResponse } from '@/core/api.ts';
import type {
AccountCreateRequest,
AccountDeleteRequest,
AccountHideRequest,
AccountInfoResponse,
AccountModifyRequest,
AccountMoveRequest,
AccountUpdateLastReconciledTimeRequest
} from '@/models/account.ts';
import axios from 'axios';

import { type ApiResponsePromise } from './transport.ts';

export const accountsServices = {
    getAllAccounts: ({ visibleOnly }: { visibleOnly: boolean }): ApiResponsePromise<AccountInfoResponse[]> => {
        return axios.get<ApiResponse<AccountInfoResponse[]>>('v1/accounts/list.json?visible_only=' + visibleOnly);
    },
    getAccount: ({ id }: { id: string }): ApiResponsePromise<AccountInfoResponse> => {
        return axios.get<ApiResponse<AccountInfoResponse>>('v1/accounts/get.json?id=' + id);
    },
    addAccount: (req: AccountCreateRequest): ApiResponsePromise<AccountInfoResponse> => {
        return axios.post<ApiResponse<AccountInfoResponse>>('v1/accounts/add.json', req);
    },
    modifyAccount: (req: AccountModifyRequest): ApiResponsePromise<AccountInfoResponse> => {
        return axios.post<ApiResponse<AccountInfoResponse>>('v1/accounts/modify.json', req);
    },
    updateAccountLastReconciledTime: (req: AccountUpdateLastReconciledTimeRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/accounts/update/last_reconciled_time.json', req);
    },
    hideAccount: (req: AccountHideRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/accounts/hide.json', req);
    },
    moveAccount: (req: AccountMoveRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/accounts/move.json', req);
    },
    deleteAccount: (req: AccountDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/accounts/delete.json', req);
    },
    deleteSubAccount: (req: AccountDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/accounts/sub_account/delete.json', req);
    }
};
