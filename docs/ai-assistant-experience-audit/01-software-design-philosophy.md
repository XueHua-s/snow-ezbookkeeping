# 01 · Software Design Philosophy 串行审查

日期：2026-10-16。基线：`91081998` 相对 `origin/main`。审查者：`gate01_design_philosophy`。本文件保留首轮发现，并记录 Gate02–14 后的最终设计复审；不是发布验收，也不代表已证明零 bug。

**最新：产品实现及验证工具设计评分 10/10；Gate18 最终复审已关闭 D8 构建/冒烟子进程生命周期缺口。** 初始 7.5/10 → 首轮修复 8.5/10 → Gate14 后产品设计复审 10/10 → Gate15–17 工具增量 9.5/10 → Gate18 修复复审 10/10；D4–D6、D8 均已关闭。见文末最新增量复审。 10/10 仅评价本项软件设计原则与已检查的实现边界，不是全部功能、实际部署或外部模型验收评分。没有启动子 agent，没有 commit/push，没有运行 Docker 或全仓测试矩阵。

## 技能依据与审查范围

完整读取 `skills/software-design-philosophy/SKILL.md` 及六份 references：complexity-symptoms、deep-modules、information-hiding、general-vs-special、comments-as-design、strategic-programming。结合 `.codex/skills/ezbookkeeping-project-spec/SKILL.md` 与 global-prompt。前端修改前读取 `skills/vue-best-practices/SKILL.md` 及 reactivity、sfc、component-data-flow、composables 四份必读参考；定向测试读取 Vitest skill。

检查了 PR 全部变更路径清单和后端服务/API 接口索引；详细阅读下表关键边界与调用链。图片、locale 和测试工具仅归类，并未将路径清单扫描声称为逐行审查全部 272 个文件。其余实现由后续 spec/structure/AI/Vue/UI 串行门独立审核，最终复审需结合其证据。

| 原则 | 重点阅读文件/调用链 | 结论 |
| --- | --- | --- |
| 深模块 | `ai_assistant_storage{,_commands,_dependencies}.go`，`ai_assistant_runs{,_execute}.go` | 事务模块隐藏数据库重试、CAS、回执；worker 隐藏租约与 checkpoint。不是简单 repository 套壳。补充了 Tx 生命周期契约。 |
| 信息隐藏 | `ai_assistant_context.go`，`ai_assistant_conversations_projection.go`，`ai_assistant_conversations_export.go`，`ai_assistant_memories{,_management}.go` | 模型请求预算和历史投影集中；发现 memory 列表/导出投影重复（D6）。 |
| 复杂度下沉 | `large_language_models_assistant_conversations{,_runtime,_stream}.go`，`ai_assistant_cleanup_sweep.go` | HTTP adapter 负责鉴权/参数/限流；SSE 为持久快照只读投影；清理采用有界批次。多个同 receiver Go 文件本身不是 classitis。 |
| 单一事实来源 | `src/models/ai_assistant_conversation.ts`，`src/lib/assistant_conversation_stream.ts`，`src/stores/assistant/types.ts` | 修复重复的运行终态知识（D1）。 |
| 状态边界 | `src/stores/assistant.ts`，`assistant/{dataClear,reference}.ts`，`src/lib/services/assistant_conversations.ts` | 幂等命令/本地 pending 保留在 store；独立抽出列表窗口知识，并修复 await 后越界写入（D2/D3）。 |
| UI 模块边界 | `AssistantCloudWorkspace.vue`，`useAssistantCloudWorkspace.ts`，`useAssistantReference.ts`，共享组件目录 | 双平台共用产品命令合理；CloudWorkspace 同时管理多个对话框、导出、路由与焦点，且极度压缩模板，仍不达满分（D4）。 |
| 注释/战略设计 | 上述公共接口与生命周期 | 关键不变量已有注释；补充 Prepare 必须在事务外、授权需后续重验。避免为 500 行要求继续合并语句。 |

## 已修复

### D1 · 终态契约在 transport/store 重复

此前 SSE parser 和 store 各维护一份 `completed/incomplete/failed/canceled/interrupted`，类型本身还有第三份同样知识。现在模型模块定义终态和全部状态，类型从其推导，parser 与 store 使用同一 `isAssistantRunTerminal`。运行状态扩展无需同步修改两个消费者中的字符串数组。未新建通用框架或额外薄包装。

