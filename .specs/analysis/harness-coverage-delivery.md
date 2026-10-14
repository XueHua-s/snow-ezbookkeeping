# Harness 方案到落地独立审查

审查日期：2026-10-14。输入：原方案 v1.1（重点 §11–17）、主计划及全部 14 子任务、plan-task SKILL 全部规范；抽查当前源码/测试入口。只审核并写本报告，未实施代码、未移动 todo、未运行功能测试。本报告不继承以前的 judge 分数。

## 结论与优先级

文档的主链路和交付门禁已相当完整，但还不能以“没有遗漏”结束：**D1 为 High 的源要求落地遗漏，D2 为 Medium 的 skill 分档不符合，D3 为 Low 的精确路径不足。** D1 应在确认计划完整前修复；D2/D3 可同次补齐。主计划引述源方案为规范契约，避免了语义丢失，但只靠总引述不足以证明每一项都有实施 owner 与验收。

### D1 — High：运行可观测性没有形成实施—断言—证据闭环

证据：原方案 §12.3/第 439 行要求数据库工具耗时、读取行数、返回字节数、模型轮次、工具调用次数、首个状态时间、总耗时、失败/取消原因。主计划约第 364 行仅承诺 datastore request ID/耗时/计数脱敏元数据；Step 08/14 记录 benchmark 内存、扫描行数、耗时；Step 06 的调用预算计数是控制行为，未要求对外形成可审核运行记录。全文没有覆盖完整八项的实施条目、采集口径和测试。

建议最小修订（无需新增通用遥测框架）：

- Step 01：冻结纯内部、与 HTTP/provider 私有状态分开的轻量观测字段/回调或现有日志接口约定；只允许 request ID、工具名、计数、字节、时长、受控原因枚举等。不要在业务 queries 中返回内部观测详情。
- Step 02/04：datastore 记录数据库执行时长和实际读取行数（明确是应用实际读取数，不冒称数据库优化器扫描行数）；04 统一工具 envelope 生成处记录最终序列化返回字节及工具耗时。页大小变化/空集/错误须使用真实值，避免每个工具重复实现。
- Step 06：唯一 runner owner 记录实际 provider 轮次、实际工具执行次数（缓存命中另计或说明不算执行）、整体耗时、预算/取消/超时/provider/协议/数据库失败的受控原因；所有退出路径完成且仅完成一次记录。不得直接记录 error.Error()（可能含 SQL/参数/正文）。
- Step 10：单 SSE 发送点记录首次成功状态写入相对请求开始时间；JSON 无中间状态时记 null/不适用，不能伪造 0；心跳不能作为首状态。将观测传回请求级汇总，避免第二个请求状态 owner。
- CK-18 或新增 CK：列完整八项；HR-2 保持日志禁止财务正文/备注/密钥/游标/provider 私有状态。
- 自动测试：固定时钟/记录捕获器覆盖成功、空集、多轮/切备、缓存复用、预算耗尽、数据库超时、客户端取消、SSE 写失败；断言次数/字节/首状态/单次终结记录，嵌入敏感 canary 检查任何观测输出都不含这些值。DB 行数与工具结果行数分别命名。
- Step 14：记录上述运行字段样例及采集口径，联系 10 万条基准和取消案例；环境未运行仍标待验证。

### D2 — Medium：若严格按 plan-task，模型分档理由需要逐步重算

Skill Selection Rules 的 mandatory precedence 指明 shared contract changes、data integrity、concurrency/new subsystem 命中 Complex/opus，不能以“已冻结”忽略其他更高匹配。当前 Note 大量复用“常规实现、契约冻结”，不足以解释如下步骤：

