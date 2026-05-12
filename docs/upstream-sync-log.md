# Upstream Sync Log

This file records the upstream `QuantumNous/new-api` commit that has been reviewed or integrated into this repository.

## 2026-05-13

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `543cc64ea3805a3f2291b86525ad83771cb61423`
- Upstream `main` reviewed and synced through: `aa56667b8f233958aec0f593b3960c04230de6d3`
- Local integration scope:
  - Integrated default frontend regression fixes from `ba474393`, `2b89989f`, `fde2cac9`, and `469d3747`.
  - Integrated performance metrics handling/UI updates from `19fc384e` and `03d53732`.
  - Integrated user-group ratio display fix from `7fe896d2`.
  - Integrated classic dependency/resource refresh from `3856b9d2` and `428e3d91` where it applies to the current tree.
  - Integrated upstream request ID tracking and response-header protection from `aa56667b`.
- Skipped or equivalent:
  - `5fa103fa` was already equivalent for `THIRD-PARTY-LICENSES.md`; this sync still kept the new frontend `node_modules`/`dist` Docker ignores.
  - `3057f04a` had no file delta against this branch.
  - Merge commit `a720064d` was reviewed but not replayed as a merge object.
- Notes:
  - Kept the local absence of `README.en.md` and `web/classic/src/i18n/locales/zh.json`.
  - Preserved local OmniRouters customizations around model marketplace pricing, mail settings, external page header behavior, and top-up notification templates.
  - Quota notification top-up links now use the upstream `PaymentReturnURL` helper while retaining the local remaining-quota email values.

## 2026-05-10

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `a7475a1e67ff0d5a46c683f3c24430cf83d25f50`
- Upstream `main` synced through: `543cc64ea3805a3f2291b86525ad83771cb61423`
- Local integration commits:
  - `f2226194` `✨ feat: Add model performance metrics to dashboard`
  - `60a65e2b` `🎨 fix(theme): align UI controls with global radius tokens`
  - `944eca94` `feat: add DeepChat deeplink support (#4668)`
  - `6acdb666` `⚖️ chore(web/default): add reusable copyright header tooling`
  - `c0f3c18a` `feat(licenses): add LICENSE, NOTICE, and THIRD-PARTY-LICENSES files to Docker images`
- Notes:
  - The copyright-header commit was conflict-resolved in local OmniRouters-modified files so model marketplace, Sora pricing, and email settings customizations remain intact.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-05-07 (theme follow-up)

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Upstream `main` synced through: `a7475a1e67ff0d5a46c683f3c24430cf83d25f50`
- Previous upstream sync point: `415d21d07100fbf767255cecde6e5c48752daa35`
- Local integration scope:
  - Integrated the upstream theme-token and preset alignment across the default frontend.
  - Updated status badges, copy/group badges, sidebar, and AI element UI helpers to use the new theme palette behavior.
  - Synced dashboard chart and color utilities so overview/model/user charts follow the new preset-aware colors.
- Notes:
  - This upstream commit is UI-only and does not interfere with the local Sora pricing or model marketplace changes currently in progress.

## 2026-05-07 (follow-up)

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Upstream `main` synced through: `415d21d07100fbf767255cecde6e5c48752daa35`
- Local integration commits:
  - `ce108dc6e756598c121d40dc527d06e9314e52f8` `♻️ refactor(layout): rename workspace switcher to system brand`
  - `c4a9a313277517e7495c750525200abb545c31b3` `✨ feat(default): redesign dashboard overview`
- Notes:
  - `abc255dd6dfb08fddefd531f984048b554bbd437` was empty on this branch because the equivalent `SectionPageLayout` description-slot behavior was already aligned.
  - `415d21d0` introduces `system-brand.tsx`, removes `workspace-switcher.tsx`, and churns related layout i18n keys.
  - `a7d019e3` is the larger default-dashboard refresh, touching overview panels, stat cards, filters, and system-settings routing.
  - These upstream changes are now synced into OmniRouters main.

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
