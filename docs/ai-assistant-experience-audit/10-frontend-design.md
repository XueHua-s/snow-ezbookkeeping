# 10 · Frontend design 串行审查

2026-10-16；基线 `91081998` 相对 `origin/main`，包含 Gate01–09 未提交修复。初始 **6/10**；首轮修复后 **9/10**；下游证据闭环后的当前本地适用范围 **10/10**（详见文末复审，不代表外部设备/全产品认证）。评分来自本次审查契约，不是技能自带量表。真实移动布局和缩放标题缺陷已关闭；首轮发送按钮对比度/标题全文问题已由 Gate11 修复，并经 Gate16 最终产品镜像验证。下文首轮记录保持原时点，当前结论以文末复审为准；物理移动键盘/手工读屏不计 PASS。

## 依据、方向及组件边界

实读 `skills/frontend-design/SKILL.md`、Vue best practices 与 reactivity/sfc/component-data-flow/composables 四必读参考、project-spec/global-prompt、software-design-philosophy、计划第3–4节、主 audit manifest 及05–09报告；定向回归读取 Vitest/vue-testing-best-practices。设计方向为 **安静的账务阅读工作台**：会话标题与财务正文优先，暖棕仅用于既有主动作，细边框区分账单依据，保持用户字体和字号。没有新增字体请求、依赖、品牌装饰或动画。

职责不变：Toolbar 独占工具栏密度/菜单和焦点；CloudWorkspace 只组合区域并界定根布局；Composer 保持输入高度和发送位置；MessageList/AutoScroll 继续独占阅读锚点。没有增加 props/emits、store 状态或异步接口，没有削弱 scope/confirmation/export/route/selection 保护。

| 技能/计划条款 | 结果 |
| --- | --- |
| Context-specific direction / restraint | 沿用金融产品主题；删除工具栏重复边框噪声，保持清楚的标题、正文、依据层级。 |
| Typography / 用户设置 | font:inherit；正文行高1.7；账单和表格 tabular-nums；长词 anywhere；未替换用户字体。 |
| Spatial composition / 自适应密度 | 桌面宽屏保留历史/新会话文字；≤960px及移动收为44px图标；记忆统一置More。标题使用余下空间，完整title和DOM文本保留。 |
| Theme / states | currentColor图标，继承light/dark前景；hover/active背景、focus-visible轮廓；未增加颜色专用语义。发送按钮现有对比度待下一门精准验证。 |
| Vue / complexity | scoped class；toolbar-mobile避免通用根类冲突；根布局限定.cloud-workspace；所有旧用户意图事件保留。 |
| Plan 4.1–4.3 / concrete review | 实际桌面/移动route、API、SQLite、确定性provider截图；实际查看全部18张，不以overflow断言替代视觉判断。 |

## 具体前后变化

1. 旧 `desktop-zoom200.png` 标题被全长工具按钮挤至约两字符。现在真实200%缩放（1280×900物理像素，640×450 CSS，DPR2）标题宽 **430px**，能完整读出测试会话标题。主工具均有完整aria-label/title，图标aria-hidden，尺寸44×44；DOM顺序保持历史→标题→新会话→More，RTL镜像自然布局。
2. 实际查看第一轮移动截图发现正文被挤成几乎零高。父组件裸`.mobile` scoped规则也作用到子组件根，给Toolbar/Composer施加工作区高度。父复审确认Composer同样受影响；已将父根规则限定`.cloud-workspace.mobile/.desktop`，Toolbar变体采用toolbar-mobile并不参与flex收缩。最终390×844移动工具栏 **53px**，五行草稿时正文 **535px**；568×320横屏正文 **130px**，textarea44px可内滚。
3. 桌面工作区底部预留从固定150px改为96–150px随视高变化的clamp；200%时正文151px，输入框完全可达。900px高仍保持原150px留白。
4. More菜单宽度约束在视口内，最大高度70dvh且内部滚动；记忆入口统一在More，菜单动作仍先关闭、随后进入记忆。原Esc与焦点归还保持。未新增本地化文案。
5. 回归测试使用aria-label定位原生summary More，不再依赖图标按钮的可见文字；memory lifecycle测试先打开菜单。现有E2E flows已经按isVisible判断并打开More，无需加隐藏重复控件。