### D2 · inaccessible 清理越过身份/导航边界

`synchronize()` 确认不可访问后等待本地清理，原实现随后无条件 `applyLocal`/`open()`。期间用户切换会话或账号会被旧完成事件覆盖。现在保存 storage owner，并在 await 后校验 epoch/navigation/owner。新增两个延迟清理回归，分别切换会话与身份，断言新会话、草稿、本地快照不被旧结果覆盖。

### D3 · 历史列表知识与过期失败处理

原 store 同时管理 history/search ticket、归档筛选、分页上限、可见时刷新。抽为 `assistant/history.ts`，接口隐藏这些细节，store 只传身份 generation。它同时丢弃过期成功与过期失败；父审指出旧 history 401 若漏过 catch 会 reset 新账号，此问题也已修复。新增旧 history 401、旧 search 失败在身份/请求变更后的回归。

这是按领域知识分割，不是按请求的“读取/解析/保存”时序拆模块。接口注释解释为什么 sidebar visibility 跨 identity reset 保留。

### D7 · 隐含接口前提

`AssistantStorageTx` 增加回调内有效、不可逃逸异步任务、所有直接 SQL 必须限定 State.Uid 的契约；`AssistantContextService.Prepare` 增加会调用模型/压缩历史、不能放入数据库事务、依赖还必须在生成与 checkpoint 时重验的契约。仅改注释。

## 首轮未关闭问题及交接（已由文末复审关闭）

### D4 · CloudWorkspace 承担多项独立职责，且压缩模板妨碍审查（中）

`src/components/common/assistant/AssistantCloudWorkspace.vue` 的 toolbar、pending 恢复、reference notice、export dialog 与 confirmation dialog 大量写成整行；脚本中包含导出下载、焦点管理、路由同步、viewport、确认 promise。该组件已使用共享产品 hook，但仍不是清晰的组合界面。

交给 Vue/frontend-design 串行门：将有独立状态/生命周期的 dialog/export 处理封装到合适组件或 composable，保持 props/emits 与必要 imperative API 小而明确；展开多属性模板及 CSS。不得仅把代码移到多个只有转发的组件，也不得改变焦点恢复、账号 fencing、草稿恢复契约。改后用已有浏览器用例与相关新回归验证。

### D5 · store 控制流仍过度压缩（中）

`src/stores/assistant.ts` 从 500 降至 491 行，但仍有 `reset`、subscription、accept、executeOnce 等一行多个变更/异步分支。行数减少不等于复杂度问题已解决。尤其 epoch/navigation/owner 检查必须清楚可见；本次发现的 D2 说明压缩有实际维护成本。

交给 Vue/Pinia 串行门：按一致的生命周期或命令事务知识整理必要边界，展开关键异步控制流；若触及 500 行，提取真正拥有状态与约束的模块，避免十几个 callback 参数的薄 helper。不得用压缩语句通过行数门禁。最终本项复审接口是否比实现简单。

### D6 · memory 公共投影存在重复（低，但阻止满分）

`ai_assistant_memories.go:List` 与 `ai_assistant_conversations_export.go:exportMemories` 分别解析 SourcesJSON、构建 AssistantMemoryDTO、生成 sourceMessageIds。已经有细微漂移：List 忽略空 MessageId，export 不忽略；List 使用当前 schema 常量，export 使用记录 schema。当前规范记录通常掩盖差异，但后续演进存在修改放大风险。

交给后端 spec/AI 串行门：以一个投影函数收敛公共格式，明确 schema 与空 source ID 契约；调用方可以保留列表忽略损坏记录、导出报错的差异。补与公开契约有关的测试，不只测试函数逐字段照抄。不要合并两个权限/IO职责不同的服务。

## 定向验证

