# 09 · VueUse 串行审查

2026-10-16；基线 `91081998` 相对 `origin/main`，包含 Gate01–08 未提交修复。**初始 7/10；本门代码合规与定向验证 10/10，整体运行时验收仍为 9/10。** 这是主审查契约的范围评分，不是 VueUse 自带评分；提升整体评分仍需后续真实应用 E2E、视觉与最终矩阵，本门没有用局部浏览器测试替代它们。

## 实读技能与条款映射

- `.agents/skills/vueuse-functions/SKILL.md`：先判断合适 composable，再决定是否使用；实读 `useResizeObserver`、`useEventListener`、`useScroll`、`useTextSelection`、`useRafFn`、`useWindowSize` 的 Usage/类型参考。项目已安装 `@vueuse/core` 15.0.0，无新增依赖。
- Vue best practices 及 reactivity、sfc、component-data-flow、composables 四必读参考；software-design-philosophy；project-spec/global-prompt；主 audit manifest、05–08 报告。保留 shallowRef/readonly、typed imperative reading API 和既有组件边界。
- 定向测试读取 Vitest、vue-testing-best-practices 与 browser-vs-node、blackbox、async/flushPromises、composable-host 参考。实际布局用 Chromium；happy-dom 仅承担 teardown/命令/生命周期回归。

实施前职责判断：MessageList 负责列表与 typed props/emits；AutoScroll 独占跟随、消息锚点、滚动监听、RAF 和布局反应；WorkspacePanels 独占双端 viewport/focus。修复不改变 store、路由、owner/epoch lease、异步命令或 attachment 的所有权。

| 候选/条款 | 判断及实际结果 |
| --- | --- |
| useResizeObserver / scope cleanup | 保留现有 content + panel 两个 observer；二者分别覆盖内容排版和视口尺寸。实读安装包实现：target watcher 会先 disconnect 旧 observer，stop 同时 disconnect/stopWatch，tryOnScopeDispose(stop) 自动收尾。无需再造 observer 包装。 |
| useScroll | 可提供 x/y/arrivedState，但不能直接表达消息 ID + 偏移锚点、显式跟随和程序滚动抑制。换用后仍需这些状态与恢复逻辑，增加双重权威，因此保留目前单一状态实现。 |
| useTextSelection | 当前只在布局事件同步判断选区，不需要响应式 selection 文本、rects、ranges 或全局 selectionchange listener。直接检查 ownerDocument 当前 ranges，可准确识别跨节点选区是否与面板相交。 |
| useRafFn | 当前 RAF 只结束一次 programmatic 抑制，不做动画循环；既有单句柄替换取消和卸载取消清晰。没有新增 pause/resume 层或无消费者的 callback API。Gate08 已删除 preserveWhile，本门确认无调用者残留。 |
| useEventListener / useWindowSize | 现有滚动 watcher 显式更换旧/新 panel listener，卸载移除当前 listener；viewport mount/unmount 成对维护 window 与 visualViewport resize。useWindowSize 单个 type 无法同时替代 innerWidth 断点及 visualViewport.height 高度策略；两次调用再 computed 并不简化这个小模块。按用户“先判断而非机械替换”要求保留清楚的现有实现。 |
| 深模块/最小接口 | 自动布局选择保护集中在 AutoScroll 的私有 layoutChanged；公共 API 无新增方法、参数或 store 依赖。contentChanged 是内部两种 observer 来源，用于避免仅 panel resize 冒充新内容。 |

## 已证实并修复的问题

1. **选中文本仍强制到底。** 内容 observer 原先检测非空 selection 后走 restore(anchor)，但 anchor.following=true 又调用 scrollToLatest。Chromium 使用真实 Selection/Range 与实际尺寸，从底部 900px 选中文字，内容增长后错误滚至 1200px。现在相交的非折叠选区将 follow 和原消息锚点转为 paused，再恢复原相对位置；不能将刚选中的阅读状态解释为跟随。
2. **panel resize 绕开 selection。** 原 panel observer 直接 restore(anchor)，缩小面板把 scrollTop 从 900px 推至 1000px。现在两个 observer 共享 selection/anchor 保护，但只有 content resize 标记 unread。仅 resize 时也保持 900px 和 paused 状态。
3. **选区边界与恢复。** 根据 range.intersectsNode(panel) 检查，包括选区端点在不同节点的情况；页面外文本不会暂停跟随。清除选区不会偷偷恢复跟随，用户按 Back to latest 明确恢复；产品原有滚动到近底部恢复逻辑不变。主动新会话 restore(undefined) 和主动 Back to latest 不被被动布局保护拦截。

