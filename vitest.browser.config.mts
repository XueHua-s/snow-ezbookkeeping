import { defineConfig } from 'vitest/config';
import vue from '@vitejs/plugin-vue';
import { playwright } from '@vitest/browser-playwright';
import path from 'node:path';

export default defineConfig({
    plugins: [vue()],
    resolve: { alias: { '@': path.resolve(import.meta.dirname, './src') } },
    test: {
        include: ['src/**/*.browser.test.ts'],
        clearMocks: true,
        browser: {
            enabled: true,
            headless: true,
            provider: playwright(),
            instances: [{ browser: 'chromium' }]
        }
    }
});
