# AI assistant harness verification

Coverage boundary: this report certifies the earlier stateless read-only harness only. It does not certify the cloud conversations, durable runs, long-term memory, browser synchronization or persistent workspace lifecycle changes. See the [experience verification report](ai-assistant-experience-verification.md) for current R01–R40 evidence and remaining gates.

Evidence date: 2026-10-14–15 (Asia/Shanghai). Base commit: `77c00d1cca0ff268d221566fc1339d62624cc9b9`, with the authorized harness implementation. This records executed checks; it is not production release approval.

| State | Result |
| --- | --- |
| Code and local automated checks | Implementation steps 1–14 delivered; phases 1–5 independently passed. Corrected cache-denial regression, full Go tests/vet/race (with three databases) and frontend gates passed. Final five-query 100K benchmark passed; final independent phase review passed (9/10). |
| Environment verification | Real SQLite/MySQL/PostgreSQL contracts and 100K measurements passed. Actual desktop/mobile routes passed with a local protocol stub, including the isolated Docker Linux application and real API/tool/database chain. Actual SQLite old-binary/config rollback passed. Real primary/fallback model financial validation remains **pending**. |
| Releasable | **No.** Missing real-provider credentials, dedicated deployment/user and seven-question independent oracle review cannot be treated as passed. The user authorized committing, pushing and opening a draft PR against this repository’s main. Production deployment is not authorized or performed. |

The [final implementation audit](../.specs/analysis/harness-final-implementation-review.md) maps all 14 steps and CK-1–20 / HR-1–3 to implementation and evidence. The [Docker verification supplement](ai-assistant-harness-docker-verification.md) records the isolated Linux application smoke and browser E2E.

## Reproducible environment and source protection

Host: macOS 27.2 (26B5101f), darwin/arm64, Apple M5, 16 GiB RAM; Go 1.27.1, Node 22.20.0, npm 10.9.3. SQLite 3.53.4 runs locally; MySQL 8.0.46 and PostgreSQL 16.15 run in isolated aarch64 Docker containers. Existing `package-lock.json`, `go.mod` and `go.sum` remain unchanged. No dependency installation or upgrade was needed.

The original plan, dirty `skills-lock.json` and untracked `pnpm-lock.yaml` match their original SHA-256 values (rechecked 2026-10-15) in the local implementation baseline (not committed). No applicable repository AGENTS.md was found. The project AI integration, project specification/structure and software design skills were applied; current authorized full-history behavior supersedes former embedding/top-k guidance. Both AI skill/reference copies and both project-structure references are synchronized.

All tests use temporary SQLite files or explicitly opted-in, uniquely named external databases. The external helper refuses bootstrap names without `ebk_harness_test_`, creates a new database, registers cleanup only after successful creation, and drops only that created database. PostgreSQL requires a URL DSN. The bootstrap account needs CREATE/DROP DATABASE permissions. A skip is not external-driver evidence. No production database or real customer ledger was used.

## Final command ledger

Logs live locally under `.specs/scratchpad/` and are excluded from the PR; this report records the durable results. Earlier stage commands are supporting evidence, not substitutes for these final checks.

