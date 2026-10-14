import { describe, expect, it, vi } from 'vitest';
import { createI18n } from 'vue-i18n';
import type { Composer } from 'vue-i18n';
import en from '@/locales/en.json';
import { ALL_LANGUAGES } from '@/locales/index.ts';

let composer: Composer;
vi.mock('vue-i18n', async importOriginal => ({
    ...await importOriginal<typeof import('vue-i18n')>(),
    useI18n: () => composer
}));
vi.mock('@/stores/setting.ts', () => ({ useSettingsStore: () => ({}) }));
vi.mock('@/stores/user.ts', () => ({ useUserStore: () => ({}) }));
vi.mock('@/stores/exchangeRates.ts', () => ({ useExchangeRatesStore: () => ({}) }));
vi.mock('@/lib/services.ts', () => ({ default: {} }));
import { useI18n } from '@/locales/helpers.ts';

const assistantErrors = [
    'ai assistant request failed',
    'ai assistant execution budget exhausted',
    'ai assistant query timed out',
    'ai assistant continuation is invalid or expired',
    'ai assistant data is not found or not accessible',
    'ai assistant provider does not support this request',
    'ai assistant query arguments are invalid',
    'ai assistant request canceled'
] as const;

describe.each(Object.entries(ALL_LANGUAGES))('assistant errors in %s', (language, info) => {
    it('translates all new backend messages through the real snackbar/toast error helper', () => {
        composer = createI18n({ legacy: false, locale: language, fallbackLocale: 'en', messages: { [language]: info.content as typeof en, en } }).global as Composer;
        const { te } = useI18n();
        const content = info.content as Record<string, string | Record<string, string>>;
        const nested = content['error'] as Record<string, string>;
        assistantErrors.forEach((message, index) => {
            // Desktop SnackBar.showError and mobile showToast call te(message).
            expect(te(message)).toBe(content[message]);
            expect(te(message)).not.toBe(message);
            // Preserve the helper's structured error envelope path as well.
            expect(te({ error: { success: false, errorCode: 215014 + index, errorMessage: message, path: '/api/v1/llm/assistant/chat' } })).toBe(nested[message]);
            expect(nested[message]).toBe(content[message]);
        });
        expect(te('Unable to get AI assistant response')).toBe(content['Unable to get AI assistant response']);
        expect(te('Unable to get AI assistant response')).toBeTruthy();
        if (language === 'zh-Hans' || language === 'zh-Hant') {
            for (const message of assistantErrors) expect(te(message)).toMatch(/[\u3400-\u9fff]/);
        }
    });
});
