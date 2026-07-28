# Upstream Sync Log

This file records the upstream `QuantumNous/new-api` commit that has been reviewed or integrated into this repository.

## 2026-07-28

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `3e1e728279884d83358811aec00980dd55f6ad4e`
- Upstream `main` reviewed and selectively synced through: `afe16c64cd73853da1eda3bf236f15d69637b4bf`
- Local integration commits:
  - `f9531cc27` parameterizes and sanitizes slow/error SQL logging, adds `SQL_SLOW_THRESHOLD_MS`, and removes the ineffective copied `CustomEvent` mutex.
  - `b0a13c1b7` adds configurable tool surcharges, Alpha Search billing, and the Sub2API channel while preserving checked quota conversion.
  - `b27dcec4a` adds New API channel type `60`; numeric type `58` remains reserved for upstream compatibility without enabling Advanced Custom.
  - `fdace003a`, `10b3a8056`, `b28cc9777`, and `249468137` migrate protocol conversion into the standalone RelayKit module, restore host-owned pricing/container types, configure trusted proxies in middleware, and harden Gemini stream state and terminal handling.
  - `32b1b9beb` adds per-channel HTTP protocol and HTTP/2 connection-shard controls across validation, client caching, runtime transport selection, tests, channel editing, and locales.
- Deferred, skipped, or equivalent:
  - Skipped `bc14c18f6` because it removes OmniRouters refund-pending, atomic refund-claim, reconciliation retry, and legacy-cutoff protections.
  - Skipped `f51dd4d80` because Advanced Custom remains disabled by product decision; its DTO compatibility plumbing stays dormant inside RelayKit.
  - Deferred `b27b2b1d6` because no equivalent iPad-specific session regression has been established locally.
  - `60a1acb70`, `c3db41407`, and the applicable part of `8e2bfe278` are already represented by the local migration and logging commits.
  - Skipped GitCode release workflows, upstream CI/rules changes, and RelayKit documentation-only commits because OmniRouters keeps its own GitHub/GHCR source-build workflow and project documentation policy.
- Notes:
  - Preserved Sora per-request pricing, the fixed per-call `audio_generation` surcharge, checked/saturated quota accounting, async OpenAI image tasks, task refund reconciliation, tickets, mail settings, wallet behavior, and model-mapping privacy.
  - HTTP transport settings validate `auto` or forced HTTP/1.1 and allow 1-8 HTTP/2 connection shards; Advanced Custom validation was intentionally excluded.
  - Locale JSON parsing, targeted channel formatting, conflict-marker checks, and source diff checks passed. Frontend typecheck no longer reports New API channel errors and only retains the previously known local `yace` installation, dynamic-price null-index, and duplicate `StatusBadge` issues.
  - Go tests could not run because Go is unavailable locally. The source-only workflow was followed; GitHub Actions remains responsible for dependency installation, tests, and image publishing after push.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-07-26

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `1721144221ec5c94dd87891a7ae1bee228e7bb63`
- Upstream `main` reviewed and selectively synced through: `3e1e728279884d83358811aec00980dd55f6ad4e`
- Local integration commits:
  - `f12f5a017` protects existing model pricing during create/rename, keeps open model forms stable during option refreshes, adds Gemini image GA models, fixes clearable/minimum top-up input behavior, debounces user search, and safely logs empty upstream errors.
  - `5127f76af` stabilizes debounced channel priority/weight updates and table row IDs, and adds Tencent TokenHub single-key routing while preserving native TC3 credentials.
  - `087c453a1` migrates real JSON configuration fields to the Yace-based unified editor with syntax highlighting, line numbers, formatting, copy, cursor location, form integration, and focused utility tests.
- Deferred, skipped, or excluded:
  - Skipped `cb96ab020` and `cbd9b30aa` because they only replace upstream GitHub issue templates and do not affect OmniRouters runtime behavior.
  - Skipped `257223be2` and `5ede832d8` because they only change upstream README badges.
  - Did not add `happy-dom`; it is only needed by upstream DOM-specific editor tests. OmniRouters keeps focused dependency-light utility coverage instead.
