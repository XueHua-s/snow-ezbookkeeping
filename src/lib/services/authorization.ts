import type { ApiResponse } from '@/core/api.ts';
import {
    getCurrentToken
} from '@/lib/userstate.ts';
import type {
AuthResponse,
RegisterResponse
} from '@/models/auth_response.ts';
import type {
ForgetPasswordRequest
} from '@/models/forget_password.ts';
import type {
OAuth2CallbackLoginRequest
} from '@/models/oauth2.ts';
import type {
TokenGenerateAPIRequest,
TokenGenerateAPIResponse,
TokenGenerateMCPRequest,
TokenGenerateMCPResponse,
TokenInfoResponse,
TokenRevokeRequest
} from '@/models/token.ts';
import type {
UserLoginRequest,
UserRegisterRequest,
UserResendVerifyEmailRequest,
UserVerifyEmailResponse
} from '@/models/user.ts';
import type {
UserExternalAuthInfoResponse,
UserExternalAuthUnlinkRequest
} from '@/models/user_external_auth.ts';
import axios from 'axios';

import { type ApiRequestConfig, type ApiResponsePromise } from './transport.ts';

export const authorizationServices = {
    authorize: (data: UserLoginRequest): ApiResponsePromise<AuthResponse> => {
        return axios.post<ApiResponse<AuthResponse>>('authorize.json', data);
    },
    authorize2FA: ({ passcode, token }: { passcode: string, token: string }): ApiResponsePromise<AuthResponse> => {
        return axios.post<ApiResponse<AuthResponse>>('2fa/authorize.json', {
            passcode: passcode
        }, {
            noAuth: true,
            headers: {
                Authorization: `Bearer ${token}`
            }
        } as ApiRequestConfig);
    },
    authorize2FAByBackupCode: ({ recoveryCode, token }: { recoveryCode: string, token: string }): ApiResponsePromise<AuthResponse> => {
        return axios.post<ApiResponse<AuthResponse>>('2fa/recovery.json', {
            recoveryCode: recoveryCode
        }, {
            noAuth: true,
            headers: {
                Authorization: `Bearer ${token}`
            }
        } as ApiRequestConfig);
    },
    authorizeOAuth2: ({ password, passcode, callbackToken }: { password?: string, passcode?: string, callbackToken: string }): ApiResponsePromise<AuthResponse> => {
        const req: OAuth2CallbackLoginRequest = {
            password,
            passcode,
            token: getCurrentToken() || undefined
        };

        return axios.post<ApiResponse<AuthResponse>>('oauth2/authorize.json', req, {
            noAuth: true,
            headers: {
                Authorization: `Bearer ${callbackToken}`
            }
        } as ApiRequestConfig);
    },
    register: (req: UserRegisterRequest): ApiResponsePromise<RegisterResponse> => {
        return axios.post<ApiResponse<RegisterResponse>>('register.json', req);
    },
    verifyEmail: ({ token, requestNewToken }: { token: string, requestNewToken: boolean }): ApiResponsePromise<UserVerifyEmailResponse> => {
        return axios.post<ApiResponse<UserVerifyEmailResponse>>('verify_email/by_token.json?token=' + token, {
            requestNewToken: requestNewToken
        }, {
            noAuth: true,
            ignoreError: true
        } as ApiRequestConfig);
    },
    resendVerifyEmailByUnloginUser: (req: UserResendVerifyEmailRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('verify_email/resend.json', req);
    },
    requestResetPassword: (req: ForgetPasswordRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('forget_password/request.json', req);
    },
    resetPassword: ({ email, token, password }: { email: string, token: string, password: string }): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('forget_password/reset/by_token.json?token=' + token, {
            email: email,
            password: password
        }, {
            noAuth: true,
            ignoreError: true
        } as ApiRequestConfig);
    },
    logout: (): ApiResponsePromise<boolean> => {
        return axios.get<ApiResponse<boolean>>('logout.json');
    },
    getExternalAuths: (): ApiResponsePromise<UserExternalAuthInfoResponse[]> => {
        return axios.get<ApiResponse<UserExternalAuthInfoResponse[]>>('v1/users/external_auth/list.json');
    },
    unlinkExternalAuth: (req: UserExternalAuthUnlinkRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/users/external_auth/unlink.json', req);
    },
    getTokens: (): ApiResponsePromise<TokenInfoResponse[]> => {
        return axios.get<ApiResponse<TokenInfoResponse[]>>('v1/tokens/list.json');
    },
    generateAPIToken: (req: TokenGenerateAPIRequest): ApiResponsePromise<TokenGenerateAPIResponse> => {
        return axios.post<ApiResponse<TokenGenerateAPIResponse>>('v1/tokens/generate/api.json', req);
    },
    generateMCPToken: (req: TokenGenerateMCPRequest): ApiResponsePromise<TokenGenerateMCPResponse> => {
        return axios.post<ApiResponse<TokenGenerateMCPResponse>>('v1/tokens/generate/mcp.json', req);
    },
    revokeToken: ({ tokenId, ignoreError }: { tokenId: string, ignoreError?: boolean }): ApiResponsePromise<boolean> => {
        const req: TokenRevokeRequest = {
            tokenId: tokenId
        };

        return axios.post<ApiResponse<boolean>>('v1/tokens/revoke.json', req, {
            ignoreError: !!ignoreError,
            expectedToken: getCurrentToken()
        } as ApiRequestConfig);
    },
    revokeAllTokens: (): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/tokens/revoke_all.json');
    }
};