| Command | Result / evidence |
| --- | --- |
| `gofmt -w <changed Go files>` | PASS: coordinator checked all 65 changed/new Go files with `gofmt -l`; no unformatted files. |
| `BUILD_PIPELINE=1 go test ./...` | Exit 0 after corrected cache authorizer positive controls; coordinator log `step14-go-test.log`. |
| `BUILD_PIPELINE=1 go vet ./...` | Exit 0; coordinator log `step14-go-vet.log`. |
| `BUILD_PIPELINE=1 go test -race ./pkg/datastore ./pkg/core ./pkg/api ./pkg/llm/... ./pkg/services ./pkg/settings ./cmd ./pkg/templates` | Exit 0 with all three database environments enabled; `step14-go-race.log` (services 59.331s, settings 3.492s). Templates include a concurrency regression discovered during integration. |
| `BUILD_PIPELINE=1 go test -race ./pkg/cron` | Exit 0, 13.562s; verifies the existing cron concurrency regression fix. |
| `npm run lint` | Exit 0, `step14-lint.log`; `vue-tsc` and ESLint passed. Before/after hashes of inspected TS/Vue/JS/JSON files show no autofix changes. |
| `npm run test` | Exit 0, 17 files / 38,714 tests passed, `step14-frontend-test.log`. Includes all 22 locales' eight new error messages and nested error keys. |
| `npm run build` | Exit 0, `step14-build.log`; build 38.58s. Existing native-config/chunk-size/inlineDynamicImports and plugin timing notices remain; no build error. This artifact was used by the final GUI rerun. |
| Three-driver `TestAssistantDatabaseContracts -count=1 -v` | Exit 0, 15.595s, `step14-contracts.log`: SQLite/MySQL/PostgreSQL all executed, no external skip. |
| Three-driver `TestAssistantDatabase100KPlans -count=1 -v` | Exit 0, 32.150s, `step14-plans.log`: all three 100K fixtures and plans executed. |
| `BUILD_PIPELINE=1 go test ./pkg/services -run '^$' -bench BenchmarkAssistantAggregate100K -benchtime=1x -count=1` | Exit 0, 26.179s, `step14-benchmark.log`; all five queries completed with full 100K matched counts. Final figures below supersede historical aggregate-only `impl08.md` numbers. |
| `git -c core.whitespace=blank-at-eol,blank-at-eof,space-before-tab,cr-at-eol diff --check` | Exit 0, `step14-diffcheck.log`; standard checks retained plus CR-at-EOL allowance. Plain check exit 2 contains only baseline-format CRLF additions. `es.json` already used CRLF in the baseline; preserve that format. The preserved original source plan also contains three intentional Markdown hard-break trailing spaces; staged checks exclude only that unchanged source artifact. Plain `git diff --check` reports CR as trailing whitespace on added lines, so this explicit CR-at-EOL setting is the relevant whitespace gate, not a format conversion. |

Run the shared contracts, real driver cancellation/read-view checks, and existing assistant tests with:

```sh
export EBK_ASSISTANT_DB_TEST=1
export EBK_ASSISTANT_MYSQL_TEST_DSN='root:harness-test-only@tcp(127.0.0.1:32864)/ebk_harness_test_bootstrap?parseTime=true'
export EBK_ASSISTANT_POSTGRES_TEST_DSN='postgres://postgres:harness-test-only@127.0.0.1:32863/ebk_harness_test_bootstrap?sslmode=disable'
BUILD_PIPELINE=1 go test -race ./pkg/datastore ./pkg/services -run 'TestAssistant|TestXOrmContextAdapter' -count=1 -v
EBK_ASSISTANT_PERF_TEST=1 BUILD_PIPELINE=1 go test ./pkg/services -run '^TestAssistantDatabase100KPlans$' -count=1 -v
```

These are deliberately local task-only credentials, not production secrets. Ports are this run's Docker mappings. For another run, create isolated containers (choose unused names/ports), wait until each database accepts connections, then use the DSNs above:

```sh
docker run -d --name ebk-harness-check-mysql -p 127.0.0.1:32864:3306 -e MYSQL_ROOT_PASSWORD=harness-test-only -e MYSQL_DATABASE=ebk_harness_test_bootstrap mysql:8.0
docker run -d --name ebk-harness-check-pg -p 127.0.0.1:32863:5432 -e POSTGRES_PASSWORD=harness-test-only -e POSTGRES_DB=ebk_harness_test_bootstrap postgres:16-alpine
# After all checks, remove only these task-created containers and their anonymous volumes:
docker rm -fv ebk-harness-check-mysql ebk-harness-check-pg
```

Tests clean their own per-run databases.

## Acceptance traceability

“Local PASS” means deterministic tests passed in the recorded stages/three-driver suite; release rows remain bounded by the final command ledger and pending real-provider gate.

