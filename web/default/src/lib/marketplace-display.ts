export type MarketplaceDisplayModel = {
  is_new?: number | boolean | string | null;
  discount_enabled?: number | boolean | string | null;
  discount_percent?: number | string | null;
  discount_label?: string | null;
  promotion_note?: string | null;
  display_original_price_source?: string | null;
  display_original_price_group?: string | null;
};

export function isDisplayEnabled(value: unknown) {
  return value === 1 || value === true || value === "1";
}

export function getDiscountLabel(
  model: MarketplaceDisplayModel,
  fallback = "Discount",
) {
  if (!isDisplayEnabled(model.discount_enabled)) return "";

  const customLabel = String(model.discount_label || "").trim();
  if (customLabel) return customLabel;

  const percent = Number(model.discount_percent);
  if (Number.isFinite(percent) && percent > 0) {
    return `${Number.isInteger(percent) ? percent : percent.toFixed(2)}% OFF`;
  }

  return fallback;
}

export function getMarketplaceDisplay(
  model: MarketplaceDisplayModel,
  fallback = "Discount",
) {
  return {
    showNew: isDisplayEnabled(model.is_new),
    discountLabel: getDiscountLabel(model, fallback),
    promotionNote: String(model.promotion_note || "").trim(),
  };
}
