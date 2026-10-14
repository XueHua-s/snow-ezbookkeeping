# AI Assistant Harness 研究记录

日期：2026-10-14。关联任务：[draft 计划](../tasks/in-progress/ai-assistant-harness.feature.md)。来源：[原方案](../../AI_ASSISTANT_HARNESS_PLAN.md) v1.1。研究范围为文档与现有源码；没有实现、运行新功能测试、进入 todo 或 git 暂存。

## 结论与推荐

沿用现有 Go、xorm、HTTP provider 和 Vue 共享页面层，不引入 agent 框架、外部 MCP 跳转或独立数据存储。实现七个扁平只读工具，以服务端 UID、精确统计、读事务和请求预算约束模型执行。原方案第 15 节作为契约源，第 17 节只描述未来实施交接，不授权本次计划任务实施。

关键先决条件：先修复 context 传播并实现独立一致性读入口，再连接查询服务和 provider。xorm v1.4.1 的 `Session.Begin()` 实际调用 `DB().BeginTx(session.ctx, nil)`，不能在计划中假设已有可直接传 TxOptions 的 Begin 接口；需在 datastore 封装数据库适配并验证同一物理连接。PostgreSQL 可在事务开始后、首次业务查询前设置隔离级别；MySQL 的事务特性设置与连接绑定必须先用专用测试库证明，不能在 session 上先执行设置语句便假定与后续 Begin 同连接。必要时在 datastore 内使用现有 database/sql 连接事务能力，避免升级 xorm 或全局修改隔离级别。

## 已核查资源（16 组）

| 资源 | 源码证据/用途 | 可信边界 |
| --- | --- | --- |
| `AI_ASSISTANT_HARNESS_PLAN.md` 全文 | 7 工具、金额、游标、预算、兼容与验收冻结 | 产品与设计要求，不是运行证据 |
| `.agents/skills/plan-task/SKILL.md` Phase 2a/Judge 2a | 研究、可复用 skill、scratchpad 与评分维度 | 本次执行采用通用代理；缺少插件脚本与 judge prompt |
| `.codex/skills/ezbookkeeping-project-spec/` 及 global-prompt | 深模块、数据安全、可逆迁移 | 当前有效规则 |
| `.codex/skills/ezbookkeeping-project-structure/` 及两份 references | API/services/provider/UI 分层与现有工具链 | 版本以 manifest 为准 |
| `.codex/skills/ezbookkeeping-ai-integration/` 及 ai-feature-map | AI 配置与 prompt/provider/UI 入口 | 旧 top-k 要求与本任务冲突；实施 P4 更新两处副本 |
| `go.mod` | Go 1.27.1；xorm 1.4.1；jsonschema 0.14.0；mysql 1.10.1；pq 1.12.3；sqlite3 1.14.52；testify 1.12.1 | 当前声明版本，未在本次重新下载或验证 |
| `package.json` | Vue 3.5.43；Vitest 5.0.0；现有 lint/test/build | lint 包含 `eslint . --fix`，计划阶段不运行以免修改源码 |
| `pkg/datastore/query_context.go` | Deadline/Done/Err 均为空；Value 仅 request ID | 取消阻断已由源码确认 |
| `pkg/datastore/database.go` | NewSession 走 engineGroup；DoTransaction Begin 无显式隔离配置 | 不可直接证明只读、一致性快照或主库固定 |
| 本地模块 `xorm.io/xorm@v1.4.1/session_tx.go`、`engine_group.go` | BeginTx(ctx,nil)；EngineGroup.Master() 存在 | 库接口已核查，跨数据库实际行为待集成测试 |
| `pkg/api/large_language_models_assistant.go` | 1800 条、top 18、年份窗口；结构化回复类型；embedding 准备 | 替换旧路径，不保留 RAG 双轨 |
| `pkg/mcp/mcp_container.go` 与 query handler | 已使用 jsonschema Reflector，旧工具名称与分页约定 | 可借鉴 schema 生成；不可开放写工具或照搬旧 name-only 输出 |
| `pkg/utils/converter.go` | ParseAmount 最终整数乘 100 | 输入范围需在乘法前验证；累计用 math/big |
| `pkg/llm/provider/openai/` | Chat Completions/Responses adapter 与已有协议测试 | 独立可选工具接口；既有识图路径保留 |
| `src/lib/services.ts` | 解析失败 fallbackToDoneEvent 返回 done；固定 LLM timeout | 必须修复终态与客户端整体预算 |
| 原方案第 15.7 节 | 6 笔可手算真值与临时数据库要求 | 新实现测试预期，不冒充测试已通过 |