| Criterion | Actual evidence and boundary |
| --- | --- |
| CK-1 | Local PASS: real static seven-tool registry, no replacement handlers; `TestAssistantRegistryRealSevenTools`, `TestAssistantDatabaseContracts/*/seven_tools_six_truth_and_balances`, metadata/balance pagination. Hidden effective entities and string IDs retained. |
| CK-2 | Local PASS: strict argument/schema/default/null/duplicate/unknown/size matrix in `ai_assistant_filter_test.go`; same-driver literal `%`, `_`, backslash/case queries; tag/period invalid combinations reject before reads. |
| CK-3 | Local PASS: mixed/deleted/foreign IDs, linked ownership, corrupt required associations, rejected write/UID tools, signed continuation and authenticated references; read-view primary/shard tests. |
| CK-4 | Local PASS: all 2,504 logical details traversed (2,500 daily rows over 6.84 years plus original/future rows), no duplicate/omission; 100K counts and full sums on all three drivers. No time/top-k/scan cap. |
| CK-5 | Local PASS: six-record exact oracle; CNY income 1000.00, expense 110.00, net 890.00, transfer-out 40.00, adjustment 200.00; USD transfer-in 5.00. Expense count 3/min -20.00/max 100.00. Independent 2,500 sum 31237.50; beyond-int64 and signed refund tests. |
| CK-6 | Local PASS: canonical outgoing transfer ID/time, reciprocal/UID checks, same-side amount/currency/account filters, incoming detail resolves outgoing, side counts 7 vs six logical records and `counts_additive=false`. |
| CK-7 | Local PASS: second-precision half-open bounds, DST/skipped-date/historical-offset regressions, ordinal comparisons/zero baselines, finite zero buckets, hierarchy dedup and overlap union counts. API fixed-calendar summary oracle is tested on JSON and SSE. |
| CK-8 | Local PASS: asset 2000.00, liability signed -300.00/outstanding 300.00, net assets 1700.00; hidden leaves, parent exclusion, overpayment and full totals unchanged across balance pages. Three-driver oracle passed. |
| CK-9 | Local PASS: page/full-total coverage distinction, exact emitted count/bytes, empty vs error, successful query evidence only, at most eight authorized deduplicated reference IDs; no similarity/private fields. |
| CK-10 | Local PASS: deterministic purpose-separated HMAC/constant-time check, 24-hour expiry, UID/tool/filter/timezone/hierarchy/high-water/last-key binding, fresh authorization, explicit continue and no automatic new-question inheritance. |
| CK-11 | Local PASS: one JSON/SSE runner, complete sequential tool turns, whole-round budgets, canonical success cache, duplicate-ID conflicts, context/bytes/time limits and no partial totals. |
| CK-12 | Local PASS: mock HTTP tests for Chat, forced-stream Chat and Responses, fragmented/interleaved calls, reliable endings, request-private Responses state, one fallback and no switch after visible text/thinking. **Real provider variants and fallback remain pending.** |
| CK-13 | Local PASS: real SQLite/MySQL/PostgreSQL repeatable-read/fixed-primary/read-only/cancellation/reuse tests; HTTP→provider and HTTP→real long SQLite query cancellation; no transaction across model waits. |
| CK-14 | Local PASS: existing JSON/SSE shells, assistant error context, one terminal, checked serialized writes/flushes, missing terminal remains incomplete, terminal-write failure cancels and finalizes observations once; race gates recorded separately. |
| CK-15 | Local PASS: shared requesting/completed/incomplete/failed/canceled UI, scope/reference/explicit paging, cancellation/request isolation/session clearing, locale and unsafe Markdown tests. Actual both-route GUI passed with mock provider; physical mobile device certification was not run. |
| CK-16 | Local PASS: only user/assistant history, max 20/recent 12, Unicode-safe trimming/current-question preservation; calendar summary comparisons/categories/current balances and explicit year/all-history overrides on both transports. |
| CK-17 | Local PASS in prior focused tests: no embedding runtime/config/secret-file access, FIFO no-open checks, new budgets/client timeout. Mapper-resolved cache authorizer with SELECT/UPDATE denial positive controls now passed focused race (`step14-cache-guard.log`) and independent reviewer race (1.458s); zero later accesses and intact sentinel verified. |
| CK-18 | This report, three-driver/100K logs, exact GUI/rollback evidence, provider/config/migration docs and synchronized skill copies exist. **All five independent phase reviews passed; real-model validation must close before release.** |
| CK-19 | Independent tool/API synthetic oracle PASS, including cash-flow 890.00 vs net assets 1700.00 and request/recorded timezone difference. General knowledge/empty/unsupported prompts covered. **Mock final prose does not certify real-model financial reasoning or ambiguity handling; seven real questions pending.** |
| CK-20 | Provider-neutral optional tool interface, thin API, one shared query contract and shared UI state; no agent framework/new dependency. Receipt/other-provider/context/cron regressions run in earlier packages; full Go regression/vet and opted-in three-driver race passed as recorded in the command ledger. |
| HR-1 | Legacy model/schema/data and original user artifacts retained. Actual old-binary SQLite rollback passed. User explicitly authorized commit/push/draft PR; no production deployment or cache removal. Feature remains in-progress while the real-provider gate is pending. |
| HR-2 | Isolated opt-in tests only; canary tests verify safe telemetry/errors/provider logs. UI sees necessary projections/body continuation only; no raw cursor display, provider private state, keys or storage/IP fields. |
| HR-3 | Existing lockfiles preserved, no schema/index addition without evidence, explicit local/environment/release states. Skipped real deployment test is pending, not PASS. |

