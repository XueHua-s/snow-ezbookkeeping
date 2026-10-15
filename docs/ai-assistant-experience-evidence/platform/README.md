# Platform command integration evidence

The sanitized `platform-commands.json` records the successful isolated run on 2026-10-15.

Command:

```sh
EBK_E2E_ARTIFACTS=.vitest/assistant-platform node scripts/assistant-platform-command-smoke.mjs
```

This runs a real Go HTTP server, a disposable SQLite database, authenticated synthetic users and 12 synthetic transactions. The configured model service is a deterministic protocol fixture, including structured context summaries and authorized continuation handles. Requested page size is 8; the application may reduce it to fit the actual model request budget.

All 11 recorded command groups passed. In particular, a history-query branch recovers its source after a fresh read, its next page contains references disjoint from the first page, and an explicit restart uses the authorized saved query scope. Editing branches create a visible empty target before their separately authorized first generation. Memory changes, version selection, viewed-version export, rename/archive, delete and clear use the real API.

These are **HTTP command integration checks**, corresponding to commands used by the shared frontend hook. They are **not evidence that every command was clicked through on both platforms**. Separate actual-route browser and UI-review reports cover platform navigation, rendered controls, clipboard, focus and bill-detail return. The shared-hook unit tests cover edit cancellation/failure preservation, explicit confirmation and draft protection. This fixture does not establish real-model reasoning quality, physical mobile keyboard behavior, database portability or long-history performance.

The JSON contains only timestamps, assertion labels and fixture configuration. It contains no credentials, user IDs, financial message bodies, raw cursors or provider prompts.
