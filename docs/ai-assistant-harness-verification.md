# AI assistant harness verification

Coverage boundary: the initial sections record the earlier stateless read-only harness. Later dated additions record their stated persistent-assistant changes and supersede the historical release status below. See the [experience verification report](ai-assistant-experience-verification.md) for R01–R40 evidence and remaining gates.

Evidence date: 2026-10-14–15 (Asia/Shanghai). Base commit: `77c00d1cca0ff268d221566fc1339d62624cc9b9`, with the authorized harness implementation. This records executed checks; it is not production release approval.

| State | Result |
| --- | --- |
| Code and local automated checks | Implementation steps 1–14 delivered; phases 1–5 independently passed. Corrected cache-denial regression, full Go tests/vet/race (with three databases) and frontend gates passed. Final five-query 100K benchmark passed; final independent phase review passed (9/10). |
| Environment verification | Real SQLite/MySQL/PostgreSQL contracts and 100K measurements passed. Actual desktop/mobile routes passed with a local protocol stub, including the isolated Docker Linux application and real API/tool/database chain. Actual SQLite old-binary/config rollback passed. Real primary/fallback model financial validation remains **pending**. |
| Releasable | **No.** Missing real-provider credentials, dedicated deployment/user and seven-question independent oracle review cannot be treated as passed. The user authorized committing, pushing and opening a draft PR against this repository’s main. Production deployment is not authorized or performed. |

The [Docker verification supplement](ai-assistant-harness-docker-verification.md) records the isolated Linux application smoke and browser E2E.

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

## 2026-10-16 实时正文与生成状态回归

此前 provider 只向展示层发送思考增量，正文在完整回答结束后才送达。Chat Completions 与 Responses 现在直接传递正文增量，runner 统一处理工具前言撤回、最终补尾、异常终态和禁止已展示正文后的备用重试。交互式请求自动启用上游流式传输。

双端先显示处理摘要与生成状态，正文更新保留 DOM；结束或取消立即停止活动指示，过渡遵循 reduced-motion。

验证：Go 全套测试及 vet、LLM 包 race、lint、前端单元测试 47 文件 / 38,924 项、真实浏览器测试 10 文件 / 74 项通过。首次浏览器全套出现一次对话框点击超时，独立复测 7 项和全套重跑 74 项通过。`node scripts/assistant-streaming-e2e.mjs` 四场景使用真正阻塞的合成 HTTP 上游，验证桌面/手机在上游结束前显示正文、刷新恢复、完成不重复，以及停止时关闭上游且保留部分回答。测试不调用生产模型，不代表供应商可用性认证。

两轮串行独立复审通过；检查范围为本次流式传输与生成体验变更，不能据此保证所有输入均无故障。

## 2026-10-16 首轮大聚合与工作上下文复审

用户反馈的新会话在首轮跨年消费问题上失败；记录显示完整范围聚合已成功，随后模型循环因 `context_budget` 停止，尚无正文。旧日志未区分本地请求预算和供应商明确上下文拒绝，因此不能仅凭该日志断言是哪一个上游分支。

此次审查确认了此前测试遗漏：工具结果尽量填满余量、同批后续工具仅留错误占位、工具轮之间没有整理，以及分页预测使用虚构 query_id。单纯历史摘要无法解决新会话的此类失败。

修复在工具轮之间使用无损 `column_rows` 表格编码，保留每个值、完整汇总、查询范围、分页标记及调用配对；原始缓存和证据不变。工具页按剩余预算分配，并给后续分析留空间。分页预测使用最终 query_id/read time，模型请求以实际供应商序列化检查为准。供应商明确拒绝上下文且尚无正文时，只允许一次减少输出预留的重试，仍受原始期限和轮数限制；不会删除财务事实或无限重试。脱敏 observation 区分 local_input、tool_result、upstream_input 及恢复次数。

查询接口仍是七个 UID 绑定的只读工具。总体统计直接聚合，元数据仅用于解析具体名称过滤；消费变化先比较总额，再按需分析分类。明细页不能充当完整账本总额或完整类别排名的证据，不提供任意 SQL 执行能力。

