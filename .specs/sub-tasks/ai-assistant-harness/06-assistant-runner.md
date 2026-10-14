# Step 06：实现唯一请求级循环、预算、fallback 与证据归集

**Task File:** `.specs/tasks/in-progress/ai-assistant-harness.feature.md`
**Phase:** 2
**Model:** gpt-6-astra
**Agent:** general-purpose
**Depends on:** 01, 05
**Parallel with:** 04
**Note:** 高复杂度档（等效 opus）：新执行循环、预算状态机、取消/事件终结竞争，命中 new subsystem 与 concurrency。 本轮已获用户实施授权；仅实施本步骤及其测试，遵守主任务当前状态与文件所有权。依赖编号指同目录对应文件，接口以主任务 Architecture Overview 及根目录 `AI_ASSISTANT_HARNESS_PLAN.md` §15 为准。
**Goal:** 实现唯一请求级循环、预算、fallback 与证据归集。

实施边界：pkg/llm/assistant_runner.go；pkg/llm/assistant_runner_test.go。只修改本步骤责任文件；共享协议变更先集中协调，不与其他步骤同时编辑。阅读主任务 Acceptance Criteria、Architecture Overview 及源方案所引章节；不以相邻步骤的摘要替代输入契约。原始工作区改动必须保留。

#### Expected Output

上述路径的可编译实现、针对实际行为的自动测试及必要测试辅助；相关原有功能保持可运行。新增文件名为计划命名，若按项目命名微调须更新实际验证记录。

#### Success Criteria

- 用 mock executor 可独立运行多轮完整会话；状态与调用次数确定；无 llm→services import；不依赖 embedding。
- 当前步骤测试使用隔离夹具，不能访问/清空生产数据；未具备外部环境的验证记录为待验证。

#### Subtasks

- [x] 按已冻结契约实现 AssistantRunner.Run，依赖 ToolExecutor 而非 services；整轮成功后顺序调用，完整调用结果配对再请求模型。
- [x] 集中计数 12轮/24调用、64KiB结果/128KiB请求、重复成功仅复用一次、连续两次相同错误退出；一轮额度不足整轮不执行，重试切备计轮次。
- [x] 请求级主备最多切一次，正文/可展示思考发出后禁止切换；status 不阻止；失败轮不执行；取消/权限/参数错误不触发供应商切换。
- [x] 最终无工具轮完成才发送正文；EventSink 失败取消；汇集成功 queries/最多8去重引用与最后候选continuation；不持久化状态、不把工具错误当空结果。
- [x] 写 scripted provider/executor 测试：纠错、多工具、超额前置拦截、缓存/重复ID异参、fallback 前后事件、预算未完成、context取消、正文缓冲、证据排除错误结果、一般知识零工具。

- [x] 上游显式上下文超限映射 BUDGET_EXHAUSTED/incomplete，保留已确认 queries/continuation；不通过删减工具对重试，不触发无数据财务回答。测试超限发生于已有成功工具之后，断言证据保留、工具对不丢失、不会继续模型请求或切备。
- [x] runner 独占请求运行观测汇总：request ID、工具名、工具耗时/计数/返回字节/错误码、模型轮次、工具调用次数、失败或取消原因。仅记录元数据；缓存命中与实际执行次数区分。日志捕获测试同时断言所需字段存在与账单/余额/备注/提示词/游标/私有状态均不存在。

- [x] 汇总01内部观测约定，实际provider轮次包括重试/切备，实际工具执行次数不包含缓存命中（命中另计），工具DB读取行数与结果行数分别记录；受控原因枚举区分预算、取消、超时、provider、协议和数据库失败，禁止直接记录 error.Error()。整体耗时由唯一runner计时并在所有退出路径终结一次。
- [x] 用固定时钟/捕获器验证成功、空集、多轮切备、缓存复用、预算、DB超时、客户端取消、SSE写失败的次数/字节/首状态及一次终结；植入账单备注/余额/密钥/游标/私有状态敏感canary，所有观测输出均不得出现。

#### Blockers & Risks

| Type | Item | Impact | Likelihood | Resolution / Mitigation |
| --- | --- | --- | --- | --- |
| Risk | 风险：成功工具对被裁剪或同 ID 重复执行；测试配对与规范参数缓存，超限明确 incomplete。 | High | Medium | 将上述针对性测试作为本阶段验收证据；失败必须修复或明确环境阻塞，不能降低财务/隔离契约。 |
| Blocker | 依赖 01, 05 的已冻结接口与测试产物未完成 | Medium | Low | 等待真实前置接口交付；使用mock开发不能跳过最终集成 |
