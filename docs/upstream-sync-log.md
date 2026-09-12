# Upstream Sync Log

## 2026-09-12 - Normalize text performance cache input

- Local reason: the text settlement performance path passed total prompt tokens plus cached tokens, double-counting cached input in the cache-rate denominator.
- Changed: pass uncached prompt tokens separately, matching the ordinary settlement path; cache rate now uses cached / (cached + uncached).
- Preserved: billing, quota settlement, usage logs, routing, and provider usage parsing.
- Validation: `git diff --check`; Go compilation and browser verification unavailable locally.
- Commit/push: pending.

## 2026-09-12 - Preserve cache counters in performance group results

- Local reason: performance cards returned an absent cache rate even when time buckets contained cache observations. `buildQueryResult` omitted input/cache counters from group totals.
- Changed: sum both counters across buckets before computing the group cache rate. Added regressions for persisted-shaped and live buckets, hit-only denominators, and missing observations.
- Preserved: request recording, database schema, pricing and settlement, token normalization, percentile behavior, and no-cache display. This addresses missing aggregation, not historical denominator accuracy or all provider mappings.
- Validation: source review and diff check; Go tests cannot be executed without a local Go toolchain. No production API or browser verification performed.
- Commit/push: local changes only; not committed or pushed.

## 2026-09-12 (TPOT percentile unit display)

- Local reason: TPOT percentile values were formatted as generic latency, causing millisecond-per-token values to display with the `s` suffix.
- Changed: TPOT P95/P99 use the dedicated `ms/token` formatter; TTFT P95/P99 continue using latency formatting.
- Validation: `git diff --check`; Go compilation and browser verification unavailable locally.
- Local commit and push status: pending.

## 2026-09-12 (historical TPOT sample compatibility)

- Local reason: previously persisted TPOT percentile samples used the old seconds-to-milliseconds conversion twice.
- Changed: performance buckets carry a sample-format version; legacy samples are converted on read while new samples use the corrected unit.
- Preserved: numeric aggregates and all billing behavior.
- Validation: `git diff --check`; Go compilation and browser verification unavailable locally.
- Local commit and push status: pending.

## 2026-09-12 (performance TPOT and percentile validity)

- Local reason: TPOT was multiplied by 1000 although generation duration was already stored in milliseconds, and P95/P99 were shown for very small samples.
- Changed: TPOT and its percentile samples now use milliseconds per output token; percentile values remain hidden until at least 10 valid samples exist.
- Preserved: billing, cache accounting, routing, and raw latency/TTFT measurements.
- Validation: `git diff --check`; Go compilation and browser verification unavailable locally.
- Local commit and push status: pending.

## 2026-09-12 (performance cache input mapping)

- Local reason: usage logs contained cache tokens, but the standard relay performance recording path passed only uncached prompt tokens to cache-rate aggregation.
- Changed: the performance sample now passes total input tokens (uncached prompt plus cached input) while retaining cached tokens separately, so cache rate matches the usage log.
- Preserved: billing, quota settlement, routing, and usage-log accounting.
- Validation: `git diff --check`; Go compilation and browser verification unavailable locally.
- Local commit and push status: pending.

## 2026-09-11 (dynamic card request price)

- Local reason: model cards displayed dynamic request prices in internal quota units while the detail table used currency units.
- Changed: dynamic card request-price ranges now convert the internal millionth-unit value before formatting, producing values such as `$0.068-$0.080/request`; group discounts remain applied.
- Validation: `git diff --check`; Go compilation and browser verification unavailable locally.
- Local commit and push status: pending.

## 2026-09-11 (dynamic tier detail formatting)

- Local reason: dynamic expression tier prices were displayed as internal quota values and omitted the request unit.
- Changed: numeric tier prices are displayed as currency per request in the dynamic pricing table; billing and group multiplier calculations are unchanged.
- Validation: `git diff --check`; Go compilation and browser verification unavailable locally.
- Local commit and push status: pending.

## 2026-09-11 (dynamic tier price display)

- Local reason: expressions using `tier("label", 68000)` were billed correctly but displayed no model-square price because the frontend parser only recognized `fixed(68000)`.
- Changed: frontend tier parsing now recognizes numeric tier prices as request prices and displays their range; billing evaluation and accounting are unchanged.
- Validation: JSON locale parsing and `git diff --check`; browser verification and Go compilation unavailable locally.
- Local commit and push status: pending.

## 2026-09-11 (dynamic request price formatting)

- Local reason: numeric `tier` prices were incorrectly repeated as the base price and displayed in internal quota units.
- Changed: request prices are converted to currency units and shown only in the tier breakdown/range; the base-price section remains labeled dynamic pricing. Billing evaluation is unchanged.
- Validation: `git diff --check`; Go compilation and browser verification unavailable locally.
- Local commit and push status: pending.

## 2026-09-11 (bounded TTFT/TPOT percentile sampling)

- Local reason: add scalable P95/P99 performance indicators for high request volume.
- Changed: added bounded reservoir sampling for TTFT and TPOT and exposed group P95/P99 values in the performance response. Each bucket retains at most 256 samples per metric; samples are stored as bounded JSON snapshots on the existing performance bucket and no per-request database rows are created.
- Preserved: numeric aggregates, billing, relay routing, cache accounting, and the existing retention cleanup. Database migration is additive through GORM and remains compatible with SQLite, MySQL, and PostgreSQL.
- Validation: `git diff --check`; targeted Go test was added but Go is unavailable locally; Bun typecheck ran and reports pre-existing repository errors outside this change.
- Local commit and push status: pending.

## 2026-09-11 (TPOT unit display)

- Local reason: TPOT was displayed with the generic latency formatter and could be read as seconds without the per-token unit.
- Changed: added a dedicated TPOT formatter that treats the value as milliseconds and displays `ms/token` or `s/token`.
- Preserved: TPOT calculation, TPS, TTFT, latency, billing, and routing behavior.
- Validation: `git diff --check`; frontend build and browser verification were unavailable locally.
- Local commit and push status: pending.

## 2026-09-11 (cache rate hit-only sampling)

- Local reason: cache-rate statistics should exclude requests with no cache hit from the cache-rate denominator.
- Changed: performance aggregation records cached-token samples only when cached read tokens are positive; cache rate is cached tokens divided by total input tokens (cached plus uncached), so a request with 1,073 uncached and 229,376 cached tokens reports 99.53%.
- Preserved: billing calculations and all non-cache performance metrics.
- Validation: `git diff --check`; Go/Bun and browser verification were unavailable locally.
- Local commit and push status: pending.

## 2026-09-11 (overview panel height alignment)

- Local reason: API information and announcement cards rendered with different heights in the overview grid.
- Changed: aligned the API information content area with the announcement area at `h-96` and made both grid items fill the shared row height.
- Preserved: API checks, announcement rendering, data contracts, and interactions.
- Validation: `git diff --check`; browser and frontend build verification were unavailable locally.
- Local commit and push status: pending.

## 2026-09-11 (performance metric translations)

- Local reason: the TTFT, TPOT, and cache-rate cards contained untranslated labels and descriptions.
- Changed: connected TTFT and TPOT labels to i18n; the cache-rate label and all three descriptions already use translation lookup keys and will use locale fallback until translated entries are synchronized.
- Validation: `git diff --check`; frontend build was unavailable locally.
- Local commit and push status: pending.

## 2026-09-11 (performance metrics build fix)

- Local reason: GitHub Actions reported missing input/cache fields on `PerfMetricSummaryBucket` after adding cache-rate aggregation.
- Changed: added the fields to the summary bucket DTO and included them in the bucket aggregation query so the performance package compiles and receives persisted cache counters.
- Validation: `git diff --check`; GitHub Actions will provide the Go build verification.
- Local commit and push status: pending.

## 2026-09-11 (model performance TPOT and cache rate)

- Local reason: add TPOT to the model details performance view.
- Changed: exposed real TPOT values from the existing output-token and generation-duration aggregates, and added input/cache-token counters to calculate cache hit rate by model and group. The performance page now shows six cards.
- Preserved: billing, request routing, existing TTFT/latency/success/TPS calculations, and database compatibility.
- Validation: `git diff --check`; Go/Bun and browser verification were unavailable locally.
- Local commit and push status: pending.

## 2026-09-11 (overview announcement rich content)

- Local reason: overview announcements displayed raw Markdown image syntax and HTML instead of rendered content.
- Changed: overview announcements use the existing sanitized RichContent renderer on the complete content, preserving Markdown blocks and HTML structure. A separate detail button avoids nesting content links inside a button.
- Preserved: announcement storage, notification popover, timeline, and detail dialog.
- Validation: `git diff --check` passed; browser rendering and frontend build not verified locally.
- Local commit and push status: pending.

### 2026-09-09 Sync upstream options, Volcengine, and Alibaba Wan fixes

- Integrated `4fc9d1f1f`: added a cross-database `options` table primary-key repair migration that deduplicates legacy rows, preserves the old table as a backup, and runs before normal option auto-migration. This protects persisted settings such as Waffo Pancake unit price from duplicate-key/readback corruption.
- Integrated `876903a8e`: corrected Volcengine upstream model discovery from `/v1/models` to `/api/v3/models`, while retaining configured special-base handling.
- Integrated `a20574136`: updated the Alibaba Wan task plugin with model capability profiles, image-to-video routing, Wan 2.7/Wan 3.0 media validation, supported duration/resolution handling, completion usage facts, nested video artifact URLs, and the corresponding protocol regression tests.
- Already present and deliberately preserved: the local `Option.Key` primary-key model declaration, Waffo Pancake save validation and payment flow, Sora per-request billing, `audio_generation` surcharge billing, plugin registry/runtime, and task polling/settlement behavior.
- Deliberately not copied from `4fc9d1f1f`: upstream's new model-pricing configuration file and its associated architecture-specific pricing mutation changes, because the local pricing implementation has separate group/customer discount and Sora/audio boundaries. The compatible database repair portion is integrated instead.
- Deferred for separate evaluation: `c79b74b68` (upstream frontend status-query changes), `12be9975c` (large frontend error-notification rewrite), and `7cf9b473f` (upstream `web/src` pricing scroll layout). `d52bdc0b4` and `064ed943e` are recorded below as audited/integrated for this round.
- Validation: source review and `git diff --check`; no local Go/frontend compilation, Docker build, or image publication per the source-only workflow. Local commit and push are pending.

## 2026-09-07 (first-round polling audit)

- Reviewed upstream `9df450fe5` against the plugin task subsystem.
- Already present locally: persisted task query contexts, response status classification, bounded consecutive poll failures, CAS-safe failure updates, plugin state persistence limits, and task refund/settlement integration. No upstream polling code was copied.
- Preserved: local Sora/audio billing, plugin registry, task artifacts, and provider-specific polling adapters.
- Validation: source comparison and `git diff --check`; full Go tests unavailable because Go is not installed locally.

## 2026-09-08 (authentication batch: duplicate OAuth route safety)

- Removed duplicate Telegram OAuth route registrations in `router/api-router.go`, which could cause Gin to panic during startup even when the Go build succeeded.
- Added a router regression test covering the Telegram-specific and provider wildcard GET routes. Existing authentication, OAuth, billing, payment, mail, plugin, and frontend behavior remains unchanged.
- Validation: source review and `git diff --check` only; no local compilation or tests, per the source-only workflow. GitHub Actions is the verification gate.
- Local commit/push: pending on `main`.

### 2026-09-09 Add standalone audit log page

- Added the administrator-only `/audit-logs` page, typed `/api/audit/` query, and sidebar entry.
- Existing backend audit records and usage-log audit details remain unchanged.
- Validation: source review and `git diff --check` only; not pushed yet.

### 2026-09-09 Save Waffo Pancake unit price independently

- Fixed unit-price changes to use the normal option update path instead of requiring a Store/Product binding save. Changing only `1` to `1.05` no longer fails because the catalog binding is incomplete.
- Validation: source review and `git diff --check` only; not pushed yet.

### 2026-09-09 Fix audit log response unwrapping

- Fixed the standalone audit page API client to unwrap the common `{ success, data }` response envelope before reading paginated items.
- Added an empty-page fallback and null-safe rendering to prevent a successful API response from causing a frontend runtime 500.
- Validation: source review and `git diff --check` only; not pushed yet.

### 2026-09-09 Fix Waffo Pancake unit-price change detection

- Fixed the payment settings sanitizer to include `WaffoPancakeUnitPrice`; changing `1` to `1.05` now enters the save path instead of incorrectly reporting no changes.
- Validation: source review and `git diff --check` only; not pushed yet.

### 2026-09-09 Fix Waffo Pancake unit price label translation

- Added the exact translation key used by the Waffo Pancake unit-price field. The displayed `1.05` indicates the saved value matches the current form, so the no-changes message is expected.
- Validation: source review and `git diff --check` only; not pushed yet.

### 2026-09-09 Fix Passkey status with dashboard sessions

- Updated Passkey user resolution to accept the validated user ID from the dashboard authentication middleware, while retaining legacy cookie-session compatibility.
- This prevents the profile page from reporting Passkey status failure or session expiry after normal password login; Passkey remains optional.
- Validation: source review and `git diff --check` only; not pushed yet.

### 2026-09-09 Fix Waffo Pancake unit price persistence

- Fixed the configuration save payload and backend persistence path to include `WaffoPancakeUnitPrice`; values such as `1.05` are now saved instead of reverting to the default `1`.
- Preserved the existing checkout pricing behavior and added the change to the source-level synchronization log.
- Validation: source review and `git diff --check` only; not pushed yet.

### 2026-09-09 Fix audit log page endpoint and translations

- Corrected the audit page request from `/api/audit/` to the registered `/api/audit` route.
- Added missing English and Chinese translation entries for the standalone audit page.
- Validation: source review and `git diff --check` only; not pushed yet.

### 2026-09-09 Fix dashboard login identity header

- Fixed login and refresh success handling so the returned user ID is persisted as `uid`, allowing subsequent requests to send the required `New-Api-User` header.
- This addresses the observed `login 200` followed by `self 401` error without changing authentication or billing logic.
- Validation: source review and `git diff --check` only; not pushed yet.

### 2026-09-09 Sync upstream `9bf328d97`: preserve Sora provider fields

- Integrated the safe portion of the upstream Sora rendering fix: persisted provider response fields are returned when available, and the host preserves extensions while overriding public identity and lifecycle fields.
- Deliberately did not import the unrelated task-adaptor API refactor because it overlaps local Sora/audio billing, polling, plugin state, and channel compatibility behavior.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication.
- Local functionality deliberately preserved: Sora per-request billing, audio surcharge, task polling, plugin center, and provider compatibility.
- Local commit/push: pending on `main`.

### 2026-09-08 Model/vendor/pricing compatibility round 4 final review

- Reviewed the database and serialization boundary for the upstream model/vendor/pricing rewrite.
- Local model, vendor, and pricing data use GORM-compatible fields and migrations. No PostgreSQL-only JSONB, database-specific function, or destructive schema change is required for this compatibility round.
- The upstream rewrite remains intentionally not cherry-picked because it would replace local Sora/audio billing, dynamic pricing, customer discounts, plugin-center behavior, and wallet/payment presentation.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication.
- Result: the compatible model/vendor/pricing functionality is already present locally; remaining differences are architectural upstream refactors, not required missing behavior.
- Local commit/push: pending on `main`.

### 2026-09-08 Model management compatibility round 3

- Rechecked upstream model-management changes against the local model metadata, vendor management, model synchronization, and pricing refresh paths.
- The local implementation already supports vendor IDs/metadata, endpoint and modality metadata, model visibility, matching rules, channel-derived models, pricing refresh, and audit-aware model deletion. The upstream rewrite is therefore not applied wholesale.
- No safe missing behavior was found that can be added without changing local billing or plugin behavior.
- Validation: source review and `git diff --check` only; no local build or tests.
- Local commit/push: pending on `main`.

### 2026-09-08 Model/vendor/pricing compatibility round 2

- Compared the upstream pricing controller and model metadata API with the local implementation.
- The local `/api/pricing` already exposes model metadata, vendor data, endpoint capabilities, pricing version, billing mode/expression, group ratios, and usable groups. No missing non-breaking API field was identified.
- No source overwrite was performed because the upstream controller would replace local Sora/audio pricing and customer/group discount behavior.
- Validation: source review and `git diff --check` only; no local build or tests.
- Local functionality deliberately preserved: model/vendor management, pricing snapshots, Sora/audio billing, dynamic billing, group/customer discounts, and plugin-center behavior.
- Local commit/push: pending on `main`.

### 2026-09-08 Model/vendor/pricing refactor compatibility round 1

- Reviewed upstream `0c76e4dae`, `75e533209`, `0e0ba152b`, and `71c1fd7ca`.
- The local branch already contains compatible model/vendor entities, pricing snapshots, site-currency display, dynamic billing, Sora per-request pricing, audio surcharge, group/customer discounts, and model management APIs.
- The upstream rewrite was not copied wholesale because it would overwrite those local behaviors. This round records the compatibility boundary; implementation of dual-read/dual-write migration remains a separate controlled round.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication.
- Local functionality deliberately preserved: all local billing, plugin-center, wallet/payment, mail, and frontend customizations.
- Local commit/push: pending on `main`.

