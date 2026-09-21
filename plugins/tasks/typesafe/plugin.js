const MAX_INPUT_TOKENS = 65536;
const MODELS = ["jev-1.13.0", "jev-latest", "jev-preview", "laya"];

export const meta = {
  apiVersion: 1,
  key: "typesafe",
  name: "TypeSafe",
  version: "1.1.1",
  author: { name: "OmniRouters" },
  models: MODELS,
  fetchMode: "per_task",
  upstreams: ["vendor", "new_api"],
  routes: [{ method: "POST", path: "/typesafe/v1/systemone", type: "submit", decode: "decodeSystemOne", render: "renderSystemOne", retainResult: false }],
  usageSchema: { input_tokens: { type: "number", unit: "token" } },
  usageExamples: [
    { label: "Short request", facts: { input_tokens: 402 } },
    { label: "Maximum context", facts: { input_tokens: 65536 } }
  ]
};

function declaredModel(model) {
  if (!MODELS.includes(model)) throw new Error("Unsupported TypeSafe model");
  return model;
}

function systemOneRequest(ctx) {
  const model = declaredModel(ctx.model);
  const body = ctx.requestBody;
  return { model: ctx.upstreamModel || model, state: body.state, questions: body.questions };
}

export function decodeSystemOne(ctx) {
  if (!ctx.body || ctx.body.kind !== "json") throw new Error("JSON body required");
  const body = ctx.body.value;
  if (!body || typeof body !== "object" || Array.isArray(body)) throw new Error("Request body must be an object");
  const model = declaredModel(body.model);
  if (body.state === undefined || body.questions === undefined) throw new Error("state and questions are required");
  return { kind: "submit", requestBody: { model, state: body.state, questions: body.questions }, model };
}

export function buildSubmitRequest(ctx) {
  const body = systemOneRequest(ctx);
  let baseUrl = ctx.baseUrl.replace(/\/+$/, "");
  if (baseUrl.endsWith("/v1")) baseUrl = baseUrl.slice(0, -3);
  return { url: baseUrl + "/v1/systemone", method: "POST", headers: { Authorization: "Bearer " + ctx.apiKey, "Content-Type": "application/json", Accept: "application/json" }, body };
}

// Required by the generic task-plugin contract; native routes normally do not query.
export function buildQueryRequest(ctx) {
  throw new Error("SystemOne is synchronous; polling is not supported");
}

export function parseSubmitResponse(ctx, response) {
  if (response.statusCode !== undefined && (response.statusCode < 200 || response.statusCode >= 300)) throw new Error("SystemOne upstream request failed");
  const body = response.body || response;
  if (!body || typeof body !== "object" || !body.answers) throw new Error("Invalid TypeSafe response");
  const taskId = body.id || ctx.publicTaskId;
  if (!taskId) throw new Error("Missing SystemOne response id");
  return { taskId, taskData: body, immediate: { status: "SUCCESS", progress: "100%" } };
}

// The task-plugin registry requires this hook even for native synchronous routes.
export function parseTaskResult() { throw new Error("SystemOne is synchronous; polling is not supported"); }

export function extractUsage() { return { input_tokens: MAX_INPUT_TOKENS }; }

export function extractUsageOnComplete(ctx, result, body) {
  const tokens = body && body.usage && body.usage.input_tokens;
  if (typeof tokens !== "number" || !Number.isFinite(tokens) || tokens < 0 || tokens > MAX_INPUT_TOKENS || !Number.isInteger(tokens)) {
    throw new Error("Invalid SystemOne input_tokens usage");
  }
  return { input_tokens: tokens };
}

export function renderSystemOne(ctx, task) { return task.data; }

export const native = { decodeSystemOne, renderSystemOne };
