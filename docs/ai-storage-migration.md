# ezBookkeeping 全量迁移与数据安全手册（生产版）

> 这是一份“整项目迁移”文档，不仅限 AI。  
> 目标：在不丢失任何业务数据的前提下完成迁移。  
> 原则：先演练、后切换；先备份、后操作；可回滚。

## 1. 必须先明确的结论

- 必迁数据是三类：`数据库` + `对象存储文件` + `运行配置（含密钥）`。
- 仅导出 CSV/TSV 不是完整备份，不能替代数据库备份。
- 交易图片附件是独立对象存储数据，必须和数据库一起迁移。
- 旧 AI 助手向量表 `a_i_assistant_embedding` 保留 schema 和记录用于回滚，随整库备份迁移；新版助手不再访问或重建缓存。
- `recognize_receipt_image.json` 识图上传默认是临时识别输入，不会自动保存为交易图片附件。

## 2. 数据资产地图（整项目）

## 2.1 数据库（核心）

迁移时应按“整库”备份与恢复。业务相关表包括但不限于：

- `user`
- `two_factor`
- `two_factor_recovery_code`
- `token_record`
- `account`
- `transaction`
- `transaction_category`
- `transaction_tag_group`
- `transaction_tag`
- `transaction_tag_index`
- `transaction_template`
- `transaction_picture_info`
- `assistant_conversation`
- `assistant_message`
- `assistant_memory`
- `assistant_user_state`
- `assistant_command_receipt`
- `a_i_assistant_embedding`
- `user_custom_exchange_rate`
- `user_application_cloud_setting`
- `user_external_auth`
- `insights_explorer`

说明：

- 表结构由 `./ezbookkeeping database update` / 启动时 `auto_update_database=true` 自动维护。
- `a_i_assistant_embedding` 是保留的旧向量缓存表，存储 JSON 向量和内容哈希。数据库更新继续注册该模型，不删除旧表或记录；助手运行不再读取、写入或清理它。

物理表名以当前数据库引擎的 mapper、前缀和 schema 为准：`engine.TableName(&models.AIAssistantEmbedding{})`。默认 mapper 产生 `a_i_assistant_embedding`，不是模型文件名 `ai_assistant_embedding.go` 去掉扩展名；抽检或备份不要假设配置了前缀的部署仍使用默认名字。

## 2.2 对象存储（文件）

由 `[storage]` 决定后端类型（`local_filesystem` / `minio` / `webdav`）。

- 用户头像前缀：`avatar/`
- 交易图片前缀：`transaction/`
- 交易图片路径格式：`transaction/{uid}/{pictureId}.{ext}`

说明：

- 交易图片的“可见性”以数据库 `transaction_picture_info` 为准，文件迁移建议做前缀全量迁移。
- 即使某些图片在业务上已逻辑删除，也建议保留完整对象存储快照，避免误删引发历史数据不可恢复。

## 2.3 配置与密钥（必须同步）

至少需要完整迁移 `conf/ezbookkeeping.ini`（或你实际使用的 conf 文件），重点字段：

- `[database]`：数据库连接/路径
- `[storage]`：对象存储配置
- `[security] secret_key`：签名密钥（不一致会导致会话/令牌行为变化）
- `[uuid] server_id`：多实例场景必须保证唯一性
- `[auth]`：OAuth2/OIDC 相关配置
- `[ai]`、`[llm_image_recognition]`、`[llm_image_recognition_fallback]`、`[llm_assistant]`、`[llm_assistant_fallback]`：AI 功能相关配置

AI 配置建议重点核对（防止“能启动但行为偏差”）：

- `llm_image_recognition.llm_provider`
- `llm_image_recognition.openai_api_key`
- `llm_image_recognition.openai_base_url`
- `llm_image_recognition.openai_model_id`
- `llm_image_recognition.chat_completions_stream`
- `llm_image_recognition_fallback.llm_provider`
- `llm_image_recognition_fallback.openai_api_key`
- `llm_image_recognition_fallback.openai_base_url`
- `llm_image_recognition_fallback.openai_model_id`
- `llm_image_recognition_fallback.chat_completions_stream`
- `llm_assistant.llm_provider`
- `llm_assistant.openai_api_key`
- `llm_assistant.openai_base_url`
- `llm_assistant.openai_model_id`
- `llm_assistant.run_timeout`（默认 180000ms）
- `llm_assistant.query_timeout`（默认 30000ms）
- `llm_assistant.max_rounds`（默认 12）
- `llm_assistant.chat_completions_stream`
- `llm_assistant_fallback.llm_provider`
- `llm_assistant_fallback.openai_api_key`
- `llm_assistant_fallback.openai_base_url`
- `llm_assistant_fallback.openai_model_id`
- `llm_assistant_fallback.chat_completions_stream`

