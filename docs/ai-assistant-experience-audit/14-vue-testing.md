# 14 · Vue testing 串行审查

2026-10-16；审查 `91081998` 相对 `origin/main` 的 assistant experience PR，并保留 Gate01–13 工作树修改。初始 **6.5/10**，本门首次冻结时修复后 **9/10**、设计简化 **9/10**；最终证据闭环后的当前本地适用评分及设计简化均为 **10/10**（范围见文末，历史评分保留）。本门未委派、commit/push、修改主 manifest、操作 Docker 或读取真实用户数据/密钥。真实应用验证均用临时 SQLite、任务自建 Go 服务、确定性 provider 和合成账号。

## 实读与测试边界

实读 vue-testing-best-practices 及 blackbox、async、browser-vs-node、Pinia、Teleport、E2E 参考；vue-best-practices 与四个必读参考；vitest、project-spec/global-prompt、Vue Router cleanup 参考、主 manifest 和 Gate05–13 报告。按可观察行为区分 Node/组件、Chromium 与真实应用验证，没有把 happy-dom 当作浏览器布局，也没有把模拟 storage 事件称为原生跨标签事件。

组件测试显式创建和 dispose Pinia、卸载 wrapper。既有确认、导出、读序和 workspace lifecycle 回归保留；新增测试关注身份边界、控件实际可用和后端数据，而非仅匹配内部函数调用。

## 证实缺陷与修复

### Framework7 连续身份切换使历史入口滑出屏幕

实际移动端应用在连续 owner 切换、会话替换期间出现两个 `.page-current`，两者均保持约 -78px transform，历史按钮在 viewport 外。截图 `.vitest/gate14-red/mobile-identity-failure.png` 已目视检查，原始日志 `/tmp/gate14-mobile-red.log`、`/tmp/gate14-identity-ui.log`。

内部 replace 且当前路径已是 `/assistant` 时，沿用现有 Framework7 页面，通过 `updateCurrentUrl` 加 `refreshRoute` 更新选中会话。用户普通历史导航仍用原动画；same-target 判断严格区分 undefined（恢复最近会话）和空字符串（新会话），没有引入额外导航状态机。

浏览器回归运行真实 Framework7，两个 owner 使用不同 conversation ID，mock get 对其他 owner ID 返回404。它通过受控模拟身份事件复现时序，断言选中会话正确、只有一个当前页面、历史控件位于视口中；此项不是原生 storage 证据。临时恢复旧 adapter 后该回归实际失败，再恢复最终修复通过。

### 多标签刷新令牌导致未处理拒绝和队列挂起

原生双标签重复运行暴露间歇性 `pageerror Object`：同一旧 token 的并发 refresh 一次200、另一次401/202002。旧 Promise 包装没有向外 reject，阻塞请求也不释放。没有通过忽略 pageerror 或 mock refresh 隐藏问题。

transport 使用单一 refresh Promise 合并同标签请求，并在成功/失败时释放或拒绝队列。仅 refresh 在同 owner token 已改变时最多尝试3次，普通记账/助手 POST 不自动重放；旧 Authorization 的晚到401不清除新 token，当前 token 的401仍按既有规则退出。晚到成功只有 owner/token 仍匹配才调用现有 app-lock-aware `updateCurrentToken`。store 合并整个 bootstrap，检查 transport→store continuation 的 owner/token，再应用用户设置；best-effort revoke 和应用启动/解锁的拒绝均有处理。

Axios 的异步 request interceptor 还有独立微任务窗口。refresh/revoke 捕获严格 `expectedToken`，身份改变时拒绝派发。助手 shared get/post/export 捕获 `expectedOwner`，避免旧 payload 借用新 owner token，同时允许正常同 owner refresh 后的排队请求使用新 token。SSE fetch 原本同步捕获 token，无需另加适配层。没有重构全仓其他 API。

真实 Axios adapter 回归覆盖失败后队列恢复、同标签合并、旧401/当前401、同 owner 晚到成功、跨 owner 晚到成功/失败、logout、重试上限、普通 POST 不重试、refresh/revoke/assistant 的 interceptor 前微任务切换，以及助手正常刷新队列。服务 mock 契约同步补充 owner，未用 mock 替代这些 transport 回归。

## 新增真实应用身份流程

