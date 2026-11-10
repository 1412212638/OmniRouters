// AstraFlow video enhancement task plugin.
// Reserve ten seconds up front and reconcile against upstream duration.
const MODEL = "ai-mediakit-video-enhancement-std";
const OMNI_MODEL = "gemini-omni-1.2-flash";
const OMNI_ACTION = "omni_video";
const OMNI_TOKEN_LIMIT = 2147483647;
const OMNI_SCHEMA = {
  input_tokens: { type: "number", unit: "token", description: { en: "Input tokens", zh: "输入词元" } },
  text_output_tokens: { type: "number", unit: "token", description: { en: "Text output tokens", zh: "文本输出词元" } },
  video_output_tokens: { type: "number", unit: "token", description: { en: "Video output tokens", zh: "视频输出词元" } },
};
const RESOLUTIONS = ["720p", "1080p", "2k", "4k", "6k", "8k"];
const FPS_OPTIONS = [30, 60, 120];
const PRECHARGE_SECONDS = 10;

function trimmed(value) {
  return String(value == null ? "" : value).trim();
}

function objectValue(value, name) {
  if (!value || typeof value !== "object" || Array.isArray(value)) throw new Error(name + " must be an object");
  return value;
}

function positiveDuration(value, name) {
  const number = Number(value);
  if (!Number.isFinite(number) || number <= 0 || number > 3600) throw new Error(name + " must be between 0 and 3600 seconds");
  return number;
}

function fpsValue(parameters) {
  if (!Object.prototype.hasOwnProperty.call(parameters, "fps")) throw new Error("parameters.fps is required");
  const fps = Number(parameters.fps);
  if (!Number.isFinite(fps) || !FPS_OPTIONS.includes(fps)) throw new Error("parameters.fps must be one of 30, 60, or 120");
  return fps;
}

function resolutionValue(parameters) {
  const resolution = trimmed(parameters.resolution);
  if (!RESOLUTIONS.includes(resolution)) throw new Error("parameters.resolution must be one of " + RESOLUTIONS.join(", "));
  return resolution;
}