- Notes:
  - Preserved OmniRouters Sora per-request pricing, `audio_generation` fixed surcharge, payment and corporate-transfer settings, custom OAuth behavior, tickets, member tiers, mail settings, wallet behavior, and model-mapping privacy.
  - Kept the product decision to exclude paid-feature compliance confirmation.
  - Source checks passed for staged whitespace, locale JSON parsing, protected headers, and targeted frontend lint. Full typecheck still requires Actions to install the new `yace` dependency and continues to expose the previously known dynamic-price and upstream-ratio-sync errors.
  - Followed the source-only workflow; no local Docker image was built and GitHub Actions remains responsible for dependency installation and image publishing after a later push.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-07-22

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `7c28993f6bd9e92616f3f578212577f8b7c40b45`
- Upstream `main` reviewed and selectively synced through: `1721144221ec5c94dd87891a7ae1bee228e7bb63`
- Local integration scope:
  - Integrated authentication-data cleanup on hard user deletion, Telegram callback verification hardening, fail-closed 2FA queries, and concurrency-safe authentication updates.
  - Integrated server-side user sorting, immediate DataTable column updates, action-column sizing, fixed-height Playground model groups, unlimited-key used-quota display, and high-risk retry confirmation copy.
  - Integrated Responses-to-Chat streaming tool-call deduplication, MiniMax vendor inference, realtime GA header/model updates, and auto-group model listing while preserving local `ExtraGroups` behavior.
  - Integrated Codex upstream model discovery, compact model variants, proxy URL validation, canonical proxy-client caching, and precise cache invalidation while retaining the complete local static Codex model list.
  - Integrated async task failure compare-and-swap handling and refund reconciliation. OmniRouters additionally records `refund_pending` so only failures produced by the new transition path are eligible for reconciliation; historical failed tasks cannot be refunded a second time.
- Deferred, skipped, or excluded:
  - Deferred `a6cf42c0f` and `5a6c53d49` because Advanced Custom upstream model fetching and its wording changes remain outside the current channel policy.
  - Deferred `31d70fca3` and dependent `172114422` because the stateless-auth and `web/default` to `web` architecture migration is a high-impact authentication/frontend restructuring that requires a dedicated migration and rollout plan.
  - Skipped `d0e23e1e0` because it only changes repository language-statistics configuration.
- Notes:
  - Preserved OmniRouters Sora per-request pricing, `audio_generation` fixed surcharge and refund accounting, tickets, member tiers, extra user groups, email settings, marketing mail, wallet behavior, and model-mapping privacy.
  - Task refunds use the persisted total task quota and never reprice Sora or audio surcharges during recovery.
  - Source-level checks passed for whitespace, locale JSON parsing, and targeted frontend formatting. Go tests could not be run because Go is unavailable locally; GitHub Actions remains responsible for the full build and test run after push.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-07-13

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `8739c05c0e2aa96d69faec3b9f76b4d2c7f66108`
- Upstream `main` reviewed and selectively synced through: `7c28993f6bd9e92616f3f578212577f8b7c40b45`
- Local integration scope:
  - Integrated table resizing, Playground selector synchronization, GORM v2 row locks, case-sensitive custom model names, external iframe navigation, the `x/crypto` update, Codex field synchronization, model filtering, and referral-copy fixes.
  - Integrated upstream pricing-sync improvements, the unset-price channel-model workflow and hardening, stream/timing log details, and stale-instance actions while preserving local pricing columns.
  - Integrated saturated pre-consume rejection, image-count billing corrections, OpenAI image-stream disconnect protection, and the local Sora surcharge overflow audit path.
  - Selectively migrated `c36418c86` protocol conversion into the registry architecture with `BillingUsage` propagation across OpenAI, Responses, Claude, and Gemini formats.
  - Integrated native OpenAI `cache_write_tokens` parsing and cache-creation pricing, non-negative uncached remainder handling, tiered-expression propagation, and Compact `prompt_cache_key` forwarding from `48068ce92` and `92d3c9d18`.
- Deferred, skipped, or excluded:
  - Excluded Advanced Custom channel backend/frontend routing, channel-settings DTOs, model ability/cache/pricing rewrites, related translations, poster deletion, and upstream workflow changes from `c36418c86`.
  - Did not retain the broad design-system refactors in `262ab9312`, `0918bdb49`, and `9d1ca545e`; upstream subsequently reverted that line in `337169e0a` and `1b1b23d1d`.
  - Excluded unrelated dashboard/theme, sidebar hover-color, model-card unit, mobile log-card width, and branding/rules churn bundled with otherwise selected fixes.
  - Did not replay merge wrapper `ad900bbba`; selected underlying changes were integrated directly.
