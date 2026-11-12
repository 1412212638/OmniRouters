import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { CreditCard } from 'lucide-react'
import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import { useTranslation } from 'react-i18next'

import { SectionPageLayout } from '@/components/layout'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { getOrganization } from '@/features/organizations/api'
import type { OrganizationWorkspace } from '@/features/organizations/types'
import { useActiveOrganizationId } from '@/features/organizations/workspace-store'
import { useStatus } from '@/hooks/use-status'
import { useSystemConfig } from '@/hooks/use-system-config'
import { getSelf } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'

import { getInvoiceProfile } from './api'
import { AffiliateRewardsCard } from './components/affiliate-rewards-card'
import { CorporateTransferCard } from './components/corporate-transfer-card'
import { BillingHistoryDialog } from './components/dialogs/billing-history-dialog'
import { CreemConfirmDialog } from './components/dialogs/creem-confirm-dialog'
import { PaymentConfirmDialog } from './components/dialogs/payment-confirm-dialog'
import { TransferDialog } from './components/dialogs/transfer-dialog'
import { WalletFundingSummary } from './components/funding-summary'
import { InvoiceProfileCard } from './components/invoice-profile-card'
import { RechargeFormCard } from './components/recharge-form-card'
import { WalletStatsCard } from './components/wallet-stats-card'
import { DEFAULT_DISCOUNT_RATE } from './constants'
import { getFundingSummary } from './funding'
import {
  useTopupInfo,
  usePayment,
  useAffiliate,
  useRedemption,
  useCreemPayment,
  useWaffoPayment,
  useWaffoPancakePayment,
} from './hooks'
import { getMinTopupAmount, isWaffoPancakePayment } from './lib'
import type {
  UserWalletData,
  PaymentMethod,
  PresetAmount,
  CreemProduct,
} from './types'

interface WalletProps {
  initialShowHistory?: boolean
}

export function Wallet(props: WalletProps) {
  const userId = useAuthStore((state) => state.auth.user?.id)
  const organizationId = useActiveOrganizationId()
  const { t } = useTranslation()
  const workspace = useQuery({
    queryKey: ['organization', userId, organizationId],
    queryFn: () => getOrganization(organizationId as number),
    enabled: organizationId !== null,
  })
  if (organizationId && workspace.isPending) {
    return <Skeleton className='h-48 w-full' />
  }
  if (
    organizationId &&
    (!workspace.data ||
      !['owner', 'admin'].includes(workspace.data.membership.role))
  ) {
    return (
      <SectionPageLayout>
        <SectionPageLayout.Title>
          {t('Organization wallet')}
        </SectionPageLayout.Title>
        <SectionPageLayout.Content>
          <Alert>
            <AlertDescription>
              {t(
                'Only organization owners and admins can manage the shared wallet.'
              )}
            </AlertDescription>
          </Alert>
        </SectionPageLayout.Content>
      </SectionPageLayout>
    )
  }
  return (
    <WalletContent
      key={organizationId ?? 'personal'}
      {...props}
      workspace={organizationId ? workspace.data : undefined}
    />
  )
}

