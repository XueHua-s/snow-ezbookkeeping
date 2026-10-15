import type { ApiResponse } from '@/core/api.ts';
import type {
InsightsExplorerCreateRequest,
InsightsExplorerDeleteRequest,
InsightsExplorerHideRequest,
InsightsExplorerInfoResponse,
InsightsExplorerModifyRequest,
InsightsExplorerMoveRequest,
} from '@/models/explorer.ts';
import axios from 'axios';

import { type ApiResponsePromise } from './transport.ts';

export const explorersServices = {
    getAllExplorations: (): ApiResponsePromise<InsightsExplorerInfoResponse[]> => {
        return axios.get<ApiResponse<InsightsExplorerInfoResponse[]>>('v1/insights/explorers/list.json');
    },
    getExploration: ({ id }: { id: string }): ApiResponsePromise<InsightsExplorerInfoResponse> => {
        return axios.get<ApiResponse<InsightsExplorerInfoResponse>>('v1/insights/explorers/get.json?id=' + id);
    },
    addExploration: (req: InsightsExplorerCreateRequest): ApiResponsePromise<InsightsExplorerInfoResponse> => {
        return axios.post<ApiResponse<InsightsExplorerInfoResponse>>('v1/insights/explorers/add.json', req);
    },
    modifyExploration: (req: InsightsExplorerModifyRequest): ApiResponsePromise<InsightsExplorerInfoResponse> => {
        return axios.post<ApiResponse<InsightsExplorerInfoResponse>>('v1/insights/explorers/modify.json', req);
    },
    hideExploration: (req: InsightsExplorerHideRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/insights/explorers/hide.json', req);
    },
    moveExploration: (req: InsightsExplorerMoveRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/insights/explorers/move.json', req);
    },
    deleteExploration: (req: InsightsExplorerDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/insights/explorers/delete.json', req);
    }
};
