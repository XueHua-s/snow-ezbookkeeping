# 16 · Final container E2E and compatibility

2026-10-16, serial child review. No delegation, commit or push. All users,
ledger rows, conversations, memories and provider responses are synthetic.
Raw application logs and failure diagnostics remain under `.vitest`.

## Runtime provenance

Final frontend image:
`sha256:bd5b1c31834ad9d917679b9bb9007edae4b9553201a2af48939f6796d1d339e1`.
The fresh build uses Node22.20.0/npm10.9.3 on the host, Go1.27.1 linux/arm64,
Alpine3.24.2, the actual project entrypoint and UID:GID1000:1000. Chromium is
156.0.8078.4. Runtime containers publish loopback ports and have no host mounts.
The deterministic loopback provider is reached through Colima's
`host.lima.internal`; the Docker bridge retains outbound access.

This remains a validation image using the cached Go/Alpine base, including
extra Go toolchain files. It is **not** a stock Dockerfile release build or
its Node26 frontend stage. Build details are in
[build/build-provenance.json](../ai-assistant-experience-evidence/audit-e2e/build/build-provenance.json).
The previous Gate15 image/provenance was not overwritten.

The only product change in this stage is frontend capability error ownership
in `src/stores/assistant/session.ts` and `src/stores/assistant.ts`. The final
frontend artifact digest is `ba13bcd1e34f10a293a2717b7a48558c524c9ad16340cc07ae9614b971e5eead`.
Both old and new images have the same backend source digest
`4753fce8adb6f0f7a09cae95365840b77fbbe82f183b9018edb431eff61dce7a`
and exact Linux executable SHA256
`e5af7ae3ab00bf3930964e0f7c02aacd5b12dca41910effa15ecb967c56c5f63`.
Thus backend-only proxy/platform/upstream-cancel results below retain their
actual original image ID `sha256:990c1f23742049c78c43832cdde54e69a575412675524f4a8ac88d4b80bc0dff`;
they are not relabelled as executions of the new image. Parent independently
verified both equal backend hashes.

## Defect found by real compatibility testing

The new production frontend connected to a task-only build of old main
`1a21bc9bfd2b155a3ed9b8ac57caff7c76bbb293`. Real HTTP capabilities404 produced
a generic request error instead of the required server-upgrade notice.
This was neither a fake response nor an old-backend SPA200 fallback.

Initial native `pageshow` foreground synchronization can supersede the
`initialize` capability ticket. The obsolete initialize failure was correctly
ignored, but compatibility detection existed only in its catch. The current
synchronize failure never set `legacyBackend`. Classification now belongs to
`session.checkIdentity`, after its existing current ticket/epoch/disposal
fence. `initialize` no longer misclassifies a history404 as missing capabilities.
There is no new watcher, state machine or public API. The existing shared
session module owns this protocol detail for both platforms.

Four behavioral regressions cover initialize/synchronize overlap, successful
upgrade recovery, obsolete404 after a newer success, and unrelated history404.
Before repair: **3 FAIL/1 PASS**. After repair with existing store/lifecycle/read
ordering: **4 files/65 PASS**. Parent separately reran all four successfully.
The new frontend was rebuilt; main production E2E, references and applicable
visual/host checks were then rerun. Gate07 and Gate14 contain follow-up notes;
this report does not silently replace their previous scoped assessments.

## Harness improvements

- Native identity first-frame checks now reject old owner title and saved
  memory text as well as answer, unsent inputs and dialogs. These assertions
  run inside the full final main suite, not a synthetic StorageEvent test.
- A genuine persisted `pageshow` schedules capture in the **first receiving
  RAF**. If restoration occurs, the test rejects old answer/title/draft there.
  Both final Chromium probes returned persisted=false; navigation type was
  `back_forward`, and `notRestoredReasons` exposed only `masked`. Both remain
  **NOT RUN**. Ordinary history restore and trusted native cross-tab storage
  transitions are separately passed; no synthetic bfcache event substitutes.
- Reference actionability uses Playwright `click({trial:true})` before capturing
  the anchor, allowing real route rerender/attachment to settle without a click.
  The first older-image reference run failed in `scrollIntoViewIfNeeded` when
  Framework7 replaced the DOM after Back. Its raw failure is retained; the
  business assertions were not weakened. Complete reference reruns followed.
- Compatibility explicitly rejects a current Docker override, probes actual
  old capabilities404, and keeps browser launch inside cleanup. Reference,
  store and visual runners likewise clean partial browser startup. Result
  provenance separates production Docker from Vite source-module tests.
- Host and Docker environments both remove their own mkdtemp directory on
  close. External `EBK_E2E_BINARY` paths are never removed by that helper.
  A new normal-close regression checks directory removal and external binary
  retention; every partial-startup cleanup test also checks directory absence.
  Restoring the former host cleanup behavior failed the new regression.
  Final helper suite: **3 files/11 PASS**.
