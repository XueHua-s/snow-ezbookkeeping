# AI assistant: serial skill audit and container verification

Started 2026-10-16. Scope: the assistant experience PR, initially commit
`910819988e77e5dbf393ab69c95aa047c8157922`, relative to `origin/main`.
Follow-up fixes are included in each subsequent review. Existing unrelated
application features and user Docker containers are not cleanup targets.

## Review contract

Only one child agent runs at a time. Each skill is reviewed in sequence; a
reviewer reads the actual skill and its required references, identifies
applicable requirements, fixes substantiated defects, and records evidence.
The parent independently checks the findings and changes before advancing.
Later changes that invalidate an earlier review require re-review.

The software-design-philosophy skill defines a 0–10 score. For other skills,
10/10 is this audit's assessment that all applicable reviewed requirements
are satisfied, not a scoring system invented by the skill itself. A score
does not replace execution results or certify absence of all possible bugs.
Required checks that cannot run remain explicitly unverified.

Focused regression checks may support a repair during review. The full local
Docker smoke, application E2E, unit-test matrix and final parent validation
run serially after the skill review stages. No commit or push occurs until
the requested local verification and parent review finish.

## Current local execution evidence

All14 applicable skill gates are closed at **10/10 within their documented local scope** after serial reviewer closure and parent review. No known implementation defect remains open in this audit. External release acceptance remains incomplete; the PR stays draft.

The tested product frontend and Linux executable are byte-identical between
Gate16's full application E2E image and Gate18's rebuilt validation image.
Gate18 changes only the test/build harness. Earlier counts and pending notes
below remain historical; this snapshot and each review's latest closure govern
the current state.

| Verification | Final observed result | Evidence |
| --- | --- | --- |
| Local Docker build and smoke | Build exit 0; 10 smoke scenarios PASS; explicit timeout/container cleanup PASS | [18](ai-assistant-experience-audit/18-harness-finalization.md) |
| Desktop/mobile container E2E | 47 PASS / 2 actual-bfcache NOT RUN; references 10 PASS; old-backend compatibility 2 PASS | [16](ai-assistant-experience-audit/16-container-e2e.md) |
| Final Vitest | 44 files / 38,917 PASS, parent independent run | [18](ai-assistant-experience-audit/18-harness-finalization.md) |
| Chromium component suites | 7 files / 62 PASS | [17](ai-assistant-experience-audit/17-final-test-matrix.md) |
| Whole-repository Go / required race scope | 1,746 / 754 leaf PASS; 4 opt-in SKIP per broad run, performance subsequently executed | [17](ai-assistant-experience-audit/17-final-test-matrix.md) |
| Final services after test parameterization | 441 leaf PASS / 2 performance SKIP; performance separately PASS | [17](ai-assistant-experience-audit/17-final-test-matrix.md) |
| Exact three-engine regressions | Ledger 42 and memory 9 PASS in both normal/race modes | [17](ai-assistant-experience-audit/17-final-test-matrix.md) |
| Performance and native migration | Six three-engine performance leaves and one HTTP export PASS; upgrade/backup/restore/rollback/re-upgrade PASS on all three engines | [17](ai-assistant-experience-audit/17-final-test-matrix.md) |
| Static/build checks | Lint/types, build, go vet, gofmt, source-size and diff checks PASS | [17](ai-assistant-experience-audit/17-final-test-matrix.md), [18](ai-assistant-experience-audit/18-harness-finalization.md) |

Repeated or overlapping modes are not added as unique coverage. Commercial
model deployment remains the unexecuted opt-in Go integration. The Docker image
uses the explicitly documented cached-Go/Alpine fallback and host Node22 build;
stock Dockerfile/Node26, amd64, WebKit/Firefox, physical phone keyboard, manual
screen reader and actual bfcache restoration are unverified. Local review
scores do not remove these release-acceptance gaps.

## Ordered skill reviews

