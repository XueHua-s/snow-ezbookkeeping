# ezbookkeeping Project Structure Map

## Top-Level Layout

- `cmd/`: app wiring, startup, web server, database bootstrap.
- `pkg/`: backend application code.
- `src/`: Vue frontend code (desktop + mobile).
- `conf/`: config template (`ezbookkeeping.ini`).
- `templates/`: server-side templates, including AI prompt templates.
- `docs/`: migration and operation docs.
- `vite.config.mts`: ESM production/dev build, assets, worker and PWA configuration.
- `vitest.config.mts`, `vitest.browser.config.mts`: ESM Node/component and browser test configurations; `.mts` preserves the package’s existing CommonJS default.

## Backend Modules (`pkg`)

- `pkg/api/`: HTTP handlers and request/response orchestration.
- `pkg/services/`: domain services and data access composition.
- `pkg/models/`: database models and API payload structs.
- `pkg/settings/`: config schema, defaults, and config-file/env loading.
- `pkg/llm/`: provider-neutral runner/contracts, provider container and provider adapters.
- `pkg/datastore/`: database sessions, shard selection and connection/transaction primitives; assistant business rules remain in services.
- `pkg/templates/`: prompt/template loading.
- `pkg/storage/`: object storage adapters.

## Frontend Modules (`src`)

- `src/views/base/`: shared page logic for desktop/mobile pages.
- `src/views/desktop/`: desktop Vuetify pages.
- `src/views/mobile/`: mobile Framework7 pages.
- `src/components/common/assistant/`: shared assistant rendering; `src/components/desktop/assistant/` and `src/components/mobile/assistant/`: platform navigation/reference adapters.
- `src/stores/assistant.ts` and `src/stores/assistant/`: shared conversation state and reconciliation.
- `src/lib/services/`: endpoint-family clients and shared transport; `src/lib/services.ts`: existing facade.
- `src/lib/server_settings.ts`: frontend accessors for server-provided feature/config flags.
- `src/router/desktop.ts`, `src/router/mobile.ts`: route entry points.

## AI Feature Touchpoints

- Receipt image recognition endpoint: `pkg/api/large_language_models.go`.
- Legacy stateless assistant JSON/SSE endpoints: `pkg/api/large_language_models_assistant_harness.go`; mode/locale helpers: `large_language_models_assistant.go`.
- Assistant runner/protocol: `pkg/llm/assistant_runner.go`, `pkg/llm/data/*`, `pkg/llm/provider/openai/openai_assistant_tools.go`.
- UID-bound read-only tool queries: `pkg/services/ai_assistant_queries.go` and its query helpers; fixed-connection read views: `pkg/datastore/assistant_read_view.go`.
- Legacy assistant adapter/SSE/evidence: `src/views/base/assistant/AssistantPageBase.ts`, `src/lib/assistant_stream.ts`, `src/components/common/AssistantMessageEvidence.vue`.
- AI provider adapters: `pkg/llm/provider/*`.
- AI prompt templates: `templates/prompt/*.tmpl`.
- Legacy embedding schema/rows retained for rollback: `pkg/models/ai_assistant_embedding.go`, `cmd/database.go`; no cache service or active assistant embedding config/calls.
- Assistant config/operations: `docs/ai-providers.md`, `docs/ai-storage-migration.md`, `docs/ai-assistant-harness-verification.md`.

## Safe Change Pattern

1. Adjust schema, defaults, loading and validation in `pkg/settings/`: `setting_config.go` holds Config, `setting_load.go` wires loading, and `setting_ai_experience.go` owns persistent assistant settings.
2. Wire config to provider/api layer in `pkg/llm` or `pkg/api`.
3. Expose general runtime flags in `pkg/api/server_settings.go`; persistent assistant availability, effective limits and owner/epoch state belong to `AssistantCapabilitiesHandler`, not duplicated general flags.
4. Update the corresponding frontend accessor/client and shared state/page logic before platform adapters.
5. Update `conf/ezbookkeeping.ini` and relevant docs in the same commit.

## Persistent Assistant Ownership

- `pkg/services/ai_assistant_storage*.go`: per-user transactions, receipts, epoch fencing and quota.
- `pkg/services/ai_assistant_conversations*.go`: history/search/export/versions/branch; `ai_assistant_runs*.go`: application-owned worker lifecycle.
- `pkg/services/ai_assistant_memories_projection.go`: pure public memory projection shared by listing/export; ownership queries and error policy remain with callers.
- `pkg/services/ai_assistant_context*.go`, `ai_assistant_compaction.go`, `ai_assistant_memories*.go`: bounded context and revocable memory.
- `pkg/services/ai_assistant_cleanup*.go`, `pkg/services/users_access.go`, `pkg/api/data_managements_clear.go`: online deletion, account fence and physical cleanup.
- `src/stores/assistant.ts`, `src/stores/assistant/*`: shared cloud conversation state; `src/lib/assistant_local_storage.ts`: isolated drafts and pending commands. Platform components do not own independent histories.
- `pkg/api/large_language_models_assistant_conversations*.go`: persistent API/runtime; generation availability does not hide history routes.
- `AI_ASSISTANT_EXPERIENCE_PLAN.md` and `docs/ai-assistant-experience-verification.md`: scope and actual evidence. Full database backups include all five new assistant tables on every user-data shard; retained embedding data remains untouched.

## Persistent Request and Lifecycle Trace