`llm_assistant_fallback` 为只读工具助手提供请求内最多一次备用切换。新版忽略整个 `[llm_assistant_embedding]`、`openai_embedding_model_id` 及对应的 `EBK_` / `EBKCFP_` 覆盖，不读取旧密钥文件；旧段无效或文件缺失不会阻止启动。若上游只接受流式 Chat Completions，请设置对应 section 的 `chat_completions_stream = true`；Responses 使用 `llm_provider = openai_responses_compatible` 及 `openai_compatible_*` 配置；浏览器 JSON/SSE 两条路由共享所配置的上游协议。主备模型均须支持对应工具协议。预算只配置在主助手段，不能限制历史范围。浏览器超时为整体预算加 10 秒，反向代理读取超时至少覆盖该值，关闭 SSE 缓冲。

同时检查环境变量覆盖：

- `EBK_WORK_DIR`（影响相对路径解析）
- `EBK_<SECTION>_<ITEM>`（直接覆盖配置项）
- `EBKCFP_<SECTION>_<ITEM>`（从文件读取覆盖值）

如果迁移后环境变量缺失或不同，会导致“同一份 ini 但行为不同”。

## 3. 迁移前准备（强制）

1. 先做一次“演练迁移”（同版本、同流程、非生产数据）。  
2. 约定维护窗口，迁移期间停止写入。  
3. 记录源端版本号、提交号、配置来源（文件+环境变量）。  
4. 目标端预先准备：
   - 应用二进制/镜像版本
   - 数据库实例
   - 对象存储实例
   - conf 文件和密钥
5. 确保有可回滚备份，且备份可读可恢复（不是只“备了个文件”）。

## 4. 标准迁移流程（推荐）

## 4.1 冻结写入

- 停止应用实例（或摘流量并确认无写请求）。
- 禁止旧实例与新实例同时写同一套数据。

## 4.2 备份配置

- 备份 conf 文件：
  - `conf/ezbookkeeping.ini` 或自定义 `--conf-path` 文件
- 备份部署环境变量（尤其 `EBK_*` / `EBKCFP_*` / `EBK_WORK_DIR`）。

## 4.3 备份数据库（重点）

### SQLite

```bash
sqlite3 /path/to/ezbookkeeping.db "PRAGMA wal_checkpoint(FULL);"
sqlite3 /path/to/ezbookkeeping.db ".backup '/backup/ezbookkeeping.db.bak'"
sqlite3 /backup/ezbookkeeping.db.bak "PRAGMA integrity_check;"
```

预期：`integrity_check` 返回 `ok`。

### MySQL

```bash
mysqldump \
  --single-transaction \
  --quick \
  --routines \
  --triggers \
  --events \
  --hex-blob \
  --set-gtid-purged=OFF \
  -h <host> -P <port> -u <user> -p <db_name> \
  > /backup/ezbookkeeping.sql
```

### PostgreSQL

```bash
pg_dump \
  -h <host> -p <port> -U <user> -d <db_name> \
  -Fc \
  -f /backup/ezbookkeeping.dump
```

## 4.4 备份对象存储

### `local_filesystem`

- 完整备份 storage 根目录（默认 `storage/`）：
  - `storage/transaction/`
  - `storage/avatar/`

示例：

```bash
rsync -aHAX --numeric-ids --info=progress2 /src/storage/ /backup/storage/
```

### `minio` / `webdav`

- 按 bucket/root path 做全量同步（至少包含 `transaction/`、`avatar/` 前缀）。
- 建议保留迁移前快照或版本化备份。

## 4.5 传输到目标端并恢复

1. 恢复数据库备份。  
2. 恢复对象存储数据。  
3. 恢复 conf 与环境变量。  
4. 核对目标端路径：
   - 相对路径以 `EBK_WORK_DIR` 或进程工作目录为基准。

## 4.6 数据库结构对齐（必须执行）

在目标端执行：

```bash
./ezbookkeeping --conf-path=/path/to/ezbookkeeping.ini database update
```

说明：

- 该命令会同步表结构（含 `a_i_assistant_embedding`）。
- 即使 `auto_update_database=true`，生产迁移仍建议先手动执行一次，再启动服务。

## 4.7 启动服务并验证

```bash
./ezbookkeeping --conf-path=/path/to/ezbookkeeping.ini server run
```

## 5. 迁移后校验清单（必须逐条勾）

## 5.1 功能校验

- 能正常登录（历史用户可登录）
- 交易列表、分类、账户、标签、模板数据正常
- 历史交易图片可打开
- 新上传交易图片可显示
- AI 助手可正常响应（启用时）