function requestParts(body) {
  const input = objectValue(body.input, "input");
  const parameters = objectValue(body.parameters, "parameters");
  const videoURL = trimmed(input.video_url);
  if (!/^https?:\/\//i.test(videoURL)) throw new Error("input.video_url must be an HTTP or HTTPS URL");
  return { input: input, parameters: parameters, videoURL: videoURL, resolution: resolutionValue(parameters), fps: fpsValue(parameters) };
}

function estimateDuration(_body) {
  return PRECHARGE_SECONDS;
}

function usageFacts(body) {
  const parts = requestParts(body);
  return { seconds: estimateDuration(body), resolution: parts.resolution, fps: String(parts.fps) };
}

function isOmni(ctx) {
  return ctx.action === OMNI_ACTION || ctx.model === OMNI_MODEL || ctx.upstreamModel === OMNI_MODEL || (ctx.requestBody && ctx.requestBody.model === OMNI_MODEL);
}

function validateOmni(body) {
  objectValue(body, "request body");
  if (!(typeof body.input === "string" && body.input.trim()) && !(Array.isArray(body.input) && body.input.length)) throw new Error("input must be non-empty text or an array");
  const config = objectValue(body.generation_config, "generation_config");
  const video = objectValue(config.video_config, "generation_config.video_config");
  if (!["text_to_video", "image_to_video", "reference_to_video", "edit"].includes(video.task)) throw new Error("unsupported generation_config.video_config.task");
  if (body.background !== undefined && typeof body.background !== "boolean") throw new Error("background must be a boolean");
  if (body.response_format !== undefined) {
    const format = objectValue(body.response_format, "response_format");
    if (format.type !== "video") throw new Error("response_format.type must be video");
    if (format.aspect_ratio !== undefined && !["16:9", "9:16"].includes(format.aspect_ratio)) throw new Error("response_format.aspect_ratio must be 16:9 or 9:16");
  }
  // Walk prior interaction steps too, without flattening or rewriting them.
  let images = 0, videos = 0, mediaBytes = 0, nodes = 0;
  function visit(value, depth) {
    if (++nodes > 4096 || depth > 32) throw new Error("input is too complex");
    if (!value || typeof value !== "object") return;
    if (Array.isArray(value)) { value.forEach(function (item) { visit(item, depth + 1); }); return; }
    if (value.type === "image" || value.type === "video") {
      if (typeof value.data !== "string" || !value.data || !/^[A-Za-z0-9+/]*={0,2}$/.test(value.data) || value.data.length % 4 === 1) throw new Error("media data must be raw Base64");
      if (typeof value.mime_type !== "string" || !value.mime_type.startsWith(value.type + "/")) throw new Error("media mime_type must match its type");
      if (value.type === "image") images++; else videos++;
      mediaBytes += value.data.length;
      if (images > 10 || videos > 3 || mediaBytes > 48 * 1024 * 1024) throw new Error("input exceeds 10 images, 3 videos, or 48 MiB of Base64 media");
    }
    Object.keys(value).forEach(function (key) { if (key !== "data") visit(value[key], depth + 1); });
  }
  visit(body.input, 0);
  return body;
}

function tokenCount(value, name) {
  if (typeof value !== "number" || !Number.isSafeInteger(value) || value < 0 || value > OMNI_TOKEN_LIMIT) throw new Error(name + " must be a non-negative bounded integer");
  return value;
}

function omniUsage(body) {
  const usage = objectValue(body.usage, "usage");
  const input = tokenCount(usage.total_input_tokens, "usage.total_input_tokens");
  const output = tokenCount(usage.total_output_tokens, "usage.total_output_tokens");
  if (!Array.isArray(usage.output_tokens_by_modality)) throw new Error("usage.output_tokens_by_modality is required; cannot infer video tokens");
  let text = 0, video = 0, reported = 0;
  usage.output_tokens_by_modality.forEach(function (item) {
    objectValue(item, "output token modality");
    const count = tokenCount(item.tokens, "modality.tokens");
    reported = tokenCount(reported + count, "modality total");
    if (item.modality === "text") text += count;
    else if (item.modality === "video") video += count;
    else if (count > 0) throw new Error("unsupported billable output modality");
  });
  if (reported !== output) throw new Error("output modality tokens must equal total_output_tokens");
  return { input_tokens: input, text_output_tokens: text, video_output_tokens: video,
    completionTokens: output, totalTokens: tokenCount(usage.total_tokens, "usage.total_tokens") };
}

function omniResult(body) {
  const statuses = { queued: "QUEUED", in_progress: "IN_PROGRESS", processing: "IN_PROGRESS", completed: "SUCCESS", failed: "FAILURE" };
  const status = statuses[body.status];
  if (!status) return { status: "UNKNOWN", reason: "unrecognized interaction status" };
  const result = { status: status, progress: status === "SUCCESS" || status === "FAILURE" ? "100%" : "0%" };
  if (status === "FAILURE") result.reason = trimmed(body.error && body.error.message) || "video generation failed";
  return result;
}

function omniVideo(body) {
  const steps = body && Array.isArray(body.steps) ? body.steps : [];
  for (let i = steps.length - 1; i >= 0; i--) {
    if (steps[i].type !== "model_output" || !Array.isArray(steps[i].content)) continue;
    const video = steps[i].content.find(function (item) { return item.type === "video" && typeof item.data === "string" && item.data; });
    if (video) return video;
  }
  return null;
}

export const meta = {
  apiVersion: 1,
  key: "astraflow",
  name: "AstraFlow",
  icon: "Volcengine.Color",
  description: {
    en: "AstraFlow video enhancement and Omni video generation",
    zh: "AstraFlow 视频增强与 Omni 视频生成",
  },
  version: "1.1.0",
  author: { name: "QuantumNous" },
  models: [MODEL, OMNI_MODEL],
  fetchMode: "per_task",
  upstreams: ["vendor", "new_api"],
  usageSchema: {
    seconds: {
      type: "number",
      unit: "second",
      description: { en: "Input video duration in seconds", zh: "输入视频时长（秒）" },
    },
    resolution: {
      enum: RESOLUTIONS,
      description: { en: "Target output resolution", zh: "目标输出分辨率" },
    },
    fps: {
      enum: FPS_OPTIONS.map(String),
      description: { en: "Target output frames per second", zh: "目标输出帧率" },
    },
  },
  usageExamples: [
    { label: "1080p · 30 fps · 15s", facts: { seconds: 15, resolution: "1080p", fps: "30" } },
    { label: "1080p · 60 fps · 15s", facts: { seconds: 15, resolution: "1080p", fps: "60" } },
  ],
  usageProfiles: [
    { models: [MODEL], schema: {
      seconds: { type: "number", unit: "second", description: { en: "Input video duration in seconds", zh: "输入视频时长（秒）" } },
      resolution: { enum: RESOLUTIONS, description: { en: "Target output resolution", zh: "目标输出分辨率" } },
      fps: { enum: FPS_OPTIONS.map(String), description: { en: "Target output frames per second", zh: "目标输出帧率" } },
    }, examples: [{ label: "1080p · 30 fps · 15s", facts: { seconds: 15, resolution: "1080p", fps: "30" } }] },
    { models: [OMNI_MODEL], schema: OMNI_SCHEMA, examples: [
      { label: "Video output", facts: { input_tokens: 19, text_output_tokens: 0, video_output_tokens: 57920 } },
      { label: "Text and video output", facts: { input_tokens: 57815, text_output_tokens: 100, video_output_tokens: 57920 } },
    ] },
  ],
  protocols: [{ name: "openai_responses", models: [MODEL], supports: ["stream", "sync", "background"] }, "openai_video"],
  routes: [
    { method: "POST", path: "/v1beta/models/interactions", type: "submit", models: [OMNI_MODEL], decode: "createInteraction", render: "interaction" },
    { method: "GET", path: "/v1beta/models/interactions/:id", type: "query", taskIdParam: "id", render: "interaction" },
    {
      method: "POST",
      path: "/v1/tasks/submit",
      type: "submit",
      models: [MODEL],
      decode: "createTask",
      render: "taskCreated",
    },
    {
      method: "GET",
      path: "/v1/tasks/status",
      type: "query",
      taskIdQuery: true,
      render: "taskStatus",
    },
  ],
};

export const native = {
  createInteraction: function (ctx) {
    if (!ctx.body || ctx.body.kind !== "json") throw new Error("JSON body required");
    const body = validateOmni(ctx.body.value);
    if (body.model !== OMNI_MODEL) throw new Error("model must be " + OMNI_MODEL);
    return { kind: "submit", model: body.model, action: OMNI_ACTION, requestBody: body };
  },
  interaction: function (_ctx, task) {
    const body = task.data && typeof task.data === "object" ? task.data : {};
    const statuses = { QUEUED: "queued", IN_PROGRESS: "in_progress", SUCCESS: "completed", FAILURE: "failed" };
    const result = Object.assign({}, body, { id: task.task_id, model: task.model || OMNI_MODEL, status: statuses[task.status] || body.status });
    if (task.status === "FAILURE" && !result.error) result.error = { code: "video_generation_failed", message: task.fail_reason || "video generation failed" };
    return result;
  },
  createTask: function (ctx) {
    if (!ctx.body || ctx.body.kind !== "json") throw new Error("JSON body required");
    const body = ctx.body.value;
    if (!body || typeof body !== "object" || Array.isArray(body)) throw new Error("request body must be an object");
    const model = trimmed(body.model);
    if (model !== MODEL) throw new Error("model must be " + MODEL);
    requestParts(body);
    estimateDuration(body);
    return { kind: "submit", model: model, action: "video_enhancement", requestBody: body };
  },
  taskCreated: function (ctx, task) {
    const body = task.data && typeof task.data === "object" ? task.data : {};
    return { output: { task_id: task.task_id }, request_id: body.request_id || "" };
  },
  taskStatus: function (ctx, task) {
    const body = task.data && typeof task.data === "object" ? task.data : {};
    const output = Object.assign({}, body.output || {}, { task_id: task.task_id });
    return Object.assign({}, body, { output: output });
  },
};

export function buildSubmitRequest(ctx) {
  const request = ctx.requestBody || {};
  if (isOmni(ctx)) {
    validateOmni(request);
    return { url: ctx.baseUrl + "/v1beta/models/interactions", method: "POST",
      headers: { "Content-Type": "application/json", Accept: "application/json", Authorization: "Bearer " + ctx.apiKey },
      body: Object.assign({}, request, { model: ctx.upstreamModel || OMNI_MODEL }), action: OMNI_ACTION };
  }
  const parts = requestParts(request);
  const body = {
    model: ctx.upstreamModel || MODEL,
    input: parts.input,
    parameters: parts.parameters,
  };
  return {
    url: ctx.baseUrl + "/v1/tasks/submit",
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "application/json", Authorization: "Bearer " + ctx.apiKey },
    body: body,
    action: "video_enhancement",
  };
}

