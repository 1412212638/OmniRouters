import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import type { PricingModel } from '../types'
import { parseTiersFromExpr } from './billing-expr'
import { getCatalogTaskPriceRows } from './catalog-task-price'
import {
  getDynamicPriceEntries,
  getDynamicPricingSummary,
  getTaskImagePriceEntries,
  getTaskPriceEntries,
  getTaskVideoPriceEntry,
} from './dynamic-price'
import { getDisplayedModelName } from './model-helpers'

function testPricingModel(model: Partial<PricingModel>): PricingModel {
  return {
    id: 0,
    model_name: 'test-model',
    quota_type: 0,
    model_ratio: 1,
    completion_ratio: 1,
    enable_groups: [],
    ...model,
  }
}

describe('dynamic fixed request prices', () => {
  test('displays fixed prices as USD per request without token scaling', () => {
    const [tier] = parseTiersFromExpr('tier("request", fixed(0.005))')

    const [entry] = getDynamicPriceEntries(tier, {
      tokenUnit: 'M',
    })

    assert.equal(entry?.displayUnit, 'request')
    assert.equal(entry?.value, 0.005)
    assert.equal(entry?.formatted, '$0.005')

    const model = testPricingModel({
      billing_mode: 'tiered_expr',
      billing_expr: 'tier("request", fixed(0.005))',
    })
    const summary = getDynamicPricingSummary(model, { tokenUnit: 'M' })

    assert.equal(summary?.requestPriceRange, '$0.005')
  })
})

