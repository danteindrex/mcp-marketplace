'use client'

import { useState, useEffect, useCallback } from 'react'
import { Download, CreditCard, AlertCircle, CheckCircle2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { AppShell } from '@/components/app-shell'
import { LoadingState } from '@/components/empty-state'
import {
  createStripeTopUpSession,
  fetchBilling,
  fetchBuyerPaymentControls,
  fetchBuyerWalletTopUps,
  fetchInvoices,
  updateBuyerPaymentControls,
} from '@/lib/api-client'
import { CurrencyTransfer } from '@/components/kokonut'
import { StripeOnrampWidget } from '@/components/stripe-onramp-widget'

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

export default function BillingPage() {
  const [isLoading, setIsLoading] = useState(true)
  const [billing, setBilling] = useState<any>({ plan: 'free', monthlySpend: 0, currentBalance: 0, nextBillingDate: new Date(), status: 'active' })
  const [invoices, setInvoices] = useState<any[]>([])
  const [controls, setControls] = useState<any | null>(null)
  const [topups, setTopups] = useState<any[]>([])
  const [isSavingCaps, setIsSavingCaps] = useState(false)
  const [isCreatingTopup, setIsCreatingTopup] = useState(false)
  const [topupAmountUsd, setTopupAmountUsd] = useState(50)
  const [topupNotice, setTopupNotice] = useState('')
  const [onrampClientSecret, setOnrampClientSecret] = useState('')
  const [showOnrampEmbed, setShowOnrampEmbed] = useState(false)

  const stripePublishableKey = process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY || ''

  const refreshWalletData = useCallback(async () => {
    const [bill, ctl, topupRes] = await Promise.all([
      fetchBilling(),
      fetchBuyerPaymentControls(),
      fetchBuyerWalletTopUps(),
    ])
    setBilling(bill as any)
    setControls(ctl as any)
    setTopups((topupRes as any).items || [])
  }, [])

  useEffect(() => {
    const loadData = async () => {
      setIsLoading(true)
      try {
        const [bill, invs, ctl] = await Promise.all([
          fetchBilling(),
          fetchInvoices(),
          fetchBuyerPaymentControls(),
        ])
        const topupRes = await fetchBuyerWalletTopUps()
        setBilling(bill as any)
        setInvoices(invs as any)
        setControls(ctl as any)
        setTopups((topupRes as any).items || [])
        if ((topupRes as any).defaultTopUpUsd) {
          setTopupAmountUsd(Number((topupRes as any).defaultTopUpUsd))
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

  return (
    <AppShell role="buyer">
      <div className="p-6 space-y-8">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold mb-2">Billing & Invoices</h1>
          <p className="text-muted-foreground">Manage your subscription and payment information</p>
        </div>

        {/* Payment Demo */}
        <Card className="p-8 bg-gradient-to-br from-green-500/5 to-green-500/10 border-green-500/20">
          <h2 className="text-xl font-bold mb-6">Recent Payment</h2>
          <CurrencyTransfer
            fromAmount={99.00}
            fromCurrency="USD"
            toCurrency="USD"
            toAmount={99.00}
            conversionRate={1.0}
            isComplete={true}
          />
        </Card>

        {/* Current Plan */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <Card className="lg:col-span-2 p-8">
            <h2 className="text-2xl font-bold mb-4 capitalize">{billing.plan} Plan</h2>

            {billing.status === 'past_due' && (
              <div className="mb-6 p-4 bg-red-500/10 border border-red-200 dark:border-red-800 rounded-lg flex items-start gap-3">
                <AlertCircle className="w-5 h-5 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="font-semibold text-red-900 dark:text-red-100">Payment Due</p>
                  <p className="text-sm text-red-800 dark:text-red-200 mt-1">
                    Your payment is overdue. Please update your payment method to avoid service interruption.
                  </p>
                </div>
              </div>
            )}

            <div className="space-y-6">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <div>
                  <p className="text-sm text-muted-foreground mb-1">Monthly Cost</p>
                  <p className="text-3xl font-bold">
                    $
                    {(planFeatures[billing.plan as keyof typeof planFeatures]?.price || 0).toFixed(
                      2
                    )}
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
                  {(planFeatures[billing.plan as keyof typeof planFeatures]?.features || []).map(
                    (feature, i) => (
                      <li key={i} className="flex items-center gap-2 text-sm">
                        <CheckCircle2 className="w-4 h-4 text-green-600 dark:text-green-400 flex-shrink-0" />
                        {feature}
                      </li>
                    )
                  )}
                </ul>
              </div>

              <Button variant="outline" className="w-full">
                Change Plan
              </Button>
            </div>
          </Card>

          {/* Payment Method */}
          <Card className="p-8">
            <h3 className="text-xl font-bold mb-4">Payment Method</h3>

            {billing.paymentMethod ? (
              <div className="space-y-4">
                <div className="p-4 bg-muted rounded-lg border border-border">
                  <div className="flex items-start gap-3">
                    <CreditCard className="w-6 h-6 text-muted-foreground mt-1" />
                    <div>
                      <p className="font-semibold text-sm">Credit Card</p>
                      <p className="text-sm text-muted-foreground">{billing.paymentMethod}</p>
                      <p className="text-xs text-muted-foreground mt-1">Primary card</p>
                    </div>
                  </div>
                </div>

                <Button variant="outline" className="w-full">
                  Update Payment Method
                </Button>

                <Button variant="ghost" className="w-full">
                  Add Another Card
                </Button>
              </div>
            ) : (
              <div className="text-center py-6 space-y-4">
                <AlertCircle className="w-12 h-12 text-muted-foreground/50 mx-auto" />
                <p className="text-sm text-muted-foreground">No payment method on file</p>
                <Button className="w-full">
                  <CreditCard className="w-4 h-4 mr-2" />
                  Add Payment Method
                </Button>
              </div>
            )}
          </Card>
        </div>

        {/* Account Balance */}
        <Card className="bg-gradient-to-br from-primary/5 to-primary/10 border-primary/20 p-8">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            <div>
              <p className="text-sm text-muted-foreground mb-2">Current Balance</p>
              <p className="text-4xl font-bold mb-4">${billing.currentBalance.toFixed(2)}</p>
              <p className="text-sm text-muted-foreground">
                Available credit to use for future charges
              </p>
              <p className="text-xs text-muted-foreground mt-2">
                Minimum balance target: ${Number(controls?.policy?.minimumBalanceUsdc || billing?.caps?.minimumBalanceUsdc || 0).toFixed(2)}
              </p>
            </div>
            <div className="space-y-3">
              <p className="text-xs text-muted-foreground">Top up with Stripe Onramp (one-time, no subscription)</p>
              <div className="flex gap-2">
                <Input
                  type="number"
                  min={1}
                  value={topupAmountUsd}
                  onChange={e => setTopupAmountUsd(Number(e.target.value || 0))}
                />
                <Button
                  size="lg"
                  disabled={isCreatingTopup}
                  onClick={async () => {
                    setIsCreatingTopup(true)
                    setTopupNotice('')
                    try {
                      const res: any = await createStripeTopUpSession({
                        amountUsd: Number(topupAmountUsd || 0),
                        walletAddress: controls?.policy?.walletAddress || controls?.policy?.siwxWallet || '',
                        paymentMethod: 'stripe_onramp',
                      })
                      if (res?.topup) {
                        setTopups((prev: any[]) => [res.topup, ...prev].slice(0, 20))
                      }
                      if (res?.stripe?.clientSecret && stripePublishableKey) {
                        setOnrampClientSecret(String(res.stripe.clientSecret))
                        setShowOnrampEmbed(true)
                        setTopupNotice('Stripe onramp session created. Complete payment below.')
                      } else if (res?.stripe?.hostedUrl) {
                        window.open(res.stripe.hostedUrl, '_blank', 'noopener,noreferrer')
                        setTopupNotice('Stripe session created. Complete checkout in the opened window.')
                      } else if (res?.stripe?.clientSecret) {
                        setTopupNotice('Stripe session created but publishable key is missing for embed.')
                      } else {
                        setTopupNotice('Top-up intent created.')
                      }
                    } catch (err: any) {
                      setTopupNotice(err?.message || 'Failed to create top-up session.')
                    } finally {
                      setIsCreatingTopup(false)
                    }
                  }}
                >
                  {isCreatingTopup ? 'Creating...' : 'Top Up'}
                </Button>
              </div>
              {topupNotice && (
                <p className="text-xs text-muted-foreground">{topupNotice}</p>
              )}
            </div>
          </div>
        </Card>

        {showOnrampEmbed && onrampClientSecret && (
          <Card className="p-6 space-y-4">
            <div className="flex items-center justify-between gap-3">
              <div>
                <h3 className="text-lg font-semibold">Complete Stripe Onramp</h3>
                <p className="text-xs text-muted-foreground">
                  This is a one-time top-up flow. No subscription is created.
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
                onSessionUpdate={async (session) => {
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
                Missing `NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY`, cannot render embedded onramp.
              </p>
            )}
          </Card>
        )}

        {/* Payment Controls */}
        {controls && (
          <Card className="p-6 space-y-5">
            <div>
              <h3 className="text-xl font-bold">Payment Controls (Caps)</h3>
              <p className="text-sm text-muted-foreground mt-1">
                Set hard limits for AI spending and choose which payment methods are allowed.
              </p>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <p className="text-xs text-muted-foreground mb-1">Per Call Cap (USDC)</p>
                <Input
                  type="number"
                  value={controls.policy?.perCallCapUsdc ?? ''}
                  onChange={e => setControls((prev: any) => ({ ...prev, policy: { ...prev.policy, perCallCapUsdc: Number(e.target.value || 0) } }))}
                />
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-1">Daily Cap (USDC)</p>
                <Input
                  type="number"
                  value={controls.policy?.dailySpendCapUsdc ?? ''}
                  onChange={e => setControls((prev: any) => ({ ...prev, policy: { ...prev.policy, dailySpendCapUsdc: Number(e.target.value || 0) } }))}
                />
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-1">Monthly Cap (USDC)</p>
                <Input
                  type="number"
                  value={controls.policy?.monthlySpendCapUsdc ?? ''}
                  onChange={e => setControls((prev: any) => ({ ...prev, policy: { ...prev.policy, monthlySpendCapUsdc: Number(e.target.value || 0) } }))}
                />
              </div>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <p className="text-xs text-muted-foreground mb-1">Minimum Wallet Balance (USDC)</p>
                <Input
                  type="number"
                  value={controls.policy?.minimumBalanceUsdc ?? 0}
                  onChange={e => setControls((prev: any) => ({ ...prev, policy: { ...prev.policy, minimumBalanceUsdc: Number(e.target.value || 0) } }))}
                />
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-1">Auto Top-Up Trigger (USD)</p>
                <Input
                  type="number"
                  value={controls.policy?.autoTopUpTriggerUsd ?? 0}
                  onChange={e => setControls((prev: any) => ({ ...prev, policy: { ...prev.policy, autoTopUpTriggerUsd: Number(e.target.value || 0) } }))}
                />
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-1">Auto Top-Up Amount (USD)</p>
                <Input
                  type="number"
                  value={controls.policy?.autoTopUpAmountUsd ?? 0}
                  onChange={e => setControls((prev: any) => ({ ...prev, policy: { ...prev.policy, autoTopUpAmountUsd: Number(e.target.value || 0) } }))}
                />
              </div>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-muted-foreground mb-1">Destination Wallet Address</p>
                <Input
                  value={controls.policy?.walletAddress || controls.policy?.siwxWallet || ''}
                  onChange={e => setControls((prev: any) => ({ ...prev, policy: { ...prev.policy, walletAddress: e.target.value } }))}
                />
              </div>
              <div className="space-y-2">
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={Boolean(controls.policy?.hardStopOnLowFunds)}
                    onChange={e => setControls((prev: any) => ({ ...prev, policy: { ...prev.policy, hardStopOnLowFunds: e.target.checked } }))}
                  />
                  Block paid tool calls if minimum balance would be breached
                </label>
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={Boolean(controls.policy?.autoTopUpEnabled)}
                    onChange={e => setControls((prev: any) => ({ ...prev, policy: { ...prev.policy, autoTopUpEnabled: e.target.checked } }))}
                  />
                  Enable auto top-up rule (non-subscription threshold based)
                </label>
              </div>
            </div>
            <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <span>Daily spend: {Number(controls.dailySpendUsdc || 0).toFixed(2)} USDC</span>
              <span>Monthly spend: {Number(controls.monthlySpendUsdc || 0).toFixed(2)} USDC</span>
              <span>Facilitator mode: {controls.facilitatorMode || 'mock'}</span>
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
                    setControls((prev: any) => ({ ...(prev || {}), ...(updated as any) }))
                  } finally {
                    setIsSavingCaps(false)
                  }
                }}
              >
                {isSavingCaps ? 'Saving...' : 'Save Caps'}
              </Button>
            </div>
            <div className="space-y-2">
              <p className="font-semibold text-sm">Supported Payment Methods</p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                {(controls.methods || []).map((m: any) => (
                  <div key={m.id} className="border border-border rounded-md p-3">
                    <p className="font-semibold text-sm">{m.displayName}</p>
                    <p className="text-xs text-muted-foreground mt-1">{m.integration}</p>
                    <p className="text-xs mt-1">{m.configured ? 'Configured' : 'Not configured'}</p>
                  </div>
                ))}
              </div>
            </div>
          </Card>
        )}

        {/* Top-Up History */}
        <Card className="p-6 space-y-4">
          <h3 className="text-xl font-bold">Wallet Top-Up History</h3>
          {topups.length === 0 ? (
            <p className="text-sm text-muted-foreground">No top-ups yet.</p>
          ) : (
            <div className="space-y-2">
              {topups.slice(0, 10).map((item: any) => (
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

        {/* Invoices */}
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
                            `data:text/plain;charset=utf-8,Invoice ${invoice.id}\nDate: ${new Date(invoice.date).toLocaleDateString()}\nAmount: $${invoice.amount.toFixed(2)}`
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

        {/* Metering Info */}
        <Card className="p-6 space-y-4">
          <h3 className="font-bold text-lg">Usage-Based Pricing</h3>
          <p className="text-sm text-muted-foreground">
            Some servers charge per API call or per transaction. Your usage is tracked in real-time and added to your
            invoice at the end of the billing cycle.
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4 border-t border-border">
            <div>
              <p className="text-sm text-muted-foreground mb-2">Current Metering Usage</p>
              <p className="text-2xl font-bold">1,247 calls</p>
              <p className="text-xs text-muted-foreground mt-1">This month</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground mb-2">Estimated Cost</p>
              <p className="text-2xl font-bold">$12.47</p>
              <p className="text-xs text-muted-foreground mt-1">Based on current usage</p>
            </div>
          </div>
        </Card>
      </div>
    </AppShell>
  )
}