## 新鲜视觉证据与观察

图片目录：`docs/ai-assistant-ui-review/gate10/`。所有截图来自本次当前代码的真实平台入口、真实router/store/API、临时SQLite与合成账本；后端本次从工作树构建。没有读用户配置或真实账本，环境保留双EBK_/EBKCFP_过滤。

- `desktop-empty/answer/history/memory.png`：主列限宽，空态不挤输入；依据卡片低于正文层级；历史焦点轮廓清晰，记忆设置、关闭和新增有稳定间距。
- `mobile-empty/answer/history/memory.png`：标题单行，正文恢复，输入位于底部；记忆文本自然换行；历史键盘焦点清楚。
- `desktop-dark/rtl/narrow/zoom200.png`：深色前景清楚且无额外发光装饰；RTL动作与文本方向一致；窄屏/200%标题显著恢复，不再只剩两字符。
- `mobile-dark/rtl/small/landscape.png`：320px标题仍有160px可用宽，长标题省略合理；横屏保留可读消息和输入；黑底代码/依据均可辨。
- `desktop-long-focus/mobile-long-focus.png`：长引用自然换行，表格留在消息容器内，New conversation键盘轮廓明确。table computed overflow-x:auto及tabular-nums；本fixture表格实际可换行到容器宽，未声称触发横向滚动条。长文脚本等待Framework7进入布局稳定（正文>300px）后拍摄，避免进入动画中间态截图误判。

`review.json`记录双端剪贴板原Markdown/代码、历史18次Tab包含、Esc焦点归还及记忆关闭焦点；`extended-review.json`记录8种主题/方向/尺寸实测；`long-review.json`记录双端标题、触控尺寸、表格和页面宽度。全部pageerror为0。200%使用隔离扩展chrome.tabs.setZoom(2)，不是CSS zoom；RTL用真实语言preview与移动RTL入口，主题用应用设置。

本人通过view_image查看全部18张接触表，并分别查看移动回答、深色桌面、200%、最终移动长文原图；父另独立查看200%、移动回答/横屏、深色桌面和记忆五张，确认标题及正文恢复。旧截图只作对照。

## 验证及边界

- focused Vitest：4 files / **40 tests PASS**，0 skip/unhandled，`/tmp/gate10-focused.log`；workspace lifecycle、Experience、Confirmation、Export。
- Typecheck PASS：`/tmp/gate10-types.log`。Vue/TS scoped ESLint PASS：`/tmp/gate10-lint-final.log`；脚本在项目ESLint ignore范围，因此另行node --check，不冒称脚本lint覆盖。
- source-lines：1693 files / 0 exceed500；CloudWorkspace仍497行，未压缩绕门。git diff --check PASS。
- 长文真实route定向脚本 `scripts/assistant-experience-design-review.mjs` 双端PASS；环境新增可选answerContent仅控制合成provider正文，默认行为不变。其他真实视觉脚本按既有review/matrix适配在/tmp/gate10-review.mjs与/tmp/gate10-matrix.mjs执行，fixture凭据只在0600临时文件。
- 首轮browser回归13 PASS / 4 FAIL：旧getByText('More')找不到改成图标后的按钮；改成role/button仍不能定位原生summary，最终使用getByLabelText；焦点断言同步检查aria-label。后续中间回归2 PASS/2 FAIL因旧textContent焦点断言，修正后结果见下方最终记录。
- native config ESM warning保留交Gate12，没有隐藏。未运行Docker/全仓矩阵、未commit/push、未修改主manifest、未委派。

Gate11须复核enabled Send按钮：截图中现有暖棕约RGB177/115/63配白字，估算3.89:1，低于普通文字4.5:1；需用computed颜色而非截图估计完成整改。标题完整文本现由原生title提供，移动触屏长标题全文的可发现性也应进一步审查。手工屏幕阅读器、物理iOS/Android键盘、所有语言/字体组合、真实模型质量、最终Docker/E2E/三库与全套测试仍未执行，不以本门截图替代。

最终回归：`/tmp/gate10-browser-complete.log` **2 files / 17 tests PASS，exit 0，0 skip、0 unhandled error**；包含双端原生确认/导出四项和Experience展示十三项。More定位用aria-label，取消后明确断言activeElement的aria-label仍为More。最后新增测试修改后typecheck、scoped ESLint再次PASS，`/tmp/gate10-types-final.log`、`/tmp/gate10-browser-lint-final.log`；source-lines再次1693/0、diff-check PASS。本门冻结交父审及Gate11。


