# AI assistant harness 代码影响分析

任务：[draft 计划](../tasks/in-progress/ai-assistant-harness.feature.md)。依据根目录 `AI_ASSISTANT_HARNESS_PLAN.md` v1.1（完整阅读，重点第 15 节）；2026-10-14 实地代码核查。本文仅为计划，不表示实现或测试已经完成。

## 边界与仓库约束

只读、绑定当前 UID、全部有效历史、隐藏实体纳入、无 embedding 运行依赖；保留旧表和 schema 注册。没有新增数据库、长期聊天存储、MCP 写权限、agent 框架或生产发布。原方案中执行授权示例不构成本轮实施授权；本轮只写规划，任务留在 draft。

检查仓库及 `/Users` 至工作目录祖先，未发现适用 AGENTS.md。已读 `.codex/skills/ezbookkeeping-{project-structure,ai-integration,project-spec}/SKILL.md` 及其 references。AI 技能中的稳定 top-k 要求与本次方向冲突，以用户及方案为准；实施阶段同步 `skills/` 副本。现有 `skills-lock.json` 修改及未跟踪 `pnpm-lock.yaml`、plan-task 技能、原方案不属于本任务可覆盖内容。

## 现有接口与调用链（真实符号）

| 路径 | 已核查接口/行为 | 影响 |
|---|---|---|
| `pkg/api/large_language_models_assistant.go` | `(*LargeLanguageModelsApi).AssistantChatHandler(c *core.WebContext) (any, *errs.Error)`；`AssistantChatStreamHandler(c *core.WebContext) *errs.Error`；`prepareAIAssistantPromptContext(c *core.WebContext, currentConfig *settings.Config) (*aiAssistantPreparedPromptContext, *errs.Error)` | 两路共用新 runner，保留认证、语言、模式；删除旧准备流程 |
| 同上 | `getAIAssistantTransactionsForKnowledge(c *core.WebContext, uid int64, clientTimezone *time.Location) ([]*models.Transaction, error)`；`getAIAssistantKnowledgeCoverageStartUnixTime(clientTimezone *time.Location) int64`；`selectTopAIAssistantKnowledgeItems(queryEmbedding []float64, knowledgeItems []*aiAssistantKnowledgeItem, topK int) []*aiAssistantRetrievedKnowledgeItem`；`buildAIAssistantFinancialSnapshot(knowledgeItems []*aiAssistantKnowledgeItem, clientTimezone *time.Location) string` | 固定窗口、向量和受限快照都必须移除，不能留下另一条回答分支 |
| `pkg/services/transactions.go` | `GetTransactionByTransactionId(c core.Context, uid int64, transactionId int64) (*models.Transaction, error)`；`GetTransactionsByTransactionIds(c core.Context, uid int64, transactionIds []int64) ([]*models.Transaction, error)` | UID 查询规则可复用；现有 helper 自建 session，不能直接拼出跨语句一致快照 |
| 同上 | `GetAccountsTotalIncomeAndExpense(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, excludeAccountIds []int64, excludeCategoryIds []int64, tagFilters []*models.TransactionTagFilter, clientTimezone *time.Location, useTransactionTimezone bool) (map[int64]*big.Int, map[int64]*big.Int, error)` | 使用 big.Int，但先取 allTransactions；不适合作为新全历史内存受控聚合的原样实现 |
| 同上 | `buildTransactionQueryCondition`、`GetTransactionsByMaxTime`、`GetTransactionCount`、`appendFilterTagIdsConditionToQuery` | 旧边界为 `<= max`、`>= min` 且仅正数启用；有按账户选择转账记录的旧口径。新半开时间、负时间、规范逻辑 ID、同侧金额条件必须集中实现，不能盲目接旧参数 |
| `pkg/services/accounts.go` | `GetAllAccountsByUid(c core.Context, uid int64) ([]*models.Account, error)`；`GetAccountAndSubAccountsByAccountId(c core.Context, uid int64, accountId int64) ([]*models.Account, error)` | 查询有 UID/deleted 且不过滤 hidden，适合规则来源；新元数据分页需 ID 次序而非现 display_order |
| `pkg/datastore/database.go` | `NewSession(c core.Context) *xorm.Session`；`DoTransaction(c core.Context, fn func(sess *xorm.Session) error) (err error)` | 当前没有显式只读、REPEATABLE READ 或主库固定保证，新增助手专用一致读入口，不改变全部业务事务默认 |
| `pkg/datastore/query_context.go` | `NewXOrmContextAdapter(c core.Context) *XOrmContextAdapter`；Deadline/Done/Err 全空，Value 只传 SessionID | 确认取消断链；必须委托父 context 并保留日志值 |
| `pkg/core/context.go` / `context_web.go` | Context 嵌入 context.Context；WebContext 嵌入 Gin，有禁止加字段注释 | 新小适配器组合 Request.Context 与已捕获元数据；不将 Gin writer 留给异步任务 |
| `pkg/llm/provider/large_language_model_provider.go` | `GetJsonResponse(c core.Context, uid int64, request *data.LargeLanguageModelRequest) (*data.LargeLanguageModelTextualResponse, error)`；独立 `StreamTextResponse(c core.Context, uid int64, request *data.LargeLanguageModelRequest, callback data.LargeLanguageModelStreamCallback) (*data.LargeLanguageModelStreamResponse, error)` | 追加可选工具能力接口，避免要求所有识图 provider 实现工具 |
| `pkg/llm/provider/openai/openai_common_compatible_large_language_model_adapter.go` | Chat Completions JSON/强制流式解析 | 必须增加 call ID/index 分片和有效结束判定，不能只提取文字 |
| `pkg/llm/provider/openai/openai_responses_api_large_language_model_provider.go` | `(*openAILargeLanguageModelProvider).StreamTextResponse(...)`；`processOpenAIResponsesStreamEvent` | Responses 必须携带工具结果和必要 output/reasoning items；保持 store=false |
| `pkg/llm/large_language_model_provider_container.go` | `InitializeLargeLanguageModelProvider(config *settings.Config) error`；fallback provider 的 GetJsonResponse/StreamTextResponse | 当前 fallback 是单次请求包装；runner 必须拥有整个用户请求最多一次切换状态 |
| `src/lib/services.ts` | `chatWithAIAssistant`；流内 `parseRawStreamEvent(rawEvent: string, fallbackToDoneEvent: boolean)` | 已确认解析失败尾帧返回 `{type:'done'}`；移除、跟踪唯一终态、EOF 无终态失败 |
| `src/views/base/assistant/AssistantPageBase.ts` | `useAssistantPageBase()`；`clearConversation()`；`cancelCurrentRequest()`；`getHistoryPayload()` | 现 clear 只清消息；history 直接截最近 12。需取消与请求 ID 隔离、状态过滤、保留错误 context/continuation |

