# Step 11：实现共享可靠流解析、状态、历史与续查

**Task File:** `.specs/tasks/in-progress/ai-assistant-harness.feature.md`
**Phase:** 4
**Model:** gpt-6-astra
**Agent:** general-purpose
**Depends on:** 01, 03
**Parallel with:** 10
**Note:** 高复杂度档（等效 opus）：新UTF-8/SSE增量解析和终态状态机、取消旧回调竞争，命中 new subsystem 与 concurrency。 本轮已获用户实施授权；仅实施本步骤及其测试，遵守主任务当前状态与文件所有权。依赖编号指同目录对应文件，接口以主任务 Architecture Overview 及根目录 `AI_ASSISTANT_HARNESS_PLAN.md` §15 为准。
**Goal:** 实现共享可靠流解析、状态、历史与续查。

实施边界：src/lib/services.ts；src/lib/assistant_stream.ts（按需抽取）；src/views/base/assistant/AssistantPageBase.ts；对应 .test.ts。只修改本步骤责任文件；共享协议变更先集中协调，不与其他步骤同时编辑。阅读主任务 Acceptance Criteria、Architecture Overview 及源方案所引章节；不以相邻步骤的摘要替代输入契约。原始工作区改动必须保留。

#### Expected Output

上述路径的可编译实现、针对实际行为的自动测试及必要测试辅助；相关原有功能保持可运行。新增文件名为计划命名，若按项目命名微调须更新实际验证记录。

#### Success Criteria

- 解析故障绝不完成；错误范围/续查不丢；mock传输证明共享状态隔离；旧页面在新composable默认值下仍可编译。
- 当前步骤测试使用隔离夹具，不能访问/清空生产数据；未具备外部环境的验证记录为待验证。

#### Subtasks

- [x] 移除fallbackToDoneEvent，增量UTF8/SSE parser处理CRLF/多行/合法无空行尾帧，跟踪terminalSeen；损坏数据和EOF无终态失败，未知非终态可忽略，重复终态拒绝。
- [x] 保持错误context.assistant及queries/continuation；JSON和SSE截止run_timeout+10秒，Abort取消与旧UUID隔离。
- [x] composable维护 requesting/completed/incomplete/failed/canceled、唯一请求标识；取消/清空/离开/换用户清理及中止；历史12且未完成内容不当可信事实，Unicode历史截取并标省略。
- [x] 续查入口只在页面内存，最新问题决定是否使用；下一页发送明确消息；暴露两端所需scope/reference/status，不新增全局store。
- [x] 写 src/lib/assistant_stream.test.ts、src/views/base/assistant/AssistantPageBase.test.ts：UTF8截断/坏JSON/EOF/CRLF/多行/重复终态、错误元数据、取消后旧回调、用户切换/清空/超时、历史和候选续查；按Vue测试技能采用仓库既有mock设施。

#### Blockers & Risks

| Type | Item | Impact | Likelihood | Resolution / Mitigation |
| --- | --- | --- | --- | --- |
| Risk | 风险：组件测试环境不足；优先mock纯parser与composable依赖，必要设施最小添加并记录，不引入全新测试栈。 | High | Medium | 将上述针对性测试作为本阶段验收证据；失败必须修复或明确环境阻塞，不能降低财务/隔离契约。 |
| Blocker | 依赖 01, 03 的已冻结接口与测试产物未完成 | Medium | Low | 等待真实前置接口交付；使用mock开发不能跳过最终集成 |