| Step | 当前 | 命中依据 | 建议 |
| --- | --- | --- | --- |
| 01 | sol/sonnet | 新增并冻结跨 llm/provider/models/frontend 的共享契约 | astra/opus |
| 02 | astra | context、读事务与连接清理并发 | 保持 |
| 03 | sol | 新配置与非敏感 settings accessor；若只依冻结映射，不含新共享协议设计 | 可保持，明确局部已定规则 |
| 04 | astra | UID/归属安全、财务一致规则/游标 | 保持 |
| 05 | sol | 跨 provider/container/httpclient 新工具协议和私有状态新子系统 | astra/opus；写明协议状态设计触发 |
| 06 | sol | 新执行循环、事件/取消并发、预算状态机 | astra/opus |
| 07 | sol | 财务逻辑交易归一/去重/关联完整性 | astra/opus |
| 08 | astra | 财务精确聚合与非平凡算法 | 保持 |
| 09 | sol | 余额、父子去重、分币种财务数据完整性 | astra/opus |
| 10 | astra | API 认证、取消/SSE 并发链路 | 保持 |
| 11 | sol | 新 SSE 增量 parser、UTF-8/终态状态机与取消竞争 | astra/opus；至少须明确是否仅按已冻结算法机械实现，当前未给算法 |
| 12 | sol | 基于共享 composable 的 UI 展示与交互 | 可保持 |
| 13 | sol | 删除已无引用运行分支、兼容配置及文档；保留 schema/数据，无破坏性迁移 | 可保持 |
| 14 | sol | 条件性三库索引/schema 兼容变更及数据完整性验收 | astra/opus（该实现职责已在本步，不是纯证据整理） |

这不是对某模型实际能力的评价，是对计划自称 sonnet/opus 对应及 skill 强制选择规则的一致性审查。全部 reviewer 已是 astra，可保持；表格、子任务 Model/Note 与主计划说明须同步。

### D3 — Low：Step 05 的真实 httpclient 修改位置仍未精确列出

Step 05 实施边界写“实际 httpclient context 包装”。当前真实文件是 `pkg/httpclient/http_request_context.go` 和 `pkg/httpclient/http_client.go`；`CommonHttpLargeLanguageModelProvider.getTextualResponse` 调用 `httpclient.CustomHttpResponseLog`。建议直接列这两个文件的核查范围，按实际是否取消断链决定修改，保留其他 HTTP 功能回归。无需将不存在的新函数误称已有。

## 源方案→实施→验收覆盖

“覆盖”指文档有实施步骤及验收要求，不表示代码已验证。

| 源要求 | 实施 owner | CK/HR 与阶段 | 证据/命令状态 |
| --- | --- | --- | --- |
| §11 P0 契约、schema、真值 | 01/04 | CK-2、CK-5；P1/P2 | 契约测试与六笔 fixture 已计划；新入口待创建 |
| §11 P1 七工具/全历史/隔离 | 04/07/08/09 | CK-1–10；P3 | services/datastore 命令明确；七 Definitions→Execute 汇合测试明确 |
| §11 P2 三协议/runner | 05/06 | CK-11/12；P2 | mock/协议/race 已计划；provider 旧接口兼容明确 |
| §11 P3 API/UI | 10/11/12 | CK-9–16/19/20；P4 | go test/race、npm lint/test/build；GUI 手测必须记录 |
| §11 P4 替换/迁移 | 13 | CK-17/18、HR-1/3；P5 | 配置升级三路径、禁止删表、保留回滚均明确 |
| §11 P5 发布验收 | 14 | 全 CK/HR；P5 | 源 §12 全必要项通过才可发布；未环境验证不计通过 |
| §12.1 独立数据/权限 | 01/07/08/09/14 | CK-1–10/13/19 | 六笔、2500跨五年、十万数据；A/B/空用户齐全 |
| §12.2 模拟与真实七问题 | 05/06/10/14 | CK-11/12/18/19/20 | 真实部署测试 opt-in 已有入口；新行为须先改写 |
| §12.3 性能与检查 | 08/14 | CK-4/18、HR-2/3 | 四种基准场景主计划已有；全仓命令准确；运行观测缺 D1 |
| §13 全链路完成定义 | 10/13/14 | DoD 全部 | 非只改提示词/单工具；无旧 RAG fallback |
| §14 原始基线 | 所有步骤/14 | HR-1/3 | 明示旧成绩不能复用；保留实际工作区状态 |
| §15.1 严格参数/schema | 01/04 | CK-2；P2/P3 | 默认/空数组/null/重复键/字节/金额时间边界测试明确 |
| §15.2 财务/转账/比较/投影 | 04/07/08/09 | CK-5–9；P3 | 独立真值、同侧匹配、ordinal/补零/余额等测试已计划 |
| §15.3 游标/预算/历史 | 04/06/10/11 | CK-10/11/16；P2–P4 | 参数绑定、24h/64KiB、配对、Unicode、256KiB 等已计划 |
| §15.4 context/数据库 | 02/04/10/14 | CK-13/20；P1/P4/P5 | 三库/主库固定/只读/取消真实驱动验证；外库待环境 |
| §15.5 协议/证据终态 | 05/06/10 | CK-9/11/12/19 | 正文缓冲/错误不空集/8引用/私有状态隔离明确 |
| §15.6 HTTP/SSE/前端 | 01/10/11/12 | CK-14/15/16 | 唯一终态/错误元数据/parser/取消/i18n/安全渲染齐全 |
| §15.7 真值/环境 | 01/08/09/14 | CK-5/8/18、HR-2/3 | 不虚构新测试入口；要求创建真实命令及变量说明 |
| §16 R01–R14 | 上述对应步骤 | 对应 CK | R01/02→02；R03→11；R04–08→07/08/09；R09→04/11；R10/11→05/06；R12→03/13；R13→07；R14→14/DoD |
| §17 开始/完成交接 | Regular Checks/01/13/14 | HR-1–3、CK-18/20 | 技能/依赖/用户改动/不提交部署、报告三状态明确 |

