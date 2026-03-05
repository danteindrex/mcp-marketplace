'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { ArrowRight, Zap, Shield, TrendingUp, Users, Grid, Sparkles, MessageCircle, Lock, Zap as ZapIcon, Cpu } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { fetchFeaturedServers, type Server } from '@/lib/api-client'
import { TypeWriter, ScrollText, DynamicText, MouseEffectCard, SpotlightCard, BeamsBackground } from '@/components/kokonut'
import { Accordion } from '@/components/animate-ui/accordion'
import { LightModeOnly, DarkModeOnly } from '@/components/theme-aware'
import { Button as RetroButton, Card as RetroCard, Text as RetroText, Badge as RetroBadge } from '@/components/retroui'

const features = [
  {
    icon: Zap,
    title: 'Easy Installation',
    description: 'One-click setup with automatic scope negotiation and permission management',
  },
  {
    icon: Shield,
    title: 'Secure Integration',
    description: 'Enterprise-grade security with OAuth2, token rotation, and audit logging',
  },
  {
    icon: TrendingUp,
    title: 'Usage Analytics',
    description: 'Real-time metrics and performance monitoring for all your connections',
  },
  {
    icon: Users,
    title: 'Community Driven',
    description: 'Discover and share servers built by the global developer community',
  },
]

