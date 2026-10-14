# AI Assistant Harness 最终实施覆盖审查

审查日期：2026-10-15（Asia/Shanghai）。实施基线：`77c00d1cca0ff268d221566fc1339d62624cc9b9`。依据[原方案](../../AI_ASSISTANT_HARNESS_PLAN.md) §§4–8、10、12、15–17、[执行任务](../tasks/in-progress/ai-assistant-harness.feature.md)、全部 14 份子任务及实际代码、测试、阶段审查和最终验证记录。

**原 14 步及新增 Docker 整应用 smoke/E2E 的本地实现与验证：PASS，9/10。未发现未解决的实质性本地缺陷或丢失的实施要求。真实主备模型财务验收仍待完成，不可据此发布生产。** 新增应用容器的构建、HTTP 和双页面交互已单独执行，未以原有三库 Docker 测试替代。

用户已另外授权向当前仓库 `XueHua-s/ezbookkeeping` 的 `main` 提交草稿 PR；这不等于授权生产发布，也不把真实模型待验证项标记为通过。此次遗漏审计未修改生产代码，未重复运行已通过的全套测试。

本文件保存可随代码审查交付的结论和关键证据，不依赖本地 scratchpad 日志才能理解状态。详细操作矩阵见[验收记录](../../docs/ai-assistant-harness-verification.md)，提供商和回滚步骤见 [AI provider 文档](../../docs/ai-providers.md)与[迁移文档](../../docs/ai-storage-migration.md)。

## 14 步逐项落地

| 步骤 | 实际实现与验证入口 | 审查结果 |
| --- | --- | --- |
| 01 中立契约 | `pkg/llm/data/large_language_model_data.go`、`assistant_contract.go`、Go/TS HTTP DTO、`pkg/services/testdata/ai_assistant_truth.json`；契约序列化和兼容测试 | 工具 snake_case/HTTP camelCase 分离；金额和 ID 为字符串；内部观测不进入用户证据；既有 provider 接口保留。 |
| 02 Context/只读视图 | `pkg/core/context_assistant.go`、`pkg/datastore/query_context.go`、`assistant_read_view.go` 及三库/清理故障测试 | 真正委托 deadline/cancel；固定主库连接和一致性只读事务；SQLite 恢复/回滚次序及坏连接丢弃；真实长查询取消和连接复用通过。 |
| 03 配置 | `pkg/settings/ai_assistant_runtime_test.go`、`setting.go`、后端及前端 server settings | 主段统一正数预算，默认/零/负值/溢出/EBK/EBKCFP 均覆盖；前端仅获运行超时，JSON/SSE 使用加 10 秒截止。 |
| 04 查询基础 | `ai_assistant_queries.go`、`filter.go`、`cursor.go`、`read_helpers.go` 及对应测试 | 固定七工具、同源 schema/执行校验、UID/删除/隐藏语义、字面关键词、同侧转账、用途隔离签名游标和真实观测；无残留未实现工厂。 |
| 05 Provider | `pkg/llm/provider/openai/openai_assistant_tools.go`、`openai_assistant_tool_protocol.go`、协议测试及容器 | Chat JSON、强制流式 Chat、Responses 的完整调用、结束信号、choice 隔离、私有状态、角色、缓冲/上下文限制、取消/超时均有模拟协议证据。真实模型另列待验证。 |
| 06 Runner | `pkg/llm/assistant_runner.go` 及 scripted provider/executor 测试 | 单一请求级 owner；顺序完整轮、预算前置扣减、配对历史、成功缓存、重复调用保护、一次切备、最终正文缓冲及单一观测终结。 |
| 07 明细 | `pkg/services/ai_assistant_transactions.go` 及测试 | 五年以上及未来账单、逻辑转账、详情鉴权、稳定时间/金额排序、静态分页无重漏、字节缩页真实末键；无附件路径/IP/similarity 投影。 |
| 08 聚合 | `ai_assistant_aggregate.go`、聚合测试和 100K benchmark | 全匹配范围、任意精度有符号累加、分币种/flow、非可加计数、ordinal 比较、非正基期、DST/跳日、有限零桶及深分页；不靠缩短查询范围达成性能。 |
| 09 元数据/余额 | `ai_assistant_metadata.go`、`ai_assistant_balances.go`、真实七工具 registry 测试 | 元数据和余额均可分页；隐藏有效叶子纳入、删除排除、父节点不重算；完整币种余额总计与页大小无关。 |
| 10 API/SSE | `pkg/api/large_language_models_assistant_harness.go`、对应测试、系统提示词 | 两路接同一 runner；历史/体积/角色验证，可信候选续查；四部分 summary、固定时钟、真实 HTTP/SQL 取消、串行写入及错误证据；已移除活动 RAG 路径。 |
| 11 共享前端 | `src/lib/assistant_stream.ts`、`services.ts`、`AssistantPageBase.ts` 及测试 | 无伪造 done；UTF-8/CRLF/坏帧/尾帧/重复终态检查；错误 context、取消/UUID 隔离、短期完整历史、summary 追问和明确下一页。 |
| 12 双端 UI | 桌面/移动 `AssistantPage.vue`、`AssistantMessageEvidence.vue`、真实 SFC 测试和 22 locale 测试 | 同一共享状态；范围/完整性/读取时间、隐藏金额、鉴权引用、安全 Markdown、取消和会话清理；实际双路浏览器 smoke 已执行。 |
| 13 退役/迁移 | 删除旧 embedding service/窗口/top-k/hash/snapshot；settings/API 精简；缓存拒访、FIFO 密钥零读取、实际 schema 更新保留测试；文档和两套技能 | 无旧配置/密钥/网络/cache 运行依赖，保留旧模型/schema/记录；已修复实际表名导致的测试假阳性，见下文。 |
| 14 全量证据 | `ai_assistant_database_integration_test.go`、benchmark、部署测试、最终验证文档 | 三库同一契约、100K 计划/性能、全部本地门禁、实际 GUI/SQLite 旧版本恢复、观测与 CK/HR 映射齐全；真实主备模型明确未通过。 |

