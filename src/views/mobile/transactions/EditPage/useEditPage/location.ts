import logger from '@/lib/logger.ts';
import {
    GeoLocationStatus,
    TransactionEditPageMode
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { useEditPageState } from './state.ts';

interface Context {
    isSupportGeoLocation: ReturnType<typeof useEditPageState>['isSupportGeoLocation'];
    showToast: ReturnType<typeof useEditPageState>['showToast'];
    geoLocationStatus: ReturnType<typeof useEditPageState>['geoLocationStatus'];
    transaction: ReturnType<typeof useEditPageState>['transaction'];
    mode: ReturnType<typeof useEditPageState>['mode'];
    showTimeInDefaultTimezone: ReturnType<typeof useEditPageState>['showTimeInDefaultTimezone'];
    transactionDateTimeSheetMode: ReturnType<typeof useEditPageState>['transactionDateTimeSheetMode'];
    showTransactionDateTimeSheet: ReturnType<typeof useEditPageState>['showTransactionDateTimeSheet'];
}

export function useEditPageLocation(context: Context) {
    const { isSupportGeoLocation, showToast, geoLocationStatus, transaction, mode, showTimeInDefaultTimezone, transactionDateTimeSheetMode, showTransactionDateTimeSheet } = context;

    function updateGeoLocation(forceUpdate: boolean): void {
        if (!isSupportGeoLocation) {
            logger.warn('this browser does not support geo location');

            if (forceUpdate) {
                showToast('Unable to retrieve current position');
            }
            return;
        }

        navigator.geolocation.getCurrentPosition(function (position) {
            if (!position || !position.coords) {
                logger.error('current position is null');
                geoLocationStatus.value = GeoLocationStatus.Error;

                if (forceUpdate) {
                    showToast('Unable to retrieve current position');
                }

                return;
            }

            geoLocationStatus.value = GeoLocationStatus.Success;

            transaction.value.setLatitudeAndLongitude(position.coords.latitude, position.coords.longitude);
        }, function (err) {
            logger.error('cannot retrieve current position', err);
            geoLocationStatus.value = GeoLocationStatus.Error;

            if (forceUpdate) {
                showToast('Unable to retrieve current position');
            }
        });

        geoLocationStatus.value = GeoLocationStatus.Getting;
    }

    function clearGeoLocation(): void {
        geoLocationStatus.value = null;
        transaction.value.removeGeoLocation();
    }

    function showDateTimeDialog(sheetMode: string): void {
        if (mode.value === TransactionEditPageMode.View) {
            showTimeInDefaultTimezone.value = !showTimeInDefaultTimezone.value;
        } else {
            transactionDateTimeSheetMode.value = sheetMode;
            showTransactionDateTimeSheet.value = true;
        }
    }

    return { updateGeoLocation, clearGeoLocation, showDateTimeDialog };
}