- Notes:
  - Preserved OmniRouters Sora per-request and `audio_generation` fixed surcharge billing, asynchronous OpenAI image handling, email settings and marketing mail, wallet/corporate-transfer behavior, model-mapping privacy, external homepage/about integration, tickets, custom OAuth/login behavior, and marketplace pricing labels.
  - Local validation was source-only because Go and Bun are not installed; conflict-marker checks, staged diff checks, relay-converter export checks, and all frontend locale JSON parsing passed.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-07-08 (follow-up)

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `6ce7305cd36f16506fb6a2c3c524a5a318539ba7`
- Upstream `main` reviewed and synced through: `8739c05c0e2aa96d69faec3b9f76b4d2c7f66108`
- Local integration scope:
  - Integrated default channel connection paste restoration and OAuth callback URL display/copy helpers.
  - Integrated browser-translation protection on React roots.
  - Integrated group-aware model-square dynamic pricing display, preserving OmniRouters Sora per-request pricing and `audio_generation` fixed surcharge display.
  - Integrated channel table manual column resizing.
  - Added playground per-model parameter settings, with Claude/Anthropic-like models defaulting `temperature` off.
- Notes:
  - Preserved local deferred Advanced Custom channel policy, custom OAuth behavior, model marketplace labels, Sora/audio surcharge pricing, model-mapping privacy, and the no local Docker image build policy.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-07-08

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `fc26b88fd131fecf179c94eacb86775748c60fd1`
- Upstream `main` reviewed and synced through: `6ce7305cd36f16506fb6a2c3c524a5a318539ba7`
- Local integration scope:
  - Integrated stale stream-write protection after client disconnects, quantity parameter validation, saturating quota conversions, quota saturation admin indicators, transactional row locking, adjusted quota bounds, and the tiered-expression default token estimate.
  - Integrated the PriceData other-ratio encapsulation while preserving local fixed-quota support for Sora and `audio_generation` surcharges.
  - Integrated subscription quota reset actions, stale system-instance cleanup, wallet reward-transfer quota-unit handling, zh-TW locale resources, Intl locale normalization, Chinese browser-language detection mapping, and the classic build source fix.
  - Restored the upstream classic-builder filtered dependency install so classic builds do not resolve against the default frontend `date-fns` dependency set.
  - Integrated the 5.6 preparation, group-ratio decimal draft editing fix, and GPT-5.6 token ratios.
- Deferred, skipped, or equivalent:
  - Skipped `5cbb7b0be17c7258def3ec835e0fee26076caac0` because it only updates upstream README architecture requirements.
  - Skipped `8bc4bf1d6b1fd7d117100edadcf4257d3a4eb479` because Docker cosign signing and publishing permissions are handled by the GitHub image workflow policy rather than local source sync.
- Notes:
  - Preserved OmniRouters email settings, marketing mail, wallet/corporate-transfer display, Sora per-request and `audio_generation` fixed-surcharge billing, model-mapping privacy, external homepage/about integration, tickets, custom OAuth behavior, and local no-payment-compliance-gate policy.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-07-06

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `b6e8ff9dd5d22399ca558193ced497e4aa81680a`
- Upstream `main` reviewed and synced through: `fc26b88fd131fecf179c94eacb86775748c60fd1`
- Local integration scope:
  - Integrated the classic build source fix, redemption-code filtering and cleanup, username normalization, Codex passthrough controls, mobile user-card ordering, Shadow DOM theme synchronization, and authenticated sign-up redirect.
  - Integrated account email/password hardening, disabled-token read-only protection, opt-in Secure session cookies, and their regression coverage.
  - Integrated async task quota persistence and Ali video-duration fallback while preserving local Sora per-request and `audio_generation` fixed surcharges.
  - Integrated dial-time SSRF protection for user-controlled downloads, webhooks, notifications, and media proxies without applying it to operator-configured model upstreams.
  - Integrated compact dynamic-pricing log display and the group-ratio editor redesign while retaining local fixed-price columns, Sora/audio pricing, and model-mapping privacy.
- Deferred, skipped, or equivalent:
  - Skipped `a1301039` because it only renames local Makefile development targets.
  - Skipped `86021d8e` because the classic deprecation banner and broad generated i18n refresh are not required and include unrelated compliance-copy churn.
  - Skipped `1e80ce03` because it only changes legacy audit-warning copy and conflicts broadly with local translation files.
  - `2f91d8cc` is already superseded by the local `ExternalContentFrame` theme, language, and scroll bridge.
  - Did not replay merge wrappers `722d0366` and `1ae75747`; their selected underlying changes were integrated directly.