- `npx vitest run src/stores/assistant/assistant.test.ts src/lib/assistant_conversation_stream.test.ts`：2 files / **50 tests PASS**；新增 5 个竞态回归。
- 回归有效性：visible-history 401 测试使用下一事件循环 tick 排空完整异步错误链；临时移除过期失败 guard 后测试按预期 FAIL（authenticated false），恢复后 50 tests PASS。红灯日志 `/tmp/assistant-design-red.log`。
- 针对修改的 6 个 TS 文件运行 ESLint：PASS。
- `npx vue-tsc --noEmit`：PASS。
- `npm run check:source-lines`：1663 files，0 exceed 500。
- `git diff --check`：PASS。
- 本项 Go 变更仅为接口注释，没有重新跑 Go 全仓；最终串行测试阶段负责 Go tests/vet/race、Docker、真实路由 E2E。

首轮这些测试不覆盖当时尚未整改的 D4–D6；以下最终复审独立检查后续实现，不以旧测试代替整改证明。


## Gate02–14 后最终串行设计复审

重新完整读取实际 `skills/software-design-philosophy/SKILL.md` 和本报告，实际阅读 Gate13、Gate14 报告以及 Gate04 重新打开的清账性能修复记录。没有根据其他 agent 的评分直接沿用结论。本轮未改生产代码：没有发现必须为设计门额外修改的实质缺陷，不为达到分数引入新的框架或层级。

### D4 关闭：组合界面与 UI 生命周期分开拥有知识

实读 `AssistantCloudWorkspace.vue`（398行）、`AssistantConfirmationDialog.vue`、`AssistantExportDialog.vue`、`useAssistantWorkspaceScope.ts`、`useAssistantWorkspacePanels.ts`、`useAssistantWorkspaceLifecycle.ts` 及产品命令 hook。

- Workspace 现在主要组合 history、toolbar、reference、recovery、notices、messages、composer 和 dialog，模板属性与异步控制流已经展开。独立 CSS 并不新增运行时接口；不以文件行数作为设计证明。
- Confirmation 的唯一 imperative 接口 `confirm(message)` 隐藏 pending promise 替换、context 失效、unmount 拒绝、下一 tick 展示和焦点恢复；Export 的 `open(item, all)` 隐藏选中版本快照、下载取消、晚到结果屏蔽、对象 URL 生命周期和文件名约束。这两个模块都拥有实际行为，不是仅转发 props 的包装。
- Panels 拥有响应式测量、原生窄屏 dialog、键盘和焦点恢复；Lifecycle 拥有路由/branch 恢复/attachment 与卸载。传入的 navigate/restoreAnchor/run 分别是平台适配、DOM能力、产品错误呈现，不要求调用方管理内部 generation 或监听器。
- WorkspaceScope 统一 UI authority 的 away-and-back/销毁判定；产品命令确认返回可再检查的 lease，不把早先 boolean 确认当作永久授权。保存、导出和导航后的副作用各自明确检查当前 scope。

这些拆分按知识归属而不是“读→处理→写”时序；没有再把同一套 draft/command 持久化规则复制进组件。D4 关闭。

### D5 关闭：深的身份、当前会话和命令模块

实读 `src/stores/assistant.ts`（313行）、`assistant/session.ts`（237行）、`conversation.ts`（433行）、`commands.ts`（416行）、history/dataClear 与各调用点。

- Session 拥有身份 generation、local owner、服务端时钟偏移、草稿修订、单调快照及销毁；capture 统一调用方需要的 authority 判断。
- Conversation 拥有导航 generation、90条窗口、版本选择、reading positions、SSE/backoff/polling/attachment，调用方无需重建这些资源策略。
- Commands 拥有 exact persisted request、in-flight 去重、delivery、receipt reconciliation、branch target acknowledgment。send/retry/branch 等领域入口保留是为了不同语义的校验，不能简单合成不透明参数包。
- Store 是协调入口；模块之间保留的 synchronize 回调用于命令完成、终态和前台恢复的重新同步，具有实际跨模块职责，不是为了凑行数传十几个字段/setter。现有界面适配的 broad return surface 对应真实功能，没有增加同功能层层委托。
- await 后 stale success/error 检查、reset/dispose 及 subscription cleanup 现在能够顺序阅读。State ownership 仍在同一 Pinia store 生命周期内，未建立第二份业务真相。

D5 关闭。较长文件本身不是浅模块证据，也不再要求为长度继续切碎。

### D6 关闭：memory 公共投影单一权威

