# Step 03：加入执行预算并建立客户端可读的非敏感预算配置

**Task File:** `.specs/tasks/in-progress/ai-assistant-harness.feature.md`
**Phase:** 1
**Model:** gpt-6.1-sol
**Agent:** general-purpose
**Depends on:** None
**Parallel with:** 01, 02
**Note:** 常规实现档（等效 sonnet）：契约/路径已冻结，以局部实现与确定性测试为主。 本轮已获用户实施授权；仅实施本步骤及其测试，遵守主任务当前状态与文件所有权。依赖编号指同目录对应文件，接口以主任务 Architecture Overview 及根目录 `AI_ASSISTANT_HARNESS_PLAN.md` §15 为准。
**Goal:** 加入执行预算并建立客户端可读的非敏感预算配置。

实施边界：pkg/settings/setting.go；pkg/api/server_settings.go；src/lib/server_settings.ts；conf/ezbookkeeping.ini；pkg/settings 对应测试。只修改本步骤责任文件；共享协议变更先集中协调，不与其他步骤同时编辑。阅读主任务 Acceptance Criteria、Architecture Overview 及源方案所引章节；不以相邻步骤的摘要替代输入契约。原始工作区改动必须保留。

#### Expected Output

上述路径的可编译实现、针对实际行为的自动测试及必要测试辅助；相关原有功能保持可运行。新增文件名为计划命名，若按项目命名微调须更新实际验证记录。

#### Success Criteria

- 新配置可独立加载，预算只有一个权威来源；旧路由行为保留；前端 accessor 类型检查通过。
- 当前步骤测试使用隔离夹具，不能访问/清空生产数据；未具备外部环境的验证记录为待验证。

#### Subtasks

- [x] 在主 llm_assistant 集中增加 run_timeout/query_timeout/max_rounds，按源 §10.1 默认与正数/溢出规则解析，不放在 fallback 配置；此步保留旧 embedding loader 直至切换清理。
- [x] server settings 仅传运行预算，不传供应商密钥；增加前端 accessor，供后续 JSON/SSE 使用 run_timeout+10 秒。
- [x] 编写 pkg/settings/ai_assistant_runtime_test.go 与 server settings 测试，覆盖默认/0/负数/溢出、EBK/EBKCFP 覆盖和敏感字段不泄漏；当前助手仍可启动。

#### Blockers & Risks

| Type | Item | Impact | Likelihood | Resolution / Mitigation |
| --- | --- | --- | --- | --- |
| Risk | 风险：配置副作用过早移除旧依赖；本步只增加预算，移除 embedding 交由 13。 | High | Medium | 将上述针对性测试作为本阶段验收证据；失败必须修复或明确环境阻塞，不能降低财务/隔离契约。 |
| Blocker | None | None | None | None |
