import type { ApiRequestConfig } from '@/lib/services/transport.ts';
import { transactionsServices } from '@/lib/services/transactions.ts';
import type { TransactionInfoResponse } from '@/models/transaction.ts';
import { result, statusCode } from './types.ts';

/** Current ledger data is transient; historical evidence and reading state remain immutable. */
export function createReferenceReader(identity: () => number, unauthorized: (failure: unknown) => void) {
    return async (id: string): Promise<TransactionInfoResponse> => {
        const generation = identity();
        try {
            const transaction = await result(transactionsServices.getTransaction({ id, withPictures: false }, { ignoreError: true } as ApiRequestConfig));
            if (generation !== identity()) throw new DOMException('Assistant identity changed', 'AbortError');
            return transaction;
        } catch (failure) {
            if (generation !== identity()) throw new DOMException('Assistant identity changed', 'AbortError');
            if (statusCode(failure) === 401) unauthorized(failure);
            throw failure;
        }
    };
}
