# Step 13：完成旧向量运行代码/配置退役和升级回滚文档

**Task File:** `.specs/tasks/in-progress/ai-assistant-harness.feature.md`
**Phase:** 5
**Model:** gpt-6.1-sol
**Agent:** general-purpose
**Depends on:** 10, 12
**Parallel with:** None
**Note:** 常规实现档（等效 sonnet）：契约/路径已冻结，以局部实现与确定性测试为主。 本轮已获用户实施授权；仅实施本步骤及其测试，遵守主任务当前状态与文件所有权。依赖编号指同目录对应文件，接口以主任务 Architecture Overview 及根目录 `AI_ASSISTANT_HARNESS_PLAN.md` §15 为准。
**Goal:** 完成旧向量运行代码/配置退役和升级回滚文档。

实施边界：pkg/services/ai_assistant_embeddings.go；pkg/api/large_language_models_assistant.go；pkg/api/large_language_models.go；pkg/settings/setting.go；cmd/initializer.go；conf/ezbookkeeping.ini；pkg/api/ai_deployment_integration_test.go；docs/ai-providers.md；docs/ai-storage-migration.md；两套技能副本。只修改本步骤责任文件；共享协议变更先集中协调，不与其他步骤同时编辑。阅读主任务 Acceptance Criteria、Architecture Overview 及源方案所引章节；不以相邻步骤的摘要替代输入契约。原始工作区改动必须保留。

#### Expected Output

上述路径的可编译实现、针对实际行为的自动测试及必要测试辅助；相关原有功能保持可运行。新增文件名为计划命名，若按项目命名微调须更新实际验证记录。

#### Success Criteria

- 无embedding可启动并完成工具问答；旧坏配置或旧密钥缺失不再访问文件；保留旧表及回滚说明；所有技能不再要求本助手top-k。
- 当前步骤测试使用隔离夹具，不能访问/清空生产数据；未具备外部环境的验证记录为待验证。

#### Subtasks

- [x] 清零引用后删除 embedding service、知识窗口/hash/topk/snapshot/similarityScore死代码与失效测试；不删除 models/ai_assistant_embedding.go 或 cmd/database.go schema注册与真实记录。
- [x] 停止旧embedding配置段解析/校验/EBKCFP读取，清理getter/字段/initializer脱敏列表；保留主备/开关/新预算，旧无效段不影响启动。
- [x] 改部署测试为无embedding工具检查，保留识图子测试；新增无段/无效段/缺旧密钥文件/无网络embedding与无缓存读写断言，旧表schema保留回归。
- [x] 更新docs配置/能力/代理超时/旧表保留/备份与回滚；同步 .codex/skills/ezbookkeeping-ai-integration 和 skills/ezbookkeeping-ai-integration 各SKILL.md与references/ai-feature-map.md，以及两份项目结构 references。
- [x] 做定向rg审计引用并写迁移断言；使用实际package-lock/go.mod，不改用户pnpm-lock等既有内容。

#### Blockers & Risks

| Type | Item | Impact | Likelihood | Resolution / Mitigation |
| --- | --- | --- | --- | --- |
| Risk | 风险：删除通用字段误伤其他功能；先全仓引用调查再删，仅清助手专用依赖，识图/其他provider回归。 | High | Medium | 将上述针对性测试作为本阶段验收证据；失败必须修复或明确环境阻塞，不能降低财务/隔离契约。 |
| Blocker | 依赖 10, 12 的已冻结接口与测试产物未完成 | Medium | Low | 等待真实前置接口交付；使用mock开发不能跳过最终集成 |
