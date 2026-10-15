import { UTC_TIMEZONE } from './utc.ts';
import type { TimezoneInfo } from '@/core/timezone.ts';

export const ALL_TIMEZONES_WESTERN: TimezoneInfo[] = [
    // UTC-12:00
    {
        displayName: 'International Date Line West',
        timezoneName: 'Etc/GMT+12'
    },
    // UTC-11:00
    {
        displayName: 'Coordinated Universal Time-11',
        timezoneName: 'Etc/GMT+11'
    },
    // UTC-10:00
    {
        displayName: 'Aleutian Islands',
        timezoneName: 'America/Adak'
    },
    {
        displayName: 'Hawaii',
        timezoneName: 'Pacific/Honolulu'
    },
    // UTC-09:30
    {
        displayName: 'Marquesas Islands',
        timezoneName: 'Pacific/Marquesas'
    },
    // UTC-09:00
    {
        displayName: 'Alaska',
        timezoneName: 'America/Anchorage'
    },
    {
        displayName: 'Coordinated Universal Time-09',
        timezoneName: 'Etc/GMT+9'
    },
    // UTC-08:00
    {
        displayName: 'Baja California',
        timezoneName: 'America/Tijuana'
    },
    {
        displayName: 'Coordinated Universal Time-08',
        timezoneName: 'Etc/GMT+8'
    },
    {
        displayName: 'Pacific Time (US & Canada)',
        timezoneName: 'America/Los_Angeles'
    },
    // UTC-07:00
    {
        displayName: 'Arizona',
        timezoneName: 'America/Phoenix'
    },
    {
        displayName: 'La Paz, Mazatlan',
        timezoneName: 'America/Chihuahua'
    },
    {
        displayName: 'Mountain Time (US & Canada)',
        timezoneName: 'America/Denver'
    },
    {
        displayName: 'Yukon',
        timezoneName: 'America/Whitehorse'
    },
    // UTC-06:00
    {
        displayName: 'Central America',
        timezoneName: 'America/Guatemala'
    },
    {
        displayName: 'Central Time (US & Canada)',
        timezoneName: 'America/Chicago'
    },
    {
        displayName: 'Easter Island',
        timezoneName: 'Pacific/Easter'
    },
    {
        displayName: 'Guadalajara, Mexico City, Monterrey',
        timezoneName: 'America/Mexico_City'
    },
    {
        displayName: 'Saskatchewan',
        timezoneName: 'America/Regina'
    },
    // UTC-05:00
    {
        displayName: 'Bogota, Lima, Quito, Rio Branco',
        timezoneName: 'America/Bogota'
    },
    {
        displayName: 'Chetumal',
        timezoneName: 'America/Cancun'
    },
    {
        displayName: 'Eastern Time (US & Canada)',
        timezoneName: 'America/New_York'
    },
    {
        displayName: 'Haiti',
        timezoneName: 'America/Port-au-Prince'
    },
    {
        displayName: 'Havana',
        timezoneName: 'America/Havana'
    },
    {
        displayName: 'Indiana (East)',
        timezoneName: 'America/Indianapolis'
    },
    {
        displayName: 'Turks and Caicos',
        timezoneName: 'America/Grand_Turk'
    },
    // UTC-04:00
    {
        displayName: 'Asuncion',
        timezoneName: 'America/Asuncion'
    },
    {
        displayName: 'Atlantic Time (Canada)',
        timezoneName: 'America/Halifax'
    },
    {
        displayName: 'Caracas',
        timezoneName: 'America/Caracas'
    },
    {
        displayName: 'Cuiaba',
        timezoneName: 'America/Cuiaba'
    },
    {
        displayName: 'Georgetown, La Paz, Manaus, San Juan',
        timezoneName: 'America/La_Paz'
    },
    {
        displayName: 'Santiago',
        timezoneName: 'America/Santiago'
    },
    // UTC-03:30
    {
        displayName: 'Newfoundland',
        timezoneName: 'America/St_Johns'
    },
    // UTC-03:00
    {
        displayName: 'Araguaina',
        timezoneName: 'America/Araguaina'
    },
    {
        displayName: 'Brasilia',
        timezoneName: 'America/Sao_Paulo'
    },
    {
        displayName: 'Cayenne, Fortaleza',
        timezoneName: 'America/Cayenne'
    },
    {
        displayName: 'City of Buenos Aires',
        timezoneName: 'America/Buenos_Aires'
    },
    {
        displayName: 'Montevideo',
        timezoneName: 'America/Montevideo'
    },
    {
        displayName: 'Punta Arenas',
        timezoneName: 'America/Punta_Arenas'
    },
    {
        displayName: 'Saint Pierre and Miquelon',
        timezoneName: 'America/Miquelon'
    },
    {
        displayName: 'Salvador',
        timezoneName: 'America/Bahia'
    },
    // UTC-02:00
    {
        displayName: 'Coordinated Universal Time-02',
        timezoneName: 'Etc/GMT+2'
    },
    {
        displayName: 'Greenland',
        timezoneName: 'America/Godthab'
    },
    // UTC-01:00
    {
        displayName: 'Azores',
        timezoneName: 'Atlantic/Azores'
    },
    {
        displayName: 'Cabo Verde Is',
        timezoneName: 'Atlantic/Cape_Verde'
    },
    // UTC
    UTC_TIMEZONE,
    // UTC+00:00
    {
        displayName: 'Dublin, Edinburgh, Lisbon, London',
        timezoneName: 'Europe/London'
    },
    {
        displayName: 'Monrovia, Reykjavik',
        timezoneName: 'Atlantic/Reykjavik'
    },
    {
        displayName: 'Sao Tome',
        timezoneName: 'Africa/Sao_Tome'
    },
    // UTC+01:00
    {
        displayName: 'Casablanca',
        timezoneName: 'Africa/Casablanca'
    },
    {
        displayName: 'Amsterdam, Berlin, Bern, Rome, Stockholm, Vienna',
        timezoneName: 'Europe/Berlin'
    },
    {
        displayName: 'Belgrade, Bratislava, Budapest, Ljubljana, Prague',
        timezoneName: 'Europe/Budapest'
    },
    {
        displayName: 'Brussels, Copenhagen, Madrid, Paris',
        timezoneName: 'Europe/Paris'
    },
    {
        displayName: 'Sarajevo, Skopje, Warsaw, Zagreb',
        timezoneName: 'Europe/Warsaw'
    },
    {
        displayName: 'West Central Africa',
        timezoneName: 'Africa/Lagos'
    },
    // UTC+02:00
    {
        displayName: 'Athens, Bucharest',
        timezoneName: 'Europe/Bucharest'
    },
    {
        displayName: 'Beirut',
        timezoneName: 'Asia/Beirut'
    },
];
