# 04 · AI integration 串行审查

2026-10-16；基线 `91081998` 相对 `origin/main`，包含 gate01–03 工作区修复。

**最终闭环：本地适用 AI integration 门 10/10，通过。** Gate15–18 的实际执行已补齐本门原定的数据库、性能、迁移、双端与质量条件；见文末最终核验。下文8/9/9.5分及SKIP保留为历史，不改写成当时已通过。真实商业模型等外部发布验收仍为 NOT RUN，本结论不等于计划全部发布验收通过。

**初始 8/10；修复后 9/10。** 已确认的清账并发缺陷已修复且有红绿回归；本门代码审查范围未发现其他已证实未修缺陷。完整质量矩阵、三库并发、双端真实交互、真实主备模型与大规模账户清账性能仍未验证，不能据局部测试给整个 AI integration 门 10/10。评分采用本审查约定，技能本身未定义量表。

未派子 agent、commit/push、运行 Docker/全仓矩阵；未读私人配置、凭据或用户数据库。新增数据库测试使用临时 SQLite。

## 依据与范围

读取总 manifest、01/02/03 报告、`skills/ezbookkeeping-ai-integration/SKILL.md` 与必读 `references/ai-feature-map.md`，并读取 `.codex` 同技能及 project-spec/global-prompt；结合 `docs/ai-providers.md` 的隔离测试约束与 harness 验收记录引用。审查对象是当前持久化体验，同时检查其复用的 legacy runner/tool/provider 边界；收据识别只核对配置隔离，未改识图实现。

明确逐段阅读了 queries、cursor、transactions、aggregate/results、balances、read_helpers_scan、datastore read-view，runner、OpenAI tool provider，settings AI/load，experience runtime，storage transaction/commands/dependencies/operations，send/retry，cleanup，worker claim/execute/checkpoint，compaction 与 background memory 提交。对辅助 read_helpers 使用 UID/SQL 调用检索并结合现有隔离回归。没有把路径扫描称为逐行审完 PR 所有源文件；Vue/Pinia/UI 和全协议组合留给后续专门门及最终矩阵。

| 技能约束 | 代码/运行证据与结论 |
| --- | --- |
| 固定七工具、UID 只读 | `AssistantQueryService` 在 trusted constructor 固定 UID/shard，registry 校验七个 handler；tool 参数不控制 UID。`read_helpers_transactions` 和 metadata SQL 限定 UID/deleted；`WithAssistantReadView` 使用 primary 固定连接、SQLite query_only 或只读事务并恢复/丢弃连接。七工具、跨 UID、隐藏/删除关系、拒绝写工具回归通过。 |
| 完整匹配范围 | `ScanTransactions` 用 transaction_id storage batch 读到尽头；detail page heap 仅限制投影内存，不把 detail page 当总额；aggregate 扫描后再分页分组，余额先按全体 leaf 累计再分页。2500 条与 full-history 回归通过；未在本轮重跑 100K 三库计划。 |
| 精确金额、transfer、证据/错误 | 金额以整数字符串投影，aggregate/balance 使用 big.Int，按币种/flow分组；transfer 归一为 logical outgoing 并校验另一侧。runner unresolved tool error 阻止模型把失败说成成功空账本。overflow、最小负数、transfer、持久化 evidence 和 error correction 回归通过。 |
| 一个 runner、取消和预算 | 持久化 `Prepare` 创建同一个 `AssistantRunner`；legacy JSON/SSE 也共用该 runner。worker 应用级运行，HTTP SSE 只读快照；checkpoint/final transaction 重验 owner/lease/依赖，取消传播上游。runner fallback 仅一次且 visible output 后禁止；本轮 provider cancellation、runner sink/tool cancellation、late ledger-clear completion 回归通过。 |
| 显式 continuation | Send next_page 校验当前/branch source；signed cursor 绑定 UID/tool/timezone/hierarchy/expiry 并重验。Prepare 使用 run-local saved-page handle，不把新提问自动当下一页；cursor 与跨请求 continuation 回归通过。 |
| provider/config 边界 | model/URL 来自 LLMConfig/provider builder；transport 在 OpenAI adapter，用既有 httpclient、配置 proxy/TLS/timeout。Responses store=false、adapter-private state 校验 owner/prefix；预算包含 wire bytes。经验限制在 experience config，runner runtime在 primary段，capabilities 输出有效限制。三协议 fixture、private-state isolation、配置验证通过。真实供应商行为仍未验证。 |
| retired embedding/cache | load 不解析 retired 段和 retired secret-file overrides；生产工具/runner无 embedding 访问。legacy 模型保留；cleanup retention 测试保留其记录。配置 FIFO secret-file不打开、production container拒绝cache访问、cleanup retention 回归通过。前端 service-worker cache 属后续 Vite/UI验证，不冒充本轮 Go 证据。 |
| 持久化与失效 | mutation/receipt/epoch/CAS 在 service；compaction、memory background、worker final 都在同一 user-state CAS 下重验。发现 early invalidate 与 ledger delete 分离窗口，本轮修复如下。Gate02 的 memory projection 与超4任务撤销修复保留。 |
| 双端一致性与文档 | 两端复用持久化 runtime/接口；本轮没有前端变更，不能凭共用代码宣称手机/桌面实际操作通过。修订两套 AI skill 的配置/接口地图事实错误，迁移文档同步删除事务语义。 |

