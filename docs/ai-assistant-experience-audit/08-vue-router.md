# 08 · Vue Router 串行审查

2026-10-16 最终复审；当前工作树包含 Gate01–14 与父审后续修复，基线仍为 `91081998` 相对 `origin/main`。
**当前代码合规与定向路由行为范围：10/10。历史首次评分：7/10 → 9/10，原始证据保留在下文。** 本评分仅表示本门适用的路由复用、异步意图、平台适配和清理要求已关闭；不是完整运行时、Docker、bfcache、物理设备或全矩阵验收评分。

## 最终复审补记（Gate14 后）

重新实读 Vue Router skill、param-change/lifecycle-cleanup 参考，Vue best practices 及 reactivity/sfc/component-data-flow/composables 四必读，08 与 14 报告；核对当前双端路由适配、WorkspaceLifecycle、AutoScroll 与 Framework7 安装版本的 `router-class.js:updateCurrentUrl`。本轮只更新本报告，未改产品代码、测试或 master manifest。

- **移动端内部 replace 已关闭重叠页面缺陷。** Gate14 的 native 应用流程发现连续 owner 变化与 history transition 后出现两个 `.page-current`、约 -78px transform；其旧 adapter 回放明确红灯。当前仅在 replace 且 current.path 为 `/assistant` 时调用既有 `updateCurrentUrl(target)`，然后 `refreshRoute()`。独立读取 F7 实现确认它替换 history 末项、更新 currentRoute 的 query/hash/params/url/path、在 browserHistory 启用时调用 History.replace、saveHistory 并发出 routeUrlUpdate；无需重建页面或自行修改 history 数组。
- **路由响应与普通导航保留。** refreshRoute 创建新的 shallowRef query 投影，Workspace 的字段 watcher 立即失效旧 generation，再于 nextTick 读取最终参数。首次 send 正式 ID、404 空 ID、branch recovery 均经同一路径；不需要依赖重挂载来恢复锚点或确认 branch。普通 history 选择仍调用 F7 navigate 并保留动画；page beforeout 卸载 workspace、beforein 恢复当前 query 的行为不变。exact conversationId 比较区分 undefined（恢复最近）与空字符串（显式新建）；同目标刷新不会因字段值相同产生 open 循环。
- **原生命周期约束继续成立。** 初始化失败仍 attach 当前实例的恢复监听，卸载撤销旧工作且只 detach 自己的资源；初始化后的 open 读取当前 route，没有重新使用旧 owner payload。Gate07 后续启动 caller lease 不改变这里的路由适配。branch marker 仍只在 URL 目标和已加载 detail 一致后 acknowledge；阅读位置仍由 scoped store 和 MessageList/AutoScroll 管理。Gate09 已关闭选区/resize lead，相关真实 MessageList selection/Markdown 行为本轮随 reading suite 回归。
- **本轮独立执行：** `npx vitest run --config vitest.browser.config.mts src/views/base/assistant/AssistantPages.browser.test.ts src/components/common/assistant/AssistantReadingRoute.browser.test.ts`：**2 files / 14 PASS，exit 0，0 skip，0 unhandled error**，`/tmp/gate08-rereview-browser-final.log`。覆盖双端 query/message/version、leave/back、first-send replace、404、初次 offline 恢复、重复 owner/F7 transition 控件位置，以及复用阅读/selection/Markdown。workspace lifecycle：**1 file / 18 PASS，exit 0**，`/tmp/gate08-rereview-lifecycle.log`，包括异步卸载、参数合并、初始化微任务、branch acknowledge。`git diff --check` PASS。首次 browser 命令沿用历史 `.ts` 配置名，因 Gate12 已改为 `.mts` 启动失败；更正后得到上述实际成功结果，未把启动失败计入测试。
- **证据来源分开：** Gate14 的两端 native 14 场景连续两轮 PASS、父独立 Framework7/workspace 16 PASS 见 [14 报告](14-vue-testing.md) 和 `/tmp/ebk-parent-gate14-browser.log`；本轮未重复运行那些应用流程，也不将本轮 synthetic identity/online 通知称为 native storage 证据。
- **仍保留的最终验收边界：** 完整 Docker、完整应用 E2E/Go/三库/unit matrix 尚待最终阶段；bfcache 需要实际 pageshow.persisted 命中，普通 Back 不替代；物理设备键盘/触摸、商业模型、手工屏幕阅读器仍无本门证明。原下文“后续交接”为首次审查历史，其中 Gate09/14 已完成的部分以上述补记与对应报告为准。

