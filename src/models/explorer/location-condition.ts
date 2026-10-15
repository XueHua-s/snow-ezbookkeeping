import type { PartialRecord } from '@/core/base.ts';
import {
    TransactionExplorerConditionFieldType,
    TransactionExplorerConditionOperatorType
} from '@/core/explorer.ts';
import type { TransactionInsightDataItem } from '@/models/transaction.ts';
import type { TransactionExplorerCondition } from './condition-types.ts';

export type GeoLocationConditionOperator = TransactionExplorerConditionOperatorType.IsEmpty |
    TransactionExplorerConditionOperatorType.IsNotEmpty |
    TransactionExplorerConditionOperatorType.LatitudeBetween |
    TransactionExplorerConditionOperatorType.LatitudeNotBetween |
    TransactionExplorerConditionOperatorType.LongitudeBetween |
    TransactionExplorerConditionOperatorType.LongitudeNotBetween;

export class TransactionExplorerGeoLocationCondition implements TransactionExplorerCondition<TransactionExplorerConditionFieldType.GeoLocation, [number, number, number, number]> {
    public static readonly supportedOperators: PartialRecord<TransactionExplorerConditionOperatorType, true> = {
        [TransactionExplorerConditionOperatorType.IsEmpty]: true,
        [TransactionExplorerConditionOperatorType.IsNotEmpty]: true,
        [TransactionExplorerConditionOperatorType.LatitudeBetween]: true,
        [TransactionExplorerConditionOperatorType.LatitudeNotBetween]: true,
        [TransactionExplorerConditionOperatorType.LongitudeBetween]: true,
        [TransactionExplorerConditionOperatorType.LongitudeNotBetween]: true
    };
    public static readonly MIN_LATITUDE: number = -90.0;
    public static readonly MAX_LATITUDE: number = 90.0;
    public static readonly MIN_LONGITUDE: number = -180.0;
    public static readonly MAX_LONGITUDE: number = 180.0;

    public readonly field = TransactionExplorerConditionFieldType.GeoLocation;
    public readonly operator: GeoLocationConditionOperator = TransactionExplorerConditionOperatorType.IsNotEmpty;
    public value: [number, number, number, number];

    constructor(operator: GeoLocationConditionOperator, value: [number, number, number, number]) {
        this.operator = operator;
        this.value = [
            value[0] ?? TransactionExplorerGeoLocationCondition.MIN_LATITUDE,
            value[1] ?? TransactionExplorerGeoLocationCondition.MAX_LATITUDE,
            value[2] ?? TransactionExplorerGeoLocationCondition.MIN_LONGITUDE,
            value[3] ?? TransactionExplorerGeoLocationCondition.MAX_LONGITUDE
        ];
    }

    public getValueForStore(): [number, number, number, number] {
        return [
            this.value[0] ?? TransactionExplorerGeoLocationCondition.MIN_LATITUDE,
            this.value[1] ?? TransactionExplorerGeoLocationCondition.MAX_LATITUDE,
            this.value[2] ?? TransactionExplorerGeoLocationCondition.MIN_LONGITUDE,
            this.value[3] ?? TransactionExplorerGeoLocationCondition.MAX_LONGITUDE
        ];
    }

    public match(transaction: TransactionInsightDataItem): boolean {
        if (this.operator === TransactionExplorerConditionOperatorType.IsEmpty) {
            return !transaction.geoLocation;
        } else if (this.operator === TransactionExplorerConditionOperatorType.IsNotEmpty) {
            return !!transaction.geoLocation;
        } else if (this.operator === TransactionExplorerConditionOperatorType.LatitudeBetween) {
            if (!transaction.geoLocation) {
                return false;
            }

            const latitude = transaction.geoLocation.latitude;

            if (typeof(this.value[0]) === 'number' && latitude < this.value[0]) {
                return false;
            }

            if (typeof(this.value[1]) === 'number' && latitude > this.value[1]) {
                return false;
            }

            return true;
        } else if (this.operator === TransactionExplorerConditionOperatorType.LatitudeNotBetween) {
            if (!transaction.geoLocation) {
                return false;
            }

            const latitude = transaction.geoLocation.latitude;

            if (typeof(this.value[0]) === 'number' && typeof(this.value[1]) === 'number' && latitude >= this.value[0] && latitude <= this.value[1]) {
                return false;
            }

            return true;
        } else if (this.operator === TransactionExplorerConditionOperatorType.LongitudeBetween) {
            if (!transaction.geoLocation) {
                return false;
            }

            const longitude = transaction.geoLocation.longitude;

            if (typeof(this.value[2]) === 'number' && longitude < this.value[2]) {
                return false;
            }

            if (typeof(this.value[3]) === 'number' && longitude > this.value[3]) {
                return false;
            }

            return true;
        } else if (this.operator === TransactionExplorerConditionOperatorType.LongitudeNotBetween) {
            if (!transaction.geoLocation) {
                return false;
            }

            const longitude = transaction.geoLocation.longitude;

            if (typeof(this.value[2]) === 'number' && typeof(this.value[3]) === 'number' && longitude >= this.value[2] && longitude <= this.value[3]) {
                return false;
            }

            return true;
        }

        return false;
    }

    public toExpression(): string {
        if (this.operator === TransactionExplorerConditionOperatorType.IsEmpty) {
            return `geo_location IS EMPTY`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.IsNotEmpty) {
            return `geo_location IS NOT EMPTY`;
        } else if (this.operator === TransactionExplorerConditionOperatorType.LatitudeBetween || this.operator === TransactionExplorerConditionOperatorType.LatitudeNotBetween) {
            const minLatitude: number = this.value[0] ?? TransactionExplorerGeoLocationCondition.MIN_LATITUDE;
            const maxLatitude: number = this.value[1] ?? TransactionExplorerGeoLocationCondition.MAX_LATITUDE;

            if (this.operator === TransactionExplorerConditionOperatorType.LatitudeBetween) {
                return `(geo_location.latitude >= ${minLatitude} AND geo_location.latitude <= ${maxLatitude})`;
            } else if (this.operator === TransactionExplorerConditionOperatorType.LatitudeNotBetween) {
                return `(geo_location.latitude < ${minLatitude} OR geo_location.latitude > ${maxLatitude})`;
            }
        } else if (this.operator === TransactionExplorerConditionOperatorType.LongitudeBetween || this.operator === TransactionExplorerConditionOperatorType.LongitudeNotBetween) {
            const minLongitude: number = this.value[2] ?? TransactionExplorerGeoLocationCondition.MIN_LONGITUDE;
            const maxLongitude: number = this.value[3] ?? TransactionExplorerGeoLocationCondition.MAX_LONGITUDE;

            if (this.operator === TransactionExplorerConditionOperatorType.LongitudeBetween) {
                return `(geo_location.longitude >= ${minLongitude} AND geo_location.longitude <= ${maxLongitude})`;
            } else if (this.operator === TransactionExplorerConditionOperatorType.LongitudeNotBetween) {
                return `(geo_location.longitude < ${minLongitude} OR geo_location.longitude > ${maxLongitude})`;
            }
        }

        return '';
    }
}
