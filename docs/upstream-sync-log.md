# Upstream Sync Log

This file records the upstream `QuantumNous/new-api` commit that has been reviewed or integrated into this repository.

## 2026-05-07 (follow-up)

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Upstream `main` reviewed through: `415d21d07100fbf767255cecde6e5c48752daa35`
- Local `main`: `4422a3e0b47b2518c23b03f0078644756501797c`
- Pending upstream-only commits:
  - `415d21d07100fbf767255cecde6e5c48752daa35` `♻️ refactor(layout): rename workspace switcher to system brand`
  - `abc255dd6dfb08fddefd531f984048b554bbd437` `☀ fix(default): keep SectionPageLayout description slot hidden`
  - `a7d019e3a9ff6e2f2661a38df00b35a2a58cfab5` `✨ feat(default): redesign dashboard overview`
- Notes:
  - `415d21d0` introduces `system-brand.tsx`, removes `workspace-switcher.tsx`, and churns related layout i18n keys.
  - `a7d019e3` is the larger default-dashboard refresh, touching overview panels, stat cards, filters, and system-settings routing.
  - These upstream changes are tracked here but not yet synced into OmniRouters.

## 2026-05-07

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `38a3314b9b7ba504773056c8408d5072da3a3aff`
- Synced through upstream `main`: `e8cfb546fa7e1d5bf266c5998181c0021826e045`
- Local integration branch: `codex/direct-main-upstream-sync`
- Local integration scope:
  - Integrated upstream `d98f0e8ac3d31bdbc22afd234e7428b4b6a91d45` through `e8cfb546fa7e1d5bf266c5998181c0021826e045`.
  - Integrated the Base UI Select items API migration across the default frontend.
  - Integrated model performance summary APIs and model square performance badges.
  - Preserved local OmniRouters model marketplace NEW, discount, and promotion badges while adding upstream performance badges.
- Notes:
  - Upstream commits were cherry-picked and conflict-resolved to preserve local OmniRouters customizations.

## 2026-05-06

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `dac55f0fdeb16bbbc2bdc472bda14e60431f3845`
- Synced through upstream `main`: `38a3314b9b7ba504773056c8408d5072da3a3aff`
- Local integration branch: `codex/direct-main-upstream-sync`
- Local integration scope:
  - Integrated upstream `8b2b03d2761f08195a35b9a71c82f96b2f3b2ea7` through `38a3314b9b7ba504773056c8408d5072da3a3aff`.
  - Integrated the Base UI/default frontend overhaul, model performance metrics, real rankings data, channel table server-side sorting, subscription payment display fixes, billing settings form improvements, `topup_link` API relocation, and OpenAI image edit reference field preservation.
  - Preserved local OmniRouters customizations for the classic auth shell, model marketplace NEW/discount badges, mail settings templates, marketing email sending, top-up success email notifications, and payment fee-rate display.
- Notes:
  - Upstream commits were cherry-picked and conflict-resolved to preserve local OmniRouters customizations.

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