### 2026-09-08 (final review of upstream batch `ea7cb0ba4`)

- Final review completed for the upstream range `bee45b58a..ea7cb0ba4`.
- Integrated: the safe backend redemption batch deletion endpoint from `524455fac`. Already present locally: usage-log group filtering/mobile layout, API-key and quota display behavior, plugin metadata/icon support, plugin marketplace/channel management, site-currency display, model/pricing editors, Sora/audio pricing, group-model ratios, and customer-specific discounts.
- Deliberately deferred: legacy `web`-only UI rewrites that do not map directly to `web/default`; model/vendor pricing rewrite portions that would conflict with local pricing; Go convention/document-only commits; and any change that would remove the administrator plugin switch or alter local billing behavior.
- No remaining safe source change was identified in this range. Local billing, payment, mail, plugin, Sora/audio, customer discount, and frontend customizations remain preserved.
- Validation: source/history review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication. GitHub Actions is the verification gate.
- Local commit/push: pending on `main`.

### 2026-09-08 (redemption batch deletion)

- Integrated the safe backend portion of upstream `524455fac`: administrators can delete up to 1000 selected redemption codes through `POST /api/redemption/batch`, with input validation and management-audit recording.
- File export and legacy frontend changes remain deferred for a separate `web/default` adaptation. Existing redemption crediting, wallet, payment, billing, mail, plugin, and frontend custom behavior remains unchanged.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication. GitHub Actions remains the verification gate.
- Local commit/push: pending on `main`.

### 2026-09-08 (final upstream backlog reconciliation)

- Reconciled the remaining low-scope upstream items: OpenAI generation/capability handling, Responses cached-token billing, expression-pricing support, database migration compatibility, and task-plugin diagnostics are already represented in the local branch.
- Deliberately kept `/messages/count_tokens` enabled because the upstream temporary disable would regress existing Claude-compatible clients. The custom plugin switch and large model/pricing rewrite remain intentional local divergences documented in earlier entries.
- No source change was required. Local Sora/audio billing, customer discounts, wallet/payment display, mail templates, plugin center, and frontend customizations remain preserved.
- Validation: source/history review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication. GitHub Actions remains the verification gate.
- Local commit/push: pending on `main`.

### 2026-09-08 (model and pricing batch assessment)

- Assessed upstream `0c76e4dae` model/vendor/pricing rewrite and its dependent pricing API/frontend changes. The rewrite requires new model-pricing snapshots, vendor metadata, storage contracts, and editor UI that are not present as an isolated compatible layer locally.
- Deliberately deferred this batch rather than introducing an unusable endpoint or bypassing OmniRouters pricing behavior. Existing group ratios, customer-specific discounts, Sora per-request pricing, `audio_generation` surcharge, wallet/payment display, and billing-expression safeguards remain authoritative.
- Validation: source/dependency review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication. GitHub Actions remains the verification gate.
- Local commit/push: pending on `main`.

### 2026-09-08 (model and pricing batch assessment)

- Assessed upstream `0c76e4dae` model/vendor/pricing rewrite and its dependent pricing API/frontend changes. The rewrite requires a new model-pricing snapshot, vendor metadata, storage contracts, and editor UI that are not present as an isolated compatible layer locally.
- Deliberately deferred this batch rather than introducing an unusable endpoint or bypassing OmniRouters pricing behavior. Existing group ratios, customer-specific discounts, Sora per-request pricing, `audio_generation` surcharge, wallet/payment display, and billing-expression safeguards remain authoritative.
- Validation: source/dependency review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication. GitHub Actions remains the verification gate.
- Local commit/push: pending on `main`.

### 2026-09-08 (task plugin and polling batch reconciliation)

- Rechecked the task-plugin and polling changes through `bee45b58a`. The local branch already contains the upstream polling contract hardening, bounded failures, status classification, CAS terminal handling, plugin state persistence, Hailuo/MiniMax usage accounting, and settlement/refund integration.
- The upstream custom-plugin switch removal (`210734bb7`) remains intentionally skipped because OmniRouters exposes that switch in its administrator plugin center. Removing it would change a local operational control rather than provide a compatible fix.
- No task source change was required in this reconciliation. Sora per-request billing, `audio_generation` surcharge, plugin billing, pre-consume, settlement, refunds, payment, mail, and frontend customizations remain preserved.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication. GitHub Actions remains the verification gate.
- Local commit/push: pending on `main`.
- Commit/push: pending.

## 2026-09-07 (first-round database audit)

- Reviewed upstream `1751f43ee`, `6eb6f35ed`, `69a41eead`, `2b6f1dfef`, and `27ff6a876` against the current database layer.
- Already present locally: SQLite WAL, pragma busy timeout, immediate transaction locking, JSON `[]byte`/`string` scanning, `prefill_groups` constraint migration, and Token Key uniqueness migration. These were retained because the local implementations include existing cross-database and project-specific safeguards.
- No upstream database code was copied in this audit. PostgreSQL pooler handling and migration behavior remain under targeted verification before the next batch.
- Validation: source comparison and `git diff --check`; real PostgreSQL/MySQL runtime tests remain unavailable in this environment.
- Commit/push: pending.

## 2026-09-07 (OpenAI model generation compatibility)

- Integrated the safe portion of upstream `49ec46966`: OpenAI GPT generation detection now recognizes `gpt-5` and later major generations, including `gpt-6-astra`, for the existing max-completion-token, sampling-parameter and developer-role compatibility path.
- Preserved: reasoning suffix parsing, channel-specific conversion, Sora/audio pricing, plugin flows and billing.
- Deferred: upstream's larger capability matrix and extensive tests require model-by-model validation against our custom conversion behavior.
- Validation: `git diff --check`; Go tests unavailable because the local Go toolchain is not installed.
- Commit/push: pending.

## 2026-09-07 (Ali image response format)

- Integrated upstream `0bee5d441`: Ali image responses now read `response_format` from the validated image request stored in `RelayInfo`, so `url` and `b64_json` are honored consistently.
- Preserved: image quantity, Sora/audio pricing, plugin task flows, payment behavior, and all quota settlement paths.
- Already present: upstream response-header timeout, transport safeguards, and the broader plugin task lifecycle changes were compared and retained in their local implementations.
- Validation: `git diff --check`; Go tests unavailable because the local Go toolchain is not installed.
- Commit/push: pending.

## 2026-09-07 (wallet fee display adjustment)

- Local request: remove the inline fee preview beside the custom amount and show the fee-inclusive amount in the confirmation dialog.
- Changed frontend display only: confirmation `You Pay` now shows base amount plus the displayed fee; no backend order, credit, or billing behavior changed.
- Validation: `git diff --check`; frontend build unavailable locally.
- Commit/push: pending.

## 2026-09-07 (Waffo Pancake unit price setting)

- Added the existing `WaffoPancakeUnitPrice` option to the Waffo Pancake admin section; `1.05` makes a `$10` recharge quote `$10.50`.
- Preserved store/product binding, checkout callbacks, settlement and quota credit behavior.
- Validation: `git diff --check`; frontend build unavailable locally.
- Commit/push: pending.

## 2026-09-06 (Pancake USD quote display)

- Local bug: Pancake returns USD quotes but the shared wallet formatter divided them by the configured exchange rate again, displaying a USD 10 quote as approximately USD 1.39.
- Added provider-aware quote formatting for wallet quoted amounts and confirmation amounts/discounts/display fees. Pancake uses its checkout currency USD directly with two decimals; other providers retain existing formatting. Preset estimates retain their existing base-currency formatting.
- Preserved: quote APIs, order creation, payment totals, credits, callbacks and all backend billing. Combined with pending display-only fee changes, no fee is added to the quoted total.
- Validation: targeted Node formatter regression checks and git diff --check; full frontend build/browser verification unavailable locally (dependencies absent).
- Commit/push: integrated in `d6dc48aaa`, pushed to `origin/main` on 2026-09-06. This status update follows in a documentation-only commit.

## 2026-09-06 (per-payment-method display fees)

- Local request: configure wallet display fees independently for payment methods without changing checkout charges or credited quota.
- Added an optional 0-100 percent field to the payment-method editor and a fee column; stored as a string ratio in existing PayMethods JSON, preserving the backend map-of-strings contract. Applies to the configurable Epay, Stripe and Waffo Pancake entries, not separate Creem product or legacy Waffo editors.
- Wallet payment buttons and confirmation resolve the method rate before the legacy global fallback. Explicit zero disables the display; omitted values retain the global fallback. Invalid rates are suppressed on display. Payment totals now show the backend quote without adding the informational fee.
- Preserved: all backend payment handlers, callbacks, discounts, order amounts, credited quota, and corporate transfer behavior. No database migration.
- Validation: diff whitespace check, JSON parsing and targeted Node checks of rate normalization; full frontend typecheck and browser checks unavailable because local frontend dependencies are absent.
- Commit/push: integrated in `d6dc48aaa`, pushed to `origin/main` on 2026-09-06. This status update follows in a documentation-only commit.

## 2026-09-05 (notification timeline alignment)

- Local reason: align the timeline with the date column in the supplied reference.
- Changed: `web/default/src/components/notification-popover.tsx` places the dashed connector behind centered date badges; undated entries use a dot on the same axis. Removed duplicate title-side dots and the content border, and corrected the rich-content title wrapper to a div.
- Preserved: announcement content, expansion controls, read state, and backend contracts.
- Validation: `git diff --check`; no browser visual verification or frontend build performed locally.
- Local commit and push status: pending.

## 2026-09-05 (announcement backend length correction)

- Local reason: saving announcements still failed because the earlier inspection missed the backend 500-character check.
- Changed: `setting/console_setting/validation.go` now accepts up to 5000 UTF-16 code units, matching the default frontend, and reports the correct limit. Added boundary regression tests in `announcement_validation_test.go` for ASCII, Chinese, and supplementary Unicode characters.
- Preserved: announcement count, date/type validation, extra-field limits, storage format, and all billing behavior.
- Validation: `git diff --check` passed; Go and gofmt are unavailable locally, so the regression tests could not be executed here.
- Local commit and push status: pending.

## 2026-09-05 (announcement content length)

- Local reason: the previous 500-character announcement limit was too restrictive for Markdown content, image URLs, and detailed notices.
- Changed: raised the default frontend announcement content limit to 5000 characters and updated the form guidance in all maintained frontend locales. The backend 500-character limit was mistakenly overlooked in this change; corrected in the follow-up entry above.
- Validation: all maintained locale files pass `JSON.parse`, and `git diff --check` passes; frontend build/type validation was not run because Bun is unavailable locally.
- Local commit and push status: pending.

## 2026-09-05 (notification timeline default state)

- Local reason: correct the notification timeline interaction after visual review.
- Changed: notifications and announcements now start expanded; the unrelated external-link icon was removed from the expand/collapse control. Existing content, read state, and data contracts remain unchanged.
- Validation: pending; frontend build/type validation may be unavailable locally if Bun is not installed.
- Local commit and push status: pending.

## 2026-09-05 (notification timeline display)

- Local reason: align the notification popover with the dashboard announcement timeline presentation requested by the user.
- Changed: `web/default/src/components/notification-popover.tsx` now uses date labels, timeline nodes, dashed connectors, title/body previews, and expand/collapse controls for notices and announcements. Existing notification tabs, read-state inputs, rich-content rendering, and data contracts are preserved.
- Validation: pending; frontend build/type validation may be unavailable locally if Bun is not installed.
- Local commit and push status: pending.

## 2026-09-05 (announcement timeline display)

- Local reason: update the dashboard announcement presentation to a compact product-update timeline based on the requested reference layout.
- Changed: `web/default/src/features/dashboard/components/overview/announcements-panel.tsx` now displays date labels, dashed timeline nodes, announcement titles/body previews, and a detail affordance; the existing announcement data shape, detail dialog, settings editor, and click behavior are preserved.
- Validation: `git diff --check`; frontend build/type validation was not run because the local Bun toolchain is unavailable.
- Local commit and push status: pending.

## 2026-09-04 (remaining upstream audit closeout)

- Audit scope: upstream history through `32c261923a9786c64d2af087327ef057e7bde7e3`, with emphasis on `0ed497f06`, `bbd97446c`, `057f71c23`, `9f506dd7f`, `219c9e063`, `aece11d2f`, and `73afad588`.
- Integrated or already present: hosted-tool conversion and billing integrity are covered by the local conversion/usage commits (`b250536ca`, `0f884cfa9`, `eb85ad878` and existing relaykit changes); log privacy and projection are covered by `eb85ad878`; task lock and quota-stat fixes are already present; MiniMax-H3 and H3 media usage were integrated in `e2d2cf65f`; public-content ETag behavior and plugin no-channel diagnostics were integrated in `5b9c39048`.
- Intentionally skipped: upstream anonymous session-hint cookies and route-session revalidation do not map to this branch's local auth architecture, which uses the existing localStorage-based frontend auth flow and does not expose the upstream refresh-session helpers. Importing the hint alone would be inert and misleading.
- Preserved: all OmniRouters Sora/audio pricing, dynamic/group pricing, task plugin polling and billing, mail/wallet/payment features, model marketplace, and log visibility rules.
- Validation: `node --check plugins/tasks/hailuo/plugin.js`, `git diff --check`, static symbol checks, and staged-file review pass. Go/gofmt/Bun are unavailable locally; GitHub Actions must provide backend/frontend build and test verification.
- Local commit and push status: pending.

## 2026-09-04 (MiniMax-H3 task plugin follow-up)

- Upstream source reviewed: `aece11d2f` and `73afad588`.
- Integrated: the Hailuo task plugin now supports `MiniMax-H3`, including multimodal content validation, frame/reference media limits, 768P/2K resolution, 4-15 second duration, ratio handling, `/v2/video_generation` submission and polling, H3 result/error envelopes, public CDN artifacts, and resolution-based usage metadata. Input media accounting from `73afad588` was included.
- Preserved: the local task-plugin polling context, normalized actions, bounded failure handling, task identity mapping, and OmniRouters Sora/Suno/plugin billing paths were retained; upstream host/adaptor hunks were not allowed to replace those newer local implementations.
- Intentionally skipped: the upstream anonymous session-hint optimization because this branch uses the existing local authentication architecture without the upstream refresh-session flow; importing only that cookie would not be meaningful. Broad protocol-conversion and unrelated upstream changes remain separate audit items.
- Validation: `node --check plugins/tasks/hailuo/plugin.js` and `git diff --check` pass. Go/gofmt/Bun are unavailable locally; production build and backend tests remain for GitHub Actions.
- Local commit and push status: pending.

## 2026-09-04 (upstream plugin error, ETag, and release version fixes)

- Upstream source reviewed: `32c261923a9786c64d2af087327ef057e7bde7e3`, `8f5ab8e4048a90d88b20ae1e6d5228b04233d3b8`, and `36dbbf0f77e710455e745048f4a32e8120ad3fd2`.
- Integrated: plugin-claimed models now return a localized 503 message naming the claiming plugin when no enabled channel is available; public Notice/About/legal/home content endpoints now use stable weak ETags with RFC weak matching while preserving the existing `{success,message,data}` response envelope; Release workflow version resolution now prefers the triggering tag and only falls back to version tags for manual runs.
- Deliberately adapted: the upstream ETag implementation was made compatible with the current local JSON wrapper and Go toolchain; only the version-resolution portion of the CI commit was copied, preserving OmniRouters' existing default and classic frontend release builds and action pins.
- Already present: plugin routing, plugin channel binding, local billing/plugin behavior, and the existing public content endpoints were retained rather than replaced.
- Deferred/skipped: unrelated upstream CI action upgrades, frontend build restructuring, compliance-message changes, and broad upstream refactors were not included.
- Preserved: local Sora/audio billing, dynamic/group pricing, mail settings/templates, wallet/payment behavior, model marketplace, task plugins, and all other OmniRouters customizations.
- Validation: `git diff --check` passes; Go/gofmt are unavailable locally, so Go tests and the production build remain for GitHub Actions; workflow structure and changed references were inspected manually. Untracked local helper/report files were excluded.
- Local commit and push status: pending.

## 2026-09-03 (plugin polling compatibility complete)

- Upstream source reviewed: `9df450fe54e1a874a5339b7c38a61014217f02c3`.
- Integrated: `9df450fe54e1a874a5339b7c38a61014217f02c3` plugin polling contract, including persisted state, real query contexts, batch task contexts, HTTP response metadata, `UNKNOWN` status handling, bounded consecutive poll failures, CAS-protected terminal transitions, and refund reconciliation.
- Follow-up hardening: restored the upstream canonical task-action vocabulary through `constant.NormalizeTaskAction` while retaining legacy native action values, counted successful batch responses that omit a requested task as one unresolved poll failure, and kept the legacy adapter bridge from emitting a literal `<nil>` task ID.
- Integrated: the ten built-in task plugins were updated for the new context and explicit unknown-state behavior; Jimeng persists `req_key`, Suno batch polling receives task contexts, and Google/Vertex preserve missing-`done` in-progress semantics.
- Already present and retained: PostgreSQL JSON string writes, string/byte JSON scanning compatibility, refund-pending markers, atomic refund claims, and local task billing protections were kept as the surrounding implementation rather than overwritten by upstream variants.
- Intentionally skipped: upstream-only broad refactors and unrelated authentication, ticket, frontend, and CI changes from the upstream branch are outside this sync boundary.
- Preserved: local Sora per-call billing, fixed `audio_generation` surcharge, dynamic/group pricing, plugin billing, pre-consume/settlement/refunds, quota-saturation auditing, Suno batch polling, and all native task adaptors.
- Validation: `git diff --check` passes; all embedded task plugin files pass `node --check`; canonical action and batch omission regression coverage was added. Go, `gofmt`, and Bun are unavailable in this local environment, so Go tests and the production build must be verified by GitHub Actions. Existing local test history includes the SQLite fixture limitation (`tickets` table is not created) in `go test ./service`.
- Local commit and push status: integrated in `f5735642d` and pushed to `origin/main`; the log-only status correction follows in the next commit.

