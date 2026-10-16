# 03 · Project structure 串行审查

2026-10-16；基线 `91081998` 相对 `origin/main`，包含 gate01/02 未提交修复。

**初始 8/10；修复后 10/10（本门模块归属/地图范围通过）。** 文档漂移已修复，未发现需搬迁产品代码的确证。最新串行复核已覆盖后续模块拆分；gate01 已关闭 D4–D6，gate07/08 最终 scope 已通过。若继续修改模块边界或路径，本门需要复审。整个 PR 的本地跨层质量检查已完成，执行证据见15–18；外部发布限制仍保留。此评分是本次审查约定，技能本身没有评分量表；不把静态路径核对当作运行通过。未启动子 agent、Docker、全仓测试或 build，未 commit/push、未访问私有配置。

## 最终增量闭环（Gate15–18 后）

**当前范围最终 10/10，结构地图闭环并冻结。** 仅文档修正，无产品/总审查表修改、委派、提交或测试重跑。读取15–18报告的模块、执行及限制说明；批量报告输出部分截断后，定向补读16复现/隔离段，不声称重新验证全部执行证据。实际读取 bounded-command、build-resources、migration runtime 的职责实现、compatibility 四个测试场景，并搜索其真实调用点和 session 中404分类；核对 `.test.mjs` 的 Vitest imports 与 `vitest.config.mts`。

补全构建/容器 smoke/E2E、POSIX限时命令、task build resources、Python迁移运行边界及 compatibility 回归的必要入口，分别归于 scripts 验证基础设施和共享 session。无需移动产品模块。

**纠正本门上轮错误：** environment/cleanup `.test.mjs` 实际使用 Vitest，不能用 `node --test`。已修订结构/dependencies镜像为 `npm run test` 或定向 `npx vitest run`；新 bounded-command/build-resources 回归同属 Vitest。Python helper 仍用独立 unittest 命令。下方上轮记录保留为历史，其“Node harness tests 和 Vitest 的区别”表述由此更正，不视为有效命令指导。

验证：4对镜像逐字节相等，**178个具体路径引用存在**（含重复），git diff --check PASS。没有重跑15–18测试，也未将其既有 PASS 转记为本门执行结果；商业模型、物理设备与真实bfcache等边界仍按各报告保留。此前“最终质量阶段未完成”的描述是当时状态，最新已执行项目以15–18和父审为准。

## 最新结构复核（后续串行门整改后）

**最终维持 10/10，仅限结构归属和地图准确性。** 本轮只改文档，冻结后交回父审；不修改总审查表、不新增产品实现、不委派、不执行 Docker/full matrix、不 commit/push。

重新读取 actual structure SKILL 与两份 references，按 `rg --files` 核对实际文件名；读取 workspace Scope/Panels/Lifecycle 全文、session/conversation/commands 的职责与状态入口、ledger fence/transaction bridge 全文、token caller lease 定义和 transport/unlock 引用、package.json、review scripts 入口及 AI feature map。此前 D4/D5 的“待关闭”描述属于首轮历史；gate01 的最终复审已记录 D4–D6 关闭。本轮确认新模块仍位于 shared UI/base/store/service 正确层，没有发现新的归属缺陷。

本轮地图增量：

- 补 shared dialogs/toolbar/notices/recovery/CSS、workspace 生命周期/作用域/面板分工，以及 session/conversation/commands/history 所有权。
- 补通用 token refresh caller lease 和 auth/local identity 回归入口；ledger-clear 同事务 fence、version reuse 及 memory projection 契约测试归属。
- 保留已由 gate12 同步的 `.mts` 配置路径；补 npm/lockfile/CommonJS 约定与 Node harness tests 和 Vitest 的区别。
- 修正 AI feature map 的旧 stateless adapter/list 被误写为当前 persistent 页面入口，以及 facade 被误写为 transport 的漂移；两份镜像同步。保持简洁分层地图，不列所有测试函数。

验证：**4 对镜像逐字节一致；151 个具体路径引用（含重复）及 7 个相对组件/CSS 路径存在；git diff --check PASS。** 当前 manifests 差异仅 package.json 的 browser config `.ts` → `.mts` 脚本路径，go.mod/go.sum/package-lock.json 无变化；首轮“manifests 无变化”结论保留为当时证据，不代表当前脚本完全无差异。

没有执行新增测试/host review scripts。脚本路径是定位入口，不能作为 Docker/E2E 验收通过证据；最终验收入口后续若新增需补图。真实模型/手机键盘/screen reader/bfcache 状态仍以最终证据为准，本门未验证这些运行项。

## 首轮记录（以下为当时结论，后续状态以上方复核为准）

## 依据及实际读取

完整读取 `skills/ezbookkeeping-project-structure/SKILL.md`、`references/project-structure.md`、`references/dependencies.md` 和 `.codex/skills/ezbookkeeping-project-structure/` 下三个同名副本。先读取总审查契约、01/02 报告；扫描本 PR 路径清单，核对 dependency manifests 差异。此门修改六份技能 Markdown 和本报告，没有 Vue/TS/Go 源码修改，因此不触发 Vue 实现/测试参考必读；平台组件只作结构追踪。

明确阅读的关键实现：

