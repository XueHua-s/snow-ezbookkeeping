# Step 14：完成可执行验收矩阵、性能/真实环境证据及发布状态报告

**Task File:** `.specs/tasks/in-progress/ai-assistant-harness.feature.md`
**Phase:** 5
**Model:** gpt-6-astra
**Agent:** general-purpose
**Depends on:** 13
**Parallel with:** None
**Note:** 高复杂度档（等效 opus）：包含条件性三库索引/schema兼容实现及财务完整性验收，命中 data integrity；不是纯证据整理。 本轮已获用户实施授权；仅实施本步骤及其测试，遵守主任务当前状态与文件所有权。依赖编号指同目录对应文件，接口以主任务 Architecture Overview 及根目录 `AI_ASSISTANT_HARNESS_PLAN.md` §15 为准。
**Goal:** 完成可执行验收矩阵、性能/真实环境证据及发布状态报告。

**Current status (2026-10-15):** 原计划本地实施与验证已完成，Phase 5 独立审查 PASS（9/10）；新增 Docker 整应用构建/HTTP/双页面 E2E 也通过独立证据复核，未以三库 Docker 测试替代。真实主备模型七题仍待环境。最新用户明确授权 commit/push 并通过 gh 向当前 origin `XueHua-s/ezbookkeeping` 的 `main` 建立草稿 PR；该授权覆盖早期提交限制，不含生产发布或旧表删除。持久交付结论见[最终实施审查](../../analysis/harness-final-implementation-review.md)及[验收记录](../../../docs/ai-assistant-harness-verification.md)。

实施边界：docs/ai-assistant-harness-verification.md；pkg/services/ai_assistant_database_integration_test.go；pkg/services/ai_assistant_aggregate_benchmark_test.go；pkg/api/ai_deployment_integration_test.go；必要回归测试。仅在执行计划/10 万条基准证明需要索引时，本步也是该条件性实现的唯一 owner，可修改 pkg/models/transaction.go、相关账务关联模型（例如 pkg/models/transaction_tag_index.go；实际文件名须先 rg 核查）、cmd/database.go 及现有 schema/迁移辅助路径，并新增对应迁移测试；不得扩展为业务字段重构。只修改本步骤责任文件；共享协议变更先集中协调，不与其他步骤同时编辑。阅读主任务 Acceptance Criteria、Architecture Overview 及源方案所引章节；不以相邻步骤的摘要替代输入契约。原始工作区改动必须保留。

#### Expected Output

上述路径的可编译实现、针对实际行为的自动测试及必要测试辅助；相关原有功能保持可运行。新增文件名为计划命名，若按项目命名微调须更新实际验证记录。

#### Success Criteria

- CK-1—20/HR-1—3全部有对应证据或明确待环境验证；独立真值通过；环境缺失不是可发布；完整交接可复核且原用户工作区改动保留。
- 当前步骤测试使用隔离夹具，不能访问/清空生产数据；未具备外部环境的验证记录为待验证。

#### Subtasks

- [x] 独立6笔/2500跨5年/10万数据集、三库同一契约集、权限/数值/分页/取消/连接状态隔离均通过；修复本阶段动态表名拒访测试假阳性，不以共享生产 helper 当唯一 oracle。
- [x] 三库命令/环境变量、opt-in、临时库标识检查及清理入口已交付；双端真实浏览器/实际 API 与模拟 provider 按12清单通过；缺失真实模型输入及命令已记录。
- [ ] 使用真实主备模型、专用隔离测试用户和独立财务真值逐题核对源§12.2七问题；缺真实凭据/专用部署，保持待环境。
- [x] 主任务全部本地 Regular Checks：65个Go文件格式、全仓 test/vet、要求的 race 包、npm lint/test/build、严格CRLF-aware diff check通过；无无关 lint/lockfile/secret/data变化。
- [x] 硬件/DB版本、退出码、fixture、读取计数/内存/耗时/执行计划、真实取消及隔离SQLite旧二进制恢复均记录。三库100K证据不支持新增索引，因此未修改schema或账务数据；条件性新索引迁移/三库降级验证不适用，不能声称已做不存在的迁移。现有schema重复同步和旧缓存哨兵保留通过。
- [x] 文档区分本地完成/待环境/不可发布，未执行不计通过；任务仍in-progress。提交/推送/当前仓库草稿PR依最新明确授权执行，不部署生产。

- [x] 审查并记录工具名/耗时/计数/字节/错误码/request ID、轮次/调用次数、首状态/总耗时、失败/取消原因；正常、多工具、缓存、预算和取消的确定性轨迹/脱敏测试结果已纳入报告。

- [x] 本地真实工具+模拟最终回答同时验证890.00收支净额与1700.00当前净资产、请求/记录时区跨月说明；三库稳定深页、counts_additive=false、删除/标签/期间负例已记录。真实模型回答仍归入上方未完成项，未以模拟文字替代。
- [x] 应用实际读取与计划扫描估计、结果记录与字节分开；JSON首状态null、心跳不计；正常/取消等样例通过canary脱敏。
- [x] 新增Docker整应用构建/HTTP/双页面smoke/E2E均通过；持久[Docker报告](../../../docs/ai-assistant-harness-docker-verification.md)和独立审阅已交付。实际Linux后端+最终dist测试镜像，不宣称stock前端stage重建、实体手机或真实模型验收。

#### Blockers & Risks

| Type | Item | Impact | Likelihood | Resolution / Mitigation |
| --- | --- | --- | --- | --- |
| Risk | 风险：把真实模型连通当财务正确或跳过当通过；逐题核对金额/范围/工具轨迹，分开环境矩阵和发布门禁。 | High | Medium | 将上述针对性测试作为本阶段验收证据；失败必须修复或明确环境阻塞，不能降低财务/隔离契约。 |
| Resolved | 依赖13的接口与测试产物 | None | None | 已稳定交付并完成最终集成；缓存拒访正向控制经独立race复验 |
| Pending environment | 真实主备模型凭据及专用测试部署/用户 | High | Current | 保留七题补测清单和可执行命令；不宣称可发布 |