没有改写滚动库、增加通用抽象或更改存储/数据结构。没有新增 resume/dispose callback API。原 programmatic 标志避免自己恢复锚点产生的 scroll 事件被当成用户滚动；布局恢复只改变 scrollTop，不改变观察元素尺寸。真实 MessageList 的 Back to latest 按钮出现会改变 panel 尺寸，浏览器回归覆盖该次额外 observer 反应并验证最终锚点稳定。

## 真实布局与回归证据

- `/tmp/gate09-red.log`：修复前 **1 file / 2 FAIL + 2 PASS**，0 unhandled error；两个失败分别为 content 900→1200 和 panel 900→1000。外部 selection 与 prepend/image 用例原先已通过，没有把它们写成修复前失败。
- `/tmp/gate09-browser-final.log`：**2 files / 7 tests PASS，exit 0，0 skip，0 unhandled error**。AutoScroll host 四项：面板选区+增长+显式恢复、仅 panel resize、页面外选区、上滚后 prepend + 实际 image.decode 后 180px 布局增长 + resize。host 禁用原生 overflow-anchor，证明消息锚点恢复本身有效。
- 同一 browser run 的真实 MessageList 三项：A→B→back→forward 保持 Gate08 契约；实际 Selection 后新增消息保留被选文本的屏幕 top（<2px）并点击真正 Back to latest 按钮恢复到底；真实 Markdown renderer 增长时跟随，用户上滚后再次 Markdown 增长保持 messageId/offset/following。后两项使用实际组件 CSS，不模拟几何值。selection 由 DOM Range 建立，不声称物理鼠标拖选；图片为本地 data SVG，不是远程图片网络 E2E。
- `/tmp/gate09-focused.log`：**4 files / 55 tests PASS，exit 0，0 skip，0 unhandled error**：AutoScroll teardown、workspace commands、workspace lifecycle、store readOrdering。Gate08 的重复 RAF 取消、卸载取消、晚 observer/restore 无副作用继续通过；Gate05/07/08 的相关身份/route fencing 未改动。
- `npx vue-tsc --noEmit` PASS：`/tmp/gate09-typecheck-final.log`；三个修改/新增 TS 文件 scoped ESLint PASS：`/tmp/gate09-lint-final.log`。
- `npm run check:source-lines`：1692 files / 0 exceed 500；`git diff --check` PASS。没有为 500 行限制压缩实现或测试。
- 中途新增 Markdown 回归首次沿用带 references 的 fixture，但 locale test stub 仅提供 tt，因缺 formatAmountToLocalizedNumerals 失败；将这个只验证排版的 fixture references 设为空后重跑上述 7 项全部通过。没有为 fixture 改产品代码或吞掉错误。
- 父独立重跑当前两份真实 scroll/MessageList browser suite：**2 files / 7 tests PASS**，`/tmp/ebk-parent-gate09-browser.log`（父消息）；父也确认 panel/content resize 统一及 range 交集判断。
- native Vite config ESM warning 未隐藏，继续交 Gate12。没有运行 Docker、完整测试矩阵、数据库或商业模型。

## 后续边界

本门关闭 selection/resize 的已证实缺陷，并对 VueUse 资源所有权给出代码与定向证据。浏览器 coverage 限 Chromium 组件环境；native 多 tab/storage event、实际 bfcache 命中、真实模型效果、物理移动键盘、手工读屏不在本门 PASS 范围。

Gate10 frontend-design 接续检查双端视觉/字体/密度/主题、200% zoom toolbar、真实 viewport 与移动键盘布局；Gate11 检查完整键盘/读屏。最终真实服务端 E2E 和全部数据库/单元矩阵仍按父任务串行流程执行。本门未改 master manifest、未子委派、未 commit/push。提交父审后冻结。
