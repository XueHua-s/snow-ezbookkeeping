# Assistant SSE reverse-proxy verification

Executed 2026-10-15 with `node scripts/assistant-proxy-smoke.mjs`: **PASS**.

The test starts the actual application with a new isolated SQLite database and registered test user, a controlled HTTP model provider, and a real Docker nginx container. It sends one assistant command, opens four SSE subscriptions through nginx, polls the original durable run after a proxy disconnect, and reconnects to that same completed run. This is API/transport evidence; the provider is a deterministic protocol fixture, not a real-model evaluation.

- nginx: **1.28.3**, image `nginx:1.28-alpine`.
- Image digest: `sha256:a8b39bd9cf0f83869a2162827a0caf6137ddf759d50a171451b335cecc87d236`.
- Completed run: `ISOLATED_RUN_1` (sanitized isolated database ID), revision **6**.
- Send commands: **1**. Provider calls: **2** (one tool request and one final answer). Subscriptions, status polling and reconnection made no additional model calls.

| Actual nginx configuration | First body bytes | Stream result |
| --- | ---: | --- |
| `proxy_buffering off; proxy_read_timeout 35s` | 3 ms | One completed terminal, 10,504 ms |
| `proxy_buffering on; proxy_read_timeout 35s`, honoring upstream `X-Accel-Buffering: no` | 20 ms | One completed terminal, 10,522 ms |
| Buffering on, `proxy_ignore_headers X-Accel-Buffering`, 16 KiB buffers, 35s timeout | 10 ms | One completed terminal, 10,511 ms |
| Buffering off, deliberately insufficient `proxy_read_timeout 1s` | 4 ms | Actual connection termination after 2,000 ms; no terminal received |

The insufficient timeout produced an nginx `upstream timed out ... while reading upstream` error. The application continued generating independently. Status polling found the original run completed; reconnection with `afterRevision=6` returned exactly one completed snapshot at revision 6 and then EOF. All successful responses retained `Cache-Control: no-store` and the SSE content type.

The buffering-on test did **not** observe a buffering delay for this small, flushed response. Buffering configuration alone does not guarantee delayed delivery; the report does not infer a delay from the configuration. The test captures `nginx -T` to prove which settings actually ran. Slow TCP readers and byte bounds are tested separately by `TestAssistantSnapshotHTTPSlowReaderBounded` and `TestAssistantSnapshotFrameExactBytes`.

The committed evidence includes [sanitized results](ai-assistant-experience-evidence/proxy/results.json), [nginx input configuration](ai-assistant-experience-evidence/proxy/nginx.conf), and [effective nginx -T output](ai-assistant-experience-evidence/proxy/nginx-active.conf). Only the isolated run ID and ephemeral upstream port were replaced; timings, revisions, call counts and proxy directives are unchanged. No token or application log is included.

The script writes actual config, `nginx -T`, nginx logs, application logs and full timing/state results under `.vitest/assistant-proxy/`. It removes its nginx container and stops its isolated application on completion. Set `EBK_E2E_ARTIFACTS` to retain results elsewhere. It requires Docker and builds the current source unless `EBK_E2E_BINARY` is explicitly supplied.

Deployment settings should disable buffering on the SSE route and allow an idle timeout covering at least two 15-second heartbeat intervals. This controlled local nginx test does not claim coverage of every production CDN or proxy.
