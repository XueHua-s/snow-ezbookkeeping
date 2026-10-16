# 12 · Vite 串行审查

2026-10-16；范围为 `91081998` 相对 `origin/main` 的 assistant experience PR 加 Gate01–11 未提交修复。初始 **8/10**，初审完成时 **9/10**；最终闭环为 **10/10（本地 Vite 适用范围）**：配置契约、实际生产产物与本机 Docker 验证链已闭环；stock Dockerfile/Node26 release build 仍 NOT RUN。下文初审时的未执行项保留为历史记录，最终状态见末节。本分数是审查契约评价，不是技能自带量表，也不宣称整仓无缺陷。

## 实读与适用规则

实读 `.agents/skills/vite/SKILL.md` 及 core-config、core-features、core-plugin-api、build-and-ssr、rolldown-migration；project-spec/global-prompt；vue-best-practices 与 reactivity、sfc、component-data-flow、composables；Vitest 及 core-config/core-cli（仅用于配置和定向运行）。阅读主 audit manifest、Gate11完整报告、Gate03路径结论与相关PR差异。

实际安装 Vite 8.2.2、Vitest 5.0.0、Node 22.20.0；技能的 Vite 8 beta/Vitest 3 示例只作原则参考，以实际 API 为准。配置、资源和 worker 插件归 Vite；Workspace 是组合界面，样式仍由其 scoped src 所有；业务状态及导航归已审 composable/store/双端适配器。此次无需变更这些 Vue 边界。无 SSR、library mode、自定义 runtime 环境，因此不引入相应配置或升级依赖。

## 证实问题与修复

- `vite.config.ts`、`vitest.config.ts`、`vitest.browser.config.ts` 改为同名 `.mts`，明确仅配置模块的 ESM 身份；package 未增加 type:module，现有 `.mjs`/`.cjs` 契约保留。Vite/Vitest默认发现正常，browser script 与 tsconfig 显式引用同步。
- 主配置 JSON imports 补 `with { type: 'json' }`；主/Node测试配置的 `__dirname` 改 `import.meta.dirname`。LICENSE 从配置 URL 解析，Node 内建 imports 使用 `node:`。
- native 实测首次失败揭示 `git-rev-sync` 无参 `short()` 依赖 CommonJS `module.parent.filename`。改为 `git.short(import.meta.dirname)`，保持同一仓库版本标识，无 shim 或警告屏蔽。
- `scripts/check-source-lines.mjs` 纳入 `.mts`、`.cts`，重命名不绕过500行规则。
- `skills/` 与 `.codex/skills/` project-structure 地图同步补三个当前配置事实路径。全仓非依赖/非历史证据消费搜索未发现其他旧路径；历史报告不改写成未实际执行的新命令。

## 构建边界核对

- 保留 `base: './'`、三个HTML入口、静态目录、hash资源名、Framework7 CSS注入、production define和现有浏览器targets；没有把Node配置依赖引入src。
- Worker仍使用Vite的import-meta-url流水线和独立Terser插件；实际构建输出 `sw.monaco-editor-*` 与 `sw.monaco-editor-lang-ts-*`，未改变worker hash/minify契约。
- injectManifest完成，`dist/sw.js`无残留 `__WB_MANIFEST`，precache 60 entries / 1244.25 KiB。assistant GET NetworkOnly/no-store在静态和可配置map规则前注册；activate清理旧应用cache中的assistant响应；API handler/stream亦设置no-store。非GET不被Workbox默认GET缓存路由接管。定向策略测试覆盖部署前缀、query误匹配和选择性旧cache清理，实际离线/更新行为仍留最终E2E。
- 浏览器src搜索未见裸 require/Node imports。保留 `rehype-mathjax/chtml` 浏览器noop alias，避免parser静态导入带入Node实现；Node smoke脚本的createRequire属于Node边界。构建通过不等于所有运行时路径已验证，最终真实Markdown/双端页面仍须容器E2E。
- `AssistantCloudWorkspace.vue` 的 `<style scoped src="./assistant-workspace.css" />` 经实际production编译，产物含 `.cloud-workspace[data-v-…]`、dark全局主题selector，未残留 `:deep(`/`:global(`。没有为了500行压缩CSS或取消scoped。

## 本门执行证据（全部 host，非 Docker）

