'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { AlertCircle, CheckCircle2, Download, RefreshCw, Wallet } from 'lucide-react'
import { AppShell } from '@/components/app-shell'
import { LoadingState } from '@/components/empty-state'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { StripeOnrampWidget } from '@/components/stripe-onramp-widget'
import {
  createStripeTopUpSession,
  fetchBilling,
  fetchBuyerPaymentControls,
  fetchBuyerWalletTopUps,
  fetchRuntimeConfig,
  fetchX402Intents,
  fetchInvoices,
  type Billing,
  type BuyerPaymentControls,
  type PaymentMethodCatalogItem,
  type WalletTopUp,
  type X402Intent,
  settleX402Intent,
  updateBuyerPaymentControls,
} from '@/lib/api-client'
import {
  getPaymentMethodDescription,
  getPaymentMethodStatus,
  getPaymentMethodTitle,
} from '@/lib/payment-methods'
import { toast } from 'sonner'

const planFeatures: Record<
  string,
  { description: string; price: number; features: string[] }
> = {
  free: {
    description: 'For getting started',
    price: 0,
    features: ['5 servers', 'Basic analytics', 'Community support'],
  },
  starter: {
    description: 'For small teams',
    price: 29,
    features: ['25 servers', 'Advanced analytics', 'Email support', 'API access'],
  },
  professional: {
    description: 'For growing businesses',
    price: 99,
    features: [
      'Unlimited servers',
      'Real-time analytics',
      'Priority support',
      'Custom integrations',
      'SSO',
    ],
  },
  enterprise: {
    description: 'Custom for large organizations',
    price: 0,
    features: [
      'Everything in Professional',
      'Dedicated support',
      'Custom SLA',
      'Advanced security',
      'Audit logs',
    ],
  },
}

interface InvoiceRecord {
  id: string
  date: Date
  amount: number
  status: string
}

interface WalletTopUpSummary {
  items: WalletTopUp[]
  count: number
  minimumTopUpUsd: number
  defaultTopUpUsd: number
  stripeConfigured: boolean
}

interface StripeTopUpSessionResponse {
  topup?: WalletTopUp
  stripe?: {
    clientSecret?: string
    hostedUrl?: string
  }
}

interface X402IntentSummary {
  items: X402Intent[]
  count: number
}

interface IntentFeedback {
  tone: 'success' | 'error'
  message: string
}

const emptyBilling: Billing = {
  id: '',
  userId: '',
  plan: 'free',
  monthlySpend: 0,
  dailySpend: 0,
  currentBalance: 0,
  nextBillingDate: new Date(),
  allowedMethods: [],
  caps: {},
  wallet: {},
  status: 'active',
}

const emptyTopUpSummary: WalletTopUpSummary = {
  items: [],
  count: 0,
  minimumTopUpUsd: 1,
  defaultTopUpUsd: 50,
  stripeConfigured: false,
}

const emptyX402IntentSummary: X402IntentSummary = {
  items: [],
  count: 0,
}

function formatDateTime(value?: string | Date | null, fallback = 'Not available') {
  if (!value) return fallback
  const date = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(date.getTime()) || date.getUTCFullYear() <= 1) return fallback
  return date.toLocaleString()
}

function humanizeMethodId(value?: string | null) {
  if (!value) return 'Not provided'
  return value
    .split('_')
    .filter(Boolean)
    .map(part => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}

function getIntentStatusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (status) {
    case 'settled':
      return 'default'
    case 'pending':
      return 'secondary'
    case 'failed':
    case 'rejected':
      return 'destructive'
    default:
      return 'outline'
  }
}

function parseIntentChallenge(intent: X402Intent) {
  if (!intent.challenge) return null
  try {
    const parsed = JSON.parse(intent.challenge)
    const first = Array.isArray(parsed) ? parsed[0] : parsed
    return first && typeof first === 'object' ? (first as Record<string, unknown>) : null
  } catch {
    return null
  }
}

function PaymentMethodStatusCard({
  method,
  allowedMethods,
}: {
  method: PaymentMethodCatalogItem
  allowedMethods: string[]
}) {
  const readiness = getPaymentMethodStatus(method, allowedMethods)

  return (
    <div className="rounded-lg border border-border p-4 space-y-3">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="font-semibold text-sm">{getPaymentMethodTitle(method)}</p>
          <p className="text-xs text-muted-foreground mt-1">{method.integration}</p>
        </div>
        <Badge variant={readiness.tone}>{readiness.label}</Badge>
      </div>
      <p className="text-sm text-muted-foreground">{getPaymentMethodDescription(method)}</p>
      <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
        {method.network && <span>Network: {method.network}</span>}
        {method.asset && <span>Asset: {method.asset}</span>}
        <span>{readiness.description}</span>
      </div>
      {method.docs && (
        <a
          href={method.docs}
          target="_blank"
          rel="noreferrer"
          className="text-xs text-primary underline-offset-4 hover:underline"
        >
          View setup docs
        </a>
      )}
    </div>
  )
}

