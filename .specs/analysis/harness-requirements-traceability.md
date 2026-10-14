# AI Assistant Harness 需求追踪与缺口修订总索引

当前状态（2026-10-15）：原14步本地实施、五阶段独立审查及全部本地门禁已完成，最终本地结论PASS（9/10），无未解决的实质性实施缺陷；新增Docker整应用构建/HTTP/双页面smoke/E2E也已通过独立证据复核。真实主备模型七题待环境，**不可发布**。任务保持in-progress；用户最新授权commit/push及向当前origin `XueHua-s/ezbookkeeping` 的 `main` 创建草稿PR，不授权生产部署/删除旧缓存。实际实现覆盖、已关闭问题、门禁和限制见[最终实施审查](harness-final-implementation-review.md)与[验收记录](../../docs/ai-assistant-harness-verification.md)。

以下内容是2026-10-14的**计划阶段历史审查**：当时只证明计划覆盖与可执行交接，不表示当时已经实现或通过测试。其“draft/未运行业务测试”等历史描述不代表上述当前状态。原方案及用户原始文件保持；14步/5阶段结构未改变。

## 输入和逐条来源矩阵

- [原方案 v1.1](../../AI_ASSISTANT_HARNESS_PLAN.md)：来源规范，不改写源设计。
- [主计划](../tasks/in-progress/ai-assistant-harness.feature.md)：CK-1—20、HR-1—3、六rubric、阶段门禁。
- [财务逐条矩阵](harness-coverage-finance.md)：§1—6、§15.1—15.4、§15.7，涵盖完整数据、schema、金额、交易、余额、筛选、权限、游标、事务与fixture。
- [运行链逐条矩阵](harness-coverage-runtime.md)：§7—10、§15.3、§15.5—15.6、§17，涵盖provider/harness/API/UI/config/迁移。
- [交付及plan-task审计](harness-coverage-delivery.md)：§11—17与skill结构、DAG、所有权、模型分档和验证交付。

三份审计是修订前发现的独立记录，不覆盖原文或被静默改成PASS；其gap项的最新状态以本索引的修订映射及报告末尾独立复核为准。交叉覆盖不矛盾：准确章节绑定可以证明要求已承接，但缺具名测试/owner仍在本轮补齐。

| 原方案范围 | 详细矩阵 | 主验收 | 实施阶段 |
| --- | --- | --- | --- |
| §1—3 目标/现状/边界 | finance §1—3各行 | CK-1/3/4/17/20，HR | 1—5 |
| §4 工具/schema/筛选/分页/结果/比较 | finance §4.1—4.5各行 | CK-1—10 | 1—3；证据4 |
| §5 UID/只读/隐私 | finance §5各行 + runtime R-G6 | CK-3/10/18，HR-2 | 1—5 |
| §6 完整性/财务口径 | finance §6各行 | CK-4—9/19 | 3—5 |
| §7 Harness/预算/provider/证据 | runtime §7各行 | CK-9/11—13/19/20 | 2/4/5 |
| §8 请求/摘要/短期上下文/SSE | runtime §8各行 | CK-14—16/19 | 4/5 |
| §9 文件复用 | runtime §9 + delivery路径核查 | CK-20 | 1—5 |
| §10 配置/旧表/升级回滚 | runtime §10各行 | CK-17/18，HR-1/3 | 1/5 |
| §11—14 阶段/门禁/交付/基线 | delivery覆盖表 | 全CK/HR | 1—5 |
| §15.1—15.2 冻结参数/财务 | finance对应各行 | CK-1—9 | 1—3 |
| §15.3 游标/预算/历史 | finance + runtime各行 | CK-10/11/16 | 2—4 |
| §15.4 context/事务 | finance对应各行 | CK-13/20 | 1/3/4/5 |
| §15.5—15.6 provider/API/UI | runtime对应各行 | CK-9/12/14—16 | 2/4 |
| §15.7 fixtures/环境 | finance + delivery对应行 | CK-5/8/18/19，HR-2/3 | 1/3/5 |
| §16 R01—R14 | delivery对应映射 | 对应全部CK | 1—5 |
| §17 执行与交付边界 | runtime + delivery对应行 | HR-1—3/CK-18/20 | 全程 |

## 缺口修订闭环（已独立复核）