参考 [Codex MidTurn 自动压缩](https://github.com/openai/codex/blob/4eb4a65407d98cebc4c83bf13be440b9fbc78a7c/codex-rs/core/src/session/turn.rs#L659) 与 [压缩请求超限处理](https://github.com/openai/codex/blob/4eb4a65407d98cebc4c83bf13be440b9fbc78a7c/codex-rs/core/src/compact.rs#L359)，以及[官方配置说明](https://developers.openai.com/codex/config-reference/)中的上下文、自动压缩和工具输出预算。这里借鉴工具轮内整理；没有直接移植删除最旧财务证据的策略。

新增 `scripts/assistant-context-e2e.mjs`：真实 SQLite、HTTP adapter、桌面/手机，50 分类×跨年月份比较与第二个总额工具同批调用，核对完整范围精确金额和分页标记。使用合成账本和确定性模型协议，不代表商业模型工具规划质量验收。空正文失败不再误报“已显示部分内容不可靠”，同会话问题与重试入口保留。

本轮执行：`BUILD_PIPELINE=1 go test ./...`、`go vet ./...`、`npm run lint` 通过；前端 47 文件 / 38,924 单测、10 文件 / 75 浏览器测试通过；双端原有 40 场景、流式 4 场景、大聚合 2 场景通过。新增后端真实 SQLite 9,183 笔交易测试包含大聚合、后续精确总额查询、一次明确上游拒绝及有界恢复；相关 race 检查通过。真实模型授权尚未确认，因此未进行供应商推理调用，也未重放用户的生产对话。

## 2026-10-16 多轮查询工作集与失败意图续接

后续生产案例暴露了上一轮验证的边界：两段较长历史回答之后，总额、同期、分类、月份与明细查询会连续增加工作上下文。查询本身可以成功，但下一工具的结果页无法再进入模型输入，最终以 `tool_result/context_budget` 停止。此前单次大聚合回归并不能覆盖这条路径。

本次将单页工作集预算改为总输入额度的最多八分之一，并先保留完整汇总与查询元数据，再分配可分页的 Data/Comparison；准备阶段在历史达到约一半预算时开始现有的有界摘要。工具轮最多执行两次确定性 checkpoint 整理：所有 system/user 消息、调用/结果配对、完整 Totals、MatchedCount、Scope、Periods、Coverage、Warnings 和 QueryID 保留。旧 Data/Comparison/References 标记为已归档、当前不可引用；模型需要时可重读原七工具。原执行缓存、账本与展示证据不变，不使用模型自由摘要替代工具中的精确金额。重建后清空 opaque provider state；同一归档结果最多获得一次额外缓存重读机会，原轮数、查询期限与取消约束继续生效。最终不能恢复的请求仍保留明确失败状态。

历史续接同时修复了“继续查询”丢失未完成问题的缺口：失败回答不能成为事实，但原用户问题会按时间顺序保留为 question-only turn，并进入同一历史摘要/`SummaryThrough` 管理，避免积累在较新摘要后覆盖用户的新更正。无需修改数据库结构或迁移账本。

新增合成回归经过 Prepare、真实 SQLite 查询和实际 HTTP 序列化：两段长历史后执行总体、同期、分类、月份、分类元数据、交易明细六个工具轮，再输出结果；逐轮检查输入预算、完整范围金额、负退款、独立币种以及逻辑转账。同一测试复制到独立的修复前 HEAD 源码后，在第五轮稳定触发 `tool_result/context_budget`；修复后普通路径与一次明确上游上下文拒绝路径均完成。另有并行 call/result 配对、原缓存重读、证据去重、循环上限、取消以及未完成问题摘要的回归。

展示层移除技术性的“查询依据”展开区，保留真实账单引用、去重后的数据警告和分页范围提示；发送命令捕获用户语言与时区。提示词要求用用户可读的日期说明必要比较范围，默认不输出内部工具名、queried_at 或 RFC3339 查询时间；内部精确范围与证据不变。

验证仍使用合成账本和确定性 HTTP 模型协议，未调用真实供应商模型、未重放用户账本。不能据此声称商业模型对任意问题的工具规划都已验证。

本轮验证：`BUILD_PIPELINE=1 go test ./...`、`go vet ./...` 通过；新增精确金额独立算术断言定向回归通过；LLM/服务层工作集、runner、摘要与未完成意图相关 race 回归通过。前端 lint、生产构建、38,927 项单测通过；浏览器全套 11 文件 / 78 项通过。生产模式双端体验 E2E 为 43 PASS、2 NOT RUN：Chromium 桌面和手机没有进入 bfcache，因此这两项不计通过，普通历史恢复通过。另行执行的双端上下文 E2E 2 场景与流式 E2E 4 场景通过。执行记录保存在本地测试 artifacts；审查与生产部署结果由后续独立复审记录补充。

独立复审完成：前端末页范围提示遗漏已修复并再次通过；后端与最终交叉审查未发现新的阻断项，复审独立运行工作集恢复、失败意图与六轮 HTTP 回归通过。发布脚本同步补齐手动回滚前的后台助手任务排空检查；发布/回滚只切换代码，使用当前数据库的新备份，不恢复旧账本。


## 2026-10-16 简化消息操作

移除消息编辑与分支图标，以及对应的 props/emits、编辑状态和输入框编辑模式。保留复制、重试、重新生成、版本切换、账单分页；容量续接和历史分页仍使用原分支持久化能力。历史列表的会话重命名保留。

两轮独立只读复审通过。lint、生产构建、48 文件 / 38,925 项单元测试及 11 文件 / 78 项浏览器测试通过；两项仅覆盖已删除编辑模式的单元测试随功能移除。生产构建双端 E2E 为 43 PASS、2 NOT RUN（bfcache 未触发），验证消息操作入口消失、复制/重试/历史/删除及身份隔离；原编辑分支 E2E 改为按钮不存在检查，分支回执和持久化的 store 测试保留。后端、模型配置和数据库结构未修改。

## 2026-10-17 分类主题发现与历史使用情况

“餐饮”字面匹配可能只找到从未使用的旧分类，不能据此推断“食品”等实际记账分类没有支出。`list_categories` 新增字符串计数 `direct_transaction_count` 和 `subtree_transaction_count`：前者统计直接归类记录，后者包含所有后代分类。两者覆盖全部有效历史、隐藏记录与所有币种；退款计一条记录，逻辑转账只计一次，删除记录和其他用户排除。父分类直接计数为零不代表整个子树未使用，金额相抵为零也不代表没有记录。这些计数不是用户所选月份的支出金额。

计数复用同一 UID 只读快照中的分批逻辑交易扫描，再统一计算分类树汇总，不执行逐分类查询，不新增同义词表、不改变账本。扫描工作量随有效交易量增长，内存按已有扫描批次和分类树约束；每次分类工具读取保持当前快照一致性，受现有查询超时和取消约束。损坏转账或缺失关系明确失败，不伪装成分类未使用。此实现没有增加性能预算或声称百万级账本性能已验证。

保留在实际模型输入中的工具描述及系统提示明确：keyword 只是名称字面匹配，不是主题分类。广义主题遇到未使用的名称匹配或零匹配交易时，先进行一次保持日期、币种、类型、账户、标签、金额等限制的 category 聚合发现，只移除暂定主题的分类限制；看到实际分类后说明统计口径，再聚合或澄清真正的歧义。不得将主题替换为交易备注 keyword，不得无限重试。显式分类 ID 或用户确认的“两类合计”必须保持原范围，真实为零就报告该范围的零，不静默并入食品等其他分类。

新增回归使用合成 SQLite 与确定性 HTTP 协议：旧分类字面匹配→零结果→同范围分类发现→明确食品/饮料/水果范围聚合得到 CNY 45.00；另测显式两类只返回原范围零，以及同范围确实没有交易的零。逐轮检查账户、标签、币种、金额和时间限制不变，分类使用计数经过 `column_rows` 压缩仍完整。元数据回归覆盖隐藏有/无流水、父子孙层级、退款抵消、多币种、同名不同 ID、删除/UID 隔离、分页、取消和损坏关系。增加 100 个空分类只增加 100 个元数据读取行，交易扫描不会按分类数重复。

`BUILD_PIPELINE=1 go test ./...`、`go vet ./...`、新增分类定向回归及分类/元数据/多轮相关 race 检查通过。以上验证工具数据与协议约束，不等于真实模型的语义选类认证；未调用真实模型或修改生产数据。

## 2026-10-17 正常发送提示闪烁

正常发送曾把持久化 pending 当作异常恢复展示，顶部短暂出现绿色“正在确认更改”；新建会话收到回执后的读取又会显示绿色 Loading。现在持久化之前记录 sending，恢复面板只展示真正待确认、未送达或重载后未知的请求，正常发送由输入区反馈。后台核对跳过当前在途请求，持久化/拒绝异常的清理和身份代际校验确保恢复入口不会永久隐藏，也不会被旧请求清除新重试状态。正常发送时的新会话加载不再显示顶部提示，普通历史加载保留。

生产构建结合隔离本地 API，在桌面回车与手机发送按钮分别检查新建会话和后续提问，MutationObserver 捕获发送全过程的瞬态提示：修复前两端均出现 Confirming changes 和 Loading；修复后两端均为零，运行时错误为零。自动测试覆盖持久化等待、网络等待、重复发送、实际失败、刷新恢复、拒绝状态保存异常，以及重置后同请求 ID 重试的迟到回执。前端 51 文件 / 38,954 单元测试、12 文件 / 81 浏览器测试通过；没有调用真实模型。

## 2026-10-17 摘要局部超时与前台请求恢复

压缩阶段有独立摘要期限。此前该子期限或供应商局部 `DeadlineExceeded` 被直接上传，又被 worker 当成整个 run 超时；即使原完整历史仍在模型预算内，也会反复停在 preparing。修复将父 run 的取消/期限与摘要子请求失败分开：父 context 优先终止；局部摘要失败只在全部原历史通过实际主/备供应商输入预算检查时，保留原文并继续同一 run，失败调用仍计入原模型轮数。失败摘要不替换会话摘要、覆盖范围或来源依赖。

原历史确实超预算时，最多允许一次缩小完整历史前缀的压缩恢复；只处理更少的完整 turn，不裁切或删除其余约束。恢复仍使用原 run deadline、摘要子期限和调用轮数，并给前台回答预留轮数。第一次调用前以及失败后的恢复/再次调用之前，重新检查 owner/lease、cancel、conversation version、active run 和来源依赖；成功摘要仍需验证结构、大小并通过 CAS 提交。没有足够可缩小前缀或预算时明确终止，不无限重试。

worker 仅在自己的 run deadline 真正耗尽时标记 `run_timeout`；查询的局部超时通过既有可重试 `execution_limit` 返回，内部保留 `ToolErrorQueryTimeout` 分类，供应商局部失败使用 `provider_unavailable`。下一页授权检查优先保留父取消/超时，局部查询超时不再伪装成 cursor 失效；正常路径保留精确签名 cursor 与 Scope，显式重新查询仍需重新授权。

新增 allowlist 观测记录 compaction 请求开始/结束、阶段、耗时、实际模型调用数、历史 turn 数、失败来源和恢复分支，不记录消息内容、财务金额、工具参数或原始供应商错误。`request_finished` 仅表示请求结束；摘要生效由无失败的 `commit` 阶段表示，无效 JSON、超大摘要或 CAS 拒绝不能作为成功摘要。

参考固定版本 Codex [compact.rs](https://github.com/openai/codex/blob/fa46ef3af641f21d0facc69bd11e562fce6a3f9a/codex-rs/core/src/compact.rs)、[session_async_compaction.rs](https://github.com/openai/codex/blob/fa46ef3af641f21d0facc69bd11e562fce6a3f9a/codex-rs/core/src/session/async_compaction.rs)、[session/turn.rs](https://github.com/openai/codex/blob/fa46ef3af641f21d0facc69bd11e562fce6a3f9a/codex-rs/core/src/session/turn.rs)。借鉴的是分层失败、取消优先、投机失败不损害已有上下文及有界恢复原则；本项目没有实现异步压缩，也没有采用删除最旧财务约束的策略。Codex 必需的前置压缩失败仍可能终止 turn，不能概括为所有压缩失败均继续。

新增 Go 回归覆盖真实 1 秒摘要子 deadline、供应商局部 Deadline、原完整历史保留、同 run 真实 SQLite 工具查询和逐段正文输出、父取消/期限、硬预算前缀恢复上限、失败期间版本/取消/删除授权变化、worker 超时分类以及 next-page 精确 cursor/scope。

双端生产构建 E2E 使用 native 后端、隔离 SQLite 与本地 HTTP 模型协议：桌面/手机各强制 1 秒摘要子 deadline 耗尽，当前 run 自动恢复原完整 history，经真实 query_transactions 后，在上游结束前可见 partial stream，终态 completed，且无 pageerror。结果位于本地 `~/Library/Caches/ebk-timeout-release/timeout-smoke/results.json`。这不是商业模型质量认证，未调用真实供应商或重放用户账本。

本轮后端定向测试（summary、Prepare、compaction、历史、超时层级、授权变化、continuation 及 models contract）通过。为兼容旧缓存前端的严格白名单，没有新增 reason 或 HTTP 错误索引；局部查询超时使用旧版本已认识的 execution_limit，新的中性文案说明执行限制并保留同会话重试。内部保留 ToolErrorQueryTimeout 分类；准备阶段失败维持 failed，运行器返回 incomplete 时仍保留 incomplete，不为 execution_limit 引入新的终态语义。

最终两位独立 agent 串行复审均无阻断项。`BUILD_PIPELINE=1 go test ./...`、`go vet ./...`、相关 race、lint 通过；前端 51 文件 / 38,954 单元测试、12 文件 / 82 浏览器测试通过，文案兼容调整后的定向测试亦通过。CI 模式按现有规则跳过实时第三方汇率；一次非 CI 全套运行中仅马来西亚央行接口返回 EOF，该单项随后重新执行通过，未将跳过项算作外部集成通过。