export default function BillingPage() {
  const [isLoading, setIsLoading] = useState(true)
  const [billing, setBilling] = useState<Billing>(emptyBilling)
  const [invoices, setInvoices] = useState<InvoiceRecord[]>([])
  const [controls, setControls] = useState<BuyerPaymentControls | null>(null)
  const [topUpSummary, setTopUpSummary] = useState<WalletTopUpSummary>(emptyTopUpSummary)
  const [x402IntentSummary, setX402IntentSummary] = useState<X402IntentSummary>(emptyX402IntentSummary)
  const [isSavingCaps, setIsSavingCaps] = useState(false)
  const [isCreatingTopup, setIsCreatingTopup] = useState(false)
  const [settlingIntentId, setSettlingIntentId] = useState('')
  const [intentFeedback, setIntentFeedback] = useState<Record<string, IntentFeedback>>({})
  const [topupAmountUsd, setTopupAmountUsd] = useState(50)
  const [topupNotice, setTopupNotice] = useState('')
  const [onrampClientSecret, setOnrampClientSecret] = useState('')
  const [showOnrampEmbed, setShowOnrampEmbed] = useState(false)
  const [stripePublishableKey, setStripePublishableKey] = useState('')

  const allowedMethods = useMemo(
    () => controls?.policy?.allowedMethods || billing.allowedMethods || [],
    [billing.allowedMethods, controls?.policy?.allowedMethods],
  )
  const supportedMethods = useMemo(() => controls?.methods || [], [controls?.methods])
  const paymentMethodLabels = useMemo(
    () => new Map(supportedMethods.map(method => [method.id, getPaymentMethodTitle(method)])),
    [supportedMethods],
  )
  const readyAllowedMethods = useMemo(
    () =>
      supportedMethods.filter(method => getPaymentMethodStatus(method, allowedMethods).status === 'ready'),
    [allowedMethods, supportedMethods],
  )
  const walletBalanceUsdc = Number(
    controls?.wallet?.balanceUsdc ?? controls?.policy?.walletBalanceUsdc ?? billing.wallet?.balanceUsdc ?? billing.currentBalance ?? 0,
  )
  const managedWalletAddress =
    controls?.managedWallet?.address || controls?.policy?.walletAddress || controls?.policy?.siwxWallet || billing.wallet?.walletAddress || ''
  const stripeMethod = supportedMethods.find(method => method.id === 'stripe_onramp')
  const stripeTopUpReady = Boolean(
    topUpSummary.stripeConfigured && stripeMethod?.enabled && stripeMethod.configured,
  )

  const refreshWalletData = useCallback(async () => {
    const [bill, ctl, topupRes, x402Res, runtimeConfig] = await Promise.all([
      fetchBilling(),
      fetchBuyerPaymentControls(),
      fetchBuyerWalletTopUps(),
      fetchX402Intents(),
      fetchRuntimeConfig(),
    ])
    setBilling(bill)
    setControls(ctl)
    setTopUpSummary(topupRes)
    setX402IntentSummary(x402Res)
    setStripePublishableKey(runtimeConfig.stripe?.publishableKey || '')
    if (topupRes.defaultTopUpUsd) {
      setTopupAmountUsd(Number(topupRes.defaultTopUpUsd))
    }
  }, [])

  useEffect(() => {
    const loadData = async () => {
      setIsLoading(true)
      try {
        const [bill, invs, ctl, topupRes, x402Res, runtimeConfig] = await Promise.all([
          fetchBilling(),
          fetchInvoices(),
          fetchBuyerPaymentControls(),
          fetchBuyerWalletTopUps(),
          fetchX402Intents(),
          fetchRuntimeConfig(),
        ])
        setBilling(bill)
        setInvoices(invs)
        setControls(ctl)
        setTopUpSummary(topupRes)
        setX402IntentSummary(x402Res)
        setStripePublishableKey(runtimeConfig.stripe?.publishableKey || '')
        if (topupRes.defaultTopUpUsd) {
          setTopupAmountUsd(Number(topupRes.defaultTopUpUsd))
        }
      } finally {
        setIsLoading(false)
      }
    }

    loadData()
  }, [])

  if (isLoading) {
    return (
      <AppShell role="buyer">
        <LoadingState />
      </AppShell>
    )
  }

  const handleSettleWalletIntent = async (intent: X402Intent) => {
    setSettlingIntentId(intent.id)
    setIntentFeedback(prev => {
      const next = { ...prev }
      delete next[intent.id]
      return next
    })

    try {
      const settled = await settleX402Intent(intent.id)
      if (settled?.status !== 'settled') {
        const message = 'Settlement did not complete. Intent remains pending.'
        setIntentFeedback(prev => ({ ...prev, [intent.id]: { tone: 'error', message } }))
        toast.error(message)
        return
      }

      await refreshWalletData()
      const settledAt = formatDateTime(settled?.settledAt, 'just now')
      const message = `Wallet debit settled at ${settledAt}.`
      setIntentFeedback(prev => ({ ...prev, [intent.id]: { tone: 'success', message } }))
      toast.success(message)
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Intent settlement failed.'
      setIntentFeedback(prev => ({ ...prev, [intent.id]: { tone: 'error', message } }))
      toast.error(message)
    } finally {
      setSettlingIntentId('')
    }
  }

  return (
    <AppShell role="buyer">
      <div className="p-6 space-y-8">
        <div>
          <h1 className="text-3xl font-bold mb-2">Billing & Invoices</h1>
          <p className="text-muted-foreground">Manage your wallet, billing limits, and the real payment methods exposed by the backend.</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <Card className="lg:col-span-2 p-8">
            <h2 className="text-2xl font-bold mb-4 capitalize">{billing.plan} Plan</h2>
            <p className="text-sm text-muted-foreground mb-6">
              {planFeatures[billing.plan as keyof typeof planFeatures]?.description || 'Billing plan details'}
            </p>

            {billing.status === 'past_due' && (
              <div className="mb-6 p-4 bg-red-500/10 border border-red-200 dark:border-red-800 rounded-lg flex items-start gap-3">
                <AlertCircle className="w-5 h-5 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="font-semibold text-red-900 dark:text-red-100">Payment Due</p>
                  <p className="text-sm text-red-800 dark:text-red-200 mt-1">
                    Paid usage may pause until your allowed payment methods are ready again.
                  </p>
                </div>
              </div>
            )}

            <div className="space-y-6">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <div>
                  <p className="text-sm text-muted-foreground mb-1">Monthly Cost</p>
                  <p className="text-3xl font-bold">
                    ${' '}
                    {(planFeatures[billing.plan as keyof typeof planFeatures]?.price || 0).toFixed(2)}
                  </p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground mb-1">Next Billing Date</p>
                  <p className="text-3xl font-bold">{billing.nextBillingDate.toLocaleDateString()}</p>
                </div>
              </div>

              <div>
                <h3 className="font-semibold mb-3">Plan Features</h3>
                <ul className="space-y-2">
                  {(planFeatures[billing.plan as keyof typeof planFeatures]?.features || []).map(feature => (
                    <li key={feature} className="flex items-center gap-2 text-sm">
                      <CheckCircle2 className="w-4 h-4 text-green-600 dark:text-green-400 flex-shrink-0" />
                      {feature}
                    </li>
                  ))}
                </ul>
              </div>

              <Button variant="outline" className="w-full">
                Change Plan
              </Button>
            </div>
          </Card>

          <Card className="p-8 space-y-5">
            <div>
              <h3 className="text-xl font-bold">Payment Readiness</h3>
              <p className="text-sm text-muted-foreground mt-1">
                Buyer payment methods come from `/v1/buyer/payments/controls`, not from placeholder card data.
              </p>
            </div>

            <div className="rounded-lg border border-border bg-muted/40 p-4 space-y-2">
              <div className="flex items-center gap-2">
                <Wallet className="w-4 h-4 text-muted-foreground" />
                <p className="font-medium text-sm">Wallet balance</p>
              </div>
              <p className="text-2xl font-bold">{walletBalanceUsdc.toFixed(2)} USDC</p>
              <p className="text-xs text-muted-foreground">
                Funding method: {controls?.wallet?.fundingMethod || controls?.policy?.fundingMethod || billing.wallet?.fundingMethod || 'Not set'}
              </p>
              {managedWalletAddress && (
                <p className="text-xs text-muted-foreground break-all">
                  Marketplace wallet: {managedWalletAddress}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                Backend: {controls?.walletConfig?.activeProvider || controls?.managedWallet?.provider || 'not selected'} | Auto-pay {controls?.walletConfig?.managedAutoPayEnabled ? 'on' : 'off'} | Legacy mode {controls?.walletConfig?.legacyPaymentModeEnabled ? 'on' : 'off'}
              </p>
              <p className="text-xs text-muted-foreground">
                External wallets: {controls?.walletConfig?.externalWalletsEnabled ? 'advanced mode exposed' : 'hidden'}
              </p>
            </div>

            {supportedMethods.length === 0 ? (
              <div className="text-center py-6 space-y-3">
                <AlertCircle className="w-10 h-10 text-muted-foreground/60 mx-auto" />
                <p className="text-sm text-muted-foreground">No payment methods were returned by the buyer payments API.</p>
              </div>
            ) : (
              <div className="space-y-3">
                {supportedMethods.map(method => (
                  <PaymentMethodStatusCard
                    key={method.id}
                    method={method}
                    allowedMethods={allowedMethods}
                  />
                ))}
              </div>
            )}

            <div className="rounded-lg border border-dashed border-border p-4 space-y-2">
              <p className="font-medium text-sm">Allowed for buyer payments</p>
              {allowedMethods.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {allowedMethods.map(methodId => {
                    const method = supportedMethods.find(item => item.id === methodId)
                    return (
                      <Badge key={methodId} variant="outline">
                        {method ? getPaymentMethodTitle(method) : methodId}
                      </Badge>
                    )
                  })}
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">No buyer payment methods are currently allowed.</p>
              )}
              {readyAllowedMethods.length === 0 && supportedMethods.length > 0 && (
                <p className="text-xs text-amber-600 dark:text-amber-400">
                  No allowed method is ready yet. Paid installs may require wallet funding or external settlement after you update controls.
                </p>
              )}
            </div>
          </Card>
        </div>

        <Card className="bg-gradient-to-br from-primary/5 to-primary/10 border-primary/20 p-8">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            <div>
              <p className="text-sm text-muted-foreground mb-2">Current Balance</p>
              <p className="text-4xl font-bold mb-4">${billing.currentBalance.toFixed(2)}</p>
              <p className="text-sm text-muted-foreground">
                Available credit to use for future charges.
              </p>
              <p className="text-xs text-muted-foreground mt-2">
                Minimum balance target: ${Number(controls?.policy?.minimumBalanceUsdc || billing?.caps?.minimumBalanceUsdc || 0).toFixed(2)}
              </p>
            </div>
            <div className="space-y-3">
              <p className="text-xs text-muted-foreground">Top up with Stripe onramp when the backend reports it as configured.</p>
              <div className="flex gap-2">
                <Input
                  type="number"
                  min={topUpSummary.minimumTopUpUsd || 1}
                  value={topupAmountUsd}
                  onChange={e => setTopupAmountUsd(Number(e.target.value || 0))}
                />
                <Button
                  size="lg"
                  disabled={isCreatingTopup || !stripeTopUpReady || topupAmountUsd < (topUpSummary.minimumTopUpUsd || 1)}
                  onClick={async () => {
                    setIsCreatingTopup(true)
                    setTopupNotice('')
                    try {
                      const res = await createStripeTopUpSession({
                        amountUsd: Number(topupAmountUsd || 0),
                        walletAddress: managedWalletAddress,
                        paymentMethod: 'stripe_onramp',
                      }) as StripeTopUpSessionResponse
                      if (res?.topup) {
                        setTopUpSummary(prev => ({
                          ...prev,
                          items: [res.topup as WalletTopUp, ...prev.items].slice(0, 20),
                          count: prev.count + 1,
                        }))
                      }
                      if (res?.stripe?.clientSecret && stripePublishableKey) {
                        setOnrampClientSecret(String(res.stripe.clientSecret))
                        setShowOnrampEmbed(true)
                        setTopupNotice('Stripe onramp session created. Complete payment below.')
                      } else if (res?.stripe?.hostedUrl) {
                        window.open(res.stripe.hostedUrl, '_blank', 'noopener,noreferrer')
                        setTopupNotice('Stripe session created. Complete checkout in the opened window.')
                      } else if (res?.stripe?.clientSecret) {
                        setTopupNotice('Stripe session created but no runtime Stripe publishable key is configured for the embed.')
                      } else {
                        setTopupNotice('Top-up intent created.')
                      }
                    } catch (err: unknown) {
                      setTopupNotice(err instanceof Error ? err.message : 'Failed to create top-up session.')
                    } finally {
                      setIsCreatingTopup(false)
                    }
                  }}
                >
                  {isCreatingTopup ? 'Creating...' : 'Top Up'}
                </Button>
              </div>
              {!stripeTopUpReady && (
                <p className="text-xs text-amber-600 dark:text-amber-400">
                  Stripe onramp is not ready in backend payment controls yet, so wallet top-up is disabled here.
                </p>
              )}
              {topupNotice && <p className="text-xs text-muted-foreground">{topupNotice}</p>}
            </div>
          </div>
        </Card>

        {showOnrampEmbed && onrampClientSecret && (
          <Card className="p-6 space-y-4">
            <div className="flex items-center justify-between gap-3">
              <div>
                <h3 className="text-lg font-semibold">Complete Stripe Onramp</h3>
                <p className="text-xs text-muted-foreground">
                  This is a one-time wallet funding flow. No subscription is created.
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setShowOnrampEmbed(false)
                  setOnrampClientSecret('')
                }}
              >
                Close
              </Button>
            </div>
            {stripePublishableKey ? (
              <StripeOnrampWidget
                publishableKey={stripePublishableKey}
                clientSecret={onrampClientSecret}
                onSessionUpdate={async session => {
                  const status = String(session?.status || '').toLowerCase()
                  if (status === 'fulfillment_complete') {
                    setTopupNotice('Top-up completed. Wallet balance refreshed.')
                    await refreshWalletData()
                    setShowOnrampEmbed(false)
                    setOnrampClientSecret('')
                  } else if (status === 'rejected') {
                    setTopupNotice('Stripe onramp was rejected. No funds were added.')
                  }
                }}
              />
            ) : (
              <p className="text-sm text-muted-foreground">
                Missing runtime Stripe publishable key, so the embedded onramp cannot render.
              </p>
            )}
          </Card>
        )}

        {controls && (
          <Card className="p-6 space-y-5">
            <div>
              <h3 className="text-xl font-bold">Payment Controls</h3>
              <p className="text-sm text-muted-foreground mt-1">
                Set spend limits and choose the backend-exposed methods you actually want to allow for buyer payments.
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <p className="text-xs text-muted-foreground mb-1">Per Call Cap (USDC)</p>
                <Input
                  type="number"
                  value={controls.policy?.perCallCapUsdc ?? ''}
                  onChange={e => setControls(prev => prev ? ({ ...prev, policy: { ...prev.policy, perCallCapUsdc: Number(e.target.value || 0) } }) : prev)}
                />
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-1">Daily Cap (USDC)</p>
                <Input
                  type="number"
                  value={controls.policy?.dailySpendCapUsdc ?? ''}
                  onChange={e => setControls(prev => prev ? ({ ...prev, policy: { ...prev.policy, dailySpendCapUsdc: Number(e.target.value || 0) } }) : prev)}
                />
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-1">Monthly Cap (USDC)</p>
                <Input
                  type="number"
                  value={controls.policy?.monthlySpendCapUsdc ?? ''}
                  onChange={e => setControls(prev => prev ? ({ ...prev, policy: { ...prev.policy, monthlySpendCapUsdc: Number(e.target.value || 0) } }) : prev)}
                />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <p className="text-xs text-muted-foreground mb-1">Minimum Wallet Balance (USDC)</p>
                <Input
                  type="number"
                  value={controls.policy?.minimumBalanceUsdc ?? 0}
                  onChange={e => setControls(prev => prev ? ({ ...prev, policy: { ...prev.policy, minimumBalanceUsdc: Number(e.target.value || 0) } }) : prev)}
                />
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-1">Auto Top-Up Trigger (USD)</p>
                <Input
                  type="number"
                  value={controls.policy?.autoTopUpTriggerUsd ?? 0}
                  onChange={e => setControls(prev => prev ? ({ ...prev, policy: { ...prev.policy, autoTopUpTriggerUsd: Number(e.target.value || 0) } }) : prev)}
                />
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-1">Auto Top-Up Amount (USD)</p>
                <Input
                  type="number"
                  value={controls.policy?.autoTopUpAmountUsd ?? 0}
                  onChange={e => setControls(prev => prev ? ({ ...prev, policy: { ...prev.policy, autoTopUpAmountUsd: Number(e.target.value || 0) } }) : prev)}
                />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-muted-foreground mb-1">Marketplace Wallet Address</p>
                <Input
                  value={managedWalletAddress}
                  readOnly
                />
              </div>
              <div className="space-y-2">
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={Boolean(controls.policy?.hardStopOnLowFunds)}
                    onChange={e => setControls(prev => prev ? ({ ...prev, policy: { ...prev.policy, hardStopOnLowFunds: e.target.checked } }) : prev)}
                  />
                  Block paid tool calls if minimum balance would be breached
                </label>
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={Boolean(controls.policy?.autoTopUpEnabled)}
                    onChange={e => setControls(prev => prev ? ({ ...prev, policy: { ...prev.policy, autoTopUpEnabled: e.target.checked } }) : prev)}
                  />
                  Enable auto top-up rule (non-subscription threshold based)
                </label>
              </div>
            </div>

            <div className="space-y-3">
              <div>
                <p className="font-semibold text-sm">Allowed Payment Methods</p>
                <p className="text-xs text-muted-foreground mt-1">
                  These checkboxes write directly to `policy.allowedMethods` and only enable methods the backend says are active.
                </p>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                {supportedMethods.map(method => {
                  const readiness = getPaymentMethodStatus(method, allowedMethods)
                  const checked = allowedMethods.includes(method.id)

                  return (
                    <label
                      key={method.id}
                      className="flex items-start gap-3 rounded-lg border border-border p-4"
                    >
                      <input
                        type="checkbox"
                        checked={checked}
                        disabled={!method.enabled}
                        onChange={event => {
                          const nextAllowedMethods = event.target.checked
                            ? [...new Set([...allowedMethods, method.id])]
                            : allowedMethods.filter(item => item !== method.id)
                          setControls(prev => prev ? ({ ...prev, policy: { ...prev.policy, allowedMethods: nextAllowedMethods } }) : prev)
                        }}
                      />
                      <div className="space-y-2">
                        <div className="flex items-center gap-2 flex-wrap">
                          <p className="font-medium text-sm">{getPaymentMethodTitle(method)}</p>
                          <Badge variant={readiness.tone}>{readiness.label}</Badge>
                        </div>
                        <p className="text-xs text-muted-foreground">{getPaymentMethodDescription(method)}</p>
                      </div>
                    </label>
                  )
                })}
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <span>Daily spend: {Number(controls.dailySpendUsdc || 0).toFixed(2)} USDC</span>
              <span>Monthly spend: {Number(controls.monthlySpendUsdc || 0).toFixed(2)} USDC</span>
              <span>Facilitator mode: {controls.facilitatorMode || 'not configured'}</span>
              <span>Wallet balance: {Number(controls.wallet?.balanceUsdc ?? controls.policy?.walletBalanceUsdc ?? 0).toFixed(2)} USDC</span>
            </div>

            <div className="flex gap-2">
              <Button
                disabled={isSavingCaps}
                onClick={async () => {
                  setIsSavingCaps(true)
                  try {
                    const updated = await updateBuyerPaymentControls({
                      perCallCapUsdc: controls.policy?.perCallCapUsdc,
                      dailySpendCapUsdc: controls.policy?.dailySpendCapUsdc,
                      monthlySpendCapUsdc: controls.policy?.monthlySpendCapUsdc,
                      allowedMethods: controls.policy?.allowedMethods || [],
                      siwxWallet: controls.policy?.siwxWallet || '',
                      minimumBalanceUsdc: controls.policy?.minimumBalanceUsdc || 0,
                      hardStopOnLowFunds: Boolean(controls.policy?.hardStopOnLowFunds),
                      autoTopUpEnabled: Boolean(controls.policy?.autoTopUpEnabled),
                      autoTopUpAmountUsd: controls.policy?.autoTopUpAmountUsd || 0,
                      autoTopUpTriggerUsd: controls.policy?.autoTopUpTriggerUsd || 0,
                      fundingMethod: controls.policy?.fundingMethod || 'stripe_onramp',
                      walletAddress: controls.policy?.walletAddress || '',
                    })
                    setControls(prev => prev ? ({ ...prev, ...(updated as BuyerPaymentControls) }) : (updated as BuyerPaymentControls))
                  } finally {
                    setIsSavingCaps(false)
                  }
                }}
              >
                {isSavingCaps ? 'Saving...' : 'Save Caps'}
              </Button>
            </div>
          </Card>
        )}

        <Card className="p-6 space-y-4">
          <h3 className="text-xl font-bold">Wallet Top-Up History</h3>
          {topUpSummary.items.length === 0 ? (
            <p className="text-sm text-muted-foreground">No top-ups yet.</p>
          ) : (
            <div className="space-y-2">
              {topUpSummary.items.slice(0, 10).map(item => (
                <div key={item.id} className="border border-border rounded-md p-3 text-sm flex flex-col md:flex-row md:items-center md:justify-between gap-2">
                  <div>
                    <p className="font-semibold">{item.provider} {item.status}</p>
                    <p className="text-xs text-muted-foreground">
                      {new Date(item.createdAt).toLocaleString()} | ${Number(item.sourceAmount || 0).toFixed(2)} {item.sourceCurrency || 'USD'}
                      {item.destinationAmount ? ` -> ${Number(item.destinationAmount).toFixed(2)} ${item.destinationAsset || 'USDC'}` : ''}
                    </p>
                  </div>
                  {item.hostedUrl && (
                    <Button size="sm" variant="outline" onClick={() => window.open(item.hostedUrl, '_blank', 'noopener,noreferrer')}>
                      Continue
                    </Button>
                  )}
                </div>
              ))}
            </div>
          )}
        </Card>

        <Card className="p-6 space-y-4">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h3 className="text-xl font-bold">x402 Intent History</h3>
              <p className="text-sm text-muted-foreground mt-1">
                Recovery state comes from `/v1/billing/x402/intents`. Pending wallet debits can be retried here when funding becomes available.
              </p>
            </div>
            <Badge variant="outline">{x402IntentSummary.count} total</Badge>
          </div>

          {x402IntentSummary.items.length === 0 ? (
            <p className="text-sm text-muted-foreground">No x402 billing intents yet.</p>
          ) : (
            <div className="space-y-3">
              {x402IntentSummary.items.map(intent => {
                const challenge = parseIntentChallenge(intent)
                const paymentMethodLabel = paymentMethodLabels.get(intent.paymentMethod || '') || humanizeMethodId(intent.paymentMethod)
                const createdAt = formatDateTime(intent.createdAt)
                const settledAt = formatDateTime(intent.settledAt, intent.status === 'settled' ? 'Timestamp unavailable' : 'Not settled')
                const challengeContext = [
                  { label: 'Tool', value: String(challenge?.toolName || intent.toolName || '') },
                  { label: 'Server', value: String(challenge?.serverSlug || challenge?.serverId || intent.serverId || '') },
                  { label: 'Resource', value: String(challenge?.resource || intent.resource || '') },
                  { label: 'Settlement scope', value: String(challenge?.settlementScope || '') },
                  { label: 'Payment address', value: String(challenge?.paymentAddress || '') },
                  { label: 'Idempotency key', value: String(challenge?.idempotencyKey || intent.idempotencyKey || '') },
                ].filter(item => item.value)
                const canRetryWalletSettlement = intent.status === 'pending' && intent.paymentMethod === 'wallet_balance'
                const feedback = intentFeedback[intent.id]

                return (
                  <div key={intent.id} className="rounded-xl border border-border p-4 space-y-4">
                    <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                      <div className="space-y-2">
                        <div className="flex flex-wrap items-center gap-2">
                          <p className="font-semibold">{intent.toolName || 'Untitled intent'}</p>
                          <Badge variant={getIntentStatusVariant(intent.status)}>{intent.status}</Badge>
                          {intent.verificationStatus && <Badge variant="outline">Verification: {intent.verificationStatus}</Badge>}
                        </div>
                        <p className="text-xs text-muted-foreground break-all">Intent ID: {intent.id}</p>
                        <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm text-muted-foreground">
                          <span>Payment method: {paymentMethodLabel}</span>
                          <span>Amount: {Number(intent.amountUsdc || 0).toFixed(2)} USDC</span>
                          <span>Created: {createdAt}</span>
                          <span>Settled: {settledAt}</span>
                        </div>
                      </div>

                      <div className="flex flex-wrap gap-2">
                        {canRetryWalletSettlement ? (
                          <Button
                            size="sm"
                            variant="outline"
                            disabled={settlingIntentId === intent.id}
                            onClick={() => handleSettleWalletIntent(intent)}
                          >
                            <RefreshCw className={`w-4 h-4 mr-2 ${settlingIntentId === intent.id ? 'animate-spin' : ''}`} />
                            {settlingIntentId === intent.id ? 'Retrying...' : 'Retry Wallet Settle'}
                          </Button>
                        ) : intent.status === 'pending' ? (
                          <p className="text-xs text-amber-600 dark:text-amber-400 max-w-xs">
                            Pending external settlement. Reuse the original x402 challenge in the install flow or payer wallet before retrying.
                          </p>
                        ) : null}
                      </div>
                    </div>

                    {(challengeContext.length > 0 || intent.verificationNote || intent.paymentIdentifier || intent.facilitatorTx) && (
                      <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                        {challengeContext.length > 0 && (
                          <div className="rounded-lg bg-muted/40 p-3 space-y-2">
                            <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Requirement Context</p>
                            <div className="space-y-1 text-sm">
                              {challengeContext.map(item => (
                                <p key={`${intent.id}-${item.label}`} className="break-all">
                                  <span className="text-muted-foreground">{item.label}:</span> {item.value}
                                </p>
                              ))}
                            </div>
                          </div>
                        )}

                        {(intent.verificationNote || intent.paymentIdentifier || intent.facilitatorTx) && (
                          <div className="rounded-lg bg-muted/40 p-3 space-y-2">
                            <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Settlement Details</p>
                            <div className="space-y-1 text-sm">
                              {intent.verificationNote && <p>{intent.verificationNote}</p>}
                              {intent.paymentIdentifier && <p className="break-all"><span className="text-muted-foreground">Payment ID:</span> {intent.paymentIdentifier}</p>}
                              {intent.facilitatorTx && <p className="break-all"><span className="text-muted-foreground">Facilitator TX:</span> {intent.facilitatorTx}</p>}
                            </div>
                          </div>
                        )}
                      </div>
                    )}

                    {feedback && (
                      <div className={`rounded-lg border p-3 text-sm ${feedback.tone === 'error' ? 'border-red-200 bg-red-500/10 text-red-800 dark:border-red-900 dark:text-red-200' : 'border-green-200 bg-green-500/10 text-green-800 dark:border-green-900 dark:text-green-200'}`}>
                        {feedback.message}
                      </div>
                    )}

                    {intent.challenge && !challenge && (
                      <div className="rounded-lg border border-dashed border-border p-3 text-xs text-muted-foreground break-all">
                        Challenge payload is stored on this intent, but it could not be parsed into structured recovery details.
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </Card>

        <div className="space-y-4">
          <h2 className="text-2xl font-bold">Invoices</h2>

          <div className="space-y-2 overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left py-3 px-4 font-semibold text-muted-foreground">Date</th>
                  <th className="text-left py-3 px-4 font-semibold text-muted-foreground">Amount</th>
                  <th className="text-left py-3 px-4 font-semibold text-muted-foreground">Status</th>
                  <th className="text-left py-3 px-4 font-semibold text-muted-foreground">Action</th>
                </tr>
              </thead>
              <tbody>
                {invoices.map(invoice => (
                  <tr key={invoice.id} className="border-b border-border hover:bg-muted/50 transition-colors">
                    <td className="py-4 px-4">{new Date(invoice.date).toLocaleDateString()}</td>
                    <td className="py-4 px-4 font-semibold">${invoice.amount.toFixed(2)}</td>
                    <td className="py-4 px-4">
                      {invoice.status === 'paid' ? (
                        <span className="inline-flex items-center gap-1 text-xs bg-green-500/20 text-green-700 dark:text-green-400 px-2 py-1 rounded">
                          <CheckCircle2 className="w-3 h-3" />
                          Paid
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1 text-xs bg-amber-500/20 text-amber-700 dark:text-amber-400 px-2 py-1 rounded">
                          <AlertCircle className="w-3 h-3" />
                          {invoice.status === 'pending' ? 'Pending' : 'Failed'}
                        </span>
                      )}
                    </td>
                    <td className="py-4 px-4">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => {
                          const element = document.createElement('a')
                          element.setAttribute(
                            'href',
                            `data:text/plain;charset=utf-8,Invoice ${invoice.id}\nDate: ${new Date(invoice.date).toLocaleDateString()}\nAmount: $${invoice.amount.toFixed(2)}`,
                          )
                          element.setAttribute('download', `invoice-${invoice.id}.txt`)
                          element.style.display = 'none'
                          document.body.appendChild(element)
                          element.click()
                          document.body.removeChild(element)
                        }}
                      >
                        <Download className="w-4 h-4 mr-2" />
                        Download
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        <Card className="p-6 space-y-4">
          <h3 className="font-bold text-lg">Usage-Based Pricing</h3>
          <p className="text-sm text-muted-foreground">
            Some servers charge per API call or per transaction. Those charges respect your caps and allowed payment methods before they settle.
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4 border-t border-border">
            <div>
              <p className="text-sm text-muted-foreground mb-2">Current Spend</p>
              <p className="text-2xl font-bold">${Number(billing.monthlySpend || 0).toFixed(2)}</p>
              <p className="text-xs text-muted-foreground mt-1">This billing period</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground mb-2">Daily Spend</p>
              <p className="text-2xl font-bold">${Number(billing.dailySpend || 0).toFixed(2)}</p>
              <p className="text-xs text-muted-foreground mt-1">Current day</p>
            </div>
          </div>
        </Card>
      </div>
    </AppShell>
  )
}
