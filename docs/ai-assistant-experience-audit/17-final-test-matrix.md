# 17 · Final unit, backend, database and migration matrix

2026-10-16. Serial child stage, following frozen Gate16. No delegation,
commit/push or shared-container cleanup. All database rows/configuration are
synthetic; only task-owned MySQL8/PostgreSQL16 containers are used. This
report does not replace Gate16's application E2E or claim zero possible bugs.

## Scope and repairs

Read the plan §13, master audit and Gates02/04/13–16; applied project-spec
and global design guidance, AI-integration/feature map, Vitest and Vue testing
skills. Existing Vue production architecture is unchanged. The new regression
work targets test isolation and actual database behavior, not snapshots.

The migration CLI previously inherited both `EBK_*` and `EBKCFP_*` settings,
allowing host overrides to supersede its synthetic INI. It now strips both
prefixes and injects only the fixture `EBK_WORK_DIR` (plus BUILD_PIPELINE).
A synthetic environment passed to a real Python child reproduces the former
failure and passes after repair; no private config or secret file is read.
The subprocess boundary has 180-second command /600-second Go-build bounds,
starts its own process group and kills that group on timeout. A real parent
and spawned child regression verifies both are terminated; zombie states,
when present, mean terminated pending OS reaping, not running descendants.
Timeout reporting omits raw command arguments. The task directory is0700,
INI files0600, and exit cleanup removes fixture configs/databases/backups,
source trees and executables while retaining synthetic manifests/logs.

Gate02's exact memory list/export projection and >4-task revocation tests
were still SQLite-only. They are now parameterized across SQLite/MySQL/
PostgreSQL, preserving all original assertions:9 actual leaves instead of3.
No product source was changed in this stage. The initial full Go invocation
had already compiled the earlier test shape; a complete current services
rerun replaces that package's evidence rather than inflating unique coverage.

## Frontend and artifact identity

- `npm run lint`: PASS, including vue-tsc and ESLint autofix.
- `npm run test`:42 files /38,907 tests PASS, no skip. This includes the
  repository's extensive table-driven utilities; the count is not an assistant
  coverage percentage.
- `npm run test:browser`:7 files /62 Chromium tests PASS, no skip.
- `npm run build`:PASS. Existing inlineDynamicImports deprecation and large
  bundle warnings remain visible; warnings were not suppressed.
- 107 dist files hash to `ba13bcd1e34f10a293a2717b7a48558c524c9ad16340cc07ae9614b971e5eead`
  using the build harness's sorted path + NUL + content algorithm. This exactly
  matches Gate16's final tested image, independently showing no frontend drift.

Host tools remain Go1.27.1 darwin/arm64, Node22.20.0/npm10.9.3. Docker database
engines run Linux/arm64 through Colima. The app image remains Gate16's validation
image, not a stock Dockerfile/Node26 release build. Test-only Go source edits
change a broad source-tree hash without changing the product executable.

## Requirement mapping and limits

| Requirement IDs | Matrix contribution |
| --- | --- |
| R01–R07 | Unit/component history, draft limits, Unicode, clipboard, rendering and real browser layout; physical input stays unverified |
| R08–R15 | State transitions, durable acceptance/receipts, worker cancel/failure/recovery, version/branch and exact query/amount/continuation contracts |
| R16–R21 | Deterministic context/compaction/memory lifecycle, revocation, budgets and source projection; no claim about commercial model reasoning |
| R22–R24 | UID/isolation, malformed protocol, lease/CAS, race detector and authorization boundary regressions |
| R25–R28 | Three-engine upgrade/backup/rollback, near-quota history/export and100k ledger measurements; manual screen reader/model cost remain NOT RUN |
| R29–R36 | Feature-off behavior, compatibility units, save failure, clear epoch/receipt barriers, stale reads/commands and concurrent source fencing |
| R37–R40 | Browser IndexedDB isolation, cache-policy units, bounded subscriptions/export, service cleanup; Gate16 owns actual HTTP/proxy/navigation evidence |

Commercial primary/fallback providers, real financial/long-chat/memory quality,
physical phone keyboards, manual screen readers, actual bfcache restoration,
other browser engines, amd64 and stock Docker release build remain NOT RUN.
No skipped test is counted as passing; repeated normal/race executions are
separate verification modes, not additional unique scenario coverage.

## Backend results (actual leaf counts)

All invocations below use `BUILD_PIPELINE=1`; normal/race database suites also
set `EBK_ASSISTANT_DB_TEST=1` and task-only MySQL/Postgres DSNs whose bootstrap
names start `ebk_harness_test_`. Helpers create/drop unique fixture databases.

| Command | Result |
| --- | --- |
| `go test -json -count=1 ./...` |1746 PASS /4 opt-in SKIP;47 tested packages PASS,19 packages without tests;49.70s |
| `go vet ./...` |PASS;1.27s |
| `go test -race -json -count=1 ./pkg/api ./pkg/services ./pkg/llm/... ./pkg/datastore ./pkg/core ./pkg/settings` |754 PASS /4 opt-in SKIP;12 tested packages PASS,2 without tests;108.22s |
| `go test -json -count=1 ./pkg/services` after test parameterization |441 PASS /2 performance SKIP;31.28s |
| `go test -json -count=1 -timeout 20m ./pkg/services -run 'TestAssistant(ExperienceHistoryPerformance\|Database100KPlans)$'` with both performance flags1 |6 PASS /0 SKIP;121.22s |
| `go test -json -count=1 ./pkg/api -run 'TestAssistantExperienceHTTPLargeExport$'` with history performance flag1 |1 PASS /0 SKIP;2.80s |
| `python3 -m unittest discover -s scripts -p assistant_migration_runtime_test.py -v` |2 PASS; real child environment and process-group timeout |