条件性索引/迁移实现没有遗漏：已有索引在实测查询计划中满足测试数据，未产生新增索引的证据，因此未添加业务字段、表或索引；无需虚构一次不存在的 schema 迁移。未来索引工作仍须覆盖三库、分库、幂等和回滚。

## CK/HR 完整性复核

| 验收项 | 已验证内容与未验证边界 |
| --- | --- |
| CK-1 | 七工具真实 Definitions→Execute、元数据/余额分页、隐藏有效实体及字符串 ID。 |
| CK-2 | 严格同源 schema、默认/null/空数组/重复键/未知字段/字节边界、时间金额和字面关键词；非法组合在查询前拒绝。 |
| CK-3 | 主查询、关联、实体、详情、游标、引用和分库均绑定 UID；混合用户 ID 不退化为无筛选；无写工具。 |
| CK-4 | 2,504 笔完整明细遍历包含 2,500 笔跨 6.84 年交易及未来日期；100K 全量聚合；无历史/top-k/扫描上限。 |
| CK-5 | 六笔独立真值、退款符号、精确大整数与两位金额；各币种独立，极大值不溢出。 |
| CK-6 | 转出规范 ID/time、同侧联合过滤、逻辑去重、双 flow；缺失/跨 UID 关联显式失败。 |
| CK-7 | 秒级半开区间、层级展开、日历/DST/跳日、ordinal/币种/flow 配对、零桶、非正基期 null。 |
| CK-8 | 资产 2000.00、负债 signed -300.00、净资产 1700.00；隐藏叶子/溢缴/父节点不重算，页间总计不变。 |
| CK-9 | 真实 scope/coverage/queried_at/计数/全量 totals，错误不变为空集，成功证据和最多八个去重逻辑引用。 |
| CK-10 | HMAC 用途隔离、恒时比较、完整筛选/层级/时区/排序/高水位绑定、24 小时过期、重新验权及候选续查。 |
| CK-11 | 单一 JSON/SSE runner、完整轮顺序执行、全轮预算、缓存/重复 ID、防无进展、配对上下文及显式终态。 |
| CK-12 | 三协议及切备模拟通过；请求私有状态隔离、可见正文/思考后不切备。真实主备协议/模型表现待验证。 |
| CK-13 | 三个真实驱动的只读一致性、固定主连接、长查询取消/复用；HTTP→provider 和 API→真实 SQLite 查询取消。 |
| CK-14 | 原 JSON/SSE 外壳、错误 context、唯一终态、损坏流不成功、写/flush 失败取消及竞态测试。 |
| CK-15 | 共享五状态、范围/引用/下一页、请求隔离、离页/换用户清理、国际化及安全 Markdown；真实双页面 smoke，但不是实体手机认证。 |
| CK-16 | 历史只允许 user/assistant、最多 20/使用最近 12、Unicode 安全、当前问题完整；summary 四部分/同进度/显式范围及追问。 |
| CK-17 | 零 embedding 配置/文件/网络/cache 依赖；默认和正数预算、前端超时；旧表/记录保留。 |
| CK-18 | 具名测试、三库、100K 性能/执行计划、配置/迁移/两套技能、真实模型补测命令和未执行清单；新增 Docker 整应用 HTTP/双页面验证已完成。 |
| CK-19 | 独立工具和 API 真值，890.00 收支净额与 1700.00 当前净资产分开，时区跨月口径说明；模拟最终文字不等于真实模型推理，七题仍待验证。 |
| CK-20 | 中立 provider、薄 API、同一查询规则与共享前端；未引入 agent 框架；识图/其他 provider/context/cron 回归。 |
| HR-1 | 旧模型/schema/记录和用户原始文件保留；真实隔离旧版本恢复；最新用户授权仅允许提交当前仓库草稿 PR，不含生产发布。 |
| HR-2 | 临时库/显式 opt-in/测试 UID；观测和日志 canary 脱敏；前端无密钥、私有 provider 状态、存储路径/IP、游标原文展示。 |
| HR-3 | 实际 lockfile 保持、未做无证据迁移；本地/环境/可发布三状态分开；跳过未计通过。 |