本轮复审完成并冻结；无新增阻断，不以额外重构扩大范围。

## 首次审查记录（历史 9/10）

2026-10-16；当时工作树包含前七门未提交修复。已关闭当时证实的 route reuse、query 更新和 404 路由同步问题；当时尚未完成的 native 跨 tab 和 selection 证据分别由后续 Gate14、Gate09补齐。

## 实读依据与职责边界

- `.agents/skills/vue-router-best-practices/SKILL.md`；适用参考 `router-param-change-no-lifecycle`、`router-beforeenter-no-param-trigger`、`router-guard-async-await-pattern`、`router-simple-routing-cleanup`。既有平台路由不改写，不引入手写 hash routing 或 next callback guard。
- `.agents/skills/vue-best-practices/SKILL.md` 及 `reactivity`、`sfc`、`component-data-flow`、`composables` 四必读参考。
- project-spec / global-prompt、software-design-philosophy，主 audit manifest、05/06/07 报告，主计划 3.5/4.1/4.2 与 R01/R07/R32 路由/阅读契约。
- 定向测试读取 Vitest、vue-testing-best-practices，以及 browser-vs-node、async/flushPromises、blackbox 三参考。

实施前职责判断：桌面适配器只将 Vue Router query 投影到 workspace props，移动适配器只通过 Framework7 page beforein/beforeout 控制 workspace 生命周期；Lifecycle 负责最新 route 意图和实例 attachment；MessageList 负责 DOM 滚动视图；AutoScroll 负责跟随、锚点与自己的浏览器资源。保持 typed props/emits 和 imperative reading API，不新增组件层级或 store facade。

## 已关闭的问题与行为证据

1. **复用列表继承前一个会话的停止跟随状态。** `restore(undefined)` 原先直接返回，A 已上滚后切到没有保存锚点的 B，仍保留 A 的 following=false。新增 Chromium 测试使用真实 MessageList、真实 Vue Router memory history、24 条真实排版消息；同一个滚动 DOM 元素从 A 切 B，距底仍为 **3595px**，明确红灯。现在 undefined 表示新会话默认跟随最新，清 anchor/unread/following 后定位到底部。浏览器验证 B 到底、back 回 A 恢复原 messageId 和像素偏移、forward 回 B 再到底。测试通过 scrollTop 写入触发浏览器滚动及真实布局；未声称物理触摸或完整应用 E2E。
2. **同步 watcher 读到 props 的中间态。** 一次 `conversationId/messageId/versionId` 更新原先触发三次 `open`，包含旧参数组合。现在 sync watcher 立即增加 generation 撤销旧续程，但在 nextTick 后紧贴实际 open 检查 generation，再读取最终 props。红灯为预期一次、实际三次；修复后仅一次并携带全部新参数。owner/epoch 更新也不会在 session 同步写入栈中发起新的 route read。没有删除同步失效保护。
3. **404 清空 id 未更新 URL。** `store.open` 404 将 conversationId 置空时，id watcher 因 navigating=true 跳过；finally 清标志不会重新触发 watcher。openRoute 完成并确认当前 generation 后显式把这种空 id replace 到 `/assistant?conversationId=`。受控测试原先无 navigate，修复后发出 replace；真实 Vue Router 与 Framework7 均验证 404 后 URL 和空态一致。没有扩大到网络错误的强制清空。
4. **延迟 DOM 恢复与资源所有权。** MessageList anchor watcher 用 onCleanup 撤销 await nextTick 后的旧恢复；AutoScroll 对 late restore/observer callback 增加 disposed 防护，单一 RAF 句柄在替换/卸载时 cancel。资源测试验证第二次滚动取消第一帧、卸载取消第二帧，晚 observer/restore/latest 不再改变 detached panel。移除全仓无调用者的 preserveWhile；保留已有 VueUse ResizeObserver，不再增加包装层。
5. **初始化与 attachment。** initialize 之后没有使用预捕获的 route/owner payload：在实际 openRoute 边界读取最新路由，后续每个 await 用当前 generation 验证。新增 helper→caller 微任务切 route 测试只打开新 message intent。attachment 是当前组件持有的恢复监听资源，不能把初次加载成功当作 attachment 授权；保留失败仍 attach 与卸载仅 detach 自己 attachment 的行为。没有新增 store.initialize 返回值或削弱 Gate07 session lease。双端浏览器用初次 capability rejection、受控 offline 状态和 synthetic online 通知，验证实际监听恢复 composer、不自动发送；这不是 native 网络切换或 native storage event 证据。
6. **已核对的路由契约。** 真实平台测试覆盖同 path 的 conversation/message/version query 变化、版本读取参数、leave/back 后历史与草稿、首次 send 后正式 id 的 replace（back 到先前 away 页面，而非重复空白助手页）、404 和初次离线恢复。桌面是 Vue Router memory history，移动是实际 Framework7 router/page lifecycle；API 为协议 fixtures，外围 layout/reference dialog 为既有测试替身。branch recovery 的 controlled workspace 回归证明只有 target props 已到达且对应 detail 已加载后才 acknowledge marker；原 store pending/branch、owner/epoch、404 cleanup、逆序 version 保护由本次 focused store suites 回归。没有把这些替身测试写成真实 API/数据库 E2E。