| Order | Skill | Scope | Status |
| --- | --- | --- | --- |
| 01 | software-design-philosophy | Module depth, information hiding, complexity, interfaces and comments | [10/10 final review; product and D8 build/smoke boundaries closed](ai-assistant-experience-audit/01-software-design-philosophy.md) |
| 02 | ezbookkeeping-project-spec | Data safety, migration, shared authority and required gates | [10/10 final local review; D4/D5, quality and migration evidence closed](ai-assistant-experience-audit/02-project-spec.md) |
| 03 | ezbookkeeping-project-structure | Layer ownership, entry points and documentation consistency | [10/10 final structure review; maps and new harness boundaries synchronized](ai-assistant-experience-audit/03-project-structure.md) |
| 04 | ezbookkeeping-ai-integration | UID-bound read-only tools, exact amounts, providers and lifecycle | [10/10 final local review; database, performance and runtime evidence closed](ai-assistant-experience-audit/04-ai-integration.md) |
| 05 | vue-best-practices | Component boundaries, reactivity, contracts and rendering cost | [10/10 final local review; component, lifecycle and downstream evidence closed](ai-assistant-experience-audit/05-vue-best-practices.md) |
| 06 | vue | Composition API, lifecycle and template correctness | [10/10 final local review; route, visual and matrix conditions closed](ai-assistant-experience-audit/06-vue.md) |
| 07 | pinia | Store ownership, actions, identity and disposal | [10/10 final local review; bootstrap leases and capability-404 increment verified](ai-assistant-experience-audit/07-pinia.md) |
| 08 | vue-router-best-practices | Route reuse, async navigation, guards and cleanup | [10/10 code/router review; final E2E evidence verified by parent](ai-assistant-experience-audit/08-vue-router.md) |
| 09 | vueuse-functions | Appropriate existing composables and resource cleanup | [10/10 code/composable review; final runtime evidence verified by parent](ai-assistant-experience-audit/09-vueuse.md) |
| 10 | frontend-design | Coherent finance UI, typography, density, themes and responsive layout | [10/10 final local review; current Docker visuals and contrast follow-up verified](ai-assistant-experience-audit/10-frontend-design.md) |
| 11 | web-design-guidelines | Current guidelines, keyboard access, semantics and interaction | [10/10 final local review; Docker contrast/modal/focus verified; manual input remains unverified](ai-assistant-experience-audit/11-web-design-guidelines.md) |
| 12 | vite | Production build boundaries, assets and service-worker integration | [10/10 final local Vite review; build/artifact/runtime verified; stock release build unverified](ai-assistant-experience-audit/12-vite.md) |
| 13 | vitest | Test isolation, behavior assertions, async races and coverage gaps | [10/10 final local test review; native tabs, full matrix and harness failure paths verified](ai-assistant-experience-audit/13-vitest.md) |
| 14 | vue-testing-best-practices | Real component/route behavior, browser assertions and E2E scope | [10/10 final local test review; native identity, container E2E and full matrix verified](ai-assistant-experience-audit/14-vue-testing.md) |

## Applicability exclusions

The repo uses Vue/Vite and npm; this work does not introduce Nuxt, pnpm,
UnoCSS, Turborepo, tsdown, Slidev or VitePress. Those skills do not impose
requirements on this change. The ezbookkeeping API-tools skill serves
personal ledger operations; this task tests isolated application APIs
directly. antfu does not replace the repository's existing toolchain or
conventions during this review. Artifact-generation and unrelated product
skills are outside the implementation scope.

## Verification stages after reviews

1. A dedicated child agent builds and runs the current application in local
   Docker, with task-owned resources and isolated data; verifies startup,
   migrations, health, authentication, representative assistant flows and
   persistence/restart behavior.
2. A dedicated child agent closes E2E gaps and runs desktop/mobile flows
   against the containerized application, including recovery, identity,
   destructive-command fencing, history, bill references and cache behavior.
3. A dedicated child agent closes meaningful unit-test gaps and runs Vitest,
   browser component tests, Go tests/vet/race and required database checks.
4. The parent reviews code, results and coverage against the plan, resolves
   remaining defects, revalidates affected earlier gates, and only then
   commits and pushes to the existing PR.

Real commercial-model quality, physical mobile keyboards, manual screen
readers and actual bfcache restoration require their own positive evidence.
Docker or deterministic fixtures alone do not prove those requirements.

