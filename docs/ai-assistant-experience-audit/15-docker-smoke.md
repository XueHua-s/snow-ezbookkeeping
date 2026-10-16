# 15 · Current application Docker smoke

2026-10-16. Serial Docker stage; no delegation, commit/push, production data,
private provider configuration, or unrelated container changes. Implementation
uses project-spec/global-prompt and AI integration/feature-map; focused helper
tests use the Vitest skill. No Vue/product source was changed by this stage.

## Reusable environment

`EBK_E2E_DOCKER_IMAGE` accepts the immutable image ID as well as a tag and opts the existing `environment({production:true})` into
`scripts/assistant-experience-docker-runtime.mjs`. Default host behavior and
`url/base/api/user/setGeneration/calls/close/artifacts/directory` remain intact.
Docker mode requires production routes; it serves the image's built `public`
directory and never starts Vite or a host Go backend. `setGeneration` actually
stops, copies the updated synthetic config, and starts **the same container**;
SQLite and signing secret survive the restart. No host config override is
forwarded. Config is copied as readable synthetic data into the container,
while the host temporary directory remains private and is removed on close.

Each run creates a random task-labelled container, publishes only the app on
127.0.0.1, and mounts no host directory. The existing deterministic Node
provider listens on host 127.0.0.1 and is reached via Colima's
`host.lima.internal`; a real container reachability probe succeeded. This is a
Colima-specific test runtime path. The default bridge retains outbound access;
there is no claim of total network isolation. No commercial provider is used.

Cleanup collects raw application logs under the chosen artifact path, then
actually removes the container (not merely a docker-logs process). It also
closes the fixture/provider and removes temporary synthetic credentials.
CLI calls have 30-second bounds; graceful app stop has a 3-second limit. A
synchronous bounded exit fallback targets only that random container. Partial
startup is included in cleanup. No network resource is created, and no shared
cache/image/container is pruned. `codex-s03-build` remains untouched.

## Build provenance and constraints

`scripts/build-assistant-docker-validation.mjs` executes a fresh host
`npm run build`, copies an allowlist of backend/module/template/entrypoint,
sample INI/LICENSE and generated dist files into a new shared cache directory,
compares source/artifact hashes, and compiles the current backend inside
`golang:1.27.1-alpine3.24` with CGO. It mounts only that task source tree and the
read-only Go module cache; it never sends the private workspace root as a
Docker build context. The temporary source copy is removed afterwards.

The official Alpine APK CDN returned TLS errors; HTTPS mirrors.aliyun.com
worked with normal package signature validation. Docker Hub alpine:3.24.2
manifest fetch repeatedly returned EOF; daocloud failed with 502/EOF. These
failed attempts are not application test failures. The validation runtime
therefore uses the cached Go base, whose actual `/etc/alpine-release` is
3.24.2, with extra Go toolchain files. It uses the exact project entrypoint and
UID:GID 1000:1000. **This is not the stock runtime image or the stock Node
26 frontend stage.** Host frontend tools are probed and recorded rather than
inferred from Dockerfile declarations. No product Dockerfile was changed.

Build/source/frontend/binary hashes, actual image/base IDs, OS/architecture,
toolchain output, commands and successful exit codes are in
`../ai-assistant-experience-evidence/audit-docker/build-provenance.json`.

## Reproduction

From the repository root, with current dependencies, Colima and the cached
Go base available:

```sh
node scripts/build-assistant-docker-validation.mjs
EBK_E2E_DOCKER_IMAGE=ebk-assistant-validation:20261016 \
EBK_E2E_ARTIFACTS=.vitest/audit-docker \
node scripts/assistant-experience-docker-smoke.mjs
```

Subsequent full E2E uses the **same adapter**, automatically creating its own
fresh isolated container and closing it on completion:

```sh
EBK_E2E_DOCKER_IMAGE=ebk-assistant-validation:20261016 \
EBK_E2E_PRODUCTION=1 EBK_E2E_ARTIFACTS=.vitest/final-docker-e2e \
node scripts/assistant-experience-e2e.mjs
```