- Real nginx uses task-only create/copy/start with no host bind mount and bounded
  CLI calls. App cleanup still runs if proxy cleanup fails. No shared Docker
  resources are pruned or reconfigured.

## Results

Scenario groups below are not a claim that every Cartesian combination or all
R01–R40 requirements are proved by one browser suite.

| Execution | Result | Environment / evidence |
| --- | --- | --- |
| Final full application suite | **47 PASS, 2 NOT RUN**, exit0;26 provider calls | New image; `main/results.json`; includes production login smoke,14 native identity cases, both full platform flows |
| Bill reference checks | **10 PASS**, exit0;8 provider calls | New image; `reference/reference-results.json` |
| Old-backend compatibility | **2 PASS**, exit0;0 provider calls | Final production dist + actual isolated old-main host binary; each platform2 capabilities404 requests,5-second stable navigation observation |
| Platform command HTTP checks | **11 PASS groups**, exit0 | Original image, same exact backend executable; `platform/platform-commands.json` |
| Real nginx proxy | **7 checks PASS**, exit0 | nginx1.28.3,4 real configurations, one durable command, exactly2 provider calls; `proxy/results.json` |
| Upstream cancellation / smoke | **10 PASS groups**, exit0 | Original image, same exact backend executable; `upstream-cancel/smoke-results.json` |
| Real store / long history | **7 behavior groups + history PASS**, exit0 | Current host Go/Vite/isolated SQLite; intentionally not Docker production evidence |
| Visual long-content / accessibility | **2 design /4 accessibility variants PASS**, exit0 | New image, actual desktop/mobile routes; no physical screen-reader or keyboard claims |

The main suite covers trusted cross-tab owner changes with confirmation,
memory and history editors open, export held across owner/navigation changes,
obsolete clear confirmations after a real second-tab epoch advance, default
IndexedDB migration/conflicting drafts, storage refusal, send/detail/back and
refresh, real clipboard, idle second-device rename, offline non-resend,
generation-disabled restart/history/export, history search/archive/unarchive,
lost accepted send and branch responses, retry versions and explicit branch
send, and clear retaining the real ledger. It rejects all pageerrors.

References verify actual modified/deleted/foreign bills and separate transport
failure, unchanged original answer and queriedAt, persisted new reference amount
23456 after explicit regeneration, and retained draft/reading context. Desktop
also keeps the notice open through16 seconds of polling.

The provider cancellation evidence now observes `response.close` with
`writableEnded=false` **554ms after explicit cancel**, before the5-second
fixture response delay. The request entered first and the durable run reached
canceled. This is an unfinished upstream HTTP response closure, not request-body
`end`, normal response completion or closure during final environment cleanup.

Nginx buffering off/on honoring X-Accel-Buffering produced initial bytes in
6/16ms. Forced buffering ignoring that header was also executed (7ms observed;
no invented buffering delay). A1-second proxy idle timeout actually interrupted
the subscription at2019ms without a terminal frame. Polling and reconnect read
the same completed durable run; no send/retry was made and provider calls stayed2.

## Requirement coverage boundaries

| Requirement IDs | Evidence supplied here / remaining layer |
| --- | --- |
| R01–R04 | Actual two-platform navigation/history/management, native drafts and storage refusal; Unicode/IME bounds belong to unit/component matrix |
| R05–R07 | Production layouts, actual clipboard and host120-turn anchor/window test; physical phone keyboard remains NOT RUN |
| R08–R12 | Real failed/retry/cancel, response-loss reconciliation, branch and offline recovery; exhaustive worker/CAS/concurrency contracts belong to backend matrix |
| R13–R15 |10 reference cases plus real paging/restart/version commands; independent extreme-amount oracles belong to unit/database matrix |
| R16–R18 | Deterministic protocol, memory creation/forget and history only; real financial/long-context/memory model quality remains NOT RUN |
| R19–R21 | UI epoch/management and protocol paging; revocation/quotas/context budget correctness additionally requires backend matrix |
| R22–R24 | Native owner first-frame exclusion, foreign bill rejection, real cancellation/proxy reconnect; full UID/lease/malformed-frame matrices remain backend/unit scope |
| R25–R28 | Fresh-schema smoke, real scoped layout/keyboard/contrast checks and bounded history; migration/backup/performance and security units belong to next matrix; manual reader/model cost NOT RUN |
| R29–R31 | Disabled-generation history/export, repaired real old-backend compatibility and shared store recovery; worker save-failure authority remains backend scope |
| R32 | Local frontend/container checks here; final full unit/three-database/parent gates still required, real-model release checks NOT RUN |
| R33–R36 | Native clear epoch, stale confirmation/export fences, unknown branch/send persistence, remote delete; concurrent service revocation/receipt barriers remain backend scope |
| R37 | Production SW controls page; no assistant CacheStorage response, HTTP no-store, no answer/token in IndexedDB, no cached offline API; actual bfcache restoration NOT RUN |
| R38–R40 | Real proxy timeout/recovery, two-stage branch and reviewed current screenshots; slow-socket/subscription limits are backend checks, physical input/reader remain NOT RUN |

