# 07 · Pinia 串行审查

2026-10-16；范围为原 PR `91081998` 相对 `origin/main`，包含 Gate01–14 修复及本次最终 Pinia 复审。**初始 6.5/10 → 首轮修复 9/10 → 当前实现范围复审 10/10。** 10/10 仅表示本项已审查的 Pinia 状态所有权、鉴权异步边界与 store 整合没有剩余明确缺陷，不是部署或全产品验收分数。历史记录保留如下；最终 Docker/E2E/full-test/database 矩阵与外部限制仍须分别报告。未修改主 manifest，未委派、commit/push 或操作 Docker。

## 技能与实施前所有权方案

实读 `.agents/skills/pinia/SKILL.md` 及 core-stores、features-composables、features-composing-stores、best-practices-testing、best-practices-outside-component；实读 Vue best practices 与 reactivity/sfc/component-data-flow/composables 四核心参考、software-design-philosophy、project-spec/global-prompt、主 manifest 及 01/05/06 报告。测试参考 Vitest、vue-testing-best-practices。没有引入 SSR、Nuxt、Pinia plugin 或新的持久化插件；不为现有客户端应用增加这些可选体系。HMR 沿用仓库现有 store 约定，本门没有引入热替换协议。

实施前方案记于 `/tmp/gate07-plan.txt` 并向父审说明。最终按实际状态/约束所有权划分，保留 Setup Store 和已有公开 API：

| 模块 | 所有权与接口 |
| --- | --- |
| `assistant.ts` | 组合 surface、身份刷新后的只读 hydration、前台同步、浏览器身份事件。返回所有产品 refs/getters/actions，不导出存储句柄或 AbortController。 |
| `assistant/session.ts` | 身份、存储 owner、草稿/本地快照、clock offset、identity/data generation 和可撤销 lease；负责 reset/close/unsubscribe。 |
| `assistant/conversation.ts` | 有界消息窗口、版本/记忆读取意图、阅读位置、stream/poll/attachment 生命周期；仅注入 session 与同步入口。 |
| `assistant/commands.ts` | exact-payload pending、dispatch、receipt、reconcile、delivery/submitting；依赖三个领域模块和同步入口，不拆成按执行阶段排列的 callback wrappers。 |
| `assistant/history.ts` | 沿用 Gate01 的列表/搜索窗口与请求 ticket，补后台刷新不能抢占显式分页及 dispose。 |

文件均按可读控制流展开；store 约 312 行、session 约 237 行、conversation 433 行、commands 约 417 行，未用多语句压缩规避 500 行门槛。内部 generation/owner/ticket 属于生命周期权限，不是可被 hydration 或组件写入的产品状态。没有新增跨 store 循环，也没有模块顶层 `useStore()`。

## 修复与行为证据

