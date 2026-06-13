# Upstream Sync Log

This file records the upstream `QuantumNous/new-api` commit that has been reviewed or integrated into this repository.

## 2026-06-13

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `4ca47ee236fd5f09ec71d732fbe3d4270f03bec3`
- Upstream `main` reviewed through: `51475c8062e53d4f731ce9e5f7e6c8ea9ce77cb9`
- Local integration scope:
  - Integrated upstream OpenAI Images API streaming relay and image edit support from `d2576ddcd31ff752c30b54d1781e802e4021f824` plus the follow-up relay split/governance alignment from `59a93cf5c7bb4f7e428e36bfaa2458b474c281a3`.
  - Integrated channel search debounce and IME-composition handling from `30d3a3a5f7036a005cf2938a66eed5e7606bf76a`.
  - Integrated Kimi K2.6 temperature normalization from `867d8acfc3cb001cc560869bd52b9c1c21cff99a`.
  - Integrated upstream DataTable rendering/pinned-column refactor from `6f415428d3e6ac7d136f44808fc76e0f3aebab1e` and the layout/badge display follow-up from `27b2b2c4b95536fb832b9e9059ab00ec79d7e5d9`.
  - Integrated deleted-user status filtering from `1507229238f08836673c8b4a7138b1d21634b155`, adapted to the current pre-DataTable-refactor table implementation.
  - Integrated deployment-settings lazy fetch from `51475c8062e53d4f731ce9e5f7e6c8ea9ce77cb9`, keeping the local models page layout while moving deployment settings requests into the deployments-only section.
  - Integrated localized security audit logs from `d0c4305a16d168e4e3863cae36ba17b91974af14`, preserving local pricing refresh behavior and adding structured operation/login audit display.
- Deferred:
  - Deferred upstream `1292b8b2d5e90480521d45af05ff0e8b38a199f6` Codex channel update because it removes Codex OAuth surfaces that require product confirmation here.
- Notes:
  - Preserved OmniRouters Codex OAuth surfaces, model marketplace New/discount/promotion badges, Sora/audio pricing, wallet/payment display, mail settings, tickets, profile display, deleted-user filtering, deployment-settings lazy fetch, and OAuth icon URL behavior while adopting the upstream DataTable structure.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-06-07

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `32805849d683fb3347650173932d4f1f4ac613b4`
- Upstream `main` reviewed through: `4ca47ee236fd5f09ec71d732fbe3d4270f03bec3`
- Local integration scope:
  - Integrated mobile usage-log cost badge alignment while preserving the existing local row-data guard.
  - Integrated multiselect combobox theme-aware popover styling and six-decimal model-pricing editor steps.
  - Integrated public pricing/ranking hero copy simplification.
  - Integrated the shared default-frontend dialog wrapper component and the broad dialog sizing/footer migration across the default frontend.
  - Integrated the model-pricing mode tab placement improvement.
  - Manually integrated the model-pricing visual editor draft-save fix so page-level "Save model prices" first commits the open editor draft.
  - Integrated the full visual model-pricing module split, model-pricing snapshot/table-column extraction, and reusable JSON code editor migration.
  - Preserved OmniRouters Sora per-request pricing and fixed per-request `audio_generation` surcharge editing across visual editing, JSON editing, draft detection, deletion, and batch-copy flows.
- Deferred or skipped:
  - Skipped upstream `b5331936` profile user-id badge because OmniRouters already shows username, ID, user group, and role in a richer profile metadata line.
- Notes:
  - Preserved OmniRouters wallet/payment display, mail settings, ticket routes, profile display, OAuth icon URL support, subscription purchase behavior, fetch-model redirect-source filtering, and local model marketplace behavior.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-06-05

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `7aaa5332657e00fe801a2a7dd8b421e4ce4c842c`
- Upstream `main` synced through: `32805849d683fb3347650173932d4f1f4ac613b4`
- Local integration scope:
  - Integrated issue-template refinements, relay idle connection timeout configuration, and stream scanner buffer reuse in channel handlers.
  - Integrated Dify remote-image file pointer initialization, GLM Anthropic-compatible non-chunked relay behavior, and narrower OpenAI o-series adaptation.
  - Integrated video task lookup model resolution for `GET /v1/video/generations/:task_id`.
  - Integrated anonymous request body limiting for unauthenticated write callbacks and auth flows.
  - Integrated configurable channel-affinity retention when a previously affined channel becomes disabled or unusable.
  - Integrated classic quota amount display using configured exchange-rate conversion.
  - Reviewed `b0ac0429` usage-log mobile-card TypeScript fix; the same `created_at` row-data guard was already present locally.
