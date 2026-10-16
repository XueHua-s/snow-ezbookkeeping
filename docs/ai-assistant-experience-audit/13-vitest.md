# 13 · Vitest 串行审查

**最终本地门评分：10/10（2026-10-16 闭环复核）。** 下文初始 7/10、阶段修复后 9/10 及当时待办均保留为历史；最终证据和适用边界见末节。此分数是当前本地 Vitest 审查范围评价，不是技能自带量表、商业模型认证或所有环境零缺陷保证。

2026-10-16；范围为 `91081998` 相对 `origin/main` 的 assistant experience PR，包含 Gate01–12 未提交修改。初始 **7/10**，修复后 **9/10**；设计简化评分同为 **9/10**。已证实的生产身份/初始化清理缺陷已修复，完整 Docker、实际双端 E2E、最终 unit/Go/三数据库矩阵尚未执行，不能给出全链路 10/10。没有子委派、修改主 manifest、commit/push、操作 Docker 或读取私密配置。

## 实读和适用范围

实读 Vitest SKILL 与 core-config、features-mocking、advanced-environments、core-hooks、features-concurrency；project-spec/global-prompt、software-design-philosophy；Vue testing best practices 及 async-await-flushpromises、component-blackbox-approach、browser-vs-node-runners、pinia-store-setup、no-snapshot-only；Vue best practices 与四个必须参考 reactivity/sfc/component-data-flow/composables；主 manifest、Gate07/12 报告。实际运行 Vitest 5，技能中 Vitest 3 的示例不作为当前 API 契约。

测试分层保持：Node 覆盖协议、服务契约和状态机；happy-dom 覆盖组件事件/生命周期；Chromium 覆盖 IndexedDB、native dialog、布局和 selection；真实 Go 后端的应用流程归后续 E2E。未新增 Vue 组件或改变 props/emits，既有 Workspace 组合界面、session/conversation/commands/store 所有权不变。

## 证实缺陷和修复

### 生产设备身份与数据恢复

原 browser helper 恒定注入 `deviceId='device'`，掩盖生产 localStorage get/generate/set 与 IndexedDB 两份身份权威。新增独立 production initializer suite，三个回归先红：保存后删 localStorage 标识导致草稿/pending 不可见；旧双身份草稿不可恢复；受控两个初始化读取均为空时产生不一致 deviceId。不是通过测试注入 ID 来证明生产初始化。

新版用同源 IndexedDB 数据库作为设备边界，默认路径使用稳定 `browser` 命名空间；不再读取或写入 localStorage 身份。owner transaction 原子归一全部旧 ID，保留每条 draft/pending 及精确 payload、rejected/resolvedConversationId；同会话多草稿标为冲突。显式传入 deviceId 时仍按该设备隔离，是否显式传入以 undefined 判断。迁移提交增加 row version 并仅广播 owner/version，重复 load 不写入、不广播。quota、page fallback、retry 合并、epoch fence、exact pending 校验继续由同一个 owner transaction 承担。

迁移文档同时明确旧前端打开标签/回滚可见性限制：旧客户端不理解新命名空间，须刷新；记录没有删除，旧客户端稍后新增的旧 ID 记录可再次归一。没有新增 metadata store、异步身份 API 或泛用持久化 DI 层。

新增 Chromium 回归同时覆盖 legacy branch target/原文、重开后精确 pending、重复请求去重、owner 隔离、迁移后 epoch 拒绝、migration-only notice。原生 IndexedDB 原有 19 项覆盖 explicit-device、并发 quota、conflict、UTF-8 byte quota、page retry、global clear receipt 等全部保留通过。初始化竞争证据是受控旧读取屏障 + 原生 IndexedDB 并发事务，不声称真实两浏览器标签测试；实际跨 tab 流程仍归 E2E。

### IndexedDB 晚到连接

新增两个先红回归证明 onblocked 拒绝之后的 onsuccess，以及 close 发生在 await open 期间，都会留下未关闭连接。现 blocked 后成功立即 close；storage close 后晚到连接也 close 并拒绝操作，不退化为一个看似可用的 page snapshot。测试通过受控 IDBOpenDBRequest 生命周期驱动事件，断言 close 确实发生。