1. **同步 404 清理越界。** inaccessible/404 统一到一条清理路径，在 clearConversation 完成后同时校验 route lease 和 storage owner。受控清理 Promise 后切 route/identity，旧结果不能覆盖新本地快照或导航。基线同 owner 导航回归真实失败。
2. **本地快照顺序不是数据库 row version。** 同一实例的 IndexedDB queue 现在在产生 snapshot 时附带单调 `sequence`，不写入数据库。session 仅在所属 owner 的 lease 有效时接受它。真实浏览器用例证明 page fallback 连续编辑后 cleanup 可使 row version 下降；sequence 仍增加，retryPersistence 恢复 device 时正确接受保留文本。旧完成事件不能覆盖新快照。旧 fixture 没有 sequence 时只在同 scope 比较 version，不跨 owner/scope 比较。
3. **身份、错误和 finally。** initialize/checkIdentity 的旧 401、旧 reconciliation 失败、send finally、pending persistence completion 均在实际副作用边界重验。reset 清 error/loading/submitting、backoff/pollUntil/streamFailures 和私有写去重表。不同身份的新 send 不被旧 finally 清 submitting。
4. **helper→caller 微任务。** 移除 accept 内返回布尔的 save helper，在 await 后的本层直接检查 lease。受控双 queueMicrotask 复现旧 helper 已检查身份、caller 继续时身份却已变化并导航。prepareDataClear 也在 checkIdentity 返回后检查原 identity revision，独立微任务回归证明不能对新 owner reconcile/savePending。普通成功测试不能替代这两条红绿证明。
5. **pending 与产品语义。** 保存完整 canonical pending 后才 dispatch；unknown 回执从不自动重放；拒绝状态和未知送达保持分离；branch target/draft 先持久化，实际路由确认后才删除 marker。新增晚 branch receipt 用例：用户已打开新会话，仍保存恢复目标但不抢导航。全量 clear 的 incomplete 状态及明确确认新操作仍保留；receipt 不作为当前 deletionEpoch 的权威。
6. **同 route 逆序读取。** version-page 按逻辑 message ticket，answer 按返回的 userMessageId 比较请求顺序，已知版本归属在请求时登记意图；不同 message 的并发浏览不互相作废。memory 只接受最新刷新，旧成功和旧失败都忽略。分页不再将捕获的旧窗口覆盖更新的 run snapshot。history/search 原 ticket 仍有效；background refresh 不覆盖正在进行的用户分页。
7. **资源与 dispose。** dispose 关闭 owner/listener、abort stream、撤销所有旧 promise 的权限、清 timer 与 transient command state；晚 initialize 不能恢复 authenticated 或重建资源。detach/dispose 均不发 cancel，离开页面不取消服务端 run。退避按 run 隔离，新 run 不继承旧 run 的 SSE 失败惩罚。
8. **新身份/删除 epoch 的完整 hydration。** checkIdentity 返回新读 generation 的 lease，caller 必须再次检查，随后以全新只读请求重启 hydration。旧写续程仍失效。测试覆盖同 owner deletionEpoch、clear receipt、capabilities 直接发现 A→B（没有先手动 reset）的 history 和 polling；新轮询保留，dispose 后 timer 为零。

## 实际验证

- 基线回放：临时恢复旧 `assistant.ts`，新 lifecycle/readOrdering 共 **12 FAIL / 9 PASS**，`/tmp/gate07-red.log`。还出现一条基线 disposal 后调度访问已移除 document 的 unhandled rejection；修复后最终测试无 unhandled error。
- 微任务红灯：临时恢复旧 store 并去掉 dataClear 的新增检查，receipt helper 和 checkIdentity helper 两条均 **FAIL**，`/tmp/gate07-microtask-red.log`。该命令按名字筛选，另有 1 PASS/12 skipped；未把 skipped 计作验证成功。
- 最终 focused Vitest：assistant、lifecycle、readOrdering、dataClear、dataClearAuthority，加 Gate05 command/workspace lifecycle 集成：**7 files / 102 tests PASS，0 skip，0 unhandled error**，`/tmp/gate07-focused.log`。
- Chromium 原生 IndexedDB：`assistant_local_storage.browser.test.ts` **19 PASS**，包括新增真实 fallback/cleanup/recovery sequence 用例，`/tmp/gate07-storage-browser.log`。这是 storage browser suite，不是双端完整应用 E2E。
- `npx vue-tsc --noEmit` PASS，`/tmp/gate07-typecheck.log`；本门改动 TS scoped ESLint PASS，`/tmp/gate07-lint.log`。
- `npm run check:source-lines`：1689 files / 0 exceed 500；`git diff --check` PASS。
- 父审独立执行当前 store/lifecycle/readOrdering/dataClear 与其环境隔离测试：5 files / 64 tests PASS（父审消息，`/tmp/ebk-parent-gate07-final.log`）。独立结果与本门 102 测试不混算。
- 中途新 shared fixture 的 hoisted export 写法、独立 helper test 未 mock transport 导致 suite 初始化失败，均已修正并实际重跑；不把失败前的部分绿灯当作最终验证。

## 首轮交接时的后续门范围（历史）

D5 的压缩控制流与状态归属已整改，请父审/Gate01 对实际接口深度复审。剩余分数不是以测试数量补齐：Gate08 仍需真实 Vue Router/Framework7 reuse、native storage event、bfcache 和 reading anchor；Gate09 负责滚动/RAF；完整 Docker/database/unit/E2E 和 UI 视觉、屏幕阅读器证据仍未执行。Gate13/14 的 deviceIdentity 首次多 tab 竞争 lead 未在本门扩大修复范围。Vite native config ESM warning 保留交 Gate12。


