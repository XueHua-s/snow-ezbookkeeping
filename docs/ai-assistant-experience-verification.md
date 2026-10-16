# AI 助手体验计划独立验收记录

本报告保留提交 `91081998` 的基线验证记录。2026-10-16 开始的后续技能串行审查、修复及 Docker 验证见 [技能审计清单](ai-assistant-experience-skill-audit.md)；下文的通过结果不自动覆盖后续代码变更。

当前复核已有独立的 [容器 E2E 记录](ai-assistant-experience-audit/16-container-e2e.md) 和 [完整单元/三库矩阵](ai-assistant-experience-audit/17-final-test-matrix.md)：主 E2E 47 PASS / 2 bfcache NOT RUN、引用 10 PASS、旧后端兼容 2 PASS，[最终 Vitest 38,917 PASS](ai-assistant-experience-audit/18-harness-finalization.md)、Chromium 62 PASS，Go 全仓、race、三库性能及备份恢复/回滚均通过。重复测试模式不叠加为独立覆盖；14项适用技能本地终审均已闭环为10/10，构建工具整改完成；评分仅覆盖审计记录中的本地适用范围。商业模型、物理输入/读屏、真实 bfcache 与 stock Dockerfile 发布构建仍未验证。

核查日期：2026-10-15–16（Asia/Shanghai）。目标为根目录 `AI_ASSISTANT_EXPERIENCE_PLAN.md` v2.1 的全部要求。

基线状态（以下为原验收快照）：实现与隔离自动/生产验证通过，无未解决的已知实现FAIL；真实主备模型、物理软键盘/读屏及未命中的真实bfcache门槛仍未完成，最终整体验收不能标为全部通过；下面 NOT RUN 表示尚无覆盖该整项要求的执行证据，不能据此宣称完成或可发布。旧 harness 的通过结果仅作为旧链路基线，不替代新增功能验收。基线质量门在当时冻结的集成工作树运行，结果见下方原始记录；后续变化的证据见上述当前审计。

## 环境核查

- macOS arm64；Go 1.27.1；Node 22.20.0；node_modules 已安装。
- 没有发现仓库或其目录祖先中的 AGENTS.md。未修改业务代码。
- Docker daemon 可用，缓存 MySQL 8.0 / PostgreSQL 16 Alpine 镜像。初次核查时没有本任务数据库容器；随后建立下述两个任务专用容器。不得复用或删除不相关资源。
- SQLite 单元测试可运行；MySQL/PostgreSQL 已创建新隔离容器；执行测试时设置 `EBK_ASSISTANT_MYSQL_TEST_DSN` / `EBK_ASSISTANT_POSTGRES_TEST_DSN`。既有外部测试 helper 只接受 `ebk_harness_test_` bootstrap，创建唯一数据库并清理自身数据库。参见旧验证报告的启动步骤；新增助手表三库契约已有日志证据；旧账务查询测试本身不能证明新存储正确。
- 首次环境核查未发现 EBK / OPENAI / ANTHROPIC / GEMINI 相关配置变量。未发现已指定的隔离真实模型测试部署；没有读取或输出用户私有凭据。主备真实模型均 NOT RUN。
- Chromium 浏览器测试可执行；配置只启用 headless Chromium。不存在 iOS Safari / Android Chrome 物理设备或可靠设备环境的已验证证据。浏览器 390px viewport 不能单独证明软键盘体验。
- origin 是 `XueHua-s/ezbookkeeping`；upstream 是 `mayswind/ezbookkeeping`。gh 自动默认 upstream，最终 PR 必须明确 `-R XueHua-s/ezbookkeeping --base main`。提交前核查（2026-10-16）：HEAD已快进至origin/main的 `a5abfe62`，ahead/behind均为0，实施在工作区等待本次提交。当时该分支没有PR；交付目标为origin/main的draft PR，不代表已可发布。

## 已执行基线

| 命令 | 实际结果 | 覆盖边界 |
| --- | --- | --- |
| `npm run test -- src/views/base/assistant/AssistantPages.test.ts src/lib/assistant_stream.test.ts` | PASS，2 文件 63 tests，1.82s | 旧页面状态与旧 SSE；不是新持久化验收 |
| `BUILD_PIPELINE=1 go test ./pkg/services -run 'Assistant' -count=1` | PASS，3.041s | 现有账务服务；未启用外部数据库，不能记三库 PASS |
| `npm run test:browser` | PASS，1 文件 2 tests，3.83s | Chromium 实际平台组件；不是新功能、整应用路由或物理手机验收 |

## R01–R40 最终逐项证据矩阵（独立审查快照）

此矩阵区分“实现已落盘”“已有局部行为证据”“整项仍未验证”。**PARTIAL ≠ 整项 PASS；PASS（隔离验证）只表示该行契约的隔离证据充分，不表示整项目可发布。** 未跑子项明确标NOT RUN。记录来源是实际读取测试断言、运行日志、native 恢复产物；没有把 agent 声明、文件存在或被 skip 的子测试视为通过。最终自动质量门与production E2E已有通过证据；明确尚未执行的真实模型/真机/读屏/bfcache门槛仍保留，本报告不宣称全部要求通过或可发布。实现与局部自动检查不代替外部验收。