## 5.2 统计校验（建议）

使用数据统计接口/页面比对迁移前后：

- 账户数
- 交易数
- 分类数
- 标签数
- 交易图片数
- 模板数
- Insights Explorer 数

接口：`GET /api/v1/data/statistics.json`

## 5.3 数据库抽检（建议）

抽检核心表记录数（源端 vs 目标端）：

- `transaction`
- `account`
- `transaction_picture_info`
- `a_i_assistant_embedding`（保留的旧缓存）

说明：

- `a_i_assistant_embedding` 随整库迁移，核对记录数及抽样内容哈希/向量/时间字段不变；新版不会按需重建，不能以重建策略解释数据缺失。

## 5.4 对象存储抽检（建议）

- 抽样核对多个用户的图片是否可访问。
- 对于本地文件存储，可比较目录结构和文件总量。
- 文件总数允许大于“有效图片记录数”（历史逻辑删除可能仍保留文件）。

## 6. 回滚方案（必须提前准备）

触发条件示例：

- 登录异常或大面积 5xx
- 核心表计数严重不一致
- 图片大面积 404

回滚步骤：

1. 停止目标服务。  
2. 恢复目标端到迁移前快照（数据库 + 对象存储 + 配置）。  
3. 恢复源端服务。  
4. 标记故障窗口并保留日志与备份用于复盘。

## 7. 高风险点与注意事项

- 不要把 CSV/TSV 导出当作全量备份。  
  CSV/TSV 主要是交易数据导出，不覆盖全部系统状态与文件对象。

- 识图接口不是图片持久化接口。  
  `recognize_receipt_image.json` 只识别图片并返回建议字段；交易附件要走 `transaction/pictures/upload.json`。

- 配置不一致比数据错误更隐蔽。  
  同一套数据库，若 `secret_key`、对象存储配置或环境变量覆盖不同，会表现为“功能正常但行为异常”。

- 多实例并行写入需谨慎。  
  若新旧实例同时写入，必须明确数据库与对象存储一致性策略，并确保 `uuid.server_id` 不冲突。

- 大版本跨越建议分段升级。  
  先在演练环境验证每一步，再执行生产迁移。

## 8. Docker 场景补充

官方镜像工作目录通常是 `/ezbookkeeping`，常见持久化目录：

- `/ezbookkeeping/data`
- `/ezbookkeeping/storage`
- `/ezbookkeeping/conf`
- `/ezbookkeeping/log`（可选）

迁移时请确保这些目录（或等价挂载卷）一致迁移。

## 9. 只读工具助手升级与旧缓存回滚

升级前保存当前二进制或镜像版本、完整数据库（含 `a_i_assistant_embedding`）、对象存储、原 ini、环境变量和旧密钥的安全备份，并在隔离副本演练恢复。

1. 在测试副本部署新版，保留旧表和记录。配置 `[ai]` 开关、主备工具模型及正整数运行预算，识图配置继续保留。
2. 可移除实际运行环境中的旧 embedding 段/环境变量/密钥挂载；新版也允许原样留下无效旧段和不存在的旧密钥路径。回滚备份须保存原配置和旧密钥。
3. 执行 `./ezbookkeeping database update`，确认旧表仍存在，记录数及抽样字段与备份一致。助手不会迁移、读取、写入或清空缓存；无 embedding 供应商也能工作。
4. 用专用测试用户执行七工具查询、五年以上历史、分币种统计、当前余额、识图和双端取消/续查验证。查询预算耗尽必须显示未完成，不能缩短范围后报告全量。补测命令见 `docs/ai-providers.md` 和验收记录。
5. 本次升级不执行任何 `DROP/TRUNCATE/DELETE a_i_assistant_embedding`，也不自动清理“过期缓存”。删除旧表需要后续独立数据迁移授权。

若回滚，在维护窗口停止新版和业务写入，切换已备份的旧二进制/镜像、原 ini、环境变量及 embedding 密钥挂载，保留旧表。确认识图和旧助手能读取测试账本与缓存后再恢复流量。若需要恢复整个快照，先备份升级后新增的业务数据并按第 6 节恢复整套数据库/对象存储/配置，避免覆盖升级后记账。不要同时运行新旧版本写入业务库。

新版无长期运行状态表；页面历史与 continuation 不可跨版本恢复。回滚后清空页面会话并重新查询。新版不会产生新 embedding 缓存，旧版可根据其原有行为补齐新增账单的缓存；这不影响业务账单记录。

## 10. 最小可执行清单（TL;DR）