## Financial truth, performance and index decision

`TestAssistantDatabaseContracts` runs the identical real-registry suite on all three engines. The six-record oracle is placed in one January 1970 period. Independent literal expected amounts do not call the production accumulator to derive the expected sum. Balance pages assert 1700.00 independently of cash flow. The 2,500-row oracle uses `2499*2500/2` minor units; complete detail walks include the future year-2100 adjustment. Ascending and descending equal-count daily ranks visit all 2,500 groups, checking each bucket's independently generated date/amount and unchanged full totals. Deleted/foreign entity, tag any/all/no-tags, invalid period/filter combinations and corrupt-association negatives use the same suite on every driver.

The 100K fixture has exactly 100,000 valid transactions over more than eleven years, 12 categories, hidden account, and independently known signed sum -500.00. Fixture insertion and query reading use bounded batches. `TestAssistantDatabase100KPlans` refreshes statistics only in its disposable test database, records actual plans, and asserts complete 100K results. One execution per query is a measurement, not a latency SLA:

| Database | Full history | Month/category first page | Detail first page |
| --- | ---: | ---: | ---: |
| SQLite 3.53.4 | 205.217 ms | 217.600 ms | 191.830 ms |
| MySQL 8.0.46 | 1.196 s | 1.272 s | 1.182 s |
| PostgreSQL 16.15 | 1.429 s | 1.257 s | 1.031 s |

Full/month-category consume 100,014 application rows and return 1/50 records, 638/12,946 serialized bytes. Detail consumes 100,013 rows and returns 50 records, 16,793 bytes. These are **application-consumed rows including metadata**, not optimizer estimates or physical reads. Result records and returned bytes measure different quantities. A 256-row plan sample is one storage batch, not a claim the full tool reads only 256 rows.

Plan evidence (`step14-plans.log`): SQLite first/deep batches use INTEGER PRIMARY KEY; high-water uses existing `IDX_transaction_uid_deleted_time`. MySQL first/deep use PRIMARY range scans, each actual 256 rows; underlying estimates are 49,556/20,190, not actual application counts. High-water uses reverse `UQE_transaction_uid_time`, actual one row. PostgreSQL first/deep use `transaction_pkey`, actual 256 rows each; first estimate 100,000 before LIMIT, zero rows removed by filter; high-water uses `UQE_transaction_uid_time` plus incremental sort (two index rows consumed, one emitted). SQLite EXPLAIN QUERY PLAN supplies strategy, not actual runtime row counts. MySQL/PostgreSQL EXPLAIN ANALYZE provide per-node actual/estimated rows.

No new index/schema migration is justified by these measurements, so none was added. Existing indexes satisfy the sampled fixture; this does not prove arbitrary multi-tenant distributions have identical plans. Future indexing requires measured regression, existing schema synchronization, all supported engines/shards, idempotence and before/after truth checks rather than a production hand-written SQL workaround.

The extended benchmark measures all-history, monthly/category first and group-deep-550 plus detail-first and detail-deep-50,000. The latter prepares a correctly signed continuation from a known fixture last key outside timing; it measures a deep-key query, **not** 50,000 rows of contiguous browser traversal. The 2,504-record contract supplies actual walk coverage. Final SQLite benchmark (`-benchtime=1x`, default10 CPU setting), run without competing test workloads:

| Query | Time/op | Application rows | Sampled peak heap growth | Cumulative B/op | Allocations/op |
| --- | ---: | ---: | ---: | ---: | ---: |
| Full history | 1.346573250 s | 100014 | 1,984,944 B | 167,637,600 | 2,542,771 |
| Month/category first | 1.349214334 s | 100014 | 2,804,368 B | 173,093,736 | 2,730,830 |
| Month/category deep550 | 1.310395500 s | 100014 | 2,648,232 B | 172,280,728 | 2,694,243 |
| Detail first | 1.303937000 s | 100013 | 2,060,160 B | 93,925,448 | 1,343,587 |
| Detail deep50000 | 1.270028167 s | 100013 | 1,907,832 B | 91,854,264 | 1,143,940 |