## 阶段审查与已关闭发现

五阶段均在各自范围完成独立审查：Phase 1 为 9/10、Phase 2 为 8.8/10、Phase 3–5 均为 9/10。早期基础阶段的 PASS 没有被用作尚未集成部分的证明；Phase 5 再核对最终源码门禁。

| 阶段 | 已修复并复验的实质问题 |
| --- | --- |
| 1 | 关闭 SQL cursor 时内部取消可能污染成功读结果；已先冻结终态错误再取消。关闭 cursor wrapper 随批次累积；改为只保留活动 cursor 和首个终态错误，重复 race/真实驱动通过。 |
| 2 | strict nullable 参数错误重置已签名续查条件；空游标被接受；provider 局部 deadline 未进入允许的备用切换；显式删除/越权实体错误优先级漂移。修复后定向独立探针、全套与 race 通过。 |
| 3 | São Paulo 不存在的午夜和 Apia 跳日造成错误/零长时间桶，Monrovia 历史秒级偏移可能丢失真实边界。统一 civil-date 边界与序号规则，独立反例和三库/race 通过。 |
| 4 | 损坏 done 载荷被当成功；终态写/flush 错误未进入唯一运行观测；summary 缺 user 意图导致后续历史丢失；错误文案、当前余额范围和零目标转账展示。修复并通过真实 SFC/API/流解析/竞态测试。 |
| 5 | 缓存拒访测试监听模型文件名推导的错误表名，原零计数为假阳性。改为 `engine.TableName`，先断言 SELECT/UPDATE 都被拒绝，再运行 JSON/SSE 并断言零访问和旧记录完整。独立 race 1.458s 通过；迁移/provider/技能文档同步物理表名。部署验证也从 OpenAI 硬编码字段改为通用工具能力检测。 |

默认物理缓存表为 `a_i_assistant_embedding`，不等于文件名 `ai_assistant_embedding.go` 去掉扩展名；前缀/schema/mapper 可改变实际名称。旧 schema 没有被重命名。原步骤 13 的拒访测试结论已由修正后的正向控制结果取代。

## 最终执行证据摘要

