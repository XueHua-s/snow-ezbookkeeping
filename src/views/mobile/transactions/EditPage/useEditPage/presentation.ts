import { ScheduledTemplateFrequencyType } from '@/core/template.ts';
import {
    TransactionQuickSaveButtonStyle,
    TransactionType
} from '@/core/transaction.ts';
import {
    getTimezoneOffset,
    getTimezoneOffsetMinutes,
    parseDateTimeFromUnixTimeWithTimezoneOffset
} from '@/lib/datetime.ts';
import { TransactionTemplate } from '@/models/transaction_template.ts';
import {
    TransactionEditPageMode,
    TransactionEditPageType
} from '@/views/base/transactions/TransactionEditPageBase.ts';
import { computed } from 'vue';
import type { useEditPageState } from './state.ts';

interface Context {
    settingsStore: ReturnType<typeof useEditPageState>['settingsStore'];
    mode: ReturnType<typeof useEditPageState>['mode'];
    transaction: ReturnType<typeof useEditPageState>['transaction'];
    showTimeInDefaultTimezone: ReturnType<typeof useEditPageState>['showTimeInDefaultTimezone'];
    formatDateTimeToLongDate: ReturnType<typeof useEditPageState>['formatDateTimeToLongDate'];
    formatDateTimeToLongTime: ReturnType<typeof useEditPageState>['formatDateTimeToLongTime'];
    numeralSystem: ReturnType<typeof useEditPageState>['numeralSystem'];
    allTimezones: ReturnType<typeof useEditPageState>['allTimezones'];
    getTransactionPictureUrl: ReturnType<typeof useEditPageState>['getTransactionPictureUrl'];
    pageTypeAndMode: ReturnType<typeof useEditPageState>['pageTypeAndMode'];
    tt: ReturnType<typeof useEditPageState>['tt'];
    getMultiWeekdayLongNames: ReturnType<typeof useEditPageState>['getMultiWeekdayLongNames'];
    firstDayOfWeek: ReturnType<typeof useEditPageState>['firstDayOfWeek'];
    getMultiMonthdayShortNames: ReturnType<typeof useEditPageState>['getMultiMonthdayShortNames'];
    getMultiMonthAndDayLongNames: ReturnType<typeof useEditPageState>['getMultiMonthAndDayLongNames'];
    formatGregorianTextualYearMonthDayToLongDate: ReturnType<typeof useEditPageState>['formatGregorianTextualYearMonthDayToLongDate'];
}