## DAG、文件所有权、phase 与 plan-task 结构

- 14 个子任务必需字段全部存在（含 None），主计划 14 行一一对应、没有孤儿文件。相对 Markdown 引用检查未发现断链。只有一个 Acceptance Criteria，六子块顺序齐全；20 CK 都有 test case；HR 通过相关 CK cases 与 Regular Checks 覆盖；六 rubric 权重合计 1。
- 依赖表、各文件、图一致：01/02→04；01→05→06；04→07/08/09；03/06/07/08/09→10；01/03→11→12；10/12→13→14。没有环。阶段门禁明确是额外 barrier，并不伪称所有路径都是代码依赖。
- 最大并行 3；01 类型、02 datastore/core、03 settings 分离；04 先建立私有工厂骨架，07/08/09 各替换独有文件，解决共享 registry 冲突；09 的全工具测试在汇合后运行，不虚构开发期依赖。05/06 分开 provider/runner。10 后端与11/12前端分离。13/14 与早期文件有重叠但不同时执行。
- 每阶段都有可运行边界：P1 新基础+旧入口；P2 内部 mock/协议+旧入口；P3 全七工具内部完整+旧入口；P4 新 API/UI 全链路；P5 无 embedding 启动及审计。未暴露 unsupported 骨架。阶段验收区分“基础”和完整，不提前声明全部成功。
- CK-1–10 由 P3 及 P4对应续查/证据收口；CK-11/12 P2及P4；CK-13 P1/P4；CK-14–16 P4；CK-17/18 P5；CK-19 P4/P5；CK-20 P1/P2/P4/P5；HR-1–3每阶段；六 rubric 均有对应阶段且P5最终全回归。未发现映射孤儿。
- 所有子任务均包含测试 subtask、风险影响/概率/措施。风险表较模板化但主要风险本身具体，不构成阻断。
- 保留 draft 是用户明确要求，优先于 skill 自动 Promote；不是 skill 缺陷。没有必要为本次核查实施或提交代码。

## 真实代码与命令核查及不可由文档证明的事项

实际确认：`pkg/datastore/query_context.go` 的 XOrmContextAdapter 确实未传播 Deadline/Done/Err；`database.go` 存在 engineGroup/NewSession；provider 保留 GetJsonResponse/StreamTextResponse；`src/lib/services.ts` 存在 fallbackToDoneEvent；`pkg/models/transaction_tag_index.go` 确实存在；npm scripts 确为 lint=`vue-tsc --noEmit && eslint . --fix`、test=`vitest run`、build=`vite build`（经 cross-env），因此计划提醒复核 lint 自动修改准确。

已存在的真实模型测试为 `TestAIDeploymentIntegration`，使用 `EBK_AI_DEPLOYMENT_CHECK`、`EBK_AI_DEPLOYMENT_WORK_DIR`、`EBK_AI_DEPLOYMENT_CONFIG`、`EBK_AI_DEPLOYMENT_UID`。代码分析文档给出精确运行命令，并说明必须先替换旧 embedding 断言。三库新测试/benchmark 本次尚不存在，要求 Step 14 创建入口后写实际命令合理；不能把未来测试现在列为可运行/通过。

以下只能在实施时证明：三驱动取消和RR/SQLite连接状态恢复、SQL字面匹配与精度、主从一致性、10万条资源行为和索引收益、provider真实兼容、GUI真实生命周期与Markdown安全、旧版回滚、最终回答的财务正确性，以及本地全测试通过。文档当前正确保留这些证据门禁。本审查没有执行这些测试，因此不提供运行通过结论。

