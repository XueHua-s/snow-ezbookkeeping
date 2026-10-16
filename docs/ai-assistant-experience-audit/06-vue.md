# 06 · Vue 语义串行审查

2026-10-16；审查范围为 `91081998` 相对 `origin/main` 的 assistant experience PR，包含前五门工作区修复。
**初始 7/10 → 首轮修复 9/10 → 当前本地适用 Vue 语义范围 10/10。** 这是审查约定的范围评分，不是 Vue skill 自带评分，也不是外部发布认证。首轮等待的 store、真实路由/视觉和本地完整矩阵已有后续证据，闭环见文末；首轮执行记录与外部 NOT RUN 均保留。

## 实读依据与边界

- `.agents/skills/vue/SKILL.md`，`references/script-setup-macros.md`、`references/core-new-apis.md`。
- `.agents/skills/vue-best-practices/SKILL.md` 及四份必读：`reactivity.md`、`sfc.md`、`component-data-flow.md`、`composables.md`。
- `.agents/skills/software-design-philosophy/SKILL.md`、`.codex/skills/ezbookkeeping-project-spec/SKILL.md` 及 `references/global-prompt.md`。
- 主审查 manifest、05 报告。回归实现读取 Vitest / vue-testing-best-practices 两技能，以及 blackbox / async-await-flushpromises 两测试参考。
- 没有引入 Teleport、KeepAlive、Suspense、async setup 或自造指令；不因此增加这些可选功能。

实施前职责判断：History 是父查询状态的可卸载视图，query/includeVersions 使用 required named models；Workspace composable 拥有过滤请求生命周期；MemoryManager 仍仅发出来源导航意图；CopyButton 自己拥有剪贴板反馈生命周期；DTO 常量拥有 reason-code 的运行时/类型权威。保留 Gate05 的命令 lease、owner/epoch 隔离、路由 generation、独立对话框和 attachment ownership。

## 要求、发现与修复

1. **单一响应式状态 / defineModel。** 旧 History 子组件 query/includeVersions 在 v-if 卸载后重建为空，父 query 与 searchResults/cursor 却保留。真实 Workspace + History 交互回归先失败。现在两个字段由 Workspace 持有，子组件用 required `defineModel`，无默认值造成父子不同步；fixture 也显式绑定 models。编辑标题仍是短期本地 state。
2. **watch 清理与可重入打开。** debounce 归 Workspace，观察 query、includeVersions、historyOpen。关闭或重排过滤时 cleanup 清定时器；卸载由 effect scope 停止 watcher；执行前验证原 lease，owner/epoch/route 失效不得发请求。scope 变化（包括 panel 保持打开时外部 route props 变化）同步关闭 history，保留同 owner 过滤值；重新打开非空查询会重新请求，即使先前 timer 因导航被取消。无查询的初次打开仍用 store 原有 setHistoryVisible 刷新，避免重复。等待期间隐藏旧 matches、禁用分页，防止 B 输入配 A 结果/cursor。父审追加的 A→输入 B→关闭→切 conversation→重开交错也先红后绿。查询返回先后顺序的 store 权威留 Gate07。
3. **列表 identity。** Memory source 允许空 messageId；改用 conversationId + messageId 复合 key。实际重排两个 conversation-only 来源保持对应 DOM 按钮身份，并分别发出 `['22', '']`、`['11', '']`；消息来源仍发出消息 ID。Workspace 原 source→navigate 连接不改。History match key 同时包含 conversation/message/version，避免串接歧义。
4. **生命周期 / 异步反馈。** CopyButton 在 clipboard.writeText 未完成时卸载，旧代码仍创建 success timer；多次点击的旧 success 也会盖掉新 failure。加入本组件 generation，在 await 成功/失败分支紧贴反馈重验，unmount 清 timer 并失效 generation；不修改实际 clipboard 权限或点击时原文快照。
5. **事件传播。** History 内 Escape 取消 rename 时原事件会继续冒泡到 Workspace 并关闭历史面板。nav 的 Escape 使用 stop modifier，第一次只退出编辑，非编辑状态第二次关闭；组件交互覆盖。
6. **运行时 / 类型同一权威。** `assistantReasonCodes` 是 DTO 中 const tuple，AssistantReasonCode 从其派生，snapshot validator 引用同一常量；移除独立 reasons 数组。没有生成框架或额外包装层。测试接受每个声明 reason 并拒绝未知字符串/数字。
7. **SFC 可读性与响应式基础。** 展开 History、Memory、Copy、Actions、Evidence、MessageList、Capacity、fixture、CodeBlock 的压缩模板/CSS；Markdown 调整为 script→template→style 并展开，双端 assistant shell 展开。mobile shell 的替换式 route/primitive active 用 shallowRef。未格式化无关旧页面。上述源文件均 <500 行，Workspace 497 行；未压缩凑行数。