## 已修：清账提交未覆盖窗口内新任务

原 `ClearAllTransactions`/`ByAccount` 先 `InvalidateLedgerContext`，然后另一事务删除账本。新 Send 可在两步之间被受理、claim 并查询旧账本；删除之后没有再次失效。不是推测：barrier 测试在 early fence 后停住清账，受理/claim 一个新任务并让模型进入，再放行 ledger delete。旧行为删除提交后任务仍 `running`，预期为 `cancel_requested`，红灯日志 `/tmp/gate04-ledger-red.log`。

修复边界：

- `ai_assistant_cleanup.go` 提取同一 `AssistantStorageTx.invalidateLedgerContext`；保留原 early fence，并在真正删账事务内再次调用。
- `transactions_clear_context.go` 让清账 mutation 与 fence 使用同一个 user-shard session/CAS；不在 assistant callback 内嵌套另一 ledger transaction。原业务错误原样返回，rollback 同时撤销本次 ledger/fence 变更。
- `transactions_delete.go` 的两个 bulk 方法显式接收 storage，三个 API 调用点全部传入。普通 `DeleteTransaction` 仍走原事务路径；账户清账内部逐条使用 fenced path，保留原逐条提交和部分失败行为。没有 schema/new pending lock/自动重放删除。
- 任务在删账提交前读取仍存在的数据是允许的；提交时取消窗口内运行、递增 source/context 版本和 memory-generation epoch。最终答案、压缩摘要与记忆候选的后续提交分别被 cancel/version/epoch 拦截。崩溃沿用数据库 rollback/commit 和 worker lease 恢复，无新增长期锁恢复流程。

新增 `ai_assistant_ledger_clear_test.go`（224 行）覆盖：

1. 全交易清空、删账户的全数据 ledger 步骤、按账户清空三场景：early-fence/send/claim/model/delete barrier，删除后旧 checkpoint 失败、不能重新 claim、晚到正文不保存、最终 canceled；删除后新 run 可 claim。
2. mutation 失败 rollback，原 run/queue/memory epoch 保留。
3. 账户第一笔成功、第二笔因缺少转账目标账户失败：第一笔删除/fence保留，第二笔 rollback，业务错误保留；普通单条删除未扩大失效行为。
4. 同一 turn regenerate 完成后再次清账仍递增 source/context 版本，并撤销再次清账前新受理任务；避免错误的“已失效且turn未变便跳过”优化。

barrier 使用 once-safe cleanup 释放和有界等待。`all-data` 子例验证共享 ledger 删除服务的 deleteAccount=true 分支，**不是完整 ClearAllData HTTP 路径**；实际 API epoch/receipt/replay由本轮同时运行的两个既有 ClearAllData 回归验证。全数据清空仍跨多个服务分步执行，不声称整体原子。

## 定向验证

