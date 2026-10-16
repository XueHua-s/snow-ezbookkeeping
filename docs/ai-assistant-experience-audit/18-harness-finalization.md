# 18 · D8 harness lifecycle correction

2026-10-16. Serial follow-up to Gate01's D8. This stage changes only validation
scripts/tests; no product code, release Dockerfile, master audit manifest,
commit/push, delegation, private configuration or real-user data. Applied the
project-spec/global-prompt, AI integration, and Vitest guidance. Gate01 owns
final design re-review; this report does not assign a replacement design score.

## Ownership and failure behavior

`assistant-bounded-command.mjs` owns a command's POSIX process group, deadline,
stdout/stderr buffers and event handlers. Callers supply command/arguments and
only the differing time budget/environment/output destination. Default is
30 seconds; frontend, Go and Docker builds use 10 minutes, production page
smoke uses 3 minutes. sqlite3 and version/inspect probes use the default.
Timeout/abort/SIGINT/SIGTERM sends TERM to the group, escalates to KILL after
500 ms, and rejects so the caller's finally runs. Successful direct-child exit
also kills remaining group descendants. Spawn/close failures remove handlers;
process exit has a group-kill fallback. Captured output is truncated at exactly
16 MiB, and no further bytes are accumulated after a failure. This is a
macOS/Linux task helper, not a Windows process-tree implementation.

`assistant-build-resources.mjs` owns the private source-copy directory and
random UUID builder/probe names. Names are registered **before** Docker runs,
so an uncertain create is still a cleanup target. Normal finally explicitly
runs `docker rm -f` for each name, regardless of `--rm`. Only a positive
no-such-container/object result establishes absence. Unknown daemon errors
remain errors, retain exit recovery, and do not prevent directory cleanup.
The build preserves both original execution errors and cleanup errors.

Build SIGINT/SIGTERM exits through a bounded synchronous recovery: first kill
active host command groups, then remove each owned container with a 10-second
CLI limit/SIGKILL, then remove temporary source. Recovery failures are printed
and retain a nonzero exit result. No wildcard name matching, network/cache
prune or unrelated-container operations are used.

The boundary is wired into all D8 call sites: validation builder run/capture,
Docker smoke production subprocess and sqlite3, main E2E production gate,
and environment's temporary host Go build. The long-running host application
keeps its existing server lifecycle; it is not assigned a command deadline.
The existing Docker runtime is not reabstracted.

## Red/green and real-process evidence

Focused final run: **5 files / 21 PASS**. Ten new tests cover:

- Real deadline with a TERM-resistant subprocess and grandchild; both end.
- ENOENT startup and listener cleanup.
- Real SIGINT and SIGTERM, descendant cleanup and caller finally execution.
- Continuous stdout/stderr flooding after TERM; capture stays at 16 MiB.
- Controlled executable Docker CLI failure after simulated daemon creation;
  exact attempted names are removed, including a never-created probe.
- Unknown daemon cleanup failure stays observable while source is removed and
  exit recovery remains registered.
- Build signal exit recovery kills host descendants before daemon cleanup.
- Exit-fallback daemon errors print an error, preserve nonzero exit status and
  still remove source.

Three controlled mutations were actually executed and restored in finally:
removing the effective deadline caused 1 failure; omitting explicit daemon
cleanup caused 2 failures; removing output truncation caused 1 failure. Final
21-test rerun passed. The fake CLI tests are process-level failure simulations,
not claims of a real daemon outage.

An additional **real Docker** probe started cached busybox **without --rm**,
printed READY, hit the 1000 ms CLI deadline, and confirmed the daemon container
still existed after the CLI stopped. `buildResources.close()` then removed it
and its temporary source directory; inspect confirmed absence. The first
probe's absence assertion expected uppercase `No such object`; actual Docker
inspect emitted lowercase. That assertion was corrected and rerun successfully;
explicit cleanup had already worked, and no product/helper repair was needed.

## Actual build, smoke and equivalence

Executed with raw artifacts under `.vitest/final-build-cleanup/`:

```sh
EBK_E2E_ARTIFACTS=.vitest/final-build-cleanup/build \
EBK_E2E_DOCKER_IMAGE=ebk-assistant-validation:20261016-bounded \
node scripts/build-assistant-docker-validation.mjs

EBK_E2E_ARTIFACTS=.vitest/final-build-cleanup/smoke \
EBK_E2E_DOCKER_IMAGE=ebk-assistant-validation:20261016-bounded \
node scripts/assistant-experience-docker-smoke.mjs
```

Both commands exited 0. The current wrapper built fresh frontend and Linux CGO
backend, completed actual version/image probes, and cleaned builder/probe
containers. Smoke completed **10 PASS scenarios**, including production
subprocess/sqlite integration and the Gate16 upstream-response-close assertion.
No old results were substituted for this smoke.

New image:
`sha256:b4e96d3244f696663aad9ff84a9c5cfd4c850371ec0a0ed14402bda735fcbd81`.
Against Gate16's final build provenance:

| Artifact | Current SHA-256 | Comparison |
| --- | --- | --- |
| dist, 107 files | `ba13bcd1e34f10a293a2717b7a48558c524c9ad16340cc07ae9614b971e5eead` | Identical |
| Linux executable | `e5af7ae3ab00bf3930964e0f7c02aacd5b12dca41910effa15ecb967c56c5f63` | Identical |
| Source tree including Go tests, 789 files | `55a56039cf67ad2043933ff0c85ecb575d25475089beee025977766eceda0d07` | Different; source equality is not claimed |

The source-tree hash includes the existing later test edits; this stage changed
no product files. Exact executable/dist equivalence permits retaining Gate16
business E2E evidence without rerunning that whole matrix. The validation
runtime still uses cached Go/Alpine 3.24.2 with extra Go tools, actual project
entrypoint, UID/GID1000 and host Node22 frontend build. Stock Dockerfile,
Node26 stage and exact stock Alpine image remain NOT RUN.

Parent independently executed final `npm run test`: **44 files / 38,917 PASS**
(`/tmp/ebk-parent-final-unit.log`); this child did not repeat that full run.
Six runner/helper syntax checks, child and parent diff checks, and source-size
checks (1716 files, zero over 500) passed.

## Handoff

Reviewed JSON and log hashes are in
[summary.json](../ai-assistant-experience-evidence/audit-harness-finalization/summary.json)
and neighboring build/runtime/smoke JSON. Raw logs remain only in `.vitest`
and the cited parent `/tmp` files. Earlier Gate15/16 provenance is unchanged.
No task builder/probe/app container or temporary credential/source directory
remains. No network was created. `codex-s03-build` remains untouched. The new
validation image is retained; use its immutable ID if another smoke is needed.
Commercial models, physical input/manual readers and actual bfcache remain
outside this correction's evidence. Ready for parent and Gate01 re-review.
