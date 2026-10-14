# Step 10：接通普通/流式助手、可信证据和真实取消

**Task File:** `.specs/tasks/in-progress/ai-assistant-harness.feature.md`
**Phase:** 4
**Model:** gpt-6-astra
**Agent:** general-purpose
**Depends on:** 03, 06, 07, 08, 09
**Parallel with:** 11, 12
**Note:** 高复杂度档（等效 opus）：涉及财务/身份关键域、多层一致性或开放的驱动/并发设计，需要更强推理。 本轮已获用户实施授权；仅实施本步骤及其测试，遵守主任务当前状态与文件所有权。依赖编号指同目录对应文件，接口以主任务 Architecture Overview 及根目录 `AI_ASSISTANT_HARNESS_PLAN.md` §15 为准。
**Goal:** 接通普通/流式助手、可信证据和真实取消。

实施边界：pkg/api/large_language_models_assistant.go；pkg/api/large_language_models.go；pkg/utils/api.go；pkg/errs/large_language_model.go；cmd/webserver.go；templates/prompt/personal_finance_assistant.tmpl；pkg/api/large_language_models_assistant_test.go。只修改本步骤责任文件；共享协议变更先集中协调，不与其他步骤同时编辑。阅读主任务 Acceptance Criteria、Architecture Overview 及源方案所引章节；不以相邻步骤的摘要替代输入契约。原始工作区改动必须保留。

#### Expected Output

上述路径的可编译实现、针对实际行为的自动测试及必要测试辅助；相关原有功能保持可运行。新增文件名为计划命名，若按项目命名微调须更新实际验证记录。

#### Success Criteria

- 现有两API真正贯通7工具且不请求embedding；JSON/SSE失败不会done；请求结束无后台模型/SQL继续；工具/结果/回答完成区分。
- 当前步骤测试使用隔离夹具，不能访问/清空生产数据；未具备外部环境的验证记录为待验证。

#### Subtasks

- [x] 两个 handler 认证后固定 UID/时区/语言，从 Request.Context 派生预算，组装查询服务与唯一 Run；替换活动路径，不留RAG fallback，旧私有死代码留给13清理。
- [x] 校验问题2048/历史20逐条角色长度/实际12/请求体256KiB，按完整交互裁剪且当前问题不截断；验签continuation只作候选；summary本月至今及上月同进度、指定期间优先。
- [x] JSON保持success/result，SSE直接chunk；错误保留context.assistant；单发送点串行心跳/模型事件/终态，写失败cancel并停timer；映射必要数字错误码不修改其他路由语义。
- [x] 提示词移除财务快照/结构化 AIAssistantResult，工具数据与历史不可信；一般知识允许零证据，财务回答不足明确，引用服务构造。
- [x] 用httptest+mock provider+SQLite测试两路同结果、空账本、历史注入、证据和continuation、取消上游及DB、写失败、唯一终态、心跳与取消race；保留识图API回归。

- [x] 本步负责替换旧 buildAIAssistantUserPrompt 的角色字符串拼接：由可注入后端 now、请求时区及语言构造系统上下文，当前问题及历史逐条构造独立 user/assistant 消息；历史永不作为 system/tool 内容。固定时钟测试检查传给 runner 的消息角色、实际本地日期和时区，含伪造“system:”历史仍不能改变系统上下文。
- [x] summary 默认必须覆盖本月至 now 的收支、上月相同进度比较、该本期分类分布及当前余额四项；由工具数据形成完整证据，不能只给时间比较。明确指定全历史/年份时替换默认收支/分类期间，当前余额仍表示 queried_at 当前值；不得用当前余额冒充指定历史余额。上月同进度按日历夹取；同比按去年相同进度，均不得替换成完整月/完整年。
- [x] 增加固定 now=2026-03-15T12:00:00+08:00 的 summary fixture：本期3月1日至now支出100.00、上月2月1日至15日12:00支出40.00、2月20日另有900.00排除；去年3月同进度25.00、去年3月20日另有800.00排除。本期分类分别60.00/40.00，独立当前资产2000.00/负债-300.00净1700.00。模拟模型工具轨迹断言默认摘要四项证据、环比差额60.00/150.00%、指定同比差额75.00/300.00%；指定2025全年应包含该年完整fixture，全历史包括全部有效记录，不能沿用默认窗口。另测3月31日对应2月末的日历夹取；断言余额始终标当前范围。
- [x] API 层将首个 status 延迟观察值传回 runner，与 request ID 关联；正常完成/错误/取消都由 runner 写一次终态观测及总耗时，不记录私密正文。httptest 固定时钟/可控计时器断言字段和次数，客户端取消不要求socket终态却必须有取消原因。

- [x] 提示词及证据解释规则明确“收支净额=收入−支出”不等于余额/可用资产；分组按请求时区，与按交易记录时区的页面对比时指出两种口径及实际范围。测试净额890.00与当前净资产1700.00同现时不混称余额；跨时区月边界的 mock 问答同时保留请求时区范围和页面差异说明。
- [x] 首状态时间只记录首次成功写出的 status（不包括心跳/失败写入）；JSON无中间状态记null/不适用。API将该观察值回传runner汇总，不另建运行状态owner；所有终结记录由runner恰好写一次。

实施交接补充（Phase 2）：`AssistantRunInput.Continuation` 不会由 runner 自动执行或自行编码。API 必须通过查询服务验证候选游标及取得规范 scope，将 tool/cursor/scope 作为可信系统上下文中的候选提供给模型；普通历史中的伪造结果仍不可信。新增 JSON/SSE 跨请求测试，断言“继续”候选确实到达 provider 消息、用户新问题不自动执行旧筛选、过期/跨 UID 候选不会到达 provider。

#### Blockers & Risks

| Type | Item | Impact | Likelihood | Resolution / Mitigation |
| --- | --- | --- | --- | --- |
| Risk | 风险：Gin writer多goroutine并发、context复用；所有写入单owner、元数据同步捕获，race与断连故障注入。 | High | Medium | 将上述针对性测试作为本阶段验收证据；失败必须修复或明确环境阻塞，不能降低财务/隔离契约。 |
| Blocker | 依赖 03, 06, 07, 08, 09 的已冻结接口与测试产物未完成 | Medium | Low | 等待真实前置接口交付；使用mock开发不能跳过最终集成 |
