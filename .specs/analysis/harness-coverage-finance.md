# Harness 来源覆盖审计：财务、分页、事务

审计范围：完整阅读源方案、主计划及 14 子任务；逐项核对源 §1—6、§15.1—15.4、§15.7。状态只表示**计划覆盖**，不是实现通过。步骤号对应 `.specs/sub-tasks/ai-assistant-harness/`；验证列的 CK 指主计划 Test Cases to Cover，同一行列出具体断言。准确引用具体冻结节且有 owner 和对应测试的要求视为 covered；泛指整份源方案不算闭环。

初审结论：关键财务、隔离、事务安全主线已覆盖，无 High 缺口。初审发现 4 项 Medium 验收闭环缺口（F1—F4）。以下原始矩阵保留初审证据；**统一修订后四项均已独立复核闭环，当前无未解决实质缺口**，精确证据见文末复核表。此结论仅针对计划，不表示实现或测试通过。

| source 锚点 | 具体要求 | CK/HR | 步骤 owner | 验证 | 状态 |
| --- | --- | --- | --- | --- | --- |
| §1 目标 | 移除 embedding/top-k/窗口，不留 RAG 分支 | CK-4/17 | 10/13 | 无网络 embedding/缓存读写，2500 条及五年数据 | covered |
| §1 当前账户 | UID 全部账户，只有显式条件才收窄 | CK-1/3/4 | 04/07/09 | A/B 隔离、默认全账本、账户条件 | covered |
| §1 没有限制 | 任意历史、连续分页、完整统计，预算不裁范围 | CK-4/11 | 07/08/06 | 2500 条无重漏及10万条，无扫描上限 | covered |
| §1 隐藏/删除 | 隐藏默认纳入，已删除排除 | CK-1/3 | 04/07/09 | 隐藏夹具；deleted 属源固定查询契约和全范围夹具，但缺显式删除负例，见 F4 | gap Medium F4 |
| §1 权限 | 不提供写入、附件内容、模板、历史余额、密码令牌 | CK-3/19/20; HR-2 | 04/07/10 | 白名单攻击、投影敏感字段、历史余额能力提示 | covered |
| §2 现状 | 修复窗口、缓存、MCP 层级与隐藏语义、两模型链路、context/流尾问题 | CK-1/4/12/13/14/17 | 02/05/07/09/11/13 | 对应专项测试，非复制现有缺陷 | covered |
| §3 边界 | API 只认证组装；runner 无财务计算；provider 无 DB/身份决策 | CK-20 | 01/04/05/06/10 | 包依赖/SQL与解析归属审查 | covered |
| §3 服务/前端 | 查询服务固定 UID，领域服务保持 deleted；前端不持密钥 | CK-3/20; HR-2 | 04/10/11 | 分库/混合 ID/投影检查；deleted 负例见 F4 | covered（删除负例另列） |
| §3 独立性 | 单循环、七平级工具，不启外部 MCP 也可运行，MCP 独立 | CK-1/20 | 04/06/14 | 七名称真实装配测试、MCP 兼容回归 | covered |
| §4.1 list_accounts | ID/名称/parent/type/currency/hidden 全有效账户 | CK-1 | 09 | 超一页实体、同名层级与完整业务投影 | covered |
| §4.1 list_categories | 类型、层级、隐藏、稳定 ID | CK-1 | 09 | 父子/同名层级与页遍历 | covered |
| §4.1 list_tags | 标签组 ID/名称、隐藏、稳定 ID | CK-1 | 09 | 标签组及超一页遍历 | covered |
| §4.1 query/get | 明细匹配总数/范围/游标；详情关联/标签/备注/附件数量无存储密钥 | CK-4/6/9 | 07 | 完整详情、投影排除路径/IP、转入归一 | covered |
| §4.1 aggregate | 完整总计、计数、分组；通用工具不造专用问题工具/脚本 | CK-4/5/20 | 08 | 完整 fixture、白名单/架构审查 | covered |
| §4.1 balances | 当前叶子余额/资产负债/查询时间/分币种完整总计 | CK-8/9 | 09 | 2000/-300=1700，分页不改 totals | covered |
| §4.1 换汇/历史 | 原币种分开，不假造历史余额/换汇 | CK-5/19 | 08/09/10 | 多币种、能力提示测试 | covered |
| §4.2 时间/type | 同一强类型 filter，RFC3339 半开可空，无默认历史截断，四类型 | CK-2/4/7 | 04/07/08 | 全默认值/非法枚举/秒边界 | covered |
| §4.2 account/category | 父项展开有效子项、范围回写、父子去重 | CK-3/7 | 04 | 父子去重、层级变更游标测试 | covered |
| §4.2 tags | any/all，no_tags 与非空集合互斥 | CK-2 | 04 | CK-2 全部参数规则引用 §15.1/4.2；需具名组合负例，见 F4 | gap Medium F4 |
| §4.2 currency/side/amount | 同侧联合、默认 either、含边界金额 | CK-2/5/6 | 04/07/08 | source/destination/either+账户币种金额联合 | covered |
| §4.2 keyword | 参数化转义、不当 SQL/指令 | CK-2/3 | 04/14 | 三库 %/_/反斜线/大小写与注入 | covered |
| §4.2 严格 JSON | 字符串 ID，未知/重复键/非法值拒绝 | CK-2 | 01/04 | 序列化与非法输入表驱动 | covered |
| §4.2 实体 | 同名查层级再澄清，缺失/混合 ID 不放宽 | CK-3/19 | 04/10 | 同名歧义与非法集合 | covered |
| §4.3 明细排序 | 默认 time desc，逻辑 time+ID；amount 限币种和转账侧 | CK-4/6 | 07 | 静态全页无重漏、币种/侧约束 | covered |
| §4.3 分页 | 50/200，只限单页，无最大页数/年限 | CK-2/4 | 04/07/09 | 页长边界、2500 条/多页元数据 | covered |
| §4.3 游标 | UID/filter/order/边界/末键签名绑定 | CK-10 | 04/07 | 篡改、跨用户、换条件拒绝 | covered |
| §4.3 分组分页 | totals 完整、第一页不是全部组 | CK-9 | 08 | 已有“组分页不改总计”成功标准；具体稳定排名分页负例见 F2 | covered（排序测试另列） |
| §4.3 字节 | 缩页不跳行；单笔完整；异常单项 RESULT_TOO_LARGE | CK-9/11 | 07 | 大小预算、returned_count 实长、单项错误 | covered |
| §4.4 结果形状 | ok/query_id/scope/counts/cursor/coverage/time/data/warnings | CK-9 | 01/04/07/08/09 | 序列化/计数字符串/data长度、真实范围 | covered |
| §4.4 覆盖率 | page 与 totals_coverage/all_matching/groups_has_more 独立 | CK-9 | 08 | 分组页不改 totals；范围状态测试 | covered |
| §4.4 空/错误 | 空集成功0无下一页；失败不变无记录 | CK-9/11 | 04/06/10 | 空账本/工具错误 mock | covered |
| §4.4 金额/投影 | 十进制字符串币种、无浮点源、不直接 ORM 序列化 | CK-5/9; HR-2 | 01/07/08 | 格式/数值真值/敏感字段排除 | covered |
| §4.4 错误 | 指定九类代码、message/retryable/纠正信息，无 SQL/连接串/他人ID存在性 | CK-3/9; HR-2 | 01/04 | CK-9 源 §4.4 错误码全集；越权统一错误 | covered |
| §4.5 group/metrics | day/month/year/category/account 白名单，时间互斥；币种/flow 独立；count/sum/min/max | CK-2/5/7 | 04/08 | 重复维度、四指标、转账两流/普通三流 | covered |
| §4.5 periods | 替代 filter 时间、不可同时；同视图，各期实际边界 | CK-2/7/13 | 04/08 | 全入参规则、重叠 period/同工具快照 | covered |
| §4.5 comparison | 引用已定义 period，同币种组、ordinal 对齐 | CK-2/7 | 04/08 | 非法定义/ordinal/未配对桶 | covered |
| §4.5 排名 | 指标白名单、同值完整分组键稳定排序、不同币种分别排名，重算新 queried_at | CK-2/7/9 | 08 | 目前只有广义“按§4.5”及组分页标准，未列并列值跨页/币种排名测试 | gap Medium F2 |
| §4.5 补零 | 有限时间缺月补零、缺失非时间组补零 | CK-7 | 08 | 有限桶分页补零、缺失分类、不造笛卡尔积 | covered |
| §5 1—4 | 认证读 UID、请求域 context/时区/语言，schema 无 UID/SQL/表路径URL | CK-2/3/13 | 04/10 | 伪造 UID/SQL/未知字段及认证流程 | covered |
| §5 5—6 | 所有实体/游标归属，所有主/关联/子/统计 UID+deleted，用户分库 | CK-3 | 04/07/08/09 | A/B 全工具攻击；删除记录显式负例见 F4 | covered（删除负例另列） |
| §5 7—8 | 独立只读白名单；引用只来自结果且打开仍鉴权 | CK-3/9/15 | 04/06/12 | 写工具拒绝、实际逻辑ID引用及链接鉴权 | covered |
| §5 不可信内容 | 名称/备注/历史不改变身份系统工具 | CK-3/19 | 04/10 | 历史/备注注入、系统/tool角色拒绝 | covered |
| §5 缓存/日志 | 仅请求内缓存不持久化、不跨用户；脱敏 provider debug | CK-11; HR-2 | 05/06/10 | 不持久化、捕获日志无财务/秘密/游标 | covered |
| §6.1 全量 | 参数化聚合/白名单；否则精确分批直到完毕；内存批+组而非全交易 | CK-4/5 | 08/14 | 2500条/10万 benchmark，扫描计数/内存 | covered |
| §6.1 三库精度 | SUM/溢出/日期精度实测，无 float64 | CK-5/18 | 08/14 | 三库契约、极大累计值，不能先SUM溢出 | covered |
| §6.1 读视图 | 同工具多SQL短事务；不等待LLM持事务 | CK-13 | 02/04/06 | 并发修改快照/连接释放；架构明确执行后关闭再回模型 | covered |
| §6.1 多轮 | 各轮当前读取，queried_at不是会话快照；严格跨期一次查询 | CK-4/7/9 | 07/08/12 | 修改期间非全会话快照、period同视图 | covered |
| §6.1 高水位 | 实际完整筛选最大次序，包含未来；不能解决回填/编辑 | CK-4/10 | 07 | 未来日期、高水位、修改期间说明 | covered |
| §6.1 超时 | 不回部分总计，先优化不裁范围 | CK-4/11/18 | 08/14 | 超时未完成、10万性能和条件索引 | covered |
| §6.2 日期 | 后端当前时间+请求时区、半开转换、全历史/年度明确范围 | CK-4/7/16 | 04/08/10 | 固定时钟/秒端点/DST/用户期间优先 | covered |
| §6.2 页面时区口径 | 请求时区分组；与交易记录时区页面比较时注明口径差异 | CK-7/19 | 候选10/12/14 | 已有请求时区测试，但无页面口径差异提示/验收断言 | gap Medium F1 |
| §6.2 财务类型 | 转账单逻辑双金额、不入普通收支；调整单列；退款保留负数 | CK-5/6 | 07/08 | 六笔真值、负支出、两流一笔 | covered |
| §6.2 账户侧 | 筛选/金额/现金流同侧，双命中 count 一次 | CK-6 | 07/08 | 联合条件与目的USD只转入 | covered |
| §6.2 多币种/层级 | 独立总计/排序/比较，父子展开去重 | CK-5/7/8 | 04/07/08/09 | 多币种拒混、父子不重 | covered |
| §6.2 标签 | 仅过滤，不提供可加总多标签比例分组 | CK-2/7 | 04/08 | group_by白名单无tag；组合过滤负例见 F4 | covered |
| §6.2 余额/比率 | 叶子现有余额不重父；baseline<=0无增长率 | CK-7/8 | 08/09 | 2000/-300及非正基数null | covered |
| §6.2 净额 | 回答明确收入-支出不等于余额/可用资产 | CK-19 | 候选10/14 | 工具数值已分开，但最终回答尚无该语义断言 | gap Medium F1 |
| §15.1 schema | 所有层 additionalProperties=false，同领域生成；strict nullable+required，必填null拒绝 | CK-2/12 | 04/05 | 嵌套额外字段/null；strict边界适配 | covered |
| §15.1 七参数表 | 每工具冻结键及默认：ID升序、三分类类型、未来/调整、正整数详情、空group/默认metrics、叶子余额 | CK-1/2 | 04/07/08/09 | CK-2明确全默认值；各工具挂点及真实装配 | covered |
| §15.1 分页/数组 | 50/200；0负超限拒绝；显式空ID/type数组拒绝，null不筛，去重逐元素验权 | CK-2/3 | 04 | 0/负/201、空数组、混合ID | covered |
| §15.1 tag/period/group | any默认；孤立tag_match报错；period非空/唯一name/必填时间，group重复/双时间拒绝 | CK-2/7 | 04 | CK-2默认与重复维度覆盖；tag具体组合与period互斥需具名负例见F4 | gap Medium F4 |
| §15.1 sort | field/direction结构；time/amount；aggregate group_key或所选metrics；稳定唯一键 | CK-2/4 | 04/07/08 | 明细稳定页覆盖，聚合排名负例见 F2 | covered（聚合验证另列） |
| §15.1 amount/keyword | 含边界、负数合法、min>max拒绝、名称备注字面大小写不敏感 | CK-2/5 | 04 | 非法金额/三库字面检索 | covered |
| §15.1 time | 秒精度拒小数、start<end、乘1000范围、null非0、旧helper end*1000-1 | CK-2/7 | 04 | 小数秒/溢出/秒边界用例及统一转换 | covered |
| §15.1 字节 | 非cursor32KiB/总96KiB，拒绝不截列表，无动态表达式 | CK-2 | 04 | 32/96KiB边界、伪造SQL | covered |
| §15.2 金额 | 100倍、不按ISO改缩放；可负两位，无科学千分NaN；先范围后解析 | CK-2/5 | 04/08 | 长度/溢出/科学NaN拒绝、精确真值 | covered |
| §15.2 累计 | 任意精度、两位、signed min/max；空count/sum=0 min/max=null | CK-5 | 08 | 极值/退款/空minmax和六笔真值 | covered |
| §15.2 规范交易 | 筛选计数分页前转出ID/time；转入related同UID，否则一致性错 | CK-6 | 04/07/08 | 缺失/跨UID关联，不静默漏掉 | covered |
| §15.2 matched_sides/flows | 明细双侧字段且只一条，只有命中侧统计；收入等单侧 | CK-6 | 07/08 | 三侧联合匹配、两流一笔 | covered |
| §15.2 counts_additive | logical matched_count去重，组内计数不能相加，明确false | CK-6/9 | 08 | 子任务明确可加性；缺直接检查字段与双侧多组计数断言 | gap Medium F3 |
| §15.2 sort/tag | 转账amount必须明确侧；标签备注来自规范交易，两工具一致 | CK-6 | 07/08 | amount侧限制、转入标签保留 | covered |
| §15.2 aggregate形状 | period/稳定ID桶/currency/flow键，名称非主键；totals每期，matched并集，无重叠总和 | CK-7/9 | 08 | 重叠period并集、同名实体；分页totals | covered |
| §15.2 comparison | 同币种flow非时间ID/ordinal；current-baseline、正基数百分比两位 | CK-7 | 08 | ordinal/精确舍入/非正null | covered |
| §15.2 calendar | 日历夹取，非30/365；真实桶端点ordinal，未配对桶无伪比例 | CK-7 | 08 | 闰年/DST/夹取/未配对桶 | covered |
| §15.2 balances | signed/outstanding取负含溢缴；净资产叶子sum；资产负债完整分币种 | CK-8 | 09 | 2000/-300/1700、溢缴、父子/隐藏/分页 | covered |
| §15.2 output | 规范ID/RFC时间原偏移/双侧/分类标签/备注/hide_amount/附件数量，内部键仅游标 | CK-9; HR-2 | 01/07/12 | §15.2准确投影引用、敏感字段排除及hide_amount渲染 | covered |
| §15.2 空关系/N+1 | 合法null分类，必有关系缺失报错，同读视图批量元数据 | CK-6/7/13 | 04/07/08 | 无分类、缺关联、read-view helper；架构明确批量 | covered |
| §15.2 补零限制 | 有限桶按页生成，仅出现过的非时间组合，无笛卡尔积，无币种空 totals | CK-7 | 08 | 主测试明确全部这三类 | covered |
| §15.3 HTTP命名 | 工具snake_case/API camelCase；queries简洁七字段，非重复账本 | CK-9/14; HR-2 | 01/10 | DTO序列化、必要投影与API envelope | covered |
| §15.3 continuation | 单个最后可分页入口、范围展示；验签候选不自动执行、新问题不沿用、下一页明确消息 | CK-10/15 | 06/10/11/12 | frontend多工具、下一页、清空/换用户 | covered |
| §15.3 签名 | version/tool/UID/raw filter/timezone/order/highwater/last/issued/expires/层级摘要，HMAC-SHA256用途隔离 | CK-10 | 04 | 绑定全集引用准确、固定时钟/篡改/换secret/条件 | covered |
| §15.3 验签 | 规范JSON、恒定时间验签、24h/64KiB，验后仍校验实体，层级变更拒绝 | CK-10 | 04 | 时效/大小/归属/层级测试；算法审查应核对恒定时间 | covered |
| §15.3 参数重传 | 可只给cursor；并给条件必须规范相等；不入URL/日志，不新业务表 | CK-10; HR-2 | 04/10/11 | 换条件拒绝、body传输/日志捕获 | covered |
| §15.3 次数 | 执行前扣，整轮不足全停，纠错/切备计轮 | CK-11 | 06 | 整轮额度拦截、fallback和纠错计数 | covered |
| §15.3 请求预算 | 全消息schema/provider状态128KiB及上游限制，不拆工具对/删部分统计伪全量 | CK-11/12 | 05/06 | 完整配对/预算终态/上游边界 | covered |
| §15.3 页预算 | count实长，cursor真实末项不跳行，cursor/结果分别限额 | CK-9/10/11 | 04/07 | 大小预算与页遍历 | covered |
| §15.3 去重 | 请求内规范参数复用；同ID同内容不执行，异内容协议错，无进展终态 | CK-11 | 06 | 缓存/重复ID异参/重复成功 | covered |
| §15.3 历史 | 请求内、完整交互裁旧、当前问题不截断，预算不裁账本 | CK-11/16 | 06/10/11 | 配对不破坏/2048/20→12/Unicode与256KiB | covered |
| §15.3 上游 | 响应/事件独立有界缓冲，不能只限出站 | CK-11/12 | 05 | body/单事件/参数缓冲边界 | covered |
| §15.4 HTTP context | Request.Context组合元数据；handler捕获，异步不用Gin writer/context | CK-13/14 | 02/10 | core元数据/deadline；取消和SSE race | covered |
| §15.4 adapter | 委托Deadline/Done/Err/Value，SessionID保留，nil background | CK-13 | 02 | 对应逐项单元回归 | covered |
| §15.4 tool/provider | query_timeout从整体派生，HTTP和日志不丢取消 | CK-13 | 05/06/10 | 断连上游，剩余整体预算 | covered |
| §15.4 同连接 | 归属/元数据/计数/页/累计同一Tx，提交回滚Close，无另开session helper | CK-13 | 02/04/07/08/09 | 并发写快照、连接释放、helper边界审查 | covered |
| §15.4 三库 | MySQL/PG RR，SQLite同连接，方言内部，不全局改事务 | CK-13/20 | 02/14 | 默认RC库仍验证RR；SQLite query_only恢复竞态和坏连接丢弃 | covered |
| §15.4 主从 | 一工具同连接同源，优先主库不拼主计数从明细 | CK-13 | 02/14 | Master底层池固定连接及主从不混用测试 | covered |
| §15.4 取消/索引 | 真停止SQL/批次并关session；索引依据计划 | CK-13/18 | 02/08/14 | 长查询取消/后续停止；条件索引三库升级回归 | covered |
| §15.4 回归 | datastore/services/API/定时任务；不全局超时 | CK-20 | 02/10/14 | 全仓测试+具体兼容回归 | covered |
| §15.7 六笔 | 精确六笔类型/金额；CNY1000/110/890/40/200、USD5、expense3/-20/100 | CK-5/6 | 01/08/14 | 独立手算fixture，不复用生产计算oracle | covered |
| §15.7 侧/余额 | 目的USD只转入5/详情规范ID；资产2000/负债-300净1700父不重复 | CK-6/8 | 07/08/09 | 同侧联合用例及余额真值 | covered |
| §15.7 隔离库 | SQLite临时复用schema；MySQL/PG独立名清理、opt-in/标识防生产 | HR-2 | 02/14 | 三库启动入口/环境/清理和文档 | covered |
| §15.7 确定性 | 日期/TTL/超时固定时钟；取消断言调用停止/连接释放 | CK-7/10/13 | 02/04/06/14 | 固定now、长查询/可控mock | covered |
| §15.7 配置 | 默认/负零溢出/废弃未知/EBK/EBKCFP表驱动 | CK-17 | 03/13 | 预算及旧段/旧密钥缺失启动 | covered |
| §15.7 环境门禁 | 缺DB模型GUI仍完成其他项，写命令输入，跳过非通过、代码≠可发布 | CK-18; HR-3 | 14 | DoD/发布三状态与补测记录 | covered |

