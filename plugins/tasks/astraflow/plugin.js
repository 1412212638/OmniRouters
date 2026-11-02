// AstraFlow video enhancement task plugin.
// Reserve ten seconds up front and reconcile against upstream duration.
const MODEL = "ai-mediakit-video-enhancement-std";
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

export const meta = {
  apiVersion: 1,
  key: "astraflow",
  name: "AstraFlow",
  icon: "Volcengine.Color",
  description: {
    en: "AstraFlow video enhancement",
    zh: "AstraFlow 视频增强",
  },
  version: "1.0.0",
  author: { name: "QuantumNous" },
  models: [MODEL],
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
  protocols: [{ name: "openai_responses", supports: ["stream", "sync", "background"] }, "openai_video"],
  routes: [
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
  const taskId = body.output && trimmed(body.output.task_id);
  if (!taskId) throw new Error("missing output.task_id");
  return { taskId: taskId, taskData: body };
}

export function extractUsage(ctx) {
  return usageFacts(ctx.requestBody || {});
}

export function extractUsageOnComplete(_task, _taskResult, body) {
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
  return {
    url: ctx.baseUrl + "/v1/tasks/status?task_id=" + encodeURIComponent(ctx.taskId),
    method: "GET",
    headers: { Accept: "application/json", Authorization: "Bearer " + ctx.apiKey },
  };
}

export function parseTaskResult(ctx, body) {
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
    return { kind: "submit", model: ctx.model, action: "video_enhancement", requestBody: body };
  },
  render: function (ctx, task) {
    const statuses = { QUEUED: "queued", IN_PROGRESS: "in_progress", SUCCESS: "completed", FAILURE: "failed" };
    return { id: task.task_id, object: "video", model: ctx.model || MODEL, status: statuses[task.status] || "unknown", progress: Number(String(task.progress || "0").replace("%", "")), created_at: Number(task.created_at || 0) };
  },
};
