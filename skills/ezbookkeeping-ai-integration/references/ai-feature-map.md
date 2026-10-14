# AI Feature Map

## Configuration

- Global AI flags: `conf/ezbookkeeping.ini` section `[ai]`; legacy `[llm]` flags remain compatible, explicit `[ai]` values win.
- Receipt recognition provider: `[llm_image_recognition]`; optional recognition-only fallback: `[llm_image_recognition_fallback]`.
- Assistant tool providers: `[llm_assistant]` and optional `[llm_assistant_fallback]`. Primary section owns positive `run_timeout`, `query_timeout`, `max_rounds`; primary/fallback share budgets.
- Config parser/defaults: `pkg/settings/setting.go`; browser receives only assistant run timeout through server settings.
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
- Shared page-local session/cancellation/state/continuation: `src/views/base/assistant/AssistantPageBase.ts`.
- Evidence presentation: `src/components/common/AssistantMessageEvidence.vue`.
- Platform pages: `src/views/desktop/assistant/AssistantPage.vue`, `src/views/mobile/assistant/AssistantPage.vue`.

## Migration-Sensitive Data and Verification

- Preserve the legacy `AIAssistantEmbedding` model table (default physical name `a_i_assistant_embedding`; resolve the actual engine mapper/prefix/schema) and rows in full backups; current assistant does not access or rebuild them.
- Transaction attachment metadata: `transaction_picture_info`; object prefixes: `transaction/`, `avatar/`.
- Migration/rollback runbook: `docs/ai-storage-migration.md`; provider/secret/timeout guide: `docs/ai-providers.md`.
- Isolated deployment receipt/tool tests: `pkg/api/ai_deployment_integration_test.go`; cache access denial and schema/record-retention regressions run locally.
- Actual gates and pending real-model/database/GUI checks: `docs/ai-assistant-harness-verification.md`.