describe('task image price entries', () => {
  test('keeps input and output image prices and localized unit metadata', () => {
    const model = testPricingModel({
      billing_mode: 'tiered_expr',
      billing_expr:
        'tier("base", u("images_above_1_5k") * 0.6 + u("images_up_to_1_5k") * 0.3 + u("input_images") * 0.02)',
      billing_usage_schema: {
        images_above_1_5k: {
          type: 'number',
          unit: 'count',
          unitLabel: { en: 'image', zh: '张' },
          description: {
            en: 'Image generation unit price',
            zh: '图片生成单价',
          },
        },
        images_up_to_1_5k: {
          type: 'number',
          unit: 'count',
          unitLabel: { en: 'image', zh: '张' },
          description: {
            en: 'Image generation unit price',
            zh: '图片生成单价',
          },
        },
        input_images: {
          type: 'number',
          unit: 'count',
          unitLabel: { en: 'image', zh: '张' },
          description: { en: 'Input image unit price', zh: '输入图片单价' },
        },
      },
    })
    const entries = getTaskPriceEntries(model, { tokenUnit: 'M' })

    assert.deepEqual(
      entries.map(({ field, value, unitLabel }) => ({
        field,
        value,
        unitLabel,
      })),
      [
        {
          field: 'images_above_1_5k',
          value: 0.6,
          unitLabel: { en: 'image', zh: '张' },
        },
        {
          field: 'images_up_to_1_5k',
          value: 0.3,
          unitLabel: { en: 'image', zh: '张' },
        },
        {
          field: 'input_images',
          value: 0.02,
          unitLabel: { en: 'image', zh: '张' },
        },
      ]
    )
  })

  test('groups Qwen output tiers with its separately priced input images', () => {
    const model = testPricingModel({
      billing_mode: 'tiered_expr',
      billing_expr: 'tier("base", u("image_count") * 0.25)',
      billing_usage_schema: {
        image_count: {
          type: 'number',
          unit: 'count',
          unitLabel: { en: 'image', zh: '张' },
          description: {
            en: 'Image generation unit price',
            zh: '图片生成单价',
          },
        },
        output_image_type: { enum: ['qima_output_1k', 'qima_output_2k'] },
        input_image_count: {
          type: 'number',
          unit: 'count',
          unitLabel: { en: 'image', zh: '张' },
          description: { en: 'Input image unit price', zh: '输入图片单价' },
        },
      },
    })
    const entries = getTaskPriceEntries(model, { tokenUnit: 'M' }, [
      {
        label: '1K output',
        conditions: [{ field: 'output_image_type', value: 'qima_output_1k' }],
        constant: 0,
        unitPrices: { image_count: 0.25, input_image_count: 0.02 },
      },
      {
        label: '2K output',
        conditions: [{ field: 'output_image_type', value: 'qima_output_2k' }],
        constant: 0,
        unitPrices: { image_count: 0.5, input_image_count: 0.02 },
      },
    ])
    const imagePrices = getTaskImagePriceEntries(model, entries)

    assert.ok(imagePrices)
    assert.equal(imagePrices.input.value, 0.02)
    assert.equal(imagePrices.input.field, 'input_image_count')
    assert.equal(imagePrices.output.length, 1)
    assert.equal(imagePrices.output[0].value, 0.25)
    assert.equal(imagePrices.output[0].formatted, '$0.25-$0.5')
    assert.deepEqual(imagePrices.output[0].unitLabel, {
      en: 'image',
      zh: '张',
    })
  })

  test('treats resolution-tiered second prices as video output pricing', () => {
    const model = testPricingModel({
      billing_mode: 'tiered_expr',
      billing_expr: 'tier("base", u("seconds") * 0.45)',
      billing_usage_schema: {
        seconds: {
          type: 'number',
          unit: 'second',
          description: {
            en: 'Video generation unit price',
            zh: '视频生成单价',
          },
        },
        resolution: { enum: ['480P', '720P', '1080P'] },
      },
    })
    const entries = getTaskPriceEntries(model, { tokenUnit: 'M' }, [
      {
        label: '480P',
        conditions: [{ field: 'resolution', value: '480P' }],
        constant: 0,
        unitPrices: { seconds: 0.45 },
      },
      {
        label: '720P',
        conditions: [{ field: 'resolution', value: '720P' }],
        constant: 0,
        unitPrices: { seconds: 0.9 },
      },
      {
        label: '1080P',
        conditions: [{ field: 'resolution', value: '1080P' }],
        constant: 0,
        unitPrices: { seconds: 1.8 },
      },
    ])
    const videoPrice = getTaskVideoPriceEntry(model, entries)

    assert.ok(videoPrice)
    assert.equal(videoPrice.value, 0.45)
    assert.equal(videoPrice.formatted, '$0.45-$1.8')
  })

  test('recognizes video duration fields by unit and output dimensions metadata', () => {
    const model = testPricingModel({
      billing_usage_schema: {
        input_video_duration: {
          type: 'number',
          unit: 'second',
          description: {
            en: 'Input video duration (seconds)',
            zh: '输入视频时长（秒）',
          },
        },
        target_output_resolution: {
          enum: ['720p', '1080p', '2k'],
          description: { en: 'Target output resolution', zh: '目标输出分辨率' },
        },
        target_output_fps: {
          enum: ['30', '60', '120'],
          description: {
            en: 'Target output frames per second',
            zh: '目标输出帧率',
          },
        },
      },
    })
    const price = {
      key: 'task:input_video_duration',
      field: 'input_video_duration',
      label: 'input_video_duration',
      shortLabel: 'input_video_duration',
      value: 0.0125,
      formatted: '$0.0125-$1.6',
      displayUnit: 'custom' as const,
      unit: 'second',
    }

    assert.equal(getTaskVideoPriceEntry(model, [price]), price)
  })

  test('builds the catalog Seedream image price rows from task usage data', () => {
    const model = testPricingModel({
      billing_usage_schema: {
        images_above_1_5k: {
          type: 'number',
          unit: 'count',
          unitLabel: { en: 'image', zh: '张' },
          description: { en: 'Output image price', zh: '图片生成单价' },
        },
        images_up_to_1_5k: {
          type: 'number',
          unit: 'count',
          unitLabel: { en: 'image', zh: '张' },
          description: { en: 'Output image price', zh: '图片生成单价' },
        },
        input_images: {
          type: 'number',
          unit: 'count',
          unitLabel: { en: 'image', zh: '张' },
          description: { en: 'Input image unit price', zh: '输入图片单价' },
        },
      },
    })
    const entries = [
      { field: 'images_above_1_5k', value: 0.6, formatted: '¥0.6' },
      { field: 'images_up_to_1_5k', value: 0.3, formatted: '¥0.3' },
      { field: 'input_images', value: 0.02, formatted: '¥0.02' },
    ] as Parameters<typeof getCatalogTaskPriceRows>[1]

    assert.deepEqual(
      getCatalogTaskPriceRows(model, entries, 'zh', '数量', '秒'),
      {
        inputLabel: '输入图片单价',
        inputValue: '¥0.02',
        inputUnit: '张',
        outputValue: '¥0.3–¥0.6',
        outputUnit: '张',
      }
    )
  })

  test('builds catalog video rows with no input price and localized seconds', () => {
    const model = testPricingModel({
      billing_usage_schema: {
        input_video_duration: {
          type: 'number',
          unit: 'second',
          description: { en: 'Input video duration', zh: '输入视频时长（秒）' },
        },
        target_output_resolution: {
          enum: ['720p', '1080p'],
          description: { en: 'Target output resolution', zh: '目标输出分辨率' },
        },
      },
    })
    const entries = [
      {
        field: 'input_video_duration',
        value: 0.0125,
        formatted: '¥0.0125-¥1.6',
      },
    ] as Parameters<typeof getCatalogTaskPriceRows>[1]

    assert.deepEqual(
      getCatalogTaskPriceRows(model, entries, 'zh', '数量', '秒'),
      {
        inputValue: '-',
        inputUnit: '',
        outputValue: '¥0.0125–¥1.6',
        outputUnit: '秒',
      }
    )
  })
})

describe('displayed model names', () => {
  test('uses the developer instead of the channel supplier as the prefix', () => {
    assert.equal(
      getDisplayedModelName({
        model_name: 'minimax-m3.1-flash',
        developer_name: 'MiniMax',
        vendor_name: 'TencentCloud',
      } as Parameters<typeof getDisplayedModelName>[0]),
      'minimax/minimax-m3.1-flash'
    )
  })

  test('keeps supplier prefix behavior when no developer is configured', () => {
    assert.equal(
      getDisplayedModelName({
        model_name: 'gpt-4.1',
        vendor_name: 'OpenAI',
      } as Parameters<typeof getDisplayedModelName>[0]),
      'openai/gpt-4.1'
    )
  })
})