Final E2E follow-up: the bfcache probe must record actual `pageshow.persisted`
and, if restoration occurs, inspect the first receiving frame for old-owner
content. Its existing two-RAF post-navigation assertion alone does not prove
the no-flash requirement. Preserve NOT RUN when Chromium declines caching;
capture the browser's reason if available. Also handle production-smoke child
process spawn errors so a failed launch cannot leave the runner pending.

## Parent review leads for subsequent gates

- Gates 05/08: reproduce an account switch while the clear-all confirmation
  is open. `useAssistantCloudWorkspace.clear()` awaits confirmation and then
  executes against the current store; check that an old confirmation cannot
  authorize deletion for a newly active owner. Audit confirmation/export
  lifecycle and async action feedback across owner/route changes.
- Gate 07: reproduce delayed `synchronize()` HTTP-404 local cleanup during a
  same-owner navigation/draft edit. That branch only fences epoch before
  `applyLocal`, whereas the inaccessible-state branch now also fences
  navigation and storage owner. Audit stale local snapshots and stale errors
  across all asynchronous store entry points; these are leads, not claimed
  failures without reproduction.
- Gates 06/01 re-review: assess duplicated runtime/type reason-code lists in
  the snapshot parser and DTO alongside the now-centralized run states.
- Gate 12: assess the native-config-loader ESM warning emitted by Vitest;
  preserve the existing application build contract when addressing it.

- Gates 05/08: `AssistantCloudWorkspace` mount awaits initialize/open and
  then unconditionally attaches the store. Unmount during those awaits can
  let an obsolete instance attach subscriptions/listeners again. Reproduce
  with controlled promises and verify cleanup after late completion.
- Gate 07: include `send` finally (unconditional submitting=false), reset
  clearing of transient errors/loading/backoff, and asynchronous command
  acceptance tails in the identity/navigation review. Require regressions
  showing obsolete operations cannot change a current operation's state.
- Gate 10: parent inspected baseline mobile-answer and desktop-zoom200
  screenshots. At 200% zoom the desktop toolbar squeezes the conversation
  title to about two characters while long action buttons retain full text.
  Review hierarchy and adaptive toolbar density beyond horizontal-overflow
  assertions, with refreshed real-route visual evidence after refactoring.
- Historical Gate04 lead (repaired below): ledger-clear barrier regressions
  initially covered SQLite only. Extend those exact behaviors to MySQL and
  PostgreSQL. Measure account clearing with many conversations: correctness
  initially preserved per-ledger-item commits and scanned/invalidated all
  conversations for each successful deletion (O(N×C)). Re-review any
  optimization against rollback, partial failure and same-turn regeneration.

  Parent focused SQLite measurement now quantifies the amplification: deleting
  25 ledger rows took 13.33 ms with no conversations and 2,962.78 ms with
  1,000 synthetic conversations, even with no concurrent assistant mutation.
  The probe preserves real header reservations and checks all ledger rows
  are deleted. Source is `/tmp/ebk-parent-ledger-clear-probe_test.go`, log
  `/tmp/ebk-parent-ledger-clear-probe.log`; the temporary test was moved out
  of the repository. An initial under-reserved probe fixture failed with
  quota_exceeded and was corrected before measuring; that was not a product
  failure. Re-review this cost before Docker so a necessary optimization
  does not invalidate all later container evidence. Per-clear reuse is safe
  only if authoritative locked user-state version proves no intervening
  assistant transaction, and only successful commits may advance that proof.

## Parent verification during review

- Final whole-change formatting check found one missing blank line before the
  new `Prepare` API comment in `pkg/services/ai_assistant_context.go` (the stage17
  gofmt check covered its changed test file). Parent applied `gofmt`, adding
  only that blank line, then confirmed all16 changed/new Go files are clean.
  This nonsemantic formatting correction occurred after the tested images;
  their recorded hashes remain actual historical build identities, not a claim
  that a new image was rebuilt after this line-number change. No behavior or
  frontend code changed, so unchanged application tests were not repeated.
  All five modified skill-file mirror pairs also compare byte-for-byte equal.
  Staging exposed one trailing space in the new identity E2E script, which
  parent removed; the complete staged diff check then passed.

