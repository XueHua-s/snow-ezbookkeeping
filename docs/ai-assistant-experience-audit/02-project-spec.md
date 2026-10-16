# 02 · Project spec 串行审查

**当前结论：本地适用门 10/10，已闭环。最终复审及证据见文末；下文初审分数与待验描述保留为历史。**

2026-10-16；基线 `91081998` 相对 `origin/main`，包含 gate01 后续修复。
依据：完整读取 `skills/ezbookkeeping-project-spec/SKILL.md`、`references/global-prompt.md`，以及 `.codex/skills` 的同内容副本。已读取总审查契约和 gate01 报告，运行 `rg --files cmd pkg src conf docs skills` 确认布局。

**初始 8/10；修复后 9/10。代码层面的本轮已确认缺陷已修复；整个 spec 门仍待 gate01 D4/D5 可读性整改和最终质量阶段复审，不声称 10/10 或发布通过。** 无子 agent、commit/push、Docker 操作或用户数据修改；没有访问私有配置或凭证。

## 适用要求与源代码证据

| 要求 | 阅读路径与结论 |
| --- | --- |
| 小接口、深模块、阅读优先 | `ai_assistant_storage{,_commands,_dependencies,_operations,_quota}.go`；事务回调、回执与额度结算统一在 service；HTTP 不复制存储政策。D6 本轮统一 memory DTO；gate01 D4/D5 的 UI/store 可读性问题尚未消失。 |
| 配置/常量单一权威 | `pkg/core/assistant_limits.go`、`pkg/settings/setting_ai_experience.go`、`setting_load.go` diff、`pkg/api/large_language_models_assistant_conversations{,_runtime}.go`、`ai_assistant_runs_policy.go`、`conf/ezbookkeeping.ini` diff。默认值/验证在 settings；capabilities 输出有效 limits；runtime 统一注入 config；接收时保存任务 policy。发现两处硬编码撤销数量 4，已修复。provider/模型具体协议由 AI 门继续审，不把本次配置路径检查算作完整 provider 审查。 |
| 数据安全、撤销/幂等/epoch | `ai_assistant_storage_commands.go` 在时间/epoch 检查前查回执，变更和回执同事务；`ai_assistant_cleanup.go` 清空时递增 deletion/memory epochs、取消任务、结算/归零并 tombstone 回执；`ai_assistant_conversations_invalidation.go` 撤销派生来源；`ai_assistant_memories{,_management}.go` 使用 version/epoch。`ai_assistant_runs_{claim,execute,access}.go` 检查 owner/lease、依赖并在过期时 interrupted 而非重放。 |
| 保留与物理清理 | `ai_assistant_cleanup_{sweep,users}.go`、`users_access.go` diff。按 UID/shard 有界清理，保留最小回执/用户 epoch；禁用/删除先 fence。清理与模型服务可用性分离。实际24小时目标与部署规模仍需运行验证/监控。 |
| 迁移/兼容/可回滚 | `cmd/database.go` diff 为新增五表 SyncStructs，旧 embedding 注册保留；阅读 `ai_assistant_{conversation,memory,user_state,command_receipt}.go` 模型；`docs/ai-storage-migration.md` 规定全分库备份、禁止混合写入、保留新表回滚及恢复后租约 fencing。修复文档“新版无长期运行状态表”的过时断言。未执行实际备份恢复或跨库升级。 |
| 保护普通记账路径 | `pkg/api/data_managements_clear.go`、`ai_assistant_storage_operations.go`：clear-all 仅首次受理者执行既有逐服务删除，重复未完成返回409，密码不进回执。已读取只清账单的预先上下文失效路径；跨服务并发的完整覆盖交给 AI/E2E 门，不能宣称跨库原子事务。 |
| 双端产品行为对齐 | 阅读 desktop/mobile `assistant/AssistantConversation.vue`，两端使用 `AssistantCloudWorkspace.vue`，主要差异是 Vue Router/Framework7 生命周期与账单打开方式；检查共享 workspace capabilities/limits 绑定及 `useAssistantCloudWorkspace.ts` 能力状态。静态复用支持一致性，但不替代双端真实路由/键盘/焦点验证。无前端修改。 |
| 后端契约优先、文档同改、可逆 | 本轮只增加纯投影函数并替换两调用点、纠正两处撤销上限和文档；没有 schema/配置新增、迁移脚本或数据重写。存储记录不变，可回退代码。 |
| 验证优先 | 本轮 focused Go regression、gofmt、diff check；技能要求的全仓 Go test/vet、npm lint/test 按用户串行安排由最终阶段执行。不能把未执行项记作 PASS。 |

审查粒度：以上为明确读过的关键路径；不是逐行审查 PR 每一个文件。后续专门 skill 门继续独立检查其职责。

## 修复与契约

