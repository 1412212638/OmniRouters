import type { PricingModel } from '../types'
import {
  getTaskImagePriceEntries,
  getTaskVideoPriceEntry,
  type DynamicPriceEntry,
} from './dynamic-price'
import { taskPriceLabel, taskUsageUnitLabel } from './task-price-display'

export type CatalogTaskPriceRows = {
  inputLabel?: string
  inputValue: string
  inputUnit: string
  outputLabel?: string
  outputValue: string
  outputUnit: string
}

function formatCardPriceRange(value: string): string {
  const separator = value.indexOf('-')
  return separator < 0
    ? value
    : `${value.slice(0, separator)}–${value.slice(separator + 1)}`
}

export function getCatalogTaskPriceRows(
  model: PricingModel,
  entries: DynamicPriceEntry[],
  language: string,
  countFallback: string,
  secondUnit: string
): CatalogTaskPriceRows | null {
  const imagePrices = getTaskImagePriceEntries(model, entries)
  if (imagePrices) {
    const schema = model.billing_usage_schema || {}
    const inputDefinition = schema[imagePrices.input.field]
    const outputDefinition = schema[imagePrices.output[0].field]
    const lowest = imagePrices.output[0]
    const highest = imagePrices.output.at(-1) || lowest
    const outputValue =
      lowest === highest
        ? formatCardPriceRange(lowest.formatted)
        : `${formatCardPriceRange(lowest.formatted)}–${formatCardPriceRange(highest.formatted)}`
    const unit = taskUsageUnitLabel(
      inputDefinition,
      language,
      taskUsageUnitLabel(outputDefinition, language, countFallback)
    )

    return {
      inputLabel: taskPriceLabel(
        inputDefinition?.description,
        imagePrices.input.field,
        language
      ),
      inputValue: formatCardPriceRange(imagePrices.input.formatted),
      inputUnit: unit,
      outputValue,
      outputUnit: unit,
    }
  }

  const videoPrice = getTaskVideoPriceEntry(model, entries)
  if (!videoPrice) return null
  return {
    inputValue: '-',
    inputUnit: '',
    outputValue: formatCardPriceRange(videoPrice.formatted),
    outputUnit: secondUnit,
  }
}