- Notes:
  - Preserved local email templates, marketing mail, extra user groups, member tiers, tickets, GHCR publishing, wallet behavior, and external homepage/about integration.
  - Redemption failures no longer expose detailed state to users or write full redemption keys to server logs.
  - Controller JSON decoding touched by this sync now uses the shared `common/json.go` helpers.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-07-04

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `8874d1929f97bb3f7fcae2af81c9e114535044f1`
- Upstream `main` reviewed through: `b6e8ff9dd5d22399ca558193ced497e4aa81680a`
- Local integration scope:
  - Integrated `0565e626793da84606190f9ad0d4e2d0fa349202` so transient network and server errors do not invalidate authenticated sessions.
  - Integrated `bfddc5fea0ba9010a14d7f88f43f08478aac9446` and `dfc0d6324b40c1d6c2972e524409f933541bfb0f` for access-token query protection and race-safe user setting/cache updates.
  - Integrated `69c4d83df403c834c1476b732add66ca8a08b36e` and `1dcb389d008ea912224abedff3aeb8cde28e2693` for the Go network and image dependency updates.
  - Integrated `0977965d933f599b0bbed3ca501b67abce6ce712` for Ollama non-stream tool-call conversion and regression tests.
  - Integrated `a6c02012571602b9d7da8ab3a35c6471460d0b3a`, `aa334c0850b10ca91d56d1b0792927c26575f1c9`, and `70c0b37eec6ec8d31196f2a61090f861cb210a45` for the AI SDK 7 frontend dependency set and nested usage details.
  - Integrated `c1903607d5c1d22a503153c81b694ed72b50a97d`, `b35dfa32efad17a6acb01ccea6a241cc8d1ab2c5`, and `c5600f9b11b913a8898f8ac340ab56b856afd2c4` for channel filter persistence and channel test dialog improvements.
- Deferred or skipped:
  - Skipped `55858f353c95376d1e1d69af8893e9b9e182d0d0` because it replaces local image workflows with upstream Docker Hub publishing and removes workflows still used by OmniRouters.
  - Skipped the empty Electron dependency commit `917a2cff64feed0acd687298252bd400adf293e0`.
  - Did not cherry-pick merge wrappers `f5bba114` and `b6e8ff9d`; their selected underlying commits were integrated directly.
- Notes:
  - Preserved local member tiers, extra user groups, tickets, email settings, billing extensions, and GHCR publishing.
  - Extended the upstream non-quota cache refresh to include `ExtraGroups` without overwriting quota fields.
  - Extended access-token omission to local bulk user queries used by tickets and marketing email services.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-07-03

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `5bf346836273fc201a0a479aff5b257b4553e478`
- Upstream `main` reviewed through: `8874d1929f97bb3f7fcae2af81c9e114535044f1`
- Local integration scope:
  - Integrated `52858ad1e617069b708d820e1ea8a312b8077c85` for Wan2.7 image-to-video media mapping and regression coverage.
  - Integrated `759ab6bbca57074d85822f6e31b9332b383d3552` so route-local tabs retain their page state.
  - Integrated `fda8177864d7ce0f3ebf9c615e1700eb234b5b45` and `1f4d8d2b26815e836cda6ca16e8920fdabd7d069` for isolated custom HTML rendering and application style injection.
  - Integrated `986d90ae046f28e2f7377715a4feb721b57b52eb` and the code changes from `8874d1929f97bb3f7fcae2af81c9e114535044f1` for graceful shutdown, synchronous quota aggregation, and startup logging order.
  - Integrated `95e8c5eecff54dc032c7bbc9aa2ca2a37b13c766` for the Rsbuild and Tailwind build pipeline update.
- Deferred or skipped:
  - Skipped the Makefile-only developer workflow changes in `f9165e7b` and `e1fd9cc2`.
  - Skipped the upstream-only `AGENTS.md` update in `bff701b0`.
  - Excluded the poster assets bundled with `8874d192`; they are unrelated to runtime behavior.
- Notes:
  - Preserved the local `ExternalContentFrame` handling for external home and about page URLs.
  - Preserved local mock scripts, `historyApiFallback`, support ticket and member tier startup tasks, and log model-mapping privacy filtering.
  - Preserved the local unified task media normalization and added the Wan2.7 direct-request regression test without duplicating normalization logic.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-07-01

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `25f998595d2da4ac9c749f3eae8fffcf9047bc3e`
- Upstream `main` reviewed through: `5bf346836273fc201a0a479aff5b257b4553e478`
- Local integration scope:
  - Integrated `c8491b41bc4491f9cda5a440b481bb53f9d2d7e1` so Doubao Seedance 2.0 task billing reflects output resolution and video input.
  - Integrated `e514db20f762649014bce8950ef85b182f5f1b3f` for Seedance 2.0 `safety_identifier`, explicit-zero-safe `priority`, and 4K billing.
  - Added focused regression coverage for the standard/fast model billing matrix, request metadata extraction, explicit `priority: 0` forwarding, and the local blank-prompt behavior.