## Reproduction and isolation

Build with `EBK_E2E_ARTIFACTS=.vitest/final-e2e-rebuild
EBK_E2E_DOCKER_IMAGE=ebk-assistant-validation:20261016-compat
node scripts/build-assistant-docker-validation.mjs` (one shell command).
For final container frontend checks, set `EBK_E2E_DOCKER_IMAGE` to the new
immutable ID above, `EBK_E2E_PRODUCTION=1`, a distinct `EBK_E2E_ARTIFACTS`, then
run serially:

```sh
node scripts/assistant-experience-e2e.mjs
node scripts/assistant-reference-e2e.mjs
node scripts/assistant-experience-design-review.mjs
node scripts/assistant-experience-accessibility-review.mjs
```

The main suite already includes strengthened identity tests; a redundant
standalone repeat is unnecessary. Backend-only scripts are
`assistant-platform-command-smoke.mjs`, `assistant-proxy-smoke.mjs` and
`assistant-experience-docker-smoke.mjs` with their recorded image IDs.
The store/history script directly imports Vite modules and seeds its isolated
SQLite history, so run `node scripts/assistant-store-api-smoke.mjs` with neither
Docker nor prebuilt binary override. Compatibility uses an explicit old-main
binary built from `git archive` in a task directory; its source/binary were
removed after completion. No shared checkout was changed.

Only reviewed sanitized JSON/screenshots are persisted under
`docs/ai-assistant-experience-evidence/audit-e2e/`; temporary ports, paths,
container names and synthetic IDs are normalized. Summary records raw log
hashes rather than copying application logs. No private config/provider keys
were read. Final cleanup and visual/host outcomes are appended below.

## Limits and handoff

No claim of zero possible bugs, full real-model certification, stock Docker
release build, amd64, WebKit/Firefox, physical iOS/Android keyboard or manual
screen-reader execution. True bfcache restoration did not occur. The next
serial unit/backend/database stage and final parent review must run before
commit/push. The scoped compatibility design increment preserves one protocol
authority and clean interfaces; assessment10/10 for that narrow change, not a
replacement score for unexecuted release requirements.

## Final completion snapshot

All executions listed above finished before handoff. The host store rerun is
**7 behavior groups PASS**, plus a 120-turn/555,816-byte synthetic history:
three prepends retain the same message anchor within 0.407px, DOM/store window
stays bounded at 90 turns, Back to latest returns to turn 120, viewport resize
moves the selected anchor 0px, and another browser context's remote deletion
clears current DOM and draft. One ambiguous send is confirmed without resend;
a branch whose first send fails retains the same target. This is current host
Go/Vite evidence, deliberately separate from production Docker.

The new image's long-content review passed both 720px desktop and 390px mobile
variants: 44px named toolbar controls, title widths 510/230px, no document
horizontal overflow, and table tabular numerals/contained overflow. The real
accessibility review passed desktop/mobile × light/dark (4 variants): enabled
Send contrast 5.34:1 at rest/focus and 6.60:1 at hover/press, visible focus,
accessible full title in More, true modal background inertness, 18 keyboard
Tab steps contained, nested Escape focus return, responsive modal→sidebar→modal,
and new-conversation navigation. All variants report zero pageerrors.
These are targeted actual-browser assertions, not a full automated WCAG scan.

Reviewer viewed full-size long-content desktop/mobile, mobile dark history,
desktop dark send focus, final mobile answer and repaired old-backend notice.
Parent independently viewed the new-image mobile long-content and desktop dark
history screenshots: reading content stays in its panel, composer remains
visible, and modal/focus treatment has no newly observed layout defect.
Selected reviewed screenshots are persisted; other raw state captures remain
in the run directories. Earlier Gate10/11 RTL/200%/landscape evidence remains
separate; these exact configurations were not all rerun in this stage.

Final helper rerun: 11 PASS. Scoped product ESLint and 12 changed-script syntax
checks passed. Source-line gate: 1,710 files, zero over 500. `git diff --check`
passed. All task-labelled containers are absent; only unrelated
`codex-s03-build` remains running and was untouched. Runtime lifecycle JSON
ends with removal for every container run. No network was created. Own fixture
config/SQLite/binary directories are removed; old-main source/binary archive
also removed. Validation images remain for the next stage. Cleanup does not
remove raw `.vitest` logs or source-controlled evidence.

[summary.json](../ai-assistant-experience-evidence/audit-e2e/summary.json)
records exact log hashes, counts, final product file hashes and cleanup.
The main suite's 2 NOT RUN bfcache cases are not passes. Stock Docker release,
physical input/manual reader and commercial provider limitations remain open.
This stage is frozen for the next serial complete unit/backend/database matrix.
