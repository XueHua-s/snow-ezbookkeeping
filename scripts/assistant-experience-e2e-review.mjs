import { writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { environment } from './assistant-experience-e2e-environment.mjs';
// Run only for a local independent review; fixture credentials stay in a 0600 file.
const env = await environment({ languagePreview: true });
const identity = await env.user('review');
const statePath = resolve(env.directory, 'review-state.json');
const origin = new URL(env.url).origin;
await writeFile(statePath, JSON.stringify({ cookies: [], origins: [{ origin, localStorage: [
    { name: 'ebk_user_token', value: identity.token }, { name: 'ebk_user_info', value: JSON.stringify(identity.user) }
] }] }), { mode: 0o600 });
console.log(JSON.stringify({ url: env.url, storageStatePath: statePath, desktop: '/desktop.html#/assistant?conversationId=', mobile: '/mobile.html#!/assistant?conversationId=' }));
let stopping = false;
for (const signal of ['SIGTERM', 'SIGINT']) process.on(signal, async () => { if (stopping) return; stopping = true; await env.close(); process.exit(0); });
