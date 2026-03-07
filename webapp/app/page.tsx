'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { ArrowRight, Check, Zap, Shield, TrendingUp, Users, Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import {
  Card as BoldCard,
  CardContent as BoldCardContent,
  CardDescription as BoldCardDescription,
  CardFooter as BoldCardFooter,
  CardHeader as BoldCardHeader,
  CardTitle as BoldCardTitle,
} from '@/components/ui/boldkit-card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Sticker } from '@/components/ui/sticker'
import { BurstShape, LightningShape, Star5Shape } from '@/components/ui/shapes'
import { fetchFeaturedServers, type Server } from '@/lib/api-client'
import { TypeWriter, ScrollText, SpotlightCard } from '@/components/kokonut'
import { LightModeOnly, DarkModeOnly } from '@/components/theme-aware'
import { Button as RetroButton, Card as RetroCard, Badge as RetroBadge } from '@/components/retroui'
import { BarChart } from '@/components/retroui/charts/BarChart'
import { FAQSection, FooterSection } from '@/components/blocks/marketing'

const features = [
  { icon: Zap, title: 'Easy Installation', description: 'One-click setup with automatic scope negotiation and permission management' },
  { icon: Shield, title: 'Secure Integration', description: 'Enterprise-grade security with OAuth2, token rotation, and audit logging' },
  { icon: TrendingUp, title: 'Usage Analytics', description: 'Real-time metrics and performance monitoring for all your connections' },
  { icon: Users, title: 'Community Driven', description: 'Discover and share servers built by the global developer community' },
]

const featureContainerColors = [
  'bg-primary/10',
  'bg-secondary/10',
  'bg-accent/10',
  'bg-info/10',
  'bg-success/10',
  'bg-warning/10',
]

const featureIconColors = [
  'bg-primary',
  'bg-secondary',
  'bg-accent',
  'bg-info',
  'bg-success',
  'bg-warning',
]

const pricingPlans = [
  {
    title: 'Starter',
    price: '$0',
    period: '/month',
    description: 'Perfect for trying out',
    features: ['5 projects', '1 team member', 'Basic analytics', 'Email support'],
    cta: 'Get Started',
    popular: false,
  },
  {
    title: 'Pro',
    price: '$29',
    period: '/month',
    description: 'Best for growing teams',
    features: ['Unlimited projects', '10 team members', 'Advanced analytics', 'Priority support', 'Custom integrations'],
    cta: 'Get Started',
    popular: true,
  },
  {
    title: 'Enterprise',
    price: '$99',
    period: '/month',
    description: 'For large organizations',
    features: ['Everything in Pro', 'Unlimited team members', 'Dedicated account manager', 'Custom SLA', 'On-premise option'],
    cta: 'Contact Sales',
    popular: false,
  },
]

