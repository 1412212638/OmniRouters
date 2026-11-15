const MODELS = [
  "speech-2.8-hd",
  "speech-2.8-turbo",
  "speech-2.6-hd",
  "speech-2.6-turbo",
  "speech-02-hd",
  "speech-02-turbo",
  "speech-01-hd",
  "speech-01-turbo",
];
const MAX_TEXT_CHARACTERS = 10000;
const AUDIO_FORMATS = ["mp3", "wav", "flac", "aac", "pcm"];

function textValue(value, name) {
  if (typeof value !== "string" || !value.trim()) throw new Error(name + " is required");
  const text = value;
  if (Array.from(text).length > MAX_TEXT_CHARACTERS) throw new Error(name + " exceeds 10000 characters");
  return text;
}

function modelValue(value) {
  const model = String(value == null ? "" : value).trim();
  if (!MODELS.includes(model)) throw new Error("unsupported MiniMax speech model");
  return model;
}

function requestModel(value, publicModel, upstreamModel) {
  modelValue(publicModel);
  const supplied = String(value == null || value === "" ? publicModel : value).trim();
  const declared = String(upstreamModel || publicModel).trim();
  if (supplied !== publicModel && supplied !== declared) throw new Error("model must be " + publicModel);
  return publicModel;
}

function finiteNumber(value, name, min, max) {
  if (value === undefined) return undefined;
  const number = Number(value);
  if (!Number.isFinite(number) || number < min || number > max) throw new Error(name + " is out of range");
  return number;
}

function objectValue(value, name) {
  if (!value || typeof value !== "object" || Array.isArray(value)) throw new Error(name + " must be an object");
  return value;
}

function formatValue(value) {
  const format = String(value == null || value === "" ? "mp3" : value).toLowerCase();
  if (!AUDIO_FORMATS.includes(format)) throw new Error("audio format is not supported");
  return format;
}

function nativeRequest(body, expectedModel, upstreamModel) {
  const request = objectValue(body, "request body");
  const model = requestModel(request.model, expectedModel, upstreamModel);
  const text = textValue(request.text, "text");
  if (request.stream === true) throw new Error("stream=true is not supported by the synchronous speech endpoint");
  const voice = request.voice_setting === undefined ? undefined : objectValue(request.voice_setting, "voice_setting");
  const audio = request.audio_setting === undefined ? undefined : objectValue(request.audio_setting, "audio_setting");
  if (voice) {
    if (voice.voice_id !== undefined) textValue(voice.voice_id, "voice_setting.voice_id");
    finiteNumber(voice.speed, "voice_setting.speed", 0.5, 2);
    finiteNumber(voice.vol, "voice_setting.vol", 0, 10);
    finiteNumber(voice.pitch, "voice_setting.pitch", -12, 12);
  }
  if (audio) {
    if (audio.format !== undefined) formatValue(audio.format);
    finiteNumber(audio.sample_rate, "audio_setting.sample_rate", 8000, 96000);
    finiteNumber(audio.bitrate, "audio_setting.bitrate", 8000, 512000);
    finiteNumber(audio.channel, "audio_setting.channel", 1, 2);
  }
  const outputFormat = String(request.output_format == null ? "hex" : request.output_format).toLowerCase();
  if (outputFormat !== "hex" && outputFormat !== "url") throw new Error("output_format must be hex or url");
  return Object.assign({}, request, { model, text, stream: false, output_format: outputFormat });
}

function openAIRequest(ctx) {
  if (!ctx.body || ctx.body.kind !== "json") throw new Error("JSON body required");
  const body = objectValue(ctx.body.value, "request body");
  const model = requestModel(body.model, ctx.model, ctx.upstreamModel);
  const text = textValue(body.input, "input");
  if (body.stream === true) throw new Error("stream=true is not supported by the synchronous speech endpoint");
  const voiceID = textValue(body.voice, "voice");
  const format = formatValue(body.response_format);
  const voiceSetting = Object.assign({}, body.voice_setting === undefined ? {} : objectValue(body.voice_setting, "voice_setting"), { voice_id: voiceID });
  const audioSetting = Object.assign({}, body.audio_setting === undefined ? {} : objectValue(body.audio_setting, "audio_setting"), { format });
  const request = {
    model: ctx.upstreamModel || model,
    text,
    stream: false,
    voice_setting: voiceSetting,
    audio_setting: audioSetting,
    output_format: "hex",
  };
  // Keep provider extensions available to OpenAI-compatible callers without
  // copying the gateway's model/input fields into the upstream request.
  ["pronunciation_dict", "language_boost", "subtitle_enable", "voice_modify"].forEach(function (key) {
    if (body[key] !== undefined) request[key] = body[key];
  });
  return { model, text, request };
}

