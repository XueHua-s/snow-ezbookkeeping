import {
    DEFAULT_API_TIMEOUT
} from '@/consts/api.ts';
import type { ApiResponse } from '@/core/api.ts';
import type {
VersionInfo
} from '@/core/version.ts';
import {
    getExchangeRatesRequestTimeout
} from '@/lib/server_settings.ts';
import type {
LatestExchangeRateResponse,
UserCustomExchangeRateDeleteRequest,
UserCustomExchangeRateUpdateRequest,
UserCustomExchangeRateUpdateResponse
} from '@/models/exchange_rate.ts';
import axios from 'axios';

import { type ApiRequestConfig, type ApiResponsePromise } from './transport.ts';

export const exchangeServices = {
    getLatestExchangeRates: (param: { ignoreError?: boolean }): ApiResponsePromise<LatestExchangeRateResponse> => {
        return axios.get<ApiResponse<LatestExchangeRateResponse>>('v1/exchange_rates/latest.json', {
            ignoreError: !!param.ignoreError,
            timeout: getExchangeRatesRequestTimeout() || DEFAULT_API_TIMEOUT
        } as ApiRequestConfig);
    },
    updateUserCustomExchangeRate: (req: UserCustomExchangeRateUpdateRequest): ApiResponsePromise<UserCustomExchangeRateUpdateResponse> => {
        return axios.post<ApiResponse<UserCustomExchangeRateUpdateResponse>>('v1/exchange_rates/user_custom/update.json', req);
    },
    deleteUserCustomExchangeRate: (req: UserCustomExchangeRateDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/exchange_rates/user_custom/delete.json', req);
    },
    getServerVersion: (): ApiResponsePromise<VersionInfo> => {
        return axios.get<ApiResponse<VersionInfo>>('v1/systems/version.json');
    }
};