## 缺口与最小修订建议

- **F1 Medium（§6.2 回答口径）**：请求时区计算已覆盖，但与采用记录时区的页面对照时的差异说明，以及“收支净额不是余额/可用资产”目前仅在源要求里，CK-19 最终文本检查没有具名断言。让 10 负责提示词口径、14 负责一例跨时区页面对照与一例净额/余额并存问答；12 仅在 UI 提示需要时参与。
- **F2 Medium（§4.5、§15.1 聚合排序）**：08 有具体来源引用和“组分页不改总计”，因此不判实现 owner 缺失；但没有测试冻结同指标并列的稳定完整键、跨币种独立排名、按指标排序深页无重漏、重算查询时间更新。补 CK-7/9 与 08 测试任务。
- **F3 Medium（§15.2 组计数不可加）**：08 明确 owner 和实现要求“组count可加性明确”，但六笔“两流一笔”不自动检验返回的 `counts_additive=false` 及多个组 count 相加大于 logical count 时的声明。补具名返回字段断言和同一交易两侧多组夹具。
- **F4 Medium（§1/5 deleted、§4.2/15.1 筛选组合）**：已有严格 schema/全部默认值测试和源规则引用，不应称为整个筛选缺失；不过删除记录负例、tag any/all/no_tags 互斥/孤立 tag_match、period 与 filter 时间冲突、comparison 引用不存在 period 的**组合**正确性不等价于“默认值/非法枚举”测试。补 CK-2/3 的具名矩阵并由 04 校验、07/08 查询行为验证，防止全部历史误含已删或组合筛选变宽。