| 验证 | 结果 / 日志 |
| --- | --- |
| Vite loadConfigFromFile，三个配置 × native/bundle | 6 PASS；`/tmp/gate12-config-load.log`；无ESM-in-CommonJS警告 |
| `npm run build`，真实默认发现配置 | PASS exit 0；`/tmp/gate12-build.log`；4354 client modules，81 SW modules |
| 构建后入口资源引用、Framework7占位符、两worker、SW注入与scoped CSS断言 | PASS；`/tmp/gate12-assets.log` |
| `npx vitest run src/lib/assistant_cache_policy.test.ts` | 1 file / 2 PASS；`/tmp/gate12-focused.log` |
| `npm run test:browser -- src/components/common/assistant/AssistantCloudWorkspace.browser.test.ts` | Chromium 1 file / 5 PASS；`/tmp/gate12-browser.log`；显式新配置路径真正执行 |
| `npx vue-tsc --noEmit` | PASS；`/tmp/gate12-types.log` |
| 三个.mts配置 scoped ESLint | PASS；`/tmp/gate12-lint.log` |
| source line gate | 1695 files，0超过500；`/tmp/gate12-lines.log` |
| 双份地图逐字节比较 / git diff --check | PASS |

保留构建提示：server_settings.js由Go动态提供，经典script不能打包是既有契约；大chunk提示仍在（如Monaco约4MB）；PWA插件内部inlineDynamicImports弃用提示与Rolldown插件耗时提示仍在。没有通过调高阈值/关闭日志掩盖，也没有为本门升级依赖。这些不阻塞当前构建，不构成性能验收完成。

本门冻结：未子委派、未触碰master manifest、未commit/push，未启动或更改任何Docker容器，未读真实凭据/用户DB。最终Docker smoke、应用E2E、完整测试/数据库矩阵及生产页面运行时仍待后续串行阶段；本门不把host构建/组件浏览器证据当作容器证明。


## 最终闭环复核（Gate15–18之后）

本次只复核证据并更新本报告，未重跑无变化测试。已核读 Gate15–18 报告、Gate16/18 的实际 build-provenance JSON、Gate16 main/results.json、Gate18 smoke-results/summary JSON、Gate17 provenance/summary 和原始 `.vitest/audit-matrix/build.log`、`.vitest/final-build-cleanup/build.log`。未发现新的 Vite 缺陷。

- Gate17 `npm run build` PASS，原始日志保留 worker/PWA构建与已说明的耗时/弃用/体积提示；无警告屏蔽。Gate18又独立执行 fresh host frontend build、Linux CGO build及实际Docker image build，记录各命令 exit 0。
- 三份独立记录的107个dist文件摘要完全一致：`ba13bcd1e34f10a293a2717b7a48558c524c9ad16340cc07ae9614b971e5eead`。Gate16/18 Linux可执行文件摘要亦一致；source-tree摘要不同，未声称源码树逐字节相同。故Gate16的生产前端运行证据可归属于Gate18最终产物，无需重复整个业务矩阵。
- [Gate16结果](../ai-assistant-experience-evidence/audit-e2e/main/results.json) 为47 PASS、2 NOT RUN。桌面/移动均以production dist运行，SW实际控制页面，HTTP no-store、assistant CacheStorage响应为0、离线API不能回放历史；真实Markdown复制/导出和应用流程通过。两项bfcache未保留不能算PASS，原始总状态PARTIAL保留。
- [Gate18构建来源](../ai-assistant-experience-evidence/audit-harness-finalization/build-provenance.json) 标识最终镜像 `sha256:b4e96d3244f696663aad9ff84a9c5cfd4c850371ec0a0ed14402bda735fcbd81`；[最终smoke](../ai-assistant-experience-evidence/audit-harness-finalization/smoke-results.json) 10 PASS，包括双端production login无network/runtime errors，以及真实容器重启持久性。Gate15是此前真实Docker基线，不冒充最终产物。

**评分条件：10/10仅覆盖已审的本地Vite配置、打包/资源/worker/PWA边界及上述生产产物的本机容器运行证据。** 前端实际工具链为host Node22.20.0/npm10.9.3；容器为缓存Go/Alpine3.24.2 fallback、linux/arm64、实际项目entrypoint与UID/GID1000，含额外Go工具。这不是stock release build；stock Dockerfile、Node26 frontend stage、精确stock Alpine runtime、amd64仍NOT RUN。跨浏览器、真实bfcache、物理输入/读屏、商业模型质量亦不因本门评分改变。初审残留构建提示及性能边界继续有效，不把本地适用门闭环解释为发布环境认证。