- `go test -json ./pkg/services ./pkg/api ./pkg/settings ./pkg/llm ./pkg/llm/provider/openai -run '<本门明确选择的 ledger/registry/UID/amount/worker/context/memory/cache/provider/config/API 用例>' -count=1`：日志 `/tmp/gate04-focused.json`；**128 个实际叶子场景 PASS、2 SKIP、0 FAIL**。skip 是 background memory 的 MySQL/Postgres，无服务配置；不能计为兼容通过。此选择在 llm 根包无匹配测试，后续单独补跑 runner，未把空包结果计数。
- `go test -json ./pkg/llm -run 'TestAssistantRunner(RequestScopedFallback|CanonicalCacheAndDuplicateIDs|CorrectsErrorsWithoutEmptySuccess|CancellationDuringTool|CancellationAndSinkFailure|ProviderDeadlineFallback|BudgetsRetainEvidence)$' -count=1`：**32 个实际叶子场景 PASS、0 SKIP、0 FAIL**，日志 `/tmp/gate04-runner.json`。
- `go test -race ./pkg/services -run 'TestAssistantLedger' -count=1`：PASS，4 顶层测试、6 实际叶子场景；日志 `/tmp/gate04-ledger-race.log`。
- 红灯有效性：临时让事务 helper 走原 ledger-only 路径，barrier 测试 FAIL（实际running）；finally恢复修复。首次测试断言误把异步取消立即视为 canceled，修正为提交时 cancel_requested、worker结束后 canceled，并真实执行 late-model completion。
- gofmt / `git diff --check`：PASS。`npm run check:source-lines`：1667 source/test/style files，0超过500行。没有压缩代码凑门禁。

以上160个不同的定向叶子场景不是全仓测试数；race 是重复执行新增场景，不叠加成独立覆盖数。

## 剩余项与交接

- 清账并发回归实际通过临时 SQLite；重开门已把原六场景及新增场景全部参数化三库。最终数据库 agent 需要真正执行 MySQL/Postgres，不能仅依赖父测试名称 PASS 或 skip。
- 初轮按账户清账存在 O(N×C) 开销；父审实测后已在下方重开门中修复。没有采用仅按 SourceInvalidated/turn/active-run 过滤；高并发生产规模负载仍待最终阶段测量。
- 全仓 Go test/vet/race、前端 lint/test、Docker 重启恢复/三库迁移保留、双端 E2E 依用户串行安排后续执行。真实主备模型配置当前缺失，未读取私人配置补足；专用隔离真实供应商 smoke仍未执行，mock 不替代它。
- 后续若改写清账事务或 worker/memory/compaction提交边界，需重跑本门 barrier/rollback/regeneration测试；若优化账户清账性能，也需保留部分失败与取消契约。


## Gate14 后重新打开：账户清账扫描放大已修复

父审用真实 `DeleteAllTransactionsOfAccount` 测得25账单/0会话13.33ms、25账单/1000会话2962.78ms，确认原门记录的开销已经形成实际性能回退。本轮仍只做串行 focused 修复；未运行 Docker/全仓矩阵、未委派/commit/push/改主manifest。

重新阅读本报告、AI integration（含 feature-map）、project-spec/global-prompt 和 software-design-philosophy 的信息隐藏/深模块/战略设计要求。选择操作私有 guard，避免让交易调用方管理版本算术、是否重扫或 commit 后更新证明；设计范围评分 **9.5/10**，整个 AI 门仍 **9/10**，实际外部数据库和最终质量验证未完成。

### 设计与正确性边界

- 新 `ai_assistant_storage_ledger.go` 的 `assistantLedgerClear` 拥有一个用户、一个串行 clear operation 的 storage/proof；账户删除循环只创建一次。每次 `apply` 仍把实际删除置于同一个用户分库 CAS 事务，不合并多个账单为一个大事务。
- `ai_assistant_storage.go` 内私有 `withUserTransactionVersion` 返回**成功 Commit 后**的权威版本。获取写位置时的 +1、提交前 +1 及锁前版本恢复都封装在 storage；公开 `WithUserTransaction` 契约不变，其他消费者无需管理版本。
- guard 在锁已经取得后比较上次成功版本与锁前版本。相同才复用上次失效；首次、任意助手提交或版本不符都重新失效。业务 callback、最终 CAS 失败、Commit 失败均不会发布 proof；失败无论是否重试都保持原 proof。若提交结果不确定，保守不复用，不自动重放其他外部操作。
- Send/retry/regenerate、worker checkpoint/final、compaction和memory提交都经过该用户状态事务。只读查询本身不能创建持久上下文。相同turn再生成即使最终无active run也推进版本，因此不会被“turn未变化”错误跳过。
- 复杂度变为无助手写入交错时一次会话扫描加N个原有删除事务；有交错时每个观察到版本变化的边界重新扫描。最坏持续写入仍可能回到O(N×C)，这是保守正确性边界，不宣称并发负载恒定耗时。无新schema、持久锁、全局缓存或客户端变化。

### 新回归与红灯证明

