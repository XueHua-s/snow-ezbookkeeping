// Opt-in real nginx + real API/SQLite SSE verification. Requires Docker.
import assert from 'node:assert/strict';
import { execFile } from 'node:child_process';
import { promisify } from 'node:util';
import { writeFile, mkdir } from 'node:fs/promises';
import { resolve } from 'node:path';
import { randomUUID } from 'node:crypto';
import { environment } from './assistant-experience-e2e-environment.mjs';

const execute = promisify(execFile);
const exec = (file, args) => execute(file, args, { timeout: 30000, maxBuffer: 4 * 1024 * 1024 });
process.env.EBK_E2E_ARTIFACTS ||= resolve('.vitest/assistant-proxy');
await mkdir(process.env.EBK_E2E_ARTIFACTS, { recursive: true });
const env = await environment({ production: true });
const name = `ebk-assistant-proxy-${randomUUID().slice(0, 8)}`;
const image = 'nginx:1.28-alpine';
let container = false;
try {
    const port = new URL(env.base).port;
    const location = (name, directives) => `location /${name}/ { proxy_pass http://host.lima.internal:${port}/; proxy_http_version 1.1; proxy_set_header Connection ""; proxy_cache off; ${directives} }`;
    const config = `events {}\nhttp { access_log /dev/stdout; error_log /dev/stderr info; server { listen 8080;\n${[
        location('off', 'proxy_buffering off; proxy_read_timeout 35s;'),
        location('on', 'proxy_buffering on; proxy_read_timeout 35s;'),
        location('forced', 'proxy_buffering on; proxy_ignore_headers X-Accel-Buffering; proxy_buffer_size 16k; proxy_buffers 8 16k; proxy_read_timeout 35s;'),
        location('timeout', 'proxy_buffering off; proxy_read_timeout 1s;')
    ].join('\n')}\n} }\n`;
    const configPath = resolve(env.artifacts, 'nginx.conf');
    await writeFile(configPath, config);
    container = true; // Cleanup also covers a partially successful create.
    await exec('docker', ['create', '--pull', 'never', '--name', name, '--label', 'ebk.validation=assistant-experience', '-p', '127.0.0.1::8080', image]);
    await exec('docker', ['cp', configPath, `${name}:/etc/nginx/nginx.conf`]);
    await exec('docker', ['start', name]);
    const { stdout: mapping } = await exec('docker', ['port', name, '8080/tcp']);
    const proxy = `http://${mapping.trim()}`;
    let proxyReady = false;
    for (let attempt = 0; attempt < 100; attempt++) {
        try { proxyReady = (await fetch(proxy + '/off/server_settings.js')).ok; } catch { /* nginx starts asynchronously */ }
        if (proxyReady) break;
        await new Promise(resolve => setTimeout(resolve, 100));
    }
    assert(proxyReady, 'nginx must reach the real application before streaming');
    const identity = await env.user('proxy');
    const apiPrefix = 'v1/llm/assistant/';
    const caps = await env.api(apiPrefix + 'capabilities', undefined, identity.token);
    const before = env.calls();
    const command = { clientRequestId: randomUUID(), requestIssuedAt: new Date().toISOString(), expectedDeletionEpoch: caps.deletionEpoch,
        expectedContextVersion: 0, mode: 'chat', content: 'E2E_SLOW isolated proxy expense query' };
    const accepted = await env.api(apiPrefix + 'messages/send', command, identity.token);
    const started = performance.now();
    const headers = { Authorization: `Bearer ${identity.token}`, 'X-Timezone-Name': 'UTC', 'X-Timezone-Offset': '0' };
    async function subscribe(route, afterRevision = 0) {
        const response = await fetch(`${proxy}/${route}/api/${apiPrefix}runs/stream.json?runId=${accepted.runId}&afterRevision=${afterRevision}`, { headers });
        assert.equal(response.status, 200);
        assert.match(response.headers.get('content-type'), /text\/event-stream/);
        assert.equal(response.headers.get('cache-control'), 'no-store');
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let text = '', firstByteMs, disconnected = false;
        try {
            while (true) {
                const result = await reader.read();
                if (result.done) break;
                firstByteMs ??= Math.round(performance.now() - started);
                text += decoder.decode(result.value, { stream: true });
            }
        } catch { disconnected = true; }
        finally { await reader.cancel().catch(() => {}); }
        const snapshots = text.split(/\n\n/).filter(frame => frame.startsWith('event:')).map(frame => {
            const value = JSON.parse(frame.split('\n').find(line => line.startsWith('data: ')).slice(6));
            assert.equal(value.runId, accepted.runId);
            return { state: value.state, revision: value.revision, terminal: frame.startsWith('event: run.terminal') };
        });
        return { route, firstByteMs, elapsedMs: Math.round(performance.now() - started), disconnected, snapshots };
    }
    const streams = ['off', 'on', 'forced', 'timeout'].map(route => { const pending = subscribe(route); pending.catch(() => {}); return pending; });
    const timeout = await streams[3];
    assert.equal(timeout.disconnected, true, 'nginx 1s idle timeout must actually interrupt the stream');
    assert.equal(timeout.snapshots.filter(snapshot => snapshot.terminal).length, 0);
    // Recovery reads the original durable run. No send/retry request is made.
    let state;
    for (let attempt = 0; attempt < 80; attempt++) {
        state = await env.api(apiPrefix + `runs/get?runId=${accepted.runId}`, undefined, identity.token);
        if (['completed', 'failed', 'incomplete', 'canceled', 'interrupted'].includes(state.state)) break;
        await new Promise(resolve => setTimeout(resolve, 250));
    }
    assert.equal(state.state, 'completed');
    const results = await Promise.all(streams);
    for (const result of results.slice(0, 3)) {
        assert.equal(result.disconnected, false);
        assert.equal(result.snapshots.filter(snapshot => snapshot.terminal).length, 1);
        assert.equal(result.snapshots.at(-1).state, 'completed');
    }
    assert(results[0].firstByteMs < 3000 && results[1].firstByteMs < 3000, 'both supported nginx configurations must flush before provider completion');
    // Buffering on is a real nginx configuration, but does not guarantee a
    // delay for small flushed chunks. Record observed timing, never invent it.
    const reconnected = await subscribe('on', state.revision);
    assert.equal(reconnected.snapshots.length, 1);
    assert.equal(reconnected.snapshots[0].terminal, true);
    assert.equal(reconnected.snapshots[0].revision, state.revision);
    assert.equal(env.calls() - before, 2, 'one tool call + one answer; subscriptions/recovery must not regenerate');
    const { stdout: version, stderr: versionErr } = await exec('docker', ['exec', name, 'nginx', '-v']);
    const { stdout: activeConfig, stderr: configDiagnostics } = await exec('docker', ['exec', name, 'nginx', '-T']);
    await writeFile(resolve(env.artifacts, 'nginx-active.conf'), activeConfig + configDiagnostics);
    assert.match(activeConfig, /proxy_ignore_headers X-Accel-Buffering/);
    const { stdout: inspect } = await exec('docker', ['image', 'inspect', image, '--format', '{{json .RepoDigests}}']);
    const report = { status: 'PASS', nginx: (version + versionErr).trim(), imageDigests: JSON.parse(inspect), productionAPI: true, backend: process.env.EBK_E2E_DOCKER_IMAGE ? `Docker image ${process.env.EBK_E2E_DOCKER_IMAGE}; see docker-runtime.json` : 'current host Go build', database: 'isolated SQLite', provider: 'controlled HTTP protocol fixture', providerCalls: env.calls() - before,
        commandCount: 1, runId: accepted.runId, results, reconnected, completedState: state.state,
        checks: ['real nginx buffering off', 'real nginx buffering on honoring X-Accel-Buffering', 'buffering-on with X-Accel-Buffering ignored (measured timing)', 'actual proxy idle disconnect', 'same-run status polling recovery', 'terminal reconnect afterRevision equal', 'no duplicate model calls'] };
    await writeFile(resolve(env.artifacts, 'results.json'), JSON.stringify(report, null, 2));
    console.log(JSON.stringify(report, null, 2));
} finally {
    try {
    if (container) {
        const logs = await exec('docker', ['logs', name]).catch(() => ({ stdout: '', stderr: '' }));
        await writeFile(resolve(env.artifacts, 'nginx.log'), logs.stdout + logs.stderr);
        await exec('docker', ['rm', '-f', name]);
    }
    } finally { await env.close(); }
}