- Notes:
  - Corrected the previous `7aaa5332` full-hash record from an invalid hash to `7aaa5332657e00fe801a2a7dd8b421e4ce4c842c`.
  - Preserved OmniRouters wallet corporate-transfer display, payment fee/currency handling, mail settings, ticket routes, profile display, OAuth icon support, Sora/audio pricing, and local model marketplace display behavior.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-06-02

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `0c7aceb8319b2fdc67e7eb65773c30ac87174d3f`
- Upstream `main` synced through: `7aaa5332657e00fe801a2a7dd8b421e4ce4c842c`
- Local integration scope:
  - Integrated model-marketplace custom model icon display while preserving OmniRouters model card badges, pricing labels, Sora/audio pricing, and status display customizations.
  - Integrated frontend workspace dependency centralization with a shared `web/package.json` and root `web/bun.lock`.
  - Integrated classic frontend Rsbuild support and updated Docker/release/make build paths to install from the shared `web` workspace.
  - Integrated classic React 19 Semi adapter injection.
  - Integrated channel drawer advanced-settings error detection, while keeping model-mapping errors from auto-expanding the advanced section.
- Notes:
  - Preserved OmniRouters classic login/register `AuthShell` visual layout instead of reverting those pages to the upstream plain wrapper.
  - Preserved OmniRouters wallet corporate-transfer display, payment fee/currency handling, mail settings, OAuth icon URL support, ticket navigation, task-log model column, profile display tweaks, and local model marketplace display behavior.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-05-31

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `5b86ce0d7001df96bb0f7d32446926c030addf6f`
- Upstream `main` synced through: `0c7aceb8319b2fdc67e7eb65773c30ac87174d3f`
- Local integration scope:
  - Integrated Claude Opus 4.8 support across Claude, Bedrock, Vertex, default model ratios, and cache ratios.
  - Integrated admin log ordering/index fixes for `created_at` composite index usage.
  - Integrated subscription plan balance-redemption toggle while preserving the local decision to exclude paid-feature compliance locks.
  - Integrated default frontend form-validation focus behavior and API key custom `cc-switch` name persistence fix.
  - Integrated classic frontend Axios patch update and Playwright MCP artifact ignore rule.
- Notes:
  - Preserved OmniRouters wallet corporate-transfer display, payment fee/currency handling, Sora/audio pricing, mail settings, OAuth icon URL support, login-page layout/header behavior, ticket feedback routing, task-log model column, model marketplace labels, and profile display tweaks.
  - Skipped upstream merge commit `e8c836d7` because its non-merge contents were covered by `63ead2bf` and `e79cee1e`.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-05-27

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `5bc4c748139029a8856191d5820f7aff90c5ae3d`
- Upstream `main` synced through: `5b86ce0d7001df96bb0f7d32446926c030addf6f`
- Local integration scope:
  - Integrated system-settings save detection fixes for dotted option keys and safer numeric inputs.
  - Integrated the `simple-large` theme preset, `xl` font scale, and related usage-log/channel badge cleanup.
  - Integrated channel and usage-log UI improvements, including the responsive mobile log card.
  - Integrated exact log filtering behavior for usage/token log filters unless `%` is explicitly supplied.
  - Integrated batch update optimization that combines user quota, used quota, and request-count writes.