`scripts/assistant-experience-e2e-identity.mjs` 导出的 `identityFlows` 已集成主 E2E；独立入口 `scripts/assistant-experience-identity-review.mjs` 默认顺序跑 desktop/mobile。第二个真实应用 tab 修改 localStorage，接收 tab 断言 `StorageEvent.isTrusted === true`，采集的是该接收事件安排的**第一个 RAF**，不是两帧后才检查。

每端7项：

- 默认生产 local storage initializer（未注入 deviceId）下，两个真实 app tab 的并发草稿均保留并显示冲突；移除 legacy key 后 reload 仍可恢复。
- 确认框、记忆编辑、历史 rename 各自打开时切换 owner；首个接收 RAF 无旧文本/输入/弹窗；导航或 reload 前 composer 已恢复，通过实际历史控件重新打开新 owner 会话。两个 owner 的后端会话、消息和记忆快照完全不变，未发清除/删除/更新/保存请求。
- 真实 Go export 响应通过 route.fetch 获得后暂存在有 deadline/finally 的可控 barrier，分别执行 owner switch 或实际浏览器 Back；原生请求取消、无 download、后端快照不变。这里只证明浏览器原生取消；非合作晚到 Promise 仍由独立组件测试覆盖。
- 第二 app 真正执行 clear 推进 epoch；第一 tab 的旧确认失效，即使调用保留的旧 DOM Confirm，也不能删除新 epoch memory；另一 owner 快照不变，第一 tab 无写请求。

每端收集全部 pageerror 并断言零错误。诊断阶段用于定位拒绝的临时 reason/response 探针已经删除。runner 的浏览器启动/退出置于 try/finally，部分启动失败也清理本任务资源；独立测试只运行隔离 host 环境，不能当作 Docker 证明。

## 执行证据

| 检查 | 结果 / 日志 |
| --- | --- |
| 原移动端 route adapter 回归红 | 1 failed / 10 filtered；`/tmp/gate14-router-red.log`；原文件在 finally 恢复 |
| 原 transport 失败路径红 | 6 failed + 4 unhandled errors；`/tmp/gate14-transport-red.log` |
| 移除 exact-token guard 的微任务红 | 2 failed / 10 filtered；`/tmp/gate14-auth-microtask-red.log`；文件已恢复 |
| 助手 owner 微任务初始红 | 3 failed / 1 passed / 12 filtered；`/tmp/gate14-assistant-owner-red.log` |
| 最终 focused Node/happy-dom（含 auth 21项） | 10 files / 100 PASS；`/tmp/gate14-focused-complete.log` |
| 最终 Framework7/workspace Chromium | 2 files / 16 PASS；`/tmp/gate14-browser-complete.log` |
| 最终 native desktop→mobile 第1轮 | 14 PASS、零 pageerror；`/tmp/gate14-native-final1.log`；`.vitest/gate14-final1/identity-results.json` |
| 最终 native desktop→mobile 第2轮 | 14 PASS、零 pageerror；`/tmp/gate14-native-final2.log`；`.vitest/gate14-final2/identity-results.json` |
| vue-tsc --noEmit | PASS；`/tmp/gate14-type-final.log` |
| 本门改动 TS/Vue scoped ESLint | PASS；`/tmp/gate14-lint-final.log`、`/tmp/gate14-service-lint.log` |
| 四个修改/新增 mjs 的 node --check | PASS；仓库 ESLint 默认忽略 scripts，不声称脚本 lint PASS |
| source-lines / git diff --check | 1701 files / 0 超过500；`/tmp/gate14-lines-final.log`；diff check PASS |

父审独立复验 auth 21项和浏览器16项亦通过。早期 native 单轮成功后重复曾失败，故不将 `/tmp/gate14-identity-final.log` 的早期成功作为 auth 修复证据；最终两轮均在移除诊断探针、完成 exact-token/owner guard 后执行。

复跑命令：`EBK_E2E_ARTIFACTS=.vitest/gate14 node scripts/assistant-experience-identity-review.mjs`。可用 `EBK_E2E_VARIANT=mobile` 单端定位；`EBK_E2E_PRODUCTION=1` 使用 host dist，仍不等于容器验证。最终主 `node scripts/assistant-experience-e2e.mjs` 会自动包含新 identityFlows。

