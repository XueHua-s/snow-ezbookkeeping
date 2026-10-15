# AI Feature Map

## Configuration

- Global AI flags: `conf/ezbookkeeping.ini` section `[ai]`; legacy `[llm]` flags remain compatible, explicit `[ai]` values win.
- Receipt recognition provider: `[llm_image_recognition]`; optional recognition-only fallback: `[llm_image_recognition_fallback]`.
- Assistant tool providers: `[llm_assistant]` and optional `[llm_assistant_fallback]`. Primary section owns positive `run_timeout`, `query_timeout`, `max_rounds`; primary/fallback share budgets.
- Config parser/defaults: `pkg/settings/setting.go`; legacy browser receives run timeout through server settings; new capabilities expose validated `[assistant_experience]` limits.
- Retired `[llm_assistant_embedding]` and `openai_embedding_model_id`, including EBK/EBKCFP overrides, are ignored without secret-file reads.

## Backend Runtime

- Provider container: `pkg/llm/large_language_model_provider_container.go`; existing receipt/other-provider interfaces remain separate from optional tool capability.
- OpenAI tools/protocol adapters: `pkg/llm/provider/openai/*`; configured Chat Completions, forced streamed Chat Completions, or Responses serve both browser transports.
- Receipt recognition API: `pkg/api/large_language_models.go`.
- Assistant JSON/SSE orchestration: `pkg/api/large_language_models_assistant_harness.go`; locale/mode helpers: `large_language_models_assistant.go`.
- Single runner and budgets: `pkg/llm/assistant_runner.go`.
- Fixed UID seven-tool registry/filter/cursors: `pkg/services/ai_assistant_queries.go`, `ai_assistant_filter.go`, `ai_assistant_cursor.go`; transaction/statistics/balance implementations: `ai_assistant_transactions.go`, `ai_assistant_aggregate.go`, `ai_assistant_balances.go`.
- Fixed-connection read view/cancellation: `pkg/datastore/assistant_read_view.go`, `pkg/core/context_assistant.go`.
- No assistant vector service or embedding calls. Legacy model/schema: `pkg/models/ai_assistant_embedding.go`, registration in `cmd/database.go`, retained solely for rollback compatibility.

## Prompt Layer

- Assistant system prompt: `templates/prompt/personal_finance_assistant.tmpl`; summary calendar/context assembly in the harness API.
- Receipt image recognition prompt: `templates/prompt/receipt_image_recognition.tmpl`.

## Frontend Runtime

- API calls and timeout: `src/lib/services.ts`; bounded SSE validation: `src/lib/assistant_stream.ts`.
- Server settings accessors: `src/lib/server_settings.ts`.
- Persistent shared Store: `src/stores/assistant.ts` and `src/stores/assistant/*`; thin page adapter: `src/views/base/assistant/AssistantPageBase.ts`. Leaving a page unsubscribes; it never clears history or cancels a server run.
- Evidence presentation: `src/components/common/AssistantMessageEvidence.vue`.
- Shared rendering/scrolling: `src/components/common/assistant/AssistantMessageList.vue`, `src/views/base/assistant/useAssistantAutoScroll.ts`; typed session contract: `src/views/base/assistant/types.ts`.
- Platform feature containers, input and actions: `src/components/desktop/assistant/*`, `src/components/mobile/assistant/*`.
- Thin platform route pages: `src/views/desktop/assistant/AssistantPage.vue`, `src/views/mobile/assistant/AssistantPage.vue`.

## Migration-Sensitive Data and Verification

- Preserve the legacy `AIAssistantEmbedding` model table (default physical name `a_i_assistant_embedding`; resolve the actual engine mapper/prefix/schema) and rows in full backups; current assistant does not access or rebuild them.
- Transaction attachment metadata: `transaction_picture_info`; object prefixes: `transaction/`, `avatar/`.
- Migration/rollback runbook: `docs/ai-storage-migration.md`; provider/secret/timeout guide: `docs/ai-providers.md`.
- Isolated deployment receipt/tool tests: `pkg/api/ai_deployment_integration_test.go`; cache access denial and schema/record-retention regressions run locally.
- Actual gates and pending real-model/database/GUI checks: `docs/ai-assistant-harness-verification.md`.

## Persistent Experience

- Contract and five additive tables: `pkg/models/ai_assistant_*.go`; settings: `pkg/settings/setting_ai_experience.go` and `[assistant_experience]` in the INI.
- Shared transaction/CAS/receipts/quota: `pkg/services/ai_assistant_storage*.go`; history, versions, branch/search/export: `ai_assistant_conversations*.go`.
- Application-owned worker/lease/checkpoints: `ai_assistant_runs*.go`; context/compaction: `ai_assistant_context*.go`, `ai_assistant_compaction.go`; memory: `ai_assistant_memories*.go`.
- New HTTP/runtime lifecycle: `pkg/api/large_language_models_assistant_conversations*.go`; new client/SSE: `src/lib/services/assistant_conversations.ts`, `src/lib/assistant_conversation_stream.ts`.
- Local drafts/pending-command persistence: `src/lib/assistant_local_storage.ts`; cache exclusion: `src/lib/assistant_cache_policy.ts`, `src/sw.ts`.
- User deletion/disable: `pkg/services/users_access.go`; clear-all versus ledger-only clear: `pkg/api/data_managements_clear.go`; bounded physical cleanup: `pkg/services/ai_assistant_cleanup*.go`.
- Deletion keeps epoch and minimum body-free receipts; replay cannot delete newly created data. Background cleanup never accesses embedding records.
- Canonical scope: `AI_ASSISTANT_EXPERIENCE_PLAN.md`; current requirement-by-requirement evidence: `docs/ai-assistant-experience-verification.md`. Old harness reports cover only their stated scenarios.