- Deferred or skipped:
  - Kept Advanced Custom channel support deferred, so `43591fba` was not integrated.
  - Skipped the Electron-only lockfile refresh in `12fc0100`; server and GHCR builds do not consume it.
  - Skipped the 679-file frontend formatting commit `5bf34683` to avoid unrelated churn across OmniRouters custom UI and project rules.
- Notes:
  - Preserved the local Doubao behavior that omits blank prompt content.
  - Preserved OmniRouters Sora per-request and `audio_generation` fixed-surcharge billing; Seedance uses its own task adapter and `OtherRatios` path.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-06-29

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `9ba251ce5f2acdbeda9d67fac3fb7353ec666955`
- Upstream `main` reviewed and synced through: `3a506f50f08b4c11968f102972bc814e0fc9da0d`
- Local integration scope:
  - Integrated Waffo goods information and SDK updates, tiered-pricing editor refresh, theme-switch route reset, async task node attribution, ClickHouse `LIKE` filtering, and password validation copy from `79396745`, `3245b2b7`, `c0e42bfb`, `d10fc762`, `df44a75d`, and `df5ba9fa`.
  - Integrated rich HTML/Markdown rendering and sanitization from `0b48ad86` and `626dadb5`, while retaining the local external-page theme, language, and scroll bridge.
  - Integrated OpenAI Chat-to-Responses compatibility hardening from `3a506f50`.
  - Integrated the Playground experience and Markdown renderer refactor from `966af88e`, preserving group-aware OpenAI endpoint model filtering.
  - Integrated channel-scoped administrator permissions from `4aee5f7d`, preserving the local Codex OAuth/usage routes and classifying credential-changing operations as sensitive.
- Already satisfied or skipped:
  - `35074345` was already satisfied by the existing exact DOMPurify `3.4.11` package and lock entries.
  - `6c35e1ef` only updates upstream i18n skill metadata; local project skills and rules remain authoritative.
- Notes:
  - Preserved OmniRouters mail settings and marketing mail, wallet fee/currency and corporate-transfer display, Sora per-request/audio-generation surcharge pricing, model marketplace labels, custom OAuth/login behavior, ticket routes, task-log model display, OpenAI asynchronous image handling, and external homepage/about integration.
  - Added regression coverage for the local Codex OAuth channel routes and combined Playground group/endpoint filtering.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-06-29 (follow-up)

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `3a506f50f08b4c11968f102972bc814e0fc9da0d`
- Upstream `main` reviewed and synced through: `25f998595d2da4ac9c749f3eae8fffcf9047bc3e`
- Local integration scope:
  - Integrated the Responses-to-Chat and Chat-to-Responses conversion expansion from `2d5a041639e56316a45e8a2a11e5b5fb50c7a6ac`, including Gemini Responses relay support, stream handling, and conversion regression tests.
  - Integrated the channel editor section navigation and layout update from `1d166532fe954a45207dffd2924697796a984159`.
  - Integrated the channel management refinements from `25f998595d2da4ac9c749f3eae8fffcf9047bc3e`, including opt-in batch operations, consistency-repair confirmation, card/table sensitive-data masking, JSON editor initialization, and channel drawer navigation refinements.
- Adaptations:
  - Excluded Advanced Custom channel DTO, adapter, editor, and translation fragments because that channel remains intentionally deferred in OmniRouters.
  - Restored the local Codex OAuth authorization dialog and credential workflow in the refactored channel drawer.
  - Preserved channel-scoped administrator permissions, sensitive-field write protection, and the local Codex usage/operation route classifications.
