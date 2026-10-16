# 05 · Vue best practices 串行审查

2026-10-16；基线 `91081998` 相对 `origin/main`，包含前四门工作区修复。
**当前本地适用范围最终评分 10/10；历史初始 6.5/10 → 首轮修复 9/10。** 下文保留首轮记录；后续条件的实际关闭与仍未验证的外部边界见文末最终复审。 评分是本审查约定，不是技能自带评分系统。不把未执行矩阵或 happy-dom 当作真实平台证据。

## 依据与实施前边界

读取 `skills/vue-best-practices/SKILL.md`，完整四份必读 `reactivity.md`、`sfc.md`、`component-data-flow.md`、`composables.md`；读取 `skills/software-design-philosophy/SKILL.md`、project-spec 及 global-prompt；读取主审查契约、01/02 报告。定向测试读取 Vitest、vue-testing-best-practices 技能及 blackbox、async/flushPromises、composable-host 三参考。使用原生 dialog，不新增 Teleport/KeepAlive/Suspense 或 VueUse 包装，因此无此类 optional 功能要求。

实现前职责地图保存在 `/tmp/gate05-plan.txt`，落实后如下。新增展示组件用于独立 UI 区段，异步约束仍由拥有它的模块集中管理，不按请求步骤拆 wrapper。

| 边界 | 单一职责 / typed contract |
| --- | --- |
| `AssistantCloudWorkspace.vue` | 双平台组合 surface；`variant/conversationId/messageId/versionId` props；`navigate/openReference` emits。组合历史、记忆、消息、composer、dialogs 与区段；保留少量连接动作。最终 495 行，展开模板和 CSS，未压缩凑行数。 |
| `useAssistantCloudWorkspace` | 产品命令和临时反馈；`confirm(message): Promise<boolean>` 与 route-key getter；统一保护确认、异步链和反馈。返回 typed actions/refs，关键异步分支已展开。 |
| `useAssistantWorkspaceScope` | UI authority lease：owner、deletionEpoch、会话、route intent 任一改变或 dispose，旧 lease 失效；比较值而非 capabilities 对象身份。 |
| `AssistantConfirmationDialog` | 一个待决确认与焦点恢复；`context/fallbackFocus` props，暴露 `confirm(message)`；替换/失效/卸载自动 decline。 |
| `AssistantExportDialog` | 导出选择、格式、请求取消、下载与 dialog 生命周期；`context` prop，暴露 `open(item?, all?)`；其余细节不外泄。 |
| `AssistantReferenceDialog` | bill verification notice 与焦点；`context/messages` props，`openReference/retryQuestion` emits，暴露 `check`。仍使用已有 reference composable。 |
| `useAssistantWorkspaceLifecycle` | 初始化/路由同步、store attachment、锚点恢复；store + route/navigate/restore/run typed options。实例持有 attachment 才 detach。 |
| `useAssistantWorkspacePanels` | viewport、history keyboard containment、memory/history focus；visibility refs + variant/showMemories/captureScope options。 |
| `AssistantWorkspaceToolbar` | toolbar/menu 的 UI 和菜单焦点；variant/conversation/memoryEnabled props，typed 用户意图 emits，暴露 focus。 |
| `AssistantRecoveryPanel` | 持久化状态、pending 请求和草稿恢复区；只读 store 投影，typed recovery intent emits；授权/执行仍在命令层。 |
| `AssistantConversationNotices` | 来源、失效、分支继续和归档状态区；disabledReason prop，typed navigate/continue/restart/unarchive emits。 |
| `AssistantExperienceComposer` | 原有输入/提交组件；保留 defineModel、IME 契约，将自身横屏样式归回自身，scoped class selectors。 |

## 要求与修复证据