## 修改、新增、删除清单

以下按职责列路径；新增文件为建议名，实际数量随最小合理分拆变化，不承诺一个巨型文件。

- **新增核心**：`pkg/services/ai_assistant_queries.go`（7 工具/强类型 DTO/schema/执行器）、`pkg/llm/assistant_runner.go`、`pkg/core/context_assistant.go`（建议）、服务层 query/filter/cursor/aggregate 按复杂度拆文件。签名细节由 P0 固定，不能将这里的建议误记为现存函数。
- **新增测试/证据**：对应服务、runner、core、datastore 测试；`pkg/services/testdata/` 下可手算夹具与规模生成器；SSE parser/composable 测试；`docs/ai-assistant-harness-verification.md`。跨数据库 opt-in 测试入口须实施时建立并记录准确命令。
- **修改后端**：上述表所有生产文件，另有 `pkg/api/large_language_models.go`、`pkg/llm/data/large_language_model_data.go`、`pkg/models/large_language_model.go`、`pkg/services/transaction_categories.go`、`pkg/services/transaction_tags.go`、`pkg/utils/api.go`、`pkg/errs/large_language_model.go`、`cmd/webserver.go`、`pkg/api/server_settings.go`。后两领域服务仅当需要 session-aware 复用时修改，避免无意义改名。
- **修改配置/运维**：`pkg/settings/setting.go`、`conf/ezbookkeeping.ini`、`cmd/initializer.go`、`templates/prompt/personal_finance_assistant.tmpl`、`docs/ai-providers.md`、`docs/ai-storage-migration.md`、两套项目技能及 references。
- **修改前端**：`src/models/large_language_model.ts`、`src/lib/services.ts`、`src/lib/server_settings.ts`、`src/views/base/assistant/AssistantPageBase.ts`、桌面及移动 `AssistantPage.vue`、`src/locales/`。沿用共享 Composition API，不新建 store。
- **修改现有测试**：`pkg/api/large_language_models_assistant_test.go`、`pkg/api/ai_deployment_integration_test.go`、`pkg/llm/large_language_model_provider_container_test.go`、OpenAI 两种协议测试及设置/初始化相关测试。保留识图等回归。
- **删除运行实现**：`pkg/services/ai_assistant_embeddings.go`（全部引用清零后）、assistant handler 内 embedding 请求/序列化/hash/top-k/快照/相似度生产路径及对应失效测试；移除配置 load/getter/初始化脱敏列表中的旧 embedding 字段。
- **明确保留**：`pkg/models/ai_assistant_embedding.go`、`cmd/database.go` 的 SyncStructs 注册及实际表数据；MCP 对外读写协议；既有普通交易 API 行为。