## Gate14 后最终复审

重新读取实际 Pinia / Vue best practices 技能及四个 Vue 核心参考，复读本报告和 Gate14 报告；独立检查 `token.ts`、`services/{transport,authorization,assistant_conversations}.ts`、DesktopApp/MobileApp/Unlock 三个调用方与 assistant session 的交界。Gate01 最新设计复审已经关闭 D4–D6；本次没有为了评分重拆模块。

### 已确认的所有权与生命周期

- token Setup Store 在 setup 内获取 settings/user stores，没有模块顶层实例或 await 后重新寻找 active Pinia。产品状态仍由原 settings/user store 持有；refreshPromise 是私有在途合并句柄，不是另一份 token 状态。
- transport 拥有实际 refresh/queue：成功和失败均 settle/drain，只有 refresh 在同 owner 的 token 改变时最多尝试3次；普通 POST 不自动重放。当前 token 的401保留原退出规则，旧 Authorization 的晚401及旧成功不能清除/覆盖新身份。
- refresh/revoke 的 expectedToken 与助手 get/post/export 的 expectedOwner 在 Axios interceptor 实际派发前验证；队列恢复再次检查 owner/token。同 owner 正常 refresh 可以使用新 token，未把助手 payload 错绑到过期 token。
- token store 拥有 transport 返回后的用户设置应用与 best-effort revoke；assistant session 继续拥有 ownerScope/deletionEpoch/route leases、pending 与 stream 生命周期，两者没有重复保存或抢写 token。SSE fetch 在调用时同步捕获 token。
- 三个启动调用方和 best-effort revoke 都处理拒绝；没有挂起外层 Promise 或未处理的清理拒绝。

### 本次实际发现并修复：store → 启动调用方的最后一层 continuation

原 Gate14 检查覆盖 transport → token store，但 DesktopApp/MobileApp/Unlock 在 store Promise 成功后的 `.then` 仍立即应用旧 locale/notification。使用真实 `useUnlockPageBase` + 真实 token store，在 store 应用用户信息时排入 queueMicrotask 切换 owner 或同 owner token：两条都实际调用了旧 language，红灯日志 `/tmp/gate07-final-caller-red.log`（2 FAIL）。没有用 sleep 或 fake store 的布尔结果代替这条调用链。

最小修复由 token store 返回 `RefreshedSession`：保留原响应字段，增加 `isCurrent(): boolean`，闭包隐藏本次已提交的 owner/token。三个调用方只在各自 `.then` 的实际副作用边界检查该 lease；没有把 token 比较规则复制到 UI，没有通用 task framework，也没有新增持久状态。该方法属于本地 bootstrap 结果，不修改 HTTP DTO 或服务端协议。新增成功用例同时证明正常当前身份仍应用 locale/notification/cache/rates 更新。

### 本次定向证据与冻结边界

- `token.test.ts`、`transport.test.ts`、新增 `UnlockPageBase.auth.test.ts`：**3 files / 24 PASS，0 skip，0 unhandled error**，`/tmp/gate07-final-auth.log`。其中 Gate14 原有21项保留，新增3项是两个微任务拒绝与正常当前身份成功路径。
- `npx vue-tsc --noEmit` PASS，`/tmp/gate07-final-type.log`。
- 本次 token/三个调用方/新增测试 scoped ESLint PASS，`/tmp/gate07-final-lint.log`。
- source-lines：1705 files / 0 exceed 500；diff-check PASS，`/tmp/gate07-final-lines.log`、`/tmp/gate07-final-diff.log`。
- 父审独立复跑本次3 files / 24 PASS，`/tmp/ebk-parent-gate07-auth-final.log`；父已检查 lease 接口与正常路径断言。
- Gate14 的真实双端 native 14场景连续两轮零 pageerror、父独立21 auth及16 browser结果来自其报告与父审；本次未重复运行，也不冒充本次新增 lease 的完整应用 E2E。

