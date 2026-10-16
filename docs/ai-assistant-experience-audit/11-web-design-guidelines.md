# 11 · Web Interface Guidelines 串行审查

2026-10-16；`91081998` 相对 `origin/main`，包含 Gate01–10 未提交修复。初始 **7/10**；首轮 **9/10**；下游证据闭环后当前本地适用范围 **10/10**（详见文末；手工读屏、物理移动键盘等仍 NOT RUN）。这是审查契约评分，并非该技能自带评分。软件设计复杂度本门 **9/10**：主题仍由平台 CSS 变量拥有，原生 dialog 承担模态隔离，无主题 observer、全局 modal manager 或新增 store 接口；该首轮判断保留历史语义；当前本地结论以文末复审为准，不表示整体产品或外部设备认证。

## 来源与范围

实读 `.agents/skills/web-design-guidelines/SKILL.md`；通过 curl 新鲜抓取
<https://raw.githubusercontent.com/vercel-labs/web-interface-guidelines/main/command.md>，时间 **2026-10-15T21:25:15Z**（上海 2026-10-16 05:25:15）。内容 SHA-256 `d246b026f4f29b5823a9cc857f9edf3d2507002e055e32040cadeaf3b38e0234`；本次原文保存在 `/tmp/gate11-web-guidelines.md`。

另实读 Vue best practices 及 reactivity/sfc/component-data-flow/composables 四参考、frontend-design、software-design-philosophy、project-spec/global-prompt、计划第4节、主 audit manifest 和05–10报告；测试读取 Vitest/vue-testing-best-practices 及 blackbox、async、browser-vs-node 参考。范围包括 common/assistant experience 组件、Markdown、双端适配器、WorkspacePanels 与关联 store/scroll 生命周期；旧 legacy chat 与整站框架不作重写。

实施前边界：Toolbar 负责已有 More 中的全文展示；Composer 负责输入描述与按钮色；Workspace/Panels 负责真正模态抽屉及响应式切换；Confirmation/Export 负责自身语义。CloudWorkspace 的原 scoped CSS 移到 `assistant-workspace.css`，仍通过 `<style scoped src>` 加载，布局所有权不变，避免超500行或压缩代码。无新增产品文案，全部复用现有翻译键。

## 发现（最终行号；均已修复）

- `src/components/common/assistant/AssistantExperienceComposer.vue:164` — enabled 主按钮普通文字对比不足；由平台主题色局部混入25%黑、白字，hover/pressed混入35%黑；移除整体brightness，避免前景一起变暗。当前支持的双端light/dark实际computed验证见下。
- `src/components/common/assistant/AssistantExperienceComposer.vue:184` — 白色按钮focus轮廓在浅背景不清楚；改CanvasText，配合workspace实际color-scheme。输入框outline:none有composer-row:focus-within替代，不是无焦点样式。
- `src/components/common/assistant/AssistantWorkspaceToolbar.vue:53` — 长标题原来只靠鼠标title；现有More顶部显示完整可换行标题，无额外弹层、状态或动作，触屏点击和键盘均可发现。
- `src/components/common/assistant/AssistantCloudWorkspace.vue:195` / `src/views/base/assistant/useAssistantWorkspacePanels.ts:38` — 原aria-modal抽屉只inert cloud-main，应用顶栏仍可能聚焦；窄屏使用原生showModal的dialog，宽屏保持aside。浏览器top layer覆盖应用所有兄弟区域；resize替换元素，不靠异步close事件回写状态。
- `src/components/common/assistant/AssistantConversationHistory.vue:38` / `:52` — 原Escape只stopPropagation不足以阻止原生dialog的默认cancel；preventDefault保留第一次退出rename、第二次关闭历史，取消编辑后回到该行More焦点。
- `src/components/common/assistant/AssistantConfirmationDialog.vue:53` — 对话框缺accessible name/description；增加Confirm名称与实际待确认文本描述，仍默认聚焦Cancel，原generation/scope保护不变。
- `src/components/common/assistant/AssistantExportDialog.vue:87` — 对话框缺accessible name；补Export名称，原生select显式surface背景，确认/导出/引用dialog添加overscroll containment与有界高度。
- `src/components/common/assistant/AssistantExperienceComposer.vue:76` / `src/components/common/assistant/AssistantMemoryManager.vue:68` — 输入缺name/autocomplete与超限错误关联；补唯一useId、aria-invalid/aria-describedby，保留输入和原有IME/发送策略。历史搜索、rename、summary、复制文本、版本/格式选择也补name。
- `src/components/common/assistant/AssistantCopyButton.vue:39` — 复制成功仅按钮文字变化；补polite播报，内容不在live区域，避免读整段回答。
- `src/components/common/assistant/AssistantRecoveryPanel.vue:104` — 恢复草稿连续长词缺换行；补overflow-wrap:anywhere。
- `src/components/common/assistant/assistant-workspace.css:119` — 助手动作补touch-action、明确tap highlight与hover反馈；`:131`按平台现有dark class设置局部color-scheme，原生控件和CanvasText随主题变化，不修改全站主题。