### E2E 环境失败清理

原环境在 provider/backend 已启动后，ready/Vite 失败没有清理路径；build/backend spawn error 没有监听。旧代码 Vite create/listen 两个回归实际失败，provider 仍在监听。现同一 close 函数服务成功退出和部分启动失败，尝试释放 Vite、任务自建 child、provider、端口 reservation、exit hook 和 log；不会清理其他进程/容器。close 收集并报告拒绝，启动错误与清理错误同时保留。

backend 就绪 fetch 设 1000ms AbortSignal deadline；停止先 SIGTERM，3 秒未退出则仅对该 child SIGKILL，再 1 秒未结束则拒绝。测试覆盖构建 spawn、backend spawn、ready 退出、Vite 创建/监听、清理拒绝仍释放其他资源、SIGTERM 超时升级、readiness deadline 取消。没有把 spawn 错误中的环境值写入日志。INI section 查找改为匹配实际 section 行；原 indexOf+1 后检查 <0 永远不能识别缺失 section，现明确抛错。

父新增 EBK_/EBKCFP_ 环境隔离回归和 Gate10 answerContent fixture 均保留。测试只使用 synthetic binary/env 和 task-created localhost HTTP servers，没有启动真实 Go/provider 或接触用户配置。

## 测试审查与边界

- 两套 `.mts` 配置分开 Node/browser discovery，Node isolate/clearMocks/restoreMocks/mockReset 保持，browser 不冒充 happy-dom geometry。无框架或依赖升级。
- store fixture 每项 fresh Pinia，dispose、unstub globals、real timers 在 afterEach；组件 mount/unmount 和受控 deferred 断言覆盖 owner/navigation 与 helper→caller 微任务边界。已有两个 setTimeout(0) 用于耗尽完整 Promise 链，不是等待网络或任意秒数的 flaky sleep，本门未机械改写。
- 查阅 stream 的 split UTF-8、terminal/revision/amount 断言、service exact-payload/no-retry/no-store、store lifecycle/readOrdering、组件确认/导出和原生存储测试。协议 fixture 明确只模拟服务，不据此声称商业模型质量。新增测试断言可观察数据/资源结果，不是复制实现或仅快照。
- `rg` 检索 `src`/`scripts` test 文件的 `it/test/describe.skip/todo/only` 无匹配；assistant suite 无 snapshot-only 断言。红运行通过 `-t vite` 选择测试所报告的 2 skipped 是过滤结果，不是代码库存 skip。最终运行无 skip。
- 未新增“通过率”或 coverage 百分比结论；全仓覆盖率、实际 bfcache、物理输入法/屏幕阅读器和真实跨 tab 用户切换仍须各自正面证据。

## 执行证据

| 检查 | 结果 / 日志 |
| --- | --- |
| production identity 初始红 | 3 failed；`/tmp/gate13-identity-red.log` |
| 晚到 IDB connection 初始红 | 2 failed；`/tmp/gate13-resource-red.log` |
| 原版环境 Vite create/listen 初始红 | 2 failed；`/tmp/gate13-cleanup-red.log`；运行后立即恢复当前文件 |
| final focused Node/happy-dom：environment 两文件、store assistant/lifecycle/readOrdering、Workspace lifecycle/composable | 7 files / 113 PASS；`/tmp/gate13-focused-final.log` |
| final native Chromium storage + production identity | 2 files / 26 PASS；`/tmp/gate13-browser.log` |
| 环境失败路径 + 父环境隔离 | 2 files / 9 PASS；`/tmp/gate13-cleanup-final.log` |
| vue-tsc --noEmit | PASS；`/tmp/gate13-types.log` |
| 改动 TS/浏览器测试 ESLint | PASS；`/tmp/gate13-lint-final.log` |
| mjs 语法检查 | 两文件 `node --check` PASS |
| source line gate | 1697 files / 0 超过500；`/tmp/gate13-lines.log`；storage 425行，environment 最终174行 |
| git diff --check | PASS |