- Notes:
  - Preserved OmniRouters corporate-transfer payment settings, OAuth icon URL support, Telegram login icon display, task-log model column, Sora/audio pricing, mail settings, wallet fee/currency display, and model marketplace status bars.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-05-26

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `583da45296eda8a9950a346055997e53cb8a7e1e`
- Upstream `main` reviewed through: `5bc4c748139029a8856191d5820f7aff90c5ae3d`
- Local integration scope:
  - Integrated relay/backend fixes for image quality handling, oversized upstream error log truncation, Claude/Gemini tool-call conversion, and channel test user IDs.
  - Integrated default-frontend fixes for API-key search pagination, duplicate channel toast handling, badge sizing/colors, usage-log spacing/typography, chart dark-mode readability, and default-theme font fallback.
  - Integrated Anthropic theme preset and font customization while keeping the default preset on sans.
  - Integrated Waffo payment settings save-flow consolidation while preserving OmniRouters corporate-transfer settings and excluding paid-feature compliance confirmation.
  - Integrated subscription purchases with wallet balance, without adding the upstream compliance-gate check.
  - Integrated the modular channel create/edit drawer refactor, model-mapping editor improvements, and Base UI multi-select follow-up.
- Deferred:
  - `51ca897cf4136bb2ecff7b69ef66bc6a778ff661` home hero redesign remains intentionally deferred.
- Notes:
  - Preserved OmniRouters wallet corporate-transfer display, payment fee/currency handling, Sora/audio pricing, mail settings, OAuth icon URL support, login-page layout/header behavior, ticket navigation, task-log model column, and model marketplace labels.
  - The `33608826` hero-file formatting conflict was resolved by keeping the local hero because the upstream home hero redesign is still deferred.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-05-25

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `b9bc6f0e21fb3b87267475912fc7c31fa0494d6e`
- Upstream `main` synced through: `49bc3a1175fd28726bc62596cb321b546ed02e95`
- Local integration scope:
  - Integrated upstream relay memory-residency reduction for large base64 request bodies, including byte-based param override processing and disk-backed outbound JSON bodies.
  - Integrated Gemini response content assembly optimizations for large inline media payloads.
  - Integrated multi-key channel status cache handling so auto-disabled channels are evicted/re-enabled correctly when individual keys change.
  - Integrated Waffo Pancake webhook trade number handling using `OrderMerchantExternalID`, including session creation, webhook dispatch, subscription handling, and tests.
  - Integrated the upstream classic-frontend hiding of Waffo Pancake payment settings.
- Notes:
  - Preserved the local decision to exclude paid-feature compliance confirmation and avoided adding payment/reward locks.
  - Preserved OmniRouters wallet, corporate-transfer, Sora/audio pricing, mail settings, OAuth icon URL, login header, ticket workspace, and model marketplace customizations.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

### Partial follow-up from upstream `b302be30e3fd1e548d47a6add7ea2eedfd3ca6c0`

- Integrated low-risk fixes from the upstream v1 feedback batch:
  - Channel copy now inserts the cloned channel directly so the response includes the new ID.
  - `/api/status` exposes `password_login_enabled`, and the new login form respects it without changing the login page layout.
  - User search supports server-side `role` and `status` filters, with matching default-frontend query parameters.
- The rest of `b302be30` remains selectively deferred where it introduces unrelated frontend-cache behavior or paid-feature compliance locks.

### Large frontend refactor follow-up through upstream `583da45296eda8a9950a346055997e53cb8a7e1e`

- Integrated upstream frontend restructuring commits:
  - `92a0959448751d858c42f749b46be6a69820998b` drill-in sidebar/navigation refactor.
  - `b08febaa3c41b7ea1da5118744f661073a585ba5` compact system-settings layout refactor.
  - `88437a18691ad13687f1aad073b4e67b93ab90c0` default frontend dependency refresh.
  - `583da45296eda8a9950a346055997e53cb8a7e1e` usage-log filter responsiveness/mobile UX refactor and build metadata helper.
- Notes:
  - Preserved OmniRouters admin Email Settings as a standalone admin entry rather than moving SMTP back under System Settings.
  - Preserved wallet corporate transfer display, payment fee/currency handling, Sora/audio pricing, OAuth icon URL support, login-page custom layout/header behavior, ticket navigation, and model marketplace labels.
  - Excluded upstream paid-feature compliance confirmation and payment/reward locks by product decision.
  - Did not import the `initializeFrontendCache` call from the unselected remainder of `b302be30`, so no missing `frontend-cache` dependency is introduced.

