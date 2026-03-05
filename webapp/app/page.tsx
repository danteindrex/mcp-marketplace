'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { ArrowRight, Zap, Shield, TrendingUp, Users, Grid, Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Sticker } from '@/components/ui/sticker'
import { BurstShape, LightningShape, Star5Shape } from '@/components/ui/shapes'
import { fetchFeaturedServers, type Server } from '@/lib/api-client'
import { TypeWriter, ScrollText, MouseEffectCard, SpotlightCard, BeamsBackground } from '@/components/kokonut'
import { Accordion } from '@/components/animate-ui/accordion'
import { LightModeOnly, DarkModeOnly } from '@/components/theme-aware'
import { Button as RetroButton, Card as RetroCard, Badge as RetroBadge } from '@/components/retroui'
import { BarChart } from '@/components/retroui/charts/BarChart'

const features = [
  { icon: Zap, title: 'Easy Installation', description: 'One-click setup with automatic scope negotiation and permission management' },
  { icon: Shield, title: 'Secure Integration', description: 'Enterprise-grade security with OAuth2, token rotation, and audit logging' },
  { icon: TrendingUp, title: 'Usage Analytics', description: 'Real-time metrics and performance monitoring for all your connections' },
  { icon: Users, title: 'Community Driven', description: 'Discover and share servers built by the global developer community' },
]