`pkg/services/ai_assistant_memories_projection.go:assistantMemoryDTO` 同时服务 `List` 与 `exportMemories`。函数明确输出当前 public schema，空/null SourcesJSON 正规化为空数组，conversation-only 来源留在 Sources、不生成空 message ID。列表跳过损坏行、导出拒绝损坏行的区别由各调用方保留。

独立实读并执行 `TestAssistantMemoryListExportContract`：真实 list/export 的结果相等，覆盖旧 row schema、空/null 来源、conversation-only 来源和损坏记录策略。格式知识不再重复，权限/IO职责未被合并。D6 关闭。

### 后续新增边界复审

| 新边界 | 实读与判断 |
| --- | --- |
| reason/runtime 契约 | `src/models/ai_assistant_conversation.ts` 从 `assistantReasonCodes` 推导 union，SSE消费同一数组；终态也维持单一模型权威。父审首轮留下的重复 reason 观察已关闭。 |
| IndexedDB 默认身份迁移 | `src/lib/assistant_local_storage.ts` 默认 browser 命名空间由同源数据库定义；migrateDevice 在 owner readwrite transaction 内保留 draft/pending，保留 payload/rejected/branch target，冲突草稿明确标记。显式 deviceId 仍隔离；quota/fallback/recovery/epoch 不外泄到 UI。没有创建新的“先初始化身份、再打开数据库”双重协议。原子迁移设计成立；真实原生跨 tab 证据来自 Gate14，未把本轮静态审查当作重跑。 |
| auth transport/store | `transport.ts`、`stores/token.ts`、assistant service 的 expectedOwner、refresh/revoke 的 expectedToken。transport 负责单次 refresh 合并、队列释放/拒绝与限次 refresh-only retry；store 负责整次用户设置应用的合并及 continuation 身份检查。两层合并保护不同副作用，不是重复 wrapper。普通 POST 不重放，token/owner guard 的不同力度对应真实刷新需求。 |
| Framework7 adapter | `components/mobile/assistant/AssistantConversation.vue` 用 `updateCurrentUrl` 处理当前助手页的内部 replace，普通用户历史导航仍用 navigate。实现没有新增一套 route state machine，也没有让共享 Workspace 知道 Framework7 transition 细节。 |
| 账户清账版本证明 | `ai_assistant_storage_ledger.go`、storage 的 withUserTransactionVersion/userTransaction、`transactions_clear_context.go` 和 delete 调用点。guard 仅拥有单用户单串行 operation 的成功版本；锁前版本由 storage 捕获；+1/+1 CAS 算术只在 storage。apply 在 Commit 成功之后发布 proof；首次/任意介入助手写入重新失效。callback rollback、最终CAS失败、Commit失败不发布证明。普通单笔删除仍走原路径，逐笔部分提交保留。单个 conditional adapter 负责两种事务路径有语义价值，不是可消除的空转发层。 |

清账优化将无助手写入交错时的重复会话扫描移入 guard，不把“是否需重扫”推给交易调用方。持续写入交错时保守重新扫描是正确性成本，不应再通过忽略版本变化换性能。专门读了 reuse 测试的新 run、同 turn regenerate、部分失败、rollback/final CAS retry，以及实际 SQL 页数测试，未把仅调用 helper 次数当业务正确性证据。

### 本轮独立验证与评分边界

- `npx vitest run src/stores/assistant src/lib/services/transport.test.ts src/stores/token.test.ts src/lib/assistant_conversation_stream.test.ts src/components/common/assistant/AssistantConfirmationDialog.test.ts src/components/common/assistant/AssistantExportDialog.test.ts src/views/base/assistant/useAssistantCloudWorkspace.test.ts`：**11 files / 139 tests PASS**。日志 `/tmp/gate01-final-design-vitest.log`。
- `BUILD_PIPELINE=1 go test -json ./pkg/services -run 'TestAssistant(MemoryListExportContract|LedgerClearReuseDatabaseContracts|LedgerClearScaling)$' -count=1`：**9叶子PASS / 16外部数据库叶子SKIP / 0FAIL**。日志 `/tmp/gate01-final-design-go.json`。没有把 skip 计为三库通过；这是定向复核，不是全仓矩阵。
- Gate14 父独立 auth 21项、真实 Framework7/workspace 16项，以及 native 两端14场景连续两轮 PASS 为补充既有证据，本轮没有重跑或重复计数。
- 本轮只更新本报告，没有生产代码变更；不需要为文档更新重跑 build/browser。后续父审仍要串行 Pinia/auth、router 与结构地图复核。

