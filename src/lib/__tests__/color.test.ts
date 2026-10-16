import { APP_THEME } from '@/consts/theme.mts';
import { describe, expect, it } from 'vitest';

import { getContrastTextColor, getContrastIconColor } from '@/lib/color.ts';

describe('getContrastTextColor', () => {
    it('returns black for a light background', () => {
        expect(getContrastTextColor('edddcd')).toBe('000000');
    });

    it('returns white for a dark background', () => {
        expect(getContrastTextColor('112233')).toBe('ffffff');
    });
});

describe('getContrastIconColor', () => {
    it('uses the theme text color for the default light background', () => {
        expect(getContrastIconColor(APP_THEME.light.tint)).toBe(APP_THEME.light.text.slice(1));
    });

    it('uses the theme text color for the default dark background', () => {
        expect(getContrastIconColor(APP_THEME.dark.tint)).toBe(APP_THEME.dark.text.slice(1));
    });
});
