# Actual application route evidence

Executed on 2026-10-15 against the current worktree's freshly compiled Go binary, isolated SQLite database, real authentication/registration/account/category/transaction APIs, real Vue desktop and Framework7 mobile application routes. The upstream model is a deterministic protocol fixture: these results prove application persistence and interaction behavior, **not real-model financial reasoning or long-memory quality**.

Run from the repository root:

```sh
node scripts/assistant-experience-e2e.mjs
# After the final npm run build:
EBK_E2E_PRODUCTION=1 node scripts/assistant-experience-e2e.mjs
```

The helper creates a new temporary database, signing secret, dedicated users, ledger and loopback-only service/provider ports. No production data is used. Inherited `EBK_*` application overrides are removed from the child server environment so a developer’s database/provider settings cannot redirect the isolated fixture. It raises only the search/export request rate to 120/minute for the scripted traversal; storage/model budgets remain configured defaults. Every run rebuilds Go unless `EBK_E2E_BINARY` is explicitly supplied; that override is for debugging only and is labeled non-final in results. Process shutdown stops the backend/provider/Vite. Temporary fixture/config files remain local for diagnosis and are not included here.

`route-results-development.json` records **30 passing scenarios** (15 per platform), with 18 provider calls total. This run used Vite with actual application routes, no page/component mocks. Screenshots are at desktop 1440×960 and mobile 390×844, Chromium device-sized viewport (not physical-phone certification). They include empty, saved response/ledger evidence, memory, history/search, provider failure, a recovered branch awaiting explicit send, and IndexedDB refusal warning. The application log and ephemeral auth storage files are deliberately excluded.

The page component tests are separate supporting evidence: `AssistantPages.test.ts` (3 tests) exercises route lifecycle, generation-disabled history and IME with real Pinia/local storage and a protocol mock; `AssistantPages.browser.test.ts` (2 tests) uses actual Vuetify/Framework7 controls/navigation with a protocol mock. Neither is presented as API/DB E2E.

The route script found and reproduced a branch-receipt reload race: removing the pending command before persisting the recovered target/draft lost edited text. The final passing run includes the original network-fault/refresh reproduction against the atomic local recovery and route acknowledgement fix.

The development run preceded the final inclusive archived-list correction; its archive/unarchive result does not establish inclusive-list behavior. The production rerun explicitly requires an unarchived item to remain in the checked “Include archived” list.

Production service-worker cache and real bfcache observations are reported separately when the production script runs. A navigation only counts as bfcache coverage when the actual `pageshow.persisted` is true. No synthetic lifecycle event is accepted as proof of bfcache behavior. Wider dark/RTL/zoom/focus review lives in [the independent UI review](../ai-assistant-ui-review.md).

## Final production verification

The final production build (rerun after the R13 reference fix on 2026-10-16 Asia/Shanghai) passed `npm run test:production` and the real route suite: **33 PASS, 2 NOT RUN**, 18 provider calls. See `route-results-production.json`. The screenshots in this directory now come from this final production run. Both platforms passed inclusive archive/unarchive, exact reading offset restoration, lost-response send/branch reconciliation, confirmed assistant clearing with the original ledger expense retained, real service-worker control, `Cache-Control: no-store`, zero assistant responses in CacheStorage, no completed answer or authentication token in assistant IndexedDB, and offline API replay refusal.

Actual back/forward navigation was attempted on both platforms with Chromium’s default bfcache-disabling launch flag removed. Neither navigation reported `pageshow.persisted === true`, so both bfcache rows are **NOT RUN**; ordinary history restoration and the cross-tab identity boundary passed. This result does not certify a genuine bfcache restore.

The independent mixed-version smoke (`scripts/assistant-experience-compat-smoke.mjs`, requiring `EBK_E2E_BINARY` pointing to the archived old-main binary) connects this final production frontend to old main `1a21bc9bfd2b155a3ed9b8ac57caff7c76bbb293` in a separate fresh SQLite environment. Both platforms passed: actual capabilities HTTP 404, explicit server-upgrade explanation, no further navigation during the five-second observation window, only two capability probes, and zero provider calls. See `compat-results.json` and `*-old-backend.png`. This override is only compatibility evidence, separate from the freshly compiled final-backend acceptance above.

## Historical bill reference boundary (R13)

`EBK_E2E_PRODUCTION=1 node scripts/assistant-reference-e2e.mjs` passed all **10 scenarios** against the post-R13 production build and freshly compiled isolated Go/SQLite backend. Both desktop and mobile show an explicit changed-bill dialog, open the real current bill with its modified description (and amount CNY 234.56), and return with the unsent draft. Desktop additionally retains the dialog across a 16-second capability-refresh interval. Real deletion and an actual foreign-owner HTTP authorization rejection share the neutral inaccessible explanation; an injected transport abort has a distinct retry explanation. The foreign case rewrites only the preflight request ID to the second isolated user’s real transaction, without mocking its response.

Each reference preflight preserves the current reading state (bottom following or saved scroll offset) and draft without a model request. Explicit “Ask this question again” creates a second answer using the updated bill amount (`sourceAmountMinor=23456`), while the complete original answer snapshot—including `queriedAt`—remains byte-equivalent as a JSON object and the existing draft survives. The current transaction DTO exposes a positive seconds-scale `lastModifiedUnixTime`. See `reference-results-production.json` and `*-reference-*.png`. This targeted suite blocks service workers to make network fault injection deterministic; the separate full production suite verifies actual service-worker behavior.

The old-backend compatibility screenshots/results precede the R13-only reference-dialog addition; the full final-backend route suite and targeted R13 suite were rerun after it.
