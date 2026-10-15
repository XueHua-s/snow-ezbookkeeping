import pluginVue from 'eslint-plugin-vue';
import vueTsEslintConfig from '@vue/eslint-config-typescript';

export default [
    ...pluginVue.configs['flat/essential'],
    ...vueTsEslintConfig(),
    {
        languageOptions: {
            parserOptions: {
                projectService: true,
                // Keep the project service stable when switching between Vue and TS modules.
                extraFileExtensions: ['.vue'],
                tsconfigRootDir: import.meta.dirname,
            }
        },
    },
    {
        ignores: [
            'dist/**',
            '**/*.{js,jsx,cjs,mjs}'
        ]
    },
    {
        files: [
            '**/*.{vue,ts,tsx,mts,js,jsx,cjs,mjs}'
        ],
        rules: {
            'vue/valid-v-slot': ['error', {
                allowModifiers: true
            }]
        }
    },
];