## 2026-09-03 (database compatibility follow-up)

- Upstream source reviewed: `1751f43ee`, `66031a09d`, `6eb6f35ed`, with current upstream `main` at `9df450fe54e1a874a5339b7c38a61014217f02c3`.
- Integrated: SQLite WAL/busy-timeout/`_txlock=immediate` behavior was already present; PostgreSQL GORM prepared statements are disabled for transaction-pooling proxies; JSON column Valuers return strings for PostgreSQL simple protocol; JSON Scanners accept both `[]byte` and `string` while preserving NULL/empty semantics; PostgreSQL prepared-statement SQLSTATE errors now include an actionable diagnostic.
- Already present: shared JSON marshal/unmarshal wrappers, `StringList`, `SubscriptionGroupList`, and existing task JSON fields already use the required cross-database patterns.
- Deferred: `9df450fe5` task/plugin polling contract changes, legacy Token Key migration, and `prefill_groups` uniqueness migration remain separate batches because they touch polling behavior, authentication data, or schema migration semantics.
- Preserved: local Sora per-call billing, fixed `audio_generation` surcharge, dynamic/group pricing, plugin billing, refunds, quota-saturation auditing, usage logs, mail settings, and other OmniRouters customizations.
- Validation: `gofmt` and `git diff --check` are required for this batch. Targeted model tests remain blocked by pre-existing test compilation errors (`useUserCacheMiniRedis`, `AuthVersion`, and `UpdateUserAccessToken` symbols); production build result is recorded with the commit.
- Local commit and push status: pending.

## 2026-09-01 (Sora pricing detail display)

- Local reason: simplify Sora per-second pricing in the model detail view.
- Changed `web/default/src/features/pricing/components/model-details.tsx`: removed the base per-second price and resolution multiplier labels, leaving only final resolution prices; renamed the group price column to `Price`.
- Preserved: all billing calculation and stored pricing configuration behavior.
- Validation: `git diff --check`; Go/Bun tests not run because the local toolchains are unavailable.
- Validation: `git diff --check`; Go/Bun tests not run because the local toolchains were unavailable.
- Local commit and push status: this local change predates the current upstream-sync tracking entries.

## 2026-09-02 (billing settlement, phase C review)

- Upstream source reviewed: `0ed497f06`, billing and settlement-related portions.
- Result: no additional code was integrated in this phase. The current code already has `BillingSession` pre-consume/settle/refund orchestration, tiered-expression actual-usage settlement, quota saturation protection, and local Sora per-call plus fixed `audio_generation` surcharge handling.
- Deliberately skipped upstream changes that would alter local semantics: removing group-model ratios, changing quota-notification links/templates, changing subscription-to-wallet fallback rules, and the broad task/plugin billing rewrite. These require a separate compatibility design and must not overwrite OmniRouters billing behavior.
- Validation: existing targeted billing, tiered settlement, task refund, and quota safety tests were reviewed; no source changes were made in this review-only phase.
- Local commit and push status: no new code commit; the review record is included in `9c1badd67`, which is pushed to `origin/main`.

## 2026-09-02 (reasoning core, phase D review)

- Upstream source reviewed: `0ed497f06`, reasoning normalization and model-suffix handling.
- Integrated the protocol-independent reasoning intent/rendering core, suffix parsing, conversion-state DTO, and focused upstream reasoning tests. Existing legacy `TrimEffortSuffix` behavior remains available for current converters.
- Deliberately did not integrate the host helper that depends on upstream-only `RelayInfo.ReasoningConversion` and `ConvOptions.PreserveEffortTail`; it requires the broader conversion-host refactor. Responses stream conversion, Claude/Gemini host conversion, and hosted tools remain separate follow-up stages.
- Preserved local Sora per-call pricing, fixed `audio_generation` surcharge, dynamic pricing, plugin billing, refunds, and usage-log behavior.
- Validation: reasoning package tests, OpenAI Responses/OAI chat converter tests, targeted channel tests, production Go build, and `git diff --check` passed after excluding the host helper.
- Local commit: `771434efe` (`sync upstream reasoning core phase D`); remote `origin/main` now points to this commit. The push command reported a transient ref-lock race, but remote verification confirmed the commit is present.

## 2026-09-02 (OpenAI Responses to Chat conversion, phase E)

- Upstream source: `0ed497f06`, Responses-to-Chat request/response and streaming conversion improvements.
- Integrated reasoning-aware Responses request mapping, response reasoning/citation preservation, sparse stream event handling, annotation propagation, and the required Responses/Chat DTO fields.
- Deliberately deferred Responses-to-Claude/Gemini target conversion and Hosted Tools/toolconv. No billing, task plugin, Sora, or `audio_generation` behavior was changed.
- Validation: `relaykit/relayconvert/internal/oai_responses` tests, OpenAI Responses/channel tests, production Go build, and `git diff --check` passed.
- Local commit: `bafa5f80e` (`sync upstream Responses to Chat conversion`); remote verification confirms it is present on `origin/main` despite a transient `unexpected eof` reported after upload.

## 2026-09-02 (Responses to Claude conversion, phase F)

- Upstream source: `0ed497f06`, OpenAI Responses to Claude Messages conversion.
- Integrated direct Claude request conversion with portable reasoning state, Claude usage mapping, ordinary response conversion, streaming response state handling, citations, and the required host conversion metadata/DTO fields.
- Updated compatibility tests and the response golden fixture for preserved reasoning content. Hosted Tools/toolconv and Responses-to-Gemini remain separate follow-up work.
- Preserved local Sora per-call pricing, fixed `audio_generation` surcharge, dynamic pricing, plugin billing, refunds, and usage-log behavior.
- Validation: `relaykit` conversion tests, relay metadata tests, targeted channel tests, production Go build, and `git diff --check` passed.
- Local commits: `0f884cfa9` (`sync upstream Responses to Claude conversion`) and `30a12c89f` (`fix include Claude conversion dependencies`); remote verification confirms both are present on `origin/main` despite transient TLS/ref-lock messages.

## 2026-09-02 (Responses to Gemini request conversion, phase G)

- Upstream source: `0ed497f06`, OpenAI Responses to Gemini request conversion.
- Integrated reasoning-aware thinking configuration, explicit zero-value preservation for `top_p` and `max_output_tokens`, function-call response IDs, and the required Gemini DTO compatibility field.
- Deliberately deferred Gemini grounding/hosted-tool response conversion and toolconv. No billing, task plugin, Sora, or `audio_generation` behavior was changed.
- Validation: Responses converter tests, Gemini/OpenAI targeted channel tests, production Go build, and `git diff --check` passed.
- Local commit: `9283bd5fa` (`sync upstream Responses to Gemini request conversion`); remote verification confirms it is present on `origin/main` despite transient TLS/ref-lock messages.

This file records the upstream `QuantumNous/new-api` commit that has been reviewed or integrated into this repository.

## 2026-09-03 (upstream log privacy and billing follow-up)

- Upstream sources reviewed: `057f71c23` and `bbd97446c`, plus dependencies from the already integrated `0ed497f06`.
- Integrated scoped log metadata with user/admin/root visibility, conversion diagnostics, hosted-tool task metadata, and related usage-log UI/test updates. Preserved local fixed-quota/Sora/audio_generation billing fields while adapting writers to `LogOther`.
- Integrated follow-up task billing settlement behavior, including tiered task usage facts and per-call billing semantics; added the required compatible snapshot/relay metadata fields.
- Deliberately did not restore the upstream deletion of the legacy log cleanup handler and did not modify local billing formulas or refund ownership. The Hailuo/MiniMax-H3 plugin batch and database/stability batch remain separate planned stages.
- Preserved: local Sora per-call pricing, fixed `audio_generation` surcharge, dynamic pricing, task plugins, refunds, quota saturation auditing, and usage-log quota statistics. Reapplied the local `type=0` all-types sentinel and separate RPM/TPM scan after upstream merge.
- Validation: production Go build, `go test ./service ./relay -run '^$' -count=1`, focused task-settlement tests, `go test ./pkg/billingexpr ./relay/helper -count=1`, and `git diff --check` passed. The full `service` suite also exercises unrelated existing failures: missing `tickets` table in the test database and pre-existing task test fixture expectations; these are not part of this sync.
- Local commit and push status: pending commit and push.

## 2026-09-02 (usage log quota statistic preservation)

- Local reason: usage-log header showed `$0` even though individual consume logs had non-zero charges.
- Changed `model/log.go`: scan RPM/TPM into a separate result structure so the second aggregate query cannot overwrite the quota returned by the first query.
- Preserved: all-type and consume-only filtering, recent RPM/TPM window, billing writes, and quota units.
- Validation: `git diff --check`; production build and package tests pending.
- Local commit and push status: pending validation.

## 2026-09-02 (usage log statistics type sentinel)

- Local reason: usage log header statistics were always zero when the frontend selected all log types.
- Changed `model/log.go`: `SumUsedQuota` now treats `type=0` as the all-types sentinel, matching the usage-log list API; consume-only filtering remains applied when `type=2` is explicitly requested.
- Preserved: quota, RPM, and TPM aggregation filters and all billing/log write behavior.
- Validation: `git diff --check` passed; model package tests could not compile because the existing test suite references unrelated missing helpers/fields (`useUserCacheMiniRedis`, `AuthVersion`, and `UpdateUserAccessToken`).
- Local commit and push status: pending validation.

## 2026-09-02 (Hosted Tools/toolconv follow-up)

- Upstream source: `0ed497f06`, hosted tool conversion and protocol DTO dependencies.
- Integrated the portable `relaykit` tool conversion package, conversion-loss diagnostics/policy, hosted-tool response artifact handling, web-search action normalization, Gemini grounding metadata compatibility, Claude server-tool fields, and Responses hosted-tool output fields.
- Updated the Responses-to-Gemini golden fixture to preserve function-call IDs in function calls and results.
- Deliberately did not alter the host routing/relay selection layer or local task-plugin system; this stage supplies and tests protocol conversion primitives only.
- Preserved local Sora per-call pricing, fixed `audio_generation` surcharge, dynamic pricing, plugin billing, refunds, and usage-log behavior.
- Validation: `go test ./relayconvert/internal/toolconv -count=1` and all relayconvert packages except the stale golden snapshot passed; after updating the expected upstream ID behavior, the focused conversion suite and `git diff --check` are the remaining final checks.
- Local commit and push status: pending final validation and commit.

## 2026-09-02 (usage stream merging, phase B)

- Upstream source: `0ed497f06`, selectively integrated from the stream-usage portion.
- Integrated `relaykit/dto/usage_merge.go` and focused tests so partial OpenAI, Claude, and Gemini usage snapshots retain earlier non-zero fields; provider-native billing dialects are preserved when snapshots are merged.
- Updated OpenAI Chat final-frame handling, OpenAI Responses completion-event handling, and Gemini streaming metadata accumulation to merge cumulative usage instead of allowing a later sparse frame to erase token/cache/modality data.
- Deliberately deferred the same commit's hosted-tool conversion, reasoning/Responses protocol rewrites, broader Claude/Gemini conversion changes, and billing pre-consume/settlement changes. Existing Sora per-call pricing, fixed `audio_generation` surcharge, dynamic pricing, plugin billing, refunds, and quota protections remain unchanged.
- Validation: `go test ./dto` in `relaykit`, targeted usage tests for Gemini/OpenAI/Claude, production Go build, and `git diff --check` passed.
- Local commit: `9e05339da` (`sync upstream stream usage merging phase B`), included in `9c1badd67`; pushed successfully to `origin/main` after switching the Git connection to OpenSSL.

## 2026-09-02 (reasoning core, phase D)

- Upstream source: `0ed497f06`, reasoning normalization and model suffix support.
- Integrated the protocol-independent reasoning intent/rendering core for Claude and Gemini, model suffix parsing, in-process reasoning state on OpenAI request DTOs, and compatibility for existing OpenAI suffix callers.
- Deliberately deferred the host helper until the surrounding host conversion state was available; it is now represented by the compatible `RelayInfo`/`convmeta` state additions in the Claude conversion phase. Full Hosted Tools/toolconv remains separate.
- Preserved local Sora per-call pricing, fixed `audio_generation` surcharge, dynamic pricing, plugin billing, refunds, and usage-log behavior.
- Validation: all `relaykit` tests, targeted channel tests, production Go build, and `git diff --check` passed.
- Local commit: `771434efe` (`sync upstream reasoning core phase D`), pushed to `origin/main`.

## 2026-09-02 (usage normalization, phase A)

- Upstream source: `0ed497f06`, selectively integrated from the usage/billing portion only.
- Integrated provider-native usage canonicalization for OpenAI Chat/Responses, Claude Messages, and Gemini Chat in `relaykit/dto`; Gemini estimated usage now retains reasoning, cache, and modality details.
- Consolidated the service usage interpretation path onto `BillingUsage.CanonicalUsage()` and removed the duplicate conversion implementation.
- Deliberately did not integrate stream-wide usage merging, hosted-tool conversion, reasoning protocol rewrites, or any new pre-consume/settlement behavior. Existing Sora per-call pricing, fixed `audio_generation` surcharge, dynamic pricing, plugin billing, refunds, and quota protections remain authoritative.
- Validation: `go test ./dto` in the standalone `relaykit` module, targeted `service` tests, production Go build, and `git diff --check` passed.
- Local changes are prepared as phase A only; the remaining upstream commit stays split for later review.

## 2026-09-02 (system task no-op state update)

- Upstream source: `b7017c251` (`fix(model): do not treat no-op system task state writes as lock loss`).
- Integrated the lock confirmation fallback in `UpdateSystemTaskState`: when a database reports zero changed rows for an identical state payload, the task and unexpired lease are checked before returning `ErrSystemTaskLockLost`.
- Added a regression test for repeated identical state persistence and successful task completion.
- Preserved all local relay, billing, plugin, mail, branding, and usage-log audit behavior; no frontend or request path was changed.
- Validation: `gofmt`, `git diff --check`, and production Go build passed. The broader model test suite remains blocked by pre-existing missing test helpers/fields.
- Local commit: `26106919d` (`fix system task no-op state updates`).

## 2026-09-01 (usage-log request audit tabs)

- Local feature: expanded the existing usage-log detail dialog with `Basic Information`, administrator-only `Request Audit`, and administrator-only `Raw Data` tabs.
- The request audit summarizes existing request identifiers, caller identity, routing/model mapping, retry chain, timing, token usage, billing path, parameter overrides, and diagnostics. Raw data shows only the log object already returned by the backend.
- No request prompt, response body, or complete API key collection was added. Existing backend removal of `other.admin_info` for non-admin users remains authoritative, and non-admin users continue to see only the original detail view.
- Added translations for the new tabs in English, Simplified/Traditional Chinese, French, Japanese, Russian, and Vietnamese.
- Validation: locale JSON parsing, protected-header format check, and `git diff --check` passed. Full frontend build remains delegated to GitHub Actions because local frontend dependencies are not installed.
- Code commit: `b3467bb80b` (`feat: add request audit to usage log details`).

## 2026-09-01 (task quota error retry and logging)

- Local fix: task submission converted local billing errors into `TaskError` without retaining their local/skip-retry semantics. A wallet quota 403 was consequently treated as an upstream channel failure, recorded once per attempt, and retried through the same channel (`2 -> 2 -> 2 -> 2`).
- `TaskErrorFromAPIError` now marks pre-upstream API failures as local errors. Insufficient quota still returns HTTP 403 with code `insufficient_user_quota`, but task submission no longer retries it or records it as a channel error.
- Added a regression test covering status, message, error code, and local-error preservation.
- Preserved real upstream task retry behavior for 429, 307, and eligible 5xx responses, along with all billing/refund logic.
- Validation: targeted service regression test and `git diff --check` passed.
- Code commit: `81ea90fa4a` (`fix task quota errors retrying as channel failures`).

## 2026-09-01 (task usage-log route)

- Local fix: the default frontend requested admin task logs at `/api/task`, while Gin registers the admin collection route as `/api/task/`, producing a 404 for administrators.
- Updated the shared usage-log API path builder to add the collection trailing slash for admin requests; user requests continue using `/self`.
- Preserved task-log filters, model visibility rules, and all local billing/plugin behavior.
- Validation: source diff reviewed; frontend build was not run because the local frontend dependency setup is incomplete.

## 2026-08-31 (PostgreSQL compatibility and initialization dependencies)