## 可复用模式与实现注意

1. 契约先行：一份领域 schema 定义同时服务 provider 和执行端；jsonschema 是生成工具，不能替代重复键、未知字段、null/缺省、金额范围和实体归属校验。严格模式 required/nullable 映射应单独适配，不污染财务语义。
2. 数据查询：UID 构造时固定；所有主表、关联及子查询都限定 UID 和有效记录。先规范转账逻辑交易，再筛选、分页、计数。负支出保留符号，分币种统计；count 的可加总性必须显式说明。
3. 完整性：分页只限返回，不限扫描历史；完整 totals 与分页 groups 分开。读事务覆盖验证归属、元数据、计数与结果，不跨 LLM 等待。游标是签名查询参数，不是授权或数据库快照。
4. 执行循环：中立 DTO 放 `pkg/llm/data`，harness 接收执行接口；服务不调用模型。完整一轮终结后才能执行工具，整轮工具预算不足则一个也不执行。缓存仅请求内且 UID 已绑定。
5. 协议：测试参数分片、截断、重复 call ID、Responses 必需 opaque items，以及跨 provider 重建中立上下文。最终正文缓冲到确认无工具的完整轮次；status 可实时。已输出正文/思考后不得切备用。
6. 网络与 UI：一处串行 SSE writer；写失败取消上游，心跳不是完成。损坏帧或无终态 EOF 都失败；旧回调不能覆盖新请求；未完成 assistant 消息不进入可信历史。
7. 迁移：忽略旧 embedding 段及 EBKCFP 文件读取；旧表与 schema 首发保留、不运行读写。禁止通过删除旧表证明退出 embedding。

## 外部参考与尚未验证的事项

本次没有联网核验最新供应商文档。下列是实施时应核对的官方入口，不作为已验证事实或锁定兼容承诺：

- OpenAI function calling：https://platform.openai.com/docs/guides/function-calling
- OpenAI Responses：https://platform.openai.com/docs/api-reference/responses
- Go database/sql：https://pkg.go.dev/database/sql （Conn.BeginTx / TxOptions / context 取消）
- PostgreSQL transaction isolation：https://www.postgresql.org/docs/current/transaction-iso.html
- MySQL transaction isolation：https://dev.mysql.com/doc/refman/8.4/en/innodb-transaction-isolation-levels.html
- SQLite isolation：https://www.sqlite.org/isolation.html
- WHATWG SSE：https://html.spec.whatwg.org/multipage/server-sent-events.html

版本兼容上游的工具能力、实际驱动取消、主从固定、默认隔离、内存与性能均需后续实测。若需要联网核对 OpenAI 文档，应由实施任务读取 openai-docs 技能后使用官方资料；本次仅研究仓库接口，不宣称获取了最新 API 规范。

## 验证建议与门禁

先建立手算 6 笔数据：CNY income 1000.00、expense 110.00、net 890.00、transfer_out 40.00、balance_modification 200.00；USD transfer_in 5.00；6 个逻辑交易。再测两用户隔离、2500+ 跨五年数据、100000+ 性能集、未来日期、高水位、特殊字符、DST、重复 ID、断流与预算。独立参考值不得只调用与待测实现相同的统计 helper。

执行阶段跑原方案第 12.3 节命令；数据库矩阵、真实模型、GUI 缺环境时记录待验证、缺少输入及可执行补测命令，不计为通过。发布必须等全部必要门禁。计划阶段仅检查文档路径、契约覆盖与依赖闭环。

可复用工作指引：[assistant-readonly-harness](../../.claude/skills/assistant-readonly-harness/SKILL.md)。本研究不自行评分；由独立 judge 按 plan-task rubric 记录。