**最终实现设计评分 10/10**：六项原则在本轮实际审查范围内均有具体实现依据，D4–D6 无残留必须修改项，首轮 8.5 结论已由上述实际整改取代。评分不声称架构永远不可改进，也不证明没有尚未发现的运行缺陷。

最终 Docker 冒烟、完整 E2E、全仓 Go/前端质量矩阵、MySQL/PostgreSQL 实跑仍待相应串行阶段；商业主备模型、真机软键盘/手动读屏、实际 bfcache restore 不能用 mock 或设计评分替代。本门完成并冻结，不修改主 manifest，不 commit/push。

### Gate07 最后 continuation 窗口的增量复审

仅复审本次小改动：实际读取 `stores/token.ts` 的 `RefreshedSession`、DesktopApp/MobileApp/Unlock 三个调用点及 `UnlockPageBase.auth.test.ts`。**设计评分维持 10/10，无新增必须修改项。**

`isCurrent()` 由返回值闭包捕获已提交的 owner/token，调用方只需在自身 `.then` 的实际副作用前检查一次；它没有暴露 token 比较规则、增加第二份身份状态或泛用 callback 框架。store 内检查无法覆盖 Promise 返回到 caller 的最后微任务窗口，因此将有效性查询随 bootstrap 结果返回有必要，并非重复浅包装。接口注释明确调用时机；三个调用点都在 locale、notification、缓存/汇率处理之前检查，正常路径保持。

新增实际 Unlock composable + token store 测试在 storeUserBasicInfo 后排入微任务，分别改变 owner/token 并断言旧 UI 副作用不执行，另验证当前 session 正常执行。父独立验证日志 `/tmp/ebk-parent-gate07-auth-final.log` 为3文件24项PASS，本轮实读日志，不声称重新运行，也不重复之前139项测试。仅更新本报告；本增量复审完成并冻结，后续 router/结构/最终矩阵按原安排执行。


### Gate15–17 增量复审：产品归类通过，harness 生命周期 D8 待修

实际读取15/16/17报告，以及 `assistant-experience-docker-runtime.mjs`、`build-assistant-docker-validation.mjs`、`assistant-experience-docker-smoke.mjs`、environment adapter、`assistant_migration_runtime.py` 与其真实子进程测试、migration CLI 的调用/清理边界、session/store 及4个 compatibility 回归。没有重跑已通过的全矩阵，也没有改生产或工具代码。

**产品设计仍为10/10。** capabilities404 现在由 `session.checkIdentity` 在当前 ticket/epoch/disposal 检查后归类，initialize 不再把下游 history404 当旧后端；前台同步和初始化共享同一协议知识，既不增加公共接口也不新造状态机。4个回归涵盖并发、恢复、过时404和无关404，正面覆盖问题边界。

验证工具的合理设计：runtime 的 start/stop/close 隐藏容器命名、copy/start、诊断与删除，未把 Docker 协议放到业务流程；close 在 uncertain create 时检查真实存在性，仅明确 absent 才放弃清理，失败保留退出兜底。environment 复用原业务API、启动失败也关闭全部自有资源。Python bounded_run 隐藏进程组和超时终止，isolated_environment 集中清除两种应用覆盖前缀，调用方无需复制环境过滤/kill 算法；180/600秒区分命令与编译是实际需求，不是泛化框架。

**D8（必须关闭，中）：构建/冒烟新子进程的生命周期没有与上述模块同样完整封装。**

- `build-assistant-docker-validation.mjs:run` 的 spawn 没有 deadline；其多处 `capture` 也没有 timeout。npm、Docker build/run/inspect 或registry/network阻塞可以使 finally 永远无法到达。
- builder `docker run --rm --name ebk-assistant-builder-${process.pid}` 没有在失败/CLI中断后显式删除该确切容器的路径，也没有 exit recovery。`--rm` 只在容器结束后删除，不等于 CLI 被终止时容器一定结束；当前 finally 仅删除临时源码目录。
- 新 smoke 的 `npm run test:production` spawn 和 sqlite3 execFile 同样无命令期限；若子进程挂住，外层 env.close 不会执行。Smoke 应继续拥有 env.close，但被等待的命令本身应有有界终止契约。