原 `ai_assistant_ledger_clear_test.go` 六场景全部改为 SQLite/MySQL/Postgres 参数化。新增 `ai_assistant_ledger_clear_reuse_test.go` 每库六个叶子场景：1/25账单实际SQL扫描计数、新run插入、同turn完成再生成、部分失败保留先前提交、rollback及callback返回之后的最终CAS失败重试。数据库hook只计数，不开启敏感SQL日志；首次尝试用普通日志统计得到0，因为助手session正确禁用SQL日志，改用hook后才作为有效证据。

`ai_assistant_ledger_clear_performance_test.go` 将父probe转为可重复三库回归，每库0/1000会话、25账单，运行真实账户清账，断言扫描页数分别1/21并验证账单已删除；只记录耗时，不以脆弱的机器速度阈值断言。

受控临时 mutation（全部 finally 恢复，日志留 `/tmp`）：

- `/tmp/gate04-reuse-red-scan.log`：恢复逐笔失效，25账单扫描50次，预期2，FAIL。
- `/tmp/gate04-reuse-red-stale.log`：只复用第一次失效、忽略版本变化，新run仍queued，同turn再生成的sourceVersion未增长，两项FAIL。
- `/tmp/gate04-reuse-red-proof.log`：callback内提前发布proof，失败后proof从2变4，FAIL。
- 正式最终CAS重试用例在callback中变更数据库version，迫使callback返回nil之后的CAS失败，第二次重新执行；每次都断言已取消交错任务、proof尚未变化，最终只有一次epoch增量和成功删除。不是只测试callback主动返回retry错误。

### 实测比较与最终验证

同一隔离fixture、同一机器顺序复测（临时父probe文件已删除，正式scaling测试保留）：

| 25笔账单 | 每笔失效基线 | 操作内版本证明 | 说明 |
| --- | --- | --- | --- |
| 0会话 | 15.8025ms | 13.6983ms | 常规小负载无额外明显开销 |
| 1000会话 | 2705.3534ms | 119.4954ms | 约22.6倍，耗时减少95.6% |

原始对比日志 `/tmp/gate04-reuse-probe-before.log`、`/tmp/gate04-reuse-probe-after.log`。正式scaling再次运行：0会话13.7150ms/1页，1000会话127.7899ms/21页。以上是本机临时SQLite样本，不是生产SLO或三库性能认证。

- `BUILD_PIPELINE=1 go test -json ./pkg/services -run '^TestAssistant(Ledger|StorageDatabaseContracts|StorageExternalOperationAcceptance)' -count=1`：**18叶子PASS、36外部数据库叶子SKIP、0FAIL**；其中ledger 14 PASS/28 SKIP，storage 4 PASS/8 SKIP。日志 `/tmp/gate04-reuse-final.json`。SKIP不计通过。
- 同一选择加 `-race`：**18叶子PASS、36SKIP、0FAIL**，日志 `/tmp/gate04-reuse-race.json`；没有race报告，不把重复执行再算新增覆盖。
- 父独立复核ledger组：14 SQLite叶子PASS、28外部数据库SKIP；25×0=12.29ms/1页，25×1000=128.67ms/21页，日志 `/tmp/ebk-parent-gate04-reuse.log`（父报告证据，非本子agent重复执行）。
- gofmt、`git diff --check` PASS；`npm run check:source-lines`：1704文件、0超过500行。最终本轮Go文件分别191/50/16/265/240/181/52行，无压缩凑门禁。

本次重开修改/新增：`ai_assistant_storage.go`、`ai_assistant_storage_ledger.go`、`transactions_clear_context.go`、`transactions_delete.go`、原ledger_clear_test及新增reuse/performance_test，迁移说明与本报告。普通single-delete事务路径、账户部分提交/业务错误、公开assistant transaction API保持；没有改主manifest。实际三库执行、全仓质量矩阵、Docker/E2E、真实主备模型仍待对应串行阶段。此门冻结交回父审，后续设计门复核guard边界即可。


## Gate15–18 后最终闭环

2026-10-16。此次只核验现有实现与已执行证据、更新本报告；未修改代码或其它报告，未重复无变化测试，未委派或提交。

重新检查当前 `assistantLedgerClear.apply`、storage 的私有版本事务入口与账户循环：guard仍仅属于单用户单次串行操作；锁前版本比较发生在取得用户写位置后；proof仍仅在成功Commit返回后更新；普通单删保持nil guard原路径。未发现新的具体缺陷，前述并发失效和性能整改契约保持。

### 原条件的实际关闭证据

