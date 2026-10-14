# Step 07：实现全历史逻辑交易明细、详情与稳定分页

**Task File:** `.specs/tasks/in-progress/ai-assistant-harness.feature.md`
**Phase:** 3
**Model:** gpt-6-astra
**Agent:** general-purpose
**Depends on:** 04
**Parallel with:** 08, 09
**Note:** 高复杂度档（等效 opus）：财务逻辑交易归一、去重及关联一致性，命中 data integrity。 本轮已获用户实施授权；仅实施本步骤及其测试，遵守主任务当前状态与文件所有权。依赖编号指同目录对应文件，接口以主任务 Architecture Overview 及根目录 `AI_ASSISTANT_HARNESS_PLAN.md` §15 为准。
**Goal:** 实现全历史逻辑交易明细、详情与稳定分页。

实施边界：pkg/services/ai_assistant_transactions.go；pkg/services/ai_assistant_transactions_test.go（新增）。只修改本步骤责任文件；共享协议变更先集中协调，不与其他步骤同时编辑。阅读主任务 Acceptance Criteria、Architecture Overview 及源方案所引章节；不以相邻步骤的摘要替代输入契约。原始工作区改动必须保留。

本步只替换 04 已创建的 ai_assistant_transactions.go 工厂骨架体，保持签名和名称不变。

#### Expected Output

上述路径的可编译实现、针对实际行为的自动测试及必要测试辅助；相关原有功能保持可运行。新增文件名为计划命名，若按项目命名微调须更新实际验证记录。

#### Success Criteria

- 不经 LLM 能定位任意有效历史并读完整详情；所有页去重逻辑集合等于独立fixture；matched_count 与范围真实。
- 当前步骤测试使用隔离夹具，不能访问/清空生产数据；未具备外部环境的验证记录为待验证。

#### Subtasks

- [x] 仅实现 04 预留 query_transactions/get_transaction 挂点，不编辑共享 registry/filter；遵循 §4.2—4.4、§6、§15.1—15.3。
- [x] 转入解析为同 UID 转出规范ID/time；同侧账户/币种/金额联合匹配、逻辑去重、标签/附件数量投影，缺失关联明确错误，合法无分类为 null。
- [x] 同读视图计数/分页，高水位取实际匹配最大键包含未来交易；时间/金额排序稳定并满足币种/转账侧约束；按结果字节缩页不跳行，异常单条 RESULT_TOO_LARGE。
- [x] 测试五年2500条/未来日期、静态全页无重漏、source/destination/either、退款、隐藏/父子、跨用户关联、游标修改/层级变化、大小预算；断言不会输出附件路径/IP或similarityScore。

- [x] 明细/详情负例包含已删除交易与关联元数据：deleted 交易不可达且不计匹配总数；有效隐藏数据仍可达；有效交易缺失必需关联显式失败。另以多标签交易验证 any/all/no_tags 与聚合相同筛选集合，非法组合不执行查询。

#### Blockers & Risks

| Type | Item | Impact | Likelihood | Resolution / Mitigation |
| --- | --- | --- | --- | --- |
| Risk | 风险：先对物理记录分页再去重导致缺页；逻辑归一与筛选必须在分页之前。 | High | Medium | 将上述针对性测试作为本阶段验收证据；失败必须修复或明确环境阻塞，不能降低财务/隔离契约。 |
| Blocker | 依赖 04 的已冻结接口与测试产物未完成 | Medium | Low | 等待真实前置接口交付；使用mock开发不能跳过最终集成 |
