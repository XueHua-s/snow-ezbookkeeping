import type { Coordinate } from '@/core/coordinate.ts';
import type { AccountInfoResponse } from '@/models/account.ts';
import type { TransactionCategoryInfoResponse } from '@/models/transaction_category.ts';
import type { TransactionPictureInfoBasicResponse } from '@/models/transaction_picture_info.ts';
import type { TransactionTagInfoResponse } from '@/models/transaction_tag.ts';


export interface TransactionDraft {
    readonly type?: number;
    readonly categoryId?: string;
    readonly sourceAccountId?: string;
    readonly sourceAmount?: number;
    readonly destinationAccountId?: string;
    readonly destinationAmount?: number;
    readonly hideAmount?: boolean;
    readonly tagIds?: string[];
    readonly pictures?: TransactionPictureInfoBasicResponse[];
    readonly comment?: string;
}

export interface TransactionGeoLocationRequest {
    readonly latitude: number;
    readonly longitude: number;
}

export interface TransactionCreateRequest {
    readonly type: number;
    readonly categoryId: string;
    readonly time: number;
    readonly utcOffset: number;
    readonly sourceAccountId: string;
    readonly destinationAccountId: string;
    readonly sourceAmount: number;
    readonly destinationAmount: number;
    readonly hideAmount: boolean;
    readonly tagIds: string[];
    readonly pictureIds: string[];
    readonly comment: string;
    readonly geoLocation?: TransactionGeoLocationRequest;
    readonly clientSessionId: string;
}

export interface TransactionModifyRequest {
    readonly id: string;
    readonly type: number;
    readonly categoryId: string;
    readonly time: number;
    readonly utcOffset: number;
    readonly sourceAccountId: string;
    readonly destinationAccountId: string;
    readonly sourceAmount: number;
    readonly destinationAmount: number;
    readonly hideAmount: boolean;
    readonly tagIds: string[];
    readonly pictureIds: string[];
    readonly comment: string;
    readonly geoLocation?: TransactionGeoLocationRequest;
}

export interface TransactionBatchUpdateCategoryRequest {
    readonly transactionIds: string[];
    readonly categoryId: string;
}

export interface TransactionBatchUpdateAccountRequest {
    readonly transactionIds: string[];
    readonly accountId: string;
    readonly isDestinationAccount: boolean;
}

export interface TransactionBatchAddTagsRequest {
    readonly transactionIds: string[];
    readonly tagIds: string[];
}

export interface TransactionBatchRemoveTagsRequest {
    readonly transactionIds: string[];
    readonly tagIds: string[];
}

export interface TransactionBatchClearTagsRequest {
    readonly transactionIds: string[];
}

export interface TransactionMoveBetweenAccountsRequest {
    readonly fromAccountId: string;
    readonly toAccountId: string;
}

export interface TransactionDeleteRequest {
    readonly id: string;
}

export interface TransactionBatchDeleteRequest {
    readonly ids: string[];
    readonly password: string;
}

export interface TransactionImportRequest {
    readonly transactions: TransactionCreateRequest[];
    readonly clientSessionId: string;
}

export interface TransactionListByMaxTimeRequest {
    readonly maxTime: number;
    readonly minTime: number;
    readonly count: number;
    readonly page: number;
    readonly withCount: boolean;
    readonly type: number;
    readonly categoryIds: string;
    readonly accountIds: string;
    readonly tagFilter: string;
    readonly amountFilter: string;
    readonly keyword: string;
    readonly matchMode: number;
    readonly mustHavePictures?: boolean;
    readonly withPictures?: boolean;
}

export interface TransactionListInMonthByPageRequest {
    readonly year: number;
    readonly month: number; // 1-based (1 = January, 12 = December)
    readonly type: number;
    readonly categoryIds: string;
    readonly accountIds: string;
    readonly tagFilter: string;
    readonly amountFilter: string;
    readonly keyword: string;
    readonly matchMode: number;
    readonly mustHavePictures?: boolean;
    readonly withPictures?: boolean;
}

export interface TransactionAllListRequest {
    readonly startTime: number;
    readonly endTime: number;
    readonly withPictures?: boolean;
}

export interface TransactionReconciliationStatementRequest {
    readonly accountId: string;
    readonly startTime: number;
    readonly endTime: number;
}

export type TransactionGeoLocationResponse = Coordinate;

export interface TransactionInfoResponse {
    readonly lastModifiedUnixTime?: number;
    readonly id: string;
    readonly timeSequenceId: string;
    readonly type: number;
    readonly categoryId: string;
    readonly category?: TransactionCategoryInfoResponse;
    readonly time: number;
    readonly utcOffset: number;
    readonly sourceAccountId: string;
    readonly sourceAccount?: AccountInfoResponse;
    readonly destinationAccountId: string;
    readonly destinationAccount?: AccountInfoResponse;
    readonly sourceAmount: number;
    readonly destinationAmount: number;
    readonly hideAmount: boolean;
    readonly tagIds: string[];
    readonly tags?: TransactionTagInfoResponse[];
    readonly pictures?: TransactionPictureInfoBasicResponse[];
    readonly comment: string;
    readonly geoLocation?: TransactionGeoLocationResponse;
    readonly editable: boolean;
}
