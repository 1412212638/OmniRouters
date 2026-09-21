const MAX_INPUT_TOKENS = 65536;
const MODELS = ["jev-1.13.0", "jev-latest", "jev-preview", "laya"];

export const meta = {
  apiVersion: 1,
  key: "typesafe",
  name: "TypeSafe",
  version: "1.1.0",
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
  return { model: ctx.upstreamModel || model, state: ctx.body.state, questions: ctx.body.questions };
}

export function decodeSystemOne(ctx) {
  const request = systemOneRequest(ctx);
  return { requestBody: request, model: request.model };
}

export function buildSubmitRequest(ctx) {
  const body = systemOneRequest(ctx);
  let baseUrl = ctx.baseUrl.replace(/\/+$/, "");
  if (baseUrl.endsWith("/v1")) baseUrl = baseUrl.slice(0, -3);
  const prefix = ctx.upstream && ctx.upstream.kind === "new_api" ? "/typesafe" : "";
  return { url: baseUrl + prefix + "/v1/systemone", method: "POST", headers: { Authorization: "Bearer " + ctx.apiKey, "Content-Type": "application/json", Accept: "application/json" }, body };
}

export function parseSubmitResponse(ctx, response) {
  const body = response.body || response;
  if (!body || typeof body !== "object" || !body.answers) throw new Error("Invalid TypeSafe response");
  const usage = body.usage || {};
  const inputTokens = Math.max(0, Math.min(MAX_INPUT_TOKENS, Number(usage.input_tokens) || 0));
  return { taskId: body.id || "systemone", status: "SUCCESS", data: body, usage: { input_tokens: inputTokens } };
}

// The task-plugin registry requires this hook even for native synchronous routes.
export function parseTaskResult(ctx, response) { return parseSubmitResponse(ctx, response); }

export function renderSystemOne(ctx, task) { return task.data; }