- `cmd/webserver_api_routes.go` registers authenticated assistant routes independently of generation availability. `messages/send.json` → `AssistantMessageSendHandler` → `AssistantConversationService.Send` → `AssistantStorageService.ExecuteCommand` persists question/run and receipt in a user transaction. Run state is stored in `AssistantMessage`, not a separate run table.
- `cmd/database.go` registers five assistant models on UserDataStore. `cmd/webserver_assistant_lifecycle.go` owns application startup/shutdown; API runtime composition injects shared storage/config into services and starts workers/maintenance. A disconnected SSE subscriber does not own generation.
- `AssistantContextService` prepares owned history and UID-bound tools; `pkg/llm/assistant_runner.go` orchestrates provider-neutral turns. `pkg/llm/provider/openai/openai_assistant_tools.go` and protocol helpers own upstream HTTP/wire formats and reuse `pkg/httpclient/`.
- Both `/assistant` routes load platform `AssistantPage.vue` → platform `AssistantConversation.vue` → common `AssistantCloudWorkspace.vue` → `useAssistantCloudWorkspace.ts`/assistant store → `src/lib/services/assistant_conversations.ts`. The client reuses Axios transport for JSON and a bounded stream reader for SSE.
- `src/stores/assistant/history.ts` owns list/search windows and request fencing; `src/lib/assistant_local_storage.ts` owns local draft/pending persistence. These are not separate platform histories or server data authorities.

## Current Shared Modules and Verification Entry Points

- `src/components/common/assistant/AssistantCloudWorkspace.vue` composes dialogs (`AssistantConfirmationDialog.vue`, `AssistantExportDialog.vue`, `AssistantReferenceDialog.vue`), toolbar/notices/recovery (`AssistantWorkspaceToolbar.vue`, `AssistantConversationNotices.vue`, `AssistantRecoveryPanel.vue`) and `assistant-workspace.css` in that directory. Dialogs own confirmation/export/reference interaction; platform adapters retain navigation and bill opening.
- `src/views/base/assistant/useAssistantWorkspaceScope.ts` fences asynchronous UI effects; `src/views/base/assistant/useAssistantWorkspacePanels.ts` owns responsive panel/focus behavior; `src/views/base/assistant/useAssistantWorkspaceLifecycle.ts` owns route synchronization and store attachment. Product commands remain in `src/views/base/assistant/useAssistantCloudWorkspace.ts`.
- `src/stores/assistant/session.ts` owns identity/capabilities/local recovery leases; `src/stores/assistant/conversation.ts` owns read windows and stream/poll lifetime; `src/stores/assistant/commands.ts` owns exact pending commands/receipt reconciliation; `src/stores/assistant/history.ts` owns list/search windows. The root store composes these modules.
- Authentication refresh authority belongs to `src/stores/token.ts` (`RefreshedSession` caller lease), with consumers in `src/lib/services/transport.ts` and `src/views/base/UnlockPageBase.ts`; it is not an assistant-only transport mechanism. Regression entry points include `src/stores/token.test.ts`, `src/views/base/UnlockPageBase.auth.test.ts` and `src/lib/assistant_local_identity.browser.test.ts`.
- `pkg/services/ai_assistant_storage_ledger.go` owns the per-operation ledger-clear context fence and committed-version reuse proof; `pkg/services/transactions_clear_context.go` couples ledger deletion and invalidation in the same user transaction. Regression families are `pkg/services/ai_assistant_ledger_clear*_test.go`; memory list/export contract coverage is `pkg/services/ai_assistant_memories_projection_test.go`.
- Component/lifecycle coverage starts at `src/components/common/assistant/AssistantCloudWorkspace.browser.test.ts` and `src/components/common/assistant/AssistantCloudWorkspace.lifecycle.test.ts`. Host browser review entry points are `scripts/assistant-experience-identity-review.mjs`, `scripts/assistant-experience-design-review.mjs` and `scripts/assistant-experience-accessibility-review.mjs`; shared identity flows live in `scripts/assistant-experience-e2e-identity.mjs`.
- E2E environment/cleanup regression entry points: `scripts/assistant-experience-e2e-environment.test.mjs`, `scripts/assistant-experience-e2e-cleanup.test.mjs`. These use Vitest in the Node environment and are included in `npm run test`. A script's presence is not execution evidence; final Docker/E2E acceptance must record its actual target and results.

## Final Validation Harness Boundaries

- `scripts/build-assistant-docker-validation.mjs` builds the isolated validation image; `scripts/assistant-experience-docker-runtime.mjs` adapts the E2E environment to task-owned containers; `scripts/assistant-experience-docker-smoke.mjs` and `scripts/assistant-experience-e2e.mjs` are smoke/business entry points. Validation image provenance is distinct from a stock Dockerfile release build.
- `scripts/assistant-bounded-command.mjs` owns POSIX command process groups, deadlines, output bounds and cancellation; `scripts/assistant-build-resources.mjs` owns private build source directories and registered builder/probe container cleanup. They are validation infrastructure, not application worker services. Matching Vitest regressions: `scripts/assistant-bounded-command.test.mjs`, `scripts/assistant-build-resources.test.mjs`.
- `scripts/assistant-experience-migration-check.py` orchestrates synthetic upgrade/backup/rollback verification; `scripts/assistant_migration_runtime.py` owns isolated application environment and bounded subprocess execution. Its separate Python unittest entry point is `scripts/assistant_migration_runtime_test.py`.
- `src/stores/assistant/compatibility.test.ts` covers capability initialization/synchronization races, upgrade recovery and unrelated history404. Capability error classification belongs to `src/stores/assistant/session.ts`, shared by both platforms, not route/UI adapters.
- Current execution/provenance/limitations are recorded in `docs/ai-assistant-experience-audit/15-docker-smoke.md`, `docs/ai-assistant-experience-audit/16-container-e2e.md`, `docs/ai-assistant-experience-audit/17-final-test-matrix.md` and `docs/ai-assistant-experience-audit/18-harness-finalization.md`. Consult these reports for actual results rather than inferring a pass from an entry point's existence.