| 原待验证条件 | 最终核验结果 |
| --- | --- |
| ledger guard三库、回滚/重试/交错与扫描开销 | 独立读取并计数 [最终services叶子清单](../ai-assistant-experience-evidence/audit-matrix/go-services-final-assistant-leaves.json) 与 [race叶子清单](../ai-assistant-experience-evidence/audit-matrix/go-race-assistant-leaves.json)：各自 **14×3=42 ledger PASS、0 SKIP**；覆盖原barrier、部分失败、同turn再生成、最终CAS重试和scaling。25账单/1000会话三库都断言21扫描页。历史外部SKIP由这些后续实际执行关闭，未被改写为PASS。 |
| Gate02 memory投影及超默认并发撤销 | 同两份清单独立计数：normal/race各 **3×3=9 PASS、0 SKIP**，保留统一投影和超过4任务的撤销契约。 |
| 完整匹配查询、规模和持久化容量 | [Gate17](17-final-test-matrix.md) 与其 [summary](../ai-assistant-experience-evidence/audit-matrix/summary.json)：100k ledger、约97.78MiB逻辑历史在三库实际执行，6性能叶子全PASS；全范围金额/coverage、双计数/字节分页和导出断言通过。另实际大HTTP导出1 PASS。性能测量非race数据，也非生产容量承诺。 |
| schema/cache retention与恢复 | [三库迁移结果](../ai-assistant-experience-evidence/audit-matrix/migration.json) 全PASS：旧版初始化、保留ledger/embedding的两次升级、原生备份恢复、旧版回滚/再升级及过期lease恢复；不再仅凭fresh schema推断旧记录保留。 |
| 统一运行生命周期、真实HTTP取消及双端 | [Gate15](15-docker-smoke.md) 的启动/持久化smoke由 [Gate16](16-container-e2e.md) 补足两端完整流程、引用、兼容与proxy。上游取消 [原始结果](../ai-assistant-experience-evidence/audit-e2e/upstream-cancel/smoke-results.json) 明确 `upstreamResponseCompleted=false`、554ms内观察到未完成HTTP响应关闭；这是实际网络取消，但供应商仍是确定性fixture。完整业务主组47 PASS/2 NOT RUN，引用10 PASS，两个NOT RUN是实际bfcache恢复，未计通过。 |
| 技能质量命令及最终产物 | Gate17实际全仓Go/vet、相关包race、lint、browser/build通过；全仓初始opt-in skip中的性能与HTTP导出已另行执行，商业deployment仍NOT RUN。 [Gate18](18-harness-finalization.md) 及其 [summary](../ai-assistant-experience-evidence/audit-harness-finalization/summary.json) 记录最终父独立unit **44文件/38,917 PASS**、当前harness重建与10场景smoke；dist/Linux executable与Gate16逐字节相同，支持保留双端证据。未将测试源码哈希差异说成完全源树相同，也未将normal/race重复执行累加为独立覆盖。 |

**本地适用门最终10/10。** 固定UID只读工具、完整精确查询、统一runner/cancellation/continuation、provider/config/cache边界，以及持久化失效/恢复，本门已审要求具有相应代码、回归和实际隔离环境证据；此前扣分的本地条件已关闭。评分是有范围的审查结论，不保证不存在所有未知缺陷。

### 仍未执行的外部验收及评分范围

AI integration技能明确写的是 **“Real-model smoke is opt-in with a dedicated isolated deployment/test user”**。未提供专用真实商业主备模型配置时，不擅读私人凭据、不执行该opt-in smoke；因此它不作为本地适用技能门的假失败，也绝不登记为PASS。

另一方面，[体验验收记录](../ai-assistant-experience-verification.md) 的真实模型验收门槛要求真实模型与独立账本oracle、主备服务和成本/延迟证据，并要求供应商不可用时保留发布阻断。真实财务/长对话压缩/记忆质量及主备商业模型测试仍 **NOT RUN，发布验收阻断保留**；本地10/10不能代替这些要求。真实fixture HTTP关闭证据也不能证明商业供应商的取消、质量或计费行为。

实际bfcache恢复、物理手机键盘、人工screen reader、其它浏览器/amd64、stock Dockerfile及Node26阶段仍按Gate16–18标为NOT RUN。当前验证镜像是已记录的Go/Alpine验证环境，不冒充stock发布镜像。它们属于尚未关闭的跨门/发布验收边界，不由本次仅文档闭环悄然豁免。