## 规则逐类核对与适用边界

| 规则组 | 本范围结论 |
| --- | --- |
| Accessibility / semantic actions | 图标按钮具aria-label/title，装饰SVG aria-hidden；button/summary/label/select原生键盘语义；确认/导出/引用及抽屉有名称。Markdown链接使用真实a、安全scheme；外部图片显示链接，不加载图片，因此图片尺寸/懒加载不适用。既有应用壳的全站skip-navigation/heading层级不在本门改造范围，未宣称整站WCAG认证。 |
| Focus / modal | focus-visible或输入组focus-within可见；真正模态阻止背景焦点；历史关闭恢复入口，rename退出恢复行More，确认恢复Toolbar More；跨owner/route的旧操作保护继续回归。 |
| Forms | 明确label/name、non-auth autocomplete off；无paste拦截；checkbox与label同hit target。超限/空输入禁用保留既有“不提交不可接受问题”的产品规则，错误关联而非照搬“任何时候都启用submit”。确认Cancel autofocus是删除安全默认且不弹软键盘。 |
| Async / live / copy | 历史、搜索、消息分页、记忆、引用、导出有loading/empty/error/retry或相应取消；生成只播报state字符串而非token正文；错误与恢复说明持续存在，不只用颜色或toast。未把代码检查说成手工读屏。 |
| Motion / typography | 本次无动画、transition:all、平滑强制滚动或autoplay；浏览器验证设置reducedMotion=reduce。沿用字体和字号、金额tabular-nums、长词换行。既有翻译字典Loading等文案未做全语言标点重写。 |
| Content / layout | More全文自然换行、菜单有界滚动；消息/代码/表格局部overflow和min-width:0保留；触控动作44px、composer safe-area保留。Gate10的RTL/200%/长表格证据作为既有基线，本门没有冒充重新执行它们。 |
| Performance | 不新增font/CDN/preload请求；无渲染路径布局读取。会话/历史窗口既有90条硬上限，未机械加入content-visibility/虚拟化破坏Gate09的消息锚点/选择几何；这属于有界阅读窗口取舍，不能称无限列表性能证明。 |
| Navigation / data | 真实conversation/message/version route同步与返回由既有适配器拥有；短期More/编辑状态不写URL。草稿持久化及page-only beforeunload保护保留；删除仍确认。未改变scope、export、owner/epoch、store或selection契约。 |
| Dark / locale / hydration | 两平台light/dark真实computed color-scheme及按钮状态通过；日期既有toLocaleString/toLocaleDateString、金额本地化helper保留，来源证据保留原时间/时区文本以便核对。客户端应用，无SSR hydration；模型/代码不引入自动翻译层。meta theme-color仍由平台现有入口拥有。 |
| Media / drag | 无音视频、GIF、拖拽或gesture-only动作；无需新媒体替代/拖拽inert。 |

