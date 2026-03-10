'use client'

import { use, useCallback, useEffect, useState } from 'react'
import Link from 'next/link'
import { ArrowLeft, Copy } from 'lucide-react'
import { AppShell } from '@/components/app-shell'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  fetchServerPricing,
  publishMerchantServer,
  updateMerchantServer,
  updateMerchantServerPaymentConfig,
  type MerchantServerPricingResponse,
} from '@/lib/api-client'
import { toast } from 'sonner'

type PricingFormState = {
  pricingType: string
  pricingAmount: string
  paymentAddress: string
  paymentMethods: string[]
  perCallCapUsdc: string
  dailyCapUsdc: string
  monthlyCapUsdc: string
}

function toNumber(value: string) {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

export default function PricingPage({ params }: { params: Promise<{ serverId: string }> }) {
  const { serverId } = use(params)
  const [data, setData] = useState<MerchantServerPricingResponse | null>(null)
  const [form, setForm] = useState<PricingFormState | null>(null)
  const [isSaving, setIsSaving] = useState(false)
  const [isPublishing, setIsPublishing] = useState(false)

  const load = useCallback(async () => {
    const next = await fetchServerPricing(serverId)
    setData(next)
    setForm({
      pricingType: next.pricing.type || 'x402',
      pricingAmount: String(next.pricing.amount ?? 0),
      paymentAddress: next.pricing.paymentAddress || '',
      paymentMethods: next.pricing.methods || [],
      perCallCapUsdc: String(next.pricing.caps?.perCallCapUsdc ?? 0),
      dailyCapUsdc: String(next.pricing.caps?.dailyCapUsdc ?? 0),
      monthlyCapUsdc: String(next.pricing.caps?.monthlyCapUsdc ?? 0),
    })
  }, [serverId])

  useEffect(() => {
    load().catch(error => {
      toast.error(error?.message || 'Failed to load pricing')
    })
  }, [load])

  const savePricing = useCallback(async () => {
    if (!form) return

    const pricingAmount = toNumber(form.pricingAmount)
    if (pricingAmount <= 0) {
      throw new Error('Enter a positive price before saving.')
    }

    await updateMerchantServer(serverId, {
      pricingType: form.pricingType,
      pricingAmount,
    })
    await updateMerchantServerPaymentConfig(serverId, {
      paymentAddress: form.paymentAddress.trim(),
      paymentMethods: form.paymentMethods,
      perCallCapUsdc: toNumber(form.perCallCapUsdc),
      dailyCapUsdc: toNumber(form.dailyCapUsdc),
      monthlyCapUsdc: toNumber(form.monthlyCapUsdc),
    })
    await load()
  }, [form, load, serverId])

  const handleSave = async () => {
    setIsSaving(true)
    try {
      await savePricing()
      toast.success('Pricing saved')
    } catch (error: any) {
      toast.error(error?.message || 'Save failed')
    } finally {
      setIsSaving(false)
    }
  }

  const handlePublish = async () => {
    if (!form) return

    setIsPublishing(true)
    try {
      await savePricing()
      await publishMerchantServer(serverId, {
        pricingType: form.pricingType,
        pricingAmount: toNumber(form.pricingAmount),
      })
      toast.success('Published to marketplace')
      await load()
    } catch (error: any) {
      toast.error(error?.message || 'Publish failed')
    } finally {
      setIsPublishing(false)
    }
  }

  if (!data || !form) {
    return (
      <AppShell role="merchant">
        <div className="p-6">Loading...</div>
      </AppShell>
    )
  }

  return (
    <AppShell role="merchant">
      <div className="max-w-4xl mx-auto p-6 space-y-6">
        <Link
          href="/merchant/servers"
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors w-fit"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Servers
        </Link>

        <div>
          <h1 className="text-3xl font-bold mb-2">Pricing Configuration</h1>
          <p className="text-muted-foreground">
            Set a real price, save payment constraints, then publish the server.
          </p>
        </div>

        <Card className="p-8 space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <Label htmlFor="pricing-type">Pricing Type</Label>
              <select
                id="pricing-type"
                value={form.pricingType}
                onChange={event =>
                  setForm(current => (current ? { ...current, pricingType: event.target.value } : current))
                }
                className="px-3 py-2 rounded-md border border-input bg-background text-sm w-full"
              >
                <option value="x402">x402</option>
                <option value="subscription">subscription</option>
                <option value="free">free</option>
              </select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="pricing-amount">Amount (USDC)</Label>
              <Input
                id="pricing-amount"
                type="number"
                min="0"
                step="0.01"
                value={form.pricingAmount}
                onChange={event =>
                  setForm(current => (current ? { ...current, pricingAmount: event.target.value } : current))
                }
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <Label htmlFor="payment-address">Payment Address</Label>
              <Input
                id="payment-address"
                placeholder="Wallet or payout address"
                value={form.paymentAddress}
                onChange={event =>
                  setForm(current => (current ? { ...current, paymentAddress: event.target.value } : current))
                }
              />
            </div>

            <div className="space-y-2">
              <Label>Enabled Payment Methods</Label>
              <div className="space-y-2 rounded-lg border border-border p-3">
                {data.supportedMethods.map(method => (
                  <label key={method.id} className="flex items-start gap-3 text-sm">
                    <input
                      type="checkbox"
                      checked={form.paymentMethods.includes(method.id)}
                      onChange={event =>
                        setForm(current => {
                          if (!current) return current
                          const paymentMethods = event.target.checked
                            ? [...current.paymentMethods, method.id]
                            : current.paymentMethods.filter(item => item !== method.id)
                          return { ...current, paymentMethods }
                        })
                      }
                      className="mt-1"
                    />
                    <div>
                      <p className="font-medium">{method.displayName}</p>
                      <p className="text-xs text-muted-foreground">
                        {method.integration}
                        {method.enabled ? '' : ' | disabled'}
                        {method.configured ? '' : ' | not configured'}
                      </p>
                    </div>
                  </label>
                ))}
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label htmlFor="per-call-cap">Per Call Cap</Label>
              <Input
                id="per-call-cap"
                type="number"
                min="0"
                step="0.01"
                value={form.perCallCapUsdc}
                onChange={event =>
                  setForm(current => (current ? { ...current, perCallCapUsdc: event.target.value } : current))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="daily-cap">Daily Cap</Label>
              <Input
                id="daily-cap"
                type="number"
                min="0"
                step="0.01"
                value={form.dailyCapUsdc}
                onChange={event =>
                  setForm(current => (current ? { ...current, dailyCapUsdc: event.target.value } : current))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="monthly-cap">Monthly Cap</Label>
              <Input
                id="monthly-cap"
                type="number"
                min="0"
                step="0.01"
                value={form.monthlyCapUsdc}
                onChange={event =>
                  setForm(current => (current ? { ...current, monthlyCapUsdc: event.target.value } : current))
                }
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <p className="text-xs text-muted-foreground mb-1">Marketplace Status</p>
              <p className="font-medium capitalize">{data.lifecycle?.marketplaceStatus || 'draft'}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground mb-1">Deployment Status</p>
              <p className="font-medium capitalize">{data.lifecycle?.deploymentStatus || 'not_deployed'}</p>
            </div>
          </div>

          <div className="border-t border-border pt-4 space-y-3">
            <p className="text-sm font-semibold">x402 Configuration</p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div>
                <p className="text-muted-foreground">Version</p>
                <p>{data.pricing.x402.version}</p>
              </div>
              <div>
                <p className="text-muted-foreground">Network</p>
                <p>{data.pricing.x402.network}</p>
              </div>
              <div>
                <p className="text-muted-foreground">Asset</p>
                <p>{data.pricing.x402.asset}</p>
              </div>
              <div>
                <p className="text-muted-foreground">CAIP-2</p>
                <p className="font-mono">{data.pricing.x402.caip2}</p>
              </div>
            </div>
            <Button
              variant="outline"
              onClick={() => {
                navigator.clipboard.writeText(JSON.stringify(data.pricing.x402))
                toast.success('x402 config copied')
              }}
            >
              <Copy className="w-4 h-4 mr-2" />
              Copy x402 Config
            </Button>
          </div>

          {(data.lifecycle?.blockingReasons || []).length > 0 && (
            <div className="rounded-lg border border-border bg-muted/40 p-4 space-y-2">
              <p className="font-medium">Publish blockers</p>
              <ul className="space-y-1 text-sm text-muted-foreground">
                {(data.lifecycle.blockingReasons || []).map(reason => (
                  <li key={`${reason.stage}-${reason.code}`}>{reason.message}</li>
                ))}
              </ul>
            </div>
          )}

          <div className="border-t border-border pt-4 flex flex-col gap-3 md:flex-row">
            <Button onClick={handleSave} disabled={isSaving || isPublishing}>
              {isSaving ? 'Saving...' : 'Save Pricing'}
            </Button>
            <Button onClick={handlePublish} disabled={isSaving || isPublishing}>
              {isPublishing ? 'Publishing...' : 'Save and Publish'}
            </Button>
          </div>
        </Card>
      </div>
    </AppShell>
  )
}