export function parseSubmitResponse(ctx, resp) {
  const body = resp.body || {};
  if (isOmni(ctx)) {
    if (!trimmed(body.id)) throw new Error("missing interaction id");
    const result = omniResult(body);
    if (result.status === "UNKNOWN") throw new Error(result.reason);
    return { taskId: trimmed(body.id), taskData: body, immediate: result.status === "SUCCESS" || result.status === "FAILURE" ? result : undefined };
  }
  const taskId = body.output && trimmed(body.output.task_id);
  if (!taskId) throw new Error("missing output.task_id");
  return { taskId: taskId, taskData: body };
}

export function extractUsage(ctx) {
  if (isOmni(ctx)) {
    validateOmni(ctx.requestBody || {});
    // A reservation, not an asserted duration or actual usage. Reconciled once on completion.
    return { input_tokens: 65536, text_output_tokens: 4096, video_output_tokens: 65536 };
  }
  return usageFacts(ctx.requestBody || {});
}

export function extractUsageOnComplete(_task, _taskResult, body) {
  if (isOmni(_task)) return omniUsage(body);
  const output = body && body.output ? body.output : {};
  const usage = body && body.usage ? body.usage : {};
  const duration = usage.output_video_duration ?? usage.duration;
  const facts = {};
  if (duration !== undefined && duration !== null) facts.seconds = positiveDuration(duration, "usage.duration");
  if (usage.size !== undefined && RESOLUTIONS.includes(trimmed(usage.size))) facts.resolution = trimmed(usage.size);
  if (usage.output_video_fps !== undefined && usage.output_video_fps !== null) {
    const fps = Number(usage.output_video_fps);
    if (!Number.isFinite(fps) || !FPS_OPTIONS.includes(fps)) throw new Error("usage.output_video_fps must be one of 30, 60, or 120");
    facts.fps = String(fps);
  }
  return facts;
}

