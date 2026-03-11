'use client'

import { useMemo, useState } from 'react'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { AppShell } from '@/components/app-shell'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { createMerchantServer } from '@/lib/api-client'
import { toast } from 'sonner'

type CreateServerForm = {
  dockerImage: string
  containerPort: string
  name: string
  slug: string
  description: string
  category: string
  canonicalResourceUri: string
  requiredScopes: string
  pricingType: string
  pricingAmount: string
  supportsLocal: boolean
  supportsCloud: boolean
}

function toSlug(value: string) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 50)
}

function deriveServerName(imageRef: string) {
  const imageNameWithTag = imageRef.split('/').pop() || 'mcp-server'
  const imageName = imageNameWithTag.split(':')[0] || 'mcp-server'
  return (
    imageName
      .split(/[-_]/)
      .filter(Boolean)
      .map(part => part.charAt(0).toUpperCase() + part.slice(1))
      .join(' ')
      .trim() || 'MCP Server'
  )
}

function parseScopes(raw: string) {
  return raw
    .split(',')
    .map(scope => scope.trim())
    .filter(Boolean)
}

export default function ImportDockerPage() {
  const [form, setForm] = useState<CreateServerForm>({
    dockerImage: '',
    containerPort: '3000',
    name: '',
    slug: '',
    description: '',
    category: 'automation',
    canonicalResourceUri: '',
    requiredScopes: 'agent:invoke',
    pricingType: 'x402',
    pricingAmount: '1',
    supportsLocal: true,
    supportsCloud: true,
  })
  const [isSubmitting, setIsSubmitting] = useState(false)

  const parsedScopes = useMemo(() => parseScopes(form.requiredScopes), [form.requiredScopes])

  const applyImageDefaults = (value: string) => {
    const trimmed = value.trim()
    const nextName = form.name || deriveServerName(trimmed)
    const nextSlug = form.slug || toSlug(nextName)
    setForm(current => ({
      ...current,
      dockerImage: value,
      name: current.name || nextName,
      slug: current.slug || nextSlug,
      description:
        current.description || (nextName ? `${nextName} deployed from ${trimmed}` : ''),
    }))
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    const dockerImage = form.dockerImage.trim()
    const name = form.name.trim()
    const slug = toSlug(form.slug)
    const canonicalResourceUri = form.canonicalResourceUri.trim()
    const containerPort = Number(form.containerPort)
    const pricingAmount = Number(form.pricingAmount)

    if (!dockerImage || !name || !slug) {
      toast.error('Docker image, name, and slug are required.')
      return
    }
    if (!Number.isInteger(containerPort) || containerPort <= 0 || containerPort > 65535) {
      toast.error('Enter a valid exposed container port.')
      return
    }

    if (!Number.isFinite(pricingAmount) || pricingAmount <= 0) {
      toast.error('Enter a positive price before creating the server.')
      return
    }

    setIsSubmitting(true)
    try {
      const created = await createMerchantServer({
        name,
        slug,
        description: form.description.trim(),
        category: form.category.trim() || 'automation',
        dockerImage,
        containerPort,
        canonicalResourceUri,
        requiredScopes: parsedScopes,
        pricingType: form.pricingType,
        pricingAmount,
        supportsCloud: form.supportsCloud,
        supportsLocal: form.supportsLocal,
        status: 'draft',
      })
      toast.success('Server created as a draft. Continue to deployments.')
      window.location.href = `/merchant/servers/${created.id}/deployments`
    } catch (error: any) {
      toast.error(error?.message || 'Server creation failed')
    } finally {
      setIsSubmitting(false)
    }
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
          <h1 className="text-3xl font-bold mb-2">Create Server from Docker Image</h1>
          <p className="text-muted-foreground">
            Enter the real metadata the backend needs and create the draft directly. No mock scan
            or generated review data is used here.
          </p>
        </div>

        <Card className="p-8">
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="docker-image">Docker Image</Label>
              <Input
                id="docker-image"
                placeholder="docker.io/myorg/mcp-server:1.0.0"
                value={form.dockerImage}
                onChange={event => applyImageDefaults(event.target.value)}
                className="font-mono text-sm"
              />
              <p className="text-xs text-muted-foreground">
                Use a real image reference that your deployment workflow can pull.
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="container-port">Exposed Container Port</Label>
              <Input
                id="container-port"
                type="number"
                min="1"
                max="65535"
                value={form.containerPort}
                onChange={event =>
                  setForm(current => ({ ...current, containerPort: event.target.value }))
                }
              />
              <p className="text-xs text-muted-foreground">
                Required for real container deployment. For example, `3000` for `samanhappy/mcphub`.
              </p>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="server-name">Name</Label>
                <Input
                  id="server-name"
                  value={form.name}
                  onChange={event => setForm(current => ({ ...current, name: event.target.value }))}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="server-slug">Slug</Label>
                <Input
                  id="server-slug"
                  value={form.slug}
                  onChange={event =>
                    setForm(current => ({ ...current, slug: toSlug(event.target.value) }))
                  }
                />
              </div>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="category">Category</Label>
                <Input
                  id="category"
                  value={form.category}
                  onChange={event =>
                    setForm(current => ({ ...current, category: event.target.value }))
                  }
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="canonical-uri">Canonical Resource URI</Label>
                <Input
                  id="canonical-uri"
                  placeholder="Optional before deployment"
                  value={form.canonicalResourceUri}
                  onChange={event =>
                    setForm(current => ({ ...current, canonicalResourceUri: event.target.value }))
                  }
                />
                <p className="text-xs text-muted-foreground">
                  Leave this blank for local-docker drafts. Buyer installs use the marketplace hub;
                  deployment can populate the upstream runtime URL later.
                </p>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                rows={4}
                value={form.description}
                onChange={event =>
                  setForm(current => ({ ...current, description: event.target.value }))
                }
              />
            </div>

            <div className="grid gap-6 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="required-scopes">Required Scopes</Label>
                <Input
                  id="required-scopes"
                  placeholder="agent:invoke,files:read"
                  value={form.requiredScopes}
                  onChange={event =>
                    setForm(current => ({ ...current, requiredScopes: event.target.value }))
                  }
                />
                <p className="text-xs text-muted-foreground">
                  Comma-separated. Parsed as {parsedScopes.length} scope
                  {parsedScopes.length === 1 ? '' : 's'}.
                </p>
              </div>
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="pricing-type">Pricing Type</Label>
                  <select
                    id="pricing-type"
                    value={form.pricingType}
                    onChange={event =>
                      setForm(current => ({ ...current, pricingType: event.target.value }))
                    }
                    className="px-3 py-2 rounded-md border border-input bg-background text-sm w-full"
                  >
                    <option value="x402">x402</option>
                    <option value="subscription">subscription</option>
                    <option value="free">free</option>
                  </select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="pricing-amount">Price (USDC)</Label>
                  <Input
                    id="pricing-amount"
                    type="number"
                    min="0"
                    step="0.01"
                    value={form.pricingAmount}
                    onChange={event =>
                      setForm(current => ({ ...current, pricingAmount: event.target.value }))
                    }
                  />
                </div>
              </div>
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              <label className="flex items-start gap-3 rounded-lg border border-border p-4">
                <input
                  type="checkbox"
                  checked={form.supportsLocal}
                  onChange={event =>
                    setForm(current => ({ ...current, supportsLocal: event.target.checked }))
                  }
                  className="mt-1"
                />
                <div>
                  <p className="font-medium">Supports local install</p>
                  <p className="text-sm text-muted-foreground">
                    Buyers can connect this server to local MCP clients.
                  </p>
                </div>
              </label>

              <label className="flex items-start gap-3 rounded-lg border border-border p-4">
                <input
                  type="checkbox"
                  checked={form.supportsCloud}
                  onChange={event =>
                    setForm(current => ({ ...current, supportsCloud: event.target.checked }))
                  }
                  className="mt-1"
                />
                <div>
                  <p className="font-medium">Supports cloud deployment</p>
                  <p className="text-sm text-muted-foreground">
                    Deployment workflows can target hosted runtimes for this image.
                  </p>
                </div>
              </label>
            </div>

            <div className="rounded-lg border border-border bg-muted/40 p-4 text-sm text-muted-foreground">
              The backend creates this server immediately in draft state. Deployment, payment
              configuration, and publish still happen on the later seller pages.
            </div>

            <Button type="submit" disabled={isSubmitting} className="w-full">
              {isSubmitting ? 'Creating draft...' : 'Create Draft Server'}
            </Button>
          </form>
        </Card>
      </div>
    </AppShell>
  )
}