证据目录 `E=/tmp/ebk-experience-verification-20261015`：`t02-storage-tests.log`、`t03-conversations-tests.log`、`runs-race-three-databases.log`、`history-performance-97mib.log` 已逐项读取；迁移证据见本文后续及 JSON。前端单测断言在相关 `src/**/assistant*.test.ts`；此前 `npm-test-all.log` 38,760 项通过是较早快照，后续变更需 final 日志。已独立读取持久化 `docs/ai-assistant-experience-evidence/route-results-development.json`：当前工作区构建的真实应用两端各 15 项、共 30 项 PASS，providerCalls=18，供应商为确定性协议 fixture。覆盖详情返回/刷新保留草稿、双设备读取与空闲改名同步、原始 Markdown 剪贴板、离线不自动发送、导出、创建/忘记记忆、搜索/归档、丢失受理响应后刷新去重、清空不删账单。随后production套件33项PASS/2项NOT RUN已落盘：两端SW真实控制、no-store、CacheStorage无助手响应、IndexedDB不落答案/token及离线不能重放API均PASS；真实pageshow.persisted两端均false，因此bfcache仍NOT RUN。

| ID | 状态 | 实现入口/机制 | 已核验行为与证据 | 覆盖边界/剩余门槛 |
| --- | --- | --- | --- | --- |
| R01 | PASS（隔离验证） | 共享 Store、平台路由与阅读锚点 | AssistantPages/browser 与 anchor 单测覆盖离页不 cancel、草稿/位置恢复 | 两端真实路由详情/返回/刷新 PASS；120轮真实API/DOM三次prepend原锚点漂移<0.5px、resize原锚点不变PASS；继续追问财务条件需真实模型 |
| R02 | PASS（隔离验证） | 服务器会话/消息、ownerScope、恢复 | HTTPContracts 保存/读回，三库 ConversationsContracts；真实 IndexedDB reopen 测试入口 | 双设备与刷新真实E2E PASS；跨tab换用户前一owner在下一帧前清除、原owner重开正式历史PASS；真实bfcache未命中，NOT RUN |
| R03 | PASS（隔离验证） | 历史/搜索/更新/归档/删除/导出 | 三库 ConversationsContracts、ExportBatches 与 1,000 条搜索分页实跑 | 两端搜索/归档/取消归档/导出/清空真实E2E PASS；单会话删除由真实store/API用例验证，loading/error/retry由共享组件与store回归补证 |
| R04 | PASS（隔离验证） | Composer、统一本地存储、发送状态 | Experience 单测精确 Unicode/byte/IME；Store 保留新草稿；IndexedDB 浏览器 fixture | 真实双tab草稿冲突与IndexedDB拒绝警告PASS；物理输入法/软键盘未验 |
| R05 | PARTIAL / NOT RUN子项 | 桌面限宽历史栏、手机 flex/safe-area | 组件浏览器布局与 viewport fixture | 真实Chromium横竖/小屏/landscape已有证据；iOS/Android可靠设备软键盘 NOT RUN |
| R06 | PASS（隔离验证） | 原始 Markdown/代码复制与fallback | Experience.test 精确快照、拒绝权限、代码缩进断言 | 两端真实原始Markdown剪贴板 PASS；完整格式与权限失败由组件精确断言补证 |
| R07 | PASS（隔离验证） | messageId/offset 锚点与回到最新 | Experience anchor prepend 单测、Store bounded-window 测试 | 120轮长历史三次prepend、viewport reflow、90条DOM上限、回到第120轮及详情返回PASS；有限场景覆盖对应契约，不声称穷尽所有浏览器资源组合 |
| R08 | PASS（隔离验证） | durable worker 状态/显式 cancel | runs-race-three-databases.log：取消竞争、应用 stop、空正文/超限、保存失败 | 运行状态组件fixtures与真实失败/重试/后台恢复共同覆盖；worker组合fixture修复已复验 |
| R09 | PASS（隔离验证） | Retry/版本选择、成功上下文链 | ConversationsContracts 与 Context chain；Store历史版本只读断言 | 两端真实provider失败与显式retry不重复question PASS；失败新版本后继续上下文由服务回归证明 |
| R10 | PASS（隔离验证） | Branch冻结前缀/版本指针 | ConversationByteBudgetAndCompactedAncestor；workspace edit不自动send | 真实分支刷新E2E、HTTP branch上下文断言及共享hook编辑状态共同覆盖 |
| R11 | PASS（隔离验证） | 命令确认/订阅退避/同run polling | Store unknown outcome、401、三次流失败回退，HTTP断流重连 | 两端真实已受理响应丢失/刷新原命令去重与离线恢复不发模型PASS；401单元回归通过；真实bfcache未命中，NOT RUN |
| R12 | PASS（隔离验证） | 统一command receipt、CAS、epoch | 三库 StorageContracts 重放/hash冲突/并发quota/过期旧请求；HTTP同ID重放 | 真实HTTP middleware暂停未受理send→clear→释放返回409、admission/ID/object未增加；race重复5次PASS |
| R13 | PASS（隔离验证） | 只读详情与历史查询时间 | 引用投影保存 queriedAt，平台详情入口实现 | 共享preflight/native dialog修复后真实production两端10场景PASS：改/删/foreign/network明确反馈，16秒poll不dismiss，draft/reading/原snapshot queriedAt不变，显式重查新金额23456 |
| R14 | PASS（隔离验证） | minor decimal string、Decimal显示 | 快照int64 extrema、Experience安全number fallback/隐藏金额；旧三库独立账务oracle | 独立账务oracle、int64 DTO、Decimal投影与真实账单引用视图共同验证 |
| R15 | PASS（隔离验证） | sourceMessageId绑定、restart query | BranchQueryRestartAuthorization、workspace next_page/restart 单测 | 11组真实HTTP含查询分支/过期重查与共享hook显式source绑定共同覆盖 |
| R16 | PARTIAL / NOT RUN子项 | 财务summary独立mode/期间确认 | 旧财务回归与workspace期间确认不覆盖草稿 | 主备真实模型四部分/日期/财务oracle NOT RUN |
| R17 | PARTIAL / NOT RUN子项 | 服务端摘要/CAS/保留原文 | ContextThirtyTurnsTwoCompactions、Prepare32轮自动压缩、CAS回归 | 主备真实模型长对话条件保留 NOT RUN |
| R18 | PARTIAL / NOT RUN子项 | 长期记忆管理/候选/来源/会话off | MemoryConfirmForget、Candidates、BackgroundIdleBudget；三库memory日志 | 两端实际管理UI创建/确认遗忘PASS；实际主备抽取/矛盾/过期目标语义NOT RUN |
| R19 | PASS（隔离验证） | generation/use epoch与来源失效 | 三库StorageRevocation、MemoryBackground disable fence、来源off跨run回归 | 受控provider整理屏障验证撤销；真实模型语义另属R17/R18 |
| R20 | PASS（隔离验证） | 原子预留/结算/固定投影/接受快照 | 三库并发quota与一次settlement、RunAcceptancePolicy、ByteBudget；容量fixture | 原子服务边界、容量UI fixture与明确续聊hook验证；不要求所有容量×主题笛卡尔积 |
| R21 | PASS（隔离验证） | 输入/工具/输出保守预算与主备window | ConservativeWindow、ActualWireFits、OversizedContext-before-model | 预算修复后11组真实HTTP含page8动态缩页/branch continuation PASS；主备真实模型NOT RUN |
| R22 | PASS（隔离验证） | UID绑定全对象读取与写命令 | HTTP他人get/export与SSE 404；三库foreign消息/branch/memory/receipt回归 | 两端真实跨tab换身份下一帧前清旧正文PASS；真实bfcache未命中，NOT RUN |
| R23 | PASS（隔离验证） | lease owner fencing、过期中断、不重放 | 三库run race+native restore真实worker过期恢复、provider eligibility零调用 | 双执行者fencing及真实恢复后进程扫描覆盖租约恢复；不声称任意网络分区恰好一次计费 |
| R24 | PASS（隔离验证） | snapshot revision/schema/UTF8限帧 | TS解析20项；真HTTP completed重连/断流；Store拒旧rev/终态回退 | 真实HTTP流与TS parser/store重连断言共同验证，不要求所有畸形帧跨全部viewport重放 |
| R25 | PASS（隔离验证） | 五表升级/清理/native backup restore | migration evidence逐字段哈希、旧binary维护；三库Lifecycle清理与ExportBatches | 最新operation/订阅schema变更后已重跑三库PASS；全量export由流式service/真实HTTP验证，当前会话两端导出由production E2E验证 |
| R26 | PASS（隔离验证） | 安全Markdown/SQL日志脱敏/只读权限 | Experience script/img/svg/data链接断言；三库logging canary；HTTP拒uid/history注入 | 安全Markdown负面断言、传输边界、SQL canary与HTTP严格字段共同覆盖 |
| R27 | PARTIAL / NOT RUN子项 | label/focus/RTL/touch/theme | Experience浏览器RTL/键盘fixture，semantic controls实现 | 实际dark/RTL/narrow/200%browser zoom及focus已PASS；各语言fallback有检查；物理读屏NOT RUN |
| R28 | PARTIAL / NOT RUN子项 | keyset/byte预算/stream export/有界缓存 | 97.78MiB三库fixture：首屏p95 43/65/65ms、搜索602/262/79ms、export堆<3.7MB | 数据为service非HTTP；120轮浏览器DOM保持90条与最新窗口恢复PASS；90条DOM与有界Store提供客户端容量证据，export heap实测有界；真实模型延迟/tokens未验，不额外要求无限时长cross-product |
| R29 | PASS（隔离验证） | history与generation能力分离 | HTTP generationAvailable=false仍history=true；AccountFence及UI禁生成单测 | 真实后端关闭generation重启，两端history/draft/export保留且provider零新调用PASS |
| R30 | PASS（隔离验证） | 旧协议保留/能力探测/旧binary回滚 | 旧harness回归；native迁移旧binary不改五表/legacy行 | 最终dist→隔离old-main两端capabilities404/升级提示、5秒无循环、0provider PASS，见compat-results.json |
| R31 | PASS（隔离验证） | 服务端authority/薄平台/安全终态 | HTTP拒history/state字段、WorkerStorageFailure、corrupt snapshot测试；父/独立结构审查 | worker保存失败不发假completed、projection损坏拒绝、共享store UI错误反馈已验证 |
| R32 | PARTIAL / NOT RUN子项 | 最终全质量门与真实环境 | go-test-all与npm-test-all初轮通过；各独立三库日志已读 | 早期race/worker失败已有修复与定向回归；final-go-test/vet/race已PASS（含archive修正）；lint与browser33项已PASS；build已PASS（58.39s，SW171ms）；production33项PASS/2项bfcache NOT RUN；真实主备模型NOT RUN |
| R33 | PASS（隔离验证） | 所有写命令共用回执/epoch/global clear fence | StorageOperations+ClearAllDataReplayNeverReentersLedgerServices；三库bulk清空 | 真实HTTP旧send屏障拒绝、fresh send后旧clear回执不删新正文/epoch保持1；race5次PASS |
| R34 | PASS（隔离验证） | IndexedDB原命令/冲突/容量/拒绝状态 | local_storage.browser 事务quota、reopen、denied、canonical ID断言；Store未知branch保留 | 两端真实未知branch刷新恢复、共享IDB冲突与拒绝PASS；配额上限浏览器事务测试与真实拒绝存储UI共同验证，不静默淘汰原文 |
| R35 | PASS（隔离验证） | 15s/30s同步、身份/离线/路由 | Store offline不send/旧回调、local_storage stale-tab；实际E2E两端离线与双设备改名同步PASS | 双设备空闲rename+独立context远端delete清DOM/draft、双tab冲突及offline PASS；真实bfcache未命中，NOT RUN |
| R36 | PASS（隔离验证） | 已用memory/source撤销与提交fence | MemoryConfirmForgetRunningRevocation、来源off跨run、Run dependency/epoch测试 | 受控provider提交屏障、跨来源撤销测试及两端实际管理UI共同验证；真实主备语义另属R18 |
| R37 | PARTIAL / NOT RUN子项 | no-store/SW排除/身份清理 | HTTP每response no-store；cache-policy URL与清理单测；Store401隐藏 | 生产SW/CacheStorage/IDB/离线缓存边界PASS；两端真实bfcache未命中（pageshow.persisted=false），该项NOT RUN |
| R38 | PASS（隔离验证） | 限帧/DB订阅租约/写超时/429/polling | 真TCP慢读者断开、HTTP429 Retry-After、三库订阅并发租约、同run polling | 真实nginx四配置/超时断流/同run重连PASS（2次模型调用、1命令）；订阅并发配额与慢socket写截止另有回归；覆盖范围见proxy报告 |
| R39 | PASS（隔离验证） | 两命令分支/冻结前缀/有界seed | CompactedAncestor、BranchQueryRestart、Store未知branch复用/保留文本 | 两端真实lost branch acceptance+刷新恢复单目标/编辑草稿且仅显式send调用provider PASS；store真实first-send失败也保留目标 |
| R40 | PASS（隔离验证） | 两端状态fixture/语义操作和焦点 | 真实两端6图与review.json已审阅；具体范围见ai-assistant-ui-review.md | 真实两端6张状态图及focus/clipboard review已PASS；实际dark/RTL/narrow/真正200%browser zoom回答场景PASS；568×320 landscape修复真实复跑PASS（消息130px、textarea44px）；真实设备/读屏未验 |