1. **D6：记忆公开投影收敛。** 新增 `ai_assistant_memories_projection.go`，列表和 JSON 导出调用同一函数。公开 DTO 使用当前 schema 常量，空/JSON null 来源规范为数组；会话级依赖保留于 Sources，但不会生成空 SourceMessageIds。列表仍跳过损坏记录；导出仍返回 storage_unavailable。调用方保留各自 UID 查询、分页、写流及错误语义。文档同步说明。新增公共 List/Export 契约回归，含旧行 schema=0、空来源、null、会话级与消息级混合来源、损坏来源。
2. **撤销数量不应固定为默认并发4。** `assistantRevokeMemoryRuns` 以及 `InvalidateSources` 的间接 memory 失效阶段，都改以事务中的 `RunningCount + QueuedCount` 有界查询。依据：ReserveRun 同事务递增 queued；MarkRunRunning 原子转移 queued→running；SettleRun 按持久 ReservationSettled 防重复并减对应计数；cancel_requested 尚未结算仍计入 running。因此该上限覆盖已接收任务，也不会因配置调小漏掉旧任务。两个6任务回归分别验证关闭记忆/来源重验后全部 canceled、reservationSettled=true、QueuedCount=0。
3. **迁移文档纠错。** 把早期无状态版本和当前持久化版本明确区分；回滚保留五表，不再指导把新历史误认为只在页面内。

## 定向验证

- `gofmt -w pkg/services/ai_assistant_memories.go pkg/services/ai_assistant_memories_management.go pkg/services/ai_assistant_memories_projection.go pkg/services/ai_assistant_memories_projection_test.go pkg/services/ai_assistant_conversations_export.go pkg/services/ai_assistant_conversations_invalidation.go`：完成。
- `go test ./pkg/services -run 'TestAssistantMemory(ListExportContract|RevocationBeyondDefaultConcurrency)$' -count=1`：PASS，2 个顶层测试/3 个实际叶子场景。
- 红灯验证：临时把两个撤销上限恢复4，运行 `go test ./pkg/services -run 'TestAssistantMemoryRevocationBeyondDefaultConcurrency$' -count=1`：两个子场景均按预期 FAIL（仍 queued），随后 finally 恢复修复；日志 `/tmp/gate02-revocation-red.log`。
- `go test -json ./pkg/services -run 'TestAssistant(Memory|StorageDatabaseContracts|ConversationExportBatchesKeepQuestionVersionGroups)' -count=1 > /tmp/gate02-focused-tests.json`：exit 0；**10 个实际叶子场景 PASS，14 个外部数据库叶子场景 SKIP，0 FAIL**。Go JSON 有19条 pass事件，包括父级聚合，不能把这些当作19个实际测试；尤其 MySQL/Postgres 父级 pass 下的子项全部 skip，不证明数据库兼容通过。
- 首次新增测试因测试 fixture 使用手写 JSON 列名而失败（xorm acronym mapper）；改为读取记录后 AllCols 更新，避免错误列名。此为测试构造修正，不把初始失败隐藏成产品缺陷。
- `git diff --check`：PASS。

## 后续门交接与复审条件

- gate01/05/07/10：D4、D5 仍开放，修完后 project-spec 阅读优先要求需复审。
- AI integration：审查完整工具 UID/精度/供应商与运行依赖；建议针对清账单预先 invalidate 与后续逐服务删除之间新任务交错补充证明，本轮不据静态代码声称跨服务并发已验证，也未证明存在越权读写。
- structure：确认新的纯 projection 文件符合服务层组织，文档早期无状态叙述已更正。
- 最终串行阶段：执行全仓 `go test ./...`、`go vet ./...`、`npm run lint`、`npm run test`，Go race/三数据库实际服务、重复迁移/备份恢复/旧表数据对比、Docker restart/lease、双端 E2E。并对本轮撤销测试及 D6 再验。结果齐备且 D4/D5 关闭后方可给整个 spec 门满分；本轮不以授权分阶段运行测试为理由伪造通过。

## 最终闭环复审（Gate15–18 后）

2026-10-16。本节保留并取代上文“9/10、待最终阶段”的当前状态判断；历史发现、修复和当时 SKIP 记录不改写为已通过。**当前本地实现及验证范围最终评分 10/10，Gate02 关闭。** 这表示本 skill 适用的已审实现要求和本地质量门已满足，不是外部发布认证或零缺陷保证。

本轮完整阅读 Gate15–18 报告（长输出的 Gate16 尾段另读），并检查 `audit-matrix/{summary,migration}.json`、`audit-harness-finalization/summary.json` 及实际测试日志。独立重算 normal/services/race、vet、lint、browser、migration 七份原始日志 SHA-256，均与 matrix summary 一致；父最终 `/tmp/ebk-parent-final-unit.log` 哈希与 Gate18 summary 一致。未重复执行无变化的完整测试。