1. **D4 职责和可读性。** 导出、确认、reference dialog、route/attach、viewport/focus 各归属其组件/composable。toolbar、恢复区和上下文通知拆为实质 UI 区段。移除模板内 pending filter / latest-message find / memory busy some，改 computed。新增源文件、测试和样式均低于 500 行。
2. **确认跨 owner/epoch/route 无效。** clear/remove/forget/历史 nextPage/local draft 均获得 scope lease，紧贴 store 命令调用重新验证。父审提出的额外微任务竞态也实际复现：`decide(true); queueMicrotask(changeOwner)` 会越过原先返回 boolean 的 helper；现返回 lease，最外层执行前重验。不是只在 helper 内 await 后检查。
3. **导出绑定原选择。** open 时捕获 id/title/displayed versions；提交时捕获 format/options。owner、deletionEpoch、route、取消、卸载会清掉 selection 并 abort。返回后检查 generation，即使实现不响应 AbortSignal，旧 blob 也不能创建 URL 或 click 下载。旧 finally 不影响新导出。
4. **run 的反馈和异步 tail。** owner/route 变化重置 error/working；旧失败/返回/finally 均忽略。计数覆盖同 scope 并发操作，旧操作不能提前清 working。continue/loadLatest→branch、viewVersion→selectVersion、summary/edit completion，以及外层 rename/save editor close、滚动/焦点都在最终副作用前校验。
5. **受控 mount/unmount。** 原实现等待 initialize/open 后无条件 attach，两个真实受控 Promise 回归均先红灯。现卸载后不再 open/attach/restore；路由 generation 防止旧 open 恢复新页面锚点；未 attach 的实例不会 detach 别的实例 attachment。
6. **编辑和分页生命周期。** route intent 改变清 message edit/retry state 和 version cursors；晚分页返回不能重新污染 cursor。owner/deletion 变化清 query/includeVersions/archived/summaryPeriod、关闭 history/memory；两个编辑器按 owner+epoch key 隔离，同 tick 重开也不能复用旧表单。实际 MemoryManager 输入旧 owner 文本后切账号/重开回归为空。
7. **Vue 基础。** primitive shallowRef、typed props/emits/template refs、script→template→scoped style、原 composer defineModel；新组件使用 class selectors；computed 无副作用。未新增依赖；只临时执行 Prettier CLI 格式化，没有改变 package/lockfile。未新增自造网络或 IndexedDB 抽象，也未进行没有行为证据的性能优化。

## 定向验证（实际执行）

- 首次 command 红灯：5 fail / 10 pass（四类确认切账号 + 旧 run error），`/tmp/gate05-command-red.log`。
- mount 红灯：initialize/open 两个 delayed unmount 场景均 FAIL，`/tmp/gate05-mount-red.log`。
- confirmation 微任务红灯：1 fail / 22 pass，`/tmp/gate05-confirm-microtask-red.log`。
- export 回归有效性：临时仅去掉 late-response generation guard，cancel/owner/route/unmount 四场景均 FAIL（实际 anchor click），4 fail / 4 pass；finally 恢复，`/tmp/gate05-export-red.log`。
- 最终 focused Vitest：`useAssistantCloudWorkspace.test.ts`、`AssistantCloudWorkspace.lifecycle.test.ts`、`AssistantExportDialog.test.ts`、`AssistantConfirmationDialog.test.ts`、既有 `AssistantPages.test.ts` 与 `AssistantExperience.test.ts`：**6 files / 55 tests PASS，0 skip，0 unhandled error**，`/tmp/gate05-focused.log`。中途新增 shallow component 测试缺少 exposed restoreAnchor stub 导致两条 unhandled rejection，补齐测试替身后重跑得到上述干净结果；没有把此前文件 pass 当作成功。
- Chromium `AssistantExperience.browser.test.ts`：**13 PASS**，`/tmp/gate05-browser.log`；只覆盖展示 fixture、输入/布局，不是完整应用 E2E。
- 新增 Chromium `AssistantCloudWorkspace.browser.test.ts`：**4 PASS**（desktop/mobile 原生 dialog：取消恢复 More 焦点、切 owner 自动 decline、route prop 变化取消导出、非协作晚响应不 click），`/tmp/gate05-workspace-browser.log`。这是直接 workspace component browser test，不是完整真实平台 route E2E。首次使用 getByRole 定位关闭后不在 accessibility tree 的按钮再断言 not-visible，4 条因 locator 找不到而超时；改为检查原生 `dialog[open]` 为零，重跑全绿，没有产品代码为迁就测试改变。
- `npx vue-tsc --noEmit`：PASS，`/tmp/gate05-typecheck.log`。
- 对本门所有修改/新增 Vue/TS 运行 ESLint：PASS，`/tmp/gate05-lint.log`。
- `npm run check:source-lines`：1679 files / 0 exceed 500；`git diff --check` PASS。后续增加的短 browser test 不改变任一源文件上限。
- Vitest native config loader ESM warning 仍存在，交 Gate12；未静默隐藏。

## 后续边界与复审

- Gate01/02 父审：D4 的主要职责/压缩模板问题已整改，请对实际接口和组合深度复审；评分不替代此审。
- Gate06：按主 manifest 继续 history query 父子 remount 漂移、memory source key、reason code 单一权威，以及本次新组件 Vue 语义复审。
- Gate07：store 的深层 owner/navigation、pending acceptance、send finally、404 清理及 D5 仍未由本门改动；UI lease 不替代 store 自身隔离。尤其运行到 store 内部 await 后的权限仍由 store 负责。
- Gate08：完整真实 Vue Router / Framework7 page reuse、back/forward、bfcache 和账号切换流程。这里证据是受控 props/lifecycle 与既有 happy-dom route tests，不冒充完整真实路由验收。
- Gate10/11：toolbar 200% zoom 密度、双端刷新截图、完整键盘/屏幕阅读器和视觉 review；本门仅保持现有设计并拆清所有权。
- 后续专门阶段执行 Docker、完整 unit/browser/Go/数据库/E2E matrix。本门没有 Docker、commit/push、主 manifest 修改或子 agent 委派。