- Gates08/09 retain their completed 10/10 code/targeted reviews: no later
  route/scroll/composable product edits invalidate them. Parent read Gate16's
  native desktop/mobile navigation and identity results plus host long-history
  measurements (120 turns, at most90 mounted items, three prepend drifts at most
  0.407px and zero resize drift). Gate17's 62 Chromium tests and Gate18's final
  full unit run include their unchanged suites. This closes their previously
  pending final-matrix handoff; actual bfcache/physical-input limits remain.

- Final handoff preflight: parent independently verified all 43 referenced raw-log
  SHA-256 entries in the Docker/E2E/matrix/harness evidence; all matched and
  none were missing. All newly retained JSON files parse. The new-file inventory
  contains no database, backup, private key, INI, environment or raw log files;
  a credential-pattern scan found no provider/GitHub keys, JWTs, private keys
  or credential-bearing URLs. This is a bounded artifact review, not a universal
  secret-detection guarantee. Only the two pre-existing unrelated Docker
  containers remain, and `git diff --check` passes.

- D8 build/smoke follow-up: a shared bounded command runner now owns POSIX
  process-group deadlines, signal handling and capped capture; build resources
  own unique attempted container names and cleanup. Parent read both boundaries
  and required the capture cap to remain effective under continuous output.
  Real subprocess/descendant/signal and controlled CLI regressions pass. The
  actual build and smoke reran; frontend and Linux executable hashes remain
  identical to stage 16. Parent independently reran the complete final Vitest:
  **44 files / 38,917 PASS** (`/tmp/ebk-parent-final-unit.log`), plus source size
  **1716 files / 0 over 500** and diff checks. This supersedes stage 17's earlier
  frontend unit total, rather than adding duplicate tests to it. Gate01 independently re-reviewed the implementation and evidence: D8 is
  closed and the combined design score is 10/10.

- Final matrix stage 17: [complete local quality evidence](ai-assistant-experience-audit/17-final-test-matrix.md)
  records Vitest 38,907 PASS, Chromium 62 PASS, whole-repository Go 1,746
  leaf PASS, final services 441 PASS and required race scope 754 PASS.
  These overlapping modes are not added as unique coverage. Opt-in performance
  skips were subsequently exercised: three-engine ledger/history performance
  6 PASS and actual HTTP large export 1 PASS. Commercial deployment remains
  NOT RUN. Exact ledger-clear cases are 14 × 3 engines and memory projection/
  revocation cases 3 × 3, all PASS in normal and race modes. Native upgrade,
  backup/restore, old-binary rollback/re-upgrade and lease recovery passed all
  three databases. Parent read actual command exit records and raw test/perf
  results, independently ran both Python migration-runtime regressions
  (`/tmp/ebk-parent-migration-runtime.log`), and checked final container inventory:
  only the two pre-existing unrelated containers remain. Lint/type/build/vet,
  gofmt, source size and diff checks pass. Final dist hash equals the tested
  stage 16 image. Only test parameterization and migration-harness repairs
  changed in stage 17; no product source drift. Original skill reviewers now
  close their conditional checks serially against these final results.

- Container E2E stage 16: [final application evidence](ai-assistant-experience-audit/16-container-e2e.md)
  records 47 PASS and 2 actual-bfcache NOT RUN, references 10 PASS and
  old-backend compatibility 2 PASS. Real compatibility testing found a
  capability-404 race between initialization and foreground synchronization.
  Classification now lives behind the current ticket/epoch fence in session;
  unrelated history 404s no longer imply an old backend. Parent independently
  reviewed and ran its four regressions: 4 PASS
  (`/tmp/ebk-parent-compatibility.log`). The final frontend image is
  `sha256:bd5b1c31834ad9d917679b9bb9007edae4b9553201a2af48939f6796d1d339e1`.
  Parent verified unchanged Go source and identical Linux binary hashes versus
  stage 15; frontend artifacts changed and all affected application checks
  reran. Parent read final runtime/results/cleanup records and inspected
  desktop/mobile answer, mobile long-content and dark history screenshots.
  New first-frame assertions also exclude old title and saved-memory text.
  Actual upstream cancellation closed an unfinished provider response in
  554 ms, adding evidence beyond stage 15's persisted canceled-state assertion.
  Full unit/backend/three-database verification follows serially in stage 17.

