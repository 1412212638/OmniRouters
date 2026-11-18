# Seedance 2.5 billing and callbacks

The Doubao task plugin supports the native submit/query endpoints at
`/api/v3/contents/generations/tasks` and the existing `/doubao/api/v3/...` aliases.
This update retains provider validation of business parameters. `service_tier`
and `priority` remain passthrough fields. Web search does not introduce an
additional charge.

## Pricing and Draft

- Both `doubao-seedance-2.5` and `doubao-seedance-2-5-260628` expose 480p, 720p,
  1080p and 4k in their task usage schema. Flash retains 480p/720p.
- Configure the new 4k prices, with and without reference-video input, in task
  pricing. Saved expressions and prices are not rewritten automatically.
- Submission estimates use requested resolution and duration. Automatic or
  unspecified duration on recognized 2.5 models reserves up to 30 seconds;
  completion settles using `usage.completion_tokens`, falling back to
  `usage.total_tokens`, and the actual returned resolution.
- Draft previews and final renders are separate tasks, each with its own
  reservation and settlement. A final request using `draft_task` inherits only
  the reference-video billing selector from the owned original task. Its output
  resolution, actual usage and current pricing snapshot determine its own cost.
  The provider inherits prompt/material/generation parameters; the gateway does
  not add them to the final request.
- New tasks preserve the input selector in private plugin state as well as in
  expression billing snapshots. Older Drafts with neither source cannot be
  billed reliably and are rejected before reservation instead of guessing a
  price tier; generate a new Draft in that case. Existing Drafts with a saved
  expression usage snapshot continue to work.
- Existing group/customer multipliers, organization/subscription/wallet funding,
  charge reconciliation and failure refunds continue through the existing task
  billing pipeline. This update does not change configured unit prices.

## Customer callback_url

Include an HTTP(S) `callback_url` in the native JSON request (or in video
metadata when using a compatible endpoint):

```json
{
  "model": "doubao-seedance-2.5",
  "content": [{ "type": "text", "text": "A cat walking in the garden" }],
  "resolution": "720p",
  "duration": 5,
  "callback_url": "https://customer.example/video-result"
}
```

Configure the public server address, or the existing custom callback address,
to the externally reachable gateway URL. The gateway replaces the upstream
callback URL with `/api/v3/task-notifications/{public_task_id}?token=...`. The
random token and customer URL are stored privately in the shared database. All
application nodes must use the same database, and the reverse proxy must allow
POST requests to that path. No API-key header is required on that notification
receiver; it validates the per-task token and expected upstream task ID.

Provider notifications are wake-up signals, never evidence of billable usage.
The gateway fetches the task using the original provider credentials and feeds
the confirmed response into the existing status CAS and settlement pipeline.
Invalid tokens/IDs cannot alter billing. Notifications arriving before durable
submission and initial settlement receive a non-2xx response so the provider
can retry; normal polling remains a fallback.

Customer notifications are JSON POST requests using the native query response
shape, with `id` replaced by the gateway's `task_...` ID and `model` set to the
public requested model. Other provider result fields, including `content` and
`usage`, are retained. The gateway relay URL/token, API keys, cookies and
private billing fields are not included. Customer URLs follow the existing
SSRF settings, credentials/fragments in the URL are rejected, and delivery
redirects are not followed. HTTP 2xx acknowledges delivery.

A persistent outbox retries unsuccessful deliveries with exponential backoff
(30 seconds initially, capped at 64 minutes), coordinated across nodes using a
database lease. The existing task-polling system job also dispatches the outbox
after polling, so delivery can recover after a process restart or missing
provider notification. Successful delivery suppresses repeat sends for that
stored status. Rapid intermediate states may be coalesced to the latest
confirmed state. Delivery is at least once: a receiver should handle duplicate
notifications idempotently by task ID and status, especially if its response was
lost or the gateway restarted before recording acknowledgement. Callback
delivery never charges or refunds funds.

Outbox delivery retries are retained for 15 days after submission; records are
cleaned up after 16 days. Each customer request has a five-second timeout, and
scheduled dispatch uses eight workers with a ten-second run budget. Polling and
notification latency still depend on the configured task polling schedule,
system rate limits, network availability and the provider's query endpoint.
