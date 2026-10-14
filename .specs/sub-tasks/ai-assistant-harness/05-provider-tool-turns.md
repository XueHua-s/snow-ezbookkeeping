# Step 05：实现三种 OpenAI 工具轮次协议与请求内状态

**Task File:** `.specs/tasks/in-progress/ai-assistant-harness.feature.md`
**Phase:** 2
**Model:** gpt-6-astra
**Agent:** general-purpose
**Depends on:** 01
**Parallel with:** 04
**Note:** 高复杂度档（等效 opus）：新增跨 provider/container/httpclient 的工具轮次及私有状态子系统，命中 new subsystem 与协议状态设计。 本轮已获用户实施授权；仅实施本步骤及其测试，遵守主任务当前状态与文件所有权。依赖编号指同目录对应文件，接口以主任务 Architecture Overview 及根目录 `AI_ASSISTANT_HARNESS_PLAN.md` §15 为准。
**Goal:** 实现三种 OpenAI 工具轮次协议与请求内状态。

实施边界：pkg/llm/provider/openai/ 下工具 adapter 新文件及既有两链路；pkg/llm/large_language_model_provider_container.go；pkg/llm/provider/common/common_http_large_language_model_provider.go；pkg/httpclient/http_request_context.go；pkg/httpclient/http_client.go（核查 CustomHttpResponseLog/HTTP context 包装，按需修改并保留其他 HTTP 功能回归）。只修改本步骤责任文件；共享协议变更先集中协调，不与其他步骤同时编辑。阅读主任务 Acceptance Criteria、Architecture Overview 及源方案所引章节；不以相邻步骤的摘要替代输入契约。原始工作区改动必须保留。

#### Expected Output

上述路径的可编译实现、针对实际行为的自动测试及必要测试辅助；相关原有功能保持可运行。新增文件名为计划命名，若按项目命名微调须更新实际验证记录。

#### Success Criteria

- 模拟上游三协议能交付同一中立完整 TurnResponse；失败轮不出现可执行调用；既有 provider 测试通过且无财务日志。
- 当前步骤测试使用隔离夹具，不能访问/清空生产数据；未具备外部环境的验证记录为待验证。

#### Subtasks

- [x] 实现可选 RunToolTurn，分别支持 Chat Completions JSON/强制 stream、Responses；schema strict required/nullable 在 provider 边界适配。
- [x] 按 call ID/index/item 组装完整参数；只有有效终结且无截断时交付 ToolCall，正文+工具轮不提前执行；Responses store=false 并保留必要 output/reasoning items。
- [x] 向 runner 提供主/备用工具能力，不在 provider 私自做请求级切换；中立上下文转换失败明确 unsupported；保留识图结构化接口。
- [x] 助手上游 body/单事件/参数缓冲有界，HTTP/log context 不丢取消，助手财务正文/私有状态不写日志。
- [x] 新增 openai_assistant_tools_test.go 与现有 adapter/container 回归：三路径分片/多调用/坏JSON/重复ID/断流/截断/取消/能力不支持，保留其他 provider 与识图测试。

- [x] Chat 多候选流先按 choice index 隔离，再按 call ID/index 组装；首期选定一个有效候选并在整轮保持不变，其他候选不能污染参数、正文或触发工具。添加交错两个 choices、相同 call index 不同参数、未选候选错误/截断测试，断言只返回选定候选且不跨候选拼接。
- [x] 新工具 HTTP 路径继续使用该 provider 的 LargeLanguageModelAPIRequestTimeout；整体 deadline 额外生效，两者取最早截止。用可控 mock 上游分别验证 provider timeout 短于/长于 run_timeout，取消后上游调用停止。
- [x] 将供应商明确的 context-length/context-window 错误归一为可识别的上下文预算错误，交 runner 输出未完成；保留原始规范化工具对，不擅自删消息或转为空财务回答。加入普通/流式该错误的模拟协议测试。
- [x] 协议测试检查 ConversationMessage 独立角色在 Chat 与 Responses 均正确编码；包含“system:”文本的历史仍是 user/assistant 内容，备用模型中立重建也不能提升权限。

#### Blockers & Risks

| Type | Item | Impact | Likelihood | Resolution / Mitigation |
| --- | --- | --- | --- | --- |
| Risk | 风险：Responses 不透明状态错误重放；隔离适配器状态、跨 provider 只中立重建。 | High | Medium | 将上述针对性测试作为本阶段验收证据；失败必须修复或明确环境阻塞，不能降低财务/隔离契约。 |
| Blocker | 依赖 01 的已冻结接口与测试产物未完成 | Medium | Low | 等待真实前置接口交付；使用mock开发不能跳过最终集成 |