- Docker stage 15: [actual container smoke](ai-assistant-experience-audit/15-docker-smoke.md)
  reports 10 PASS. Parent independently read the runtime/build/smoke adapters,
  required explicit query evidence and disabled-provider assertions, and
  identified uncertain Docker-create cleanup being silently treated as absence.
  That cleanup defect is repaired and tested. Parent independently ran all
  three harness test files: 10 PASS (`/tmp/ebk-parent-docker-helpers.log`).
  Recomputed 789 source-file and 107 dist-file hashes exactly match the build
  provenance. Task containers were removed; unrelated containers retained.
  The image uses cached Go/Alpine 3.24.2 and host-built frontend, not the stock
  Dockerfile runtime/frontend stages. Full container E2E now follows serially.

- Gate07 final auth re-review found and reproduced the last store-to-caller
  microtask window. The token store now returns a `RefreshedSession.isCurrent`
  lease; DesktopApp, MobileApp and Unlock recheck before applying locale,
  notification or cache settings. Parent read the implementation and actual
  Unlock/composed-store tests, then independently ran all auth regressions:
  3 files / 24 PASS (`/tmp/ebk-parent-gate07-auth-final.log`). The positive
  current-session case also passes, guarding against premature negative
  assertions. Gate01 receives this small interface increment for re-review.

- Gate04 performance re-review: operation-local proof now reuses invalidation
  only when the locked database version proves no intervening assistant write.
  Storage owns version arithmetic and returns proof only after Commit succeeds.
  Parent independently ran all ledger-clear/reuse/scaling regressions:
  14 SQLite leaf tests PASS, 28 external-database leaf tests SKIP, explicitly
  not passes (`/tmp/ebk-parent-gate04-reuse.log`). With 25 ledger rows,
  0 conversations took 12.29 ms (1 scan); 1,000 took 128.67 ms (21 scans,
  exactly one paginated traversal). Tests now parameterize all three databases;
  real MySQL/PostgreSQL execution remains the final database stage. The tests
  cover interleaved new runs, completed same-turn regeneration, mutation
  rollback, final CAS retry and partial account-clear failure. The new proof
  boundary must also receive Gate01's design re-review before Docker.

- Gate04: independently read the transaction helper/call sites and new ledger
  regressions; `BUILD_PIPELINE=1 go test ./pkg/services -run
  '^TestAssistantLedger' -count=1` PASS; `git diff --check` PASS. This is
  SQLite evidence, not a substitute for the final database matrix.

- Gate06: memory source buttons currently key only by `source.messageId`;
  Gate02 projection intentionally permits conversation-only sources (empty
  message ID). Check stable composite keys and navigation for multiple such
  sources. History panel local `query`/`includeVersions` resets on v-if remount
  while parent query remains: reproduce closing/reopening after search and
  make rendered query, visible results and next-page source agree.

- Final E2E stage: existing `identityBoundary` uses synthetic StorageEvent
  on its non-bfcache path; bfcache probes navigate away before switching.
  Add live two-tab identity changes while confirmation/export/edit UI is open
  using actual storage events, and verify both server data and no download/
  stale UI effect. Name evidence accurately; a dispatched event is not by
  itself proof of a native cross-tab or actual bfcache transition.

- Gate13/14 local-storage review lead: `deviceIdentity()` does separate
  localStorage get/generate/set before owner rows filter by deviceId. Assess
  simultaneous first use in two tabs, or removal of that ID while IndexedDB
  survives: no saved drafts/pending may silently become unreachable on reload.
  Existing browser fixtures inject a shared deviceId and therefore do not
  cover the production identity initializer. This is not yet a reproduced bug.

  Parent follow-up: the missing-ID path is now reproduced in an isolated
  Chromium profile using the actual production storage factory (no deviceId
  override). Save a synthetic draft, close storage, remove only
  `assistant-device-id`, and reopen the same owner: beforeCount=1,
  afterCount=0, both snapshots still report scope=device. IndexedDB was not
  cleared. `/tmp/ebk-parent-local-identity-probe.log` records the failing
  assertion (exit 1); browser/server were closed. Gate13 must repair and
  add persistent regression coverage. Simultaneous first use remains a lead.

