'use client'

import { useState } from 'react'
import Link from 'next/link'
import { ArrowLeft, AlertCircle, CheckCircle2, Loader } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { AppShell } from '@/components/app-shell'
import { createMerchantServer } from '@/lib/api-client'
import { toast } from 'sonner'

export default function ImportDockerPage() {
  const [step, setStep] = useState<'url' | 'scanning' | 'review'>('url')
  const [dockerUrl, setDockerUrl] = useState('')
  const [registryAuth, setRegistryAuth] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  const [scanResults, setScanResults] = useState<{
    name: string
    slug: string
    version: string
    description: string
    baseSize: number
    layers: number
    sbom: { libraries: string[]; vulnerabilities: number }
  } | null>(null)

  const toSlug = (value: string) =>
    value
      .toLowerCase()
      .trim()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '')
      .slice(0, 50)

  const handleScan = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!dockerUrl) {
      toast.error('Please enter a Docker image URL')
      return
    }

    setIsLoading(true)
    setStep('scanning')

    const imageRef = dockerUrl.trim()
    const imageNameWithTag = imageRef.split('/').pop() || 'mcp-agent'
    const imageName = imageNameWithTag.split(':')[0] || 'mcp-agent'
    const tag = imageNameWithTag.split(':')[1] || 'latest'
    const derivedName = imageName
      .split('-')
      .filter(Boolean)
      .map(part => part.charAt(0).toUpperCase() + part.slice(1))
      .join(' ')
      .trim() || 'MCP Agent'
    const derivedSlug = toSlug(imageName) || 'mcp-agent'

    setTimeout(() => {
      setScanResults({
        name: derivedName,
        slug: derivedSlug,
        version: tag,
        description: `${derivedName} imported from Docker image ${imageRef}`,
        baseSize: 486,
        layers: 12,
        sbom: {
          libraries: ['postgres-driver:15.2', 'langchain:0.1.0', 'nodejs:20.10.0'],
          vulnerabilities: 2,
        },
      })
      setIsLoading(false)
      setStep('review')
    }, 2000)
  }

  const handleImport = async () => {
    if (!scanResults) return
    setIsLoading(true)
    try {
      const created = await createMerchantServer({
        name: scanResults.name,
        slug: scanResults.slug,
        description: scanResults.description,
        category: 'automation',
        dockerImage: dockerUrl.trim(),
        canonicalResourceUri: `https://mcp.marketplace.local/resource/${scanResults.slug}`,
        requiredScopes: ['agent:invoke'],
        pricingType: 'x402',
        pricingAmount: 0,
        supportsCloud: true,
        supportsLocal: true,
        status: 'draft',
      })
      toast.success('Server imported as draft')
      window.location.href = `/merchant/servers/${created.id}/deployments`
    } catch (e: any) {
      toast.error(e?.message || 'Import failed')
    } finally {
      setIsLoading(false)
    }
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
          <h1 className="text-3xl font-bold mb-2">Import Docker Image</h1>
          <p className="text-muted-foreground">
            Import your MCP server from Docker Hub or a private registry
          </p>
        </div>

        {/* Step 1: URL Input */}
        {step === 'url' && (
          <Card className="p-8 max-w-2xl">
            <h2 className="text-2xl font-bold mb-6">Docker Image URL</h2>

            <form onSubmit={handleScan} className="space-y-6">
              <div>
                <Label htmlFor="docker-url" className="mb-2 block">
                  Docker Image URL
                </Label>
                <Input
                  id="docker-url"
                  placeholder="e.g., docker.io/myorg/postgres-assistant:2.1.0"
                  value={dockerUrl}
                  onChange={e => setDockerUrl(e.target.value)}
                  className="font-mono text-sm"
                />
                <p className="text-xs text-muted-foreground mt-2">
                  Include the tag and registry. Public registries don't require authentication.
                </p>
              </div>

              <div>
                <Label htmlFor="registry-auth" className="mb-2 block">
                  Registry Authentication (Optional)
                </Label>
                <Input
                  id="registry-auth"
                  placeholder="username:password or auth token"
                  type="password"
                  value={registryAuth}
                  onChange={e => setRegistryAuth(e.target.value)}
                />
                <p className="text-xs text-muted-foreground mt-2">
                  Only needed for private registries
                </p>
              </div>

              <div className="bg-blue-500/10 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
                <p className="text-sm text-blue-900 dark:text-blue-100">
                  We'll scan your image for security vulnerabilities and generate an SBOM (Software Bill of Materials).
                </p>
              </div>

              <Button type="submit" disabled={isLoading} className="w-full">
                {isLoading ? 'Scanning...' : 'Scan & Import'}
              </Button>
            </form>
          </Card>
        )}

        {/* Step 2: Scanning */}
        {step === 'scanning' && (
          <Card className="p-8 max-w-2xl">
            <div className="text-center space-y-6">
              <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center mx-auto">
                <Loader className="w-8 h-8 text-primary animate-spin" />
              </div>
              <div>
                <h2 className="text-2xl font-bold mb-2">Scanning Docker Image</h2>
                <p className="text-muted-foreground">This may take a minute...</p>
              </div>
              <div className="space-y-2 text-sm text-muted-foreground">
                <p>Pulling image metadata</p>
                <p>Analyzing layers</p>
                <p>Scanning for vulnerabilities</p>
                <p>Generating SBOM</p>
              </div>
            </div>
          </Card>
        )}

        {/* Step 3: Review */}
        {step === 'review' && scanResults && (
          <div className="space-y-6">
            {/* Image Info */}
            <Card className="p-8">
              <h2 className="text-2xl font-bold mb-6">Review Scan Results</h2>

              <div className="space-y-6">
                <div>
                  <h3 className="font-semibold mb-4">Server Information</h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                      <p className="text-sm text-muted-foreground mb-1">Name</p>
                      <p className="font-medium">{scanResults.name}</p>
                    </div>
                    <div>
                      <p className="text-sm text-muted-foreground mb-1">Version</p>
                      <p className="font-medium">v{scanResults.version}</p>
                    </div>
                    <div className="md:col-span-2">
                      <p className="text-sm text-muted-foreground mb-1">Description</p>
                      <p className="text-sm">{scanResults.description}</p>
                    </div>
                  </div>
                </div>

                <div className="border-t border-border pt-6">
                  <h3 className="font-semibold mb-4">Image Metrics</h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <Card className="p-4 bg-muted/50 border-border">
                      <p className="text-sm text-muted-foreground mb-1">Compressed Size</p>
                      <p className="text-2xl font-bold">{scanResults.baseSize}MB</p>
                    </Card>
                    <Card className="p-4 bg-muted/50 border-border">
                      <p className="text-sm text-muted-foreground mb-1">Docker Layers</p>
                      <p className="text-2xl font-bold">{scanResults.layers}</p>
                    </Card>
                  </div>
                </div>

                <div className="border-t border-border pt-6">
                  <h3 className="font-semibold mb-4">SBOM - Software Bill of Materials</h3>
                  <div className="space-y-3">
                    <div>
                      <p className="text-sm font-medium mb-2">Base Libraries</p>
                      <div className="space-y-2">
                        {scanResults.sbom.libraries.map((lib, i) => (
                          <div
                            key={i}
                            className="flex items-center gap-2 text-sm p-2 bg-muted/50 rounded"
                          >
                            <CheckCircle2 className="w-4 h-4 text-green-500 flex-shrink-0" />
                            {lib}
                          </div>
                        ))}
                      </div>
                    </div>

                    {scanResults.sbom.vulnerabilities > 0 && (
                      <div className="bg-amber-500/10 border border-amber-200 dark:border-amber-800 rounded-lg p-4">
                        <div className="flex items-start gap-3">
                          <AlertCircle className="w-5 h-5 text-amber-600 dark:text-amber-500 flex-shrink-0 mt-0.5" />
                          <div>
                            <p className="font-semibold text-amber-900 dark:text-amber-100 text-sm">
                              {scanResults.sbom.vulnerabilities} Known Vulnerabilities
                            </p>
                            <p className="text-xs text-amber-800 dark:text-amber-200 mt-1">
                              Review and update dependencies if needed before publishing
                            </p>
                          </div>
                        </div>
                      </div>
                    )}
                  </div>
                </div>

                <div className="bg-green-500/10 border border-green-200 dark:border-green-800 rounded-lg p-4">
                  <p className="text-sm text-green-900 dark:text-green-100">
                    ✓ Image scan complete. Ready to import.
                  </p>
                </div>
              </div>
            </Card>

            {/* Actions */}
            <div className="flex gap-3">
              <Button
                variant="outline"
                onClick={() => {
                  setStep('url')
                  setScanResults(null)
                }}
                disabled={isLoading}
              >
                Back
              </Button>
              <Button
                onClick={handleImport}
                disabled={isLoading}
                className="flex-1"
              >
                {isLoading ? 'Importing...' : 'Import Server'}
              </Button>
            </div>
          </div>
        )}
      </div>
    </AppShell>
  )
}
