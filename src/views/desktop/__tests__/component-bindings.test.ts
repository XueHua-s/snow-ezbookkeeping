import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { compileScript, compileTemplate, parse } from '@vue/compiler-sfc';
import { describe, expect, it } from 'vitest';

describe('extracted page component bindings', () => {
    it.each([
        ['insights/ExplorerPage.vue', 'ExplorerChartTab'],
        ['settings/PreferencesSettingsPage.vue', 'ChartColorSchemeDialog'],
        ['transactions/import/ImportDialog.vue', 'ImportTransactionCheckDataTab'],
        ['transactions/import/ImportDialog.vue', 'ImportTransactionDefineColumnTab'],
        ['transactions/import/ImportDialog.vue', 'ImportTransactionExecuteCustomScriptTab']
    ])('%s renders %s rather than its template ref', (file, component) => {
        const filename = path.resolve('src/views/desktop', file);
        const { descriptor } = parse(readFileSync(filename, 'utf8'), { filename });
        const script = compileScript(descriptor, {
            id: file,
            fs: {
                fileExists: existsSync,
                readFile: file => readFileSync(file, 'utf8')
            }
        });
        const template = compileTemplate({
            id: file,
            filename,
            source: descriptor.template!.content,
            compilerOptions: {
                bindingMetadata: script.bindings,
                expressionPlugins: ['typescript']
            }
        });
        expect(template.errors).toEqual([]);
        expect(template.code).toContain(`$setup["${component}"]`);
    });
});