## 其余语义检查

- 新增 Gate05 Confirmation/Export/Reference dialogs、Toolbar、Notices、RecoveryPanel 与 Scope/Lifecycle/Panels 已复读：typed props/emits、imperative template ref API、computed 派生、同步身份 watcher、卸载清理继续成立。
- Composer 的 required model、code-point/byte computed、IME composition guard 与 nextTick textarea resize；MessageList 的 keyed messages、typed events、anchor/scroll composable；Evidence 的 computed 金额/引用投影；Markdown 的转义内容、安全链接与 parser-backed code 均保留。
- 没有解构 reactive props 导致脱离依赖、computed 内写入、在同一元素叠 v-if/v-for、增加不可信 v-html、无效双向 prop mutation。元素 DOM refs 使用 useTemplateRef；新状态 primitive 用 shallowRef。没有为微小渲染计算增加缓存框架。
- 原有 legacy 页面仅审 PR 修改点；不把全项目旧 SFC 的样式/section 顺序当作本门全局重写任务。部分本 PR scoped CSS 仍用元素选择器；本次展开保持样式行为，后续设计门可在真实视觉验收中统一 class selectors，本门不以纯格式化宣称全部风格要求满分。

## 执行证据

- `/tmp/gate06-red.log`：history remount 行为 **1 FAIL / 5 PASS**，失败为输入值从 budget 变空。
- `/tmp/gate06-contracts-red.log`：临时恢复 messageId-only key、移除 CopyButton completion guards 后 **3 FAIL / 0 PASS**；随后恢复修复。分别捕获 DOM identity、卸载后 timer、旧成功覆盖新失败。
- `/tmp/gate06-route-filter-red.log`：父审补充 route/filter 交错 **1 FAIL / 11 PASS**，旧 Result A 被展示。
- `/tmp/gate06-visible-route-red.log`：未先关闭 panel 的外部 route 变化 **1 FAIL / 12 PASS**；修复后两种导航方式均通过。
- `/tmp/gate06-focused.log` 最终 **8 files / 82 tests PASS，0 skip，0 unhandled error**：AssistantVueContracts、AssistantCloudWorkspace.lifecycle、assistant_conversation_stream、useAssistantCloudWorkspace、AssistantExperience、AssistantPages、AssistantConfirmationDialog、AssistantExportDialog。首轮曾传错 AssistantPages 路径只跑到 7 files，最终已更正并实际覆盖 8 files。
- `npx vue-tsc --noEmit` PASS，`/tmp/gate06-typecheck.log`。初次测试新增输入元素断言缺少 HTMLInputElement 类型，修正后重跑。
- 本门所有修改 Vue/TS 的 scoped ESLint PASS，`/tmp/gate06-lint.log`。
- `npm run check:source-lines`：1681 source/test/style files，0 exceed 500；`git diff --check` PASS。
- native Vite config ESM warning 未隐藏，仍交 Gate12。本门未运行 Docker、全仓 matrix 或真实 browser/route E2E；没有将历史 Gate05 browser 结果算作本次运行。

## 首轮后续交接（历史）