function taskDataBody(task) {
  return task && task.data && typeof task.data === "object" ? task.data : {};
}

function responseUsage(body) {
  const usage = body && body.extra_info;
  if (!usage || usage.usage_characters === undefined) return {};
  const characters = Number(usage.usage_characters);
  if (!Number.isFinite(characters) || !Number.isSafeInteger(characters) || characters < 0 || characters > MAX_TEXT_CHARACTERS) {
    throw new Error("extra_info.usage_characters is invalid");
  }
  return { characters };
}

function responseBody(resp) {
  const body = resp && resp.body ? resp.body : {};
  const base = body.base_resp || {};
  if (base.status_code !== undefined && Number(base.status_code) !== 0) {
    throw new Error(String(base.status_msg || "MiniMax speech request failed"));
  }
  if (!body.data || typeof body.data.audio !== "string" || !body.data.audio) throw new Error("MiniMax response does not contain audio");
  return body;
}

export const meta = {
  apiVersion: 1,
  key: "minimax-speech",
  name: "MiniMax Speech",
  description: {
    en: "MiniMax synchronous speech synthesis with OpenAI-compatible and native endpoints",
    zh: "MiniMax 同步语音合成，支持 OpenAI 兼容与官方原生接口",
  },
  version: "1.0.0",
  author: { name: "OmniRouters" },
  models: MODELS,
  fetchMode: "per_task",
  upstreams: ["vendor", "new_api"],
  usageSchema: {
    characters: {
      type: "number",
      unit: "count",
      max: MAX_TEXT_CHARACTERS,
      unitLabel: { en: "characters", zh: "字符" },
      description: { en: "Characters synthesized", zh: "合成的字符数" },
    },
  },
  usageExamples: [
    { label: "Short speech", facts: { characters: 100 } },
    { label: "Long speech", facts: { characters: 1000 } },
  ],
  protocols: [
    { name: "openai_audio", supports: ["sync"] },
    { name: "minimax_t2a", supports: ["sync"] },
  ],
};

export function buildSubmitRequest(ctx) {
  const request = ctx.requestBody || {};
  const body = Object.assign({}, request, { model: ctx.upstreamModel || ctx.model || request.model });
  return {
    url: String(ctx.baseUrl || "").replace(/\/+$/, "") + "/v1/t2a_v2",
    method: "POST",
    headers: {
      Authorization: "Bearer " + ctx.apiKey,
      "Content-Type": "application/json",
      Accept: "application/json",
    },
    body: body,
  };
}

export function parseSubmitResponse(ctx, resp) {
  if (resp.statusCode !== undefined && (resp.statusCode < 200 || resp.statusCode >= 300)) throw new Error("MiniMax speech upstream request failed");
  const body = responseBody(resp);
  const taskId = String(body.trace_id || ctx.publicTaskId || "").trim();
  if (!taskId) throw new Error("MiniMax speech response is missing trace_id");
  return { taskId, taskData: body, immediate: { status: "SUCCESS", progress: "100%", taskId } };
}

export function extractUsage(ctx) {
  const body = ctx && ctx.requestBody ? ctx.requestBody : {};
  const text = body.text !== undefined ? body.text : body.input;
  return { characters: Array.from(textValue(text, "text")).length };
}

export function extractUsageOnSubmit(_ctx, body) {
  return responseUsage(body);
}

export function extractUsageOnComplete(_task, _taskResult, body) {
  return responseUsage(body);
}

export function buildQueryRequest() {
  throw new Error("MiniMax speech is synchronous; polling is not supported");
}

export function parseTaskResult() {
  throw new Error("MiniMax speech is synchronous; polling is not supported");
}

export const protocols = {
  openai_audio: {
    decodeRequest: function (ctx) {
      const parsed = openAIRequest(ctx);
      return { kind: "submit", model: parsed.model, requestBody: parsed.request };
    },
    render: function (ctx, task) {
      const body = taskDataBody(task);
      const request = ctx && ctx.body && ctx.body.value ? ctx.body.value : {};
      return {
        audioHex: body.data && body.data.audio,
        format: body.extra_info && body.extra_info.audio_format ? body.extra_info.audio_format : formatValue(request.response_format),
      };
    },
  },
  minimax_t2a: {
    decodeRequest: function (ctx) {
      if (!ctx.body || ctx.body.kind !== "json") throw new Error("JSON body required");
      const body = nativeRequest(ctx.body.value, ctx.model, ctx.upstreamModel);
      return { kind: "submit", model: body.model, requestBody: body };
    },
    render: function (_ctx, task) {
      return taskDataBody(task);
    },
  },
};
