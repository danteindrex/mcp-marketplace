'use client'

import { useState } from 'react'
import Link from 'next/link'
import { ArrowLeft, Plus, Trash2, Copy } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { AppShell } from '@/components/app-shell'
import { toast } from 'sonner'

const pricingModels = [
  {
    id: 'free',
    label: 'Free',
    description: 'No charges for users',
    color: 'bg-green-500/10 border-green-200 dark:border-green-800',
  },
  {
    id: 'subscription',
    label: 'Subscription',
    description: 'Monthly or annual billing',
    color: 'bg-blue-500/10 border-blue-200 dark:border-blue-800',
  },
  {
    id: 'flat',
    label: 'Flat Rate',
    description: 'Fixed price per user/month',
    color: 'bg-purple-500/10 border-purple-200 dark:border-purple-800',
  },
  {
    id: 'x402',
    label: 'Per-Call (X402)',
    description: 'HTTP 402 micropayments per call',
    color: 'bg-amber-500/10 border-amber-200 dark:border-amber-800',
  },
]

export default function PricingPage({ params }: any) {
  const [pricingType, setPricingType] = useState('free')
  const [models, setModels] = useState<any[]>([
    {
      id: '1',
      type: 'free',
      displayName: 'Free Tier',
      basePrice: 0,
      perCallEnabled: false,
    },
  ])
  const [editingId, setEditingId] = useState<string | null>(null)

  const handleAddModel = () => {
    const newModel = {
      id: Date.now().toString(),
      type: pricingType,
      displayName: `${pricingType === 'x402' ? 'Per-Call Model' : 'Pricing Model'} ${models.length + 1}`,
      basePrice: 0,
      currency: 'USD',
      perCallEnabled: pricingType === 'x402',
      perCallAmount: pricingType === 'x402' ? 0.01 : undefined,
    }
    setModels([...models, newModel])
    toast.success('Pricing model added')
  }

  const handleDeleteModel = (id: string) => {
    setModels(models.filter(m => m.id !== id))
    toast.success('Pricing model removed')
  }

  const handleUpdateModel = (id: string, field: string, value: any) => {
    setModels(
      models.map(m => (m.id === id ? { ...m, [field]: value } : m))
    )
  }

  return (
    <AppShell role="merchant">
      <div className="max-w-4xl mx-auto p-6 space-y-6">
        {/* Header */}
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
            Set up how users will pay for your server
          </p>
        </div>

        {/* Pricing Model Selection */}
        <Card className="p-8">
          <h2 className="text-xl font-bold mb-6">Select a Pricing Model</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
            {pricingModels.map(model => (
              <button
                key={model.id}
                onClick={() => setPricingType(model.id)}
                className={`p-4 rounded-lg border-2 transition-all text-left ${
                  pricingType === model.id
                    ? `border-primary ${model.color}`
                    : 'border-border hover:border-primary/50'
                }`}
              >
                <p className="font-semibold text-sm">{model.label}</p>
                <p className="text-xs text-muted-foreground mt-1">
                  {model.description}
                </p>
              </button>
            ))}
          </div>

          <div className="bg-blue-500/10 border border-blue-200 dark:border-blue-800 rounded-lg p-4 text-sm text-blue-900 dark:text-blue-100">
            <p>
              {pricingType === 'free' &&
                'Your server will be available at no cost. You can add per-call charges later.'}
              {pricingType === 'subscription' &&
                'Charge users a recurring monthly or annual fee.'}
              {pricingType === 'flat' &&
                'Set a fixed monthly price per user or organization.'}
              {pricingType === 'x402' &&
                'Use HTTP 402 Payment Required for micropayments per API call.'}
            </p>
          </div>
        </Card>

        {/* Pricing Models */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-bold">Pricing Models</h2>
            <Button size="sm" onClick={handleAddModel}>
              <Plus className="w-4 h-4 mr-2" />
              Add Model
            </Button>
          </div>

          {models.length === 0 ? (
            <Card className="p-8 text-center text-muted-foreground">
              <p>No pricing models yet. Add one to get started.</p>
            </Card>
          ) : (
            <div className="space-y-4">
              {models.map(model => (
                <Card
                  key={model.id}
                  className="p-6 space-y-4"
                  onMouseEnter={() => setEditingId(model.id)}
                  onMouseLeave={() => setEditingId(null)}
                >
                  <div className="grid grid-cols-1 md:grid-cols-4 gap-4 items-end">
                    {/* Display Name */}
                    <div>
                      <Label className="text-xs mb-2 block">Model Name</Label>
                      <Input
                        value={model.displayName}
                        onChange={e =>
                          handleUpdateModel(model.id, 'displayName', e.target.value)
                        }
                        placeholder="e.g., Professional Plan"
                      />
                    </div>

                    {/* Type */}
                    <div>
                      <Label className="text-xs mb-2 block">Type</Label>
                      <Select
                        value={model.type}
                        onValueChange={value =>
                          handleUpdateModel(model.id, 'type', value)
                        }
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="free">Free</SelectItem>
                          <SelectItem value="subscription">Subscription</SelectItem>
                          <SelectItem value="flat">Flat Rate</SelectItem>
                          <SelectItem value="x402">Per-Call</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    {/* Base Price */}
                    {model.type !== 'free' && (
                      <div>
                        <Label className="text-xs mb-2 block">Base Price</Label>
                        <div className="flex gap-2">
                          <Input
                            type="number"
                            step="0.01"
                            value={model.basePrice}
                            onChange={e =>
                              handleUpdateModel(
                                model.id,
                                'basePrice',
                                parseFloat(e.target.value)
                              )
                            }
                            placeholder="0.00"
                          />
                          <Select
                            value={model.currency || 'USD'}
                            onValueChange={value =>
                              handleUpdateModel(model.id, 'currency', value)
                            }
                          >
                            <SelectTrigger className="w-20">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="USD">USD</SelectItem>
                              <SelectItem value="EUR">EUR</SelectItem>
                              <SelectItem value="GBP">GBP</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                      </div>
                    )}

                    {/* Actions */}
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => {
                          navigator.clipboard.writeText(JSON.stringify(model))
                          toast.success('Model copied')
                        }}
                      >
                        <Copy className="w-4 h-4" />
                      </Button>
                      <Button
                        size="sm"
                        variant="destructive"
                        onClick={() => handleDeleteModel(model.id)}
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>

                  {/* Per-Call Config */}
                  {model.type === 'x402' && (
                    <div className="border-t border-border pt-4 space-y-4">
                      <h4 className="font-semibold text-sm">Per-Call Configuration</h4>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                          <Label className="text-xs mb-2 block">Amount per Call</Label>
                          <Input
                            type="number"
                            step="0.0001"
                            value={model.perCallAmount || 0}
                            onChange={e =>
                              handleUpdateModel(
                                model.id,
                                'perCallAmount',
                                parseFloat(e.target.value)
                              )
                            }
                            placeholder="0.0001"
                          />
                        </div>
                        <div>
                          <Label className="text-xs mb-2 block">CAIP-2 Asset</Label>
                          <Input
                            value={model.caip2 || ''}
                            onChange={e =>
                              handleUpdateModel(model.id, 'caip2', e.target.value)
                            }
                            placeholder="eip155:1/erc20:0x..."
                          />
                        </div>
                        <div>
                          <Label className="text-xs mb-2 block">Payment Address</Label>
                          <Input
                            value={model.paymentAddress || ''}
                            onChange={e =>
                              handleUpdateModel(
                                model.id,
                                'paymentAddress',
                                e.target.value
                              )
                            }
                            placeholder="0x..."
                            className="font-mono text-xs"
                          />
                        </div>
                      </div>
                    </div>
                  )}
                </Card>
              ))}
            </div>
          )}
        </div>

        {/* Help Section */}
        <Card className="bg-blue-500/10 border-blue-200 dark:border-blue-800 p-6 space-y-3">
          <h3 className="font-semibold text-sm">Pricing Best Practices</h3>
          <ul className="text-sm space-y-2 text-foreground/80">
            <li>• Start with a free tier to attract early adopters</li>
            <li>• Use tiered pricing to segment different user groups</li>
            <li>• For X402, set micro-transaction amounts that are economically viable</li>
            <li>• Monitor adoption rates and adjust pricing quarterly</li>
            <li>• Document what each tier includes in your server description</li>
          </ul>
        </Card>

        {/* Actions */}
        <div className="flex gap-3">
          <Button variant="outline" asChild>
            <Link href="/merchant/servers">Cancel</Link>
          </Button>
          <Button className="flex-1">Save Pricing Configuration</Button>
        </div>
      </div>
    </AppShell>
  )
}