- Gate07：D5、store synchronize 404、send finally、命令 acceptance、同 owner 逆序读/搜索/版本/记忆竞争仍须 store 自身 fencing；UI generation 不替代这一层。搜索结果窗口必须与最近用户意图一致。
- Gate08：Framework7 page reuse、Vue Router back/forward、真实 native storage identity 和 bfcache；特别验证保留搜索过滤条件的导航返回。当前是受控 props 导航与 happy-dom 证据。
- Gate09：AutoScroll 的 resize observer、requestAnimationFrame 与 preserveWhile 资源/生命周期检查留该门；本门未把其 RAF 取消当作已完成。
- Gates10/11：200% zoom toolbar、移动端密度、真实键盘/屏幕阅读器与展开后的样式视觉核验。
- Gates12–14 / 最终阶段：build config、完整测试基础设施与应用 E2E/database matrix。未改主 manifest、未委派、未 commit/push。


## 最终本地闭环复审（Gate18 后，2026-10-16）

本轮按父审限定只更新本报告，未重新执行无变化测试、未改产品代码、主报告或其他门评分。复核 Gate07–11、15–18 的相关结果与当前 History/Memory/Copy、Workspace 过滤 watcher、WorkspaceLifecycle、WorkspacePanels、AutoScroll、移动适配器、reason-code DTO/validator；未发现新的明确 Vue 语义缺陷。首轮9分的后续本地条件现已关闭：

- **状态/异步所有权：** [Gate07](07-pinia.md) 已完成 store reset、dispose、404、同 route 逆序读取与 caller lease；当前 History required models、owner/epoch reset、过滤 debounce cleanup、scope 变化关闭面板及重开查询仍保留。Copy generation、memory composite key 和 reason-code 单一常量未被后续改动撤销。
- **实际路由/清理：** [Gate08](08-vue-router.md) 最终复审记录14 Chromium路由/阅读测试和18 lifecycle测试；当前同步 watcher 先撤销旧generation，再nextTick读取最终props，Framework7内部replace使用updateCurrentUrl+新的shallowRef投影。原Gate06外部导航关闭history并保留同owner过滤词策略继续成立。[Gate09](09-vueuse.md) 已验证RAF取消、observer/选区保护，并移除无消费者preserveWhile。
- **真实布局/交互：** [Gate10](10-frontend-design.md) 完成双端、200%缩放、RTL、长文与移动横屏视觉检查；[Gate11](11-web-design-guidelines.md) 完成原生模态、焦点、主题对比和表单语义，且修正native dialog默认cancel下的Escape。当前post-flush template-ref watcher只在实际dialog挂载后showModal，History取消rename同时阻止默认行为。workspace样式以scoped src保留所有权，根选择器已限定以免影响子组件。首轮保留的部分元素选择器属于scoped样式偏好，不是剩余Vue语义缺陷；不为分数作无行为收益的全量样式重写。
- **完整本地证据：** [Gate17](17-final-test-matrix.md) 记录完整lint/typecheck、build、7 files /62 Chromium PASS及Go/三库/迁移结果。[Gate18](18-harness-finalization.md) 记录父审最终Vitest **44 files /38,917 PASS**（包含后续harness新增测试），替代Gate17历史38,907总数；两者不叠加。本轮未冒称独立重跑。
- **容器应用及产物一致性：** [Gate16](16-container-e2e.md) 主应用矩阵 **47 PASS /2 bfcache NOT RUN**，另有真实双端design/accessibility、native跨tab identity与兼容性证据；[Gate15](15-docker-smoke.md) 明确验证镜像限制。Gate18新鲜构建与10项smoke通过，summary确认dist与Linux可执行文件均与Gate16最终产物完全相同（source tree含后续测试，不能声称源码哈希相同）。因此本地应用证据未因harness-only修改失效。

**当前10/10仅指本地适用、已审查的Vue Composition API、响应式数据流、SFC/macros、列表身份、watch/lifecycle和组件契约条件闭环。** 原生bfcache实际pageshow.persisted未命中，两项仍NOT RUN；物理iOS/Android输入/键盘、手工屏幕阅读器、其他浏览器引擎、商业模型质量、stock Dockerfile/Node26及amd64发布构建仍无本门证明。真实storage事件或普通Back不替代bfcache，自动键盘/focus和截图不替代物理输入/读屏。本报告冻结供父审，不把这些外部边界转记PASS。