- Gate07: same-route competing reads also need review: two answer-version
  selections, memory refreshes and history windows can finish out of order
  without changing identity/navigation. Check user intent ordering and
  monotonic snapshots, not only cross-owner fences; prove any repair using
  reversed-response tests instead of time-based sleeps.

- Async review rule (Gates05/07/08): a nested async helper returning a
  scope-checked boolean/result does not protect its caller's next await
  continuation. Test a scope change queued between helper resolution and
  caller continuation; guard at the actual command/state/DOM side effect.
  Parent identified this interleaving in confirmCurrent → clear; Gate05 is
  adding an explicit microtask-order regression.

- Gate05 parent check: independent command/lifecycle/export/confirmation
  Vitest run passed 43 tests across four files. `git diff --check` and source
  line gate passed (1679 checked, 0 over 500). Parent review additionally
  required balanced attachment ownership, old memory-editor teardown,
  helper-to-caller microtask fencing, stable identity on capability refresh,
  and generation-fenced delayed focus restoration; added regressions cover
  these boundaries. Full route/visual evidence remains a later stage.

- Gates08/09: check reading state on reused MessageList when moving from a
  scrolled-up conversation into a different conversation without saved anchor.
  `restore(undefined)` currently does nothing, so prior following=false and
  message anchor may survive. Reproduce with actual layout and verify fresh
  target follows latest while returning to old history restores its position.
  AutoScroll also exposes unused `preserveWhile`; assess whether that surface
  is justified and whether queued RAF/layout work respects teardown.

- Gate09 additional scrolling lead: `layoutChanged()` skips the direct
  follow branch for a nonempty browser selection but then calls
  `restore(anchor)`. If that captured anchor has `following: true`, restore
  may still scroll to latest. Reproduce selecting text near the bottom and
  appending content; the plan explicitly preserves reading/selection. Also
  distinguish a selection inside this message panel from unrelated page text.

- Gate11: Gate10 screenshot inspection measured enabled Send white text on
  warm brown near RGB(177,115,63), approximately 3.89:1. Confirm the actual
  computed foreground/background and repair normal-text contrast to 4.5:1
  where applicable, including both theme variants and relevant button states.
  Disabled controls are a different criterion; do not confuse their opacity
  with enabled-control evidence.

- Gate06 parent check: independent lifecycle/VueContracts/snapshot-parser run
  passed 32 tests in three files; diff and source line checks PASS (1681/0).
  Parent-provided closed-panel and still-visible-panel route/filter races
  both reproduced red and were fixed. Same-owner navigation now closes the
  history view while retaining filters; reopening queries current filters.

- Gate07 parent check: independently reviewed the session/conversation/command
  module boundaries and required regressions for actual capabilities A→B,
  epoch hydration/polling, per-run backoff, and independent message versions.
  Store/lifecycle/readOrdering/dataClear plus environment-isolation tests:
  5 files / 64 tests PASS (`/tmp/ebk-parent-gate07-final.log`); diff-check PASS.
  D5 is repaired pending Gate01's final design re-review. This is focused
  evidence, not the full Docker or application-route matrix.

- Gate08 parent check: read lifecycle coalescing/404 reconciliation, fresh
  reading anchors, watcher invalidation and RAF cleanup. Independent focused
  workspace lifecycle/AutoScroll/readOrdering run: 3 files / 29 tests PASS
  (`/tmp/ebk-parent-gate08-focused.log`). Reviewed final browser log: 2 files /
  11 tests PASS. Initial offline recovery retains attachment; store API was
  not expanded. Real application E2E remains a later stage.

- Gate09 parent check: read unified panel/content layout handling and actual
  range intersection; independently ran AutoScroll and MessageList browser
  suites: 2 files / 7 tests PASS (`/tmp/ebk-parent-gate09-browser.log`).
  Scope is Chromium component geometry/selection, not physical input or a
  full application E2E. Gate09 closed the recorded selection/resize lead.