## 实际执行安排与必留产物

1. 三库：同一新会话服务测试集分别连接 SQLite、隔离 MySQL 与隔离 PostgreSQL，保留保存/CAS/跨会话原子配额/幂等/来源删除/任务认领结果。数据库升级前后保存账本与 embedding 行哈希，演练备份恢复与旧二进制回滚。
2. 真实 HTTP：启动隔离应用与可控 provider，针对受理前后断线、双执行者、租约过期、保存失败、取消竞争、代理缓冲、慢读者/订阅限额测试。记录 provider 请求次数，证明重连不会重新调用模型。
3. 真实模型：按 `docs/ai-assistant-harness-verification.md` 的 `TestAIDeploymentIntegration` opt-in 设置隔离目录、配置和 UID；主模型与 `EBK_AI_DEPLOYMENT_FORCE_FALLBACK=1` 分别执行。旧 smoke 只证明两个工具和协议，新方案还需独立数值 oracle、30 轮/至少两次压缩及跨会话记忆 fixture；财务/记忆正确性不得让被测模型自行评分。
4. 浏览器：两用户、两 context 与双 tab，在真实应用路由覆盖详情往返、刷新/关页重开、前进后退/深链、离线恢复、IndexedDB 故障/容量/冲突与 CacheStorage/bfcache。冻结 provider 交错时序并检查没有自动新发请求。
5. 性能：生成隔离 1,000 段、约 100 MiB 总历史、单段 200 轮/接近 2 MiB/多个版本 fixture。测首屏 p95（目标 ≤300ms）、搜索 p95（目标 ≤1s）、响应条数/字节、DOM 挂载数、连续浏览内存、订阅及导出内存上界；记录硬件、样本数和负载。
6. 视觉：空态、正常回答、表格/账单/依据、失败重试、历史搜索、记忆管理、容量续聊 × 桌面/手机 × 浅色/深色；尺寸、缩放和截图路径齐全。独立 reviewer 检查操作顺序、44px 触控、对比度、焦点/Esc、RTL、读屏及设备键盘遮挡。