- Notes:
  - Preserved OmniRouters mail settings and marketing mail, wallet fee/currency and corporate-transfer display, Sora per-request/audio-generation surcharge pricing, model marketplace labels, custom OAuth/login behavior, tickets, task-log model privacy, OpenAI asynchronous image handling, and external homepage/about integration.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-06-25

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `f8cfbfa4dfd92e0a34f1e62225af2223dad48f62`
- Upstream `main` reviewed and synced through: `9ba251ce5f2acdbeda9d67fac3fb7353ec666955`
- Local integration scope:
  - Integrated SMTP STARTTLS/NTLM support, the SMTP PLAIN TLS guard, ClickHouse dependency updates, and DOMPurify dependency hardening from `2f23a667`, `cf6ae6fd`, `993d67eb`, `acb52d0f`, and `0bf42781`.
  - Integrated small default-frontend fixes for channel card selection, wallet-menu visibility, user token limits, async polling delay, token-limit translations, and neutral drawing task labels from `de0d6ac9`, `0b2cf43e`, `5d943281`, `d2dcbc31`, `5814ca90`, and `b191f473`.
  - Integrated the persistent system task runner and system instance info panel from `53771922` and `2cbdfa03`, while preserving OmniRouters ticket maintenance and member-tier scheduled tasks.
  - Integrated the expanded default Markdown renderer from `f4473d96`, keeping DOMPurify locked to `3.4.11`.
  - Integrated the DataTable row-action/destructive-dialog refactor from `9ba251ce`, preserving OmniRouters API-key CC Switch/chat actions and keeping subscription row actions free of the upstream paid-feature compliance gate.
- Deferred or skipped:
  - Reviewed but did not integrate upstream agent/skill metadata rewrites from `9fc9c8f1`, `72b3f345`, and `ad35ab1d`; local `.agents` skills and project rules remain the source of truth.
  - Did not take the date-fns/classic release build commits `64eafc94`, `48da37a3`, `69b0f0b5`, and `c12e5db4` because OmniRouters already carries the local classic date-fns resolution and Docker/GHCR image builds are handled by GitHub Actions.
- Notes:
  - Preserved OmniRouters mail settings under the admin console mail-settings page, existing ticket/member-tier background tasks, local model marketplace pricing/badge behavior, Sora/audio-generation pricing customizations, and API-key chat/CC Switch integrations.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-06-23

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `0229dc20573f728ec1140543cdee291197d67e8c`
- Upstream `main` reviewed through: `f8cfbfa4`
- Local integration scope:
  - Integrated small backend/frontend fixes from `2c2d9211`, `fae39cd9`, `74091744`, `0c6c1b37`, `d58029c6`, `1f1da553`, `0b7ae4ea`, and `354d0fed`.
  - Integrated routing reliability, channel-card performance, channel test environment toggles, and passive channel monitoring from `91ab664c`, `f9e508bd`, `44e0e686`, and `efd6c445`.
  - Integrated Dashboard traffic-flow Sankey chart, node limits, persistent filters, interactive highlighting, and sensitive-data toggle from `a68041f7`, `06194801`, `8ad83bf6`, and `5e866446`.
  - Integrated ClickHouse log database support and log cleanup task improvements from `6dc4030f`, `a162163b`, `f84b7d59`, and `f8cfbfa4`.
  - Integrated the default-frontend `tsgo` typecheck tooling migration from `e5694748`.
- Deferred or skipped:
  - Skipped `cb841850` because the referenced channel type is not present locally.
  - Did not introduce upstream `allow_wallet_overflow` / downgrade fields from `dfcb74b5`; only retained applicable low-risk migration cleanup already compatible with local behavior.
  - Preserved local `AGENTS.md` and `CLAUDE.md` project rules.
  - Kept OmniRouters mail settings under the admin console mail-settings area instead of moving SMTP/mail options back into upstream operations settings.
  - Kept the legacy synchronous `DELETE /api/log/` handler and route for the classic frontend while the default frontend now uses the persistent `/api/system-task/log-cleanup` task flow.
