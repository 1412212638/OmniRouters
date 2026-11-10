import assert from 'node:assert/strict';
import { test } from 'node:test';
import * as plugin from './tasks/astraflow/plugin.js';

const model = 'gemini-omni-1.2-flash';
const context = { model, upstreamModel: model, action: 'omni_video', baseUrl: 'https://provider.example', apiKey: 'fixture-key' };
const request = { model, input: '生成故宫的视频', generation_config: { video_config: { task: 'text_to_video' } }, response_format: { type: 'video', aspect_ratio: '16:9' }, background: true };
const usage = { total_input_tokens: 19, total_output_tokens: 57930, total_tokens: 58295, total_thought_tokens: 346,
  output_tokens_by_modality: [{ modality: 'text', tokens: 10 }, { modality: 'video', tokens: 57920 }] };

test('native paths, model scopes, and pricing profiles preserve enhancement', () => {
  assert.deepEqual(plugin.meta.models, ['ai-mediakit-video-enhancement-std', model]);
  assert.equal(plugin.meta.routes.find(r => r.path === '/v1beta/models/interactions/:id').taskIdParam, 'id');
  assert.equal(plugin.meta.usageProfiles.find(p => p.models.includes(model)).schema.video_output_tokens.unit, 'token');
  assert.equal(plugin.meta.usageProfiles.find(p => p.models.includes('ai-mediakit-video-enhancement-std')).schema.seconds.unit, 'second');
  const facts = plugin.extractUsage({ requestBody: { model: 'ai-mediakit-video-enhancement-std', input: { video_url: 'https://example.com/input.mp4' }, parameters: { resolution: '1080p', fps: 30 } } });
  assert.deepEqual(facts, { seconds: 10, resolution: '1080p', fps: '30' });
  assert.equal(plugin.parseTaskResult({}, { output: { task_status: 'Success', urls: ['https://example.com/out.mp4'] } }).url, 'https://example.com/out.mp4');
});

test('all four tasks preserve native input and explicit background=false', () => {
  for (const task of ['text_to_video', 'image_to_video', 'reference_to_video', 'edit']) {
    const body = { ...request, background: false, generation_config: { video_config: { task } },
      input: [{ type: 'thought', signature: 'opaque-signature' }, { type: 'model_output', content: [{ type: 'video', data: 'dmlkZW8=', mime_type: 'video/mp4' }] }, { type: 'user_input', content: [{ type: 'text', text: '继续' }] }] };
    const decoded = plugin.native.createInteraction({ body: { kind: 'json', value: body } });
    const upstream = plugin.buildSubmitRequest({ ...context, requestBody: decoded.requestBody });
    assert.equal(upstream.url, 'https://provider.example/v1beta/models/interactions');
    assert.equal(upstream.headers.Authorization, 'Bearer fixture-key');
    assert.deepEqual(upstream.body, body);
    const defaultBody = { ...body };
    delete defaultBody.background;
    assert.equal(Object.hasOwn(plugin.buildSubmitRequest({ ...context, requestBody: defaultBody }).body, 'background'), false);
  }
  assert.throws(() => plugin.native.createInteraction({ body: { kind: 'json', value: { ...request, model: 'gemini-omni-1.1-flash-preview' } } }));
});

test('async states and immediate sync results use the same task lifecycle', () => {
  const parsed = plugin.parseSubmitResponse(context, { body: { id: 'private-id', status: 'in_progress' } });
  assert.equal(parsed.taskId, 'private-id');
  assert.equal(parsed.immediate, undefined);
  assert.equal(plugin.parseSubmitResponse(context, { body: { id: 'private-id', status: 'completed', usage } }).immediate.status, 'SUCCESS');
  assert.equal(plugin.parseSubmitResponse(context, { body: { id: 'private-id', status: 'failed' } }).immediate.status, 'FAILURE');
  for (const state of ['processing', 'in_progress']) assert.equal(plugin.parseTaskResult(context, { status: state }).status, 'IN_PROGRESS');
  assert.equal(plugin.buildQueryRequest({ ...context, taskId: 'private/id' }).url, 'https://provider.example/v1beta/models/interactions/private%2Fid');
});

test('usage keeps text and video separate; malformed/missing modalities never infer video', () => {
  assert.deepEqual(plugin.extractUsageOnComplete(context, {}, { usage }), {
    input_tokens: 19, text_output_tokens: 10, video_output_tokens: 57920, completionTokens: 57930, totalTokens: 58295,
  });
  const invalid = [undefined, [], [{ modality: 'video', tokens: -1 }], [{ modality: 'video', tokens: 2 ** 63 }], [{ modality: 'video', tokens: Infinity }], [{ modality: 'video', tokens: '57930' }], [{ modality: 'audio', tokens: 57930 }], [{ modality: 'video', tokens: 1.5 }]];
  for (const modalities of invalid) assert.throws(() => plugin.extractUsageOnComplete(context, {}, { usage: { ...usage, output_tokens_by_modality: modalities } }));
  for (const count of [-1, NaN, Infinity, 2 ** 63, '19']) assert.throws(() => plugin.extractUsageOnComplete(context, {}, { usage: { ...usage, total_input_tokens: count } }));
  assert.throws(() => plugin.extractUsageOnComplete(context, {}, {}));
});

test('native result retains steps but publishes gateway id and model; preview uses output media', () => {
  const steps = [{ type: 'thought', signature: 'keep' }, { type: 'model_output', content: [{ type: 'video', mime_type: 'video/mp4', data: 'dmlkZW8=' }] }];
  const data = { id: 'private-id', model: 'provider-model', status: 'completed', steps, usage };
  const result = plugin.native.interaction({}, { task_id: 'public-id', model, status: 'SUCCESS', data });
  assert.equal(result.id, 'public-id');
  assert.equal(result.model, model);
  assert.deepEqual(result.steps, steps);
  assert.deepEqual(result.usage, usage);
  assert.deepEqual(plugin.buildContentRequest({ ...context, data, artifactKey: 'video', clientRequest: { method: 'GET' } }), { inlineBase64: 'dmlkZW8=', mimeType: 'video/mp4', method: 'GET' });
  assert.equal(plugin.native.interaction({}, { task_id: 'public-id', status: 'FAILURE', data: { status: 'in_progress' } }).status, 'failed');
});

test('rejects invalid tasks, ratios, media and excessive reference counts', () => {
  for (const body of [
    { ...request, background: 0 }, { ...request, input: '' },
    { ...request, generation_config: { video_config: { task: 'unsupported' } } },
    { ...request, response_format: { type: 'video', aspect_ratio: '1:1' } },
    { ...request, input: [{ type: 'video', data: 'https://example.com/video.mp4', mime_type: 'video/mp4' }] },
    { ...request, input: Array.from({ length: 11 }, () => ({ type: 'image', data: 'eA==', mime_type: 'image/png' })) },
    { ...request, input: Array.from({ length: 4 }, () => ({ type: 'video', data: 'eA==', mime_type: 'video/mp4' })) },
  ]) assert.throws(() => plugin.native.createInteraction({ body: { kind: 'json', value: body } }));
});
