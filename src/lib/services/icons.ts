import {
    DEFAULT_UPLOAD_API_TIMEOUT
} from '@/consts/api.ts';
import type { ApiResponse } from '@/core/api.ts';
import type {
UserCustomIconDeleteRequest,
UserCustomIconInfoResponse,
UserCustomIconMoveRequest
} from '@/models/user_custom_icon.ts';
import axios from 'axios';

import { type ApiRequestConfig, type ApiResponsePromise } from './transport.ts';

export const iconsServices = {
    getAllUserCustomIcons: (): ApiResponsePromise<UserCustomIconInfoResponse[]> => {
        return axios.get<ApiResponse<UserCustomIconInfoResponse[]>>('v1/custom_icons/list.json');
    },
    uploadUserCustomIcon: ({ iconFile, clientSessionId }: { iconFile: File, clientSessionId?: string }): ApiResponsePromise<UserCustomIconInfoResponse> => {
        return axios.postForm<ApiResponse<UserCustomIconInfoResponse>>('v1/custom_icons/upload.json', {
            icon: iconFile,
            clientSessionId: clientSessionId
        }, {
            timeout: DEFAULT_UPLOAD_API_TIMEOUT
        } as ApiRequestConfig);
    },
    moveUserCustomIcons: (req: UserCustomIconMoveRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/custom_icons/move.json', req);
    },
    deleteUserCustomIcon: (req: UserCustomIconDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/custom_icons/delete.json', req);
    }
};