- Notes:
  - Preserved OmniRouters tickets, member-tier tasks, wallet/payment display, mail settings, marketplace labels, Sora/audio pricing, route privacy, channel sensitive-info masking, and local model metadata migrations while integrating this batch.
  - `perf_metrics_setting.excluded_status_codes` remains exposed in Monitoring & Alerts after upstream moved performance metrics out of Performance settings.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-06-20

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `f7dae5cbe620c4f643b0802ab9221e221036e996`
- Upstream `main` reviewed through: `0229dc20573f728ec1140543cdee291197d67e8c`
- Local integration scope:
  - Integrated first-pass low-risk frontend updates from `6ad5dbb641c40771b8ca4c2e75f8ee8f56d831c2`, `490395b2f8a62c95ca376d25cb104ab170b523f8`, `0c806db9961c3686db4fbd63764f70bbfcf899fe`, and `6bd69f3edfdd21bcd8817adbed4485f2295e1319`.
  - Selectively integrated usage-log retry-chain and multi-key index display from `0467d5401430680b7b09a6ce83a5cea2136f0a95`, without taking the upstream frontend branding-copy changes.
  - Adapted the channel sensitive-info toggle from `9100e15e4e9c9b31bb46dc1348d99ec7c9e1383d` and then completed the DataTable card/table view, channel-card, mobile-card, toolbar, and model-drawer refinements from `685855689232474b55d05dc6d6c9f0ea8b523d4e`, `29c3dcb9c8add1b49d6ba78ce2be0f72d6391ecc`, `a0de4b56090f170ebf69c3ce87f048a236d2cf5c`, `50b8f2a2399322bb3b46ef15086ae64205eb6506`, `4206d7fd7d81def24e7a6f8ad36d0d2d37766c62`, and `0229dc20573f728ec1140543cdee291197d67e8c`.
  - Integrated the Codex usage UI/backend refactor from `3fcd741c4eb9e620186e9f6f2fb1c1c45dded19b`, preserving the local channel sensitive-info masking in the default frontend dialog.
  - Integrated the default-frontend OXC tooling migration from `5b4839fa370af1444424f9383e0150805cd4ee36`, plus the later `curly` rule relaxation from `50b8f2a2399322bb3b46ef15086ae64205eb6506`.
- Deferred or skipped:
  - Skipped the upstream frontend branding-copy changes from `0467d5401430680b7b09a6ce83a5cea2136f0a95`.
  - Preserved local `AGENTS.md` and `CLAUDE.md` rules instead of taking the upstream restructuring from `490395b2f8a62c95ca376d25cb104ab170b523f8`.
  - Did not add channel type `58` to the default-frontend display order because that channel type is not present locally.
  - Did not apply the later Codex dialog import/style-only reshuffle from `50b8f2a2399322bb3b46ef15086ae64205eb6506`; retained the functional `3fcd741c4eb9e620186e9f6f2fb1c1c45dded19b` implementation to keep protected headers intact.
- Notes:
  - Preserved OmniRouters-specific marketplace, wallet, mail, Sora/audio pricing, route privacy, and channel sensitive-info behavior while integrating this batch.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-06-18

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `eb8631160423f4e4548c1949590d34b07f95ccd8`
- Upstream `main` reviewed through: `f7dae5cbe620c4f643b0802ab9221e221036e996`
- Local integration scope:
  - Integrated first-batch low-risk fixes from `17e342738e6cc28d8252e566e15ac9d977db48ca`, `34287afec74392f4518d07873e5c148377566e0a`, `43c7e30a4732590e7bd684ad7fba7bab0af8e7b0`, `b798e3496fa2ebbfd438f6eda3ab0ff674d58aa7`, `502858d35d053c89e68e891a8e5ad53f1c86bfa7`, `97eadbefa667e5405aac352c758dfe7538043f17`, `1aa77e6698f70d1f7e0659653aa5c5c0c354ac54`, and `122a730a7748942ec4a0ec890aa1e9d089f97c6f`.
  - Integrated second-batch usage-log dialog and channel test dialog fixes from `9b9b19e9d2b42d1ddeaf72ba8b41c692c8d1caab`, `a2f3ac02e4268c3a880f277b993900c13f6e6a83`, `3cc2b1bea44f25dfe2aee70215ed43a25c7b3769`, and `a37ce3d6b5d0b129357b23818447475854ae0628`.
  - Integrated third-batch performance and model-details updates from `21636fc1b5a5d7bba21a399ce0c09c69363caa1c`, `a95655a245522a9d3b433f7fe3f7c2cb4424e99a`, `208d86e96e5a9f6c08f7e9372bb1b6f066b4a748`, `df013946d77a802c03b0688a9a677059371c6e35`, and `1414569b636c1e66bf88df2d27d82df5f252973e`.
  - Integrated fourth-batch OAuth/test/payment cleanup from `3b345cfd45a685bd90b8704b1dfe830d89758d15`, `8d87d5fd52b5c9ec5be093870886485632f5c5e0`, `2154fce08d97388e9bf8805801a9c6f7b2d59757`, and `06f9dec915f0cab308bd4bb04a4db2b454c2edee`.
  - Integrated fifth-batch usage-log layout and number/timestamp formatting fixes from `cfc9bbcdbf4a21c96093afe008d45b3156e9a8d0`, `f4575fe6641e4703a4ce36c9477dde1f31fc2c66`, and `f7dae5cbe620c4f643b0802ab9221e221036e996`.
  - Preserved local ticket test tables while adding upstream OAuth binding cleanup coverage.
  - Kept the Claude empty tool-call-arguments behavior while using the project JSON wrapper for the touched parse path.
  - Preserved the local usage-log privacy rule that only admins can see the actual mapped upstream model.
  - Preserved OmniRouters Sora/audio-generation pricing display, marketplace NEW/discount badges, and three-bar model status display while adopting upstream model details and recent performance metrics.
  - Preserved OmniRouters wallet currency/fee/corporate-transfer display while adding payment-method icon support and clearer minimum-top-up button labels.
  - Kept the local decision to skip upstream paid-feature compliance confirmation UI while integrating the safe payment icon and recharge display pieces.