1. 停服务，冻结写入。  
2. 备份 conf + 环境变量。  
3. 备份整库。  
4. 备份对象存储（`transaction/`、`avatar/`）。  
5. 恢复到目标端。  
6. 执行 `database update`。  
7. 启动服务。  
8. 做统计与抽样校验。  
9. 通过后切流量，失败就回滚。

合并新版上游后，AI 全局开关位于 `[ai]`；识别功能和个人财务助手仍兼容旧 `[llm]` 下的开关及图片大小限制，但 `[ai]` 中的显式设置优先。助手配置继续使用 `[llm_assistant]`；浏览器 JSON/SSE 共用执行器，按 provider 配置使用 Chat Completions、强制流式 Chat Completions 或 Responses 工具协议。备用模型保持有效；独立 embedding 配置已忽略，旧表仍保留。

官方 OpenAI 的文本识别和编程助手沿用上游 Responses API；票据识别及普通财务助手对话保留本仓库的 Chat Completions 兼容与备用模型机制。


## 持久化助手升级、保留与恢复

升级采用新增表/索引，五张助手表位于用户数据分库，必须备份每一个分库；单独备份用户库、CSV/TSV 或会话导出均不是完整备份。本文的 SQLite `.backup`、MySQL 整库 `mysqldump` 与 PostgreSQL 整库 `pg_dump` 会包含新表，无需另建一份不一致的助手导出。表名受 mapper、前缀、schema 影响，以运行引擎解析为准。

- `assistant_conversation`：标题、摘要、冻结来源、上下文版本及占用。
- `assistant_message`：完整提问、全部回答尝试、依据、任务状态、租约与结算。
- `assistant_memory`：有效条目、待确认候选、来源与遗忘排除。
- `assistant_user_state`：额度计数、删除与记忆 epoch、账号不可用 fence。
- `assistant_command_receipt`：无正文幂等回执；不能为节省备份空间单独丢弃。

顺序为维护备份 → 可重复新增表/索引 → 同时支持旧无状态接口与新会话接口的后端 → 新前端 → 验证能力与任务。旧二进制应忽略新增表；回滚先切旧前端并停止新任务，再回旧后端，保留全部新表和旧 embedding 表，不运行删表/重建脚本。不要在同一库上并行运行不兼容的新旧写入者。

恢复前停止所有应用实例，恢复同一时间点的全部分库与配置。恢复后启动新版本，过期的已开始任务通过租约扫描转为 interrupted，不重放上游请求；旧 owner 的提交被 fencing 拒绝。过期查询游标不能续查，用户需明确重新查询。先在隔离副本验证五表记录数、会话/版本、回执和 epoch、额度结算及账本/embedding 不变，再恢复流量。仅复制 SQL 文件而未实际恢复不算完成演练。

会话删除先使在线读取和写回立即不可访问，再由独立后台清理物理正文、引用、尝试和派生缓存；目标在 24 小时内完成。候选记忆最长保留 7 天，无正文回执在对象删除后至少保留 48 小时，清空仍保留用户状态与 epoch 防止迟到命令复活。后台清理按正文 tombstone、过期候选、过期回执分别使用 UID keyset 有界扫描全部用户数据分库，跳过无待清记录的用户；日志只记录批次大小、是否满批及超期受控 ID，不调用模型，也不访问旧 embedding 表；应监控清理滞后，部署规模不能让扫描周期超过删除目标。

清空全部用户数据包括助手数据；只清账单保留聊天可读历史，但撤销旧账本上下文并停止旧生成，下次回答重新查询。账号禁用保留历史但封住新写与旧 worker；账号删除同时撤销任务和清理助手正文。数据统计包含助手已用、预留字节与会话数；归档和全部回答版本仍占容量。

在线删除不是备份永久擦除。运维须公开备份保留天数、轮转位置、销毁周期和恢复权限；已经导出的用户副本不受服务端删除控制。从删除前快照恢复可能重新引入旧内容，恢复前应核对删除审计并重新执行适用删除，不向用户声称已同时抹除历史备份。

全局“清空全部数据”的新客户端附带 `assistantCommand`（仅命令元数据在浏览器保留，密码不落盘）。旧客户端没有命令字段时，助手清理采用令牌绑定的稳定回执、签发时间和初始 epoch 0；旧令牌超过 24 小时、或已经发生其他清空后，会拒绝该旧请求，需升级客户端并主动重新确认。此兼容保护防止旧清空请求误删后来创建的数据。账本清理仍沿用原有逐服务流程，但只有首次受理者可以执行；全部步骤成功后，回执另记操作完成。重复请求若已完成直接返回，未完成则返回 409 并说明“助手已清空，其他数据结果待检查”，不再次删除。用户检查现状后，只有明确发起新操作、新命令 ID 才会再次清理；不要把助手步骤受理当成全账本跨服务事务成功证明。
