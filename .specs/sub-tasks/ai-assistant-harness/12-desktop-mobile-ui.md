# Step 12：双端统一呈现范围、状态、引用与下一页

**Task File:** `.specs/tasks/in-progress/ai-assistant-harness.feature.md`
**Phase:** 4
**Model:** gpt-6.1-sol
**Agent:** general-purpose
**Depends on:** 11
**Parallel with:** 10
**Note:** 常规实现档（等效 sonnet）：契约/路径已冻结，以局部实现与确定性测试为主。 本轮已获用户实施授权；仅实施本步骤及其测试，遵守主任务当前状态与文件所有权。依赖编号指同目录对应文件，接口以主任务 Architecture Overview 及根目录 `AI_ASSISTANT_HARNESS_PLAN.md` §15 为准。
**Goal:** 双端统一呈现范围、状态、引用与下一页。

实施边界：src/views/desktop/assistant/AssistantPage.vue；src/views/mobile/assistant/AssistantPage.vue；src/locales/；对应页面测试。只修改本步骤责任文件；共享协议变更先集中协调，不与其他步骤同时编辑。阅读主任务 Acceptance Criteria、Architecture Overview 及源方案所引章节；不以相邻步骤的摘要替代输入契约。原始工作区改动必须保留。

#### Expected Output

上述路径的可编译实现、针对实际行为的自动测试及必要测试辅助；相关原有功能保持可运行。新增文件名为计划命名，若按项目命名微调须更新实际验证记录。

#### Success Criteria

- 两端可完成追问/范围展示/续查/引用/取消相同行为；所有新文案走i18n；npm类型/测试/build可跑。
- 当前步骤测试使用隔离夹具，不能访问/清空生产数据；未具备外部环境的验证记录为待验证。

#### Subtasks

- [x] 读取 frontend-design/vue-best-practices/vue 测试相关技能，沿用现有视觉组件与安全Markdown；仅渲染11共享状态不复制执行逻辑。
- [x] 增加进度、范围/完整性/读取时间、未完成/失败/取消、续查与鉴权引用展示；不展示tool原参数或cursor原文；尊重hide_amount。
- [x] 各语言资源新增文案与可用回退；页面卸载/用户切换接共享取消；确认桌面移动同一行为。
- [x] 写两页面可用的交互测试或组件渲染断言：状态、范围、下一页触发、hide_amount、恶意Markdown/链接不绕过安全渲染；准备双端真实smoke操作清单。

#### Blockers & Risks

| Type | Item | Impact | Likelihood | Resolution / Mitigation |
| --- | --- | --- | --- | --- |
| Risk | 风险：双端生命周期不同导致仅一端取消；共同composable加两端离开场景断言。 | High | Medium | 将上述针对性测试作为本阶段验收证据；失败必须修复或明确环境阻塞，不能降低财务/隔离契约。 |
| Blocker | 依赖 11 的已冻结接口与测试产物未完成 | Medium | Low | 等待真实前置接口交付；使用mock开发不能跳过最终集成 |