- Deferred:
  - Still deferred from previous reviews: column visibility persistence (`4e8b5e9b`), subscription wallet-overflow/downgrade support (`f6c26043` and `fd557064`), setup guide expansion tweak (`b8f5ba4a`), subscription product GET endpoints (`50784c10`), rankings period removal (`21d4d18d`), and advanced custom channel support (`3f2c0aed` plus `55b00fcf`).
- Notes:
  - Legacy OAuth controller files were already absent locally and the API router already uses the unified OAuth handler.
  - Some reward-hacking test cleanup from `8d87d5fd` was not directly applicable because local test files have diverged or are absent; the applicable redundant/random/timing-heavy cases were removed.
  - Existing local `web/default/src/routeTree.gen.ts` changes and untracked `bin/sync-upstream.ps1` were not part of this upstream batch.

## 2026-06-16

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `9bc1a53dea0d9e5c9e6f165d3bd8a681fd925fc1`
- Upstream `main` synced through: `eb8631160423f4e4548c1949590d34b07f95ccd8`
- Local integration scope:
  - Integrated channel test dialog status refresh from `be60e25a94dad0d24188ed28c13c93db9ee7cb7b` and the follow-up state-reset fix from `c67c6fc740a6ce58f625167fe36d9e2404b39201`.
  - Integrated DataTable header-selector, row-selection memoization, and column-class comparison fixes from `a59e0eb59afeef807095c318b8b9f66e6d5dad76`, `8477f6288b7277388c1f46fd4788d8799c3094e9`, and `179f69dfcaa232600157e8602412eb6a8a744732`.
  - Integrated wallet/profile/table interaction alignment from `eb8631160423f4e4548c1949590d34b07f95ccd8`.
- Skipped:
  - `426c9664e98a36aa5fca308e2cf02f37e23d9d0b` was already satisfied locally after resolving `data-table-row.tsx` with a trailing newline.
- Notes:
  - Preserved OmniRouters custom profile metadata line (`username`, `ID`, user group, role), wallet corporate-transfer/payment fee/currency handling, Sora/audio-generation pricing display, custom OAuth icon URL behavior, and user-facing hidden routed-model details.
  - Did not introduce upstream paid-feature compliance disabling into subscription row actions, because OmniRouters intentionally keeps that feature disabled.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

## 2026-06-15

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `51475c8062e53d4f731ce9e5f7e6c8ea9ce77cb9`
- Upstream `main` reviewed through: `9bc1a53dea0d9e5c9e6f165d3bd8a681fd925fc1`
- Local integration scope:
  - Integrated authentication-method tracking in operation audit logs from `1ac0f5807a8a538fa9a1ec1d86e692210243ed6a`, including default-frontend log detail display and locale entries.
  - Integrated API key form option refresh behavior from `aeea3fae9bf6deb01a90c71016a7361fb5532ed1`, so the drawer refetches models/groups when opened.
  - Integrated the CC Switch model selector fix from `9bc1a53dea0d9e5c9e6f165d3bd8a681fd925fc1`, keeping `CC Switch` as an untranslated product name.
  - Integrated the DataTable text/badge overflow fixes from `3c1bb0a7d3e3b4039dd1caf2220f10967587d841`, including shared truncated/badge cell wrappers and affected list-table columns.
- Notes:
  - Preserved OmniRouters custom model-marketplace labels, Sora/audio pricing, wallet/payment display, mail settings, tickets, profile display, OAuth icon URL behavior, task-log model display, and Codex OAuth surfaces.
  - Kept the local API info settings state/save flow while applying only the upstream table overflow display wrappers there.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

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