## 集成后质量门

以下全部质量门已有R13修复后最终一轮通过证据，早期失败及修复记录保留在历史章节。R13最后修复专项10场景及修复后的全production套件33场景均已通过，另2项真实bfcache未命中保持NOT RUN；不把修复前通过等同于修复后通过。

```sh
BUILD_PIPELINE=1 go test ./...
BUILD_PIPELINE=1 go vet ./...
BUILD_PIPELINE=1 go test -race ./pkg/api ./pkg/services ./pkg/llm/... ./pkg/datastore ./pkg/core ./pkg/settings
npm run lint
npm run test
npm run test:browser
npm run build
npm run test:production
git diff --check
```

修改的 Go 文件须 gofmt；lint 含 autofix，核对没有无关修改；source-lines 必须满足 500 行。最终文档应同步 providers、storage migration、前端功能说明及两套 skill/reference，并明确旧报告覆盖边界。

## 当前缺口与交付边界

- R13原实现缺口已修复：真实两端专项10场景PASS；同一最后build的全production套件33场景PASS，2项真实bfcache未命中为NOT RUN。

- 新功能的 40 项执行证据仍未完成；基线测试通过不表示实施完成。
- 主备真实模型隔离配置/专用 UID 尚未给定，真实模型质量门 NOT RUN。
- iOS Safari / Android Chrome 可靠设备环境、读屏与软键盘证据尚缺。
- 三库新表、native 备份回滚及 97.78 MiB 性能已有证据；最新schema迁移重跑与最终视觉/自动质量门已PASS；外部门槛仍未完成。

