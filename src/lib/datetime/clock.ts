import { NumeralSystem } from '@/core/numeral.ts';
// Pure clock formatting shared by the date-time adapter and public conversion helpers.
export function getHourIn12HourFormat(hour: number): number {
    hour = hour % 12;

    if (hour === 0) {
        hour = 12;
    }

    return hour;
}

export function getUtcOffsetByUtcOffsetMinutes(utcOffsetMinutes: number): string {
    const offsetHours = Math.trunc(Math.abs(utcOffsetMinutes) / 60);
    const offsetMinutes = Math.abs(utcOffsetMinutes) - offsetHours * 60;

    const finalOffsetHours = offsetHours.toString().padStart(2, NumeralSystem.WesternArabicNumerals.digitZero);
    const finalOffsetMinutes = offsetMinutes.toString().padStart(2, NumeralSystem.WesternArabicNumerals.digitZero);

    if (utcOffsetMinutes >= 0) {
        return `+${finalOffsetHours}:${finalOffsetMinutes}`;
    } else {
        return `-${finalOffsetHours}:${finalOffsetMinutes}`;
    }
}