## 依赖与关键风险

依赖方向：API → runner/provider 和 UID 绑定查询服务；runner → provider-neutral `pkg/llm/data` + 执行器接口；services → data；禁止 llm → services。P0 契约后，查询与 provider/runner 可独立做 mock 验证；API/UI 接通需要二者完备。迁移清理必须在新路径能跑后进行。

| 风险 | 级别 | 缓解与证据 |
|---|---|---|
| 多语句计数/页/金额来自不同读视图，或读从库延迟 | 高 | datastore 固定主库连接和读事务；MySQL/PostgreSQL 显式一致隔离，SQLite 同连接；并发修改测试 |
| 通用 context 修改影响所有服务/定时任务 | 高 | 回归 deadline/Done/Err/Value/SessionID/nil，驱动取消和 session 释放，补 datastore 到 race 命令 |
| transfer 两条记录、隐藏账户、负支出、跨币种造成错误金额 | 高 | 独立六逻辑笔真值：CNY 收入 1000、支出 110、净额 890、转出 40、调整 200，USD 转入 5；不是两个共享实现互相比对 |
| schema 不拒绝重复 JSON 键/unknown，空 ID 扩大查询 | 高 | 标准 json.Unmarshal 不足；schema 与解码严格校验并统一归一化，关联表全 UID/deleted |
| 旧密钥缺文件仍触发配置读取 | 高 | 从 loadLLMConfiguration 调用源移除 embedding；测试 EBK/EBKCFP，不是只不使用返回值 |
| fallback 重放半轮工具，Responses 私有状态跨模型泄漏 | 高 | 整轮成功才执行，已完成中立上下文重建，请求级一次切换；正文/可展示思考后不切 |
| 上游/日志缓冲泄漏财务正文或失控内存 | 高 | `pkg/llm/provider/common/common_http_large_language_model_provider.go` 目前 io.ReadAll 与完整 debug response；新工具路径限流限字节并脱敏，检查 httpclient context 包装 |
| 游标错误复用/历史回填使“快照”承诺失真 | 高 | HMAC 派生 `Config.SecretKey` 用途隔离，24h/64KiB，绑定 UID/层级摘要/筛选/时区；高水位取实际最大次序，明确不是数据库版本 |
| 前端旧回调写入新会话/损坏尾帧成功 | 高 | 请求 ID/取消、唯一终态、CRLF/UTF8/EOF/重复终态测试，两平台同 composable |
| 三库金额 SUM 溢出/百万数据全驻内存 | 高 | 必要时分批 big.Int，完成才报 totals all_matching；执行计划驱动索引，不裁时间范围 |

现有 go.mod：xorm v1.4.1、invopop/jsonschema v0.14.0、mysql/pq/sqlite 驱动齐备；前端已有 Vitest、Vue、decimal.js。无需新 agent 框架。按已跟踪 package-lock.json、go.mod/go.sum 安装；不得据未跟踪 pnpm-lock 切包管理器。