| 缺口 | 原条款 | 精确计划修改位置/owner | 测试断言与验收 | 修订状态 |
| --- | --- | --- | --- | --- |
| R-G1 摘要缺分类/余额 | §8 | [10 API](../sub-tasks/ai-assistant-harness/10-api-sse-integration.md) Subtasks默认summary/固定3月15日fixture；主计划CK-16/场景4/执行器架构 | 四项证据；100/40差60与150%；同比25差75与300%；分类60/40；余额1700；月后半不污染同期；全年/全历史覆盖默认/月末夹取 | 已独立复核闭环 |
| R-G2 时钟/角色 | §7.1 | 10系统消息构造；[05 provider](../sub-tasks/ai-assistant-harness/05-provider-tool-turns.md)独立角色编码；主CK-12/16 | 固定后端now/时区/语言；历史system:仍普通内容；Chat/Responses/备用重建不提升角色 | 已独立复核闭环 |
| R-G3 多候选 | §15.5 | 05 choice隔离；主CK-12/架构 | 交错choice/同call index不同参，固定单候选、未选候选不执行 | 已独立复核闭环 |
| R-G4 游标细则 | §15.3 | [04 query foundation](../sub-tasks/ai-assistant-harness/04-query-foundation.md)；主CK-10/游标架构 | cursor-only合法；规范相等接受/变更拒绝；规范JSON/HMAC-SHA256固定向量；恒时原语审查 | 已独立复核闭环 |
| R-G5 provider限制 | §7.2/15.3 | 05超时/错误归一；[06 runner](../sub-tasks/ai-assistant-harness/06-assistant-runner.md)预算终态；主CK-11 | provider/run最早deadline两种大小；context-window错误保留证据工具对、不删消息重试/切备 | 已独立复核闭环 |
| R-G6与D1 观测 | §5/12.3 | [01契约](../sub-tasks/ai-assistant-harness/01-neutral-contracts.md)内部约定；[02 datastore](../sub-tasks/ai-assistant-harness/02-context-read-view.md)读取数/时长；04工具字节/行数；06唯一汇总；10首status；[14验收](../sub-tasks/ai-assistant-harness/14-verification-evidence.md)归档；主CK-18/架构 | 全八类指标；DB读取与结果行数区分、JSON首status null/心跳不计；成功/空/切备/缓存/预算/DB超时/取消/写失败恰好终结一次；敏感canary无泄漏 | 已独立复核闭环 |
| F1 回答口径 | §6.2 | 10提示词+mock；14真实问答/页面对照；主CK-19 | 净额890≠净资产1700；请求时区与记录时区跨月差异说明 | 已独立复核闭环 |
| F2 聚合排序 | §4.5/15.1 | [08聚合](../sub-tasks/ai-assistant-harness/08-aggregate-comparison.md)；主CK-9 | 已选metrics排序/拒非法；asc/desc并列完整键、多币种各自排名、深页无重漏、totals不变、重算queried_at更新 | 已独立复核闭环 |
| F3 计数不可加 | §15.2 | 08；主CK-9 | counts_additive=false显式字段；同转账双侧多组count和大于logical count仍逻辑去重 | 已独立复核闭环 |
| F4 删除/组合 | §1/5/4.2/15.1 | 04校验；[07明细](../sub-tasks/ai-assistant-harness/07-transactions-detail.md)/08行为；[09元数据余额](../sub-tasks/ai-assistant-harness/09-metadata-balances.md)；主CK-2/3 | deleted排除/hidden纳入；缺必需关联报错；tag any/all/no_tags/孤立tag_match；period空重名缺边界/与filter冲突/comparison引用非法；非法不执行 | 已独立复核闭环 |
| D2 模型分档 | plan-task Selection Rules | 01/05/06/07/09/11/14 Model+Note；主计划分配表/统计 | astra 11、sol 3；共享契约/财务完整性/并发/新子系统各具理由；reviewer保持astra；不改依赖阶段 | 已独立复核闭环 |
| D3 httpclient路径 | §9/15.4 | 05实施边界；主Expected Changes | 精确列pkg/httpclient/http_request_context.go与http_client.go、CustomHttpResponseLog；按需改并保留旧HTTP回归 | 已独立复核闭环 |

## 本轮文档结构检查

所有修订沿用CK-1—20/HR-1—3，未新增孤立验收编号；对应既有阶段仍覆盖：CK-2/3在Phase2基础及3完整，CK-9在3/4，CK-10在2基础及4续查，CK-11/12在2/4，CK-16/19在4，CK-18最终5。观测基础随01/02/04/06实施，Phase5以CK-18完整收口。六rubric不变，分别由财务完整性、隔离、执行可靠性、双端体验与可追溯性承接。

本轮仅运行文档字段、链接、DAG/模型表一致性检查；未运行业务测试。最终实施必须执行原计划真实自动/数据库/模型/GUI/性能/回滚门禁，环境未具备保持待验证。

## 独立复核结论

[财务复核](harness-coverage-finance.md)末尾确认F1—F4全部closed；[交付复核](harness-coverage-delivery.md)末尾确认D1—D3及R-G1—R-G6均闭环；[运行链复核](harness-coverage-runtime.md)末尾记录对应证据。R-G6与D1合并后共12项发现，均有实施owner、具名测试断言与验收映射，未发现剩余计划级实质遗漏。初审历史保留。该结论仅证明计划覆盖与可执行交接要求，不代表代码实现、业务测试、外部环境验收通过或已可发布。
