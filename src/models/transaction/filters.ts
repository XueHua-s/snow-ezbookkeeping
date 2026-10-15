import { type Coordinate, getNormalizedCoordinate } from '@/core/coordinate.ts';
import { TransactionTagFilterType } from '@/core/transaction.ts';
import type { TransactionGeoLocationRequest } from './requests.ts';

export class TransactionGeoLocation implements TransactionGeoLocationRequest {
    public latitude: number;
    public longitude: number;

    private constructor(latitude: number, longitude: number) {
        this.latitude = latitude;
        this.longitude = longitude;
    }

    public static createNewGeoLocation(latitude: number, longitude: number): TransactionGeoLocation {
        return new TransactionGeoLocation(latitude, longitude);
    }

    public static of(coordinate: Coordinate): TransactionGeoLocation {
        return new TransactionGeoLocation(coordinate.latitude, coordinate.longitude);
    }

    public toNormalizedCoordinate(): Coordinate {
        return getNormalizedCoordinate(this);
    }
}

export class TransactionTagFilter {
    public readonly tagIds: string[]
    public readonly type: TransactionTagFilterType;

    public static readonly TransactionNoTagFilterValue: string = 'none';

    private constructor(tagIds: string[], type: TransactionTagFilterType) {
        this.tagIds = tagIds;
        this.type = type;
    }

    public static create(tagIds: string[], type: TransactionTagFilterType): TransactionTagFilter {
        return new TransactionTagFilter(tagIds, type);
    }

    public static of(tagId: string): TransactionTagFilter {
        return new TransactionTagFilter([tagId], TransactionTagFilterType.HasAny);
    }

    public static parse(tagFilter: string): TransactionTagFilter[] {
        const ret: TransactionTagFilter[] = [];

        if (!tagFilter || tagFilter === TransactionTagFilter.TransactionNoTagFilterValue) {
            return ret;
        }

        const filters: string[] = tagFilter.split(';');

        for (const filter of filters) {
            const tagFilterItem: string[] = filter.split(':');

            if (tagFilterItem.length !== 2) {
                continue;
            }

            const tagFilterTypeValue: number = parseInt(tagFilterItem[0] as string, 10);

            if (Number.isNaN(tagFilterTypeValue) || !Number.isFinite(tagFilterTypeValue)) {
                continue;
            }

            const tagFilterType: TransactionTagFilterType | undefined = TransactionTagFilterType.parse(tagFilterTypeValue);

            if (!tagFilterType) {
                continue;
            }

            const tagIds: string[] = (tagFilterItem[1] as string).split(',');
            const tagFilter: TransactionTagFilter = new TransactionTagFilter(tagIds, tagFilterType);
            ret.push(tagFilter);
        }

        return ret;
    }

    public static toTextualTagFilters(tagFilters: TransactionTagFilter[]): string {
        const textualTagFilters: string[] = [];

        for (const tagFilter of tagFilters) {
            textualTagFilters.push(tagFilter.toTextualTagFilter());
        }

        return textualTagFilters.join(';');
    }

    public toTextualTagFilter(): string {
        return `${this.type.type}:${this.tagIds.join(',')}`;
    }
}