交付约定：创建明确指向origin/main的draft PR；本文记录提交前验收快照。R13修复及真实专项新日志已独立复核；真实模型、物理软键盘/读屏、真实bfcache未命中继续保留，不能在draft描述中宣称全部要求已通过。


## 本任务隔离数据库运行记录

2026-10-15 新建，`docker run -p 127.0.0.1::PORT` 由 Docker 分配本机端口：

| 容器 | 镜像 | localhost 端口 | Readiness |
| --- | --- | --- | --- |
| `ebk-experience-verify-mysql-20261015` | `mysql:8.0` | `32865` | `mysqladmin ping` → mysqld is alive |
| `ebk-experience-verify-pg-20261015` | `postgres:16-alpine` | `32866` | `pg_isready` → accepting connections |

bootstrap 数据库均为 `ebk_harness_test_bootstrap`，仅含测试数据。两容器已由父agent于2026-10-16完成测试后删除；独立复查docker ps -a对应名称为空。以下端口仅为当次证据，不是当前可连接环境。连接凭据由协调 agent 传递；本文不保存密码或现有 secrets。复现时创建自己的新容器/公开测试密码并设置：

```sh
export EBK_ASSISTANT_DB_TEST=1
export EBK_ASSISTANT_MYSQL_TEST_DSN='root:<test-password>@tcp(127.0.0.1:32865)/ebk_harness_test_bootstrap?parseTime=true'
export EBK_ASSISTANT_POSTGRES_TEST_DSN='postgres://postgres:<test-password>@127.0.0.1:32866/ebk_harness_test_bootstrap?sslmode=disable'
BUILD_PIPELINE=1 go test -race ./pkg/datastore ./pkg/services -run 'TestAssistant|TestXOrmContextAdapter' -count=1 -v
```

运行日志：`/tmp/ebk-experience-verification-20261015/three-driver-baseline.log`。实际 exit 0：datastore 2.104s，services 59.096s。三库 `TestAssistantDatabaseContracts`、外部 QueryFoundation/Transactions 与读视图/取消检查均执行通过；唯一 skip 是独立 opt-in 的 `TestAssistantDatabase100KPlans`，随后以 `EBK_ASSISTANT_PERF_TEST=1` 单独运行该测试，三库均 PASS、无 skip，exit 0（总 26.752s；SQLite 4.75s、MySQL 12.80s、PostgreSQL 8.78s）。日志为 `/tmp/ebk-experience-verification-20261015/three-driver-100k.log`；实际版本 SQLite 3.53.4、MySQL 8.0.46、PostgreSQL 16.15。该 100,000 条账单 fixture 不是 R28 要求的 100 MiB 聊天历史性能 fixture。上述旧测试证明隔离驱动/账本环境，不证明新增助手表和并发契约。

已执行以下仅针对本任务的资源清理（父agent使用docker rm -f；下列含-v为复现环境可选清理示例）：

```sh
docker rm -fv ebk-experience-verify-mysql-20261015 ebk-experience-verify-pg-20261015
```

### 远端合并提交核对

首次核查时：`origin/main` 为 `a5abfe6210fe73aff3678c7692000ea08f40ca06`，父提交 `77c00d1c` 与当前基线 `1a21bc9b`，提交信息是合并旧 harness PR #1。`HEAD^{tree}` 与 `origin/main^{tree}` 完全相同：`40163817e3ab1f58de3635dd92f9f7b80d61fb41`；`git diff --stat HEAD origin/main` 为空。协调 agent 确认无并发 Git 操作后可执行 `git merge --ff-only origin/main`，只前移分支指针；验证 agent 未执行 merge。父agent随后已快进；2026-10-16复查HEAD与origin/main均为a5abfe62，此段仅保留原始基线审查记录，不是待执行merge动作。


