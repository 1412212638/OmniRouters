import type { PricingModel } from '../types'
import {
  getTaskImageOutputPriceEntries,
  getTaskVideoPriceEntry,
  type DynamicPriceEntry,
} from './dynamic-price'
import { taskPriceLabel, taskUsageUnitLabel } from './task-price-display'

export type CatalogTaskPriceRows = {
  inputLabel?: string
  inputValue: string
  inputUnit: string
  inputFreeQuantity?: number
  inputFreeQuantityMax?: number
  outputLabel?: string
  outputValue: string
  outputUnit: string
}

function formatCardPriceRange(value: string): string {
  const separator = value.indexOf('-')
  if (separator < 0) return value
  const minimum = value.slice(0, separator)
  const maximum = value.slice(separator + 1)
  return minimum === maximum ? minimum : `${minimum}–${maximum}`
}

export function getCatalogTaskPriceRows(
  model: PricingModel,
  entries: DynamicPriceEntry[],
  language: string,
  countFallback: string,
  secondUnit: string,
  referenceVideoLabel = 'Reference video input',
  generatedVideoLabel = 'Generated video output'
): CatalogTaskPriceRows | null {
  const outputImagePrices = getTaskImageOutputPriceEntries(model, entries)
  if (outputImagePrices.length > 0) {
    const schema = model.billing_usage_schema || {}
    const input = entries.find((entry) =>
      ['input_images', 'input_image_count'].includes(entry.field)
    )
    const inputDefinition = input ? schema[input.field] : undefined
    const outputDefinition = schema[outputImagePrices[0].field]
    const lowest = outputImagePrices[0]
    const highest = outputImagePrices.at(-1) || lowest
    const lowestFormatted = formatCardPriceRange(lowest.formatted)
    const highestFormatted = formatCardPriceRange(highest.formatted)
    let outputValue = lowestFormatted
    if (
      lowest !== highest &&
      (lowest.value !== highest.value || lowestFormatted !== highestFormatted)
    ) {
      outputValue = `${lowestFormatted}–${highestFormatted}`
    }
    const outputUnit = taskUsageUnitLabel(
      outputDefinition,
      language,
      countFallback
    )

    return {
      ...(input && {
        inputLabel: taskPriceLabel(
          inputDefinition?.description,
          input.field,
          language
        ),
      }),
      inputValue: input ? formatCardPriceRange(input.formatted) : '-',
      inputUnit: input
        ? taskUsageUnitLabel(inputDefinition, language, outputUnit)
        : '',
      ...(input?.freeQuantity != null && {
        inputFreeQuantity: input.freeQuantity,
        ...(input.freeQuantityMax != null && {
          inputFreeQuantityMax: input.freeQuantityMax,
        }),
      }),
      outputValue,
      outputUnit,
    }
  }

  const videoPrice = getTaskVideoPriceEntry(model, entries)
  if (!videoPrice) return null
  if (
    model.billing_usage_schema?.[videoPrice.field]?.pricingDisplay ===
    'input_output_video_seconds'
  ) {
    return {
      inputLabel: referenceVideoLabel,
      inputValue: formatCardPriceRange(videoPrice.formatted),
      inputUnit: secondUnit,
      outputLabel: generatedVideoLabel,
      outputValue: formatCardPriceRange(videoPrice.formatted),
      outputUnit: secondUnit,
    }
  }
  return {
    inputValue: '-',
    inputUnit: '',
    outputValue: formatCardPriceRange(videoPrice.formatted),
    outputUnit: secondUnit,
  }
}