export function buildQueryRequest(ctx) {
  if (isOmni(ctx)) return { url: ctx.baseUrl + "/v1beta/models/interactions/" + encodeURIComponent(ctx.taskId), method: "GET", headers: { Accept: "application/json", Authorization: "Bearer " + ctx.apiKey } };
  return {
    url: ctx.baseUrl + "/v1/tasks/status?task_id=" + encodeURIComponent(ctx.taskId),
    method: "GET",
    headers: { Accept: "application/json", Authorization: "Bearer " + ctx.apiKey },
  };
}

export function parseTaskResult(ctx, body) {
  if (isOmni(ctx)) return omniResult(body);
  const output = body && body.output ? body.output : {};
  const status = String(output.task_status || "").toLowerCase();
  const states = {
    pending: "QUEUED",
    running: "IN_PROGRESS",
    success: "SUCCESS",
    failure: "FAILURE",
    expired: "FAILURE",
    cancelled: "FAILURE",
    unknown: "UNKNOWN",
  };
  if (!states[status]) return { status: "UNKNOWN", reason: "unrecognized task status: " + String(output.task_status || "") };
  const result = { status: states[status] };
  if (states[status] === "SUCCESS" && Array.isArray(output.urls) && output.urls.length && trimmed(output.urls[0])) {
    result.url = trimmed(output.urls[0]);
    result.remoteUrl = trimmed(output.urls[0]);
  }
  if (states[status] === "FAILURE") result.reason = trimmed(output.error_message || body.message || "video enhancement failed");
  return result;
}