## 第一轮独立服务审查（WIP，尚非验收通过）

范围：R12/R20/R22/R33/R36/R39 的命令、配额、依赖、分支及上下文事务。审查看到统一 UID CAS 事务、先回执再 epoch/时间验证、按 UID 读取对象与一次结算机制；这些结构支持要求，但不能代替完整端到端竞态证据。

发现并已交给 owner 的具体缺口：

| 发现 | 路径与复现条件 | 状态 |
| --- | --- | --- |
| 可重试事务污染外部上下文 | `ai_assistant_memories_management.go` selectContext 在 callback 中累加外部 text/dependencies；commit 重试重复正文，重试间关闭记忆会保留回滚那次选择。`ai_assistant_context_history.go` snapshot.Frozen 也在重试中累加；compaction commit 前替换 snap.Conversation 会改变后续 CAS 预期。 | 已修复并由后续代码复查/定向回归确认，见下文 |
| 二级分支丢祖先 seed | A→B 首轮 compact 后 B.FrozenPrefixJSON 清空且 SeedContext 有值；Branch(B→C) 只继承 FrozenPrefix/Sources，不继承 B.SeedContext，丢 A 的条件。 | 已修复并由后续代码复查/定向回归确认，见下文 |
| 非统一头部计量 | compaction 使用较小 struct 调用 SettleHeader，而会话服务使用完整 DTO+摘要/来源投影，可能少计元数据。 | 已修复并由后续代码复查/定向回归确认，见下文 |
| 已受理预算读取实时配置 | context.Prepare 读取 live runtime/limits，没有使用该 run 已保存预算快照；管理员调小配置会影响已预留请求。 | 已修复并由后续代码复查/定向回归确认，见下文 |
| 来源会话关闭记忆的运行中使用边界 | 来源会话改 off 后新选择拒绝其 memory，但其他会话运行中使用该 memory 的校验未检查来源会话 mode；需明确并覆盖撤销规则，不能用提升 SourceVersion 误伤独立分支 seed。 | 已修复并由后续代码复查/定向回归确认，见下文 |

新增运行测试尝试：`BUILD_PIPELINE=1 go test ./pkg/services -run '^TestAssistantRun' -count=1` 实际 exit 1（编译阶段）；`ai_assistant_cleanup_users.go:42` 引用不存在的 `models.AssistantReasonGenerationUnavailable`。这是共享 WIP 的当次结果，需修复后重跑；不将 owner 口头 PASS 作为独立运行证据。存储 owner 报告新增三库契约通过，最终需复查其日志与断言范围。


## 三库完整备份、恢复与旧版本回滚演练

2026-10-15 实际执行通过，脚本总 exit 0，SQLite/MySQL/PostgreSQL 三项 PASS。可复跑入口为 `scripts/assistant-experience-migration-check.py` 与同名 Go helper 目录。未切换共享分支；旧版本由 `git archive 1a21bc9bfd2b155a3ed9b8ac57caff7c76bbb293` 在独立临时目录编译。当前版本由当时共享工作区编译（WIP，最终提交改变相关实现后需复跑）。

执行命令：

```sh
# EBK_MIGRATION_TEST_PASSWORD 仅设置本任务公开测试密码，不使用部署密钥。
python3 scripts/assistant-experience-migration-check.py \
  --mysql-container ebk-experience-verify-mysql-20261015 \
  --postgres-container ebk-experience-verify-pg-20261015 \
  --mysql-port 32865 --postgres-port 32866
```

每库执行：旧二进制真实 `database update` 建表 → 写入旧账单/embedding sentinel → 当前二进制升级两次 → 验证旧行全字段不变 → 写入全部五类助手记录 → 原生整库备份/恢复到另一新数据库 → 比较全部助手与旧 sentinel 的规范 JSON/哈希 → 旧二进制执行维护 → 再次比较五类新记录及旧记录不变 → 新二进制重新升级 → 再次比较不变 → 启动真实 worker 扫描，过期 running 变 interrupted、已保存正文不丢、没有调用 generation eligibility/provider → 服务拒绝已过期游标。SQLite 使用 `.backup/.restore`，MySQL `mysqldump --single-transaction`，PostgreSQL `pg_dump -Fc/pg_restore --exit-on-error`。不是仅统计表是否存在。

固定合成账单金额为 `9007199254740993`，embedding 内容为 `[0.125,-0.25]`；三库升级、恢复、旧版维护与重升级后均逐字段/哈希不变。助手 fixture 包含两个会话、四条消息（运行中片段与完成回答/过期游标）、记忆、用户额度状态和命令回执。恢复后的 worker 会更改预期的会话/消息/用户计数，记忆、回执、账单与 embedding 仍逐字段不变。

