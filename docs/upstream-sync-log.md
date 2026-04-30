# Upstream Sync Log

This file records the upstream `QuantumNous/new-api` commit that has been reviewed or integrated into this repository.

## 2026-04-30

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- V1 migration branch: `codex/migrate-upstream-v1`
- Synced through upstream `main`: `dac55f0fdeb16bbbc2bdc472bda14e60431f3845`
- Local integration scope:
  - Integrated the v1.0 frontend migration line and subsequent default/classic frontend fixes.
  - Integrated Vertex custom `base_url` gateway-prefix handling.
  - Integrated tiered billing display fixes, including UTF-8 Base64 decoding and normalized tier-label matching.
  - Integrated the classic frontend switch back to the default frontend.
- Notes:
  - Upstream commits were cherry-picked and conflict-resolved to preserve local OmniRouters customizations.
  - `ghcr.io/1412212638/omnirouters:v1-ui-test` is used as the test image tag for this branch.

## 2026-04-28

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Upstream `main` reviewed through: `df14a0bf18917feaa667c250aa24704fd0c7c932`
- Local integration scope:
  - Integrated upstream `02aacb38a2523411df93077cd52c2ddccdb86560` user `created_at` / `last_login_at` tracking.
  - Integrated upstream `4e93148d9ee80ab27a334f262908a71b9f90912d` config map replacement fix.
  - Integrated upstream `bee339d279ccecbf8c8a89e14ddbbd902f78bd5d` tiered billing ratio/price fallback serialization fix.
  - Integrated upstream `9f8a4ec05010da20704c1b55aa8b9af5630df72e` tiered billing cache-price display fix.
- Deferred:
  - Upstream `a42b397607780bb0f2dbc851ae0c01043244691d` v1.0 frontend migration (`web/default` + `web/classic`) remains under evaluation.
  - Upstream CI/README-only commits were reviewed but not integrated because they do not affect runtime behavior here.

## 2026-04-27

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream baseline: `02aacb38`
- Synced through upstream `main`: `e36d191c2e31d31d0a7b8da51da1dcb93f63f681`
- Local integration commit: `477717f71cba8221f0ab8c212fe1e024d852dc69`
- Local branch pushed: `main`
- Notes:
  - Integrated the relevant upstream changes after `02aacb38` through `e36d191c2e31d31d0a7b8da51da1dcb93f63f681`.
  - The upstream commits were adapted into this repository as a local integration commit, not merged as upstream commit objects.