The shell command's regex above uses a literal `|` alternation (the table
escapes it only for Markdown). The two service performance flags are
`EBK_ASSISTANT_PERF_TEST=1` and `EBK_ASSISTANT_HISTORY_PERF=1`.
Full normal/race logs initially skip commercial `TestAIDeploymentIntegration`,
large HTTP export,100k ledger and history performance. The latter3 are executed
separately above; commercial deployment remains NOT RUN. Performance has not
been rerun under the race detector; its timings are normal-build measurements.
Package-level `skip` with no test files is reported separately from test SKIP.

Ledger clear/reuse/scaling: **14 exact scenarios ×3 engines =42 leaves PASS**,
both normal and race. Covers all-data/ledger/account barriers, rollback,
partial failure, interleaved new run, completed same-turn regeneration, final
CAS retry and operation-local proof reuse. Memory projection/revocation is
**3 scenarios ×3 =9 leaves PASS**, in final services and race. All existing
assistant query/storage/memory/context/worker/HTTP contracts in the selected
packages run; API httptest fixtures remain SQLite where defined, not a claim
that every HTTP test is independently parameterized across every engine.

## Performance measurements

Hardware: Apple M5,16GiB host RAM; MySQL8.0.46 and PostgreSQL16.15 in local
Linux/arm64 containers. This stage ran no other concurrent test workload during the performance
invocation. These are local fixture observations, not production capacity SLAs.

History fixture:1000 conversations,50 full chats of200 rounds with2 answer
versions;102,527,441 logical bytes (~97.78MiB),88,250,775 content bytes;
a full chat is2,011,706 logical bytes. All pages obey dual count/byte bounds
and visit the expected turns without omission/duplication.20 latency samples.

| Engine | First-page p95 | Rare-search p95 | Full export | Sampled peak heap growth |
| --- | ---: | ---: | ---: | ---: |
| SQLite |5.426ms |70.004ms |0.593s |2,489,504 bytes |
| MySQL |69.682ms |241.361ms |4.876s |2,563,080 bytes |
| PostgreSQL |72.322ms |87.091ms |4.566s |2,652,984 bytes |

Every export is96,841,192 bytes, max write4557 bytes, under the actual60s
context deadline. Heap growth is sampled Go heap growth from this fixture,
not OS RSS or a guarantee of peak process memory. Real SQLite HTTP export
additionally delivered95,622,642 bytes and the final JSON delimiter in504.21ms
under the configured60s HTTP limit. Normal API suite separately passes client
cancellation and deadline interruption.

100k-ledger tool measurements read100,014 application rows for aggregates /
100,013 for first detail (fixture includes additional boundary rows). Exact
amount and coverage assertions pass; SQL query plans are retained in raw logs.

| Engine | All-history aggregate | Month/category | First detail |
| --- | ---: | ---: | ---: |
| SQLite |218.05ms |242.72ms |197.06ms |
| MySQL |1863.52ms |1625.62ms |1242.67ms |
| PostgreSQL |1919.37ms |1982.10ms |1614.42ms |

Ledger clear of25 rows with1000 conversations asserts exactly21 conversation
page scans (50 records/page plus terminal page), versus1 scan with0 chats;
scan count is independent of deletion count. Timing and detailed per-engine
results are in the normal/race measurements JSON; race timing is not used for
performance acceptance.

## Native migration and cleanup

Migration script completed all3 engines in16.06s. Baseline is
`1a21bc9bfd2b155a3ed9b8ac57caff7c76bbb293`, built from `git archive` in a new
private task directory, without changing the shared checkout. For each engine:
old binary initializes → seed ledger/legacy embedding → current upgrade twice
→ compare legacy manifest → seed durable experience → native backup/restore
(SQLite backup API CLI, mysqldump/mysql, pg_dump/pg_restore) → identical full
manifest → old binary rollback with records preserved → re-upgrade with same
manifest → expired lease recovery preserving ledger/embedding/memory/receipts.

Run with `EBK_MIGRATION_TEST_PASSWORD` containing only a synthetic task password
and `python3 scripts/assistant-experience-migration-check.py
--mysql-container ebk-experience-verify-matrix-mysql
--postgres-container ebk-experience-verify-matrix-postgres
--mysql-port <task-loopback-port> --postgres-port <task-loopback-port>`.
Containers must already be dedicated fixtures; no shared DB is used.

Post-run inspection found no INI, database, backup, source tree or executable
remaining in the migration directory. Reviewed evidence stores hashes and
normalized database names, never backup files or credentials. Both database
containers were removed with `docker rm -fv` including own anonymous volumes.
No network was created, no prune ran, and unrelated `codex-s03-build` and
`new-api-codex-tls-poc` were left untouched.

## Final handoff

Final `gofmt -l` for the changed Go test is empty, `git diff --check` passes,
and source-line gate checks1712 files with0 over500. Python bytecode artifacts
were removed. No product change requires rebuilding the already identical
Gate16 frontend image. Existing helper tests and compatibility regressions are
included in the full42-file Vitest run; they are not added again to its total.

Reviewed JSON evidence is in
[the audit-matrix directory](../ai-assistant-experience-evidence/audit-matrix/):
`summary.json` contains actual leaf counts, explicit skip names, log hashes,
frontend results and cleanup; `commands.json` records exact argv/exit/time;
`provenance.json` records tool/database/image/artifact identity; per-run assistant
leaf inventories and measurements distinguish normal/race/performance;
`migration.json` records all3 outcomes and backup/manifest hashes. Raw logs stay
under `.vitest/audit-matrix`, outside committed evidence. Leaf counting excludes
parent tests that have descendants and excludes package summary events.

This stage is complete and frozen for parent review. Earlier skill scores and
master status are unchanged; required external release checks remain open.