## 真实页面证据

`scripts/assistant-experience-accessibility-review.mjs` 使用现有隔离environment，真实Go服务端、临时SQLite、合成账本、确定性provider、真实Vue Router/Framework7入口；未读取真实配置，继续过滤EBK_/EBKCFP_。Chromium四组 desktop/mobile × light/dark，均 **PASS，pageerror=0**；`docs/ai-assistant-ui-review/gate11/review.json` 保存computed颜色、Canvas解析RGB、filter、ratio、color-scheme、outlineColor和modal结果。

| enabled Send状态（四组合一致） | 前景 | 背景（Canvas量化RGB） | 对比度 |
| --- | --- | --- | --- |
| rest / focus | 255,255,255 | 149,94,54 | **5.34:1** |
| hover / pressed | 255,255,255 | 129,82,47 | **6.60:1** |

真实computed背景为CSS `color(srgb …)`，通过Canvas转换至sRGB后计算WCAG线性亮度；filter=none，测量时disabled=false。未把disabled 0.55 opacity算作失败或PASS。使用项目真实#c67e48主题权威；不宣称任意外部CSS配色覆盖仍保证4.5。已检查现有getContrastTextColor，其接受hex；本组件主题由CSS变量提供，采用局部color-mix避免新增DOM observer或第二份主题状态。

实际操作证据：四组合都用原生More打开全文，文本与标题完全相同；窄屏history `:modal=true`，真实页面外可见控件调用focus被浏览器阻止，18次Tab始终留在dialog。第一次Escape退出rename并回summary，第二次关闭且回history入口；720/390→1280宽侧栏→390模态→New conversation导航释放dialog。组件回归另验证卸载后外部焦点恢复。

共12张新截图：每组 `*-send-focus`、`*-title`、`*-history`。本人实际view_image查看全部12种画面（最终轮核心focus/深色/标题/抽屉逐张检查）：浅色黑outline、深色白outline清楚；暖棕保留；移动More全文多行可读；桌面抽屉遮罩覆盖顶栏，移动抽屉全屏；输入区保持可达。没有把Playwright键盘、DOM focus或截图当作物理键盘/手工屏幕阅读器证明。

## 定向验证与冻结

- focused：6 files / **68 tests PASS**，0 skip/unhandled，`/tmp/gate11-focused-final.log`（Workspace lifecycle、Experience、Confirmation、Export、VueContracts、Workspace commands）。
- Chromium component：2 files / **18 tests PASS**，0 skip/unhandled，`/tmp/gate11-browser-final.log`；新增真实top-layer外部焦点、宽窄切换、关闭/卸载回归。
- 真实应用四组合：`/tmp/gate11-real-final.log` + 持久化review.json，exit 0。中途一次脚本严格定位New conversation遇到背景同名按钮，改为明确作用域.cloud-history；一次color-scheme审查发现scoped/global selector写法没有应用到dark，改完整global主题选择器并新增computed断言，最终四组合通过。未隐藏这些中间失败。
- `npx vue-tsc --noEmit`、13个修改Vue/TS scoped ESLint PASS：`/tmp/gate11-types-final.log`、`/tmp/gate11-lint.log`；脚本另node --check PASS，未冒称被ESLint覆盖。
- source-lines：1695文件 / 0超过500，`/tmp/gate11-lines.log`；git diff --check PASS。没有压缩控制流规避上限。
- Vite native-config ESM warning仍保留交Gate12。未运行Docker、全仓矩阵或三库，未子委派、未改master manifest、未commit/push。

本门冻结交父审。最终Docker/E2E/单元与三库阶段仍待完成；手工读屏、物理iOS/Android输入、跨浏览器及真实模型质量未验证。Gate12接续Vite/build/service-worker审查。


