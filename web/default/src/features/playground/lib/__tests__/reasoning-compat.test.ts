import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import { applyChatCompletionResponse } from '../message/message-streaming-utils'
import { createLoadingAssistantMessage } from '../message/message-utils'
import { parseStreamMessageUpdates } from '../streaming/stream-utils'

describe('Playground reasoning response compatibility', () => {
  test('reads reasoning from stream delta.reasoning', () => {
    const updates = parseStreamMessageUpdates(
      JSON.stringify({ choices: [{ delta: { reasoning: 'private thought' } }] })
    )
    assert.deepEqual(updates, [{ type: 'reasoning', chunk: 'private thought' }])
  })

  test('reads reasoning from non-stream message.reasoning', () => {
    const message = applyChatCompletionResponse(
      createLoadingAssistantMessage(0),
      {
        id: 'chatcmpl-test',
        object: 'chat.completion',
        created: 0,
        model: 'test-model',
        choices: [
          {
            index: 0,
            message: {
              role: 'assistant',
              content: 'answer',
              reasoning: 'private thought',
            },
            finish_reason: 'stop',
          },
        ],
      }
    )
    assert.equal(message?.reasoning?.content, 'private thought')
  })
})