- Upstream sources: `6eb6f35ed` (PostgreSQL JSON compatibility) and `74158715c` (database initialization dependencies); `98d50d538` was rechecked and was already integrated.
- Integrated:
  - JSON model values now write as strings so PostgreSQL does not interpret JSON bytes as `bytea`.
  - JSON scanners accept both `[]byte` and `string`, covering SQLite/MySQL/PostgreSQL driver return types.
  - Updated `ChannelInfo`, `JSONValue`, task `Properties`, and `TaskPrivateData` while retaining the project `common` JSON wrappers.
  - Updated the GORM MySQL/PostgreSQL driver versions required by the initialization fix.
- Preserved local billing, Sora/audio-generation pricing, plugin, mail, branding, and frontend customizations.
- Validation: production Go build passed; model tests remain blocked by pre-existing missing test helpers and fields. No live PostgreSQL DSN was available, so a real PostgreSQL migration test is still pending.
- Local commit: `7d6909995` (`fix PostgreSQL JSON field compatibility`), pushed to `origin/main`.

## 2026-08-30 (task plugin migration, phase 1)

- Upstream source: `eb48396d5` (`feat(task): replace built-in task adaptors with a sandboxed JS plugin system`)
- Integrated only the dormant foundation:
  - Added the sandbox engine, plugin metadata/registry, routing primitives, request helpers, and plugin DTOs.
  - Added the task-plugin channel type and protocol timeout configuration.
  - Added `TASK_PLUGIN_ENABLED`, defaulting to `false`.
- Deliberately retained the existing Go task adapters and production task routing. No Sora, Suno, Wan, async-image, billing, settlement, or refund path is switched to plugins in this phase.
- The upstream plugin UI, plugin routers, task-plugin adaptor, built-in JS task plugins, and task lifecycle replacement remain deferred until the host bridge is implemented and tested.
- Local protections remain authoritative: Sora per-request pricing, fixed `audio_generation` surcharge, checked quota accounting, async OpenAI image handling, and task refund reconciliation.
- Validation: `git diff --check` passed. Go/Bun are unavailable locally; GitHub Actions must perform dependency resolution, compilation, and tests.

## 2026-08-29

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous local sync point: `85db3cf85`
- Reviewed upstream `main` through: `918427d8a`
- Integrated selectively:
  - Invalid relay request parameters now return HTTP 400 and skip retry (`0f2a2075a`).
  - Time-based billing rules distinguish same-day ranges from overnight ranges and reject invalid time bounds (`ac381acf4`), preserving local Sora and `audio_generation` pricing.
  - Setup status is rechecked after a page refresh instead of trusting a persistent browser cache (`98d50d538`).
  - Usage-log sensitive filters disable credential autofill while retaining the local visibility toggle (`2d8e50bf3`).
  - Ollama can pass through native Claude and OpenAI Responses requests (`ba2e9287b`).
  - Zhipu supports the OpenAI Responses endpoint (`cae3676ec`).
  - Built-in OAuth binding keys use stable provider keys without changing the local custom OAuth layout (`692e8d6ee`).
- Already present in local adaptations:
  - Recharge quota pre-validation and atomic wallet-ceiling enforcement (`2a0ce3475`, `47ba9d2c6`).
  - Ali `top_p` omission and explicit-boundary clamping (`2399de97d`).
  - Native Claude/Gemini channel testing and Gemini streaming URL handling (`b941253ae`, `4add708eb`).
  - vLLM `thinking_token_budget` passthrough (`8f6961c67`).
- Deferred:
  - Password transport/encryption changes (`b80d633cf`, `918427d8a`) require a separate authentication compatibility review.
  - Sandboxed JavaScript task plugins (`eb48396d5`) require a dedicated migration because they affect Sora, Wan, async image tasks, and settlement/refund logic.
  - Upstream CI/workflow, Bun, test-infrastructure, and development-image-only changes were not copied over the local GitHub/GHCR workflow.
- Validation notes:
  - `git diff --check` passed. Go and Bun are unavailable locally, so compilation, tests, and image publishing remain delegated to GitHub Actions.
  - Existing untracked `bin/sync-upstream.ps1` and i18n report files remain excluded.

## 2026-08-15

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `823e26304a396854ace30b52b98ec497c2dd9c36`
- Reviewed and selectively integrated through upstream `main`: `e2c7aa7b102c2075eae2377df3508658d45e88dc`
- Integrated:
  - Async task refunds now reverse user and channel accumulated usage exactly once through the existing `refund_pending` and quota CAS path (`58d4e9bd3`).
  - Ali image protocol selection uses the mapped upstream model (`93d2df85f`).
  - Recharge requests reject uncreditable amounts before payment, and settlement atomically enforces the final wallet quota ceiling across Epay, Stripe, Creem, Waffo, Waffo Pancake, and manual completion (`2a0ce3475`, `47ba9d2c6`).
  - Channel fetched-model selection follows the current unsaved form state (`15cfdedde`).
  - OpenAI Chat to Responses conversion preserves `prompt_cache_key`, and Claude conversion omits empty tools (`7d09c6954`, `4442bb302`).
  - Updated DOMPurify to `3.4.13` and added reusable custom OAuth access-policy/message templates without changing the local unified OAuth binding layout (`f250f3b58`, adapted from `116255f07`).
- Already present:
  - Gateway channel field-passthrough controls from `e90a7c48e` were already implemented in `web/default`; no duplicate UI was added.
- Skipped:
  - Electron-only dependency updates (`626058075`, `53a8739ee`, `e5efc73cd`, `cf38105a9`, `bbf67df04`) because they do not affect the server or GHCR deployment.
  - The broad Vitest migration in `e2c7aa7b`; it changes test infrastructure without runtime benefit and is independent of the fixes above.
- Validation notes:
  - Targeted frontend lint and TypeScript typecheck passed using the existing dependencies.
  - `git diff --check`, call-signature scans, and conflict-marker scans passed. Go execution remains delegated to GitHub Actions because no local Go runtime is installed.
  - `bin/sync-upstream.ps1` and i18n temporary reports remain untracked and are excluded.

## 2026-08-10

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `0ab02020603d22e5613bc4cf46bfab06f8567769`
- Reviewed and selectively integrated through upstream `main`: `823e26304`
- Integrated:
  - Native Claude/Gemini channel test requests and Gemini streaming test URL handling (`b941253ae`).
  - Per-user critical rate limits for access-token generation and affiliate quota transfer (`1da23d6b3`), retaining local route protections.
  - Ali/Qwen `top_p` behavior: omitted values remain omitted; explicit boundary values are clamped (`2399de97d`).
  - Default frontend fetched-model categorization and Qwen TTS classification (`c9bc03864`, `823e26304`).
  - Redemption-code editable precision and stale-update protection, adapted to the local currency and drawer implementation (`e926e5cac`).
  - Replayable outbound request bodies with `GetBody`, independent memory/disk readers, redirect handling, and HTTP/2 stream-reset retry support (`d6b5ce99d`, `ea4f02101`).
- Local safety adaptations:
  - Kept Sora per-request pricing, fixed per-call `audio_generation` surcharge, quota saturation audit, async image request handling, and all existing RelayInfo billing fields.
  - Merged the new Sora pass-through replay test with local audio surcharge overflow tests.
  - Moved upstream frontend tests from `web/src` to the deployed `web/default` tree.
- Skipped:
  - `5c3abffe8`, because it only changes upstream GitCode release synchronization workflow and OmniRouters uses GitHub Actions/GHCR.
- Validation notes:
  - Checked replay metadata call sites and confirmed no stale `UpstreamRequestBodySize`/`UpstreamRequestGetBody` references remain.
  - `git diff --check` and conflict-marker scans passed. Go tests require the unavailable local Go toolchain; frontend dependency/build validation remains delegated to GitHub Actions.
  - `bin/sync-upstream.ps1` and i18n temporary reports remain untracked and are excluded.

## 2026-08-05

- Upstream repository: `https://github.com/QuantumNous/new-api.git`
- Previous upstream sync point: `afe16c64cd73853da1eda3bf236f15d69637b4bf`
- Upstream `main` reviewed and selectively synced through: `0ab02020603d22e5613bc4cf46bfab06f8567769`
- Integrated:
  - Configurable OIDC display name, Qwen `thinking_budget` zero-value passthrough, zstd request decoding, owner-visible stream status, multipart image-edit preservation, safe OAuth callback-mode detection, and DeepSeek Responses support.
  - Final-group tiered retry settlement and group-switch billing hardening, preserving local checked quota conversion, saturation auditing, refund reconciliation, Sora pricing, and fixed per-call `audio_generation` surcharge.
  - Bedrock client-disconnect cancellation and effective billing-path logging.
  - Per-API-key custom Auto group ordering across token persistence/cache, channel selection, model ownership, settings, editor UI, and tests.
- Local adaptations:
  - Auto group resolution combines the token-specific order with current base and extra-group permissions; it cannot bypass either permission source.
  - Ported active frontend changes to `web/default`, retained the local DataTable timestamp implementation and drawer layout, and kept the startup nil-adapter guard for intentionally disabled Advanced Custom.
  - Preserved existing locale dictionaries and added the new Chinese Auto-order UI strings without accepting upstream locale deletions.
- Skipped:
  - `c27d1ef65` (`.gitattributes`) and `aa7d0d39a` (public-nav font-size only), because neither changes required OmniRouters runtime behavior.
- Validation notes:
  - Source conflict-marker and whitespace checks passed. TypeScript errors introduced by the Auto group port were resolved; remaining local typecheck dependency errors require the frozen Bun install used by Actions.
  - Go is unavailable locally, so GitHub Actions remains responsible for Go compilation/tests and image publishing. No local Docker image was built.
  - `bin/sync-upstream.ps1` remains an untracked local helper and was not included.

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
## 2026-09-06

- Local change: enable the default quota warning threshold consistently.
- Changed `common.QuotaRemindThreshold` from `1000` to `500000`, matching the default frontend user setting; the existing notification type and per-user overrides remain unchanged.
- Preserved: pre-consume, settlement, refund, and insufficient-quota response behavior were not modified.
- Validation: source diff review and `git diff --check`; full Go/frontend test suites were not run because the local Go/Bun toolchains are unavailable.
- Local commit/push: implemented in commits `9da5fc10d`, `b030d2065`, `49a7960d6`, `989cf467a`, `c7252725d`, `442c80721`, `bb508aed0`, `3b67f3e83`; current `origin/main` is `156fc65b2`.
- Final current-upstream review: `upstream/main` remains `0c76e4dae`. No new independent third-round patch is available beyond the already reviewed structural migrations. The model/vendor/pricing rewrite and Telegram unified OAuth remain deferred pending a dedicated compatibility branch or staged port.
- Second-round status: token lifecycle and existing account-security success events are integrated; upstream's full session-bound verification and audit schema/UI suite is not claimed as fully equivalent.
- Validation: `git diff --check`; Go/Bun toolchains unavailable locally.
- Local commit/push: pending.
- Round 3 progress: added a separate Telegram OAuth start endpoint that stores PKCE flow data in the session and returns an authorization URL; legacy Telegram routes and generic OAuth state remain unchanged.
- Validation: `git diff --check`; Go toolchain unavailable locally.
- Local commit/push: pending.
- Round 3 progress: completed the OAuth registry interface scaffold for Telegram while keeping the provider disabled until callback/session integration is complete.
- Validation: `git diff --check`; Go toolchain unavailable locally.
- Local commit/push: pending.
- Round 3 progress: registered a disabled-by-default `telegram_oauth` provider adapter behind the existing registry; it cannot authenticate or replace the legacy Telegram route until session/callback integration is complete.
- Validation: `git diff --check`; Go toolchain unavailable locally.
- Local commit/push: pending.
- Round 3 progress: added Telegram ID Token verification through OIDC JWKS with issuer/client/signature/expiry validation delegated to the verifier; only validated identity claims are returned.
- Validation: `git diff --check`; Go toolchain unavailable locally.
- Local commit/push: pending.
- Round 3 progress: added bounded Telegram authorization-code exchange with PKCE verifier/config matching and ID-token presence checks; signature verification and route integration remain pending.
- Validation: `git diff --check`; Go toolchain unavailable locally.
- Local commit/push: pending.
- Round 3 progress: added isolated Telegram OAuth PKCE flow generation and authorization URL construction; provider registration, token exchange, ID-token verification, and route switching remain intentionally deferred.
- Validation: `git diff --check`; Go toolchain unavailable locally.
- Local commit/push: pending.
- Round 3 progress: added isolated Telegram OAuth client configuration with explicit readiness checks; legacy bot configuration and login flow remain unchanged.
- Validation: `git diff --check`; Go toolchain unavailable locally.
- Local commit/push: pending.
- Round 3 progress: added idempotent legacy Telegram binding backfill into external identity claims after schema migration; duplicate ownership fails rather than being silently reassigned.
- Validation: `git diff --check`; Go toolchain unavailable locally.
- Local commit/push: pending.
- Round 3 progress: added the isolated external-identity ownership table and atomic claim helper as the first Telegram OAuth migration layer; legacy `telegram_id` login remains unchanged and no OAuth flow was replaced.
- Validation: `git diff --check`; Go toolchain unavailable locally.
- Local commit/push: pending.
- Round 2 progress: successful email binding now emits a sanitized security audit event; email address and verification code are excluded.
- Validation: `git diff --check`; Go toolchain unavailable locally.
- Local commit/push: pending.
- Round 2 progress: successful 2FA enable and backup-code regeneration now emit sanitized security audit events; secrets remain excluded.
- Validation: `git diff --check`; Go toolchain unavailable locally.
- Local commit/push: pending.
- Round 2 progress: added success audit events for WeChat binding and administrator-triggered OAuth unbinding.
- Validation: `git diff --check`; Go toolchain unavailable locally.
- Local commit/push: pending.
- Round 2 progress: successful custom OAuth unbind and Telegram bind operations now emit security audit events without storing provider credentials or Telegram secrets.
- Validation: `git diff --check`; Go toolchain unavailable locally.
- Local commit/push: pending.
- Round 2 progress: successful self-service password changes now emit a security audit event; password values are never included.
- Validation: `git diff --check`; Go toolchain unavailable locally.
- Local commit/push: pending.
- Round 2 progress: successful account deletion and 2FA disable operations now emit sanitized security audit events.
- Validation: `git diff --check`; full Go tests unavailable locally.
- Local commit/push: pending.
- Round 3 assessment: `0c76e4dae` is a broad model/vendor/pricing and frontend structure rewrite; it is not safe to cherry-pick over OmniRouters' custom pricing, plugin, wallet, and default-frontend code. Telegram OAuth and account-security changes remain separately scoped for later compatibility review.
- Round 2 progress: successful personal access-token generation/rotation now records a sanitized security event alongside revoke and request events.
- Validation: `git diff --check`; full Go tests unavailable locally.
- Local commit/push: pending.
- Round 2 progress: authenticated bearer-token requests now create sanitized audit entries containing only token fingerprint, route, status, IP, and request ID.
- Validation: `git diff --check`; full Go tests unavailable locally.
- Local commit/push: pending.
- Round 2 progress: added admin/self audit query endpoints over the isolated audit table. No usage-log endpoint was changed.
- Validation: `git diff --check`; full Go tests unavailable locally.
- Local commit/push: pending.

- Round 2 progress: added access-token creation timestamp tracking and a user-scoped token revoke endpoint.
- Validation: `git diff --check`; full Go tests unavailable in the local environment.
- Local commits/push: `c25d3b484` and `d3637953c`, both pushed to `origin/main`.
- Round 2 progress: added isolated `AuditLog` storage with token fingerprints only, automatic migration registration, and revoke-event recording. Existing usage-log audit fields remain unchanged.
- Validation: `git diff --check`; Go tests unavailable locally.
- Local commit/push: pending.

## 2026-09-07

- Third-round migration branch: `codex/migrate-upstream-structural-v3`.
- Telegram migration dependency map: `external_identity_claim`, authorization-code/PKCE flow, session-bound verification, OAuth registry/provider, system Telegram settings, and route/frontend callback changes. Legacy `telegram_id` bindings must remain readable during migration.

- Upstream review: `d8cb17744`, `3f8a50cf8`, `6f2333990`, `0973dc2b8`, `a8729b5c3`, `45c3fbe8` and related commits through `upstream/main` `0c76e4dae`.
- Planned three-round sync:
  - Round 1: database compatibility, request stability, provider/model capability fixes, and billing safety.
  - Round 2: access-token lifecycle, security verification, and independent audit records.
  - Round 3: model/provider/pricing structural refactors and Telegram OAuth updates.
- Current status: audit completed; existing administrator operation audit and privacy projection are retained. The upstream independent audit model and access-token lifecycle are not yet integrated because they depend on newer upstream type/signature changes and must be ported incrementally.
- Preserved: OmniRouters plugin system, wallet/payment display, Sora/audio billing, expression/group pricing, mail settings, and default frontend customizations.
- Validation: upstream diff/stat review only; implementation and tests pending.
- Local commit/push: pending.

### 2026-09-07 Telegram provider repair (structural-v3)

