# Task Plugin API v1

Task plugins are single-file synchronous ECMAScript modules. The host owns
HTTP connections, persistence, retries, polling, billing, settlement, and
refunds. Plugins transform requests and responses and report usage facts.

Every plugin exports `meta`, `buildSubmitRequest`, `parseSubmitResponse`, and
`parseTaskResult`. Per-task plugins also export `buildQueryRequest`; batch
plugins export `buildBatchQueryRequest` and `parseBatchResult`.

Polling hooks receive the persisted task context and upstream HTTP response
metadata. State that must survive polling rounds belongs in `state`, not the
public task data snapshot. The host bounds persisted plugin state at 1 MiB and
keeps the previous state when a hook omits it.

Use the local CLI to validate a plugin and replay deterministic fixtures:

```sh
new-api plugin lint plugin.js
new-api plugin test plugin.js --fixture golden.json
```

Unknown provider states must return `UNKNOWN`, never an `IN_PROGRESS` fallback.
This allows the host to count parser failures and eventually fail and refund a
task instead of leaving it pending forever.
