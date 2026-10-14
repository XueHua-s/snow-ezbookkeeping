# Harness 原方案运行链路覆盖审计

审计范围：原方案 §7—10、§15.3、§15.5—15.6、§17；结合主任务、全部 14 个子任务及实际入口。状态是**计划覆盖**，不代表代码实现或验证已通过。未修改原方案、计划或 todo。

判定：具体条款必须存在实施 owner 和可执行验证；概括引用整个原文不算独立闭环。covered 包括准确章节绑定且相关行为已有测试；gap 表示细节未落实到责任步骤或验证。

## 实质缺口

1. **R-G1 / High：摘要范围不完整。** 原 §8 明确默认同时查询“本月至今收支、上月相同进度比较、分类分布、当前余额”。主任务场景 4、CK-16 测试及 Step 10 仅保留时间比较与指定期间优先，未明确分类分布和余额必查，也没有默认 summary 的完整工具轨迹/结果测试。应在 Step 10 + CK-16 增加四项结果与固定时钟的断言；用户指定全历史/年份时覆盖默认期间而不漏余额。
2. **R-G2 / Medium：系统时钟与角色消息缺少落地验证。** 原 §7.1 的系统消息必须提供后端当前日期、时区、语言；历史作为独立 user/assistant 消息。Step 10 校验角色及时间，Step 01 定义 ConversationMessage，但没有明确替换现有 `buildAIAssistantUserPrompt` 字符串拼接的实现及 wire-level 测试。现有源码 `pkg/api/large_language_models_assistant.go:645` 仍将角色转大写拼字符串。应由 Step 10 构造固定时钟系统上下文和独立角色消息，Step 05/06 测试主备重建均不提升历史为 system。
3. **R-G3 / Medium：多候选协议边界没有测试。** 原 §15.5 要求按 choice/item/call ID 或 index 区分，首期只取一个有效候选。Step 05 写了 call/index/item，但遗漏 choice 隔离与“只取一个”，CK-12 没有多 choice 交错用例。可导致不同候选参数混拼或执行多个答案的工具；由 Step 05 增加交错 choices 及未选候选忽略测试。
4. **R-G4 / Medium：游标细则覆盖不足。** §15.3 的 cursor-only 查询、同时给条件必须规范化相等、canonical JSON、HMAC-SHA256、恒定时间验签主要停留在整体章节绑定。Step 04 与 CK-10 明确签名、时效、换条件拒绝，却没有 cursor-only 合法续查、等价规范参数接受、算法/恒时验证断言。补入 Step 04 精确实现与测试；否则“换条件拒绝”可能错误地禁止所有同时传参。
5. **R-G5 / Medium：供应商超时与上下文限制未形成检查。** §7.2 保留 provider HTTP timeout，整体/query deadline 额外约束；§15.3 尊重供应商明确返回的 context limit。Step 05 的取消/缓冲、Step 06 的 128 KiB 固定限制不能覆盖这两项。应增加 provider timeout 小于/大于 run_timeout 的优先截止测试，以及上游 context-length 错误→预算/未完成映射且不删工具对的测试。现有 common HTTP provider 构造器第 86 行使用 LargeLanguageModelAPIRequestTimeout，替换路径应保留。
6. **R-G6 / Medium：完整观测字段缺少 owner。** §5/12.3 要求工具名、耗时、计数、错误码、request ID，以及返回字节、模型轮次、调用次数、首状态时间、总时长、失败/取消原因。主架构有 SQL 的 request ID/耗时/计数，Step 14 有扫描/内存/时间；Step 06/10 没有完整运行指标实现或断言。补 Step 06 运行摘要、Step 10 首状态/总时长、Step 14 脱敏日志断言；保留不记录财务正文。

## 逐条追踪

