import {
    DEFAULT_CLEAR_ALL_TRANSACTIONS_API_TIMEOUT,
    DEFAULT_EXPORT_API_TIMEOUT,
    DEFAULT_UPLOAD_API_TIMEOUT
} from '@/consts/api.ts';
import type { ApiResponse } from '@/core/api.ts';
import type {
ApplicationCloudSetting
} from '@/core/setting.ts';
import type {
ClearAccountTransactionsRequest,
ClearDataRequest,
DataStatisticsResponse,
ExportTransactionDataRequest
} from '@/models/data_management.ts';
import type {
TwoFactorDisableRequest,
TwoFactorEnableConfirmRequest,
TwoFactorEnableConfirmResponse,
TwoFactorEnableResponse,
TwoFactorRegenerateRecoveryCodeRequest,
TwoFactorStatusResponse
} from '@/models/two_factor.ts';
import type {
UserProfileResponse,
UserProfileUpdateRequest,
UserProfileUpdateResponse
} from '@/models/user.ts';
import type {
UserApplicationCloudSettingsUpdateRequest
} from '@/models/user_app_cloud_setting.ts';
import axios,{ type AxiosResponse } from 'axios';

import { type ApiRequestConfig, type ApiResponsePromise } from './transport.ts';

export const usersServices = {
    getProfile: (): ApiResponsePromise<UserProfileResponse> => {
        return axios.get<ApiResponse<UserProfileResponse>>('v1/users/profile/get.json');
    },
    updateProfile: (req: UserProfileUpdateRequest): ApiResponsePromise<UserProfileUpdateResponse> => {
        return axios.post<ApiResponse<UserProfileUpdateResponse>>('v1/users/profile/update.json', req);
    },
    updateAvatar: ({ avatarFile }: { avatarFile: File }): ApiResponsePromise<UserProfileResponse> => {
        return axios.postForm<ApiResponse<UserProfileResponse>>('v1/users/avatar/update.json', {
            avatar: avatarFile
        }, {
            timeout: DEFAULT_UPLOAD_API_TIMEOUT
        });
    },
    removeAvatar: (): ApiResponsePromise<UserProfileResponse> => {
        return axios.post<ApiResponse<UserProfileResponse>>('v1/users/avatar/remove.json');
    },
    resendVerifyEmailByLoginedUser: (): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/users/verify_email/resend.json');
    },
    getUserApplicationCloudSettings: (): ApiResponsePromise<ApplicationCloudSetting[] | false> => {
        return axios.get<ApiResponse<ApplicationCloudSetting[] | false>>('v1/users/settings/cloud/get.json');
    },
    updateUserApplicationCloudSettings: (req: UserApplicationCloudSettingsUpdateRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/users/settings/cloud/update.json', req);
    },
    disableUserApplicationCloudSettings: (): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/users/settings/cloud/disable.json');
    },
    get2FAStatus: (): ApiResponsePromise<TwoFactorStatusResponse> => {
        return axios.get<ApiResponse<TwoFactorStatusResponse>>('v1/users/2fa/status.json');
    },
    enable2FA: (): ApiResponsePromise<TwoFactorEnableResponse> => {
        return axios.post<ApiResponse<TwoFactorEnableResponse>>('v1/users/2fa/enable/request.json');
    },
    confirmEnable2FA: (req: TwoFactorEnableConfirmRequest): ApiResponsePromise<TwoFactorEnableConfirmResponse> => {
        return axios.post<ApiResponse<TwoFactorEnableConfirmResponse>>('v1/users/2fa/enable/confirm.json', req);
    },
    disable2FA: (req: TwoFactorDisableRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/users/2fa/disable.json', req);
    },
    regenerate2FARecoveryCode: (req: TwoFactorRegenerateRecoveryCodeRequest): ApiResponsePromise<TwoFactorEnableConfirmResponse> => {
        return axios.post<ApiResponse<TwoFactorEnableConfirmResponse>>('v1/users/2fa/recovery/regenerate.json', req);
    },
    getUserDataStatistics: (): ApiResponsePromise<DataStatisticsResponse> => {
        return axios.get<ApiResponse<DataStatisticsResponse>>('v1/data/statistics.json');
    },
    getExportedUserData: (fileType: string, req?: ExportTransactionDataRequest): Promise<AxiosResponse<BlobPart>> => {
        let params = '';

        if (req) {
            const tagFilter = encodeURIComponent(req.tagFilter);
            const amountFilter = encodeURIComponent(req.amountFilter);
            const keyword = encodeURIComponent(req.keyword);
            params = `max_time=${req.maxTime}&min_time=${req.minTime}&type=${req.type}&category_ids=${req.categoryIds}&account_ids=${req.accountIds}&tag_filter=${tagFilter}&amount_filter=${amountFilter}&keyword=${keyword}&match_mode=${req.matchMode}`;
        } else {
            params = 'max_time=0&min_time=0&type=0&category_ids=&account_ids=&tag_filter=&amount_filter=&keyword=&match_mode=0';
        }

        if (fileType === 'csv') {
            return axios.get<BlobPart>('v1/data/export.csv?' + params, {
                timeout: DEFAULT_EXPORT_API_TIMEOUT
            } as ApiRequestConfig);
        } else if (fileType === 'tsv') {
            return axios.get<BlobPart>('v1/data/export.tsv?' + params, {
                timeout: DEFAULT_EXPORT_API_TIMEOUT
            } as ApiRequestConfig);
        } else {
            return Promise.reject('Parameter Invalid');
        }
    },
    clearAllData: (req: ClearDataRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/data/clear/all.json', req, {
            ignoreError: true,
            timeout: DEFAULT_CLEAR_ALL_TRANSACTIONS_API_TIMEOUT
        } as ApiRequestConfig);
    },
    clearAllTransactions: (req: ClearDataRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/data/clear/transactions.json', req, {
            timeout: DEFAULT_CLEAR_ALL_TRANSACTIONS_API_TIMEOUT
        } as ApiRequestConfig);
    },
    clearAllTransactionsOfAccount: (req: ClearAccountTransactionsRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/data/clear/transactions/by_account.json', req, {
            timeout: DEFAULT_CLEAR_ALL_TRANSACTIONS_API_TIMEOUT
        } as ApiRequestConfig);
    }
};