环境：macOS 27.2、Apple M5/16 GiB、Go 1.27.1、Node 22.20.0、npm 10.9.3；SQLite 3.53.4、Docker MySQL 8.0.46 和 PostgreSQL 16.15。全部账本为合成测试数据。

| 门禁 | 实际结果 |
| --- | --- |
| `BUILD_PIPELINE=1 go test ./...` | Exit 0，缓存拒访修复之后执行。 |
| `BUILD_PIPELINE=1 go vet ./...` | Exit 0。 |
| `go test -race ./pkg/datastore ./pkg/core ./pkg/api ./pkg/llm/... ./pkg/services ./pkg/settings ./cmd ./pkg/templates` | 带三库显式 opt-in，exit 0；services 59.331s。另跑 cron race，exit 0，13.562s。 |
| `gofmt -l` 变更/新增 Go | 65 文件无待格式化内容。 |
| `npm run lint` / `npm run test` / `npm run build` | 全部 exit 0；17 个测试文件/38,714 项；build 38.58s。lint 前后未发现无关自动修改。 |
| 三库同一 `TestAssistantDatabaseContracts` | 所有驱动实际执行，exit 0，15.595s；非 skip。 |
| 三库 `TestAssistantDatabase100KPlans` | exit 0，32.150s；独立完整 100,000/-500.00 真值及实际执行计划。 |
| `BenchmarkAssistantAggregate100K -benchtime=1x -count=1` | 五查询 exit 0，26.179s；各查询读取 100,014 或 100,013 应用行。 |
| 严格 CRLF-aware diff check | exit 0：`git -c core.whitespace=blank-at-eol,blank-at-eof,space-before-tab,cr-at-eol diff --check`。普通检查仅报原有西语 CRLF 格式新增行，不转换用户文件行尾。 |
| 原始文件保护 | 原方案、原有 dirty `skills-lock.json`、未跟踪 `pnpm-lock.yaml` 哈希复核一致；package-lock/go.mod/go.sum 未变；三组技能/引用副本逐字节一致。 |

六笔真值：逻辑 count=6；CNY 收入 1000.00、支出 110.00、净额 890.00、转出 40.00、余额调整 200.00；USD 转入 5.00；支出 count=3/min=-20.00/max=100.00。独立余额净资产 1700.00。2,500 日数据的独立算术合计 31237.50，所有 2,504 逻辑明细和 2,500 并列日分组完整到达。

100K 五查询测量依次为全历史 1.346573s、月/分类首页 1.349214s、分组深页 1.310396s、明细首页 1.303937s、明细深键 1.270028s。采样峰值堆增长约 1.91–2.80 MB，累计分配约 91.85–173.09 MB，两者不能混称。5ms 内存采样会影响耗时；非延迟 SLA。深 50,000 键为签名测试游标准备后的一次请求，不冒称浏览器连续遍历 50,000 条。应用读取行数、计划估计/实际扫描、返回记录数和字节分别记录。

观测覆盖正常、多工具、缓存命中、预算、失败和取消，含受控原因/request ID/次数/字节/耗时；JSON 首状态为 null，心跳不计首状态。注入财务/密钥/游标/私有状态 canary 后断言观测和错误不泄漏。890.00/1700.00 及请求/记录时区说明由真实工具结果加模拟最终文本验证，未扩大为真实模型推理证明。

## 实际界面与恢复验证

最终浏览器 smoke 使用退役后重建的后端和最终生产前端构建，连接实际 API/runner/SQLite 与 localhost 模型协议桩。桌面和移动路由均通过完成状态、Asia/Shanghai 范围/读取时间、隐藏金额、安全 Markdown、明确下一页、取消清除续查、503 本地化错误、清空与摘要意图。取消/失败时临时 rejection 监听为零。鉴权引用分别打开桌面对话框和移动详情，离页清理会话。

移动路由运行于桌面嵌入式浏览器，不是实体手机。启动时无助手请求也可复现的 ECB TLS 证书/汇率 400 和未处理拒绝被记录为既有本地环境问题，未宣称控制台完全干净。

