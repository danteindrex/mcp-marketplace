'use client'

import { useEffect, useState } from 'react'
import { AppShell } from '@/components/app-shell'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  fetchAdminIntegrations,
  type PlatformIntegrationSettingsResponse,
  updateAdminIntegrations,
} from '@/lib/api-client'

type FormState = {
  google: { clientId: string; clientSecret: string; redirectBase: string }
  github: { clientId: string; clientSecret: string; redirectBase: string }
  stripe: {
    publishableKey: string
    secretKey: string
    webhookSecret: string
    onrampReturnUrl: string
    onrampRefreshUrl: string
    onrampMinUsd: number
    onrampDefaultUsd: number
    connectReturnUrl: string
    connectRefreshUrl: string
    connectWebhookSecret: string
  }
  x402: { mode: string; facilitatorUrl: string; facilitatorApiKey: string }
  n8n: { baseUrl: string; apiKey: string; timeoutSeconds: number }
}

const emptyForm: FormState = {
  google: { clientId: '', clientSecret: '', redirectBase: '' },
  github: { clientId: '', clientSecret: '', redirectBase: '' },
  stripe: {
    publishableKey: '',
    secretKey: '',
    webhookSecret: '',
    onrampReturnUrl: '',
    onrampRefreshUrl: '',
    onrampMinUsd: 10,
    onrampDefaultUsd: 50,
    connectReturnUrl: '',
    connectRefreshUrl: '',
    connectWebhookSecret: '',
  },
  x402: { mode: 'facilitator', facilitatorUrl: '', facilitatorApiKey: '' },
  n8n: { baseUrl: '', apiKey: '', timeoutSeconds: 12 },
}

function statusTone(source?: string) {
  if (source === 'ui') return 'default'
  if (source === 'env') return 'secondary'
  return 'outline'
}

function secretPlaceholder(masked?: string, isSet?: boolean) {
  if (masked) return masked
  if (isSet) return 'Already saved'
  return ''
}