## 独立评分

按 Judge 4 八项权重重新评估：Step Quality 4.1、Testability 4.0、Risk 4.0、Completeness 3.3（D1）、Dependency 4.5、Parallel 4.4、Model Selection 2.8（D2）、Phase Design 4.4；加权 **3.99/5**。数值达默认阈值不等于“无遗漏”；D1 是源要求未形成责任闭环，必须修复后才能满足本轮用户要求。未给5分或沿用以前PASS。

## 修订后独立复核（2026-10-14）

本节是对统一修订后的主计划、14 子任务和总追踪索引重新读取核查，覆盖本报告 D1–D3 以及另一个审计人提出的 R-G1–R-G6；不沿用上文分数或 writer 自评。结论：**上述缺口均已在计划层闭环，本次复核范围未发现剩余实质缺口。** 这仍不代表实现、运行测试或发布验收已经完成。

| 缺口 | 独立核对的实施 owner | 验收/测试位置 | 复核状态 |
| --- | --- | --- | --- |
| D1 / R-G6 | 01内部契约；02独立只读DB观测；04工具统一外壳；06唯一runner汇总；10首status观察回传；14归档 | CK-18：完整字段、读取/结果/字节区分、实际调用与缓存区分、首次成功status（JSON null/心跳排除）、恰好一次终结、受控原因与敏感canary；各owner Subtasks均明确 | 已闭环 |
| D2 | 主分配表及各子任务Model/Note | 01/05/06/07/09/11/14升astra并具触发理由；11astra+3sol；5阶段reviewer均astra；独立脚本比对无不一致 | 已闭环 |
| D3 | 05及主Expected Changes | 精确列两个现存pkg/httpclient文件、CustomHttpResponseLog、按需修改及其他HTTP回归 | 已闭环 |
| R-G1 | 10 | CK-16/场景4：默认四项摘要；固定3月15日独立fixture，月后半/去年同期后记录排除，分类60/40、余额1700；指定年/全历史与月底夹取 | 已闭环 |
| R-G2 | 10消息构造+05wire协议 | CK-12/16：替换buildAIAssistantUserPrompt拼接；后端now/时区/语言；历史独立角色；Chat/Responses/备用不提升system:文本 | 已闭环 |
| R-G3 | 05 | CK-12：交错choices、同call index异参、固定单候选、未选错误/截断不污染或执行 | 已闭环 |
| R-G4 | 04 | CK-10：cursor-only、同时条件规范相等/不同、固定HMAC-SHA256签名向量/用途隔离、规范JSON；审查恒时原语而非不稳定时差测试 | 已闭环 |
| R-G5 | 05/06 | CK-11：provider/run timeout两种大小关系；明确context-window超限为预算未完成，已有证据与工具对保留，不删消息重试/切备 | 已闭环 |

结构和依赖复核：14子任务必需字段和四个小节均恰好出现一次；主计划 Acceptance Criteria/Architecture Overview/Implementation Process 均唯一；5阶段保持；主计划—子任务—总索引全部相对链接存在。模型表与文件Model逐一相同。原DAG与并行文件边界未改变；02观测明确不依赖llm/services，04汇总底层观察值，未因01内部观测契约引入02→01的隐藏代码依赖。CK/HR/rubric既有阶段映射保持，新增运行观测在CK-18最终收口，阶段基础不被冒称完成。

主计划CK-18与06/10个别新增条目有概括与细化的轻微重复，含义一致、无重复章节或冲突，不构成继续阻塞计划的理由。字段口径和责任唯一性已明确。

本次独立执行的文档脚本结果：errors=[]；models={gpt-6-astra:11,gpt-6.1-sol:3}；subtasks=14；phase_count=5。`git diff --stat -- pkg src cmd conf templates docs .specs/tasks/todo` 无输出；当前git状态未显示这些受版本控制路径有修改，任务仍在draft。本复核没有事前全工作区哈希，因此不把空tracked diff单独当作所有未跟踪内容从未变动的证明；完整前后边界以编排者保存的本轮基线核对为准。本复核自己只追加审计报告，没有修改任务、子任务、源码或todo。

可向用户说明“按来源条款完成追踪、发现项已补齐并经独立复核，未发现剩余计划级实质遗漏”；不要说明“代码全部正确/所有环境通过/已可发布”。
