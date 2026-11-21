export function getRechargeDiscountRate(topupInfo, amount) {
  const discounts = topupInfo?.discount ?? {};
  if (!topupInfo?.discount_by_threshold) {
    return discounts[amount] > 0 ? discounts[amount] : 1;
  }

  let highest = 0;
  let rate = 1;
  for (const [threshold, candidate] of Object.entries(discounts)) {
    const minimum = Number(threshold);
    if (
      minimum > highest &&
      minimum <= amount &&
      candidate > 0 &&
      candidate <= 1
    ) {
      highest = minimum;
      rate = candidate;
    }
  }
  return rate;
}