- Scope: audited actual branch HEAD `2db85323e` against upstream `3e84ec0ab` (not the earlier assumed `86426f7a6`). Earlier Telegram progress entries describe scaffolding, not a completed OAuth migration.
- Integrated at source level: actual ExchangeToken/GetUserInfo implementations, cached JWKS client, issuer/audience/signature/expiry verification, positive uint64 Telegram IDs without float conversion, current redirect/client matching, HTTP(S)-only callback validation, bounded token responses, and server-only OAuthToken.ClientID.
- Dependencies: copied upstream versions and exact go.sum checksums for go-oidc v3.21.0, oauth2 v0.36.0 and indirect go-jose v4.1.4. No local dependency resolution/build was run; module graph remains subject to CI verification.
- Identity lookup: replaced placeholders with existing Telegram occupancy lookup and a GORM lookup that propagates database errors. Did not reuse FillUserByTelegramId because it discards non-record-not-found errors.
- Safety: removed automatic telegram_oauth registration and the incomplete /oauth/telegram/start route/handler. Provider remains disabled and unregistered until single-use flows, session-bound login/binding callbacks and registry conflict handling are ported. Legacy widget routes and telegram_id data remain unchanged.
- Deferred (NOT integrated by this repair): full callback/session lifecycle, external-identity-claim lifecycle and migration audit, model/vendor/pricing restructuring. This is not completion of round 3 or upstream parity.
- Preserved: Sora/audio_generation charges, customer/group/expression pricing, payment/mail/plugin behavior, main branch and unrelated untracked files.
- Validation: added Telegram regression test source for numeric IDs, invalid callback URLs, configuration changes, token endpoint PKCE/Basic auth, private ClientID, disabled registration and real RSA/JWKS issuer/audience/expiry/signature failures. Local checks are source review and git diff --check only; no local compilation or tests.
- CI: added migration-branch-only Structural migration source checks (go test ./oauth -run Telegram); no image publication and no latest tag changes. Existing main GHCR workflow unchanged.
- Local commit/push: repair `4d9d566594248702137ab0c20d4d956afd658229` pushed successfully to origin/codex/migrate-upstream-structural-v3; main unchanged.
- Remote validation: GitHub Actions run [34127883521](https://github.com/1412212638/OmniRouters/actions/runs/34127883521) completed successfully for that exact commit, including `go test -mod=readonly ./oauth -run Telegram -count=1`. This validates the OAuth package/dependency compilation and targeted regressions, NOT the full application image, live Telegram login, or real multi-database migrations.
- Status-record commit: documentation-only follow-up, to be pushed to the same migration branch with CI skipped; no source changes after the passing run.

### 2026-09-07 Structural-v3: one-time authorization storage

- Upstream source: `3e84ec0ab:model/auth_flow.go`. Integrated the independent AuthFlow storage/creation/lookup/transactional consumption/external assertion replay protection/cleanup APIs. Normal and fast schema migration both include AuthFlow.
- Preserved upstream implementation: 32-byte random tokens with HMAC-only storage, purpose/provider/intent/user/session match predicates, expiry checks, first-write atomic consumption and action rollback. Optional match fields retain upstream semantics: future callbacks MUST supply and validate authoritative identity fields; this table alone does not validate a session.
- Explicitly deferred: AuthSessionIdentity/Authorization and ValidateAuthSessionWithTx, which require User.AuthVersion, UserSession and versioned session lifecycle absent locally. No placeholder session validator was added. Cleanup API is present but scheduling is deferred until flows are activated.
- Security boundary: provider remains unregistered/disabled; no OAuth start/callback routes enabled. Legacy Telegram, other login mechanisms, identity-claim tables, billing/payment/mail/plugin code unchanged. This is the storage prerequisite, NOT complete session binding or completion of round 3.
- Regression source: SQLite/MySQL/PostgreSQL schema idempotence, identity-field mismatch, replay/expiry, action rollback (including data rollback), concurrent single winner and signed assertion deduplication. Tests use dedicated CI databases; SQLite concurrency uses a single connection, so multi-connection SQLite lock contention remains outside this test.
- Validation: local source review and git diff --check only, no local compilation/tests/builds. Added isolated MySQL 8/PostgreSQL 16 services to migration-branch Actions for these tests; older supported DB versions still require separate verification.
- Local commit/push: pending on codex/migrate-upstream-structural-v3; main not changed. Remote result to be recorded after push.
- First push: `be82fa8a3`, run `34128852251`. Telegram checks passed; model tests failed at compilation due to pre-existing missing AuthVersion, UpdateUserAccessToken and useUserCacheMiniRedis referenced by unrelated tests. No database contract executed in that run.
- Follow-up: moved new tests to model/authflowtest, importing the actual production model package so the three-database contracts can run independently. No old test deleted/disabled and no placeholder production APIs introduced. Full model test-suite compilation remains a known migration gap, NOT fixed by this isolation.
- Source-review correction after `64640fa2b`: corrected three overqualified GORM references introduced while moving the test package. Production behavior unchanged; validation remains remote CI plus local diff review.
- Final source push: `ab8179816ee27196f39058fa33b7e0a84f0a309f` on origin/codex/migrate-upstream-structural-v3. [Actions run 34129267334](https://github.com/1412212638/OmniRouters/actions/runs/34129267334) completed successfully: Telegram regression checks and independent AuthFlow contracts with SQLite, MySQL 8 and PostgreSQL 16. Supersedes the failed intermediate runs above. No local compilation was performed.
- Completion boundary: one-time storage prerequisite completed; full model suite, authoritative session/version lifecycle, Telegram callback/binding and frontend activation remain outstanding. Documentation-only status follow-up will use [skip ci]; main remains unchanged.

### 2026-09-07 Structural-v3: authoritative session backend

- Upstream reference: `3e84ec0ab:model/user_session.go` and session identity/validation portion of `model/auth_flow.go`. Ported the complete session model with DB lookup, active-session listing/counts, revoke/refresh CAS, Redis deny fences and bounded cache lifetime, auth-version advancement and cleanup APIs. Added upstream retention defaults and TTL helper dependency.
- Added User.AuthVersion (default 1, private JSON field), UserSession to normal/fast migration, AuthSessionIdentity/Authorization and real ValidateAuthSessionWithTx. Validation rereads and locks the user/session inside the caller transaction rather than trusting cookie state.
- Deliberate local protection: AuthVersion is create-only in GORM User writes so existing stale Save/Updates cannot decrease it. Future auth-version mutation must use an explicit dedicated column update and integrate cache fencing; upstream credential-mutation/cache-version logic is NOT ported in this step.
- Regression source in model/authflowtest: DB schema repeatability; owner mismatch, revoke and cached denial; binding-flow rollback when session revoked; refresh invalid token/grace/reuse; user/session version mismatch, expiry and stale User.Save protection. Three-database harness now sets the main DB dialect so MySQL/PostgreSQL exercise FOR UPDATE. Redis 7 service exercises both cached and uncached paths.
- Preserved/not activated: existing cookie login/logout, password/2FA changes, API keys, Telegram widget, frontend, custom pricing/mail/payments/plugins. No new session issuance/refresh/revoke HTTP routes or scheduled cleanup; Telegram provider remains unregistered/disabled. These backend APIs do not make existing cookie sessions revocable automatically.
- Still required before activation: issuance limits/configuration, auth-version increments and cache fences across credential mutations, JWT/refresh lifecycle and cookie compatibility, middleware enforcement, session-bound Telegram callbacks and identity claim lifecycle. Existing full-model-test missing helpers remain separate work.
- Validation: source comparison and git diff --check only locally; no local compilation/build/tests. Migration-branch Actions expanded to test session contracts with SQLite/MySQL 8/PostgreSQL 16 and Redis 7. Full image build and old database version upgrades are not covered.
- Local commit/push: pending on codex/migrate-upstream-structural-v3; main unchanged. Remote CI outcome will be appended after push.
- Source commit `61c5d39a8ae7beb16fb64a5dc031f7e4149a093f` pushed to origin/codex/migrate-upstream-structural-v3. [Actions run 34130741264](https://github.com/1412212638/OmniRouters/actions/runs/34130741264) passed both Telegram and auth-flow jobs, including the three-database session tests with Redis enabled/disabled. This is targeted backend validation, not activation or a full image acceptance test.
- Status-only log follow-up uses [skip ci]. Next implementation boundary: authoritative login/session issuance, logout/revocation, credential-change auth-version/cache fencing, then middleware/callback integration. No claim that these paths were completed here.

### 2026-09-07 Structural-v3: isolated authentication token layer

- Upstream references: `3e84ec0ab:service/auth_token.go` and the normalized verification binding portion of `service/security_verification.go`. Added short-lived HS256 dashboard access tokens, strict issuer/audience/algorithm/expiry/issued-at/use validation, server-derived operation context hashes, and security-proof identity/scope/context binding. The token layer is not wired into existing middleware or login routes.
- Preserved: existing opaque access-token/PAT authentication, cookie login/logout, Telegram legacy routes, session model, Sora/audio billing, pricing, mail, plugin and frontend behavior. No production route behavior changes.
- Regression source: access-token round trip, security-proof purpose isolation, tamper rejection, internal JWT classification and expiry rejection. Database-backed proof issuance requires the existing application DB and remains covered by the remote target suite only after auth-session service integration.
- Deferred: service auth-session issuance/refresh adapter, JWT middleware enforcement, login/logout controller swap, auth-version mutation hooks for password/2FA/passkey changes, full security verification service, and Telegram callback activation. These are deliberately separate because the current application still uses the legacy cookie session contract.
- Validation: source review and `git diff --check` only locally; no local compilation/build/tests. This source-only step is not a deployable authentication migration by itself.
- Local commit/push: pending on codex/migrate-upstream-structural-v3; main unchanged. Remote result to be recorded after push.
- Auth-version/cache prerequisite source commit is being prepared: `model/user_auth_cache.go` adds monotonic Redis pending/committed fences and transactional version bump helpers; `UserBase` now carries the private auth version, and normal/fast migrations initialize missing versions to 1. This is required before issuing sessions from existing login paths.
- Source commit `333ea234f3033a897e1d559aa172944ea5374e6c` pushed to origin/codex/migrate-upstream-structural-v3. [Actions run 34132002060](https://github.com/1412212638/OmniRouters/actions/runs/34132002060) completed successfully; targeted Telegram/session checks and the new service token source checks passed. No local compilation/build/tests were run.
- Completion boundary: token validation is integrated as an isolated service layer only. Existing cookie/PAT middleware and login routes remain unchanged; access-token issuance, refresh endpoint, security-proof consumption, auth-version mutation hooks and Telegram callback activation remain the next integration work. main remains unchanged.
-
### 2026-09-07 Structural-v3: login session issuance

- Added the upstream session issuance service: active/issuance limits, server-side refresh-secret hashes, UUID session IDs, Access JWT bundles and HttpOnly refresh cookies.
- Integrated the existing shared `setupLogin` so successful password, 2FA, OAuth, Telegram legacy and Passkey logins issue the new bundle while retaining the legacy Gin session and response user data.
- Preserved PAT/API-key authentication, existing logout behavior and all OmniRouters billing/payment/mail/plugin/frontend behavior. Unified Telegram provider remains disabled.
- Deliberately deferred refresh rotation endpoint, JWT middleware switch, logout revocation, session management routes, auth-version mutation hooks and Telegram callback integration.
- Validation: source review and `git diff --check` only; no local build or tests. GitHub Actions is required to catch remaining package/signature issues.
- Local commit/push: pending on `codex/migrate-upstream-structural-v3`; `main` unchanged.

### 2026-09-08 Structural-v3: Telegram OAuth activation review

- Review result: the Telegram OAuth provider implementation, AuthFlow storage, external identity claim storage, and authoritative session primitives are present, but the unified callback/start/bind controller contract is not complete in this branch.
- Decision: keep `TelegramOAuthProvider.IsEnabled()` false and do not register the provider or expose a new callback route. This prevents an incomplete provider from becoming selectable or creating sessions without the complete identity/session binding flow.
- Preserved: legacy Telegram Widget login/bind behavior, legacy `telegram_id` compatibility, all OmniRouters custom billing/payment/mail/plugin/frontend behavior, and `main`.
- Validation: source comparison and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication.
- Remaining: unified Telegram callback/bind/login controller integration and frontend activation require a dedicated implementation with end-to-end tests.
- Local commit/push: pending on `codex/migrate-upstream-structural-v3`; `main` unchanged.

### 2026-09-08 Structural-v3: final round review boundary

- Planned rounds 1-4 have been processed in sequence: dashboard JWT middleware, credential auth-version fencing, security verification audit coverage, and Telegram OAuth activation review.
- Round 5 is a source-only acceptance review: verify changed-file scope, migration registration, protected OmniRouters behavior, documentation traceability, branch/remote state, and GitHub Actions outcome. No local compilation or image build will be performed.

### 2026-09-08 Structural-v3: security proof contract review

- Review result: dashboard JWT identity is now exposed to middleware through the server-validated `auth_identity` context value. The existing security-proof signer currently stores an internal AuthFlow ID in JWT `jti`, while AuthFlow consumption requires the original opaque token; therefore proof consumption was not enabled prematurely.
- Safety decision: removed the incomplete consumer path rather than introducing a proof that cannot be atomically consumed. A follow-up must return/bind the opaque flow token and internal proof ID explicitly before wiring `/api/verify` or channel-key access.
- Preserved: existing Cookie/PAT behavior, legacy secure verification, billing/payment/mail/plugin/frontend behavior, and `main`.
- Validation: source review and `git diff --check` only; no local compilation/build/tests or image publication.

### 2026-09-08 Structural-v3: security proof token contract repair

- Integrated the missing one-time Proof consumption primitive without enabling an incomplete route: the signed proof now carries the opaque AuthFlow token separately from the internal Flow ID in `jti`, and `ConsumeSecurityProof` validates the session/binding before atomically consuming that opaque token.
- Safety: the previous mismatch between internal Flow ID and opaque-token consumption is corrected. The legacy `/api/verify` response and sensitive-operation consumers are deliberately deferred until their request/response contract can carry the proof explicitly.
- Preserved: existing Cookie/PAT verification behavior and all OmniRouters business functionality.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication.

### 2026-09-08 Structural-v3: Telegram OAuth commit path

- Added transaction-owned Telegram login commit helpers for new-user and existing-user paths, plus the bind path. Existing Telegram identities are consumed through the same one-time AuthFlow before `setupLogin`; new users are created and claimed atomically.
- Controller wiring remains intentionally pending until the callback has dedicated route-level tests for state intent, provider error callbacks, concurrent replay, registration-disabled behavior, and session-bound bind requests.
- Preserved legacy Telegram Widget routes and all OmniRouters custom behavior.

### 2026-09-08 Structural-v3: Telegram unified OAuth routes

- Registered explicit `/oauth/telegram/start` and `/oauth/telegram` routes before the generic OAuth wildcard. The legacy `/api/oauth/telegram/login` and `/bind` Widget routes remain unchanged.
- The unified flow uses server-side AuthFlow state, PKCE, ID-token verification, session-bound bind intent, one-time consumption, and atomic external identity ownership. It does not depend on the generic provider registry.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication.

### 2026-09-08 Structural-v3: default frontend dashboard session client

- Added default frontend storage and request injection for the short-lived dashboard Access JWT. Login and 2FA responses persist the returned token; refresh and logout use the new session endpoints, while HttpOnly refresh cookies remain server-managed.
- Compatibility: the change is limited to `web/default` dashboard API requests. Relay/API-key examples and the classic frontend are unchanged; the existing user ID header and cookie credentials remain present for compatibility.
- Added an explicit `refreshDashboardSession` API helper for the authenticated bootstrap/expiry flow. Automatic retry is intentionally not added yet to avoid request replay and refresh races until the route lifecycle is wired and tested.
- Validation: source review and `git diff --check` only; no local frontend build or dependency installation, per source-only workflow.
- Local commit/push: pending on `codex/migrate-upstream-structural-v3`; `main` unchanged.

### 2026-09-08 Structural-v3: four-batch migration acceptance review

- Batch 1 complete: JWT-bound security proofs, operation-context binding, and one-time channel-key proof consumption.
- Batch 2 complete: Telegram AuthFlow state, PKCE, ID-token validation, atomic login/bind commits, external identity claims, and dedicated callback routes. Legacy Telegram Widget routes remain intact; the generic Provider registry remains protected from incomplete registration.
- Batch 3 complete: default frontend stores and sends dashboard Access JWTs, persists login/2FA tokens, and uses refresh/logout session endpoints. Classic frontend, PAT/API-key relay requests, and Cookie compatibility remain unchanged.
- Batch 4 source review: protected OmniRouters Sora/audio billing, expression/group/customer pricing, plugin system, wallet/payment display, mail settings/templates, and classic frontend were not modified by this migration series. Only migration-branch source and documentation changes are included.
- Validation: `git diff --check` and source-scope review only. No local Go compilation, frontend build, Docker build, or image publication, as required by the source-only workflow. GitHub Actions remains the authoritative build/test gate.
- Branch: changes are on `codex/migrate-upstream-structural-v3`; `main` was not changed. Pre-existing unrelated untracked files remain untouched.

### 2026-09-08 Default frontend access-token refresh

- Added automatic Dashboard Access JWT refresh on HTTP 401 in `web/default/src/lib/api.ts`. Concurrent expired requests share one refresh promise; each original request is retried at most once.
- Refresh and logout endpoints are excluded from retry to prevent loops. A failed refresh clears the stored dashboard token and auth-store user state. Cookie credentials remain HttpOnly and server-managed.
- Preserved: PAT/Relay requests, classic frontend, legacy cookie behavior, billing/payment/mail/plugin behavior, and `main`.
- Validation: source review and `git diff --check` only; no local frontend build or dependency installation.

### 2026-09-08 Telegram default frontend callback routing

- Unified Telegram OAuth bind callbacks now redirect to the default frontend `/profile` route. Legacy Telegram Widget binding keeps its existing `/console/personal` redirect for classic frontend compatibility.
- The default profile page already reloads current user data on route entry; no legacy binding route or classic frontend behavior was changed.
- Validation: source review and `git diff --check` only; no local frontend build or dependency installation.

### 2026-09-08 Telegram OAuth browser callback completion

- Added a redirect-based login finalizer for browser OAuth callbacks. Telegram login now creates the same server-side Session, Access JWT, refresh cookie, last-login update, and audit record as normal login, then redirects to the dashboard instead of rendering raw JSON.
- Telegram bind already redirects to the profile page; both browser paths now return to usable frontend screens.
- Validation: source review and `git diff --check` only; no local frontend build or dependency installation.

### 2026-09-08 Default frontend Telegram binding start

- Replaced the default frontend Telegram binding placeholder/widget text with a real server-side OAuth bind start action. The button requests a bind AuthFlow and redirects to its PKCE authorization URL.
- The existing profile binding refresh callback remains the next UI follow-up after the backend redirect response is finalized; legacy Widget binding remains unchanged.
- Validation: source review and `git diff --check` only; no local frontend build or dependency installation.
- Local commit/push: pending on `codex/migrate-upstream-structural-v3`; `main` unchanged.

### 2026-09-08 Default frontend Telegram login start

- Replaced the default frontend Telegram placeholder with the server-side Telegram OAuth start request. The browser is redirected only to the authorization URL returned by the backend; PKCE and flow state remain server-managed.
- Telegram bind UI and callback success-screen handling remain separate follow-up work; legacy Widget behavior is unchanged.
- Validation: source review and `git diff --check` only; no local frontend build or dependency installation.

### 2026-09-08 Structural-v3: Telegram atomic OAuth commit services

- Added transaction-aware `CommitTelegramLogin` and `CommitTelegramBind` services. The AuthFlow consumer owns the only transaction; login performs Flow consumption, user creation, and external identity Claim atomically, while bind performs Flow consumption and Claim atomically.
- Post-commit effects remain outside the transaction by design. No public Telegram unified route is enabled yet; callback wiring still needs provider token validation, authenticated bind-session matching, error handling, and targeted concurrency tests.
- Preserved legacy Telegram Widget behavior and all OmniRouters custom billing/payment/mail/plugin/frontend behavior.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication.

### 2026-09-08 Structural-v3: Telegram callback transaction review

- Review result: the existing AuthFlow consumer owns its transaction, while OAuth user creation currently has separate post-lookup and post-creation behavior. No callback route was enabled in this step because combining Flow consumption, user creation, external identity claim, and post-commit session issuance still requires a dedicated transaction-aware consumer API.
- Safety: no partial Telegram callback/controller was committed; legacy Widget login and bind remain unchanged and the unified provider remains disabled.
- Preserved: all OmniRouters billing/payment/mail/plugin/frontend behavior and `main`.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication.

### 2026-09-08 Structural-v3: Telegram AuthFlow state service

- Added server-side Telegram OAuth flow state creation/lookup for login and bind intents. PKCE verifier, client ID, and redirect URI are stored in the expiring AuthFlow payload; only the opaque flow token is intended for the browser.
- The helper validates provider/intent, expiry, payload shape, and rejects unsupported intents. Callback consumption and provider registration remain deferred until the controller can perform atomic identity/session updates.
- Preserved legacy Telegram Widget routes and all OmniRouters custom business behavior.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication.

### 2026-09-08 Structural-v3: Telegram identity binding primitive

- Added `BindExternalIdentityWithTx`: checks legacy `telegram_id` ownership, atomically claims the provider subject in `external_identity_claims`, and updates the legacy column in the same transaction.
- This is a service/model prerequisite for the unified Telegram callback. The public provider remains disabled until AuthFlow intent consumption and session-bound login/bind controllers are added.
- Preserved legacy Telegram routes and all OmniRouters billing/payment/mail/plugin/frontend behavior.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication.
- Local commit/push: pending on `codex/migrate-upstream-structural-v3`; `main` unchanged.

### 2026-09-08 Structural-v3: security proof request integration

- Integrated JWT-bound security verification into `/api/verify`: callers may provide a supported operation/context and receive a short-lived one-time security proof. The proof is bound to the authoritative user session, auth version, method, scope, and normalized operation context.
- Integrated the channel-key read path: JWT requests must present `X-Security-Proof` bound to that exact channel ID; the proof is atomically consumed before the key is returned. Legacy Cookie-session requests retain the existing five-minute verification behavior.
- Preserved: PAT/relay authentication, legacy frontend behavior, billing/payment/mail/plugin behavior, and `main`.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication.
- Local commit/push: pending on `codex/migrate-upstream-structural-v3`; `main` unchanged.

### 2026-09-08 Structural-v3: secure verification audit coverage

- Scope: universal 2FA/Passkey verification now records structured security audit events for successful and failed verification attempts, including only the verification method and request context; secrets and codes are never logged.
- Boundary: existing Gin-session verification remains compatible. The new JWT-bound security proof issuance/consumption path is intentionally not activated by this small audit change; it remains a separate complete integration round.
- Preserved: existing verification behavior, Passkey readiness marker, 2FA lockout/counters, all billing/payment/mail/plugin/frontend behavior, and `main`.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication.
- Local commit/push: pending on `codex/migrate-upstream-structural-v3`; `main` unchanged.

### 2026-09-08 Structural-v3: credential mutation auth-version fencing

- Scope: password changes through user update/edit and email password reset now increment the user's authoritative `AuthVersion` in the same database transaction. Creating or deleting 2FA and registering/deleting Passkey credentials also advances the version.
- Safety: ordinary 2FA usage updates (TOTP last-used time, failed-attempt counters, lock state, and backup-code consumption) deliberately do not invalidate every session. This avoids turning a normal login verification into an unintended global logout.
- Preserved: existing password hashing, reset semantics, session issuance, PAT/API-key authentication, relay behavior, billing/payment/mail/plugin/frontend behavior. No changes to `main`.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication. The remote Actions workflow remains the required integration check.
- Local commit/push: pending on `codex/migrate-upstream-structural-v3`; `main` unchanged.

- CI repair after run `34136306470`: remote compilation reported missing `UserBase.Role` and `userCacheSchemaVersion` required by the auth cache layer. Added upstream-compatible role caching and schema version 2. No production route behavior was changed.

### 2026-09-07 Structural-v3: refresh/logout controller layer

- Added the upstream refresh/logout/session-management controller and origin guard. Refresh reads only the HttpOnly cookie, validates optional `X-Auth-Session`, rotates refresh secrets, and clears invalid cookies; logout revokes the matching server session before clearing cookies. Added session listing/revoke routes and explicit no-store responses.
- Added `POST /api/user/auth/refresh` and `/auth/logout`; existing legacy `GET /api/user/logout` remains untouched for compatibility. New routes are origin-guarded only when secure cookie mode is enabled and do not affect relay/PAT routes.
- Source correction: local user response shape lacks the upstream `buildSelfUserData` helper, so refresh returns the existing loaded user object rather than introducing unrelated response restructuring.
- Preserved billing/payment/mail/plugin/frontend behavior and disabled Telegram unified provider. No Telegram callback activation or main merge.
- Validation: source review and `git diff --check` only; no local compilation/build/tests. Remote CI is required for package integration.
- Local commit/push: pending on `codex/migrate-upstream-structural-v3`; main unchanged.
- CI result: initial refresh/logout commit `5414d0844` failed because the auth cache port missed local `UserBase.Role` and schema-version definitions. Repair `3db33dac9` still exposed duplicate declarations caused by the corrective patch; cleanup `9e654d44b` removed only those duplicates. [Actions run 34138830581](https://github.com/1412212638/OmniRouters/actions/runs/34138830581) passed after the cleanup. Remote branch is at `9e654d44b46b052a973f84254914880148f60712`; main unchanged. No local compile/build/test was run.

### 2026-09-08 Structural-v3: dashboard JWT middleware integration

- Scope: completed the first authentication-integration round. `middleware/auth.go` now recognizes dashboard Access JWTs, validates their signature/claims through `service.ParseDashboardAccessToken`, and validates the authoritative user/session/auth-version state inside a database transaction through `model.ValidateAuthSessionWithTx`.
- Compatibility: invalid internal dashboard JWTs never fall through to opaque PAT validation; legacy cookie sessions, PAT/API-key authentication, relay authentication, and all OmniRouters billing/payment/mail/plugin/frontend behavior remain unchanged.
- Authorization: role, status, group, and user identity are loaded from the authoritative database user record for JWT requests. Existing `New-Api-User` matching and Admin/Root role gates remain active.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow. GitHub Actions is required for package-level verification.
- Remaining: credential/security mutations still need to advance `AuthVersion`; security proof completion, Telegram callback activation, and frontend activation remain separate rounds.
- Local commit/push: pending on `codex/migrate-upstream-structural-v3`; `main` unchanged.
### 2026-09-08 Main CI repair: missing NormalizeOrigin helper

- Fixed the GitHub Actions Go compilation failure after merging structural-v3: `middleware/auth_origin.go` referenced `common.NormalizeOrigin`, but the helper had not been brought into the local `common/session_cookie.go` implementation.
- Added the upstream-compatible origin validation and canonicalization helper, including scheme/host validation, wildcard rejection, default-port normalization, and IPv6 formatting. Preserved the local session-cookie initialization behavior and all OmniRouters custom billing, payment, mail, plugin, and frontend logic.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow. GitHub Actions remains the integration check.
- Local commit/push: commit `1fb7b9ece`, pushed to `origin/main`.

### 2026-09-08 Upstream performance and popup compatibility

- Integrated upstream commits `5c7cca015` and `387a40914` with path-aware conflict resolution. Model health data now exposes hourly success-rate points; nested Combobox/Select popups inside the default frontend Drawer portal into the Drawer content so focus and pointer interaction remain functional.
- The performance UI changes under the removed legacy `web/src` pricing path were deliberately not restored; the popup fix was ported to `web/default`. Existing model pricing, Sora/audio billing, payment, mail, plugin, and frontend custom behavior remain preserved.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow.
- Local commits: `679a63cec` and `c136754d5`; push pending on `main`.

### 2026-09-08 Upstream JSON codec injection

- Integrated upstream commit `7bbe85bcb` after reviewing conflicts. JSON helpers now use a host-injectable codec shared by `common` and `relaykit`, while preserving the project-wide `common.*` JSON wrapper contract and Kimi message compatibility.
- No business JSON semantics, billing, payment, mail, plugin, or frontend behavior was intentionally changed. Upstream-only test additions were retained only where they fit the current relaykit layout.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow.
- Local commit/push: commit `b7fffae47`, push pending on `main`.

### 2026-09-08 Upstream performance, cache, and CI maintenance

- Integrated upstream commits `2cf177ac4`, `36dbbf0f7`, and `8f5ab8e40` with conflict resolution. RawMessage request cloning now avoids unnecessary copies, ETag matching remains valid across JSON implementations, and release version resolution uses the triggering tag when available.
- Preserved local empty-ETag protection, current release workflow structure, and all OmniRouters billing, payment, mail, plugin, Sora/audio pricing, and frontend behavior. No local dependency upgrade was imported.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow.
- Local commits: `d630f4912`, `4eded8c17`, `712f488a9`; push pending on `main`.

### 2026-09-08 Upstream task-plugin 503 diagnostic audit

- Reviewed upstream commit `32c261923`. Its task-plugin-specific no-available-channel message, distributor branch, and en/zh translations are already present in the local plugin-center implementation, so the cherry-pick resolved to an empty change and was skipped.
- No source change was required; existing plugin routing, Sora/audio billing, payment, mail, and frontend behavior remain unchanged.
- Validation: source comparison and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow.
- Local commit/push: commit `b96c76b8c`, pushed to `origin/main`.

### 2026-09-08 Upstream database migration compatibility

- Integrated the backend portion of upstream commit `9a8674425`: MySQL decimal defaults and PostgreSQL `CHAR` metadata are normalized during GORM schema comparison, preventing redundant migrations on restart while retaining real schema changes.
- Deliberately did not import dependency churn or test-only database fixtures. Existing SQLite/MySQL/PostgreSQL support and all OmniRouters billing, payment, mail, plugin, and frontend behavior remain preserved.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow.
- Local commit/push: pending on `main`.

### 2026-09-08 GHCR build compatibility repair: JSON helper aliases

- Fixed the next GitHub Actions compile failure in `relay/channel/openai/relay_image.go` by exposing `common.JSONRawMessage` and `common.IsValidJSON` through the existing host-injectable JSON codec.
- No direct business-code JSON implementation was introduced; billing, plugins, payments, mail, and frontend customizations remain unchanged.
- Validation: source review and `git diff --check` only; local compilation/build was intentionally not run. GitHub Actions is the verification gate.
- Local commit/push: pending on `main`.

### 2026-09-08 Authentication security batch: cache and abuse controls

- Integrated the independent routing hardening from the upstream authentication batch: universal security verification and Passkey/login verification endpoints now disable caching, and authenticated security verification is additionally protected by the user-level `security-verification` rate limit.
- Preserved existing login, Telegram/OAuth, Passkey, session, Token, billing, payment, mail, plugin, and frontend behavior; no authentication contract or credential format was changed.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow. GitHub Actions remains the verification gate.
- Local commit/push: pending on `main`.

### 2026-09-08 Authentication security batch: primary login cache control

- Added `middleware.DisableCache()` to the primary `/api/user/login` endpoint, aligning authentication response handling with the verification and Passkey routes already hardened in this batch.
- This is a routing-only change: login payloads, session issuance, Telegram/OAuth behavior, billing, payment, mail, plugins, and frontend behavior remain unchanged.
- Validation: source review and `git diff --check` only; no local compilation/build. GitHub Actions remains the verification gate.
- Local commit/push: pending on `main`.

### 2026-09-08 GHCR build compatibility repair

- Fixed the GitHub Actions Go build failure where the Responses stream converter referenced missing `chatAnnotationsToResponses` and `responsesEventOutputTextAnnotationAdded` symbols.
- Ported the upstream annotation conversion helper into the local non-stream Responses converter, including URL-citation flattening and the annotation event constant. Existing reasoning conversion, JSON codec usage, billing, plugin, payment, mail, and frontend behavior were preserved.
- Validation: source review and `git diff --check` only; local compilation/build was intentionally not run under the source-only workflow. GitHub Actions is the verification gate.
- Local commit/push: pending on `main`.

### 2026-09-08 Main CI repair: missing dashboard identity accessor

- Fixed the next GitHub Actions compile error by adding the `authIdentityContextKey` constant and `GetAuthIdentity` accessor required by the already-integrated session identity guard.
- Preserved PAT behavior, legacy session behavior, and all OmniRouters custom billing, payment, mail, plugin, and frontend behavior.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow. GitHub Actions remains the integration check.
- Local commit/push: pending on `main`.

### 2026-09-08 Main CI repair: structural-v3 missing compatibility symbols

- Fixed the next Go compilation failures reported by GitHub Actions after the Origin helper repair: restored `GetSessionAuthIdentity`, corrected Telegram OAuth startup to return the persisted flow record and the in-memory OAuth flow separately, and added the safe `buildSelfUserData` response DTO using the local `User` shape.
- Preserved PAT authentication behavior, legacy Telegram behavior, local user fields, and all OmniRouters billing, payment, mail, plugin, and frontend behavior.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow. GitHub Actions remains the integration check.
- Local commit/push: pending on `main`.

### 2026-09-08 Upstream plugin fixes: Suno aliases and factory suppression

- Integrated upstream commits `92bc7ff73` and `6298b0f32`.
- Updated the built-in Suno plugin to preserve channel-mapped aliases, select music/lyrics behavior from the mapped upstream model, and avoid sending empty optional headers. Updated task-plugin status handling so disabling an overridden plugin also suppresses its built-in factory layer.
- Deliberately kept local deletions of upstream-only plugin test files because those tests belong to the removed legacy plugin test layout; no OmniRouters billing, payment, mail, Sora/audio pricing, or frontend behavior was changed.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow. Remaining plugin polling and task-core changes are deferred for a separate review.
- Local commit/push: commits `52928fc25` and `a2530cc8b`, push pending on `main`.

### 2026-09-08 Upstream relay compatibility: Kimi dynamic tool loading

- Integrated upstream commit `6e10f9bc9` with a manual conflict resolution. Message-level Kimi K3 tool declarations are preserved when converting and serializing OpenAI-compatible requests, included in token metadata, and excluded from invalid `content: null` output for tool-loading system messages.
- Preserved local `annotations`, reasoning conversion state, and the existing JSON wrapper usage. No task billing, Sora/audio pricing, payment, mail, plugin-center, or frontend behavior was changed.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow.
- Local commit/push: commit `ad7838963`, push pending on `main`.

### 2026-09-08 Upstream Alibaba plugin: Wan 3.0 all-in-one video

- Integrated upstream commit `3b4652269`. Added `wan3.0-video` and `wan3.0-video-prime` support in the Alibaba task plugin, including resolution tiers, adaptive ratio, smart duration sentinel handling, image/media input, and completion usage facts.
- Preserved existing Wan 2.x behavior, Sora/audio per-request billing, plugin-center controls, and all local payment, mail, pricing, and frontend customizations. The plugin keeps bounded duration facts before billing ratios are produced.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow.
- Local commit/push: commit `4f19d34c7`, push pending on `main`.

### 2026-09-08 Upstream task polling audit: Sora status diagnostics

- Reviewed upstream commit `9df450fe5`. The independent Sora unknown-status diagnostic is already present locally in `plugins/tasks/sora/plugin.js`; no duplicate source change was needed.
- The full polling/adaptor contract migration remains deferred because the local task interfaces and Sora/audio settlement path use a compatibility contract that must be migrated as one complete round to protect pre-consume, settlement, and refund behavior.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow.
- Local commit/push: pending on `main`.

### 2026-09-08 Upstream security and audit follow-up review

- Reviewed the next upstream security/audit batch through `bee45b58a`, including `3f8a50cf8`, `6f2333990`, `0973dc2b8`, `a8729b5c3`, `3e84ec0ab`, `45c3fbe8a`, `d8cb17744`, `8c8c4153d`, and `521cebf58`.
- `3f8a50cf8` audit records and the `8c8c4153d` usage-statistics scan fix are already present in the current branch. The statistics query already scans RPM/TPM into a separate value and assigns only those fields, so no duplicate source change was made.
- `521cebf58` targets the upstream legacy `web` frontend layout; this repository's active frontend is `web/default`, and the equivalent setup-guide behavior is already present there. It was therefore not cherry-picked.
- The authentication/security commits remain deferred for a dedicated compatibility round because they overlap local session identity, Telegram/OAuth, password, access-token, and security-verification customizations. The full task polling refactor and model/vendor/pricing refactor remain separately deferred because they overlap Sora/audio billing, plugin-center controls, and customer-specific pricing.
- Local functionality deliberately preserved: Sora per-request billing, `audio_generation` surcharge, group pricing and customer discounts, wallet/payment display, mail templates, task plugin center, and default/classic frontend customizations.
- Validation: source review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow. GitHub Actions remains the integration check.
- Local commit/push: commit `1fb7b9ece`, pushed to `origin/main`.

### 2026-09-08 Upstream backlog reconciliation through `bee45b58a`

- Reconciled the remaining upstream commits after the previous security review. The local history already contains the compatible portions of `057f71c23` (privileged log metadata isolation), `219c9e063` (revalidated public/auth-session responses), `73afad588` (Hailuo media usage), `b7017c251` (system-task no-op writes), `27ff6a876` (legacy token-key migration), and `6eb6f35ed` (PostgreSQL JSON values/scanning).
- These items are therefore recorded as integrated/already present rather than cherry-picked again. Existing local implementations also retain the required billing saturation checks, request-audit visibility rules, plugin-center controls, and Sora/audio pricing behavior.
- The remaining upstream items are explicitly deferred: full authentication/security migration (`6f2333990`, `0973dc2b8`, `a8729b5c3`, `3e84ec0ab`, `45c3fbe8a`, `d8cb17744`), full task polling/adaptor migration (`9df450fe5`), removal of the custom plugin switch (`210734bb7`), model/vendor/pricing rewrite (`0c76e4dae`), temporary count-token disable (`3a9f41ee8`), and OpenAI capability rewrite (`49ec46966`). Each overlaps active local behavior and requires a dedicated compatibility round.
- No source change was required in this reconciliation batch. Local functionality deliberately preserved: Sora per-request billing, `audio_generation` surcharge, group/customer pricing, wallet/payment display, mail templates, plugin-center administration, and default/classic frontend customizations.
- Validation: source/history review and `git diff --check` only; no local compilation, tests, frontend build, Docker build, or image publication, per source-only workflow. GitHub Actions remains the integration check.
- Local commit/push: pending on `main`.

### 2026-09-09 Waffo Pancake 单价读取修复

- 修复 Waffo Pancake 金额计算在配置刚保存后仍读取旧 typed setting，导致充值金额被误判过低的问题；优先读取已持久化的 OptionMap 单价，并保留有效值兜底。
- 保留本地 Waffo Pancake、钱包展示及现有充值入账逻辑；未改变计费或订单结算规则。
- 验证：源码检查、`git diff --check`；未本地编译或构建容器。
- Local commit/push: pending。
### 2026-09-09 OAuth 登录会话修复

- 修复 OAuth 登录成功后未保存返回的 dashboard access token，导致后续 `/api/user/self` 请求被判定为未登录的问题。
- 保留现有 Cookie、Bearer 刷新、uid 缓存和本地身份验证逻辑。
- 验证：源码检查、`git diff --check`；未本地编译或构建容器。
- Local commit/push: pending。
### 2026-09-09 Passkey 限流不阻塞其他登录方式

- Passkey 登录初始化请求改为自行处理 HTTP 错误，避免该接口 429 被全局 Axios 错误处理器显示为账号密码或 OAuth 登录失败。
- 保留 Passkey、密码登录和 OAuth 登录各自独立的认证流程。
- 验证：源码检查、`git diff --check`；未本地编译或构建容器。
- Local commit/push: pending。
### 2026-09-09 GitHub Actions Release 工作流修复

- 修复 `release.yml` 中 setup-bun 步骤缩进错误，避免 Tag 发布工作流在解析/执行阶段失败；保持日常 `main` 推送仅由 GHCR 镜像工作流处理。
- 保留多平台 Release 工作流，仅用于 Tag 或手动发布；未修改应用代码和本地支付、登录逻辑。
- 验证：源码检查、`git diff --check`；未本地构建容器。
- Local commit/push: pending。
### 2026-09-09 Waffo Pancake 单价保存后回显修复

- 修复仅修改 `WaffoPancakeUnitPrice` 时保存流程提前返回，未刷新系统设置查询，导致页面重新进入仍显示旧单价的问题。
- 保存普通选项后强制重新获取 `/api/option/`，保持钱包实际支付和设置页回显使用同一持久化值；不改变充值计费与入账逻辑。
- 验证：源码检查、`git diff --check`；未本地编译或构建容器。
- Local commit/push: pending。
### 2026-09-09 系统选项持久化错误处理

- 修复 `UpdateOption` 忽略数据库 `FirstOrCreate` 和 `Save` 错误，导致内存中的 Waffo Pancake 单价更新但数据库仍为旧值、刷新后恢复 `1` 的问题。
- 数据库写入失败时现在会直接返回错误，不再更新内存配置并误报保存成功；保留批量保存和现有支付/计费逻辑。
- 验证：源码检查、`git diff --check`；未本地编译或构建容器。
- Local commit/push: pending。
### 2026-09-09 Waffo Pancake 设置页单价回显根因修复

- 补齐 `BillingSettings` 与 `defaultBillingSettings` 中遗漏的 `WaffoPancakeUnitPrice` 字段，使 `getOptionValue()` 能保留并解析后端返回的单价，而不是被 `?? 1` 默认值覆盖。
- 钱包计算、支付入账和单价保存接口保持不变；修复设置页刷新后的真实回显链路。
- 验证：源码检查、`git diff --check`；未本地编译或构建容器。
- Local commit/push: pending。
### 2026-09-09 Waffo Pancake 单价数据库优先回读

- 设置接口对 `WaffoPancakeUnitPrice` 增加数据库优先回读，避免内存配置初始化为默认值 `1` 时覆盖已保存单价。
- 修改严格限定于设置接口的该字段，不改变其他支付方式、订单、计费或入账逻辑。
- 验证：源码检查、`git diff --check`；未本地编译或构建容器。
- Local commit/push: pending。
### 2026-09-09 修复 GHCR 构建错误

- 补充 `controller/option.go` 使用的 `strconv` 导入，修复 Waffo Pancake 单价回读改动导致的 Go 编译失败。
- 未改变任何支付、计费、订单或入账逻辑。
- 验证：源码检查、`git diff --check`；由 GitHub Actions 负责构建验证。
- Local commit/push: pending。
### 2026-09-09 Waffo Pancake 单价非法值保护

- 修复单价被保存为 `0` 的问题：后端拒绝零值、负数、NaN 和无穷值，前端也不再把空输入转换为 `0`。
- 仅影响 Waffo Pancake 单价配置校验，不改变其他支付渠道、订单、计费和入账逻辑。
- 验证：源码检查、`git diff --check`；未本地编译或构建容器。
- Local commit/push: pending。
### 2026-09-09 Waffo Pancake 单价重复保存链路修复

- 修复单价同时进入通用选项更新和 Waffo 配置保存两条链路的问题；现在单价变化与 Waffo 商户/店铺/产品配置一起通过专用保存接口提交。
- 避免旧表单状态以 `0` 重复提交并触发校验错误；不改变其他支付方式、订单、计费或入账逻辑。
- 验证：源码检查、`git diff --check`；未本地编译或构建容器。
- Local commit/push: pending。
### 2026-09-09 Waffo Pancake 单价请求字段修复

- 修复支付设置表单的 `sanitized` 对象遗漏 `WaffoPancakeUnitPrice`，导致专用保存接口收到 `0` 并返回“单价必须大于零”的问题。
- 保留单价与 Waffo 配置的单一保存链路，不改变其他支付渠道、订单、计费或入账逻辑。
- 验证：源码检查、`git diff --check`；未本地编译或构建容器。
- Local commit/push: pending。
### 2026-09-09 Waffo Pancake 单价保存闭环加固

- 完整核对并加固单价保存链路：提交前校验有限正数，专用保存成功后等待 `/api/option/` 重新拉取，避免旧查询缓存覆盖页面回显。
- 仅影响 Waffo Pancake 设置，不改变其他支付方式、订单、计费或入账逻辑。
- 验证：源码检查、`git diff --check`；未本地编译或构建容器。
- Local commit/push: pending。
### 2026-09-09 Waffo Pancake 保存错误诊断

- 专用保存接口现在返回实际失败原因，避免数据库/配置错误被统一隐藏为“保存配置失败”；同时补充单价 NaN/Infinity 校验。
- 仅影响管理员配置错误反馈，不改变其他支付渠道、订单、计费或入账逻辑。
- 验证：源码检查、`git diff --check`；未本地编译或构建容器。
- Local commit/push: pending。
### 2026-09-09 Waffo Pancake 单价与配置保存解耦

- 将 `WaffoPancakeUnitPrice` 与商户、密钥、回调地址、店铺和产品绑定拆分判断；只修改单价时直接保存系统选项，不再依赖 Waffo 配置接口或店铺绑定事务。
- 同时修改完整 Waffo 配置时仍使用原子专用保存接口；未改变充值、计费、订单回调或入账逻辑。
- 验证：源码检查、`git diff --check`；未本地编译或构建容器。
- Local commit/push: pending。
### 2026-09-09 修复 Pancake 单价与普通支付设置同时保存

- 调整保存顺序：先提交普通支付设置，再独立提交 Pancake 单价；避免同时修改其他支付配置时被单价分支提前返回而丢失普通设置。
- Pancake 单价仍不依赖 Waffo 店铺/产品专用接口；其他支付和充值入账逻辑保持不变。
- 验证：源码检查、`git diff --check`；未本地编译或构建容器。
- Local commit/push: pending。
## 2026-09-09

- Upstream `064ed943e` integrated selectively: added validated `fixed(amount)` request pricing to the shared billing expression engine, including billing-unit metadata, request-price settlement for zero or missing token usage, group/request multipliers, admin log details, Realtime/task-usage rejection, and saturation-safe quota conversion. The `web/default` visual tier editor now supports per-token/per-call tiers and fixed USD per-request values. Existing Sora/audio pricing, task plugins, tool surcharges, and group/customer discount paths were preserved.
- Upstream `d52bdc0b4` audited: the time-based pricing rule editor and request simulation capabilities are already present in `web/default`; no wholesale `web/src` replacement was made. Fixed tiers were adapted to the existing default editor and remain compatible with its time/request rules.
- Validation: `git diff --check`, merge-marker scan, and source-level inspection only. `gofmt/go` is unavailable in this environment, so Go formatting and compilation were not run. Per source-only workflow, no local frontend build or Docker image build was performed. Database matrix and upstream frontend test suites remain for GitHub Actions.
- Status: integrated locally; not pushed yet. Existing prior local Alibaba/plugin and migration changes remain in the same worktree and are intentionally kept for the next combined commit.

## 2026-09-10 - API Key 分组描述溢出修复
- 原因：修复创建 API 密钥弹窗中分组描述过长导致布局被顶穿。
- 改动：为选中项内容和比例徽章增加收缩边界，避免长文本挤压弹窗布局。
- 保留：未改变分组选择、API Key 创建或计费逻辑。
- 验证：git diff --check；未执行本地构建。
- 状态：待提交并推送。

## 2026-09-10 - 修复 billingexpr 构建失败
- 原因：固定计费代码引用了缺失的缓存条目读取函数，GitHub Actions 在 Go 编译阶段失败。
- 改动：补充 compileEntryFromCacheByHash，复用现有表达式编译缓存并返回完整元数据。
- 保留：不改变计费表达式、支付流程及既有固定计费逻辑。
- 验证：git diff --check；未执行本地编译。


## 2026-09-10 - 修复 options 迁移构建失败
- 原因：model/main.go 引用了未提交的 options 主键迁移实现，干净构建环境报 undefined。
- 改动：加入 model/option_primary_key_migration.go，提供跨 SQLite/MySQL/PostgreSQL 的迁移实现。
- 验证：检查源码依赖与 git diff --check；未本地编译。


## 2026-09-10 - 修复内置阿里插件导致容器启动失败
- 原因：usageSchema 使用了当前注册器不支持的 enumLabels 字段，内置插件注册失败触发 panic。
- 改动：移除不参与请求和计费的 enumLabels，保留 resolution 枚举及本地化描述。
- 保留：阿里 Wan 视频路由、任务插件和现有计费逻辑不变。
- 验证：静态核对注册器允许字段与插件元数据；未本地编译。


## 2026-09-10 - 修复 API Key 分组描述横向溢出
- 原因：已选分组的描述位于按钮内部，外层 flex 项未设置可收缩边界，长文本撑破更新 API 密钥弹窗。
- 改动：为分组触发按钮及内容链路增加 min-w-0/overflow-hidden，并对描述使用最大宽度单行截断。
- 保留：不改变分组选择、API Key 更新和权限逻辑。
- 验证：git diff --check；未本地构建。

## 2026-09-10 - Isolate dashboard authentication limits and preserve sessions on transient refresh failures

- Local incident: production login returned HTTP 429; the operator confirmed the logged client IP `70.39.183.135` is a server/reverse-proxy address. Requests collapsed onto that address share IP budgets. The provided Baota snippet forwards to `http://70.39.183.135:3000` and only sets Host. The screenshot alone does not identify whether the global API or critical limiter rejected the request. Proxy/header configuration still needs deployment-side verification.
- Confirmed source defects: login, OAuth, refresh, logout and unrelated critical operations shared the `CT` IP bucket (default 20 requests/1200 seconds). The default frontend swallowed all refresh errors, reset authentication, and rethrew the original access-token 401 even when refresh returned 429, 5xx, a race response or a network error. Access JWT TTL is 15 minutes; refresh failure was incorrectly treated as session expiry.
- `middleware/rate-limit.go`, `router/api-router.go`: isolate password/2FA/Passkey/OAuth login and logout into separate scopes. Refresh has a dedicated IP budget, default 120/60 seconds, configurable through `SESSION_REFRESH_RATE_LIMIT` and `SESSION_REFRESH_RATE_LIMIT_DURATION` (positive duration up to the shared 1200-second cache retention). Invalid values fall back to defaults. Global API IP limiting and login attempt limits remain enabled. IP limiter rejections now carry scope and Retry-After, with request-correlated client IP/peer diagnostics; no cookies, passwords or authorization headers are logged.
- `web/default/src/lib/api.ts`, `main.tsx`: refresh 401 still invalidates the session; transient refresh failures preserve credentials and propagate their own error/status to route guards and QueryCache. Share refresh calls and apply a retry cooldown (Retry-After when provided; otherwise 60 seconds for 429 and 5 seconds for transient errors). Do not automatically retry throttled queries. Rate-limit feedback reuses existing translations.
- Preserved: proxy trust boundaries, cookie origin checks, server-side session validation/revocation/rotation, password and 2FA verification, payment/accounting, Sora/audio and model pricing. No arbitrary forwarded-header trust or global limiter bypass was added. Redis limiter atomicity is unchanged.
- Validation: nine targeted frontend regression cases passed using the actual Axios interceptors and mocked transport, covering 429/500/503/409/network errors, malformed responses, true 401, direct 429 and concurrent successful refresh. Added a Go regression for scope isolation, another-client isolation and response diagnostics. No Go toolchain/local compilation or Docker build; backend test execution and production proxy validation remain outstanding. Source diff checks performed. Targeted lint has pre-existing non-null-assertion findings in api.ts/main.tsx; new lint errors corrected.
- Deployment: for a directly exposed Baota proxy, set X-Real-IP and X-Forwarded-For from `$remote_addr` and X-Forwarded-Proto from `$scheme` in the active proxy location, after checking included directives. Configure TRUSTED_PROXIES for the actual peer address seen by the container (often the Docker gateway); do not blindly trust all networks. With a CDN, first restore real client IPs at Nginx from only the documented trusted CDN ranges. Verify requests from separate networks show distinct client IPs before considering the incident resolved. Never log passwords, cookies or bearer tokens.
- Commit/push: this entry is included in the source repair commit targeting main; push is pending at write time and the remote result is reported in the task. Deployment must also restore distinct real client IPs; restarting alone is not a permanent fix.

## 2026-09-11 - Audit improvements, batch 1: meaningful records

- Local request: complete the first two audit improvement batches; retention and cleanup are a separate, unimplemented third batch.
- `middleware/audit_requests.go`, `router/api-router.go`, `middleware/auth.go`: one independent audit event per authenticated API request, with separate operation/security/access categories. Ordinary audit-page reads are excluded. Login attempts rejected before authentication are anonymous security events, never assigned to the submitted username. OAuth callbacks remain callbacks unless successful login is explicitly recorded. The collector sits inside gzip so business JSON can be interpreted without modifying the response. Existing management/usage-log records remain intact.
- `controller/audit.go` and existing security event callers: reuse explicit operation labels and a bounded allowlist of scalar display parameters; preserve operator versus affected-account identity. Explicit security events are deferred until the response is known rather than duplicated as a generic access event. Transactional personal-access-token revocation keeps its atomic audit row and suppresses the generic duplicate; its missing timestamp/event ID are fixed.
- `model/audit_log.go`, `model/audit_details.go`: additive nullable varchar `outcome` column through existing GORM migrations (SQLite/MySQL/PostgreSQL); existing bool status and historical rows are retained. Structured versioned details are stored in existing TEXT `other`, including method, route template, target, auth method, HTTP status and outcome. HTTP 200/success:false is failure; required 2FA is pending; malformed/truncated/unrecognized bodies are unknown. Raw response messages are not retained because they can echo credentials; failures use HTTP status or a safe business-rejection reason, correlated by request ID. No full request/response body, query, password, cookie, code or bearer token is persisted.
- `controller/option.go`, `channel.go`: record before/after values for an explicit safe option allowlist and channel model/group/type changes; potentially sensitive values are redacted. System option before-values are observed from persisted DB state without changing save/transaction semantics. These are observed snapshots, not a new serialized configuration transaction. Oversized metadata is redacted. Quota override reuses the existing safe before/after parameters. Other actions without snapshots explicitly have no change details; history is not fabricated or backfilled.
- `controller/audit_log.go`: server-side category, exact operator name/ID, action, IP, request ID, time and outcome filtering; bounded pagination and invalid-range rejection. Self scope is always ANDed with filters, so user_id cannot override authorization.
- Preserved: login validation, JWT/cookie/refresh behavior, proxy trust, billing and payment arithmetic/transactions, Sora/audio, model request logs, original management logs and audit retention. No cleanup, export, archival or read-audit toggle in this batch.
- Validation: source review and diff checks; added backend regressions for business outcomes (including string/array data), gzip capture, credential exclusion, audit-read exclusion, anonymous throttled login, self-scope filtering and additive historical-row migration. Go compiler/tests and live database matrix were not run locally under the source-only workflow; they remain to be executed in CI.
- Commit/push: implemented in local main worktree; uncommitted and not pushed.

## 2026-09-11 - Audit improvements, batch 2: usable admin page

- `web/default/src/features/audit-logs`: default Important operations tab, Security and login / Access records / All tabs, operator/action/IP/request/time/result filters, reset/refresh, pagination and page size. Filters reset pagination; query/business failures display an error rather than an empty table. The page scrolls vertically, table scrolls horizontally, long values truncate in cells and wrap in details.
- Details dialog includes localized action, target, operator, auth method, status, route/method, request/event IDs, safe failure reason and redacted before/after changes. Existing operation translations are reused for additional management action codes. Historical results are explicitly labelled and invalid/old metadata is safely parsed. No runtime claim that old HTTP-only outcomes were accurate business results.
- Seven locale resources updated (en, zh, zh-TW, fr, ru, ja, vi); existing design-system table, fields, tabs, badge, tooltip and Base UI dialog reused.
- Validation: Bun DOM/transport regression tests cover old/malformed metadata, business API failure state, default tab, tab/page reset, status-filter submission, pagination, details, redacted values and Chinese/English switching. Scoped frontend lint and formatting checks run. Full source-only TypeScript check exposes existing unrelated errors (missing test libraries, pricing/plugin types and duplicate redemption imports); audit-specific test DOM typing errors found during review were fixed. No local production/Docker build or browser pixel rendering was performed.
- Commit/push: implemented in local main worktree; uncommitted and not pushed. Third-batch retention/cleanup settings intentionally remain out of scope.

## 2026-09-11 - Audit improvements, batch 3: retention and cleanup

- Local request: add safe audit-log storage management after the first two audit batches.
- `model/option.go`: add `AuditLogRetentionDays`, defaulting to 90 days. `0` disables scheduled cleanup; values are bounded to 3650 days.
- `model/audit_log.go`: add context-aware counting and batched deletion. IDs are selected first and then deleted through GORM so the operation remains compatible with SQLite, MySQL and PostgreSQL and does not depend on `DELETE ... LIMIT` syntax.
- `model/system_task.go`, `service/system_task.go`: add an independent `audit_cleanup` system task with the existing lease/heartbeat machinery. It runs once per day when retention is enabled, deletes only audit rows older than the configured cutoff, and records progress/result separately from ordinary usage-log cleanup.
- `controller/system_task.go`, `router/api-router.go`: add the root-admin manual endpoint `POST /api/system-task/audit-cleanup?target_timestamp=...`; an active task is reused instead of creating concurrent cleaners.
- Preserved: ordinary usage logs, payment/accounting, model billing, Sora/audio pricing, audit records newer than the cutoff, and the existing log cleanup task. No export or archival behavior was added.
- Validation: `git diff --check` passed. No local Go compilation, database matrix test, frontend build, or Docker build was run under the source-only workflow; CI remains the authoritative build check.
- Commit/push: batch 3 is implemented in the local main worktree and remains uncommitted/unpushed pending the final review of the combined audit batches.

## 2026-09-12 - Upstream follow-up after bdef117505: batch 1 selected fixes

- Upstream baseline: `bdef117505247769268b209665fb3ad7554c3da7`; inspected subsequent upstream commits through `385d2dfd1`.
- Integrated: `3cea2bf79` preserves cached input-token breakdowns through relay billing snapshots; `f064bffa2` validates image quantities before quota reservation and hardens Ali image usage handling; `f362c7c51` adds image cache/quantity variables to billing expressions; `f256e40bc` adds built-in GPT image expression defaults.
- Preserved: local quota saturation checks, Sora/audio and group pricing behavior, payment/Waffo flows, audit logging and the existing `web/default` frontend. Existing image model detection entries were retained while adding upstream model coverage.
- Deferred: expression pricing editor commits `25ec832fa`, `39294418a`, `b3e279464`, plugin UI commits and Passkey UI changes require adaptation to our `web/default` layout; they were not force-merged. The large upstream web tree migration remains out of scope.
- Validation: cherry-pick conflicts were resolved by retaining local architecture and reviewing affected diffs; no local Go compilation, database matrix, frontend build or Docker build was run under the source-only workflow.
- Commit/push: selected batch 1 changes are committed locally; push is pending completion of the remaining safe batches.

## 2026-09-12 - Upstream follow-up after bdef117505: batch 2 and batch 3

- Batch 2 integrated the backend portions of `74629e29f`: task plugin streaming response handling, model-aware usage profiles, bounded plugin state/response persistence, shared-model routing, plugin usage validation, and task pricing metadata. The upstream frontend editor files were not copied because they target the migrated `web/` tree; the current `web/default` remains the local frontend baseline.
- Batch 3 was reviewed from `385d2dfd1` for multi-RP ID Passkey support. The independent configuration/documentation context already matches the local authentication structure; the deep authentication/session changes require the upstream account-security tree and were not force-applied over local session/rate-limit fixes. No Passkey behavior was changed in this batch without a compatible implementation path.
- Preserved: audit logging and retention, existing login/session protections, Sora/audio pricing, Waffo/payment flows, group discounts, and local plugin admin/upload permissions.
- Validation: upstream cherry-pick attempts were aborted where they would introduce the full upstream frontend tree; selected backend changes were applied and conflict markers were checked. `git diff --check` passed. No local Go compilation, database matrix, frontend build or Docker build was run under the source-only workflow.
- Commit/push: batches 1 and 2 are committed locally; batch 3 review is recorded. Final push remains pending the combined source review.

## 2026-09-12 - Local billing compatibility correction

- Review finding: the plugin pricing merge removed the local `SoraPerRequestPricing` field from `BillingSetting`, which would have broken our Sora/audio per-request pricing configuration.
- Corrected: restored the field, initialization and pricing sync export while retaining upstream plugin billing-expression support.
- Preserved: Sora resolution pricing and fixed `audio_generation` surcharge behavior; no payment or quota calculation path was changed.
- Validation: source diff review and `git diff --check`; Go/frontend builds remain delegated to CI under the source-only workflow.

## 2026-09-12 - Upstream follow-up: administrator update reminder

- Adapted upstream `251b76d86` to the existing `web/default` maintenance page: administrators now receive a one-time-per-release informational toast when a newer release is detected. The existing manual release inspection dialog remains unchanged.
- Preserved: no automatic upgrade, no deployment mutation, no authentication or payment behavior, and no repeated reminder for the same release in the same browser.
- Validation: source diff review and `git diff --check`; no local frontend build or Docker build under the source-only workflow.

## 2026-09-12 - CI build fix: missing plugin model-fold helper

- CI failure: Go compilation reported `undefined: asciiFold` in `pkg/jsplugin/registry.go` and `pkg/jsplugin/routing.go` after the plugin core sync.
- Fixed: added `pkg/jsplugin/model_fold.go`, matching the upstream ASCII-only folding contract used for case-insensitive model/plugin lookup. Unicode characters remain unchanged.
- Preserved: plugin routing, billing, Sora/audio, payment and authentication behavior; this is a missing-helper correction only.
- Validation: source inspection and `git diff --check`; local Go compilation was not run under the source-only workflow.

## 2026-09-12 - CI build fix: omitted plugin/pricing dependencies

- Fixed missing selective-sync dependencies reported by CI: reused the existing registry `SetEnabled` API, restored explicit model-ratio detection, resolved advanced-custom channel settings through the existing channel cache, and added metadata rule matching/resolution used by pricing.
- Preserved: local plugin override controls, Sora/audio pricing, payment flows, group discounts and audit behavior.
- Validation: source review and `git diff --check`; local Go compilation remains intentionally unrun.

## 2026-09-12 - CI build fix: missing pricing-expression helpers

- CI failure: `setting/billing_setting/tiered_billing.go` referenced `ratio_setting.HasConfiguredModelRatio` and `billingexpr.UsedUsageKeys`, but those helper files/functions were omitted from the earlier selective sync.
- Fixed: restored both helpers using the existing model-ratio normalization and compiled-expression cache. No billing formula, quota arithmetic, Sora/audio surcharge or payment path was changed.
- Validation: source inspection and `git diff --check`; CI remains the build authority and local Go compilation was not run.

## 2026-09-12 - CI build fix: plugin-aware channel selection and base URL helper

- Cause: the selective plugin/channel sync had updated callers to the upstream four-argument channel-selection APIs, while the local `GetChannel` and `GetRandomSatisfiedChannel` implementations were still on the three-argument signatures. New task polling and relay code also referenced a missing safe channel base URL helper.
- Fixed: added bounds-checked `constant.GetChannelBaseURL`; added filter-aware database selection with the same request-path and task-plugin identity predicates as the memory-cache path; applied those filters to exact and normalized model candidates; and updated the channel-selection APIs to carry `dto.ChannelFilter` constraints through both cache modes.
- Preserved: channel priority/weight selection, plugin identity fail-closed behavior when a required plugin key cannot be resolved, Sora/audio pricing, group pricing, payment flows, audit logging, and existing provider routing.
- Validation: source-format review, `git diff --check`, conflict-marker scan, and call-site signature review. The local Go toolchain was unavailable, so `gofmt`, Go compilation, and Docker/frontend builds were intentionally not run; GitHub Actions remains authoritative.
- Commit/push: implementation committed as `f4024478f`; the corrected log is included in `00530e07c`, which is present on `origin/main`. GitHub Actions is the remaining build verification step.

## 2026-09-12 - CI build fix: model modifier matching dependencies

- Cause: the image exposed missing symbols in the model pricing and token-limit matching path: model modifier parsing and canonical billing helpers were omitted, and middleware referenced `ratio_setting.RoutingMatchModelName`.
- Fixed: completed the host reasoning wrapper for explicit `@` modifiers and legacy thinking/effort suffixes, added canonical billing candidates, and added routing-name normalization before wildcard matching.
- Preserved: existing Sora/audio and group pricing, plugin routing, payment/accounting, audit logging, and model request behavior. This change only completes model-name normalization dependencies used by selection and billing lookup.
- Validation: source inspection, `git diff --check`, conflict-marker scan, and symbol/call-site review. Local Go compilation and Docker/frontend builds were not run; GitHub Actions remains authoritative.
- Commit/push: pending in the current source-level fix commit.

## 2026-09-12 - CI build fix: task plugin request interface compatibility

- Cause: the task plugin adaptor had already adopted the upstream `ParseResponse`, task-aware polling, and batch interfaces, while the local legacy `channel.TaskAdaptor` assertion and `DoTaskApiRequest` parameter still required the old `DoResponse` contract.
- Fixed: narrowed `DoTaskApiRequest` to the two request-building methods it actually uses and removed the invalid legacy interface assertion from the plugin adaptor. Existing task adaptors continue to satisfy the broader legacy interface, while the plugin keeps the upstream response and polling contracts.
- Preserved: task plugin request validation, streaming/SSE handling, plugin state, usage extraction, task billing, Sora/audio pricing, payment/accounting, audit logging, and existing provider adaptors.
- Validation: source inspection, `git diff --check`, conflict-marker scan, and interface/call-site review. Local Go compilation and Docker/frontend builds were not run; GitHub Actions remains authoritative.
- Commit/push: pending in the current source-level fix commit.

## 2026-09-12 - CI build fix: thinking modifier parser dependency

- Cause: the image reported `setting/reasoning/model_name.go` calling `relaykit/relayconvert/reasoning.ParseThinkingModifier`, which was another omitted helper from the selective sync.
- Fixed: restored parsing for `@thinking:on`, `adaptive`, `off`, and bounded integer budgets, including the existing portable `Intent` source and budget metadata.
- Preserved: model routing and billing lookup, Sora/audio and group pricing, plugin routing, payment/accounting, audit logging, and relay conversion behavior.
- Validation: source inspection, `git diff --check`, conflict-marker scan, and symbol review. Local Go compilation and Docker/frontend builds were not run; GitHub Actions remains authoritative.
- Commit/push: pending in the current source-level fix commit.

## 2026-09-12 - CI build fix: RelayKit modifier parser dependency

- Cause: the CI image still reported `reasoning.ParseModelModifiers` as undefined because the pricing helper imports RelayKit's reasoning package, whose modifier parser had not been included in the selective sync.
- Fixed: restored RelayKit's explicit trailing `@key:value` model modifier parser and its public model modifier types, while keeping opaque provider names and existing suffix parsers intact.
- Preserved: model pricing lookup, Sora/audio and group pricing, plugin routing, payment/accounting, audit logging, and all existing relay conversion behavior.
- Validation: source inspection, `git diff --check`, conflict-marker scan, and import/symbol review. Local Go compilation and Docker/frontend builds were not run; GitHub Actions remains authoritative.
- Commit/push: pending in the current source-level fix commit.

## 2026-09-12 - CI build fix: unify task plugin adaptor contract

- Cause: task relay call sites and the JavaScript task adaptor were using the
  upstream protocol (`ParseResponse`, task-aware polling, and batch polling),
  while the old legacy task adaptor contract still expected `DoResponse` and
  map-based polling. That mixed contract caused CI errors for plugin adaptor
  assignment and task polling method signatures.
- Fixed: kept `channel.TaskAdaptor` on the current plugin protocol, retained
  the pinned-plugin request routing and platform-unavailable errors, and added
  a compile-time assertion covering the JavaScript adaptor's complete task
  contract. Built-in task plugins remain the source for the migrated task
  platforms; the independent OpenAI image task path remains unchanged.
- Preserved: Sora/audio pricing, task plugin state and usage extraction,
  payment/accounting, group discounts, audit logging, advanced custom channel
  routing, and the standalone OpenAI image task implementation.
- Validation: source inspection, interface/call-site review, `git diff --check`,
  and conflict-marker scan. Local Go compilation, frontend builds, and Docker
  builds were not run under the source-only workflow; GitHub Actions remains
  the build authority.
- Commit/push: pending in the current source-level fix commit.