## 2026-05-22

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `2d1ca1538457869c8fbb650c02df628b6d943e5e`
- Upstream `main` synced through: `b9bc6f0e21fb3b87267475912fc7c31fa0494d6e`
- Local integration scope:
  - Integrated upstream performance metric active-group handling, request metadata extraction, route-link cleanup, and model `owned_by` resolution from active channels.
  - Integrated channel-test UX updates and normalized model pricing display formatting.
  - Integrated Turnstile handling on the new sign-up page.
  - Integrated Waffo Pancake wallet top-up and subscription payment support, including admin store/product binding and subscription-plan product IDs.
  - Integrated the upstream usage-log filtering revert.
- Notes:
  - Preserved the local decision to exclude paid-feature compliance confirmation and avoided adding payment/reward locks.
  - Preserved OmniRouters wallet corporate-transfer display, fee/currency handling, Sora/audio pricing, mail settings, OAuth icon URL support, login header, ticket workspace, and model marketplace custom labels.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-05-19

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `f69ceb6967901d3939f69d6467d9929ab41d8eae`
- Upstream `main` synced through: `2d1ca1538457869c8fbb650c02df628b6d943e5e`
- Local integration scope:
  - Integrated upstream auth fixes for registration visibility, affiliate payload field naming, forgot/reset password submission, and reset-confirm translations.
  - Integrated user/admin handling fixes for user deletion errors, create-user password validation, role-management checks, and sensitive access token JSON hiding.
  - Integrated default frontend fixes for wallet dark-mode preset selection, model detail tabs, table filter popover width, model-card pagination labels, API key drawer save handling, and API info color dots/save result handling.
  - Integrated upstream dashboard content visibility handling so overview panels respect the status switches.
  - Integrated upstream param override audit expansion for sensitive request body fields and log detail display.
  - Integrated upstream analytics placeholders and debug logging lazy-formatting cleanup.
- Notes:
  - Upstream commits were selectively integrated because local OmniRouters auth, wallet, pricing, i18n, and system-settings customizations touch the same files.
  - Preserved local wallet corporate transfer display, Sora/audio pricing, mail settings, OAuth icon URL support, and model marketplace custom labels.

## 2026-05-17

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `18282e610ddf3c8c39732fe84e50ded2cf6dcc7f`
- Upstream `main` synced through: `f69ceb6967901d3939f69d6467d9929ab41d8eae`
- Local integration scope:
  - Integrated `faa0f142` performance metric upsert column qualification.
  - Integrated `3caa6e46` default frontend batch fixes, including table truncation, channel fetch/edit UI, log columns, and wallet affiliate copy behavior.
  - Integrated `8f9ee9ba`, `554defe4`, `132d7b9f`, and `2d968c3e` channel remark clearing, usage-log filtering, and group-filtered channel list fixes.
  - Integrated `8a10dedb` unlimited API key quota validation handling.
  - Integrated `6f8668e4` public header navigation access control and middleware tests.
  - Integrated `68830e60` channel affinity `request_header` key source support.
  - Integrated `f69ceb69` new UI language and copy fixes.
- Notes:
  - Preserved OmniRouters external homepage/about iframe scroll handling while adding upstream public navigation auth prompts and visibility filtering.
  - Preserved OmniRouters login legal text translations and email settings / marketing email locale keys while accepting upstream i18n helper and locale updates.
  - Preserved local wallet corporate transfer display, model marketplace Sora pricing, discount badges, mail settings, and OAuth icon URL support.

## 2026-05-14

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `aa56667b8f233958aec0f593b3960c04230de6d3`
- Upstream `main` reviewed through: `18282e610ddf3c8c39732fe84e50ded2cf6dcc7f`
- Local integration scope:
  - Integrated `3e588b4d` electron lockfile `ip-address` dev dependency update.
  - Integrated `51b5cbe1` default frontend combobox focus filtering fix.
  - Integrated `18282e61` classic frontend `bun.lock` axios `1.15.2` alignment.
- Deferred:
  - `0526a226` paid-feature compliance confirmation was intentionally not integrated by product decision.

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