export default function HomePage() {
  const [featuredServers, setFeaturedServers] = useState<Server[]>([])
  const landingChart = [
    { name: 'Jan', orders: 12 },
    { name: 'Feb', orders: 32 },
    { name: 'Mar', orders: 19 },
    { name: 'Apr', orders: 35 },
    { name: 'May', orders: 40 },
    { name: 'Jun', orders: 25 },
  ]

  useEffect(() => {
    fetchFeaturedServers().then(setFeaturedServers).catch(() => setFeaturedServers([]))
  }, [])

  const categories = [
    { name: 'Data', count: featuredServers.filter(s => s.category === 'data').length },
    { name: 'Automation', count: featuredServers.filter(s => s.category === 'automation').length },
    { name: 'AI', count: featuredServers.filter(s => s.category === 'ai').length },
    { name: 'Integration', count: featuredServers.filter(s => s.category === 'integration').length },
  ]

  return (
    <BeamsBackground className="flex flex-col min-h-screen" intensity={0.4}>
      <div className="flex flex-col min-h-screen bg-gradient-to-b from-background to-background">
        <nav className="border-b border-border sticky top-0 z-50 bg-background/95 backdrop-blur">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
            <DarkModeOnly>
              <div className="flex items-center gap-2"><Sparkles className="w-6 h-6 text-primary" /><span className="text-xl font-bold">MCP Marketplace</span></div>
            </DarkModeOnly>
            <LightModeOnly>
              <RetroCard className="px-3 py-2 flex items-center gap-2"><Sparkles className="w-5 h-5" /><span className="text-lg font-black">MCP Marketplace</span></RetroCard>
            </LightModeOnly>

            <DarkModeOnly>
              <div className="hidden md:flex items-center gap-8">
                <Link href="/marketplace" className="text-sm hover:text-primary transition-colors">Browse</Link>
                <Link href="#features" className="text-sm hover:text-primary transition-colors">Features</Link>
                <Link href="#categories" className="text-sm hover:text-primary transition-colors">Categories</Link>
              </div>
            </DarkModeOnly>
            <LightModeOnly>
              <div className="hidden md:flex items-center gap-2">
                <RetroButton variant="outline" size="sm" asChild><Link href="/marketplace">Browse</Link></RetroButton>
                <RetroButton variant="outline" size="sm" asChild><Link href="#features">Features</Link></RetroButton>
                <RetroButton variant="outline" size="sm" asChild><Link href="#categories">Categories</Link></RetroButton>
              </div>
            </LightModeOnly>

            <DarkModeOnly>
              <div className="flex items-center gap-3">
                <Button variant="outline" size="sm" asChild><Link href="/login">Login</Link></Button>
                <Button size="sm" asChild><Link href="/marketplace">Get Started</Link></Button>
              </div>
            </DarkModeOnly>
            <LightModeOnly>
              <div className="flex items-center gap-2">
                <RetroButton variant="outline" size="sm" asChild><Link href="/login">Login</Link></RetroButton>
                <RetroButton size="sm" asChild><Link href="/marketplace">Get Started</Link></RetroButton>
              </div>
            </LightModeOnly>
          </div>
        </nav>

        <section className="relative px-4 sm:px-6 lg:px-8 py-20 sm:py-32 max-w-7xl mx-auto w-full overflow-hidden grid-pattern">
          <div className="pointer-events-none absolute -right-6 top-4 hidden lg:block">
            <BurstShape size={72} className="text-tertiary" />
          </div>
          <div className="pointer-events-none absolute left-8 bottom-8 hidden lg:block">
            <Star5Shape size={56} className="text-secondary" />
          </div>
          <div className="text-center space-y-8">
            <h1 className="text-5xl sm:text-6xl font-bold tracking-tight text-balance">
              <TypeWriter text="Discover & Deploy MCP Servers" speed={50} loop={false} className="bg-gradient-to-r from-primary to-primary/70 bg-clip-text text-transparent block" />
            </h1>
            <div className="text-xl text-muted-foreground max-w-2xl mx-auto text-balance h-24 overflow-hidden">
              <ScrollText speed="normal" className="h-full"><div className="py-6">The comprehensive marketplace for Model Context Protocol servers. Find, install, and manage servers securely with enterprise-grade features.</div></ScrollText>
            </div>

            <DarkModeOnly>
              <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
                <Button size="lg" asChild className="w-full sm:w-auto"><Link href="/marketplace">Browse Marketplace <ArrowRight className="ml-2 w-4 h-4" /></Link></Button>
                <Button size="lg" variant="outline" asChild className="w-full sm:w-auto"><Link href="/merchant/onboarding">Publish Server</Link></Button>
              </div>
            </DarkModeOnly>
            <LightModeOnly>
              <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
                <RetroButton size="lg" asChild className="w-full sm:w-auto"><Link href="/marketplace">Browse Marketplace</Link></RetroButton>
                <RetroButton variant="outline" size="lg" asChild className="w-full sm:w-auto !bg-[hsl(var(--tertiary))] !text-[hsl(var(--tertiary-foreground))]"><Link href="/merchant/onboarding">Publish Server</Link></RetroButton>
              </div>
            </LightModeOnly>
          </div>

          <DarkModeOnly>
            <div className="mt-16 rounded-xl border border-border bg-muted/30 backdrop-blur overflow-hidden aspect-video max-h-96 flex items-center justify-center">
              <ScrollText speed="slow"><div className="text-center"><Grid className="w-12 h-12 text-muted-foreground/30 mx-auto mb-4" /><p className="text-muted-foreground text-sm">Marketplace visualization</p></div></ScrollText>
            </div>
          </DarkModeOnly>
          <LightModeOnly>
            <RetroCard className="mt-16 overflow-hidden aspect-video max-h-96 flex items-center justify-center">
              <ScrollText speed="slow"><div className="text-center"><Grid className="w-12 h-12 text-gray-600 mx-auto mb-4" /><p className="text-sm font-semibold text-gray-700">Marketplace visualization</p></div></ScrollText>
            </RetroCard>
          </LightModeOnly>
        </section>

        <section className="px-4 sm:px-6 lg:px-8 py-16 max-w-7xl mx-auto w-full border-t border-border">
          <div className="space-y-8">
            <ScrollText speed="normal"><div className="flex items-center justify-between"><div><h2 className="text-3xl font-bold">Featured Servers</h2><p className="text-muted-foreground mt-2">Popular and verified MCP servers</p></div></div></ScrollText>
            <Card className="p-6">
              <ScrollText speed="normal"><h3 className="text-lg font-semibold mb-4">Marketplace Activity</h3></ScrollText>
              <BarChart
                data={landingChart}
                index="name"
                categories={['orders']}
              />
            </Card>
            <DarkModeOnly>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {featuredServers.slice(0, 3).map(server => (
                  <Link key={server.id} href={`/marketplace/${server.slug}`}>
                    <SpotlightCard className="p-6 h-full bg-card border border-border rounded-lg" spotlightColor="rgba(59, 130, 246, 0.3)">
                      <div className="flex items-start justify-between mb-4"><div><h3 className="font-semibold">{server.name}</h3><p className="text-sm text-muted-foreground">by {server.author}</p></div></div>
                      <p className="text-sm text-muted-foreground mb-4 line-clamp-2">{server.description}</p>
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
                      <ScrollText speed="normal">
                        <div>
                          <div className="flex items-start justify-between mb-4"><div><h3 className="font-black text-lg">{server.name}</h3><p className="text-sm font-semibold text-gray-700">by {server.author}</p></div>{server.verified && <RetroBadge variant="default">VERIFIED</RetroBadge>}</div>
                          <p className="text-sm font-semibold text-gray-800 mb-4 line-clamp-2">{server.description}</p>
                        </div>
                      </ScrollText>
                    </RetroCard>
                  </Link>
                ))}
              </div>
            </LightModeOnly>
          </div>
        </section>

        <section id="features" className="px-4 sm:px-6 lg:px-8 py-16 max-w-7xl mx-auto w-full border-t border-border">
          <ScrollText speed="normal"><div className="text-center mb-8"><h2 className="text-3xl font-bold">Why Choose MCP Marketplace?</h2></div></ScrollText>
          <DarkModeOnly>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">{features.map((feature, index) => { const Icon = feature.icon; return <MouseEffectCard key={index} className="p-6 bg-card border border-border rounded-lg" intensity={0.8}><Icon className="w-8 h-8 text-primary mb-4" /><h3 className="font-semibold mb-2">{feature.title}</h3><p className="text-sm text-muted-foreground">{feature.description}</p></MouseEffectCard> })}</div>
          </DarkModeOnly>
          <LightModeOnly>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">{features.map((feature, index) => { const Icon = feature.icon; return <RetroCard key={index} className="p-6"><ScrollText speed="normal"><div><Icon className="w-8 h-8 text-black mb-4" /><h3 className="font-black mb-2 text-lg">{feature.title}</h3><p className="text-sm font-bold text-gray-700">{feature.description}</p></div></ScrollText></RetroCard> })}</div>
          </LightModeOnly>
        </section>

        <section id="categories" className="px-4 sm:px-6 lg:px-8 py-16 max-w-7xl mx-auto w-full border-t border-border">
          <ScrollText speed="normal"><div className="mb-8"><h2 className="text-3xl font-bold">Browse by Category</h2></div></ScrollText>
          <LightModeOnly>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">{categories.map(category => <Link key={category.name} href={`/marketplace?category=${category.name.toLowerCase()}`}><RetroCard className="p-6"><ScrollText speed="normal"><div className="flex items-center justify-between"><div><h3 className="font-black">{category.name}</h3><p className="text-sm font-semibold">{category.count} servers</p></div><ArrowRight className="w-4 h-4" /></div></ScrollText></RetroCard></Link>)}</div>
          </LightModeOnly>
          <DarkModeOnly>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">{categories.map(category => <Link key={category.name} href={`/marketplace?category=${category.name.toLowerCase()}`}><Card className="p-6 cursor-pointer hover:border-primary transition-colors"><div className="flex items-center justify-between"><div><h3 className="font-semibold">{category.name}</h3><p className="text-sm opacity-75">{category.count} servers</p></div><ArrowRight className="w-4 h-4 opacity-50" /></div></Card></Link>)}</div>
          </DarkModeOnly>
        </section>

        <section className="px-4 sm:px-6 lg:px-8 py-20 max-w-4xl mx-auto w-full border-t border-border">
          <ScrollText speed="normal"><div className="text-center mb-12"><h2 className="text-4xl font-bold mb-4">Frequently Asked Questions</h2></div></ScrollText>
          <Accordion items={[{ value: 'what-is-mcp', trigger: 'What is Model Context Protocol (MCP)?', content: 'MCP is a standardized protocol that enables secure communication between AI models and tools.' }, { value: 'how-install', trigger: 'How do I install a server?', content: 'Pick a server, click install, approve scopes, and connect once to your personal hub.' }]} />
        </section>

        <section className="px-4 sm:px-6 lg:px-8 py-16 max-w-7xl mx-auto w-full border-t border-border">
          <Card className="relative overflow-hidden p-12 text-center bg-gradient-to-br from-accent to-background border-border">
            <div className="pointer-events-none absolute -right-8 top-4 hidden md:block">
              <LightningShape size={80} className="text-tertiary" />
            </div>
            <ScrollText speed="normal">
              <div>
                <h2 className="text-3xl font-black mb-4 uppercase">Ready to get started?</h2>
                <p className="text-muted-foreground mb-8 max-w-xl mx-auto">Explore thousands of verified MCP servers and integrate them into your workflow in minutes.</p>
              </div>
            </ScrollText>
            <Button size="lg" asChild className="bg-[hsl(var(--tertiary))] text-[hsl(var(--tertiary-foreground))] hover:opacity-90">
              <Link href="/marketplace">Explore Marketplace</Link>
            </Button>
          </Card>
        </section>

        <section id="contact" className="px-4 sm:px-6 lg:px-8 py-16 max-w-7xl mx-auto w-full border-t border-border">
          <div className="grid gap-6 lg:grid-cols-2">
            <Card className="relative overflow-hidden border-2 border-foreground bg-accent p-8">
              <Sticker variant="primary" className="mb-4">Contact Team</Sticker>
              <h3 className="text-3xl font-black uppercase mb-3">Need a custom deployment?</h3>
              <p className="text-sm text-foreground/80 max-w-prose">
                Tell us your tenant size, throughput goals, and compliance requirements. We can provision hosted or hybrid MCP infrastructure with role-based access and billing controls.
              </p>
              <div className="mt-6 space-y-2 text-sm font-semibold">
                <p>Sales: sales@mcp-marketplace.local</p>
                <p>Security: security@mcp-marketplace.local</p>
                <p>Support SLA: 24/7 enterprise response</p>
              </div>
              <BurstShape size={72} className="absolute -bottom-5 -right-5 text-tertiary" />
            </Card>

            <Card className="border-2 border-foreground p-8">
              <h3 className="text-2xl font-black uppercase mb-4">Send a message</h3>
              <form className="space-y-4">
                <div className="grid gap-4 sm:grid-cols-2">
                  <Input placeholder="Your name" />
                  <Input placeholder="Work email" type="email" />
                </div>
                <Input placeholder="Company / Tenant" />
                <Textarea placeholder="What do you need help building?" className="min-h-28" />
                <Button type="button" className="w-full bg-[hsl(var(--tertiary))] text-[hsl(var(--tertiary-foreground))] hover:opacity-90">Request Consultation</Button>
              </form>
            </Card>
          </div>
        </section>

        <footer className="border-t border-border mt-auto bg-foreground text-background">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
            <div className="grid gap-8 md:grid-cols-4">
              <div>
                <div className="flex items-center gap-2 mb-4">
                  <div className="h-8 w-8 border-2 border-background bg-[hsl(var(--tertiary))]" />
                  <span className="text-xl font-black uppercase">MCP Marketplace</span>
                </div>
                <p className="text-sm text-background/75">Marketplace for secure MCP discovery, deployment, billing, and tenancy control.</p>
              </div>
              <div>
                <h4 className="mb-3 text-sm font-black uppercase">Product</h4>
                <ul className="space-y-2 text-sm text-background/80">
                  <li><Link href="/marketplace" className="hover:text-background">Marketplace</Link></li>
                  <li><Link href="/buyer/dashboard" className="hover:text-background">Buyer Dashboard</Link></li>
                  <li><Link href="/merchant/onboarding" className="hover:text-background">Merchant Onboarding</Link></li>
                </ul>
              </div>
              <div>
                <h4 className="mb-3 text-sm font-black uppercase">Platform</h4>
                <ul className="space-y-2 text-sm text-background/80">
                  <li><Link href="/admin/tenants" className="hover:text-background">Admin Console</Link></li>
                  <li><Link href="/buyer/connections" className="hover:text-background">Connections</Link></li>
                  <li><Link href="/merchant/revenue" className="hover:text-background">Revenue</Link></li>
                </ul>
              </div>
              <div>
                <h4 className="mb-3 text-sm font-black uppercase">Contact</h4>
                <ul className="space-y-2 text-sm text-background/80">
                  <li>support@mcp-marketplace.local</li>
                  <li>security@mcp-marketplace.local</li>
                  <li>San Francisco, CA</li>
                </ul>
              </div>
            </div>
            <div className="mt-10 border-t border-background/25 pt-6 text-center text-sm text-background/70">
              &copy; 2026 MCP Marketplace. All rights reserved.
            </div>
          </div>
        </footer>
      </div>
    </BeamsBackground>
  )
}
