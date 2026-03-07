'use client'

import { useState, useEffect, use } from 'react'
import Link from 'next/link'
import { ArrowLeft, Star, Download, Shield, Clock, Users, Package } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { fetchServerBySlug, type Server } from '@/lib/api-client'
import { Breadcrumb, Button as RetroButton } from '@/components/retroui'
import { LightModeOnly, DarkModeOnly } from '@/components/theme-aware'

interface PageProps {
  params: Promise<{ serverId: string }>
}

export default function ServerDetailPage({ params }: PageProps) {
  const { serverId } = use(params)
  const [isInstalling, setIsInstalling] = useState(false)
  const [server, setServer] = useState<Server | null>(null)

  useEffect(() => {
    fetchServerBySlug(serverId).then(setServer)
  }, [serverId])

  if (!server) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen p-4 grid-pattern bg-background">
        <h1 className="text-2xl font-black uppercase mb-4">Server not found</h1>
        <DarkModeOnly>
          <Button asChild className="button-coral-solid">
            <Link href="/marketplace">Back to Marketplace</Link>
          </Button>
        </DarkModeOnly>
        <LightModeOnly>
          <RetroButton asChild className="button-coral-solid">
            <Link href="/marketplace">Back to Marketplace</Link>
          </RetroButton>
        </LightModeOnly>
      </div>
    )
  }

  const handleInstall = async () => {
    setIsInstalling(true)
    window.location.href = `/install/${server.slug}`
  }

  return (
    <div className="flex flex-col min-h-screen grid-pattern bg-background">
      <div className="border-b-2 border-foreground bg-background/90 backdrop-blur sticky top-0 z-40">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-6 flex items-center justify-between">
          <div className="space-y-3">
            <Breadcrumb>
              <Breadcrumb.List>
                <Breadcrumb.Item>
                  <Breadcrumb.Link asChild>
                    <Link href="/">Home</Link>
                  </Breadcrumb.Link>
                </Breadcrumb.Item>
                <Breadcrumb.Separator />
                <Breadcrumb.Item>
                  <Breadcrumb.Link asChild>
                    <Link href="/marketplace">Marketplace</Link>
                  </Breadcrumb.Link>
                </Breadcrumb.Item>
                <Breadcrumb.Separator />
                <Breadcrumb.Item>
                  <Breadcrumb.Page>{server.name}</Breadcrumb.Page>
                </Breadcrumb.Item>
              </Breadcrumb.List>
            </Breadcrumb>
            <Link href="/marketplace" className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors font-semibold">
              <ArrowLeft className="w-4 h-4" />
              Back to Marketplace
            </Link>
          </div>
          <div className="flex items-center gap-2">
            {server.verified && <Badge className="bg-green-500/20 text-green-700 dark:text-green-400 hover:bg-green-500/30">Verified</Badge>}
            {server.featured && <Badge className="bg-amber-500/20 text-amber-700 dark:text-amber-400 hover:bg-amber-500/30">Featured</Badge>}
          </div>
        </div>
      </div>

      <main className="flex-1 max-w-4xl mx-auto w-full px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2 space-y-8">
            <div className="space-y-4">
              <h1 className="text-4xl font-black uppercase">{server.name}</h1>
              <p className="text-xl text-muted-foreground font-medium">{server.description}</p>
              <div className="flex flex-wrap gap-3">
                <Badge variant="secondary">{server.category}</Badge>
                <Badge variant="outline">v{server.version}</Badge>
              </div>
              <div className="flex flex-wrap gap-6 text-sm font-semibold">
                <div className="flex items-center gap-2"><Star className="w-4 h-4 fill-amber-400 text-amber-400" /><span>{server.rating}</span></div>
                <div className="flex items-center gap-2"><Download className="w-4 h-4 text-muted-foreground" /><span>{server.installCount.toLocaleString()} installs</span></div>
                <div className="flex items-center gap-2"><Clock className="w-4 h-4 text-muted-foreground" /><span>Updated {new Date(server.updatedAt).toLocaleDateString()}</span></div>
              </div>
            </div>

            <Card className="p-6 space-y-4 border-2 border-foreground shadow-[4px_4px_0px_hsl(var(--shadow-color))]">
              <h2 className="text-xl font-black uppercase">Required Scopes</h2>
              <div className="flex flex-wrap gap-2">
                {server.requiredScopes.map(scope => <Badge key={scope} variant="outline">{scope}</Badge>)}
              </div>
            </Card>

            <Card className="p-6 space-y-4 border-2 border-foreground shadow-[4px_4px_0px_hsl(var(--shadow-color))]">
              <h2 className="text-xl font-black uppercase">Available Tools</h2>
              <div className="space-y-3">
                {server.requiredScopes.map((scope, i) => (
                  <div key={scope + i} className="flex items-start gap-3 p-3 bg-muted/50 rounded-lg border border-foreground/40">
                    <Package className="w-4 h-4 mt-1 text-primary flex-shrink-0" />
                    <div>
                      <p className="font-semibold uppercase">Tool {i + 1}</p>
                      <p className="text-sm text-muted-foreground font-medium">Scope-bound capability: {scope}</p>
                    </div>
                  </div>
                ))}
              </div>
            </Card>
          </div>

          <div className="lg:col-span-1">
            <Card className="p-6 sticky top-24 space-y-6 border-2 border-foreground shadow-[4px_4px_0px_hsl(var(--shadow-color))]">
              <div>
                <p className="text-sm text-muted-foreground mb-2 font-semibold">Price</p>
                {server.pricingType === 'free' ? <p className="text-3xl font-black uppercase">Free</p> : <p className="text-3xl font-black">${server.pricingAmount}</p>}
              </div>
              <DarkModeOnly>
                <Button onClick={handleInstall} disabled={isInstalling} size="lg" className="w-full button-orange-solid">
                  {isInstalling ? 'Installing...' : 'Install Server'}
                </Button>
              </DarkModeOnly>
              <LightModeOnly>
                <RetroButton onClick={handleInstall} disabled={isInstalling} size="lg" className="w-full button-orange-solid">
                  {isInstalling ? 'Installing...' : 'Install Server'}
                </RetroButton>
              </LightModeOnly>
              <div className="space-y-3 border-t-2 border-foreground/40 pt-6">
                <div className="flex items-start gap-3"><Shield className="w-5 h-5 text-primary mt-0.5 flex-shrink-0" /><div><p className="font-semibold text-sm">Verified & Secure</p></div></div>
                <div className="flex items-start gap-3"><Users className="w-5 h-5 text-primary mt-0.5 flex-shrink-0" /><div><p className="font-semibold text-sm">Community Trusted</p></div></div>
              </div>
            </Card>
          </div>
        </div>
      </main>
    </div>
  )
}
