# ezbookkeeping Project Structure Map

## Top-Level Layout

- `cmd/`: app wiring, startup, web server, database bootstrap.
- `pkg/`: backend application code.
- `src/`: Vue frontend code (desktop + mobile).
- `conf/`: config template (`ezbookkeeping.ini`).
- `templates/`: server-side templates, including AI prompt templates.
- `docs/`: migration and operation docs.

## Backend Modules (`pkg`)

- `pkg/api/`: HTTP handlers and request/response orchestration.
- `pkg/services/`: domain services and data access composition.
- `pkg/models/`: database models and API payload structs.
- `pkg/settings/`: config schema, defaults, and config-file/env loading.
- `pkg/llm/`: AI provider container and provider adapters.
- `pkg/templates/`: prompt/template loading.
- `pkg/storage/`: object storage adapters.

## Frontend Modules (`src`)

- `src/views/base/`: shared page logic for desktop/mobile pages.
- `src/views/desktop/`: desktop Vuetify pages.
- `src/views/mobile/`: mobile Framework7 pages.
- `src/lib/services.ts`: API client calls.
- `src/lib/server_settings.ts`: frontend accessors for server-provided feature/config flags.
- `src/router/desktop.ts`, `src/router/mobile.ts`: route entry points.

## AI Feature Touchpoints

- Receipt image recognition endpoint: `pkg/api/large_language_models.go`.
- AI assistant JSON/SSE endpoints: `pkg/api/large_language_models_assistant_harness.go`; mode/locale helpers: `large_language_models_assistant.go`.
- Assistant runner/protocol: `pkg/llm/assistant_runner.go`, `pkg/llm/data/*`, `pkg/llm/provider/openai/openai_assistant_tools.go`.
- UID-bound read-only tool queries: `pkg/services/ai_assistant_*.go`; fixed-connection read views: `pkg/datastore/assistant_read_view.go`.
- Shared assistant session/SSE/evidence: `src/views/base/assistant/AssistantPageBase.ts`, `src/lib/assistant_stream.ts`, `src/components/common/AssistantMessageEvidence.vue`.
- AI provider adapters: `pkg/llm/provider/*`.
- AI prompt templates: `templates/prompt/*.tmpl`.
- Legacy embedding schema/rows retained for rollback: `pkg/models/ai_assistant_embedding.go`, `cmd/database.go`; no cache service or active assistant embedding config/calls.
- Assistant config/operations: `docs/ai-providers.md`, `docs/ai-storage-migration.md`, `docs/ai-assistant-harness-verification.md`.

## Safe Change Pattern

1. Adjust config schema/defaults in `pkg/settings/setting.go`.
2. Wire config to provider/api layer in `pkg/llm` or `pkg/api`.
3. Expose required runtime flags in `pkg/api/server_settings.go` if frontend needs them.
4. Update frontend accessors (`src/lib/server_settings.ts`) and platform views.
5. Update `conf/ezbookkeeping.ini` and relevant docs in the same commit.
