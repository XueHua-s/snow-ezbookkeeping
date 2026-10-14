# Step 01：建立可编译的中立工具与 API 契约，冻结并行实现边界

**Task File:** `.specs/tasks/in-progress/ai-assistant-harness.feature.md`
**Phase:** 1
**Model:** gpt-6-astra
**Agent:** general-purpose
**Depends on:** None
**Parallel with:** 02, 03
**Note:** 高复杂度档（等效 opus）：新增跨 llm/provider/models/frontend 的共享契约和内部观测协议，命中 shared contract changes。 本轮已获用户实施授权；仅实施本步骤及其测试，遵守主任务当前状态与文件所有权。依赖编号指同目录对应文件，接口以主任务 Architecture Overview 及根目录 `AI_ASSISTANT_HARNESS_PLAN.md` §15 为准。
**Goal:** 建立可编译的中立工具与 API 契约，冻结并行实现边界。

实施边界：pkg/llm/data/large_language_model_data.go；pkg/llm/provider/large_language_model_provider.go；pkg/llm/assistant_contract.go（新增）；pkg/models/large_language_model.go；src/models/large_language_model.ts；pkg/services/testdata/ai_assistant_truth.json（新增）。只修改本步骤责任文件；共享协议变更先集中协调，不与其他步骤同时编辑。阅读主任务 Acceptance Criteria、Architecture Overview 及源方案所引章节；不以相邻步骤的摘要替代输入契约。原始工作区改动必须保留。

#### Expected Output

上述路径的可编译实现、针对实际行为的自动测试及必要测试辅助；相关原有功能保持可运行。新增文件名为计划命名，若按项目命名微调须更新实际验证记录。

#### Success Criteria

- 既有 GetJsonResponse/StreamTextResponse 调用仍编译；无 llm→services 依赖；六笔预期与源 §15.7 一致；接口精确定义可供并行步骤 mock。
- 当前步骤测试使用隔离夹具，不能访问/清空生产数据；未具备外部环境的验证记录为待验证。

#### Subtasks

- [x] 按源 §4、§7、§15.1—15.3 和 Architecture 定义 ToolDefinition/ToolCall/ToolResult/ConversationMessage/TurnResponse、可选 RunToolTurn、ToolExecutor/EventSink/RunInput/Result；只增加类型，不切现有入口。
- [x] 冻结 HTTP queries/continuation、状态与错误详情结构；与工具 snake_case 分开；共享接口文件在本步后由集成步骤单一负责，后续并行步骤不得改动。
- [x] 建立源 §15.7 六笔手算夹具及独立预期，不依赖生产聚合函数。
- [x] 编写 pkg/llm/data/assistant_contract_test.go、pkg/models/assistant_contract_test.go，验证字符串 ID/金额、序列化外壳及中立接口兼容；使用原 Vue 类型检查证明前端新增类型可编译。

- [x] 在 assistant_contract.go 冻结仅内部使用的轻量观测记录/回调：request ID、工具名、应用实际读取行数、结果记录数、序列化字节、时长、实际轮次/执行次数、缓存命中及受控原因枚举；不进 HTTP queries，不携带 SQL/参数/账务正文/provider 私有状态。由 runner 汇总、API只报告首状态写入，保持单一状态 owner；契约测试覆盖 JSON 无首状态的 null/不适用语义。

#### Blockers & Risks

| Type | Item | Impact | Likelihood | Resolution / Mitigation |
| --- | --- | --- | --- | --- |
| Risk | 风险：将 provider 私有状态塞入公共 DTO；隔离为适配器请求内状态。 | High | Medium | 将上述针对性测试作为本阶段验收证据；失败必须修复或明确环境阻塞，不能降低财务/隔离契约。 |
| Blocker | None | None | None | None |
