import { expect, it } from 'vitest';
import { isolatedBackendEnvironment } from './assistant-experience-e2e-environment.mjs';

it('isolates backend fixtures from both direct and file-based host configuration', () => {
    const inherited = Object.freeze({
        PATH: '/synthetic/bin',
        LANG: 'en_US.UTF-8',
        EBK_DATABASE_TYPE: 'mysql',
        EBK_WORK_DIR: '/synthetic/host-workspace',
        EBKCFP_DATABASE_HOST: '/synthetic/database-host',
        EBKCFP_LLM_ASSISTANT_OPENAI_API_KEY: '/synthetic/provider-key'
    });

    expect(isolatedBackendEnvironment(inherited)).toEqual({
        PATH: '/synthetic/bin',
        LANG: 'en_US.UTF-8'
    });
    expect(inherited.EBKCFP_DATABASE_HOST).toBe('/synthetic/database-host');
});
