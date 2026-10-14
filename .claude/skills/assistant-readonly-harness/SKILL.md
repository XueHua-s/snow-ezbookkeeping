---
name: assistant-readonly-harness
description: Plan and verify a request-scoped, read-only LLM tool harness over authenticated application data, with precise aggregation, bounded protocols, cancellation and evidence.
---

# Read-only assistant harness

Use when designing or reviewing an authenticated assistant that answers from application records. This skill is a reusable design and verification checklist, not implementation authorization. Follow the current user scope; a planning request must not execute the planned feature or promote draft tasks.

## Workflow

1. Read the product contract and current API/service/provider boundaries. Record which statements are requirements, source-confirmed facts, and unverified external assumptions.
2. Bind identity in a request-scoped service, never in model-controlled arguments. Use an explicit read-only registry separate from any general MCP registry. Validate ownership for primary and related records, cursors, lists and aggregates.
3. Define a provider-neutral tool protocol and strict typed arguments. Generate schemas from one definition; separately reject duplicate JSON keys, unknown keys, invalid nulls, integer overflow and conflicting filters. Keep provider strict-schema adaptations at the transport boundary.
4. Define the data truth before prompting: exact signed amounts, currency separation, canonical logical record identity, time intervals, hidden/deleted policy, totals coverage and zero-data behavior. Test from a hand-calculated fixture.
5. Separate complete aggregation from paginated detail. Bound output and runtime, never silently bound the matching dataset. Include normalized scope, read time and completeness in deterministic service output.
6. Validate cancellation from the HTTP request through context wrappers, HTTP clients and SQL drivers. Use a consistent short read transaction on one connection for each multi-query tool; do not retain it while waiting for a model. Inspect the installed ORM API instead of assuming isolation options exist.
7. Execute only complete, successfully terminated model turns. Count retries and fallback against budgets. Keep tool-call/result pairs intact, retain protocol-required state only in the request and rebuild neutral context on provider changes. Fail explicitly when no lossless conversion exists.
8. Enforce one terminal state. Serialize SSE writes; cancel on write errors; never translate malformed data or EOF into success. Preserve incomplete answers as incomplete and exclude them from trusted follow-up facts.
9. Treat record text and history as untrusted data. Generate references from queried records and distinguish retrieved records from proof of every sentence. Avoid logging prompts, balances, opaque provider state, secrets or raw cursors.
10. Document migration and rollback; preserve old data/schema until explicit cleanup is authorized. Separate code completion, environment validation and release readiness.

## Verification patterns

- Deterministic financial/record truth, cross-user attacks, mixed authorized/unauthorized IDs and hidden/deleted records.
- Static pagination without gaps/duplicates; signed cursor tampering, expiry, changed identity and changed parent expansion.
- Mock provider fragmentation, truncated turns, duplicate IDs, invalid arguments, retries, fallback and cancellation.
- Real-driver cancellation and concurrent updates inside each supported database's read-view contract.
- Frontend EOF without terminal, invalid JSON/UTF-8, CRLF/multiline frames, repeated terminal, stale callbacks and navigation cancellation.
- Data-volume tests that record peak memory, scanned rows, elapsed time and output size without leaking private content.

## ezbookkeeping resources

- [Research and evidence](../../../.specs/analysis/research-ai-assistant-harness.md)
- [Active task](../../../.specs/tasks/in-progress/ai-assistant-harness.feature.md)
- [Source contract](../../../AI_ASSISTANT_HARNESS_PLAN.md), especially §§15–17
- Inspect `pkg/datastore/query_context.go` and installed xorm `session_tx.go`: the current adapter drops cancellation and xorm Begin uses nil TxOptions.
- Reuse `github.com/invopop/jsonschema`, Go `math/big`, existing provider adapters and Vue shared composables; do not add an agent framework to solve a single tool loop.
- Old AI integration skill references to deterministic top-k are superseded for this replacement task; update both project copies during implementation migration work, not during plan generation.

External links and verification limits are recorded in the research artifact. Do not present unvisited documentation or unrun database/provider checks as verified.
