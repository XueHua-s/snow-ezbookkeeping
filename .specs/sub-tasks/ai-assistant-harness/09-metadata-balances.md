# Step 09：实现全量元数据发现和当前叶子账户余额

**Task File:** `.specs/tasks/in-progress/ai-assistant-harness.feature.md`
**Phase:** 3
**Model:** gpt-6-astra
**Agent:** general-purpose
**Depends on:** 04
**Parallel with:** 07, 08
**Note:** 高复杂度档（等效 opus）：余额、叶子父子去重、分币种总计的财务正确性，命中 data integrity。 本轮已获用户实施授权；仅实施本步骤及其测试，遵守主任务当前状态与文件所有权。依赖编号指同目录对应文件，接口以主任务 Architecture Overview 及根目录 `AI_ASSISTANT_HARNESS_PLAN.md` §15 为准。
**Goal:** 实现全量元数据发现和当前叶子账户余额。

实施边界：pkg/services/ai_assistant_metadata.go；pkg/services/ai_assistant_balances.go；对应 _test.go（新增）。只修改本步骤责任文件；共享协议变更先集中协调，不与其他步骤同时编辑。阅读主任务 Acceptance Criteria、Architecture Overview 及源方案所引章节；不以相邻步骤的摘要替代输入契约。原始工作区改动必须保留。

本步只替换 04 已创建的 ai_assistant_metadata.go/ai_assistant_balances.go 工厂骨架体，保持签名和名称不变。另由本步独占新增 pkg/services/ai_assistant_registry_integration_test.go：以 NewAssistantQueryService 实际构造，遍历 Definitions 七名称并通过 Execute 输入各自合法夹具，断言没有内部 unsupported；该全工具测试在 Phase 3 的 07/08/09 全部汇合后执行，本步开发期间可单测四个自有工具，不添加对 07/08 的代码依赖。

#### Expected Output

上述路径的可编译实现、针对实际行为的自动测试及必要测试辅助；相关原有功能保持可运行。新增文件名为计划命名，若按项目命名微调须更新实际验证记录。

#### Success Criteria

- 7工具中的全部元数据/余额工具独立可运行，页遍历完整，返回业务字段和真实范围；不自动换汇或重建历史余额。
- 当前步骤测试使用隔离夹具，不能访问/清空生产数据；未具备外部环境的验证记录为待验证。

#### Subtasks

- [x] 实现 list_accounts/list_categories/list_tags 与 get_account_balances 的 04 挂点；ID升序稳定分页，包含隐藏/父层/标签组，字面关键词与统一cursor。
- [x] 余额来自有效叶子 signed_balance，负债 outstanding=-signed，父节点不重复；分币种资产/负债/净资产 totals 从完整范围产生，不由当前页推导。
- [x] 测试超过一页各类实体、同名层级、空账本、跨用户混合ID、隐藏账户、父子重复、资产2000/负债-300净1700及溢缴、页变化不改totals。
- [x] 测试及业务实现仅在此步骤专属文件，不能并发编辑 04 helper；发现契约缺口记录交给阶段集成处理。

- [x] 补已删除账户/分类/标签与有效隐藏实体对照：元数据各页不返回deleted，余额不计deleted叶子；有效隐藏实体仍可达且完整总计纳入，不能将hidden等同deleted。

#### Blockers & Risks

| Type | Item | Impact | Likelihood | Resolution / Mitigation |
| --- | --- | --- | --- | --- |
| Risk | 风险：父节点余额和叶子相加重复；只对有效叶子累加并验证显式父ID展开。 | High | Medium | 将上述针对性测试作为本阶段验收证据；失败必须修复或明确环境阻塞，不能降低财务/隔离契约。 |
| Blocker | 依赖 04 的已冻结接口与测试产物未完成 | Medium | Low | 等待真实前置接口交付；使用mock开发不能跳过最终集成 |