export function useEditPagePresentation(context: Context) {
    const {
        settingsStore, mode, transaction, showTimeInDefaultTimezone, formatDateTimeToLongDate, formatDateTimeToLongTime,
        numeralSystem, allTimezones, getTransactionPictureUrl, pageTypeAndMode, tt, getMultiWeekdayLongNames, firstDayOfWeek,
        getMultiMonthdayShortNames, getMultiMonthAndDayLongNames, formatGregorianTextualYearMonthDayToLongDate,
    } = context;

    const quickSaveButtonStyleType = computed<number>(() => settingsStore.appSettings.quickSaveButtonStyleInMobileTransactionListPage);

    const quickSaveButtonFloatingPosition = computed<string>(() => {
        switch (settingsStore.appSettings.quickSaveButtonStyleInMobileTransactionListPage) {
            case TransactionQuickSaveButtonStyle.BottomLeftFloating.type:
                return 'left-bottom';
            case TransactionQuickSaveButtonStyle.BottomCenterFloating.type:
                return 'center-bottom';
            case TransactionQuickSaveButtonStyle.BottomRightFloating.type:
                return 'right-bottom';
            default:
                return 'right-bottom';
        }
    });

    const sourceAmountClass = computed<Record<string, boolean>>(() => {
        const classes: Record<string, boolean> = {
            'readonly': mode.value === TransactionEditPageMode.View,
            'text-expense': transaction.value.type === TransactionType.Expense,
            'text-income': transaction.value.type === TransactionType.Income,
            'text-color-primary': transaction.value.type === TransactionType.Transfer
        };

        classes[getFontClassByAmount(transaction.value.sourceAmount)] = true;

        return classes;
    });

    const destinationAmountClass = computed<Record<string, boolean>>(() => {
        const classes: Record<string, boolean> = {
            'readonly': mode.value === TransactionEditPageMode.View
        };

        classes[getFontClassByAmount(transaction.value.destinationAmount)] = true;

        return classes;
    });

    const transactionDisplayDate = computed<string>(() => {
        if (mode.value !== TransactionEditPageMode.View || !showTimeInDefaultTimezone.value) {
            const dateTime = parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.value.time, transaction.value.utcOffset);
            return formatDateTimeToLongDate(dateTime);
        }

        const dateTime = parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.value.time, getTimezoneOffsetMinutes(transaction.value.time));
        return formatDateTimeToLongDate(dateTime);
    });

    const transactionDisplayTime = computed<string>(() => {
        if (mode.value !== TransactionEditPageMode.View || !showTimeInDefaultTimezone.value) {
            const dateTime = parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.value.time, transaction.value.utcOffset);
            return formatDateTimeToLongTime(dateTime);
        }

        const dateTime = parseDateTimeFromUnixTimeWithTimezoneOffset(transaction.value.time, getTimezoneOffsetMinutes(transaction.value.time));
        const utcOffset = numeralSystem.value.replaceWesternArabicDigitsToLocalizedDigits(getTimezoneOffset(transaction.value.time));
        return `${formatDateTimeToLongTime(dateTime)} (UTC${utcOffset})`;
    });

    const transactionDisplayTimezoneName = computed<string>(() => {
        for (const timezone of allTimezones.value) {
            if (timezone.name === transaction.value.timeZone) {
                return timezone.displayName;
            }
        }

        return '';
    });

    const transactionPictures = computed<Record<string, string | undefined>[]>(() => {
        const thumbs: Record<string, string | undefined>[] = [];

        if (!transaction.value.pictures || !transaction.value.pictures.length) {
            return thumbs;
        }

        for (const picture of transaction.value.pictures) {
            thumbs.push({
                url: getTransactionPictureUrl(picture)
            });
        }

        return thumbs;
    });

    const transactionThumbs = computed<(string | undefined)[]>(() => {
        const thumbs: (string | undefined)[] = [];

        if (!transaction.value.pictures || !transaction.value.pictures.length) {
            return thumbs;
        }

        for (const picture of transaction.value.pictures) {
            thumbs.push(getTransactionPictureUrl(picture));
        }

        return thumbs;
    });

    const transactionDisplayScheduledFrequency = computed<string>(() => {
        if (pageTypeAndMode?.type !== TransactionEditPageType.Template) {
            return '';
        }

        const template = transaction.value as TransactionTemplate;

        if (template.scheduledFrequencyType === ScheduledTemplateFrequencyType.Disabled.type) {
            return tt('Disabled');
        }

        const items = (template.scheduledFrequency || '').split(',');
        const scheduledFrequencyValues: number[] = [];

        for (const item of items) {
            if (item) {
                scheduledFrequencyValues.push(parseInt(item));
            }
        }

        if (template.scheduledFrequencyType === ScheduledTemplateFrequencyType.Daily.type) {
            return tt('Daily');
        } else if (template.scheduledFrequencyType === ScheduledTemplateFrequencyType.EveryNDays.type) {
            if (scheduledFrequencyValues.length) {
                return tt('format.misc.everyNDays', {
                    n: scheduledFrequencyValues[0]
                });
            } else {
                return tt('Every N Days');
            }
        } else if (template.scheduledFrequencyType === ScheduledTemplateFrequencyType.Weekly.type) {
            if (scheduledFrequencyValues.length) {
                return tt('format.misc.everyMultiDaysOfWeek', {
                    days: getMultiWeekdayLongNames(scheduledFrequencyValues, firstDayOfWeek.value)
                });
            } else {
                return tt('Weekly');
            }
        } else if (template.scheduledFrequencyType === ScheduledTemplateFrequencyType.Monthly.type) {
            if (scheduledFrequencyValues.length) {
                return tt('format.misc.everyMultiDaysOfMonth', {
                    days: getMultiMonthdayShortNames(scheduledFrequencyValues)
                });
            } else {
                return tt('Monthly');
            }
        } else if (template.scheduledFrequencyType === ScheduledTemplateFrequencyType.Yearly.type) {
            if (scheduledFrequencyValues.length) {
                return tt('format.misc.everyMultiDaysOfYear', {
                    days: getMultiMonthAndDayLongNames(scheduledFrequencyValues)
                });
            } else {
                return tt('Yearly');
            }
        } else {
            return '';
        }
    });

    const transactionDisplayScheduledStartDate = computed<string>(() => {
        if (pageTypeAndMode?.type !== TransactionEditPageType.Template) {
            return '';
        }

        const template = transaction.value as TransactionTemplate;

        if (template.scheduledStartDate) {
            return formatGregorianTextualYearMonthDayToLongDate(template.scheduledStartDate);
        } else {
            return tt('No limit');
        }
    });

    const transactionDisplayScheduledEndDate = computed<string>(() => {
        if (pageTypeAndMode?.type !== TransactionEditPageType.Template) {
            return '';
        }

        const template = transaction.value as TransactionTemplate;

        if (template.scheduledEndDate) {
            return formatGregorianTextualYearMonthDayToLongDate(template.scheduledEndDate);
        } else {
            return tt('No limit');
        }
    });

    function getFontClassByAmount(amount: number): string {
        if (amount >= 10000000000 || amount <= -10000000000) {
            return 'ebk-extra-small-amount';
        } if (amount >= 100000000 || amount <= -100000000) {
            return 'ebk-small-amount';
        } else if (amount >= 1000000 || amount <= -1000000) {
            return 'ebk-normal-amount';
        } else {
            return 'ebk-large-amount';
        }
    }

    return {
        quickSaveButtonStyleType, quickSaveButtonFloatingPosition, sourceAmountClass, destinationAmountClass,
        transactionDisplayDate, transactionDisplayTime, transactionDisplayTimezoneName, transactionPictures, transactionThumbs,
        transactionDisplayScheduledFrequency, transactionDisplayScheduledStartDate, transactionDisplayScheduledEndDate,
    };
}