function WalletContent(
  props: WalletProps & { workspace?: OrganizationWorkspace }
) {
  const isOrganization = Boolean(props.workspace)
  const queryClient = useQueryClient()
  const viewerId = useAuthStore((state) => state.auth.user?.id)
  const funding = useQuery({
    queryKey: ['wallet-funding-summary', viewerId],
    queryFn: getFundingSummary,
    enabled: !isOrganization,
  })

  const { t } = useTranslation()
  const [user, setUser] = useState<UserWalletData | null>(null)
  const [userLoading, setUserLoading] = useState(true)
  const [topupAmount, setTopupAmount] = useState(0)
  const [selectedPreset, setSelectedPreset] = useState<number | null>(null)
  const [selectedPaymentMethod, setSelectedPaymentMethod] =
    useState<PaymentMethod>()
  const [paymentLoading, setPaymentLoading] = useState<string | null>(null)
  const [confirmDialogOpen, setConfirmDialogOpen] = useState(false)
  const [transferDialogOpen, setTransferDialogOpen] = useState(false)
  const [billingDialogOpen, setBillingDialogOpen] = useState(false)
  const [redemptionCode, setRedemptionCode] = useState('')
  const [creemDialogOpen, setCreemDialogOpen] = useState(false)
  const [selectedCreemProduct, setSelectedCreemProduct] =
    useState<CreemProduct | null>(null)

  const { status } = useStatus()
  const { currency, invoiceProfileEnabled = true } = useSystemConfig()
  const { topupInfo, presetAmounts, loading: topupLoading } = useTopupInfo()

  // Calculate effective exchange rate - when display type is USD, use rate of 1
  const effectiveUsdExchangeRate = useMemo(() => {
    return currency?.quotaDisplayType === 'USD'
      ? 1
      : currency?.usdExchangeRate || 1
  }, [currency?.quotaDisplayType, currency?.usdExchangeRate])
  const {
    amount: paymentAmount,
    amountBreakdown: paymentAmountBreakdown,
    calculating,
    processing,
    calculatePaymentAmount,
    processPayment,
  } = usePayment()
  const {
    affiliateLink,
    loading: affiliateLoading,
    transferQuota,
    transferring,
  } = useAffiliate(!isOrganization)
  const { redeeming, redeemCode } = useRedemption()
  const { processing: creemProcessing, processCreemPayment } = useCreemPayment()
  const { processWaffoPayment } = useWaffoPayment()
  const { processing: pancakeProcessing, processWaffoPancakePayment } =
    useWaffoPancakePayment()

  // Fetch and refresh user data
  const fetchUser = useCallback(async () => {
    try {
      setUserLoading(true)
      if (props.workspace) {
        const result = await getOrganization(props.workspace.organization.id)
        const org = result.organization
        setUser({
          id: org.id,
          username: org.name,
          quota: org.quota ?? 0,
          used_quota: org.used_quota ?? 0,
          request_count: org.request_count ?? 0,
          group: org.group,
          aff_quota: 0,
          aff_history_quota: 0,
          aff_count: 0,
        })
        return
      }
      void queryClient.invalidateQueries({
        queryKey: ['wallet-funding-summary', viewerId],
      })
      const response = await getSelf()
      const invoiceResponse = invoiceProfileEnabled
        ? await getInvoiceProfile().catch(() => null)
        : null
      if (response.success && response.data) {
        setUser({
          ...(response.data as UserWalletData),
          invoice_profile: invoiceResponse?.success
            ? invoiceResponse.data
            : null,
        })
      }
    } catch (error) {
      // eslint-disable-next-line no-console
      console.error('Failed to fetch user data:', error)
    } finally {
      setUserLoading(false)
    }
  }, [invoiceProfileEnabled, props.workspace, queryClient, viewerId])

  useEffect(() => {
    fetchUser()
  }, [fetchUser])

  useEffect(() => {
    if (props.initialShowHistory) {
      setBillingDialogOpen(true)
      window.history.replaceState({}, '', window.location.pathname)
    }
  }, [props.initialShowHistory])

  // Initialize topup amount when topup info is loaded
  const topupAmountInitializedRef = useRef(false)
  useEffect(() => {
    if (topupInfo && !topupAmountInitializedRef.current) {
      topupAmountInitializedRef.current = true
      const minTopup = getMinTopupAmount(topupInfo)
      setTopupAmount(minTopup)

      // Show the base quote until the user chooses a payment method.
      calculatePaymentAmount(minTopup, '')
    }
  }, [topupInfo, calculatePaymentAmount])

  // Payment-specific quotes and fees apply only after an explicit selection.
  const getCurrentPaymentType = useCallback(() => {
    return selectedPaymentMethod?.type || ''
  }, [selectedPaymentMethod])

  // Handle preset selection
  const handleSelectPreset = (preset: PresetAmount) => {
    setTopupAmount(preset.value)
    setSelectedPreset(preset.value)
    calculatePaymentAmount(preset.value, getCurrentPaymentType())
  }

  // Handle topup amount change
  const handleTopupAmountChange = (amount: number) => {
    setTopupAmount(amount)
    setSelectedPreset(null)
    calculatePaymentAmount(amount, getCurrentPaymentType())
  }

  // Handle payment method selection
  const handlePaymentMethodSelect = async (method: PaymentMethod) => {
    setSelectedPaymentMethod(method)
    setPaymentLoading(method.type)

    try {
      // Validate minimum topup
      const minTopup = getMinTopupAmount(topupInfo)
      if (topupAmount < minTopup) {
        return
      }

      // Calculate payment amount and show confirmation dialog
      await calculatePaymentAmount(topupAmount, method.type)
      setConfirmDialogOpen(true)
    } finally {
      setPaymentLoading(null)
    }
  }

  // Handle payment confirmation
  const handlePaymentConfirm = async () => {
    if (!selectedPaymentMethod) return

    const isPancake = isWaffoPancakePayment(selectedPaymentMethod.type)
    const success = isPancake
      ? await processWaffoPancakePayment(topupAmount)
      : await processPayment(topupAmount, selectedPaymentMethod.type)

    if (success) {
      setConfirmDialogOpen(false)
      await fetchUser()
    }
  }

  // Handle redemption
  const handleRedeem = async () => {
    if (!redemptionCode) return

    const success = await redeemCode(redemptionCode)
    if (success) {
      setRedemptionCode('')
      await fetchUser()
    }
  }

  // Handle transfer
  const handleTransfer = async (amount: number) => {
    const success = await transferQuota(amount)
    if (success) {
      await fetchUser()
    }
    return success
  }

  // Handle Creem product selection
  const handleCreemProductSelect = (product: CreemProduct) => {
    setSelectedCreemProduct(product)
    setCreemDialogOpen(true)
  }

  // Handle Creem payment confirmation
  const handleCreemConfirm = async () => {
    if (!selectedCreemProduct) return

    const success = await processCreemPayment(selectedCreemProduct.productId)
    if (success) {
      setCreemDialogOpen(false)
      setSelectedCreemProduct(null)
      await fetchUser()
    }
  }

  const handleWaffoMethodSelect = async (_method: unknown, index: number) => {
    const loadingKey = `waffo-${index}`
    setPaymentLoading(loadingKey)

    try {
      await processWaffoPayment(topupAmount, index)
    } finally {
      setPaymentLoading(null)
    }
  }

  // Get discount rate for current topup amount
  const getDiscountRate = useCallback(() => {
    return topupInfo?.discount?.[topupAmount] || DEFAULT_DISCOUNT_RATE
  }, [topupInfo, topupAmount])

  return (
    <>
      <SectionPageLayout>
        <SectionPageLayout.Title>
          {isOrganization
            ? `${props.workspace?.organization.name} · ${t('Organization wallet')}`
            : t('Wallet')}
        </SectionPageLayout.Title>
        <SectionPageLayout.Content>
          <div className='mx-auto flex w-full max-w-7xl flex-col gap-4 sm:gap-5'>
            {isOrganization && (
              <Alert>
                <AlertDescription>
                  {t(
                    'Payments made here are credited directly to this organization. Personal balances remain separate.'
                  )}
                </AlertDescription>
              </Alert>
            )}
            <WalletStatsCard user={user} loading={userLoading} />
            {!isOrganization && (
              <WalletFundingSummary
                summary={funding.data}
                loading={funding.isPending}
                failed={funding.isError}
              />
            )}
            {!isOrganization && invoiceProfileEnabled && (
              <InvoiceProfileCard
                profile={user?.invoice_profile}
                onSaved={(invoiceProfile) =>
                  setUser((current) =>
                    current
                      ? { ...current, invoice_profile: invoiceProfile }
                      : current
                  )
                }
              />
            )}

            <div className='grid gap-4'>
              <div id='wallet-add-funds' className='scroll-mt-4'>
                <RechargeFormCard
                  allowRedemption={!isOrganization}
                  topupInfo={topupInfo}
                  presetAmounts={presetAmounts}
                  selectedPreset={selectedPreset}
                  paymentType={getCurrentPaymentType()}
                  selectedFeeRate={
                    topupInfo?.pay_methods?.find(
                      (method) => method.type === getCurrentPaymentType()
                    )?.fee_rate
                  }
                  onSelectPreset={handleSelectPreset}
                  topupAmount={topupAmount}
                  onTopupAmountChange={handleTopupAmountChange}
                  paymentAmount={paymentAmount}
                  calculating={calculating}
                  onPaymentMethodSelect={handlePaymentMethodSelect}
                  paymentLoading={paymentLoading}
                  redemptionCode={redemptionCode}
                  onRedemptionCodeChange={setRedemptionCode}
                  onRedeem={handleRedeem}
                  redeeming={redeeming}
                  topupLink={topupInfo?.topup_link}
                  loading={topupLoading}
                  priceRatio={(status?.price as number) || 1}
                  usdExchangeRate={effectiveUsdExchangeRate}
                  onOpenBilling={() => setBillingDialogOpen(true)}
                  creemProducts={topupInfo?.creem_products}
                  enableCreemTopup={topupInfo?.enable_creem_topup}
                  onCreemProductSelect={handleCreemProductSelect}
                  enableWaffoTopup={topupInfo?.enable_waffo_topup}
                  waffoPayMethods={topupInfo?.waffo_pay_methods}
                  waffoMinTopup={topupInfo?.waffo_min_topup}
                  onWaffoMethodSelect={handleWaffoMethodSelect}
                  enableWaffoPancakeTopup={
                    topupInfo?.enable_waffo_pancake_topup
                  }
                />
                {!isOrganization && topupInfo?.corporate_transfer?.enabled && (
                  <div className='mt-4'>
                    <CorporateTransferCard
                      config={topupInfo.corporate_transfer}
                    />
                  </div>
                )}
              </div>
            </div>

            {!isOrganization && (
              <Card data-card-hover='false'>
                <CardContent className='flex flex-wrap items-center justify-between gap-3 p-4'>
                  <div className='flex min-w-0 items-center gap-3'>
                    <span className='bg-primary/10 text-primary flex size-9 shrink-0 items-center justify-center rounded-lg'>
                      <CreditCard className='size-4' aria-hidden='true' />
                    </span>
                    <div className='min-w-0'>
                      <p className='font-medium'>{t('Subscription Plans')}</p>
                      <p className='text-muted-foreground text-xs'>
                        {t(
                          'Choose a billing unit and subscribe to a plan for model access.'
                        )}
                      </p>
                    </div>
                  </div>
                  <Button
                    variant='outline'
                    render={<Link to='/subscription-plans' />}
                  >
                    {t('View subscription plans')}
                  </Button>
                </CardContent>
              </Card>
            )}

            {!isOrganization && (
              <AffiliateRewardsCard
                user={user}
                affiliateLink={affiliateLink}
                onTransfer={() => setTransferDialogOpen(true)}
                loading={affiliateLoading}
              />
            )}
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <PaymentConfirmDialog
        open={confirmDialogOpen}
        onOpenChange={setConfirmDialogOpen}
        onConfirm={handlePaymentConfirm}
        topupAmount={topupAmount}
        paymentAmount={paymentAmount}
        paymentBreakdown={paymentAmountBreakdown}
        paymentMethod={selectedPaymentMethod}
        calculating={calculating}
        processing={processing || pancakeProcessing}
        discountRate={getDiscountRate()}
        feeRate={
          selectedPaymentMethod
            ? (selectedPaymentMethod.fee_rate ?? topupInfo?.fee_rate)
            : 0
        }
        usdExchangeRate={effectiveUsdExchangeRate}
      />

      <TransferDialog
        open={transferDialogOpen}
        onOpenChange={setTransferDialogOpen}
        onConfirm={handleTransfer}
        availableQuota={user?.aff_quota ?? 0}
        transferring={transferring}
      />

      <BillingHistoryDialog
        open={billingDialogOpen}
        onOpenChange={setBillingDialogOpen}
      />

      <CreemConfirmDialog
        open={creemDialogOpen}
        onOpenChange={setCreemDialogOpen}
        onConfirm={handleCreemConfirm}
        product={selectedCreemProduct}
        processing={creemProcessing}
      />
    </>
  )
}
