import { ALL_TIMEZONES_WESTERN } from './timezone/western.ts';
import { ALL_TIMEZONES_EASTERN } from './timezone/eastern.ts';
import type { TimezoneInfo } from '@/core/timezone.ts';

export { UTC_TIMEZONE } from './timezone/utc.ts';

// Reference: https://github.com/nodatime/nodatime/blob/main/data/cldr/windowsZones-45.xml
// Reference: https://github.com/mattjohnsonpint/TimeZoneNames/blob/main/src/TimeZoneNames.DataBuilder/data/windows-displaynames.json
export const ALL_TIMEZONES: TimezoneInfo[] = [
    ...ALL_TIMEZONES_WESTERN,
    ...ALL_TIMEZONES_EASTERN,
];

export const WESTERNMOST_TIMEZONE_UTC_OFFSET: number = -720; // Etc/GMT+12 (UTC-12:00)
export const EASTERNMOST_TIMEZONE_UTC_OFFSET: number = 840; // Pacific/Kiritimati (UTC+14:00)