当前 Pinia 实现范围没有剩余明确阻断，报告冻结。新增 `RefreshedSession` 接口交父审/Gate01仅复核这一小段 diff。router 后续复审、最终 Docker/E2E/full-test/三数据库矩阵继续串行执行。真实 bfcache 的 persisted 证据、商业模型、物理输入法/键盘及手动屏幕阅读器仍不能从本次定向测试推导；首轮记录中的 warning/deviceIdentity 等后续门事项以相应最新报告为准。

## Final container E2E compatibility follow-up (2026-10-16)

The serial final E2E reviewer reproduced an old-backend regression with the
current production frontend and explicitly built old-main backend
`1a21bc9bfd2b155a3ed9b8ac57caff7c76bbb293`: capabilities returned real HTTP404,
but the UI showed only a generic request error. Concurrent initial `pageshow`
foreground synchronization superseded `initialize`'s capability ticket;
only `initialize` classified the 404 as a legacy backend. The current
synchronization request correctly rejected but never set the upgrade notice.

Capability compatibility is now owned by `session.checkIdentity` after its
existing ticket/epoch/disposal fence. `initialize` no longer classifies an
unrelated history404 as missing capabilities. No new state, watcher or public
API was added. Four controlled regressions cover overlapping initialize/sync,
upgrade recovery, late obsolete404 after success and unrelated history404.
Before repair:3 FAIL/1 PASS; after repair plus existing store/lifecycle/read
ordering:4 files/65 PASS. Scoped ESLint passed. Final rebuilt-image E2E and
parent review are recorded in Gate16; this addendum does not independently
raise the prior score or claim completion of external release checks.

## Gate16–18 后最终增量闭环（2026-10-16）

**当前 Pinia 实现范围维持 10/10。** 本次独立复核仅针对上段 compatibility 增量，没有重跑已通过的测试或修改产品代码。

实际检查 `session.checkIdentity`、`initialize/synchronize` 和 `compatibility.test.ts`：404 分类仅发生于 capabilities 请求的 catch，且位于 disposed/epoch/current ticket 检查之后；过期成功、过期失败不能覆盖当前身份的兼容状态。成功身份初始化清除 legacy 标志，reset 仍清理该状态。`initialize` 的通用 catch 仅处理错误，不再把 history 404 当作旧后端。initialize 与 foreground synchronize 共用同一 session 权威，没有增加重复状态、watcher 或公共接口；原命令持久化、身份 lease、导航及资源生命周期契约不变。四项回归分别覆盖并发 supersede、成功升级恢复、迟到旧404和无关 history404，未发现本增量的明确剩余缺陷。

核对现有证据而非重复执行：`/tmp/final-compatibility-green.log` 为 **4 files / 65 PASS**；Gate16 记录修复前 **3 FAIL / 1 PASS**。实读保存的 [compat-results.json](../ai-assistant-experience-evidence/audit-e2e/compat/compat-results.json)：最终 production dist 对实际隔离 old-main `1a21bc9bfd2b155a3ed9b8ac57caff7c76bbb293`，desktop/mobile **2 PASS**，各两次真实 capabilities HTTP404、零 provider 调用，显示明确升级提示，无刷新循环。该 old-main 场景是隔离 host binary 证据，不改称容器内旧后端测试。

下游发布验证状态以 [Gate15](15-docker-smoke.md)、[Gate16](16-container-e2e.md)、[Gate17](17-final-test-matrix.md)、[Gate18](18-harness-finalization.md) 的实际环境和限制为准。Gate18 记录父审最终 `npm run test` **44 files / 38,917 PASS**；这是父审执行结果，不是本次复跑。历史段落中“后续待执行”保留为当时交接记录，不表示这些已完成阶段仍被本门阻塞。

外部 NOT RUN 不变：实际 bfcache persisted、商业模型、物理键盘/输入法、手动屏幕阅读器，以及 Gate18 明列的 stock Dockerfile/Node26/精确 stock Alpine 构建均不能由 Pinia 定向回归推导为通过。本门仅更新本报告并冻结；无主 manifest 修改、委派或提交。
