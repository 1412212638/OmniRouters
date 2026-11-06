import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import type { TaskPluginOption } from '../../channels/api'
import {
  createDefaultTaskMatrixConfig,
  evaluateTaskVisualConfig,
  generateTaskExprFromConfig,
  getTaskNumberFields,
  taskMatrixToTiers,
  tryParseTaskMatrixConfig,
} from './task-expr'
import { getTaskPluginUsageSchema } from './task-plugin-usage'
import { taskUsageUnitLabel } from './task-price-display'

const plugin: Pick<TaskPluginOption, 'usageSchema' | 'usageProfiles'> = {
  usageSchema: {
    image_count: { type: 'number', unit: 'count' },
    seconds: { type: 'number', unit: 'second' },
    resolution: { enum: ['480P', '720P', '1080P'] },
  },
  usageProfiles: [
    {
      models: ['wan3.0-video', 'wan3.0-video-prime'],
      schema: {
        seconds: { type: 'number', unit: 'second' },
        resolution: { enum: ['480P', '720P', '1080P'] },
      },
    },
    {
      models: ['wan2.7-image'],
      schema: { image_count: { type: 'number', unit: 'count' } },
    },
    { models: ['no-usage'], schema: {} },
  ],
}

describe('task plugin model pricing metadata', () => {
  test('localized count units are used for model pricing display', () => {
    const definition = {
      unit: 'count',
      unitLabel: { en: 'image', zh: '张', 'zh-TW': '張' },
    }

    assert.equal(taskUsageUnitLabel(definition, 'zh', 'Count'), '张')
    assert.equal(taskUsageUnitLabel(definition, 'en', 'Count'), 'image')
    assert.equal(taskUsageUnitLabel({ unit: 'count' }, 'zh', 'Count'), 'Count')
  })

  test('Wan3 pricing generates only model-declared usage fields', () => {
    const schema = getTaskPluginUsageSchema(plugin, 'WAN3.0-VIDEO')
    assert.ok(schema)
    assert.deepEqual(
      getTaskNumberFields(schema).map(([name]) => name),
      ['seconds']
    )
    const matrix = createDefaultTaskMatrixConfig(schema)
    matrix.rows.forEach((row, index) => {
      row.unitPrices.seconds = [0.3, 0.6, 1.2][index]
    })
    const expression = generateTaskExprFromConfig(
      { tiers: taskMatrixToTiers(matrix, schema) },
      schema
    )
    assert.doesNotMatch(expression, /image_count/)
    assert.match(expression, /u\("seconds"\) \* 0\.3/)
    assert.match(expression, /u\("seconds"\) \* 0\.6/)
    assert.match(expression, /u\("seconds"\) \* 1\.2/)
    assert.match(expression, /u\("resolution"\) == "720P"/)
  })

  test('image model pricing does not inherit video fields', () => {
    const schema = getTaskPluginUsageSchema(plugin, 'wan2.7-image')
    assert.ok(schema)
    assert.deepEqual(Object.keys(schema), ['image_count'])
  })

  test('reediting the old zero image term preserves the configured video price', () => {
    const schema = getTaskPluginUsageSchema(plugin, 'wan3.0-video')
    assert.ok(schema)
    const matrix = tryParseTaskMatrixConfig(
      'tier("base", u("image_count") * 0 + u("seconds") * 0.6)',
      schema
    )
    assert.ok(matrix)
    assert.equal(matrix.rows.length, 3)
    assert.ok(matrix.rows.every((row) => row.unitPrices.seconds === 0.6))
    const expression = generateTaskExprFromConfig(
      { tiers: taskMatrixToTiers(matrix, schema) },
      schema
    )
    assert.equal(expression, 'tier("base", u("seconds") * 0.6)')
  })

  test('count pricing supports a free quantity without changing the billable expression', () => {
    const schema = {
      input_images: {
        type: 'number' as const,
        unit: 'count' as const,
        unitLabel: { en: 'image', zh: '张' },
      },
    }
    const config = createDefaultTaskMatrixConfig(schema)
    config.rows[0].unitPrices.input_images = 0.05
    config.rows[0].freeQuantities.input_images = 5
    const expression = generateTaskExprFromConfig(
      { tiers: taskMatrixToTiers(config, schema) },
      schema
    )
    assert.equal(
      expression,
      'tier("base", max(u("input_images") - 5, 0) * 0.05)'
    )
    const parsed = tryParseTaskMatrixConfig(expression, schema)
    assert.ok(parsed)
    assert.equal(parsed.rows[0].freeQuantities.input_images, 5)
    const preview = evaluateTaskVisualConfig(
      { tiers: taskMatrixToTiers(parsed, schema) },
      { input_images: 7 },
      schema
    )
    assert.equal(preview?.total, 0.1)
  })

  test('new count pricing uses the plugin free-quantity suggestion', () => {
    const schema = {
      input_images: {
        type: 'number' as const,
        unit: 'count' as const,
        freeQuantity: 1,
      },
    }
    const config = createDefaultTaskMatrixConfig(schema)
    assert.equal(config.rows[0].freeQuantities.input_images, 1)
  })

  test('legacy plugins and unmatched models retain their default schema', () => {
    assert.equal(getTaskPluginUsageSchema(plugin, 'other'), plugin.usageSchema)
    assert.equal(
      getTaskPluginUsageSchema(
        { usageSchema: plugin.usageSchema },
        'wan3.0-video'
      ),
      plugin.usageSchema
    )
  })

  test('empty model profiles replace rather than merge the default schema', () => {
    assert.deepEqual(getTaskPluginUsageSchema(plugin, 'no-usage'), {})
  })
})
