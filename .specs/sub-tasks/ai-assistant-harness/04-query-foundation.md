# Step 04：建立 UID 绑定、严格 schema、统一筛选和游标基础

**Task File:** `.specs/tasks/in-progress/ai-assistant-harness.feature.md`
**Phase:** 2
**Model:** gpt-6-astra
**Agent:** general-purpose
**Depends on:** 01, 02
**Parallel with:** 05, 06
**Note:** 高复杂度档（等效 opus）：涉及财务/身份关键域、多层一致性或开放的驱动/并发设计，需要更强推理。 本轮已获用户实施授权；仅实施本步骤及其测试，遵守主任务当前状态与文件所有权。依赖编号指同目录对应文件，接口以主任务 Architecture Overview 及根目录 `AI_ASSISTANT_HARNESS_PLAN.md` §15 为准。
**Goal:** 建立 UID 绑定、严格 schema、统一筛选和游标基础。

实施边界：pkg/services/ai_assistant_queries.go；ai_assistant_filter.go；ai_assistant_cursor.go；ai_assistant_read_helpers.go（新增）；必要时 transactions.go/accounts.go/transaction_categories.go/transaction_tags.go 纯规则抽取。只修改本步骤责任文件；共享协议变更先集中协调，不与其他步骤同时编辑。阅读主任务 Acceptance Criteria、Architecture Overview 及源方案所引章节；不以相邻步骤的摘要替代输入契约。原始工作区改动必须保留。

#### Expected Output

上述路径的可编译实现、针对实际行为的自动测试及必要测试辅助；相关原有功能保持可运行。新增文件名为计划命名，若按项目命名微调须更新实际验证记录。

#### Success Criteria

- 校验/签名/权限基础测试无需 LLM 可跑；后续 07/08/09 只新增专属工具文件，不并发编辑 registry/filter/cursor；任何未实现方法不被旧 API 访问。
- 当前步骤测试使用隔离夹具，不能访问/清空生产数据；未具备外部环境的验证记录为待验证。

#### Subtasks

- [x] 实现 NewAssistantQueryService 与唯一七工具注册白名单/Definitions，执行端严格 JSON 重复键、unknown、null/空数组及字节限制；DTO/schema 单一来源，默认严格按 §15.1。
- [x] 统一时间/金额/枚举/字面关键词、父子展开去重、ID 归属和 UID 分库；同读视图内批量读取元数据；禁止调用会另开 session 的 helper。
- [x] 冻结后续三个查询步骤共享的逻辑交易表示、同侧匹配、ReadView helper、错误/证据 envelope。采用静态工厂装配：本步在 ai_assistant_queries.go 的 NewAssistantQueryService 中一次固定调用交易、聚合、元数据/余额三个包内工厂，各返回该请求作用域的命名执行函数集合；名称重复/缺失校验由唯一 registry 完成。为 07 的 ai_assistant_transactions.go、08 的 ai_assistant_aggregate.go、09 的 ai_assistant_metadata.go/ai_assistant_balances.go 先生成可编译工厂骨架（返回对应名称的显式内部 unsupported），精确签名由本步冻结。07/08/09 随后各替换自己的骨架函数体，不替换共享方法、不改 registry、不用 init 或全局可变注册。基础骨架不暴露新路由；最终装配 owner 为 04 的静态构造器，此后无需再次写共享文件。
- [x] 实现用途隔离 HMAC cursor，绑定 §15.3 全部条件/层级摘要/时效，高水位基础接口；验签后仍验权，续查只产生候选不自动执行。
- [x] 新增 ai_assistant_filter_test.go、ai_assistant_cursor_test.go、ai_assistant_queries_test.go；覆盖重复键、混合用户 ID、金额溢出、秒精度、%/_/反斜线、固定时钟过期/篡改/换secret/层级变更及静态白名单。预置 Definitions→Execute 七名称全量贯通测试结构，基础阶段只测显式 unsupported，Phase 3 由 09 增加正式全工具测试并在并行汇合后运行，任何残留 unsupported 都失败。

- [x] 游标实现明确采用规范 JSON 编码、服务器 secret 派生专用用途的 HMAC-SHA256 与恒定时间 MAC 比较，不复用登录 token。解码后允许只有 cursor；如果同时给其他查询参数，则对显式参数应用同一默认/去重/时间归一化再与绑定条件比较，规范相等接受，任何实质变化返回 INVALID_CURSOR。
- [x] 在 ai_assistant_cursor_test.go 增加 cursor-only 跨请求续查、等价规范条件接受、实质变更拒绝、固定签名向量/不同用途不可互验测试；代码审查确认使用恒定时间比较原语而非字符串比较，不用不稳定耗时断言证明恒时。

- [x] 工具统一执行外壳采集工具耗时、读视图的应用实际读取行数、结果记录数和最终序列化返回字节；错误/空集/字节缩页均记录真实值，避免07/08/09重复埋点。新增观测测试区分DB读取行数、结果行数与字节，不将SQL参数写入记录。
- [x] 新增具名筛选组合矩阵：tag_ids 的 any/all 分别命中；no_tags=true 与非空 tag_ids、孤立 tag_match、空/重复名/缺时间 periods、periods 与 filter 起止时间同时出现、comparison 引用未定义 period 均拒绝；合法组合使用共同规范 filter，不能退化为无筛选。
- [x] 统一读 helper 保持主/关联/子查询的 UID+deleted 条件；用已删除账户/分类/标签/交易及有效隐藏实体的独立夹具证明有效范围与隐藏语义不同。合法有效交易的缺失必需关联仍报一致性错误，不以删除过滤悄悄吞交易。

#### Blockers & Risks

| Type | Item | Impact | Likelihood | Resolution / Mitigation |
| --- | --- | --- | --- | --- |
| Risk | 风险：空 ID 被归一成全部/签名被当授权；显式 [] 拒绝与新读视图重验。阻塞：read-view 接口变化；通过 02 完成交付再适配，不修改底层共享接口。 | High | Medium | 将上述针对性测试作为本阶段验收证据；失败必须修复或明确环境阻塞，不能降低财务/隔离契约。 |
| Blocker | 依赖 01, 02 的已冻结接口与测试产物未完成 | Medium | Low | 等待真实前置接口交付；使用mock开发不能跳过最终集成 |