## 可运行验证策略（本轮未执行功能测试）

已核对 package.json：lint=`vue-tsc --noEmit && eslint . --fix`，test=`vitest run`，build=Vite production。lint 会写文件，实施后必须复核 diff；本轮不运行它。

```sh
BUILD_PIPELINE=1 go test ./...
go vet ./...
go test -race ./pkg/datastore ./pkg/core ./pkg/api ./pkg/llm/... ./pkg/services ./pkg/settings ./cmd
npm run lint
npm run test
npm run build
git diff --check
```

聚焦阶段可运行 `BUILD_PIPELINE=1 go test ./pkg/services ./pkg/datastore ./pkg/core`、`BUILD_PIPELINE=1 go test ./pkg/llm/...`、`BUILD_PIPELINE=1 go test ./pkg/api ./pkg/settings ./cmd`；最终仍需全量。gofmt 仅指定实际变更 Go 文件，不全仓格式化。

既有真实部署测试 `TestAIDeploymentIntegration` 读取 `EBK_AI_DEPLOYMENT_CHECK=1`、`EBK_AI_DEPLOYMENT_WORK_DIR`、`EBK_AI_DEPLOYMENT_CONFIG`、`EBK_AI_DEPLOYMENT_UID`，默认关闭；目前仍要求 embedding，实施须先改写。改后命令为 `EBK_AI_DEPLOYMENT_CHECK=1 EBK_AI_DEPLOYMENT_WORK_DIR=/absolute/test-deployment EBK_AI_DEPLOYMENT_CONFIG=/absolute/test.ini EBK_AI_DEPLOYMENT_UID=<test_uid> go test ./pkg/api -run '^TestAIDeploymentIntegration$' -count=1 -v`。只能用于专用测试部署，不能把旧测试先指向生产验证。

三数据库新增契约集需有显式 opt-in、临时库名检查、schema 初始化/清理，固定时钟；本仓库尚无该助手 harness 测试入口，不能虚构其已可运行命令。实施交付文档必须补上确定入口/环境变量。2500 条跨 5 年真值、10 万条性能、真实模型问题矩阵、桌面/移动手测分别记录；缺环境只能标待验证，不能当作通过。本次没有修改实现，也未验证原方案旧基线所声称的测试结果。

## 增量集成与事务落地补充

P0 先新增 provider-neutral 类型、可选能力接口和 DTO 测试，保持旧助手可编译运行；P1 查询工具、P2 provider/runner 各自通过独立测试而旧路由仍可运行。P0 对 `pkg/llm/data/large_language_model_data.go`、`pkg/llm/provider/large_language_model_provider.go` 的共享接口冻结后，两支不得并发修改这些文件；接口变更集中单一集成步骤完成。查询支只编辑 services/datastore/core，协议支只编辑 provider/runner/container，API 接通在二者完成后执行。每一阶段保留可运行服务并有独立可验收里程碑，不以仅生成空壳作为里程碑。

研究阶段进一步核查 xorm v1.4.1：`Session.Begin` 内使用 `BeginTx(ctx, nil)`，没有公开 TxOptions 配置入口；先 `sess.Exec` 设置隔离再 Begin 不能直接证明同一物理连接，禁止依赖 MySQL 默认 RR 或默认连接池复用的假设。P1 datastore 步骤需选择并测试固定连接可控隔离的真实实现，必要时使用数据库内部适配/现有驱动能力，但不能以伪事务或全局修改隔离级别替代。必须覆盖 PostgreSQL 默认 READ COMMITTED 与 MySQL 被配置为 READ COMMITTED 的情形；取消应断开执行并释放连接。参见 [研究记录](research-ai-assistant-harness.md)。

需要同步去掉旧 embedding/top-k 约束的确切技能入口是 `.codex/skills/ezbookkeeping-ai-integration/SKILL.md` 与 `skills/ezbookkeeping-ai-integration/SKILL.md`，以及两目录各自的 `references/ai-feature-map.md`；项目结构两份 `references/project-structure.md` 中旧运行时 touchpoints 和 `references/dependencies.md` 中质量命令同步核查。当前 `npm run test` 确实存在，执行 `vitest run`，不是仅从技能推断。
