import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import {
  applyChatCompletionResponse,
  applyEmptyRefusalNotice,
  applyStreamingChunk,
} from '../message/message-streaming-utils'
import { createLoadingAssistantMessage } from '../message/message-utils'
import { parseStreamMessageUpdates } from '../streaming/stream-utils'

describe('Playground reasoning response compatibility', () => {
  test('reads reasoning from stream delta.reasoning', () => {
    const updates = parseStreamMessageUpdates(
      JSON.stringify({ choices: [{ delta: { reasoning: 'private thought' } }] })
    )
    assert.deepEqual(updates, [{ type: 'reasoning', chunk: 'private thought' }])
  })

  test('shows a notice when a stream ends with a refusal and no content', () => {
    const updates = parseStreamMessageUpdates(
      JSON.stringify({
        choices: [{ delta: {}, finish_reason: 'content_filter' }],
      })
    )
    assert.equal(updates.length, 1)
    assert.equal(updates[0].type, 'refusal')
    const message = applyEmptyRefusalNotice(createLoadingAssistantMessage(0))
    assert.match(message.versions[0].content, /refused to answer/i)
    const withContent = applyStreamingChunk(
      createLoadingAssistantMessage(0),
      'content',
      'existing reply'
    )
    assert.equal(
      applyEmptyRefusalNotice(withContent).versions[0].content,
      'existing reply'
    )
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

  test('shows a notice for an empty non-stream refusal response', () => {
    const message = applyChatCompletionResponse(
      createLoadingAssistantMessage(0),
      {
        id: 'chatcmpl-refusal',
        object: 'chat.completion',
        created: 0,
        model: 'claude-test',
        choices: [
          {
            index: 0,
            message: { role: 'assistant', content: '' },
            finish_reason: 'content_filter',
          },
        ],
      }
    )
    assert.match(message?.versions[0].content ?? '', /refused to answer/i)
  })
})