## 最终执行证据

- `/tmp/gate08-reading-red.log`：真实 Chromium 布局 **1 FAIL**，失败在 B 距底 3595px。
- `/tmp/gate08-lifecycle-red.log`：修复前 route coalescing 与 404 **2 FAIL / 13 skipped**；0 unhandled error，skipped 不计成功。首次编写这两个用例曾缺 exposed restoreAnchor stub，修正测试替身后重新回放，最终红灯日志是干净回放。
- `/tmp/gate08-focused-final.log`：**7 files / 108 tests PASS，exit 0，0 skip，0 unhandled error**。文件为 workspace lifecycle、AutoScroll、store lifecycle/readOrdering/assistant、workspace commands、AssistantPages。首次 focused 命令误写 assistant.test 路径，只有 5 files/69；最终更正后实际跑到 7 files/108。
- `/tmp/gate08-browser-final.log`：**2 files / 11 tests PASS，exit 0，0 skip，0 unhandled error**。AssistantPages 十项（每平台五项）和 ReadingRoute 一项。最初 online 用例尝试在 capabilities 未加载时定位 composer，但产品此时还未渲染 composer；测试改为先恢复能力再断言输入，没有为测试更改产品行为。Framework7 初始 away URL 会附带 runner 参数，back 断言按 pathname 比较。
- `npx vue-tsc --noEmit` PASS：`/tmp/gate08-typecheck-final.log`；全部本门改动 Vue/TS scoped ESLint PASS：`/tmp/gate08-lint.log`，后增测试再次检查 `/tmp/gate08-lint-final.log`。
- `npm run check:source-lines`：1691 files / 0 exceed 500；`git diff --check` PASS。控制流按可读结构展开，未用压缩凑 500 行。
- 父独立重跑 workspace lifecycle/AutoScroll/readOrdering：**3 files / 29 tests PASS**，`/tmp/ebk-parent-gate08-focused.log`（父消息；早于最后新增 branch 用例，不与本门 108 混算）。
- native config ESM warning 仍保留交 Gate12；一次中途 Framework7 first-send 运行还出现第三方 Markdown 非函数 slot warning，最终 browser 日志无该 warning。没有隐藏警告或更改第三方渲染器。

## 后续边界与交接

- Gate09：父审发现文本选择时 `layoutChanged` 的 else 分支仍可 `restore(anchor.following=true)`，可能继续滚到底；需要真实 selection/resize 红绿回归。本门只修路由默认锚点与必要 teardown，不将 selection、Markdown 动态变化、prepend/键盘全部宣称通过。审核现有 VueUse observer scope 清理与滚动资源复用。
- 最终 E2E：真实服务端、真实双 tab auth/storage event、真实浏览器 history/base/hash 配置、实际 bfcache 命中、真实 route branch recovery 与完整账单往返仍须后续阶段。未命中 bfcache 不能写 PASS；synthetic online/storage 不能充当 native 跨 tab 证据。
- Gates10/11：刷新双端视觉/200% zoom 与键盘/屏幕阅读器证据。完整 Docker、Go/database、全仓 unit/browser matrix 按原串行流程执行。
- 本门未改 master manifest、未创建子 agent、未运行 Docker/全仓矩阵，未 commit/push。提交给父审后冻结，供 Gate09 接续。