没有建议改变既定架构、增加工具或实施代码；本报告只补来源到验证闭环。


## 统一修订后的独立复核（2026-10-14）

在修订 owner 通知完成后，重新读取主计划中具体用例、相关步骤与总追踪索引，未读取或依据半成品给结论。上面的 gap 为历史初审记录；最新状态如下。

| 缺口 | 新增精确验收与 owner | 复核结果 |
| --- | --- | --- |
| F1 | 主计划 CK-19 Test Cases：净额890.00与当前净资产1700.00不能混称；请求时区/记录时区跨月差异与实际范围。Step 10 提示词及 mock 问答同时约束；Step 14 实际问答与页面对照。 | closed：不再仅依赖源章节引用，已有 owner、独立数值、文本语义断言和最终真实验证；未要求重建历史余额。 |
| F2 | 主计划 CK-9；Step 08 新增 order_by仅group_key或已选metrics、asc/desc并列完整键、多币种各自排名、同名不同ID及跨页并列夹具、深页无重漏/totals不变、修改后新queried_at；Step 14归档三库结果。 | closed：排序、分页、完整性和重算时间分别有可观察断言，保留跨请求非快照语义，没有用单页测试代替深页。 |
| F3 | 主计划 CK-9；Step 08 显式断言 counts_additive=false字段存在且为false，双侧命中转账按账户/flow多组后sum(count)>matched_count，后者仍逻辑去重；Step 14归档。 | closed：直接覆盖原先遗漏的返回字段和不可相加场景；没有把组count重新求和作为逻辑交易数。 |
| F4 | 主计划 CK-2/3；Step 04 具名tag any/all/no_tags互斥/孤立tag_match、period空重名缺边界/与filter冲突/comparison不存在；Step 07/08同夹具验证合法筛选一致、非法不执行；04/07/08/09明确deleted负例与hidden正例，缺必需关联仍一致性错误。 | closed：校验、查询行为和元数据/余额均有owner；不会以删除关联过滤静默吞有效交易，不改变合法null分类契约。 |

已查看 [总追踪索引](harness-requirements-traceability.md)：来源范围→详细矩阵→CK/HR→阶段，以及缺口→具体owner→测试断言两级映射清楚，F1—F4链接和摘要准确；索引仍标“待复审”的文本应由总协调者统一更新，不在本次越权修改。四项修订均沿用既有CK与阶段、未新增依赖或共享文件并发写冲突。**财务审计范围内未发现新的实质缺口；最终实现质量仍需执行这些测试证明。**
