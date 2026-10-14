# Step 02：实现真实取消和固定主库只读一致性读适配

**Task File:** `.specs/tasks/in-progress/ai-assistant-harness.feature.md`
**Phase:** 1
**Model:** gpt-6-astra
**Agent:** general-purpose
**Depends on:** None
**Parallel with:** 01, 03
**Note:** 高复杂度档（等效 opus）：涉及财务/身份关键域、多层一致性或开放的驱动/并发设计，需要更强推理。 本轮已获用户实施授权；仅实施本步骤及其测试，遵守主任务当前状态与文件所有权。依赖编号指同目录对应文件，接口以主任务 Architecture Overview 及根目录 `AI_ASSISTANT_HARNESS_PLAN.md` §15 为准。
**Goal:** 实现真实取消和固定主库只读一致性读适配。

实施边界：pkg/core/context_assistant.go；pkg/datastore/query_context.go；pkg/datastore/assistant_read_view.go；pkg/datastore/database.go；对应 _test.go。只修改本步骤责任文件；共享协议变更先集中协调，不与其他步骤同时编辑。阅读主任务 Acceptance Criteria、Architecture Overview 及源方案所引章节；不以相邻步骤的摘要替代输入契约。原始工作区改动必须保留。

#### Expected Output

上述路径的可编译实现、针对实际行为的自动测试及必要测试辅助；相关原有功能保持可运行。新增文件名为计划命名，若按项目命名微调须更新实际验证记录。

#### Success Criteria

- SQLite 本地真实驱动验证取消停止且连接复用无 query_only 污染；三库同连接/隔离证据可复核，缺外部库列待验证，不无声降级；公共 services/API/定时任务回归可跑。
- 当前步骤测试使用隔离夹具，不能访问/清空生产数据；未具备外部环境的验证记录为待验证。

#### Subtasks

- [x] 新增 NewAssistantContext(parent,requestID,locale,clientIP)，修复 XOrmContextAdapter 父 context 委托及 SessionID；不向 WebContext 添加字段。
- [x] 从现有 engineGroup.Master().DB().DB 固定 sql.Conn，WithAssistantReadView 使用单 Tx；MySQL/PostgreSQL BeginTx LevelRepeatableRead+ReadOnly，方言/前缀/占位符适配集中在 datastore。
- [x] SQLite 保存 query_only→开启→BeginTx→执行业务→先确保 Tx Commit/Rollback 完成→以独立有界清理 context 恢复 query_only→Close；取消触发 sql 自动回滚时须等待终结，禁止回滚/恢复并发或先还连接。无法证明可复用或恢复失败时 Conn.Raw 返回 driver.ErrBadConn 丢弃连接。
- [x] 仅暴露参数化 QueryContext/QueryRowContext 和方言元数据；各异常阶段清理 Tx/Conn；保留原 DoTransaction 默认，不在现有独立 session helper 内拼快照。
- [x] 新增 core/context_assistant_test.go、datastore/query_context_test.go、datastore/assistant_read_view_test.go：nil/Value/deadline、写保护、并发修改快照、取消长查询、回滚与 query_only 恢复顺序/竞态/失效连接丢弃/下个借用者状态；MySQL/PG 默认 READ COMMITTED 也验证实际 RR。三库测试入口显式 opt-in、临时库标识及清理。

- [x] 读视图内部采集数据库操作时长和应用实际读取行数（与结果记录数分开命名，不冒称数据库优化器扫描行数）；提供内部只读观察值供04工具层汇总，不引入对 llm/services 的反向依赖。固定夹具测试多SQL/空集/取消读取的实际计数，无SQL或参数输出。

#### Blockers & Risks

| Type | Item | Impact | Likelihood | Resolution / Mitigation |
| --- | --- | --- | --- | --- |
| Risk | 风险：sqlite 自动回滚与清理竞态导致池污染；顺序握手、独立清理 deadline、故障注入及 ErrBadConn 丢弃。阻塞：外部库缺环境；完成 SQLite/本地代码并提供 opt-in 补测命令，不宣称三库通过。 | High | Medium | 将上述针对性测试作为本阶段验收证据；失败必须修复或明确环境阻塞，不能降低财务/隔离契约。 |
| Blocker | None | None | None | None |
