---
name: ezbookkeeping-ai-integration
description: "Delivery workflow for AI assistant and receipt-recognition features in ezbookkeeping. Use when changing OpenAI/provider config, prompts, read-only assistant tools, AI API routes, or AI UI interactions."
---

# ezbookkeeping AI Integration Skill

## Quick Start

- Read `references/ai-feature-map.md` to locate config, provider, API, prompt, and frontend touchpoints.
- Confirm whether the change targets receipt recognition, assistant chat/summary, or both.
- Search for hard-coded AI settings first: `rg -n "openai|embedding|chat/completions|base_url|model_id" pkg src conf`.

## Design Constraints

- No hard-coded API keys, model IDs, or provider base URLs in business logic.
- Keep AI provider-specific transport inside provider or config layers.
- Keep assistant queries UID-bound and read-only through the fixed seven-tool registry; query and aggregate complete matching ledger ranges without history/window/vector selection limits.
- Preserve exact signed amounts, per-currency totals, logical transfers and actual query evidence; distinguish empty results from query/provider failures.
- Keep one runner for JSON/SSE with cancellation, bounded execution, explicit terminal states and verified continuation.
- Preserve the legacy embedding model/table/records for rollback; the assistant must never read/write that cache or resolve retired embedding configuration/secret-file overrides.
- Ensure desktop and mobile entry points remain behaviorally aligned.

## Execution Steps

1. Update active provider config (`LLMConfig`), assistant runtime budgets, and config template (`conf/ezbookkeeping.ini`); retired embedding keys remain ignored.
2. Wire provider endpoints through config-based URL builders.
3. Keep API handlers focused on orchestration and validation.
4. If frontend needs runtime AI config info, expose it via `pkg/api/server_settings.go` and read via `src/lib/server_settings.ts`.
5. Update docs for migration and operations when storage/config semantics change.

## Verification

- `go test ./...`
- `go vet ./...`
- `npm run lint`
- `npm run test` (if frontend behavior changed)
- Run isolated configuration/cache/schema-retention regressions; never use production data or count skipped integrations as passing.
- Manual smoke on both platforms: tools produce real evidence, errors retain explicit states, cancellation stops upstream work and next-page requires an explicit action.
- Real-model smoke is opt-in with a dedicated isolated deployment/test user; see `docs/ai-providers.md` and `docs/ai-assistant-harness-verification.md`.

## References

- `references/ai-feature-map.md`