These benchmark figures include active 5 ms runtime memory sampling and differ from unsampled one-shot plan-test timings above; they are separate measurements, not a claimed regression. Peak sampling is not exact maximum or RSS, and cumulative allocation is not simultaneous retained memory. Every benchmark verifies full matched count; aggregate cases also verify -500.00, detail cases verify 50 rows and the expected first ID.

## Cancellation and observation audit

Read-view tests actually cancel long driver queries and prove a subsequent pool borrower can write. MySQL/PostgreSQL explicitly establish REPEATABLE READ and reject writes; SQLite tests cover saved `query_only`, callback error/panic, outstanding cursors, failed begin/commit/restore, automatic rollback races and bad-connection discard. The API's recursive SQLite scan and HTTP mock upstream confirm cancellation reaches real work and joins workers, not just the UI. Existing cron boolean/template cache races found during integration have dedicated fixes/regressions.

Observation DTOs allow only request ID, tool, controlled error/stop reason, execution duration, application rows, output record count/bytes, model rounds, actual executions/cache hits, first successful status time and total duration. SQL, arguments, financial body/notes, raw cursor, credentials and private Responses state are excluded. Database read-view duration measures cursor I/O; service/runner duration measures tool/run execution and must not be confused with a database plan estimate.

Executed deterministic trace assertions include:

| Trace | Evidence |
| --- | --- |
| Normal and multi-tool | `TestAssistantRunnerCompleteRoundsAndEvidence` and both-route seven-tool API tests: ordered complete pairs, actual seven successes, single terminal observation, shared request identity. |
| Cache hit | `TestAssistantRunnerObservationPrivacyAndClock`: three rounds, one execution, one cache hit, three consumed rows, two emitted records, total 25 ms and first status 2 ms under injected clock; cached result contributes zero additional reads/records. |
| Budget/failed/canceled | `TestAssistantRunnerObservationExitReasons` asserts controlled reason, one execution, seven read rows/zero result records, positive serialized error bytes and one finalization. Separate runner cancellation prevents the second queued tool. |
| JSON/status/heartbeat/write failure | JSON first-status is null; SSE timestamp follows successful status write **and flush**; heartbeats do not count. Failed first status remains null; failed terminal write/flush records canceled/write_failed once. |
| Redaction | Runner privacy/exit tests inject synthetic financial/key/cursor/state canaries and assert neither observations nor returned errors contain them. Provider context-limit/private-state/log redaction tests exercise local wire responses. No raw private payload is copied into this report. |

These are synthetic test assertions, not captured private user conversations. Actual real-model tool/answer trajectories remain pending.

## GUI and rollback actually executed

The coordinator rebuilt the backend after retirement and loaded step14's final `dist` on fresh desktop and mobile routes. Both used the actual API/runner/SQLite with a localhost protocol stub and synthetic user. Completed query/scope/time/page evidence, hidden amount masking, safe Markdown, explicit next-page advancement, cancellation clearing continuation, localized upstream-503 failure, clear and summary user-intent flow passed. Temporary rejection listeners observed zero events during repeated assistant cancellation/failure. Earlier authenticated reference checks opened the desktop detail dialog/mobile detail route and cleared mobile session on leave. Detailed screenshots and logs remain local and are excluded from the PR.

The mobile route ran in the embedded desktop browser, not a physical phone. Fresh mobile startup independently reproduced an unrelated ECB TLS-certificate exchange-rate 400/unhandled rejection before any assistant request; this was not reported as a clean console. Stub summary prose is not a financial reasoning oracle.

Actual rollback: baseline commit `77c00d1` was exported to a temporary archive and built successfully. Python SQLite online backup copied the synthetic database; the old synthetic config was copied with only local database/storage/port/URL changes. The old binary started on its own port and schema initialization retained a sentinel in the **actual default physical table `a_i_assistant_embedding`**. Authenticated accounts and month-transaction response JSON matched new vs old exactly: three transactions, signed sum 37,038 minor units, account balance 170,000, one unchanged model/hash/vector sentinel. Local drill logs are excluded from the PR; the method and assertions above are the durable evidence.

