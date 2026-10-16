import { configDefaults, defineConfig } from 'vitest/config';
import vue from '@vitejs/plugin-vue';
import path from 'node:path';

export default defineConfig({
    plugins: [vue()],
    resolve: {
        alias: {
            '@': path.resolve(import.meta.dirname, './src')
        }
    },
    test: {
        environment: 'node',
        exclude: [...configDefaults.exclude, '**/*.browser.test.ts'],
        clearMocks: true,
        restoreMocks: true,
        mockReset: true,
        isolate: true,
        coverage: {
            provider: 'v8',
            reporter: ['text', 'html']
        }
    }
})
