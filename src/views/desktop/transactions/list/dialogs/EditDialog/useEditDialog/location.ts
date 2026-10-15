import type { Coordinate } from '@/core/coordinate.ts';
import logger from '@/lib/logger.ts';
import {
    isSupportGetGeoLocationByClick
} from '@/lib/map/index.ts';
import {
    GeoLocationStatus
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { useEditDialogState } from './state.ts';

interface Context {
    geoMenuState: ReturnType<typeof useEditDialogState>['geoMenuState'];
    isSupportGeoLocation: ReturnType<typeof useEditDialogState>['isSupportGeoLocation'];
    snackbar: ReturnType<typeof useEditDialogState>['snackbar'];
    geoLocationStatus: ReturnType<typeof useEditDialogState>['geoLocationStatus'];
    transaction: ReturnType<typeof useEditDialogState>['transaction'];
    setGeoLocationByClickMap: ReturnType<typeof useEditDialogState>['setGeoLocationByClickMap'];
    map: ReturnType<typeof useEditDialogState>['map'];
}

export function useEditDialogLocation(context: Context) {
    const { geoMenuState, isSupportGeoLocation, snackbar, geoLocationStatus, transaction, setGeoLocationByClickMap, map } = context;

    function updateGeoLocation(forceUpdate: boolean): void {
        geoMenuState.value = false;

        if (!isSupportGeoLocation) {
            logger.warn('this browser does not support geo location');

            if (forceUpdate) {
                snackbar.value?.showMessage('Unable to retrieve current position');
            }
            return;
        }

        navigator.geolocation.getCurrentPosition(function (position) {
            if (!position || !position.coords) {
                logger.error('current position is null');
                geoLocationStatus.value = GeoLocationStatus.Error;

                if (forceUpdate) {
                    snackbar.value?.showMessage('Unable to retrieve current position');
                }

                return;
            }

            geoLocationStatus.value = GeoLocationStatus.Success;

            transaction.value.setLatitudeAndLongitude(position.coords.latitude, position.coords.longitude);
        }, function (err) {
            logger.error('cannot retrieve current position', err);
            geoLocationStatus.value = GeoLocationStatus.Error;

            if (forceUpdate) {
                snackbar.value?.showMessage('Unable to retrieve current position');
            }
        });

        geoLocationStatus.value = GeoLocationStatus.Getting;
    }

    function updateSpecifiedGeoLocation(coordinate: Coordinate): void {
        if (isSupportGetGeoLocationByClick() && setGeoLocationByClickMap.value) {
            transaction.value.setLatitudeAndLongitude(coordinate.latitude, coordinate.longitude);
            map.value?.setMarkerPosition(transaction.value.geoLocation);
        }
    }

    function clearGeoLocation(): void {
        geoMenuState.value = false;
        geoLocationStatus.value = null;
        transaction.value.removeGeoLocation();
    }

    return { updateGeoLocation, updateSpecifiedGeoLocation, clearGeoLocation };
}