## Gate16–18 后最终本地闭环（2026-10-16）

仅核读最终证据、当前相关实现与 Gate10 视觉复核；没有修改产品、重复无变化测试或改动主报告。**本门当前本地适用范围10/10**：首轮已证实的对比度、触屏全文、模态/焦点及输入语义问题均已修复，原待执行的本地容器/测试矩阵已补齐；本次未发现新的本范围阻塞。此评分保留前文已说明的规则适用取舍，不宣称整站所有 Web Interface Guidelines、WCAG 或全部 R01–R40 均获认证。

- **16个真实颜色状态。** 实读 [最终 accessibility/review.json](../ai-assistant-experience-evidence/audit-e2e/accessibility/review.json)：desktop/mobile × light/dark 四组合，各含 rest/hover/focus/pressed；全部 disabled=false、filter=none，rest/focus **5.3404:1**，hover/pressed **6.6009:1**。focus浅色黑outline、深色白outline及color-scheme相符；当前Composer主题变量/color-mix实现未漂移。不从结果推断任意外部CSS配色都满足标准。
- **全文、键盘与模态。** 四组合均 titleAccessible=true、topLayer.modal/outsideFocusBlocked=true、resizeAndNavigation=PASS、errors为空。复读当前 runner 的实际断言：完整标题与Toolbar文本一致且可见；外部可见控件focus被阻止；18次Playwright Tab在dialog内；rename第一次Escape退出编辑并回summary，第二次关闭回history入口；窄→1280px aside→390px dialog→New conversation导航释放模态。不是根据“有aria-modal”推断背景隔离，也不是手工读屏证据。
- **最终视觉与历史来源。** 本次实际查看 [最终桌面深色发送焦点](../ai-assistant-experience-evidence/audit-e2e/accessibility/desktop-dark-send-focus.png) 和 [最终移动深色历史](../ai-assistant-experience-evidence/audit-e2e/accessibility/mobile-dark-history.png)：焦点轮廓和文字清楚、暖棕保留、抽屉无新增布局阻断。核读Gate10最终长内容/标题复核记录。首轮12张、RTL/200%/横屏截图仍保持原来源，不声称最终镜像重新执行所有旧组合。
- **产品字节对应。** accessibility/docker-runtime.json记录执行镜像 `sha256:bd5b1c31834ad9d917679b9bb9007edae4b9553201a2af48939f6796d1d339e1`。分别核读Gate16和Gate18 build-provenance：107个前端文件摘要均为 `ba13bcd1e34f10a293a2717b7a48558c524c9ad16340cc07ae9614b971e5eead`，Linux executable均为 `e5af7ae3ab00bf3930964e0f7c02aacd5b12dca41910effa15ecb967c56c5f63`；Gate17 provenance前端摘要也一致。Gate18镜像ID不同但产品字节相同，因此保留Gate16证据，未将它重新标成Gate18执行。
- **下游矩阵已核读。** Gate17报告及实际summary.json确认lint/typecheck/build、42文件38907单测、7文件62 browser通过；Go正常/race、三库及迁移结果明确区分opt-in skip和另行执行的性能项。Gate18 summary.json确认父最终44文件38917单测及新镜像10项smoke通过。本门仅审阅归属清晰的结果，不把它们计作自己的重跑，也不因总测试量推断无障碍覆盖比例。

**NOT RUN继续保留：** 手工屏幕阅读器、物理iOS/Android键盘/触摸、其它浏览器引擎、全部语言/字号组合、商业模型质量、实际bfcache命中、stock Dockerfile/Node26发布构建。Gate16两项bfcache未命中不计PASS；Playwright键盘、computed颜色和截图不能替代这些外部验收。首次9/10的记录和当时未执行事项未删除；当前10/10仅关闭本门可在本地核验的适用条件，整体发布结论仍归父审。

本次只更新此报告；未子委派、未启动运行环境、未提交。证据闭环完成并冻结。