持久记录：[migration evidence](ai-assistant-experience-migration-evidence.json) 保存各阶段各记录类型的 SHA-256/行数、备份哈希、恢复输出和命令日志哈希。完整 commands/stdout、native backups、合成完整 JSON 与二进制保存在 `/var/folders/pf/q768gkkd2zb8rl5rtp6w0nw40000gn/T/ebk_experience_migration_iap5gtsd`。目录配置只含本任务测试凭据，权限 0600，不提交配置或二进制。测试自建数据库名称有唯一 `ebk_experience_migration_` 前缀，完整临时证据目录保留；测试数据库随专用容器删除，不再占用运行环境。

边界：此项证明 R25/R30 的隔离数据库迁移/备份/旧版维护兼容和 R23 的过期运行恢复子场景；不证明全部生产部署环境、浏览器混合版本窗口或所有 R25 清理/导出行为，因此总矩阵仍不能将整行标为 PASS。首次工具尝试因本地 Python tarfile filter 不支持、配置模板/存储目录不完整失败，修复验证辅助工具后完整重跑，最终通过不是忽略失败。


## 服务审查修复复核

上述第一轮发现保留作为审查历史，当前已只读确认：snapshot/selectContext 重试重置临时结果；compaction 事务成功后发布快照并复用统一 header 计量；二级分支继承有界 SeedContext；来源会话 off 有选择性撤销回归。随后独立实跑：

- `BUILD_PIPELINE=1 go test ./pkg/services -run '^TestAssistantRun' -count=1` → exit 0，0.577s。
- `BUILD_PIPELINE=1 go test ./pkg/services -run '^TestAssistant(Context|Memory|Prepare|ConversationMemory)' -count=1` → exit 0，0.571s。

当次未设置外部 DSN，因此仅将 SQLite/确定性上下文结果记录为通过，不能记三库验证。先前 `GenerationUnavailable` 编译失败已修复。测试含 30 轮两次压缩及已受理 policy 的确定性回归，仍不等于真实主备模型质量认证。

R40 之外仍缺父 agent 汇总与独立确认的证据：真实主备财务 oracle/长对话/长期记忆；完整浏览器存储/身份恢复矩阵；最终集成提交全部质量门。97.78 MiB/1,000 会话性能、真实HTTP慢订阅/429与Nginx代理调用计数已有后续证据，不再作为未执行项。40 行矩阵须按完整覆盖范围逐项更新，不能把这些局部通过泛化为整项通过。


### 最新独立组合回归阻断

`E/independent-final-contracts.log` 实际 exit 1；API 子集通过，但 services 中 `TestAssistantWorkerRealRunnerOutputAndEmpty` 的 complete/empty/limit 均保留 running，cancel 用例保留 cancel_requested，运行日志终态写入返回 storage_unavailable。已交给 runs 与父 agent 调查。此前单项和三库历史 PASS 不覆盖此次组合失败，该次失败后来定位为手建worker fixture缺少会话头部预留；修复后原组合通过、真实Send→worker终态九组三库通过、services三库race重复两次30.018s通过，已撤销该worker阻断。重复组合中另一个API deadline错误类型竞态由API owner修复，最终全量门槛仍需重跑。

最新模型字段（包含 OperationCompleted 与 SubscriptionLeasesJSON）落盘后，迁移工具再次完整三库运行 exit 0，证据 JSON 已更新到 iap5gtsd 目录。


## 可审查的持久化证据摘要

[verification-summary.json](ai-assistant-experience-verification-summary.json) 保留实际日志哈希和关键摘录、真实nginx结果以及独立UI断言，不要求读者访问临时目录才能审查结论。完整native迁移摘要另见 [migration-evidence.json](ai-assistant-experience-migration-evidence.json)。

- 真实Nginx 1.28.3验证已通过，包含buffering off/on/忽略X-Accel-Buffering、实际1秒idle断流与同run状态恢复/终态重连。仅1次发送、2次provider调用，重连无额外调用。见[proxy报告](ai-assistant-proxy-verification.md)。不再将真实代理记为NOT RUN，但不推论所有CDN或多用户压力通过。
- 独立真实UI review已完成两端普通回答、历史抽屉、记忆6张图以及18次Tab焦点约束、Esc恢复、Markdown/代码剪贴板精确比对。见[UI报告](ai-assistant-ui-review.md)。仍保留实际路由全部dark/RTL/zoom/故障状态与真机/读屏边界。
- 最新三库97.78MiB fixture在导出期限修复后重跑PASS（113.525s）：首屏p95 SQLite4.64/MySQL179.70/Postgres251.01ms；搜索64.81/500.25/112.56ms；96,841,192字节流式导出0.593/4.112/7.523秒，峰值堆增长<2.8MB。测量仍为service边界。
- 父agent已执行final lint/test/vet；已读取lint零超500行、test38,771项通过的日志。后续R39/预算修复需要对应最终重跑，不把旧日志当作最终提交证明。


### 最终整合进度（新增证据，覆盖上文较早快照）

已直接读取 final-go-test.log / final-go-race.log 全包结果，均通过；final-go-vet 由父进程确认exit0；final-npm-test 38,785项PASS。早期浏览器日志曾出现连接关闭错误，R13原生dialog/独立watch修复后重跑3文件33项全部PASS（10.34s），不再保留该阻断。最终lint检查1662文件零超500行通过。最终build已exit0（58.39s、SW171ms），hash已收录summary；production33项PASS，SW/CacheStorage/IDB有真实证据；2项实际bfcache因未命中保持NOT RUN。