| 源条款 | 具体要求 | CK/HR | owner 步骤 | 验证位置/缺项 | 判定 |
| --- | --- | --- | --- | --- | --- |
| 7.1 统一执行入口 | JSON/SSE 共用 runner，任阶段可取消失败 | CK-11/14 | 06/10 | mock 多轮 + 两路同结果/唯一终态 | covered |
| 7.1 中立边界 | DTO 在 llm/data，可选 provider 接口，无 llm→services | CK-20 | 01/05/06 | 接口编译、旧识图回归、review | covered |
| 7.1 系统上下文 | 后端当前时间、时区、语言 | CK-7/16 | 应为10 | R-G2；缺 prompt/wire 固定时钟断言 | gap |
| 7.1 历史角色 | 独立 user/assistant 消息，不拼成系统指令 | CK-16 | 应为10/05 | R-G2；现有角色校验不足以证明真实传输 | gap |
| 7.1 完整轮执行 | 调用 ID/名称/完整参数且全轮成功后顺序执行 | CK-11/12 | 05/06 | 分片/截断/整轮多工具、失败轮不执行 | covered |
| 7.1 纠错 | 结果按调用 ID 配对，允许参数纠正 | CK-11 | 06 | scripted provider 纠错/工具配对 | covered |
| 7.2 预算 | 12轮/24调用/64KiB、重复成功仅复用一次、同错两次退出 | CK-11 | 06 | 边界及同ID异参测试 | covered |
| 7.2 超时 | 180s/30s及剩余整体预算，保留 provider HTTP timeout | CK-11/13 | 03/05/06/10 | 整体/query covered；provider超时 R-G5 | gap |
| 7.2 未完成 | 预算耗尽保留证据/continuation，不伪成功 | CK-9/11/14 | 06/10/11 | mock预算及错误context | covered |
| 7.2 无持久恢复 | 聚合重试不合并部分总额，无运行状态持久化 | CK-11 | 06/08 | 超时不返回部分总额、不持久化 | covered |
| 7.2 续查 | 签名经认证body回传、内存持有、清空/换用户失效 | CK-10/15 | 04/10/11 | 固定时钟与清空/换用户测试 | covered |
| 7.3 配置边界 | openai可配地址，无新增所有provider/硬编码密钥 | CK-17/20 | 05/13 | 保留现有配置与其他provider回归 | covered |
| 7.3 三协议 | Chat JSON、强制stream Chat、Responses 工具链 | CK-12 | 05 | 三协议 mock | covered |
| 7.3 Responses | store=false，必要output/reasoning保留请求内 | CK-12/HR-2 | 05 | 协议状态/日志脱敏测试 | covered |
| 7.3 能力 | 主备工具能力校验，不支持明确失败 | CK-12/19 | 05/06 | unsupported、中立上下文失败测试 | covered |
| 7.3 fallback | 整个请求最多一次；任意正文或可展示思考已发禁止；status可切 | CK-12 | 06 | 明确请求级标记与fallback前后事件测试 | covered |
| 7.3 fallback上下文 | 完成工具对保留；不透传私有状态；失败轮不执行 | CK-12 | 05/06 | 中立转换unsupported、失败轮测试 | covered |
| 7.3 排除错误 | 取消/数据库权限/参数错误不切provider | CK-12/13 | 06 | Subtasks明确，fallback事件测试需纳入该分支 | covered |
| 7.4 证据 | 个人事实先查、总额聚合、空账本可答、引用实际ID | CK-9/19 | 06/10 | 证据排除错误、知识零工具、空账本 | covered |
| 7.4 数值核验 | 真实最终金额对独立真值，不只证明tool-called | CK-19 | 14/10 | CK-19 protocol独立真值；14真实七题核金额/范围/轨迹 | covered |
| 8 路由与历史 | 保留两路/chat-summary/仅user-assistant/12短期消息 | CK-16 | 01/10/11 | 逐条角色长度/页面历史测试 | covered |
| 8 摘要 | 本月至今+上月同期+分类分布+余额；指定期间优先 | CK-16 | 10 | R-G1分类分布/余额及完整默认摘要测试缺失 | gap |
| 8 事件 | status/thinking/reply/references/done/错误语义，提示国际化 | CK-14/15 | 01/06/10/12 | 类型/状态测试，两端i18n；thinking仅可展示由架构限定 | covered |
| 9 责任文件 | API/runner/provider/query/context/UI/settings/docs技能 | CK-18/20 | 01—14 | Expected Changes及各步实施边界完整 | covered |
| 9 MCP | 不经MCP HTTP、不改变外部MCP写权限/协议 | CK-3/20 | 04/14 | 独立白名单及MCP兼容回归 | covered |
| 10.1 旧配置 | 忽略旧段、模型和旧文件密钥，不破主备/开关 | CK-17 | 13 | 无段/无效段/缺密钥启动与无网络/缓存断言 | covered |
| 10.1 新配置 | 主配置正数预算、零非法、集中常量 | CK-17 | 03/06 | 默认/0/负/溢出/EBK/EBKCFP | covered |
| 10.2 旧表 | 停所有读写清理，保留模型/schema/记录 | CK-17/HR-1 | 13 | schema保留、缓存无读写 | covered |
| 10.3 升级回滚 | 备份、三启动路径、矩阵/真模型、关闭AI后应用配置回滚 | CK-18/HR-3 | 13/14 | 文档、rollback演练、实际发布不授权 | covered |
| 15.3 API契约 | snake/camel分界、queries摘要、单continuation取最后分页 | CK-9/10 | 01/06/11 | 序列化、最后入口及候选续查测试 | covered |
| 15.3 游标签名绑定 | UID/tool/原条件/层级/时区/序/水位/末键/时效 | CK-10 | 04 | 专节绑定、篡改/层级/过期/换secret | covered |
| 15.3 游标细节 | canonical JSON、HMAC-SHA256恒时、cursor-only及条件相等 | CK-10 | 04 | R-G4 | gap |
| 15.3 请求预算 | 128KiB包括schema/provider状态、不拆工具对 | CK-11 | 05/06 | 序列化全请求预算、配对测试；provider限制R-G5 | gap |
| 15.3 字节缩页 | returned_count真实、游标不跳行、结果/游标分别预算 | CK-4/9/10 | 04/07 | cursor上限、页面大小预算/无漏页 | covered |
| 15.3 历史/上游 | 2048/20/12、256KiB体、Unicode、省略、响应/事件有界 | CK-11/16 | 05/10/11 | 输入边界、UTF8、历史、缓冲测试 | covered |
| 15.5 最终正文 | 工具轮缓冲，完整无工具终轮才输出 | CK-12 | 05/06 | 混合正文/缓冲测试 | covered |
| 15.5 多候选 | choice隔离、只取一个有效候选 | CK-12 | 05 | R-G3 | gap |
| 15.5 Markdown | 去AIAssistantResult模型序列化，保留识图JSON | CK-12/20 | 10/05 | Markdown外壳与识图回归 | covered |
| 15.5 三种完成 | 模型轮/工具结果/任务完成分离，失败不变空集 | CK-9/11/14 | 06/10 | 工具错误/终态/证据测试 | covered |
| 15.5 引用 | 最多8去重，聚合可无卡片但有完整queries | CK-9 | 06/12 | 去重/聚合证据/hide_amount测试 | covered |
| 15.6 封装 | 既有JSON/SSE/错误context，单终态，取消本地处理 | CK-14/15 | 10/11 | envelope与终态/取消测试 | covered |
| 15.6 parser | 去fallbackDone，EOF/坏JSON/UTF8/CRLF/尾帧/多行/未知/重复 | CK-14 | 11 | 明列测试矩阵 | covered |
| 15.6 生命周期 | request ID、清空/离开/换用户取消、无全局store | CK-15 | 11/12 | 旧回调、清空、双端离开测试 | covered |
| 15.6 超时 | server settings +10秒、SSE同截止、代理文档 | CK-15/17 | 03/11/13 | accessor、截止测试及文档 | covered |
| 15.6 串行写 | 单点写失败cancel，15s注释心跳，结束停timer/race | CK-14 | 10 | httptest写失败与race，架构固定15s | covered |
| 15.6 安全渲染 | Markdown/链接不绕过，引用当前UID再鉴权 | CK-3/15 | 12 | 恶意Markdown/链接与hide_amount | covered |
| 17 前置 | 工作区/AGENTS/技能/锁文件，冲突以本任务方向为准 | HR-1/3 | 全流程/13 | Regular Checks、13双副本去top-k | covered |
| 17 P0 | 可编译契约/fixture，不再次等设计批准 | CK-1/20 | 01/02/03 | Phase1 milestone与六笔fixture | covered |
| 17 交付 | 全工具/协议/取消/双端/配置/文档/验证报告/差异 | CK-1—20/HR | 14 | 三状态与缺环境补测，不把未测计通过 | covered |