恢复演练从基线提交导出旧源码到临时目录并编译成功，使用 SQLite online backup 复制合成数据库，以旧配置仅调整隔离路径/端口启动旧二进制。旧 schema 同步保留真实物理缓存表的一条 model/hash/vector 哨兵；旧/新认证账户和月账单 JSON 完全相等；仍为 3 条交易、signed sum=37038 minor units、账户余额=170000。此为实际隔离 SQLite 旧二进制/配置恢复，不是生产切流、其他数据库旧版本降级或旧 embedding 供应商可用性验证。

## 未完成环境项与新增请求

真实主/备用模型各须执行 Chat、强制流式 Chat、Responses 的 JSON/SSE 工具 smoke，并用独立合成账本真值核对：去年餐饮逐月合计；五年前大额交通；今年至今与去年同进度及分类；含隐藏账户当前余额；“只看支付宝→上个月”指代；全历史分币种收支；跨用户查询拒绝。还需核对同答 890.00/1700.00 与请求/记录时区跨月差异。缺少真实供应商凭据和专用部署/用户，因此这些项目明确待验证，不能把 mock 通过算为通过。

新增 Docker 整应用 smoke/E2E 也已完成，见下节；此项通过没有改变真实模型待环境状态。

9/10 反映清晰的 query/provider/runner/transport/UI 职责与已闭环的错误、数据和证据边界。进一步提高信心需要真实模型与部署环境证据，不靠再给模拟测试加分。当前可进入授权的草稿 PR 审核；**生产发布仍为 NO**。

## 新增 Docker 整应用验收闭环

审阅了[持久 Docker 验证报告](../../docs/ai-assistant-harness-docker-verification.md)、实际 HTTP 断言/结果、容器运行观测、源码/产物一致性记录，以及协调者的两路浏览器操作交接；独立查看桌面和移动截图，范围/时区/隐藏金额等投影与报告一致。**Docker 本地 smoke/E2E PASS，无新增实质性缺陷。**

- 构建在 Linux/arm64 的 `golang:1.27.1-alpine3.24` 中启用 CGO；实际完整后端编译 exit 0。运行镜像基于 `alpine:3.24.2`，使用项目 entrypoint、UID/GID 1000、实际 templates 和最终已验证 dist，启动/schema 同步成功。冻结源码及前端产物与 checkout 逐字节核对。
- 运行容器无 host bind mount，只含专用合成用户/SQLite/config；loopback 端口、独立网络及协议桩。镜像 ID 为 `sha256:3f7e65e9dd60d117cb6c71947f4a983dc8f9299bb11446d0ac1945306b1910c3`。
- 实际 HTTP 登录成功、未登录助手请求拒绝；JSON 返回七个实际工具证据，SSE 运行观测逐个记录七工具、两轮及一个 done。显式续查从交易 300 到 301；503 返回 failed/error 而非 done；关闭客户端 SSE socket 后协议桩观察到上游 TCP EOF，随后新请求正常完成。
- 账务仍为三笔、金额和 37038 minor units、余额 170000、旧缓存哨兵一条。实际运行日志不含合成密钥/备注 canary、完整回复或 Bearer；JSON 首状态 null，SSE 首状态和取消原因存在。
- 协调者在 Docker 发布端口实际操作桌面和移动路由：完成/范围/Asia-Shanghai 时区/当前页、隐藏金额 `***`、下一页 123.46、取消清空续查、503 本地化失败、清空和摘要意图均通过；危险链接零、助手操作未处理拒绝零。鉴权引用打开桌面对话框和移动 `id=300` 详情，移动离页清空会话。
- 验证使用最终 dist 的测试镜像，不冒称重跑 stock Dockerfile 的前端构建阶段；协议桩只覆盖此次 Docker Chat 链路，三协议另有已有协议测试。不是实体手机、amd64、真实模型或生产发布认证。报告内含可保存执行的 Python stub，避免依赖未提交的 scratchpad 文件。

最后的 PR 包装复核关闭了本地日志目录链接问题：持久验收文档自带三库创建/补测命令、恢复方法和结果，明确原始日志仅在本地；持久审查与 Docker 报告互相链接。主任务、步骤14和需求索引已同步本地完成/真实模型待验证及当前仓库草稿PR授权。可进行授权的提交/推送/草稿PR流程。