export function listArtifacts(task) {
  return task.status === "SUCCESS" ? [{ key: "video", type: "video" }] : [];
}

export function buildContentRequest(ctx) {
  if (ctx.artifactKey !== "video") throw new Error("artifact_not_found");
  if (ctx.action === OMNI_ACTION) {
    const video = omniVideo(ctx.data);
    if (!video) throw new Error("artifact_not_found");
    return { inlineBase64: video.data, mimeType: video.mime_type || "video/mp4", method: ctx.clientRequest.method };
  }
  const body = (ctx.data && ctx.data.output) || {};
  const url = Array.isArray(body.urls) ? trimmed(body.urls[0]) : "";
  if (!url) throw new Error("artifact_not_found");
  return { url: url, method: ctx.clientRequest.method, credentialless: true };
}

export const protocols = {
  openai_responses: {
    decodeRequest: function (ctx) {
      if (!ctx.body || ctx.body.kind !== "json") throw new Error("JSON body required");
      const body = ctx.body.value;
      if (!body || typeof body !== "object" || Array.isArray(body)) throw new Error("request body must be an object");
      const model = trimmed(body.model || ctx.model);
      if (model !== MODEL) throw new Error("model must be " + MODEL);
      return { kind: "submit", model: model, action: "video_enhancement", requestBody: body };
    },
    renderEvents: function (ctx, task, previousState) {
      const status = String(task.status || "UNKNOWN").toUpperCase();
      if (status === "SUCCESS") return { events: [{ type: "output", data: "Video enhancement completed." }], state: { status: status }, done: true };
      if (status === "FAILURE") return { events: [{ type: "error", code: "task_failed", message: task.fail_reason || "video enhancement failed" }], state: { status: status }, done: true };
      if (previousState && previousState.status === status) return { events: [], state: { status: status }, done: false };
      return { events: [{ type: "progress", message: status.toLowerCase() }], state: { status: status }, done: false };
    },
    renderFinal: function (_ctx, _task) {
      return { output: [{ type: "message", status: "completed", role: "assistant", content: [{ type: "output_text", text: "Video enhancement completed.", annotations: [], logprobs: [] }] }], metadata: { vendor: "astraflow" } };
    },
  },
};

protocols.openai_video = {
  decodeRequest: function (ctx) {
    if (!ctx.body || ctx.body.kind !== "json") throw new Error("JSON body required");
    const body = ctx.body.value;
    if (!body || typeof body !== "object" || Array.isArray(body)) throw new Error("request body must be an object");
    if (ctx.model === OMNI_MODEL || body.model === OMNI_MODEL) {
      // The shared video endpoint accepts the same structured interaction input.
      validateOmni(body);
      return { kind: "submit", model: ctx.model || body.model, action: OMNI_ACTION, requestBody: body };
    }
    return { kind: "submit", model: ctx.model, action: "video_enhancement", requestBody: body };
  },
  render: function (ctx, task) {
    const statuses = { QUEUED: "queued", IN_PROGRESS: "in_progress", SUCCESS: "completed", FAILURE: "failed" };
    const result = { id: task.task_id, object: "video", model: task.model || ctx.model || MODEL, status: statuses[task.status] || "unknown", progress: Number(String(task.progress || "0").replace("%", "")), created_at: Number(task.created_at || 0) };
    if (task.status === "SUCCESS") result.metadata = { url: "/v1/videos/" + encodeURIComponent(task.task_id) + "/content" };
    if (task.status === "FAILURE") result.error = { code: "video_generation_failed", message: task.fail_reason || "video generation failed" };
    return result;
  },
};