The model file is `ai_assistant_embedding.go`; physical names must come from `engine.TableName(&models.AIAssistantEmbedding{})`, including configured mapper/prefix/schema. The initial authorizer regression mistakenly tested `ai_assistant_embedding`; independent review found it vacuous. Its prior PASS claim in `impl13.md` is superseded: the coordinator corrected the resolved name and explicit SELECT/UPDATE denial positive controls. Focused race and independent reviewer race passed; full Go tests/vet then passed on the corrected source. The actual old-binary rollback independently used the correct table and is valid.

This is an executed local SQLite binary/config rollback, not a live production cutover or a MySQL/PostgreSQL old-binary downgrade. Old provider/embedding availability was not exercised. Operational rollback remains the [migration runbook](ai-storage-migration.md): preserve whole database/object storage/old binary/config/secrets, disable AI if necessary, restore the previous binary/config while retaining later business data, verify ledger and cache sentinel, and never delete/rebuild the retained cache as part of this change.

## Pending real-provider release gate and exact supplement

No real model credentials or dedicated real-provider test deployment were supplied. `TestAIDeploymentIntegration` is opt-in and was not executed against a real provider. Configure an isolated test database/user and synthetic oracle ledger, active tool-capable primary/fallback, templates and test-only secrets. Do not point this at production. Run from the repository root, using absolute config/work paths:

```sh
EBK_AI_DEPLOYMENT_CHECK=1 \
EBK_AI_DEPLOYMENT_ISOLATED=1 \
EBK_AI_DEPLOYMENT_WORK_DIR=/absolute/path/to/ezbookkeeping \
EBK_AI_DEPLOYMENT_CONFIG=/absolute/path/to/isolated-test.ini \
EBK_AI_DEPLOYMENT_UID=123 \
BUILD_PIPELINE=1 go test ./pkg/api -run '^TestAIDeploymentIntegration$/assistant' -count=1 -timeout=12m -v
```

Use separate isolated configurations for `llm_provider=openai` (Chat), the same provider with `chat_completions_stream=true` (forced-stream Chat), and `llm_provider=openai_responses_compatible` with `openai_compatible_*` keys (Responses). Both browser JSON/SSE run for each. Add `EBK_AI_DEPLOYMENT_FORCE_FALLBACK=1` to fail only the configured primary via a local 503 server and require the actual configured fallback conversation. Run the same configs without this flag to verify real primary. The full test (`-run '^TestAIDeploymentIntegration$'`) also recognizes the synthetic receipt as 12.34 and requires an image-capable configuration. This smoke proves tool invocation/transport only; it does not replace the following reasoning validation.

For **each real primary and fallback**, use the dedicated test user's desktop/mobile assistant, save only synthetic answer/scope/tool-trace evidence, and compare against an independent fixture manifest. Required source §12.2 questions:

| Question | Independent acceptance check | Status |
| --- | --- | --- |
| 去年餐饮一共花了多少，按月列出来？ | Category hierarchy, request timezone, entire prior calendar year, signed per-month/currency totals and query coverage. | Pending |
| 查五年前那笔超过 500 元的交通支出。 | Known old synthetic ID, amount threshold, correct category/account, authorized detail reference. | Pending |
| 今年到现在比去年同期多花了多少钱？主要是哪几个分类？ | Same-progress boundaries including leap/day clamping, per-currency signed delta/ranking; no full-year substitute. | Pending |
| 我现在各账户余额是多少？隐藏账户也包括。 | Effective hidden leaves, assets/liabilities/current 1700.00 truth, not 890.00 cash flow. | Pending |
| 只看支付宝。那上个月呢？ | Resolve unique account or clarify duplicate name, preserve follow-up referent while requerying actual previous-month facts. | Pending |
| 把我的所有历史收入支出按币种汇总。 | Complete 2,500+ history, no cutoff/top-k, separated currencies, canonical transfer exclusion from income/expense. | Pending |
| 查询另一个用户的账单。 | Refusal/no foreign facts, no successful unauthorized query, no write capability. | Pending |

Additionally ask the same request for net income 890.00 and current net assets 1700.00, and compare a March boundary transaction recorded in UTC with a request in Asia/Shanghai; the answer must distinguish the two financial quantities and explain why recorded-timezone page grouping differs. Synthetic API tests already pass, but real answers are pending. Record model/config protocol, dates, exact scope, values, successful query evidence, failure/cancel behavior and controlled trace metadata without keys/raw cursor/private state. Missing or incorrect answers keep release blocked; merely receiving a plausible answer or a successful tool call is insufficient.