最小整改：为这些命令提供有界执行/失败清理（可复用一个真正拥有进程组终止的 helper，不要求泛用process框架）；builder使用任务唯一名字，在不确定创建、失败/超时/正常退出时只清理该名字，保留原始错误与清理失败。给新增资源边界增加定向回归：挂起命令到期、spawn失败、Docker CLI失败后cleanup、清理失败仍移除临时文件且可观测。不要改产品Dockerfile，也不要为工具修复重跑所有业务矩阵。

这是从实际控制流确认的异常路径缺口，不否定15–17已经完成的成功执行证据；那些成功结果不能证明挂起/中断路径。**新增harness纳入后当前总设计评分9.5/10，暂不将本项全部标通过；待D8的最小修改和定向证据后复审。** 父已收到明确交接。本轮只改01报告，冻结等待下一串行安排。

15–17记录的完整单元/Go/race、三库/性能/迁移、容器E2E47PASS/2NOTRUN、reference10及old-main compatibility2均保持原证据归属，不重复计数。stock release Docker、真实模型、物理输入/手动读屏、真实bfcache恢复等外部限制也没有因本设计复审被标通过。

### Gate18 最终增量复审：D8 关闭

实际读取 `assistant-bounded-command.mjs`、`assistant-build-resources.mjs`、两份对应测试、build/smoke/main E2E/environment 调用点及 [Gate18 报告](18-harness-finalization.md)。本轮仅复审 D8 修复与既有 compatibility 归属，没有修改产品或测试代码，也没有重新执行已通过矩阵。

- boundedCommand 拥有 POSIX 进程组、deadline、TERM/KILL 升级、捕获上限及监听器释放；调用方只提供预算、环境和输出去向。失败后停止收集输出，按字节截断到预算。正常结束、启动失败、超时和中断均有明确结束路径，属于具有实际资源知识的深模块，不是 spawn 的薄包装。
- buildResources 独立拥有随机确切容器名和临时目录，在 Docker 调用前登记名字；CLI 结束与 daemon 对象销毁分开处理。close 只接受明确 absent，未知清理错误保留并继续清目录，退出兜底保留。构建 finally 汇总原始错误与清理错误。没有将容器清理协议复制给每个调用点。
- 原 D8 的 build run/capture、smoke npm/sqlite 调用均已接入有界执行；main E2E production 与临时 Go build 同样接入。长期运行的应用服务继续由既有服务生命周期管理，未错误套入短命令 deadline。原 Docker runtime 没有被再次包装。
- 定向测试确实启动真实抗 TERM 子孙进程，检查 deadline/信号/调用方 finally、spawn 失败与监听器释放、持续输出上限；受控 Docker 可执行文件模拟不确定创建及 daemon 清理失败，断言确切名字、临时目录及可观察错误。测试针对资源边界和失败行为，不是逐行复刻实现。

证据归属：实读 Gate18 的 **5 files / 21 PASS**、三个 mutation 红→绿及真实 Docker 无 `--rm` 超时后对象仍存在、close 后消失的证据；后者验证了 D8 的关键 daemon/CLI 边界。其新构建与10项冒烟PASS、dist107文件及 Linux executable 与 Gate16 完全相等的哈希可保留已有业务矩阵证据。父独立全量 Vitest **44 files / 38,917 PASS**、1716文件零超限为父审结果，本门不声称重新运行或重复计数。

**D8 关闭，本门最终设计评分 10/10。** capabilities404 的 session.checkIdentity 归属保持此前通过结论。本评分限于已审的软件设计原则与实现边界，不保证零未知 bug，不将真实商业主备模型、物理输入/手动读屏、实际 bfcache restore、stock Dockerfile/Node26 构建阶段标为通过。上述外部限制继续保持 NOT RUN。仅更新本报告，冻结交父审；不修改主 manifest，不 commit/push。