Rebuild the validation image after any product source/template/dist change.
Keep raw app logs under `.vitest`; commit only reviewed redacted JSON and
hashes. The image contains no runtime config/test credentials; it may be kept
for the next serial stage. Do not reuse tokens or real-user data.

## Execution evidence

Final smoke exited **0** with **10 PASS scenarios**. Evidence directory:
`docs/ai-assistant-experience-evidence/audit-docker/` contains only redacted
JSON; raw application/build/helper logs remain under `.vitest` and their
SHA-256 values are recorded in `verification-summary.json`.

- Fresh database initialization, real project entrypoint, actual runtime UID/GID
  1000:1000, server settings, synthetic registration/login and ledger creation.
- Anonymous request returns the existing exact 400 / errorCode 202012 contract
  (`token is empty`); an initial script assumption of 401 was corrected after
  reading `errs/token.go`, then the whole smoke reran. This was not a product bug.
- Persistent send completed through the real service tool registry. Assertions
  inspect `queries` (the public DTO field): `query_transactions`, UTC timezone,
  expense-only filter, excludesDeleted, coverage and queriedAt; the reference
  matches the created transaction and exact sourceAmountMinor `12345`.
  Persisted selected answer reads back as completed. Fixture prose is not
  evidence of real-model numerical reasoning.
- Provider 503 reaches persistent failed; cancellation occurs only after the
  fixture receives the slow request and reaches persistent canceled. **Upstream
  TCP closure was not observed and is not claimed.**
- Same-container restart with generation disabled exposes unavailable
  capabilities, preserves authenticated history, rejects new sends specifically
  with `provider_unavailable`, and does not increment fixture calls. Restart
  enabled preserves the same token/history/ledger and new generation succeeds.
- `npm run test:production` executes against this container: desktop and mobile
  production login pages render with no runtime/network errors.
- After stopping the container, a copied **synthetic** SQLite file is inspected:
  all five assistant tables plus `a_i_assistant_embedding` exist. This proves
  fresh additive schema initialization, not upgrade retention of populated legacy
  records (the latter remains the database matrix's responsibility).

Image ID:
`sha256:990c1f23742049c78c43832cdde54e69a575412675524f4a8ac88d4b80bc0dff`.
Linux executable SHA-256:
`e5af7ae3ab00bf3930964e0f7c02aacd5b12dca41910effa15ecb967c56c5f63`.
A fresh `docker run --rm` probe independently confirmed executable hash,
Alpine 3.24.2, UID/GID 1000 and Go 1.27.1 linux/arm64. Full source/dist hashes and
module/build details are in `build-provenance.json`.

Focused helpers: **3 files / 10 PASS**. A new regression covers uncertain
Docker creation followed by daemon-unavailable inspect: cleanup reports the
error and retains exit recovery; only explicit `No such object/container`
permits concluding absence. Temporarily restoring the old behavior made the
regression fail (exit1); final behavior restored and all 10 passed again.
Parent independently reran these 10 checks successfully (`/tmp/ebk-parent-docker-helpers.log`).
A separate real Docker partial-startup probe created a busybox container,
failed config copy into its absent app directory, and confirmed container
removal and no newly listening provider/reservation handles. This was not a
mock Docker result. Five script syntax checks, `git diff --check`, and the
source-size gate (1709 files, zero over 500 lines) passed.

All task app/build/probe containers and temporary credentials are gone; no
network was created, the reachability probe's host port 39876 is closed, and
`codex-s03-build` is still running. Only the validation image remains for the
next E2E stage. Scripts/report frozen at handoff; no product change requiring
a new image or renewed UI review was made by this stage.

## Explicitly not covered

Full desktop/mobile business E2E and complete unit/Go/race/three-database
matrix belong to following serial stages. Stock Dockerfile end-to-end build
and its Node26 frontend stage were NOT RUN; exact Alpine runtime image was
unavailable due to registry failures. amd64, commercial main/fallback model
quality, physical mobile keyboard, manual screen reader, and actual bfcache
restoration have no positive evidence from this stage.