export default function AdminIntegrationsPage() {
  const [data, setData] = useState<PlatformIntegrationSettingsResponse | null>(null)
  const [form, setForm] = useState<FormState>(emptyForm)
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)

  async function load() {
    setIsLoading(true)
    try {
      const res = await fetchAdminIntegrations()
      setData(res)
      setForm({
        google: {
          clientId: res.settings.google.clientId || '',
          clientSecret: '',
          redirectBase: res.settings.google.redirectBase || '',
        },
        github: {
          clientId: res.settings.github.clientId || '',
          clientSecret: '',
          redirectBase: res.settings.github.redirectBase || '',
        },
        stripe: {
          publishableKey: res.settings.stripe.publishableKey || '',
          secretKey: '',
          webhookSecret: '',
          onrampReturnUrl: res.settings.stripe.onrampReturnUrl || '',
          onrampRefreshUrl: res.settings.stripe.onrampRefreshUrl || '',
          onrampMinUsd: Number(res.settings.stripe.onrampMinUsd || 10),
          onrampDefaultUsd: Number(res.settings.stripe.onrampDefaultUsd || 50),
          connectReturnUrl: res.settings.stripe.connectReturnUrl || '',
          connectRefreshUrl: res.settings.stripe.connectRefreshUrl || '',
          connectWebhookSecret: '',
        },
        x402: {
          mode: res.settings.x402.mode || 'facilitator',
          facilitatorUrl: res.settings.x402.facilitatorUrl || '',
          facilitatorApiKey: '',
        },
        n8n: {
          baseUrl: res.settings.n8n.baseUrl || '',
          apiKey: '',
          timeoutSeconds: Number(res.settings.n8n.timeoutSeconds || 12),
        },
      })
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  return (
    <AppShell role="admin">
      <div className="p-6 space-y-6">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold mb-2">Platform Integrations</h1>
            <p className="text-muted-foreground">
              Manage OAuth, Stripe, x402, and n8n without rebuilding containers. Secrets are write-only; leave a secret field blank to keep the current value.
            </p>
          </div>
          <Button disabled={isSaving || isLoading} onClick={load} variant="outline">
            Refresh
          </Button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-5 gap-4">
          {['google', 'github', 'stripe', 'x402', 'n8n'].map(key => (
            <Card key={key} className="p-4 space-y-2">
              <p className="font-semibold capitalize">{key}</p>
              <Badge variant={statusTone(data?.status?.[key]?.source)}>
                {data?.status?.[key]?.configured ? 'Configured' : 'Not configured'}
              </Badge>
              <p className="text-xs text-muted-foreground">
                Source: {data?.status?.[key]?.source || 'none'}
              </p>
            </Card>
          ))}
        </div>

        <Card className="p-6 space-y-4">
          <h2 className="text-xl font-bold">Google OAuth</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
            <Input placeholder="Client ID" value={form.google.clientId} onChange={e => setForm(prev => ({ ...prev, google: { ...prev.google, clientId: e.target.value } }))} />
            <Input placeholder={secretPlaceholder(data?.settings.google.clientSecret.masked, data?.settings.google.clientSecret.set)} value={form.google.clientSecret} onChange={e => setForm(prev => ({ ...prev, google: { ...prev.google, clientSecret: e.target.value } }))} />
            <Input placeholder="Redirect base" value={form.google.redirectBase} onChange={e => setForm(prev => ({ ...prev, google: { ...prev.google, redirectBase: e.target.value } }))} />
          </div>
        </Card>

        <Card className="p-6 space-y-4">
          <h2 className="text-xl font-bold">GitHub OAuth</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
            <Input placeholder="Client ID" value={form.github.clientId} onChange={e => setForm(prev => ({ ...prev, github: { ...prev.github, clientId: e.target.value } }))} />
            <Input placeholder={secretPlaceholder(data?.settings.github.clientSecret.masked, data?.settings.github.clientSecret.set)} value={form.github.clientSecret} onChange={e => setForm(prev => ({ ...prev, github: { ...prev.github, clientSecret: e.target.value } }))} />
            <Input placeholder="Redirect base" value={form.github.redirectBase} onChange={e => setForm(prev => ({ ...prev, github: { ...prev.github, redirectBase: e.target.value } }))} />
          </div>
        </Card>

        <Card className="p-6 space-y-4">
          <h2 className="text-xl font-bold">Stripe</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
            <Input placeholder="Publishable key" value={form.stripe.publishableKey} onChange={e => setForm(prev => ({ ...prev, stripe: { ...prev.stripe, publishableKey: e.target.value } }))} />
            <Input placeholder={secretPlaceholder(data?.settings.stripe.secretKey.masked, data?.settings.stripe.secretKey.set)} value={form.stripe.secretKey} onChange={e => setForm(prev => ({ ...prev, stripe: { ...prev.stripe, secretKey: e.target.value } }))} />
            <Input placeholder={secretPlaceholder(data?.settings.stripe.webhookSecret.masked, data?.settings.stripe.webhookSecret.set)} value={form.stripe.webhookSecret} onChange={e => setForm(prev => ({ ...prev, stripe: { ...prev.stripe, webhookSecret: e.target.value } }))} />
            <Input placeholder="Onramp return URL" value={form.stripe.onrampReturnUrl} onChange={e => setForm(prev => ({ ...prev, stripe: { ...prev.stripe, onrampReturnUrl: e.target.value } }))} />
            <Input placeholder="Onramp refresh URL" value={form.stripe.onrampRefreshUrl} onChange={e => setForm(prev => ({ ...prev, stripe: { ...prev.stripe, onrampRefreshUrl: e.target.value } }))} />
            <Input type="number" placeholder="Onramp min USD" value={form.stripe.onrampMinUsd} onChange={e => setForm(prev => ({ ...prev, stripe: { ...prev.stripe, onrampMinUsd: Number(e.target.value || 0) } }))} />
            <Input type="number" placeholder="Onramp default USD" value={form.stripe.onrampDefaultUsd} onChange={e => setForm(prev => ({ ...prev, stripe: { ...prev.stripe, onrampDefaultUsd: Number(e.target.value || 0) } }))} />
            <Input placeholder="Connect return URL" value={form.stripe.connectReturnUrl} onChange={e => setForm(prev => ({ ...prev, stripe: { ...prev.stripe, connectReturnUrl: e.target.value } }))} />
            <Input placeholder="Connect refresh URL" value={form.stripe.connectRefreshUrl} onChange={e => setForm(prev => ({ ...prev, stripe: { ...prev.stripe, connectRefreshUrl: e.target.value } }))} />
            <Input placeholder={secretPlaceholder(data?.settings.stripe.connectWebhookSecret.masked, data?.settings.stripe.connectWebhookSecret.set)} value={form.stripe.connectWebhookSecret} onChange={e => setForm(prev => ({ ...prev, stripe: { ...prev.stripe, connectWebhookSecret: e.target.value } }))} />
          </div>
        </Card>

        <Card className="p-6 space-y-4">
          <h2 className="text-xl font-bold">x402 Facilitator</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
            <select
              value={form.x402.mode}
              onChange={e => setForm(prev => ({ ...prev, x402: { ...prev.x402, mode: e.target.value } }))}
              className="px-3 py-2 rounded-md border border-input bg-background text-sm w-full"
            >
              <option value="facilitator">facilitator</option>
              <option value="disabled">disabled</option>
              <option value="test">test</option>
            </select>
            <Input placeholder="Facilitator URL" value={form.x402.facilitatorUrl} onChange={e => setForm(prev => ({ ...prev, x402: { ...prev.x402, facilitatorUrl: e.target.value } }))} />
            <Input placeholder={secretPlaceholder(data?.settings.x402.facilitatorApiKey.masked, data?.settings.x402.facilitatorApiKey.set)} value={form.x402.facilitatorApiKey} onChange={e => setForm(prev => ({ ...prev, x402: { ...prev.x402, facilitatorApiKey: e.target.value } }))} />
          </div>
        </Card>

        <Card className="p-6 space-y-4">
          <h2 className="text-xl font-bold">n8n</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
            <Input placeholder="Base URL" value={form.n8n.baseUrl} onChange={e => setForm(prev => ({ ...prev, n8n: { ...prev.n8n, baseUrl: e.target.value } }))} />
            <Input placeholder={secretPlaceholder(data?.settings.n8n.apiKey.masked, data?.settings.n8n.apiKey.set)} value={form.n8n.apiKey} onChange={e => setForm(prev => ({ ...prev, n8n: { ...prev.n8n, apiKey: e.target.value } }))} />
            <Input type="number" placeholder="Timeout seconds" value={form.n8n.timeoutSeconds} onChange={e => setForm(prev => ({ ...prev, n8n: { ...prev.n8n, timeoutSeconds: Number(e.target.value || 0) } }))} />
          </div>
          <p className="text-xs text-muted-foreground">
            Agent Builder runtime URL: {data?.runtime?.n8n?.url || 'not configured'}
          </p>
        </Card>

        <div className="flex items-center justify-between gap-4">
          <p className="text-sm text-muted-foreground">
            Stripe publishable key and n8n URL are exposed through runtime config immediately after save.
          </p>
          <Button
            disabled={isSaving || isLoading}
            onClick={async () => {
              setIsSaving(true)
              try {
                const updated = await updateAdminIntegrations(form)
                setData(updated)
                setForm(prev => ({
                  ...prev,
                  google: { ...prev.google, clientSecret: '' },
                  github: { ...prev.github, clientSecret: '' },
                  stripe: { ...prev.stripe, secretKey: '', webhookSecret: '', connectWebhookSecret: '' },
                  x402: { ...prev.x402, facilitatorApiKey: '' },
                  n8n: { ...prev.n8n, apiKey: '' },
                }))
              } finally {
                setIsSaving(false)
              }
            }}
          >
            {isSaving ? 'Saving...' : 'Save Integrations'}
          </Button>
        </div>
      </div>
    </AppShell>
  )
}
