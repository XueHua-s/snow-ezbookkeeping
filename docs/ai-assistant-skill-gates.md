# Assistant PR #1 skill-gate remediation

Reviewed base: `afc20332a3c5579fe98c9f9feb57eb7e500c2d24`. This follow-up changes the assistant frontend and its test infrastructure. It does not change the seven-tool service, model protocols, database access or retained embedding data.

## Component and contract map

| Owner | Responsibility | Public contract |
| --- | --- | --- |
| Desktop/mobile `views/*/assistant/AssistantPage.vue` | Route entry and platform shell/props | Mobile receives the Framework7 router |
| Platform `components/*/assistant/AssistantConversation.vue` | Compose the feature, bind session actions, report platform errors and navigate to references | Mobile receives the Framework7 router; session stays local |
| Platform `AssistantComposer.vue` | Render input and send control | String `v-model`, disabled/requesting/canSend props, `send` event |
| Platform `AssistantActions.vue` | Render explicit next-page and cancellation controls | requesting/canRequestNextPage props, `next`/`cancel` events |
| Common `assistant/AssistantMessageList.vue` | Render messages, empty state and scrolling viewport | Readonly messages, enabled/platform variant, `openReference` event, empty-state slots |
| Common `AssistantMessageEvidence.vue` | Present validated query evidence and masked reference amounts | Message prop, validated reference-ID event |
| `useAssistantPageBase` | Own request identity, cancellation, completed history and continuation | Explicit `AssistantSession`; only input is writable by consumers; messages and nested evidence are deeply readonly |
| `useAssistantAutoScroll` | Follow actual content/viewport size changes and disconnect on disposal | Two native element refs; no dependency on message fields or platform component internals |
| `src/lib/assistant_stream.ts` | Validate bounded SSE frames and terminal completion | Contracts in `src/lib/types/assistant.ts` |

The shared message renderer eliminates duplicated knowledge of message fields. Platform components retain their different widget and styling contracts rather than introducing a generic widget adapter. Message/query/reference presentation is derived with computed values. Route views no longer implement the complete feature, and the duplicate desktop composer markup is replaced by a single component.

The desktop conversation viewport now has a bounded height: long replies scroll inside it instead of increasing the page indefinitely. Mobile send controls use native button semantics and intrinsic width, leaving space for the textarea instead of allowing Framework7's full-width button style to collapse it.

## Gate decisions

- **software-design-philosophy:** repaired scope meets the target design criteria (10/10 design assessment for this scope, not a correctness guarantee for the entire repository). Session/protocol complexity remains behind existing interfaces, field-sensitive scrolling knowledge is removed, and tests use supported tooling. The score is supported by the ten-point follow-up checklist below; it is not inferred from test counts.
- **vue-best-practices:** Composition API and script setup; focused feature/list/input/action boundaries; typed props/events; single-owner session state; computed presentation; managed observer and request lifetime.
- **vue-router-best-practices:** desktop uses a real leave guard and mobile uses Framework7's page lifecycle. Disposal, deactivation and user changes still invalidate stale callbacks. Framework7 is not replaced with Vue Router.
- **vue-testing-best-practices:** Vue Test Utils with Happy DOM replaces manual SFC compilation, custom host nodes and direct callback clicks. `flushPromises` replaces fixed rounds of timers. Real Chromium tests use actual Vuetify/Framework7 controls, actual routing, real keyboard input and layout; service responses remain deterministic mocks.
- **vueuse-functions:** `useResizeObserver` tracks content and viewport size, including asynchronous Markdown layout, and automatically disconnects on component disposal.
- **antfu:** production contracts and constants have dedicated files; the session API has an explicit return type; tests use describe/it and meaningful behavior assertions. Existing repository naming, lint configuration and platform layout are retained rather than imposing a repository-wide tooling migration.

## Reproduction

```sh
pnpm install --frozen-lockfile
pnpm exec playwright install chromium
pnpm run lint
pnpm test
pnpm run test:browser
pnpm run build
BUILD_PIPELINE=1 go test ./...
BUILD_PIPELINE=1 go vet ./...
```

The browser suite runs the desktop viewport at 1280×900 and mobile at 390×900. It verifies Shift+Enter versus Enter on desktop, actual mobile input/send, long-response scrolling, explicit continuation, cancellation, stale terminal callbacks, and cleanup through real Vue Router and Framework7 navigation. Screenshots are saved under the test directory's ignored `.vitest/` folder. Platform-specific global CSS is isolated to match the separate application entries.

Results on 2026-10-15: 17 frontend test files / 38,719 tests pass; both Chromium platform scenarios pass; frontend type/lint checks, production build, full Go test command and Go vet pass. Frozen pnpm installation succeeds. Screenshot inspection confirms visible input controls and bounded message areas on both platforms.

These browser tests exercise the actual feature components and widget libraries, with mocked service responses and desktop detail-dialog feedback. They are not a real-model financial acceptance run, live backend end-to-end test, physical-phone certification, or a rerun of opt-in MySQL/PostgreSQL integration suites. The original PR's real-model deployment gates remain as documented in the harness verification reports.

## Independent follow-up review

The previous 10/10 statement was premature: follow-up review found IME Enter submission, loss of unsent drafts during pagination, and shallow-only protection of nested session evidence. These have now been repaired. New regressions reproduced both interaction defects before the changes (six failing cases across session and platform tests), then passed after the fixes. A separate mutation regression verifies deep readonly protection while session actions still work.

The scoring rubric below awards one point for each verified criterion. **Current remediation score: 10/10**, limited to the six requested skills and the frontend changes covered here. This is an engineering review judgement, not proof that no undiscovered bugs exist or approval of the original PR's pending deployment gates.

| Criterion | Evidence / result | Score |
| --- | --- | --- |
| Coherent module boundaries | Route shells, feature orchestration, list, composer and actions each have one responsibility; no generic widget abstraction | 1/1 |
| Information hiding | No platform code enumerates message fields for scrolling; SSE validation stays in its owner | 1/1 |
| Explicit typed contracts | Dedicated session/protocol types and constants, typed props/events and session return type; vue-tsc passes | 1/1 |
| Single state owner | Deep readonly messages/continuation prevent nested reference tampering; mutation regression passes; actions remain effective | 1/1 |
| Safe presentation | Computed query/reference projection; hidden amounts and malicious Markdown tests pass | 1/1 |
| Input semantics | IME confirmation and WebKit code 229 do not submit; normal Enter submits; Shift+Enter remains a newline | 1/1 |
| Draft and continuation ownership | Pagination sends its own explicit question without replacing input; completed/canceled/failed cases preserve the draft on both platforms | 1/1 |
| Lifecycle and layout | Real Vue Router/Framework7 navigation cancels requests; stale callbacks are ignored; VueUse owns observer disposal; real long-content scrolling passes | 1/1 |
| Supported behavior tests | Standard DOM mounting/flushPromises plus real Chromium widgets, keyboard and geometry; no custom renderer or fixed-cycle waiting | 1/1 |
| Reproducible validation | Frozen install, full frontend suite, browser suite, lint/type checks and production build pass; prior Go checks remain applicable because no backend changes were made in this follow-up | 1/1 |

The IME regression dispatches composition keyboard events in the DOM test environment; it does not claim testing a particular operating system's physical IME implementation. Chromium coverage continues to use real key input for Enter/Shift+Enter and now also checks draft preservation with real platform widgets.