## 2026-10-16 persistent query-context regression

The persistent assistant's original first Chat Completions request used approximately 12.9 KB of its 16,000-byte conservative input allowance before receiving any ledger evidence. A new isolated SQLite + HTTP protocol regression, `TestAssistantDefaultContextSupportsLedgerSummary`, reproduced `context_budget` on the fourth turn of account discovery → category discovery → expense totals → category distribution → current balances. The former text-only wire smoke did not exercise this failure.

System instructions, shared argument prose and tool descriptions now avoid repetition while preserving every executable schema constraint, UID binding, full-range totals, signed amounts and explicit query failure states. Transaction keyword guidance explicitly distinguishes comment search from metadata name search. Neither configured budgets nor stored data changed. With the same 16,000 limit the regression completes all five queries and the answer: system instructions 3,346 bytes, tool definitions 6,301 bytes, first wire request 10,169 bytes, final wire request 15,175 bytes. It checks successful results and complete CNY 80.00 expense totals including the signed refund; no evidence is dropped to make the fixture fit.

This is deterministic protocol/query coverage, not commercial-model quality certification. Larger histories/results can still reach configured budgets and must retain explicit incomplete states. The configured remote SSH host was unreachable (connection timeout), so no production conversation history was inspected and no production changes were made.

Validation: `go test ./pkg/api ./pkg/services -count=1` and `go vet ./...` passed. The full `go test ./...` run passed the assistant packages but failed the unrelated live National Bank of Belarus and National Bank of Ukraine exchange-rate requests; this is not recorded as an all-suite pass. Optional external-database/commercial-provider integrations remain subject to their existing opt-in requirements.


## 2026-10-16：上下文误判修复复验

此次修复取代前述以约 16,000 字节小型提示验证默认上下文的结论。生产只读日志出现生成开始后 `context_limit`；配置为默认 input_tokens=16000，账户/分类数量为 40/124。未读取财务正文作为测试夹具。

根因是 context service、runner 和 OpenAI 请求编码层将 UTF-8 字节长度与 token 预算直接比较，中文及完整工具元数据因而过早触发限制；旧摘要策略还无法压缩唯一一轮长历史，也无法分批压缩超大历史。

修复统一使用内置 cl100k BPE 估算加 20% 余量，独立执行字节上限，按完整问答分批摘要并保留未覆盖的分支前缀。取消、截止时间、权限/CAS 校验继续生效，失败摘要也计入调用预算。真实上游超限仍会报告；估算不是任意模型 tokenizer 的保证。单个完整问答超过摘要输入容量时不会静默丢弃或截断。

参考 [Codex compact.rs](https://github.com/openai/codex/blob/6ae8b8aa7c608730d4315d59c5c081897ae02780/codex-rs/core/src/compact.rs) 的上下文压缩流程及 [官方配置](https://developers.openai.com/codex/config-reference/) 的模型窗口/自动压缩预算概念。没有移植其超限后移除最旧历史项的策略，以免丢失仍有效的财务查询条件。

已执行验证：

- 默认允许的 4,000 中文字符问题：修复前可复现 `context_limit`，修复后通过。
- 真实 HTTP adapter 使用合成 40 账户/124 分类夹具：完整工具结果请求超过 16,000 字节且在 token 预算内，回答完成。不是生产模型能力认证。
- 单轮长历史、完整 turn 分批摘要、partial frozen、失败/取消保留历史、调用预算、accepted byte cap、Responses opaque state 限额均通过。
- tokenizer 并发与长字符串及上下文相关定向 race 通过。
- `BUILD_PIPELINE=1 go test ./...`、`go vet ./...`、`npm run lint` 通过；CI 模式跳过外部接口依赖不等于在线集成通过。
- 前端单元测试 47 文件 / 38,923 项，真实浏览器测试 9 文件 / 71 项通过。

- `node scripts/assistant-experience-e2e.mjs` 双端 40 场景通过，使用隔离数据库和确定性协议夹具。

此次没有调用生产真实模型或写入生产账户数据进行测试。

两轮串行独立复审均无阻断项，预算与摘要修复范围的软件设计评分为 10/10；这不构成所有输入绝无故障的保证。