- `cmd/webserver_assistant_lifecycle.go`、`pkg/api/large_language_models_assistant_conversations_runtime.go`、`pkg/api/large_language_models_assistant_conversations.go`；路由文件 assistant 注册片段、router 的 assistant 入口与 database 的模型注册位置。
- `pkg/services/ai_assistant_conversations_send.go`、`pkg/services/ai_assistant_context.go`、`pkg/services/ai_assistant_memories_projection.go`、`pkg/llm/assistant_contract.go`，OpenAI tool adapter 的构造/RunToolTurn。存储 commands 和 conversation/message 模型作定向阅读，批量输出部分截断，不声称逐行完整读取其全部实现。
- `pkg/settings/setting_ai_experience.go` 全文，`setting.go` 常量片段、`setting_config.go`/`setting_load.go` 配置注册搜索；`package.json`、`go.mod` 直接依赖片段。
- 两平台 `AssistantPage.vue` 与 `AssistantConversation.vue` 全文，`useAssistantCloudWorkspace.ts` 产品命令，`src/stores/assistant/history.ts`、`src/lib/services/assistant_conversations.ts`、`src/lib/services.ts` 全文。

## 适用要求、证据与结论

| 要求 | 证据与判断 |
| --- | --- |
| 先明确 feature/data ownership、路由追到 service/model | `messages/send.json` 注册 → Handler 使用登录 UID、解码/限流 → Conversation.Send → ExecuteCommand 的事务回调创建 question/run 并返回 receipt。Send 内完成 context/version、来源、额度检查；HTTP 不直接操作 xorm。五模型注册在 UserDataStore，run 保存在 AssistantMessage。未把路径索引当作每个 endpoint 的语义审计。 |
| 最低正确层修复，职责边界明确 | 配额/回执/epoch 在 storage services；prepare 组合 owned history、memory 与 UID-bound query service；LLM runner 接收 ToolExecutor，不能任意选择数据库 UID。API runtime 为 composition root 注入服务依赖，cmd 管应用生命周期。其在 API 类型上提供初始化方法符合现有容器组织，不为纯目录美观引入新层。 |
| provider/model protocol 边界 | OpenAI adapter 负责请求构造、HTTP 执行、stream/JSON 协议，使用既有 httpclient（proxy/TLS/timeout）；context service 调 provider-neutral interface。定向 import 搜索未发现 runner/provider 反向依赖 API/services/datastore，也未发现生产 assistant service 自行发 provider HTTP。测试使用 httptest 不算边界泄漏。完整协议/精度正确性交给 AI 门。 |
| 配置单一权威 | ExperienceConfig 集中 default/load/validate；Config/setting_load 注册，API capabilities 输出生效 Limits/owner/epoch，runtime 注入配置闭包，Send 记录 policy。旧技能只指向 setting.go、只建议 LLMConfig 和 server_settings，容易把持久化配置放错层，已修正。 |
| 共享 frontend 与平台入口 | 两个 router → 平台 page → platform Conversation adapter → 同一个 CloudWorkspace。平台仅处理 Vue Router/Framework7、账单打开方式及页面激活；产品命令在 base composable、持久状态在 Pinia，wire contract 在 service client。D4/D5 的过大组件/压缩 store 仍需后续设计整治；此门确认其平台/共享层归属，不替代内部设计评分。 |
| 新 helper 的归属 | gate01 history.ts 持有分页窗口、ticket 和刷新状态，归于 store 内部模块合理；gate02 memory projection 是 service 私有纯投影，两调用方仍各自负责授权/IO/错误政策，既不应进 HTTP，也不需新通用工具包。 |
| 依赖复用 | `git diff origin/main -- go.mod go.sum package.json package-lock.json` 为空；JSON 请求复用 Axios transport，SSE 用 fetch + 专门 bounded reader，provider 复用 httpclient，未新增依赖或平行 SDK。没有依赖升级需验证的事项。 |
| 跨层质量门 | 本轮只有文档变更，执行副本/路径/diff 检查即可。整个 PR 的 Go/frontend gates 依契约留给最终串行执行；不将未执行项目记作 PASS。 |

## 本轮修正

同步修改 `skills/` 和 `.codex/skills/` 下三份文件：

1. SKILL 导航不再误称配置都在 setting.go、API 调用都在 facade、共享行为都在 base views；补 route registration 与 shared components/store 的边界。
2. project-structure 明确 legacy stateless 与当前 persistent endpoints，收窄 read-only query 文件提示，补 datastore/五表/run/生命周期/客户端完整 trace，以及新 helper 所有权。保留旧 adapter 路径作为 legacy，不将其描述为当前页面主入口。
3. dependencies 区分 provider、runner、experience config；明确 capabilities 与 general server flags，指出 npm lint 会自动修复文件并补 browser test 命令。依赖版本以 manifests 为准。

没有为了分数新增抽象、迁移文件或产品代码。

## 定向证明与后续

- Python 对三对技能文件逐字节比较：**PASS**。
- 从修订技能中提取具体 repo 路径并检查存在：**72 个引用 PASS**（重复计数，不是72个独立文件）。glob 和符号名不计入路径证明。
- `git diff --check`：**PASS**。
- Dependency manifests 对 origin/main：**无变化**。
- 未执行应用测试：本轮纯文档改动，无新增测试的必要；此前 gate01/02 测试由各自报告负责，未在本门重复宣称运行。

交接：Vue/frontend-design/Pinia 修复 D4/D5 后，复看共享组件/状态模块的接口与归属，并同步此地图如路径改变。AI gate 继续审 provider/UID/精度/并发；最终阶段执行跨层质量矩阵及 Docker/E2E。真实模型、手机键盘、screen reader 和真实 bfcache 不属于本门路径审查的验证范围，仍未验证。存储迁移/备份恢复执行与新增依赖/toolchain 选择在本轮修改中 N/A，不代表整个 PR 免除迁移验收。