## 顺序、并行与代码入口复核

- 01/02/03 分属契约、datastore/core、settings/server-settings；无明显同文件并写。05/container 与06/runner顺序明确。04静态工厂骨架交07/08/09替换独占文件，09全工具测试等待Phase3汇合，不构成实际实现依赖缺口。
- 10后端与11/12前端并行依赖已冻结01类型，契约变化由owner集中处理；13在贯通后退役旧配置保留早期可运行性。14条件索引owner明确，未发现本审计范围内的阻断性并行冲突。
- 真实旧API两handler都调用prepareAIAssistantPromptContext；Step10与13拆分活动链切换/死代码清理可覆盖。普通接口的ResponseJsonObjectType位于第108行，Step10已明确移除。
- common HTTP provider第60行包装日志context，第86行构造供应商timeout；Step05覆盖日志取消，但须补R-G5保留供应商timeout检查。
- 旧history文本拼接位于assistant API第645—650行，是R-G2的真实迁移风险，不能仅通过DTO存在认定已消除。

结论：主备请求级一次切换与任意思考增量阻止fallback、回答关键数字对独立真值这两项已有明确闭环。默认summary漏掉分类分布/当前余额是最直接产品遗漏；其余5项为协议/运维细节的owner或验证缺口。补齐后再复审，当前不应宣称逐条零遗漏。

## 修订后独立复核（2026-10-14，独立交付审计人）

已重新读取修订后的主计划及具体责任子任务，不采用修订作者自评。R-G1—R-G6 均已在**计划层**闭环：R-G1由10/CK-16四项summary及固定同期真值测试；R-G2由10独立角色/后端系统上下文和05/CK-12 wire测试；R-G3由05/CK-12多choice隔离；R-G4由04/CK-10 cursor-only、规范参数相等、规范JSON/签名向量和恒时原语审查；R-G5由05/06/CK-11最早deadline及上下文超限预算终态；R-G6由01/02/04/06/10/14与CK-18完整脱敏观测/单次终结/canary断言覆盖。每项都有具名owner、测试和验收位置，不再仅依赖原文整体引用。

详细复核证据与结构/链接检查见 [交付审查复核](harness-coverage-delivery.md) 的“修订后独立复核”节。本轮未发现这些运行链缺口的剩余实质遗漏。上述初审gap行保留作为历史发现，最新判定以本节为准。未运行业务测试，不代表实现正确或可发布；原计划三库、真实供应商、双端GUI、性能及回滚门禁继续有效。
