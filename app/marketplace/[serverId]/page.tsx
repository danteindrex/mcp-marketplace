'use client'

import { useState } from 'react'
import Link from 'next/link'
import { use } from 'react'
import { ArrowLeft, Star, Download, Shield, Clock, Users, ExternalLink, Package } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { mockServers, getServerById } from '@/lib/mock-data'

interface PageProps {
  params: Promise<{ serverId: string }>
}

export default function ServerDetailPage({ params }: PageProps) {
  const { serverId } = use(params)
  const [isInstalling, setIsInstalling] = useState(false)

  // Find server by slug
  const server = mockServers.find(s => s.slug === serverId)

  if (!server) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen p-4">
        <h1 className="text-2xl font-bold mb-4">Server not found</h1>
        <Button asChild>
          <Link href="/marketplace">Back to Marketplace</Link>
        </Button>
      </div>
    )
  }

  const handleInstall = async () => {
    setIsInstalling(true)
    // Simulate installation
    setTimeout(() => {
      window.location.href = `/install/${server.slug}`
    }, 500)
  }

  return (
    <div className="flex flex-col min-h-screen bg-background">
      {/* Header */}
      <div className="border-b border-border bg-background/50 backdrop-blur sticky top-0 z-40">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-6 flex items-center justify-between">
          <Link
            href="/marketplace"
            className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            Back to Marketplace
          </Link>
          <div className="flex items-center gap-2">
            {server.verified && <Badge className="bg-green-500/20 text-green-700 dark:text-green-400 hover:bg-green-500/30">Verified</Badge>}
            {server.featured && <Badge className="bg-amber-500/20 text-amber-700 dark:text-amber-400 hover:bg-amber-500/30">Featured</Badge>}
          </div>
        </div>
      </div>

      {/* Main Content */}
      <main className="flex-1 max-w-4xl mx-auto w-full px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left Column - Main Content */}
          <div className="lg:col-span-2 space-y-8">
            {/* Server Header */}
            <div className="space-y-4">
              <h1 className="text-4xl font-bold">{server.name}</h1>
              <p className="text-xl text-muted-foreground">{server.description}</p>

              <div className="flex flex-wrap gap-3">
                <Badge variant="secondary">{server.category}</Badge>
                <Badge variant="outline">v{server.version}</Badge>
                <Badge variant="outline">License: {server.license}</Badge>
              </div>

              <div className="flex flex-wrap gap-6 text-sm">
                <div className="flex items-center gap-2">
                  <Star className="w-4 h-4 fill-amber-400 text-amber-400" />
                  <span>{server.rating} ({Math.floor(server.installCount * 0.3)} reviews)</span>
                </div>
                <div className="flex items-center gap-2">
                  <Download className="w-4 h-4 text-muted-foreground" />
                  <span>{server.installCount.toLocaleString()} installs</span>
                </div>
                <div className="flex items-center gap-2">
                  <Clock className="w-4 h-4 text-muted-foreground" />
                  <span>Updated {server.lastUpdated.toLocaleDateString()}</span>
                </div>
              </div>
            </div>

            {/* Overview Section */}
            <Card className="p-6 space-y-4">
              <h2 className="text-xl font-semibold">Overview</h2>
              <p className="text-muted-foreground">
                {server.description} This server provides {server.toolCount} tools and requires specific OAuth2 scopes for
                integration. It supports multiple client types including VS Code, Cursor, and Claude.
              </p>

              <div className="grid grid-cols-2 gap-4 pt-4 border-t border-border">
                <div>
                  <p className="text-sm text-muted-foreground mb-1">Tools Provided</p>
                  <p className="text-2xl font-bold">{server.toolCount}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground mb-1">Latest Version</p>
                  <p className="text-2xl font-bold">v{server.version}</p>
                </div>
              </div>
            </Card>

            {/* Tools Section */}
            <Card className="p-6 space-y-4">
              <h2 className="text-xl font-semibold">Available Tools</h2>
              <div className="space-y-3">
                {['query', 'execute', 'analyze', 'transform', 'validate'].slice(0, server.toolCount).map((tool, i) => (
                  <div key={i} className="flex items-start gap-3 p-3 bg-muted/50 rounded-lg">
                    <Package className="w-4 h-4 mt-1 text-primary flex-shrink-0" />
                    <div>
                      <p className="font-medium capitalize">{tool} Tool</p>
                      <p className="text-sm text-muted-foreground">
                        {tool === 'query' && 'Execute queries against the server'}
                        {tool === 'execute' && 'Execute custom commands'}
                        {tool === 'analyze' && 'Analyze data and results'}
                        {tool === 'transform' && 'Transform data formats'}
                        {tool === 'validate' && 'Validate inputs and outputs'}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            </Card>

            {/* Scopes Section */}
            <Card className="p-6 space-y-4">
              <h2 className="text-xl font-semibold">Required Scopes</h2>
              <p className="text-sm text-muted-foreground">
                This server requires the following OAuth2 scopes to function properly:
              </p>
              <div className="flex flex-wrap gap-2">
                {server.requiredScopes.map(scope => (
                  <Badge key={scope} variant="outline">
                    {scope}
                  </Badge>
                ))}
              </div>
            </Card>

            {/* Compatibility Section */}
            <Card className="p-6 space-y-4">
              <h2 className="text-xl font-semibold">Client Compatibility</h2>
              <div className="grid grid-cols-2 gap-4">
                {[
                  { name: 'VS Code', supported: true },
                  { name: 'Cursor', supported: true },
                  { name: 'Claude', supported: true },
                  { name: 'Codex', supported: false },
                ].map(client => (
                  <div key={client.name} className="flex items-center gap-3 p-3 bg-muted/50 rounded-lg">
                    {client.supported ? (
                      <div className="w-4 h-4 rounded-full bg-green-500/20 flex items-center justify-center">
                        <div className="w-2 h-2 rounded-full bg-green-500" />
                      </div>
                    ) : (
                      <div className="w-4 h-4 rounded-full bg-muted border border-border" />
                    )}
                    <span className={client.supported ? 'text-foreground' : 'text-muted-foreground'}>
                      {client.name}
                    </span>
                  </div>
                ))}
              </div>
            </Card>

            {/* Author Section */}
            <Card className="p-6">
              <h2 className="text-xl font-semibold mb-4">About the Publisher</h2>
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold text-lg">
                  {server.author.charAt(0)}
                </div>
                <div className="flex-1">
                  <p className="font-semibold">{server.author}</p>
                  {server.homepage && (
                    <a
                      href={server.homepage}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-primary hover:underline flex items-center gap-1"
                    >
                      Visit Homepage <ExternalLink className="w-3 h-3" />
                    </a>
                  )}
                </div>
              </div>
            </Card>
          </div>

          {/* Right Column - Install Card */}
          <div className="lg:col-span-1">
            <Card className="p-6 sticky top-24 space-y-6">
              {/* Pricing */}
              <div>
                <p className="text-sm text-muted-foreground mb-2">Price</p>
                {server.pricingType === 'free' ? (
                  <p className="text-3xl font-bold">Free</p>
                ) : (
                  <div>
                    <p className="text-3xl font-bold">
                      {server.currency} {server.pricingAmount}
                    </p>
                    <p className="text-sm text-muted-foreground capitalize mt-1">{server.pricingType}</p>
                  </div>
                )}
              </div>

              {/* Install Button */}
              <Button
                onClick={handleInstall}
                disabled={isInstalling}
                size="lg"
                className="w-full"
              >
                {isInstalling ? 'Installing...' : 'Install Server'}
              </Button>

              {/* Features */}
              <div className="space-y-3 border-t border-border pt-6">
                <div className="flex items-start gap-3">
                  <Shield className="w-5 h-5 text-primary mt-0.5 flex-shrink-0" />
                  <div>
                    <p className="font-medium text-sm">Verified & Secure</p>
                    <p className="text-xs text-muted-foreground">Security audited</p>
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <Users className="w-5 h-5 text-primary mt-0.5 flex-shrink-0" />
                  <div>
                    <p className="font-medium text-sm">Community Trusted</p>
                    <p className="text-xs text-muted-foreground">{server.installCount.toLocaleString()} active users</p>
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <Clock className="w-5 h-5 text-primary mt-0.5 flex-shrink-0" />
                  <div>
                    <p className="font-medium text-sm">Actively Maintained</p>
                    <p className="text-xs text-muted-foreground">Updated regularly</p>
                  </div>
                </div>
              </div>

              {/* Stats */}
              <div className="space-y-2 border-t border-border pt-6">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Rating</span>
                  <span className="font-medium flex items-center gap-1">
                    <Star className="w-4 h-4 fill-amber-400 text-amber-400" />
                    {server.rating}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Installs</span>
                  <span className="font-medium">{server.installCount.toLocaleString()}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Tools</span>
                  <span className="font-medium">{server.toolCount}</span>
                </div>
              </div>
            </Card>
          </div>
        </div>
      </main>
    </div>
  )
}