仓库 ESLint 默认忽略 scripts/*.mjs；尝试 --no-ignore 后 TypeScript project service 明确不包含这两个文件（`/tmp/gate13-script-lint.log`）。因此不把它们称为 lint PASS，也未为本次测试修改全局 tsconfig/lint 体系；Node parser 和实际 Vitest 执行均通过。这是工具覆盖边界，不是运行失败。

本门冻结。后续 Gate14 应继续实际组件/路由行为审查；最终串行矩阵应重跑现有完整测试和真实环境启动/关闭，并纳入生产身份无注入的真实跨 tab 验证。此前各门修改没有被回滚；主报告和最终 commit/push 由父审控制。


## 最终本地闭环复核（2026-10-16）

本次仅核读并更新本报告，没有重复无变化测试、子委派、主 manifest 编辑或提交。实读 Gate14/16/17/18、当前 bounded-command/build-resources 实现及其十项测试，核对 Gate16 main/results.json、Gate17 summary.json、Gate18 summary.json 和实际原始日志。Gate18 summary 中全部日志路径的 SHA-256 与当前文件逐项相符，包括父最终全量日志；不是只依赖报告中的 PASS 文案。

原先限制本地门评分的条件已关闭：

- **生产初始化真实跨 tab**：Gate14 最终两轮各14项，Gate16 最终容器应用47 PASS/2 NOT RUN；持久结果逐端记录未注入 deviceId 的两个真实 app tabs 并发冲突草稿保留、legacy key 丢失后 reload 恢复。trusted storage 首帧身份清除、确认/编辑/导出、epoch 推进和两 owner 服务端不变均有对应结果。此前受控屏障回归保留，未冒充原生事件证明。
- **完整本地矩阵**：Gate17 前端7 files/62 Chromium PASS、完整 lint/build、Go/vet/race、SQLite/MySQL/PostgreSQL及原生备份恢复/升级回滚证据已完成；Go opt-in skips 的性能项随后显式执行，商业 TestAIDeploymentIntegration 仍 NOT RUN。测试模式和实际叶子计数按 Gate17区分，不重复累计为独立场景。
- **最终 harness 生命周期**：Gate18 将临时 Go build 和命令调用接入有界 POSIX process-group helper，并以任务自建容器名/目录资源所有者处理不确定创建与退出恢复。实读测试使用真实 resistant child/grandchild、真实信号及可执行 CLI 故障模拟，覆盖 deadline、listener cleanup、caller finally、16MiB capture、精确容器 cleanup、失败可观察性；三次 mutation 分别实际红1/2/1，最终5 files/21 PASS。真实 Docker timeout 日志确认 CLI 停止后容器仍存在、显式 cleanup 随后删除；模拟 CLI 与真实 daemon 证据没有混称。
- **最新全量测试**：父独立 `/tmp/ebk-parent-final-unit.log` 实际为44 files/38,917 PASS、0 skip，包含 Gate18新增十项；不是本 reviewer 重跑。最新 validation build/smoke10项通过，dist107文件及 Linux executable 与 Gate16完全相同，后续仅 harness/test 修改没有使既有业务 E2E失效。

因此当前 **Gate13本地适用规则 10/10**，无新增具体测试缺陷，原9分所列本地待验证条件均已闭合。早期设计9分保留为当时评估；全项目最终设计评分归 Gate01，不由本测试门替代。脚本 ESLint项目范围的既有排除仍明确保留，当前脚本有语法、行为、失败注入和实际运行证据，不改称 ESLint PASS。

外部和平台限制不变：实际 bfcache 两端均 pageshow.persisted=false，仍2 NOT RUN；商业主备模型/财务与长期对话质量、物理手机键盘/输入法、人工屏幕阅读器、其他浏览器引擎、amd64及 stock Dockerfile/Node26阶段均未获新证明。当前分数只关闭本地审查门，不将这些项目算作通过，也不宣称全链路验收完成。
