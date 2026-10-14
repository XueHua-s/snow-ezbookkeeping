# Step 08：实现完整集合精确聚合与期间比较

**Task File:** `.specs/tasks/in-progress/ai-assistant-harness.feature.md`
**Phase:** 3
**Model:** gpt-6-astra
**Agent:** general-purpose
**Depends on:** 04
**Parallel with:** 07, 09
**Note:** 高复杂度档（等效 opus）：涉及财务/身份关键域、多层一致性或开放的驱动/并发设计，需要更强推理。 本轮已获用户实施授权；仅实施本步骤及其测试，遵守主任务当前状态与文件所有权。依赖编号指同目录对应文件，接口以主任务 Architecture Overview 及根目录 `AI_ASSISTANT_HARNESS_PLAN.md` §15 为准。
**Goal:** 实现完整集合精确聚合与期间比较。

实施边界：pkg/services/ai_assistant_aggregate.go；pkg/services/ai_assistant_aggregate_test.go；pkg/services/ai_assistant_aggregate_benchmark_test.go（新增）。只修改本步骤责任文件；共享协议变更先集中协调，不与其他步骤同时编辑。阅读主任务 Acceptance Criteria、Architecture Overview 及源方案所引章节；不以相邻步骤的摘要替代输入契约。原始工作区改动必须保留。

本步只替换 04 已创建的 ai_assistant_aggregate.go 工厂骨架体，保持签名和名称不变。

#### Expected Output

上述路径的可编译实现、针对实际行为的自动测试及必要测试辅助；相关原有功能保持可运行。新增文件名为计划命名，若按项目命名微调须更新实际验证记录。

#### Success Criteria

- 总计与独立预期一致，无1800/历史/扫描上限；内存不持有所有账单；组分页不改总计；超时不伪造 all_matching。
- 当前步骤测试使用隔离夹具，不能访问/清空生产数据；未具备外部环境的验证记录为待验证。

#### Subtasks

- [x] 仅实现 04 的 aggregate 挂点，按 §4.5、§6、§15.2，共用 canonical/filter/read helper，禁止新写另一套筛选。
- [x] 完整集合批量读取或已证明安全的数据库聚合；big.Int有符号累加，分币种/flow，逻辑matched_count与组count可加性明确；取消/超时不给部分全量总计。
- [x] periods同读视图、ordinal配对、日历夹取、非正基数null、精确两位舍入、缺失组合补零；有限时间桶分页生成、不造笛卡尔积；分组页不影响totals。
- [x] 测试六笔手算真值、退款极值/空minmax、转账两流一笔、重叠period并集、DST闰年、负/零基数与未配对桶；2500条独立参考值与10万条benchmark，记录内存/扫描行数/时间；三库SUM精度不可假设。

- [x] 明确输出 counts_additive=false；同一转账命中双方、按账户/flow分成多个组时，各组 count 之和可大于逻辑 matched_count，测试字段存在且为false，matched_count仍去重，不能用组数相加作为逻辑笔数。
- [x] 补聚合排序与游标矩阵：order_by 只能使用 group_key 或已选 metrics；各指标 asc/desc 同值追加完整分组键，多币种分别排名。构造跨页并列值及同名不同ID组，按指标遍历深页静态数据无重漏、totals不变；下一请求账本修改后重算返回新的 queried_at，不能宣称游标是快照。
- [x] 使用04具名组合夹具验证 any/all/no_tags 结果与07一致，已删除交易不进入 totals/count/groups，隐藏有效交易仍纳入；验证非法 period/filter/comparison 组合不会执行聚合。

#### Blockers & Risks

| Type | Item | Impact | Likelihood | Resolution / Mitigation |
| --- | --- | --- | --- | --- |
| Risk | 风险：分组状态无界或SQL SUM先溢出；用精确批处理、执行计划与大集合测量，不能靠裁时间/明细数解决。 | High | Medium | 将上述针对性测试作为本阶段验收证据；失败必须修复或明确环境阻塞，不能降低财务/隔离契约。 |
| Blocker | 依赖 04 的已冻结接口与测试产物未完成 | Medium | Low | 等待真实前置接口交付；使用mock开发不能跳过最终集成 |