### 先前未关闭条件的实际闭环

- **D4/D5 阅读优先：** 实读当前 `AssistantCloudWorkspace.vue` 的 dialog 组合点、`AssistantConfirmationDialog.vue`、`AssistantExportDialog.vue`、`stores/assistant.ts`、`assistant/{session,commands}.ts`。确认/导出各自拥有取消、异步票据、上下文失效及 UI 生命周期；store 组合身份、会话与持久命令领域模块，重要 await 后 guard 清晰可见。不是只靠压缩行数或搬运同一段代码。Gate01 最终复审已关闭 D4–D6，Gate18 后亦关闭工具生命周期 D8；本轮未发现新的阅读优先必须修改项。
- **D6 与配置权威：** 重读 `ai_assistant_memories_projection{,_test}.go`。单一公开投影与原错误策略保留；测试现在真实参数化三库。独立解析 final services/race 原始 JSON，各有 **9 个实际 memory 叶子 PASS**，不计父级聚合、不用此前14个数据库 SKIP 替代。固定4上限修复仍有效。
- **账本清空交错：** 重读 `ai_assistant_storage_ledger.go`、storage 的 `withUserTransactionVersion/userTransaction`、`transactions_clear_context.go` 及 `transactions_delete.go` 两类批量入口。账本变更与必要上下文失效在同一用户分库事务内，guard 只在成功 Commit 后发布当前操作的版本证明；首次或任意助手写入介入后重扫，失败/回滚不发布证明。原先“只预先 invalidate”的待证边界已由这条提交边界修复。独立从 normal/services 和 race 日志确认，各有 **42 个 ledger 叶子 PASS**，包含交错、回滚、部分失败、同 turn 再生成、复用及扫描规模。并不把既有跨服务 clear-all 描述成全业务库原子事务。
- **迁移与回滚：** `migration.json` 三引擎均 PASS；每个引擎的 legacy-before/upgraded 哈希相同，before/restored/after-old-rollback/reupgraded 哈希相同。Gate17 实际执行旧版初始化→两次升级→原生备份恢复→旧二进制回滚→再升级→过期租约恢复；不再只有空库新增表证据。清理记录无剩余私有 fixture 配置/备份。迁移文档当前持久化说明仍适用。
- **后续兼容修复：** `session.checkIdentity` 只对当前 ticket/epoch/disposal 有效的 capabilities404 设置旧后端状态；同步与初始化共享该判断，避免把 history404 误当版本缺失。与 Gate16 实际旧 main 双端兼容2场景及4个竞态回归一致。
- **测试工具的数据/资源边界：** 实读 `assistant_migration_runtime.py`、`assistant-bounded-command.mjs`、`assistant-build-resources.mjs`。两种应用环境覆盖前缀统一过滤；POSIX 子进程组有期限、结束清理和输出上限；Docker 资源使用事先登记的随机确切名称，未知清理失败保留错误而不假称不存在。无共享资源 prune 或用户数据迁移。

### 已完成本地质量证据及范围

Gate17 已执行 `go test ./...`（1746实际叶子PASS，4个明确 opt-in SKIP）、`go vet ./...`、`npm run lint`、浏览器7文件62项、生产 build；服务最终重跑441叶子PASS，race754叶子PASS。最初跳过的性能/大导出后来单独执行6+1场景通过；商业 deployment 测试仍未运行，不能计入通过。Gate18 后父独立 `npm run test` 为 **44文件 / 38,917 PASS**，替代此前42文件38,907的最终前端单元状态。gofmt 和 source-line 证据仍有效；本轮仅改本报告，`git diff --check` 通过。

容器当前构建/冒烟10场景通过；Gate16 主业务双端 **47 PASS / 2 NOT RUN**、账单引用10 PASS、真实旧后端2 PASS，以及实际 proxy/cancel/原生身份切换支持双端行为一致性。Gate18 重新 build/smoke 后 dist 与 Linux executable 哈希同 Gate16；因此保留原完整业务 E2E 证据是明确的产物等价复用，不声称在新镜像重跑了所有业务场景。

### 仍然公开保留的外部发布限制

验证镜像是 Go/Alpine 缓存基础镜像加真实入口及 UID/GID1000，前端由 host Node22 构建；**stock Dockerfile/Node26 阶段、stock runtime 镜像、amd64、其他浏览器引擎未验证**。商业主备模型质量、真实手机软键盘、人工读屏和实际 bfcache 恢复仍 NOT RUN（两次 Chromium 恢复没有 persisted=true）。这些不是本地代码门的伪造 PASS，也不因10/10而消失；发布说明必须继续保留。实际生产负载下24小时清理目标仍需部署监控，局部性能数字不是 SLA。

没有新的已确认 project-spec 缺陷；本轮只更新02报告，没有子委派、生产代码/主报告修改、提交或推送。