本门没有重跑完整 E2E、Docker、完整 unit/Go 或三数据库矩阵；留给最终串行门。旧 boundaries 已复用 native helper；实际 bfcache 若 pageshow.persisted=false 仍必须记 NOT RUN，本门不以普通 Back 替代。真实商业主/备用模型、物理键盘/输入法、手动屏幕阅读器亦无新增证明。未提供虚构 coverage 百分比或全链路10分。

本门完成并冻结。后续性能/设计收敛与最终矩阵由父审安排；未 commit/push。

## Final container compatibility regression (2026-10-16)

Gate16's real old-main backend/current production frontend test exposed a
missing upgrade notice despite actual capabilities404. The isolated state
regression reproduces concurrent initial `initialize` and foreground
`synchronize`;3 of4 assertions fail before repair. Compatibility classification
now occurs only at the current capability request's session boundary rather
than the initialize caller. The4 regressions plus existing store/lifecycle/read
ordering suites pass65 tests. Full final rebuilt frontend/container evidence,
first-frame title/memory checks, actual upstream cancellation and retained
bfcache limitations are documented in Gate16. No synthetic bfcache event is
used to substitute for a real browser restoration.


## 最终串行证据闭环（2026-10-16）

本次仅核读 Gate16–18 报告、实际结果 JSON、构建 provenance、父审最终 unit 日志和新增 compatibility 四项测试及 `session.checkIdentity` 的当前 ticket/epoch/disposal 防线；未重复运行无变化测试，未修改产品、主报告或提交。没有发现新的具体缺陷。

- [容器主结果](../ai-assistant-experience-evidence/audit-e2e/main/results.json)实际为 **47 PASS / 2 NOT RUN**，包含本门14项原生跨标签身份流程和双端应用流程；当前首个接收 RAF 进一步覆盖旧标题与保存的记忆文本。[账单引用结果](../ai-assistant-experience-evidence/audit-e2e/reference/reference-results.json)为 **10 PASS**。
- [真实旧后端结果](../ai-assistant-experience-evidence/audit-e2e/compat/compat-results.json)双端 **2 PASS**，每端真实 capabilities404、明确升级提示、零 provider 调用。四个新增状态回归分别验证初始化被 foreground 同步取代、升级后恢复、旧404不能覆盖新成功、history404不能误判缺失 capabilities；这些是受控 store 测试，与实际旧后端证据分别成立。当前分类只在有效 capability 请求边界进行，原兼容性条件已关闭。
- [Gate17](17-final-test-matrix.md)完成 **7 files / 62 Chromium PASS**、lint/build 和完整 backend/race/三数据库/迁移矩阵；[矩阵命令记录](../ai-assistant-experience-evidence/audit-matrix/commands.json)所列命令均 exit0，原 opt-in skip 和后续明确执行的性能检查仍按原报告分别记载。父审最终 `/tmp/ebk-parent-final-unit.log` 实际为 **44 files / 38,917 PASS**，不把总测试数量称为助手 coverage。
- [Gate18证据](../ai-assistant-experience-evidence/audit-harness-finalization/summary.json)与两次 build provenance 确认当前107个 dist 文件 hash 及 Linux executable hash 与 Gate16 完全相同；source-tree hash 不同，未声称源码树相同。后续改动为验证脚本/测试，因此保留 Gate16 的产品 E2E 适用性，不把当前新镜像冒称为重新跑过47项业务流程。

原先待最终本地应用/测试矩阵补齐的条件已关闭，当前本地可执行范围的 Vue 测试与设计简化评为 **10/10**。此评分表示该范围内已知缺陷和测试缺口闭环，不表示零潜在缺陷或全环境覆盖；上文首次冻结记录及红运行全部保留。

两项真正 bfcache 恢复仍为 **NOT RUN**：最终证据仅有 `back_forward`、`notRestoredReasons: masked`，没有 persisted=true；普通 Back 和原生 storage 通过不能替代恢复证明。商业主/备用模型、物理输入法/键盘、手动屏幕阅读器、其他浏览器引擎、amd64、stock Dockerfile/Node26 前端构建与精确 stock Alpine 运行镜像仍无新增证明，沿用 Gate16–18 的外部/环境边界。仅更新本报告后再次冻结。