## Gate16–18 后最终证据闭环（2026-10-16）

本次仅复读 Gate11/16/17/18 报告、实际持久化 JSON、当前 Toolbar/Composer 实现并查看必要截图；没有重复无变化测试或修改产品。**当前本地适用范围10/10**：首轮两个未关闭设计问题已具备实际修复和最终运行证据，未发现新的本范围视觉缺陷。评分不把以下 NOT RUN 转成 PASS，也不替代父审。

- **对比度关闭。** 实读 [最终四组合review.json](../ai-assistant-experience-evidence/audit-e2e/accessibility/review.json)：desktop/mobile × light/dark 全部测量均 `disabled:false`，rest/focus **5.3404:1**、hover/pressed **6.6009:1**，filter:none；浅色focus黑outline、深色白outline，均有对应color-scheme。当前Composer从既有主题局部混黑、白字，与JSON一致，不引入字体/主题第二权威。不推断任意用户自定义CSS配色也成立。
- **全文发现性关闭。** 当前Toolbar More顶部实际渲染可换行`.full-title`；最终四组合JSON均 `titleAccessible:true`、`resizeAndNavigation:PASS`、`errors:[]`。查看 [Gate11移动全文截图](../ai-assistant-ui-review/gate11/mobile-light-title.png)，全文在有界菜单顶部多行完整显示；明确这是Gate11截图，最终产品的重新执行由Gate16 JSON证明，不改写截图来源。
- **最终长内容/密度。** 实读 [最终long-review.json](../ai-assistant-experience-evidence/audit-e2e/design/long-review.json)：双端PASS，桌面720px标题510px、移动390px标题230px；三个主控件均44×44；documentWidth等于viewport；表格auto/tabular-nums且错误为空。实际查看 [桌面长文](../ai-assistant-experience-evidence/audit-e2e/design/desktop-long-focus.png) 和 [移动长文](../ai-assistant-experience-evidence/audit-e2e/design/mobile-long-focus.png)：长词换行、正文/输入分区和焦点清楚，无父样式高度污染。表格此fixture没有实际横向溢出，不补写横向滚动测试。
- **主题与模态。** 实际查看 [最终深色发送焦点](../ai-assistant-experience-evidence/audit-e2e/accessibility/desktop-dark-send-focus.png) 和 [最终移动深色历史](../ai-assistant-experience-evidence/audit-e2e/accessibility/mobile-dark-history.png)：暖棕动作保留，轮廓与黑底文字清楚，抽屉布局无新阻断。四组合JSON的modal/outsideFocusBlocked均true；这是自动键盘/原生焦点证据，不是手工读屏。
- **产品身份和下游检查。** design/accessibility各自docker-runtime.json明确执行镜像为 `sha256:bd5b1c31834ad9d917679b9bb9007edae4b9553201a2af48939f6796d1d339e1`。核对Gate16及18的build-provenance和Gate17 provenance：107个前端产物摘要同为 `ba13bcd1e34f10a293a2717b7a48558c524c9ad16340cc07ae9614b971e5eead`，Linux二进制同为 `e5af7ae3ab00bf3930964e0f7c02aacd5b12dca41910effa15ecb967c56c5f63`；Gate18新镜像ID不同，但产品字节相同。保留视觉证据的原镜像身份，不称其在新镜像重新执行。Gate17完整lint/typecheck、42文件单测、7文件62项browser及build已通过；Gate18另记录父最终44文件38917单测通过。本次仅审阅这些下游结果，不计为自己重跑。

首轮RTL、窄屏、横屏及真实200%截图保持当时来源；最终长文/主题证据补充而不虚构全部组合重新拍摄。**NOT RUN继续保留**：物理iOS/Android软键盘、手工屏幕阅读器、全部语言/字体组合、其它浏览器引擎、商业模型质量、真正bfcache命中、stock Dockerfile/Node26发布构建。Gate16的两项bfcache NOT RUN不计通过。当前10/10仅表示本门可在本地审查并有证据的设计要求已闭环；完整外部发布结论仍由父审管理。