export default function HomePage() {
  const [featuredServers, setFeaturedServers] = useState<Server[]>([])
  const browseFillClass = '!bg-[hsl(0_84%_71%)] !text-[hsl(240_10%_10%)] !border-black'
  const publishFillClass = '!bg-[hsl(174_62%_56%)] !text-[hsl(240_10%_10%)] !border-black'
  const orangeBrutalButtonClass = 'bg-[hsl(var(--chart-3))] text-black border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] hover:translate-x-1 hover:translate-y-1'
  const browseBrutalButtonClass = 'bg-[hsl(0_84%_71%)] text-[hsl(240_10%_10%)] border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] hover:translate-x-1 hover:translate-y-1'
  const publishBrutalButtonClass = 'bg-[hsl(174_62%_56%)] text-[hsl(240_10%_10%)] border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] hover:translate-x-1 hover:translate-y-1'
  const heroPrimaryButtonClass = '!bg-[hsl(0_84%_71%)] !text-[hsl(240_10%_10%)] border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] hover:translate-x-1 hover:translate-y-1'
  const heroSecondaryButtonClass = '!bg-[hsl(174_62%_56%)] !text-[hsl(240_10%_10%)] border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,0.85)] hover:shadow-[2px_2px_0px_0px_rgba(0,0,0,0.85)] hover:translate-x-1 hover:translate-y-1'
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
      <div
        className="flex flex-col min-h-screen bg-background"
        style={{
          backgroundImage:
            'linear-gradient(rgba(80, 80, 80, 0.32) 1px, transparent 1px), linear-gradient(90deg, rgba(80, 80, 80, 0.32) 1px, transparent 1px)',
          backgroundSize: '40px 40px',
        }}
      >
        <nav className="border-b border-border sticky top-0 z-50 bg-background/95 backdrop-blur">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
            <DarkModeOnly>
              <div className="flex items-center gap-2"><Sparkles className="w-6 h-6 text-primary" /><span className="text-xl font-bold">MCP Marketplace</span></div>
            </DarkModeOnly>
            <LightModeOnly>
              <RetroCard className={`px-3 py-2 flex items-center gap-2 ${browseFillClass}`}><Sparkles className="w-5 h-5" /><span className="text-lg font-black">MCP Marketplace</span></RetroCard>
            </LightModeOnly>

            <DarkModeOnly>
              <div className="hidden md:flex items-center gap-8">
                <Link href="/marketplace" className="text-sm hover:text-primary transition-colors">Browse</Link>
                <Link href="#features" className="text-sm hover:text-primary transition-colors">Features</Link>
                <Link href="#categories" className="text-sm hover:text-primary transition-colors">Categories</Link>
                <Link href="/pricing" className="text-sm hover:text-primary transition-colors">Pricing</Link>
              </div>
            </DarkModeOnly>
            <LightModeOnly>
              <div className="hidden md:flex items-center gap-2">
                <RetroButton variant="outline" size="sm" asChild className={browseFillClass}><Link href="/marketplace">Browse</Link></RetroButton>
                <RetroButton variant="outline" size="sm" asChild className={browseFillClass}><Link href="#features">Features</Link></RetroButton>
                <RetroButton variant="outline" size="sm" asChild className={browseFillClass}><Link href="#categories">Categories</Link></RetroButton>
                <RetroButton variant="outline" size="sm" asChild className={browseFillClass}><Link href="/pricing">Pricing</Link></RetroButton>
              </div>
            </LightModeOnly>

            <DarkModeOnly>
              <div className="flex items-center gap-3">
                <Button variant="outline" size="sm" asChild className={publishBrutalButtonClass}><Link href="/login">Login</Link></Button>
                <Button size="sm" asChild className={browseBrutalButtonClass}><Link href="/marketplace">Get Started</Link></Button>
              </div>
            </DarkModeOnly>
            <LightModeOnly>
              <div className="flex items-center gap-2">
                <RetroButton variant="outline" size="sm" asChild className={publishFillClass}><Link href="/login">Login</Link></RetroButton>
                <RetroButton size="sm" asChild className={browseFillClass}><Link href="/marketplace">Get Started</Link></RetroButton>
              </div>
            </LightModeOnly>
          </div>
        </nav>

        <section className="relative px-4 sm:px-6 lg:px-8 py-20 sm:py-32 max-w-7xl mx-auto w-full overflow-hidden">
          <div className="pointer-events-none absolute -right-6 top-4 hidden lg:block">
            <BurstShape size={72} className="text-tertiary" />
          </div>
          <div className="pointer-events-none absolute left-8 bottom-8 hidden lg:block">
            <Star5Shape size={56} className="text-secondary" />
          </div>
          <div className="text-center space-y-8">
            <h1 className="text-5xl sm:text-6xl font-bold tracking-tight text-balance">
              <TypeWriter text="Discover & Deploy MCP Servers" speed={50} deleteSpeed={34} holdDuration={2200} loop loopDelay={450} className="bg-gradient-to-r from-primary to-primary/70 bg-clip-text text-transparent block" />
            </h1>
            <div className="text-xl text-muted-foreground max-w-2xl mx-auto text-balance h-24 overflow-hidden">
              <ScrollText speed="normal" className="h-full"><div className="py-6">The comprehensive marketplace for Model Context Protocol servers. Find, install, and manage servers securely with enterprise-grade features.</div></ScrollText>
            </div>

            <DarkModeOnly>
              <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
                <Button size="lg" asChild className={`w-full sm:w-auto ${orangeBrutalButtonClass} button-orange-solid`}><Link href="/marketplace">Browse Marketplace <ArrowRight className="ml-2 w-4 h-4" /></Link></Button>
                <Button size="lg" variant="outline" asChild className={`w-full sm:w-auto ${orangeBrutalButtonClass} button-orange-solid`}><Link href="/merchant/onboarding">Publish Server</Link></Button>
              </div>
            </DarkModeOnly>
            <LightModeOnly>
              <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
                <RetroButton variant="outline" size="lg" asChild className={`w-full sm:w-auto gap-2 ${heroPrimaryButtonClass}`}><Link href="/marketplace">Browse Marketplace <ArrowRight className="ml-2 w-4 h-4" /></Link></RetroButton>
                <RetroButton variant="outline" size="lg" asChild className={`w-full sm:w-auto ${heroSecondaryButtonClass}`}><Link href="/merchant/onboarding">Publish Server</Link></RetroButton>
              </div>
            </LightModeOnly>
          </div>

          <section id="pricing" className="mt-16 border-2 border-foreground bg-card p-6 sm:p-8 shadow-[6px_6px_0px_hsl(var(--shadow-color))]">
            <div className="mb-8 text-center">
              <p className="text-sm font-bold uppercase tracking-wide text-muted-foreground">Pricing</p>
              <h2 className="mt-2 text-3xl font-black uppercase">Simple Pricing</h2>
              <p className="mt-2 text-muted-foreground">No hidden fees. No surprises. Choose the plan that works for you.</p>
            </div>
            <div className="grid gap-6 md:grid-cols-3">
              {pricingPlans.map(plan => (
                <BoldCard
                  key={plan.title}
                  className={plan.popular ? 'border-black bg-[hsl(0_84%_71%)] text-black relative' : ''}
                >
                  {plan.popular ? (
                    <Sticker className="absolute -top-3 -right-3 z-10 bg-[hsl(0_84%_71%)] text-black border-2 border-black shadow-[3px_3px_0px_#000]">
                      Popular
                    </Sticker>
                  ) : null}
                  <BoldCardHeader className={plan.popular ? 'bg-[hsl(0_84%_71%)]' : 'bg-muted'}>
                    <BoldCardTitle>{plan.title}</BoldCardTitle>
                    <BoldCardDescription>{plan.description}</BoldCardDescription>
                    <p className="mt-2 text-3xl font-black">
                      {plan.price}
                      <span className="ml-1 text-base font-semibold">{plan.period}</span>
                    </p>
                  </BoldCardHeader>
                  <BoldCardContent className="space-y-3">
                    {plan.features.map(feature => (
                      <div key={feature} className="flex items-center gap-2 text-sm font-semibold">
                        <Check className="h-4 w-4" />
                        <span>{feature}</span>
                      </div>
                    ))}
                  </BoldCardContent>
                  <BoldCardFooter>
                    <Button className={`w-full ${orangeBrutalButtonClass}`} size="sm">
                      {plan.cta}
                    </Button>
                  </BoldCardFooter>
                </BoldCard>
              ))}
            </div>
          </section>
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
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {features.map((feature, index) => {
              const Icon = feature.icon
              return (
                <BoldCard
                  key={feature.title}
                  interactive
                  className={featureContainerColors[index % featureContainerColors.length]}
                >
                  <BoldCardHeader>
                    <BoldCardTitle>{feature.title}</BoldCardTitle>
                    <BoldCardDescription>{feature.description}</BoldCardDescription>
                  </BoldCardHeader>
                  <BoldCardContent>
                    <div className={`w-14 h-14 flex items-center justify-center ${featureIconColors[index % featureIconColors.length]} border-[3px] border-foreground shadow-[3px_3px_0px_hsl(var(--shadow-color))]`}>
                      <Icon className="h-6 w-6" />
                    </div>
                  </BoldCardContent>
                  <BoldCardFooter>
                    <p className="text-xs font-bold uppercase tracking-wide">MCP Capability</p>
                  </BoldCardFooter>
                </BoldCard>
              )
            })}
          </div>
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

        <FAQSection.Accordion
          title="Frequently Asked Questions"
          subtitle="FAQ"
          items={[
            {
              question: 'What is Model Context Protocol (MCP)?',
              answer: 'MCP is a standardized protocol that enables secure communication between AI models and tools.',
            },
            {
              question: 'How do I install a server?',
              answer: 'Pick a server, click install, approve scopes, and connect once to your personal hub.',
            },
            {
              question: 'Can I use the same account across clients?',
              answer: 'Yes. Once connected, you can use marketplace access across supported clients like Codex, VS Code, Cursor, and Claude.',
            },
          ]}
        />

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
            <Button size="lg" asChild className={orangeBrutalButtonClass}>
              <Link href="/marketplace">Explore Marketplace</Link>
            </Button>
          </Card>
        </section>

        <section id="contact" className="px-4 sm:px-6 lg:px-8 py-16 max-w-7xl mx-auto w-full border-t border-border">
          <div className="grid gap-6 lg:grid-cols-2">
            <Card className="relative overflow-hidden border-2 border-foreground bg-accent p-8">
              <Sticker className="mb-4 bg-[hsl(25_100%_55%)] text-black border-2 border-black shadow-[3px_3px_0px_#000]">Contact Team</Sticker>
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
                <Button type="button" className={`w-full ${orangeBrutalButtonClass}`}>Request Consultation</Button>
              </form>
            </Card>
          </div>
        </section>

        <FooterSection.MultiColumn
          logo={
            <div className="flex items-center gap-2">
              <div className="h-8 w-8 border-2 border-background bg-[hsl(var(--tertiary))]" />
              <span className="text-xl font-black uppercase">MCP Marketplace</span>
            </div>
          }
          description="Marketplace for secure MCP discovery, deployment, billing, and tenancy control."
          columns={[
            {
              title: 'Product',
              links: [
                { label: 'Marketplace', href: '/marketplace' },
                { label: 'Buyer Dashboard', href: '/buyer/dashboard' },
                { label: 'Merchant Onboarding', href: '/merchant/onboarding' },
              ],
            },
            {
              title: 'Platform',
              links: [
                { label: 'Admin Console', href: '/admin/tenants' },
                { label: 'Connections', href: '/buyer/connections' },
                { label: 'Revenue', href: '/merchant/revenue' },
              ],
            },
            {
              title: 'Contact',
              links: [
                { label: 'support@mcp-marketplace.local', href: 'mailto:support@mcp-marketplace.local' },
                { label: 'security@mcp-marketplace.local', href: 'mailto:security@mcp-marketplace.local' },
                { label: 'San Francisco, CA', href: '/contact' },
              ],
            },
          ]}
          socialLinks={[
            { platform: 'twitter', href: 'https://twitter.com/' },
            { platform: 'github', href: 'https://github.com/danteindrex/mcp-marketplace' },
            { platform: 'linkedin', href: 'https://linkedin.com/' },
          ]}
          copyright="© 2026 MCP Marketplace. All rights reserved."
        />
      </div>
  )
}