## Gate18 后最终闭环复审

2026-10-16。本次仅更新05报告，复读当前 workspace/commands/scope/lifecycle、confirmation/export dialogs、History models、RecoveryPanel、session.checkIdentity/initialize 及 compatibility 四回归；核对06/07/08/10/11/15–18报告、01最终设计复审，以及持久化 E2E/matrix/harness summary。未改产品代码、测试或主 manifest，未委派、提交或重跑无变化测试。以下为既有实际执行证据的复核，不能记作本轮再次执行。

### 原保留条件已关闭

- **职责与单一状态来源：** Gate01最终复审已关闭D4/D5/D6；原05组件边界仍成立。Gate06将History query/includeVersions改为父持有的required named models，并将debounce/cleanup放在Workspace composable；来源复合key与reason-code权威已收敛。当前CloudWorkspace为398行，样式由其`<style scoped src="./assistant-workspace.css">`继续拥有；这是Gate11为可读性与行数作的组织选择，不新增全局样式所有权，也未重新压缩模板。
- **异步与store边界：** Gate07的session/conversation/commands/history模块承接持久化权限、请求顺序、pending和finally；UI仍以owner/deletionEpoch/route lease在最外层副作用前重验。确认返回lease、export非协作晚响应generation guard、editor close与分页cursor guard均仍存在。当前scope逐值比较，正常capabilities刷新不误清UI。
- **真实路由与资源：** Gate08最终10/10复审含双端实际router query/first-send replace/404/离线恢复与Framework7内部replace；其2文件14条browser及18条lifecycle通过。当前Lifecycle同步失效、nextTick合并最终props、卸载检查、实例自己的attach/detach及404空id同步均保留。Gate09的选区/resize/RAF与后续完整browser矩阵补齐阅读生命周期证据。
- **视觉与交互：** Gate10/11完成200% zoom toolbar、长内容、双端light/dark、模态history、More焦点与标题全文、输入描述和实际computed对比度检查；Gate16再以最终生产dist/容器服务执行2个设计、4个accessibility组合。这些为本地浏览器和实际页面证据，不是手工读屏或物理手机证明。
- **完整本地验证：** Gate16最终生产应用47 PASS / 2 NOT RUN、bill reference10 PASS、old-backend compatibility2 PASS；主结果JSON明确包含双端native storage下旧确认/记忆/历史撤销、owner/route中止导出、另一设备clear推进epoch和真实路由恢复。Gate17持久化summary记录全量lint/build通过、42文件38,907 unit和7文件62 Chromium通过、三库/迁移/race的各自结果。Gate18工具修复后父独立全量unit为44文件38,917 PASS；新增构建/10项smoke通过，107个dist文件与Linux executable同Gate16逐字节哈希相等，因此无需重跑未变化业务E2E。上述计数不相加为独立覆盖量。

### 最终 legacy 404 改动的专项核对

`session.checkIdentity`只在当前ticket、generation及未dispose时将**capabilities请求**404归为legacy；随后成功能力读取清legacy。`initialize`的通用catch不再把history404误认作旧后端。该协议判断没有移入workspace，也没有给Vue组件新增状态或接口。

当前RecoveryPanel仍只读`store.legacyBackend`展示升级提示；WorkspaceLifecycle初始化失败后仍为当前活实例attach恢复监听，旧实例不能attach。404归类不修改owner/deletionEpoch/route lease，也不放开确认、export或旧异步tail。四个compatibility用例分别覆盖同步超越初始化、升级恢复、旧404晚于新成功、无关history404；Gate16记录修复前3 FAIL/1 PASS、修复后连同store/lifecycle/readOrdering共65 PASS。真实old-main二进制配最终生产dist的desktop/mobile各通过，稳定观察5秒、每端仅2次capabilities404，0 provider调用。结合当前调用链，未发现该最终改动破坏workspace职责或生命周期。

证据链接：[容器E2E报告](16-container-e2e.md)、[主应用结果](../ai-assistant-experience-evidence/audit-e2e/main/results.json)、[全矩阵summary](../ai-assistant-experience-evidence/audit-matrix/summary.json)、[harness最终summary及字节等价](../ai-assistant-experience-evidence/audit-harness-finalization/summary.json)。旧native-config warning交接已由Gate12的`.mts`配置归属处理；不再作为本门待办。历史首轮命令/行数/评分仍保留原时点含义。

**最终10/10限于当前PR本地适用的Vue最佳实践要求；无剩余已确认必须修改项。** 真实商业主备模型与金融/长期上下文质量、物理iOS/Android键盘触摸、手工屏幕阅读器、实际`pageshow.persisted`的bfcache命中、其他浏览器引擎/amd64、stock Dockerfile/Node26阶段及精确stock Alpine构建仍为NOT RUN。实际bfcache的两项未命中不计PASS；完整发布认证和零未知缺陷不在本评分承诺内。冻结交父审。