持久化 [store-history.json](ai-assistant-experience-evidence/store-history.json) 证明120轮API/DOM分页、三次prepend同ID锚点漂移0.406/0.406/0.063px、viewport resize同ID零偏移、90条DOM上限和回第120轮；两个独立context远端delete后ID/正文/草稿全部清理。该fixture不是120次模型调用。

[route-results-development.json](ai-assistant-experience-evidence/route-results-development.json) 30项PASS覆盖双tab真实IndexedDB冲突/拒绝、generation禁用重启、失败retry、丢失分支受理后刷新恢复、换身份下一帧前清旧正文等。development轮早于include-archived inclusive修正；最终production脚本显式断言取消归档后仍在Include archived列表，该轮已PASS。真实主备、物理手机键盘、读屏保持NOT RUN。


### 最终production结果

[route-results-production.json](ai-assistant-experience-evidence/route-results-production.json)：33 PASS、2 NOT RUN、18 provider calls；实际执行 `npm run test:production` 两端生产bundle无runtime/network错误。SW控制器确实存在，CacheStorage中助手响应数为0，HTTP no-store，IndexedDB仅约定本地草稿/待确认记录且无回答/token，离线请求不能重放云历史。已验证inclusive archived最终行为。

两端back/forward navigation均未进入浏览器bfcache，实际pageshow.persisted=false；普通history恢复与跨tab身份清理通过，但不得将其记为真实bfcache通过。真实模型、物理软键盘、读屏与真实bfcache仍为明确未验门槛，区别于不要求穷尽的全部主题×错误×设备组合。


### 最终补充闭环

R12/R33：独立检查 `TestAssistantExperienceHTTPDelayedSendCannotCrossClearEpoch` 的真实HTTP middleware屏障：请求在解码/受理/ID分配前暂停，clear提升epoch后释放，返回409 data_epoch_conflict且没有生成admission/ID/对象；之后fresh send成功，再重放旧clear仍保留新正文与epoch1。`http-epoch-barrier-race.log`重复5次PASS，1.539s。

R30：[compat-results.json](ai-assistant-experience-evidence/compat-results.json) 两端最终dist连接真实旧main基线，capabilities返回404、显示明确升级提示、观察5秒不循环刷新且providerCalls=0。旧二进制来自独立git archive构建，未切共享分支。


### 2026-10-16 R13修复后的质量门快照

独立重新读取原名覆盖的最终日志：Go全量test/vet/race通过（含R13 DTO）；npm test为27文件38,785项PASS；lint检查1662文件零超500行；browser为3文件33项PASS、10.34s；build为58.39s、SW171ms且exit0。summary已刷新这7份日志SHA-256/摘录。此轮没有由验证agent再启动测试或构建。

R13修复已包含共享preflight、时间戳/同秒字段比较、明确changed/missing/access反馈、独立身份watch和不影响布局的原生dialog。前一真实场景发现的轮询误dismiss及notice挤占anchor已修；该历史记录写入时双端最终生产场景尚在验证中；此处为专项执行前的历史状态；随后R13真实两端10场景与最终全production 33PASS/2NOT RUN均已完成，结论由后续实际产物更新，不由单测或lint替代。

两个专用数据库容器已清理，报告中的DSN/端口只是历史复现记录；没有持续运行的本任务三库环境。本文为draft PR提交前证据快照，PR状态以GitHub为准。


### R13闭环（2026-10-16）

已独立阅读 `scripts/assistant-reference-e2e.mjs` 的实际断言及持久化 [reference-results-production.json](ai-assistant-experience-evidence/reference-results-production.json)：两端共10项PASS、8次provider calls。真实修改交易后标注changed，显式查看当前账单显示新comment/23456金额；桌面等16秒能力轮询后dialog仍在。真实删除与foreign owner真实HTTP拒绝统一不可访问，网络abort另有retry文案；所有preflight保持草稿/reading且不调用模型。显式Ask again生成第2版本，完整原版本JSON（含queriedAt）deepEqual不变，新选回答reference金额23456且草稿保留。原生dialog避免挤占消息布局。R13 FAIL已关闭，不再列为发布阻断。


## 最终冻结记录（2026-10-16）

最后一次完整production运行已exit0，持久化route JSON再次核对为33 PASS、2 NOT RUN、18次provider调用；两端真实bfcache仍persisted=false。literal `npm run test:production`最后日志两端均无runtime/network错误，SHA-256已更新summary。R13专项另为10 PASS、8次provider调用。所有验证应用/浏览器服务及专用数据库容器均已结束/清理，证据文件保留。

旧后端compat两端PASS是R13修复前最终dist的独立兼容证据；R13新增的是账单preflight流程，不将该旧compat记录冒充修复后重跑。无当前未解决实现FAIL；明确未完成门槛为真实主/备模型财务与记忆oracle（含真实延迟/tokens成本）、物理手机软键盘、手工读屏、真实bfcache恢复。计划交付仍为draft PR，不宣称可发布或40项无条件全部通过。
