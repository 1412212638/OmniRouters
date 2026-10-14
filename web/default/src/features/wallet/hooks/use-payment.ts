import i18next from 'i18next'
import { useState, useCallback } from 'react'
import { toast } from 'sonner'

import {
  calculateAmount,
  calculateStripeAmount,
  calculateWaffoPancakeAmount,
  requestPayment,
  requestStripePayment,
  isApiSuccess,
} from '../api'
import {
  isStripePayment,
  isWaffoPancakePayment,
  submitPaymentForm,
} from '../lib'
import type { ApiResponse, PaymentAmountBreakdown } from '../types'

// ============================================================================
// Payment Hook
// ============================================================================

export function usePayment() {
  const [amount, setAmount] = useState<number>(0)
  const [amountBreakdown, setAmountBreakdown] =
    useState<PaymentAmountBreakdown | null>(null)
  const [calculating, setCalculating] = useState(false)
  const [processing, setProcessing] = useState(false)

  // Calculate payment amount
  const calculatePaymentAmount = useCallback(
    async (topupAmount: number, paymentType: string) => {
      try {
        setCalculating(true)
        setAmountBreakdown(null)

        const isStripe = isStripePayment(paymentType)
        const isPancake = isWaffoPancakePayment(paymentType)
        let quoteResponse: ApiResponse<string> & {
          base_amount?: string
          fee_amount?: string
        }
        if (isStripe) {
          quoteResponse = await calculateStripeAmount({ amount: topupAmount })
        } else if (isPancake) {
          quoteResponse = await calculateWaffoPancakeAmount({
            amount: topupAmount,
          })
        } else {
          quoteResponse = await calculateAmount({ amount: topupAmount })
        }

        if (isApiSuccess(quoteResponse) && quoteResponse.data) {
          const calculatedAmount = Number.parseFloat(quoteResponse.data)
          setAmount(calculatedAmount)
          if (
            isPancake &&
            'base_amount' in quoteResponse &&
            quoteResponse.base_amount != null &&
            quoteResponse.fee_amount != null
          ) {
            const baseAmount = Number(quoteResponse.base_amount)
            const feeAmount = Number(quoteResponse.fee_amount)
            if (Number.isFinite(baseAmount) && Number.isFinite(feeAmount)) {
              setAmountBreakdown({ baseAmount, feeAmount })
            }
          }
          return calculatedAmount
        }

        // Don't show error for calculation, just set to 0
        setAmount(0)
        setAmountBreakdown(null)
        return 0
      } catch {
        setAmount(0)
        setAmountBreakdown(null)
        return 0
      } finally {
        setCalculating(false)
      }
    },
    []
  )

  // Process payment
  const processPayment = useCallback(
    async (topupAmount: number, paymentType: string) => {
      try {
        setProcessing(true)

        const isStripe = isStripePayment(paymentType)
        const amount = Math.floor(topupAmount)

        const response = isStripe
          ? await requestStripePayment({
              amount,
              payment_method: 'stripe',
            })
          : await requestPayment({
              amount,
              payment_method: paymentType,
            })

        if (!isApiSuccess(response)) {
          toast.error(response.message || i18next.t('Payment request failed'))
          return false
        }

        // Handle Stripe payment
        if (isStripe && response.data?.pay_link) {
          window.open(response.data.pay_link as string, '_blank')
          toast.success(i18next.t('Redirecting to payment page...'))
          return true
        }

        // Handle non-Stripe payment
        if (!isStripe && response.data) {
          const url = (response as unknown as { url?: string }).url
          if (url) {
            submitPaymentForm(url, response.data)
            toast.success(i18next.t('Redirecting to payment page...'))
            return true
          }
        }

        return false
      } catch {
        toast.error(i18next.t('Payment request failed'))
        return false
      } finally {
        setProcessing(false)
      }
    },
    []
  )

  return {
    amount,
    amountBreakdown,
    calculating,
    processing,
    calculatePaymentAmount,
    processPayment,
    setAmount,
  }
}