- Gate10 parent check: independently viewed fresh actual-route screenshots
  desktop-zoom200, mobile-answer, mobile-landscape, desktop-dark and
  mobile-memory in `docs/ai-assistant-ui-review/gate10/`. Title space and
  message/composer geometry are restored. Parent identified that bare scoped
  `.mobile` also affects Composer roots; root selectors now explicitly target
  `.cloud-workspace.mobile/.desktop`. Final focused browser log reports
  17 PASS after accessible-name/focus selectors were updated for icon More.
  Send contrast and touch access to the full truncated title remain Gate11.

## Local verification environment (read-only inventory)

Gate14 parent check: independently reviewed the Framework7 URL replacement,
refresh coalescing/queue settlement, stale-token responses, and initiating
owner/token constraints at Axios dispatch. Parent identified the additional
service-to-interceptor microtask window; the child reproduced it before
repairing refresh/revoke and assistant read/write/export boundaries.
Independent auth tests: 2 files / 21 PASS
(`/tmp/ebk-parent-gate14-auth-final.log`); real Framework7/workspace Chromium:
2 files / 16 PASS (`/tmp/ebk-parent-gate14-browser.log`). Read both final
native app runs: desktop/mobile each 7 scenarios, 14 PASS per run, zero
pageerrors. These are host/Vite results. Gate04 now reopens serially to
repair the measured account-clear performance regression before Docker.

Gate13 parent check: independently reviewed default browser storage ownership,
atomic legacy-ID normalization, migration notifications, late connection
closure and bounded harness startup/cleanup. Independent reruns: environment
2 files / 9 PASS (`/tmp/ebk-parent-gate13-environment.log`), Chromium storage
2 files / 26 PASS (`/tmp/ebk-parent-gate13-storage.log`); diff-check PASS.
This closes the parent's missing-localStorage-ID reproduction. Controlled
constructor interleaving is not native two-tab proof; Gate14 owns that gap.

Gate12 parent check: independently reviewed the three `.mts` configurations,
JSON import attributes, config-relative paths and explicit Git directory.
Loaded all three with Vite's native loader and asserted the actual `@` alias:
3 PASS, `/tmp/ebk-parent-gate12-native.log`. Read the child's final report of
host production build, artifact checks, 2 cache-policy and 5 native-browser
tests. Container production validation remains outstanding.

Gate11 parent check: independently read native history-dialog/focus handling
and scoped CSS ownership, and viewed mobile-dark-history, desktop-light-title
and mobile-light-send-focus screenshots. No additional visual defect found.
Final focused logs show 68 tests and browser logs 18 tests PASS; four real
application variants report zero page errors and enabled Send contrast
5.34:1 (rest/focus), 6.60:1 (hover/pressed). Parent diff-check PASS. These
are host Chromium results, not Docker, manual screen-reader or physical
keyboard evidence.

Parent found and repaired an E2E isolation gap: backend startup filtered
`EBK_*` but inherited `EBKCFP_*` file-based configuration overrides. The
existing environment helper now strips both before injecting fixture-only
settings. A synthetic-env Vitest regression failed against the old filter
(`/tmp/ebk-env-isolation-red.log`) and passes with the fix (1 test). No host
secret values or files were read. Gate13 should include this regression in
the final test inventory.

At Gate07 handoff, branch/HEAD still `codex/ai-assistant-experience` /
`910819988e77e5dbf393ab69c95aa047c8157922`; no audit commit or push yet.
Host tools: Go 1.27.1 darwin/arm64, Node v22.20.0, npm 10.9.3.
Docker is available; only running container is unrelated `codex-s03-build`
and must remain untouched. Cached task-use bases include Go 1.27.1 Alpine,
MySQL 8.0, Postgres 16 Alpine, nginx 1.28 Alpine. No audit containers were
created by this inventory. The final build report must identify actual
frontend/runtime toolchains rather than infer them from Dockerfile tags.

## Delivery handoff

Final parent review is complete. The current branch is
`codex/ai-assistant-experience`; this audit adds a follow-up commit to the existing
draft PR2, without merging or changing user runtime containers. The selected
changes include product fixes, meaningful regressions, isolated verification
helpers, synchronized skill maps and sanitized evidence. Git commit/push outcome
is reported separately after executing the actual commands; no successful push
is inferred from local checks.
