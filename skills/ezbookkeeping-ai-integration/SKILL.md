---
name: ezbookkeeping-ai-integration
description: "Delivery workflow for AI assistant and receipt-recognition features in ezbookkeeping. Use when changing OpenAI/provider config, prompts, embedding retrieval, AI API routes, or AI UI interactions."
---

# ezbookkeeping AI Integration Skill

## Quick Start

- Read `references/ai-feature-map.md` to locate config, provider, API, prompt, and frontend touchpoints.
- Confirm whether the change targets receipt recognition, assistant chat/summary, or both.
- Search for hard-coded AI settings first: `rg -n "openai|embedding|chat/completions|base_url|model_id" pkg src conf`.

## Design Constraints

- No hard-coded API keys, model IDs, or provider base URLs in business logic.
- Keep AI provider-specific transport inside provider or config layers.
- Keep assistant retrieval deterministic: stable knowledge text, stable hashing, explicit top-k selection.
- Keep data migration explicit: embedding cache is data, not disposable runtime state.
- Ensure desktop and mobile entry points remain behaviorally aligned.

## Execution Steps

1. Update AI config schema (`LLMConfig`) and config file template (`conf/ezbookkeeping.ini`).
2. Wire provider endpoints through config-based URL builders.
3. Keep API handlers focused on orchestration and validation.
4. If frontend needs runtime AI config info, expose it via `pkg/api/server_settings.go` and read via `src/lib/server_settings.ts`.
5. Update docs for migration and operations when storage/config semantics change.

## Verification

- `go test ./...`
- `go vet ./...`
- `npm run lint`
- `npm run test` (if frontend behavior changed)
- Manual smoke: AI assistant page loads and requests succeed with configured provider endpoint.

## References

- `references/ai-feature-map.md`