export default function HomePage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [featuredServers, setFeaturedServers] = useState<Server[]>([])

  const categories = [
    { name: 'Data', count: featuredServers.filter(s => s.category === 'data').length, color: 'bg-blue-500/10 text-blue-700 dark:text-blue-400' },
    { name: 'Automation', count: featuredServers.filter(s => s.category === 'automation').length, color: 'bg-purple-500/10 text-purple-700 dark:text-purple-400' },
    { name: 'AI', count: featuredServers.filter(s => s.category === 'ai').length, color: 'bg-amber-500/10 text-amber-700 dark:text-amber-400' },
    { name: 'Integration', count: featuredServers.filter(s => s.category === 'integration').length, color: 'bg-green-500/10 text-green-700 dark:text-green-400' },
  ]

  useEffect(() => {
    fetchFeaturedServers().then(setFeaturedServers).catch(() => setFeaturedServers([]))
  }, [])

  return (
    <BeamsBackground className="flex flex-col min-h-screen" intensity={0.4}>
      <div className="flex flex-col min-h-screen bg-gradient-to-b from-background to-background">
        {/* Navigation */}
        <nav className="border-b border-border sticky top-0 z-50 bg-background/95 backdrop-blur">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Sparkles className="w-6 h-6 text-primary" />
            <span className="text-xl font-bold">MCP Marketplace</span>
          </div>

          <div className="hidden md:flex items-center gap-8">
            <Link href="/marketplace" className="text-sm hover:text-primary transition-colors">
              Browse
            </Link>
            <Link href="#features" className="text-sm hover:text-primary transition-colors">
              Features
            </Link>
            <Link href="#categories" className="text-sm hover:text-primary transition-colors">
              Categories
            </Link>
          </div>

          <div className="flex items-center gap-3">
            <Button variant="outline" size="sm" asChild>
              <Link href="/buyer/dashboard">Dashboard</Link>
            </Button>
            <Button size="sm" asChild>
              <Link href="/marketplace">Get Started</Link>
            </Button>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="relative px-4 sm:px-6 lg:px-8 py-20 sm:py-32 max-w-7xl mx-auto w-full">
        <div className="text-center space-y-8">
          <div className="space-y-4">
            <h1 className="text-5xl sm:text-6xl font-bold tracking-tight text-balance">
              <TypeWriter 
                text="Discover & Deploy MCP Servers"
                speed={50}
                loop={true}
                loopDelay={3000}
                className="bg-gradient-to-r from-primary to-primary/70 bg-clip-text text-transparent block"
              />
            </h1>
            <div className="text-xl text-muted-foreground max-w-2xl mx-auto text-balance h-24 overflow-hidden">
              <ScrollText speed="normal" className="h-full">
                <div className="py-6">
                  The comprehensive marketplace for Model Context Protocol servers. Find, install, and manage servers securely with enterprise-grade features.
                </div>
              </ScrollText>
            </div>
          </div>

          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
            <Button size="lg" asChild className="w-full sm:w-auto">
              <Link href="/marketplace">
                Browse Marketplace <ArrowRight className="ml-2 w-4 h-4" />
              </Link>
            </Button>
            <Button size="lg" variant="outline" asChild className="w-full sm:w-auto">
              <Link href="/merchant/onboarding">Publish Server</Link>
            </Button>
          </div>

          <div className="flex flex-col sm:flex-row gap-2 justify-center items-center text-sm text-muted-foreground">
            <span>🔒 Secure • 📊 Analytics • 🚀 Easy Setup</span>
          </div>
        </div>

        {/* Hero Image Placeholder */}
        <div className="mt-16 rounded-xl border border-border bg-muted/30 backdrop-blur overflow-hidden aspect-video max-h-96 flex items-center justify-center">
          <div className="text-center">
            <Grid className="w-12 h-12 text-muted-foreground/30 mx-auto mb-4" />
            <p className="text-muted-foreground text-sm">Marketplace visualization</p>
          </div>
        </div>
      </section>

      {/* Featured Servers */}
      <section className="px-4 sm:px-6 lg:px-8 py-16 max-w-7xl mx-auto w-full border-t border-border">
        <div className="space-y-8">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-3xl font-bold">Featured Servers</h2>
              <p className="text-muted-foreground mt-2">Popular and verified MCP servers</p>
            </div>
            <Button variant="ghost" asChild>
              <Link href="/marketplace">
                View All <ArrowRight className="ml-2 w-4 h-4" />
              </Link>
            </Button>
          </div>

          <DarkModeOnly>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {featuredServers.slice(0, 3).map(server => (
                <Link key={server.id} href={`/marketplace/${server.slug}`}>
                  <SpotlightCard 
                    className="p-6 h-full bg-card border border-border rounded-lg"
                    spotlightColor="rgba(59, 130, 246, 0.3)"
                  >
                    <div className="flex items-start justify-between mb-4">
                      <div>
                        <h3 className="font-semibold group-hover:text-primary transition-colors">
                          <DynamicText 
                            texts={[server.name, `${server.name} • Popular`]}
                            interval={4000}
                            className="text-primary"
                          />
                        </h3>
                        <p className="text-sm text-muted-foreground">by {server.author}</p>
                      </div>
                      {server.verified && <span className="text-xs bg-green-500/20 text-green-700 dark:text-green-400 px-2 py-1 rounded">Verified</span>}
                    </div>
                    <p className="text-sm text-muted-foreground mb-4 line-clamp-2">{server.description}</p>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">⭐ {server.rating}</span>
                      <span className="text-muted-foreground">{server.installCount.toLocaleString()} installs</span>
                    </div>
                  </SpotlightCard>
                </Link>
              ))}
            </div>
          </DarkModeOnly>

          <LightModeOnly>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {featuredServers.slice(0, 3).map(server => (
                <Link key={server.id} href={`/marketplace/${server.slug}`}>
                  <RetroCard className="p-6 h-full">
                    <div className="flex items-start justify-between mb-4">
                      <div>
                        <h3 className="font-black text-lg">
                          {server.name}
                        </h3>
                        <p className="text-sm font-semibold text-gray-700">by {server.author}</p>
                      </div>
                      {server.verified && <RetroBadge variant="default">VERIFIED</RetroBadge>}
                    </div>
                    <p className="text-sm font-semibold text-gray-800 mb-4 line-clamp-2">{server.description}</p>
                    <div className="flex items-center justify-between text-sm font-bold">
                      <span>⭐ {server.rating}</span>
                      <span>{server.installCount.toLocaleString()}</span>
                    </div>
                  </RetroCard>
                </Link>
              ))}
            </div>
          </LightModeOnly>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="px-4 sm:px-6 lg:px-8 py-16 max-w-7xl mx-auto w-full border-t border-border">
        <DarkModeOnly>
          <div className="space-y-12">
            <div className="text-center">
              <h2 className="text-3xl font-bold mb-4">Why Choose MCP Marketplace?</h2>
              <p className="text-muted-foreground max-w-2xl mx-auto">
                Built for developers, by developers. Enterprise-grade features for managing MCP servers at scale.
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              {features.map((feature, index) => {
                const Icon = feature.icon
                return (
                  <MouseEffectCard key={index} className="p-6 bg-card border border-border rounded-lg" intensity={0.8}>
                    <Icon className="w-8 h-8 text-primary mb-4" />
                    <h3 className="font-semibold mb-2">{feature.title}</h3>
                    <p className="text-sm text-muted-foreground">{feature.description}</p>
                  </MouseEffectCard>
                )
              })}
            </div>
          </div>
        </DarkModeOnly>

        <LightModeOnly>
          <div className="space-y-12">
            <div className="text-center">
              <h2 className="text-5xl font-black mb-4">Why Choose MCP Marketplace?</h2>
              <p className="text-base font-bold text-gray-800 max-w-2xl mx-auto">
                Built for developers, by developers. Enterprise-grade features for managing MCP servers at scale.
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              {features.map((feature, index) => {
                const Icon = feature.icon
                return (
                  <RetroCard key={index} className="p-6">
                    <Icon className="w-8 h-8 text-black mb-4" />
                    <h3 className="font-black mb-2 text-lg">{feature.title}</h3>
                    <p className="text-sm font-bold text-gray-700">{feature.description}</p>
                  </RetroCard>
                )
              })}
            </div>
          </div>
        </LightModeOnly>
      </section>

      {/* Categories Section */}
      <section id="categories" className="px-4 sm:px-6 lg:px-8 py-16 max-w-7xl mx-auto w-full border-t border-border">
        <div className="space-y-8">
          <div>
            <h2 className="text-3xl font-bold mb-4">Browse by Category</h2>
            <p className="text-muted-foreground">Find servers in your area of interest</p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            {categories.map(category => (
              <Link key={category.name} href={`/marketplace?category=${category.name.toLowerCase()}`}>
                <Card className={`p-6 cursor-pointer hover:border-primary transition-colors ${category.color} border-transparent`}>
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="font-semibold">{category.name}</h3>
                      <p className="text-sm opacity-75">{category.count} servers</p>
                    </div>
                    <ArrowRight className="w-4 h-4 opacity-50" />
                  </div>
                </Card>
              </Link>
            ))}
          </div>
        </div>
      </section>

      {/* FAQ Section */}
      <section className="px-4 sm:px-6 lg:px-8 py-20 max-w-4xl mx-auto w-full border-t border-border">
        <div className="text-center mb-12">
          <h2 className="text-4xl font-bold mb-4">Frequently Asked Questions</h2>
          <p className="text-muted-foreground max-w-2xl mx-auto">
            Everything you need to know about MCP Servers and the marketplace
          </p>
        </div>
        
        <DarkModeOnly>
          <Accordion 
            items={[
              {
                value: 'what-is-mcp',
                trigger: 'What is Model Context Protocol (MCP)?',
                content: 'MCP is a standardized protocol that enables seamless communication between AI models and external data sources, tools, and services. It provides a secure way to extend AI capabilities.'
              },
              {
                value: 'how-install',
                trigger: 'How do I install a server?',
                content: 'Installing a server takes just a few clicks. Navigate to the marketplace, find a server you want, click "Install", review the required permissions, and authorize access. The setup wizard handles the rest.'
              },
              {
                value: 'security',
                trigger: 'Is it safe to use MCP Servers?',
                content: 'Yes! All servers are verified and scanned for security issues. We implement enterprise-grade security with OAuth2, token rotation, and comprehensive audit logging.'
              },
              {
                value: 'pricing',
                trigger: 'What are the pricing options?',
                content: 'Many servers are free with optional premium tiers. Some use per-call pricing, subscriptions, or flat monthly fees. Check each server for specific pricing details.'
              },
              {
                value: 'support',
                trigger: 'How do I get support?',
                content: 'Each server has a dedicated support channel. You can also reach out to our community forums, check documentation, or contact our support team directly.'
              }
            ]}
          />
        </DarkModeOnly>

        <LightModeOnly>
          <div className="space-y-3">
            {[
              {
                value: 'what-is-mcp',
                trigger: 'What is Model Context Protocol (MCP)?',
                content: 'MCP is a standardized protocol that enables seamless communication between AI models and external data sources, tools, and services. It provides a secure way to extend AI capabilities.'
              },
              {
                value: 'how-install',
                trigger: 'How do I install a server?',
                content: 'Installing a server takes just a few clicks. Navigate to the marketplace, find a server you want, click "Install", review the required permissions, and authorize access. The setup wizard handles the rest.'
              },
              {
                value: 'security',
                trigger: 'Is it safe to use MCP Servers?',
                content: 'Yes! All servers are verified and scanned for security issues. We implement enterprise-grade security with OAuth2, token rotation, and comprehensive audit logging.'
              },
              {
                value: 'pricing',
                trigger: 'What are the pricing options?',
                content: 'Many servers are free with optional premium tiers. Some use per-call pricing, subscriptions, or flat monthly fees. Check each server for specific pricing details.'
              },
              {
                value: 'support',
                trigger: 'How do I get support?',
                content: 'Each server has a dedicated support channel. You can also reach out to our community forums, check documentation, or contact our support team directly.'
              }
            ].map((item) => (
              <RetroCard key={item.value} className="p-4">
                <details className="cursor-pointer">
                  <summary className="font-bold text-lg hover:text-gray-700 select-none">
                    {item.trigger}
                  </summary>
                  <p className="mt-3 text-gray-700">{item.content}</p>
                </details>
              </RetroCard>
            ))}
          </div>
        </LightModeOnly>
      </section>

      {/* CTA Section */}
      <section className="px-4 sm:px-6 lg:px-8 py-16 max-w-7xl mx-auto w-full border-t border-border">
        <DarkModeOnly>
          <Card className="p-12 text-center bg-gradient-to-br from-primary/5 to-primary/10 border-primary/20">
            <h2 className="text-3xl font-bold mb-4">Ready to get started?</h2>
            <p className="text-muted-foreground mb-8 max-w-xl mx-auto">
              Explore thousands of verified MCP servers and integrate them into your workflow in minutes.
            </p>
            <Button size="lg" asChild>
              <Link href="/marketplace">Explore Marketplace</Link>
            </Button>
          </Card>
        </DarkModeOnly>

        <LightModeOnly>
          <RetroCard className="p-12 text-center">
            <h2 className="text-3xl font-bold mb-4">Ready to get started?</h2>
            <p className="text-gray-700 mb-8 max-w-xl mx-auto">
              Explore thousands of verified MCP servers and integrate them into your workflow in minutes.
            </p>
            <RetroButton size="lg" asChild>
              <Link href="/marketplace">Explore Marketplace</Link>
            </RetroButton>
          </RetroCard>
        </LightModeOnly>
      </section>

      {/* Footer */}
      <footer className="border-t border-border mt-auto">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 text-center text-sm text-muted-foreground">
          <p>&copy; 2025 MCP Marketplace. All rights reserved.</p>
        </div>
      </footer>
      </div>
    </BeamsBackground>
  )
}
