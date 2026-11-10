import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import { extractTaskMedia, getTaskPreviewMedia } from './task-media'

describe('task media preview extraction', () => {
  test('offers Omni preview without copying Base64 into the task list', () => {
    const log = {
      data: {
        steps: [
          {
            type: 'model_output',
            content: [{ type: 'video', mime_type: 'video/mp4' }],
          },
        ],
      },
      status: 'SUCCESS',
      action: 'omni_video',
      task_id: 'public-id',
      fail_reason: '',
    }
    assert.deepEqual(getTaskPreviewMedia(log, 'SUCCESS', true), [
      { kind: 'video', url: '/v1/videos/public-id/content' },
    ])
    assert.deepEqual(
      getTaskPreviewMedia({ ...log, status: 'FAILURE' }, 'SUCCESS', true),
      []
    )
  })
  test('extracts image outputs using output format metadata', () => {
    assert.deepEqual(
      extractTaskMedia({
        data: [{ url: 'https://cdn.example/image', output_format: 'png' }],
      }),
      [{ kind: 'image', url: 'https://cdn.example/image' }]
    )
  })

  test('recognizes video and audio output URL fields', () => {
    assert.deepEqual(
      extractTaskMedia(
        {
          output: { urls: ['https://cdn.example/result?token=x'] },
          clips: [{ audio_url: 'https://cdn.example/song' }],
        },
        'video'
      ),
      [
        { kind: 'video', url: 'https://cdn.example/result?token=x' },
        { kind: 'audio', url: 'https://cdn.example/song' },
      ]
    )
  })

  test('uses the gateway video content URL for legacy successful tasks', () => {
    assert.deepEqual(
      getTaskPreviewMedia(
        {
          data: {},
          status: 'SUCCESS',
          action: 'generate',
          task_id: 'task/unsafe id',
          fail_reason: 'https://upstream.example/video.mp4',
        },
        'SUCCESS',
        true
      ),
      [
        {
          kind: 'video',
          url: '/v1/videos/task%2Funsafe%20id/content',
        },
      ]
    )
  })

  test('does not expose input reference media as generated output', () => {
    assert.deepEqual(
      extractTaskMedia({
        input: { image_url: 'https://cdn.example/reference.png' },
        output: { image_url: 'https://cdn.example/generated.png' },
      }),
      [{ kind: 'image', url: 'https://cdn.example/generated.png' }]
    )
  })
})
