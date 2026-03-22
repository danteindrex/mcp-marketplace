'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { ArrowRight, Check, Zap, Shield, TrendingUp, Users, Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { JsonLd } from '@/components/json-ld'
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
import { Button as RetroButton, Card as RetroCard, Badge as RetroBadge, Text } from '@/components/retroui'
import { BarChart } from '@/components/retroui/charts/BarChart'
import { FAQSection, FooterSection } from '@/components/blocks/marketing'
import { fetchCurrentUser, type CurrentUser } from '@/lib/api-client'
import { getSiteUrl } from '@/lib/site'
import { SALES_EMAIL, SECURITY_EMAIL, SUPPORT_EMAIL } from '@/lib/contact'

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

const homeFaqItems = [
  {
    question: 'What is Model Context Protocol (MCP)?',
    answer:
      'Model Context Protocol is a standard for connecting AI clients to external tools and data sources through predictable auth, transport, and capability contracts.',
  },
  {
    question: 'How does MCP Marketplace make installing a server feel seamless?',
    answer:
      'Buyers select a client, verify metadata, approve scopes, settle payment only if needed, and launch a client-specific one-click action instead of stitching the flow together manually.',
  },
  {
    question: 'Why is the install flow split into steps?',
    answer:
      'Separating client selection, metadata checks, permissions, payment, and launch removes dead ends and makes every blocker explainable before the user reaches the final install action.',
  },
  {
    question: 'Can I use the same server across clients?',
    answer:
      'Yes. Marketplace installs are designed around a buyer hub and client-specific launch actions so the same account can be reused across supported clients like Codex, VS Code, Cursor, Claude, and ChatGPT.',
  },
  {
    question: 'What happens when a paid install is blocked?',
    answer:
      'The flow shows payment as a separate step so the buyer knows whether to fund a wallet, settle an x402 challenge, or retry after the entitlement is issued.',
  },
]

export default function HomePage() {
  const [featuredServers, setFeaturedServers] = useState<Server[]>([])
  const [currentUser, setCurrentUser] = useState<CurrentUser | null>(null)
  const [authResolved, setAuthResolved] = useState(false)
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
    fetchCurrentUser()
      .then(setCurrentUser)
      .finally(() => setAuthResolved(true))
  }, [])

  const dashboardPath =
    currentUser?.role === 'admin'
      ? '/admin/tenants'
      : currentUser?.role === 'merchant'
        ? '/merchant/onboarding'
        : '/buyer/dashboard'

  const categories = [
    { name: 'Data', count: featuredServers.filter(s => s.category === 'data').length },
    { name: 'Automation', count: featuredServers.filter(s => s.category === 'automation').length },
    { name: 'AI', count: featuredServers.filter(s => s.category === 'ai').length },
    { name: 'Integration', count: featuredServers.filter(s => s.category === 'integration').length },
  ]
  const siteUrl = getSiteUrl()
  const structuredData = [
    {
      '@context': 'https://schema.org',
      '@type': 'Organization',
      name: 'MCP Marketplace',
      url: siteUrl,
      description:
        'Marketplace for discovering, buying, installing, and managing Model Context Protocol servers.',
    },
    {
      '@context': 'https://schema.org',
      '@type': 'FAQPage',
      mainEntity: homeFaqItems.map(item => ({
        '@type': 'Question',
        name: item.question,
        acceptedAnswer: {
          '@type': 'Answer',
          text: item.answer,
        },
      })),
    },
    {
      '@context': 'https://schema.org',
      '@type': 'HowTo',
      name: 'Install an MCP server with MCP Marketplace',
      description:
        'Choose a client, verify metadata, approve scopes, settle payment if needed, and launch a client-specific install action.',
      step: [
        {
          '@type': 'HowToStep',
          name: 'Choose a client',
          text: 'Select Codex, Claude, Cursor, VS Code, or ChatGPT so the marketplace can generate the right install action.',
        },
        {
          '@type': 'HowToStep',
          name: 'Verify metadata',
          text: 'Confirm CIMD, OAuth metadata, and hub readiness before continuing.',
        },
        {
          '@type': 'HowToStep',
          name: 'Review scopes',
          text: 'Approve the scopes needed for the server to operate safely.',
        },
        {
          '@type': 'HowToStep',
          name: 'Settle payment when required',
          text: 'Handle wallet or x402 settlement separately so payment never looks like an auth failure.',
        },
        {
          '@type': 'HowToStep',
          name: 'Launch the install action',
          text: 'Open the generated one-click action or local bridge flow for the selected client.',
        },
      ],
    },
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
        <JsonLd data={structuredData} />
        <nav className="sticky top-0 z-50 border-b border-border bg-background/95 backdrop-blur">
          <div className="mx-auto flex max-w-7xl flex-col gap-3 px-4 py-3 sm:h-16 sm:flex-row sm:items-center sm:justify-between sm:px-6 sm:py-0 lg:px-8">
            <DarkModeOnly>
              <div className="flex items-center gap-2">
                <Sparkles className="h-5 w-5 shrink-0 text-primary sm:h-6 sm:w-6" />
                <Text as="span" variant="h6" className="text-base sm:text-xl">MCP Marketplace</Text>
              </div>
            </DarkModeOnly>
            <LightModeOnly>
              <RetroCard className={`flex w-fit items-center gap-2 px-3 py-2 ${browseFillClass}`}>
                <Sparkles className="h-5 w-5 shrink-0" />
                <Text as="span" variant="h6" className="text-sm sm:text-lg">MCP Marketplace</Text>
              </RetroCard>
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
              <div className="grid w-full grid-cols-[minmax(0,1fr)_minmax(0,1.25fr)] gap-2 sm:flex sm:w-auto sm:items-center sm:gap-3">
                {currentUser ? (
                  <Button variant="outline" size="sm" asChild className={publishBrutalButtonClass}><Link href={dashboardPath}>Dashboard</Link></Button>
                ) : authResolved ? (
                  <Button variant="outline" size="sm" asChild className={publishBrutalButtonClass}><Link href="/login">Login</Link></Button>
                ) : null}
                <Button size="sm" asChild className={`h-12 w-full ${browseBrutalButtonClass}`}><Link href="/marketplace">{currentUser ? 'Open Marketplace' : 'Get Started'}</Link></Button>
              </div>
            </DarkModeOnly>
            <LightModeOnly>
              <div className="grid w-full grid-cols-[minmax(0,1fr)_minmax(0,1.25fr)] gap-2 sm:flex sm:w-auto sm:items-center sm:gap-2">
                {currentUser ? (
                  <RetroButton variant="outline" size="sm" asChild className={publishFillClass}><Link href={dashboardPath}>Dashboard</Link></RetroButton>
                ) : authResolved ? (
                  <RetroButton variant="outline" size="sm" asChild className={publishFillClass}><Link href="/login">Login</Link></RetroButton>
                ) : null}
                <RetroButton size="sm" asChild className={`h-12 w-full ${browseFillClass}`}><Link href="/marketplace">{currentUser ? 'Open Marketplace' : 'Get Started'}</Link></RetroButton>
              </div>
            </LightModeOnly>
          </div>
        </nav>

        <section className="relative mx-auto w-full max-w-7xl overflow-hidden px-4 py-14 sm:px-6 sm:py-24 lg:px-8 lg:py-32">
          <div className="pointer-events-none absolute -right-6 top-4 hidden lg:block">
            <BurstShape size={72} className="text-tertiary" />
          </div>
          <div className="pointer-events-none absolute left-8 bottom-8 hidden lg:block">
            <Star5Shape size={56} className="text-secondary" />
          </div>
          <div className="space-y-6 text-center sm:space-y-8">
            <Text as="h1" variant="h1" className="mx-auto max-w-5xl text-4xl leading-[0.95] tracking-tight sm:text-6xl sm:leading-tight">
              <span className="relative block min-h-[5.25rem] sm:min-h-[7.5rem]">
                <span className="pointer-events-none invisible block">
                  Discover &amp; Deploy MCP Servers
                </span>
                <span className="absolute inset-0 flex items-center justify-center">
                  <TypeWriter text="Discover & Deploy MCP Servers" speed={50} deleteSpeed={34} holdDuration={2200} loop loopDelay={450} className="block bg-gradient-to-r from-primary to-primary/70 bg-clip-text text-transparent" />
                </span>
              </span>
            </Text>
            <div className="mx-auto max-w-2xl px-2 sm:px-0">
              <ScrollText speed="normal">
                <div className="py-2 text-pretty text-sm leading-relaxed sm:py-4 sm:text-lg">
                  The comprehensive marketplace for Model Context Protocol servers. Find, install, and manage servers securely with enterprise-grade features.
                </div>
              </ScrollText>
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
              <Text variant="small" className="uppercase tracking-wide text-muted-foreground">Pricing</Text>
              <Text variant="h3" className="mt-2 uppercase">Simple Pricing</Text>
              <Text variant="body" className="mt-2 text-muted-foreground">No hidden fees. No surprises. Choose the plan that works for you.</Text>
            </div>
            <div className="grid grid-cols-3 gap-1.5 overflow-hidden md:gap-6">
              {pricingPlans.map(plan => (
                <BoldCard
                  key={plan.title}
                  className={plan.popular ? 'relative flex h-full min-w-0 flex-col border-black bg-[hsl(0_84%_71%)] text-black' : 'flex h-full min-w-0 flex-col'}
                >
                  {plan.popular ? (
                    <Sticker className="absolute -top-2 right-1 z-10 border-2 border-black bg-[hsl(0_84%_71%)] px-2 py-1 text-[9px] text-black shadow-[2px_2px_0px_#000] md:-top-3 md:-right-3 md:px-3 md:py-1.5 md:text-xs md:shadow-[3px_3px_0px_#000]">
                      Popular
                    </Sticker>
                  ) : null}
                  <BoldCardHeader className={`${plan.popular ? 'bg-[hsl(0_84%_71%)]' : 'bg-muted'} p-2.5 md:p-6`}>
                    <BoldCardTitle className="break-words text-[11px] leading-3.5 md:text-xl">{plan.title}</BoldCardTitle>
                    <BoldCardDescription className="text-[9px] leading-3 md:text-sm">{plan.description}</BoldCardDescription>
                    <Text as="p" variant="h3" className="mt-2 text-base md:text-3xl">
                      {plan.price}
                      <Text as="span" variant="body" className="ml-1 text-[9px] md:text-base">{plan.period}</Text>
                    </Text>
                  </BoldCardHeader>
                  <BoldCardContent className="flex-1 space-y-1.5 p-2.5 md:space-y-3 md:p-6">
                    {plan.features.map(feature => (
                      <div key={feature} className="flex min-h-6 items-start gap-1 text-[9px] font-semibold leading-3 md:min-h-0 md:gap-2 md:text-sm md:leading-4">
                        <Check className="mt-0.5 h-3 w-3 shrink-0 md:h-4 md:w-4" />
                        <span>{feature}</span>
                      </div>
                    ))}
                  </BoldCardContent>
                  <BoldCardFooter className="mt-auto p-2.5 md:p-6">
                    <Button className={`h-7 w-full px-1.5 text-[9px] leading-none md:h-9 md:px-4 md:text-sm ${orangeBrutalButtonClass}`} size="sm">
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
            <ScrollText speed="normal"><div className="flex items-center justify-between"><div><Text variant="h3">Featured Servers</Text><Text variant="body" className="mt-2 text-muted-foreground">Popular and verified MCP servers</Text></div></div></ScrollText>
            <Card className="p-6">
              <ScrollText speed="normal"><Text variant="h6" className="mb-4">Marketplace Activity</Text></ScrollText>
              <BarChart
                data={landingChart}
                index="name"
                categories={['orders']}
              />
            </Card>
            <DarkModeOnly>
              <div className="grid grid-cols-2 gap-2 md:grid-cols-2 md:gap-6 lg:grid-cols-3">
                {featuredServers.slice(0, 3).map(server => (
                  <Link key={server.id} href={`/marketplace/${server.slug}`}>
                    <SpotlightCard className="h-full rounded-lg border border-border bg-card p-3 md:p-6" spotlightColor="rgba(59, 130, 246, 0.3)">
                      <div className="mb-2 flex items-start justify-between md:mb-4"><div><Text variant="h6" className="text-sm md:text-lg">{server.name}</Text><Text variant="small" className="text-[10px] text-muted-foreground md:text-sm">by {server.author}</Text></div></div>
                      <Text variant="small" className="mb-2 line-clamp-3 text-[10px] leading-4 text-muted-foreground md:mb-4 md:text-sm">{server.description}</Text>
                    </SpotlightCard>
                  </Link>
                ))}
              </div>
            </DarkModeOnly>
            <LightModeOnly>
              <div className="grid grid-cols-2 gap-2 md:grid-cols-2 md:gap-6 lg:grid-cols-3">
                {featuredServers.slice(0, 3).map(server => (
                  <Link key={server.id} href={`/marketplace/${server.slug}`}>
                    <RetroCard className="h-full p-3 md:p-6">
                      <ScrollText speed="normal">
                        <div>
                          <div className="mb-2 flex items-start justify-between md:mb-4"><div><Text variant="h6" className="text-sm md:text-lg">{server.name}</Text><Text variant="small" className="text-[10px] text-gray-700 md:text-sm">by {server.author}</Text></div>{server.verified && <RetroBadge variant="default" className="px-1.5 py-0.5 text-[9px] md:px-2 md:py-1 md:text-xs">VERIFIED</RetroBadge>}</div>
                          <Text variant="small" className="mb-2 line-clamp-3 text-[10px] leading-4 text-gray-800 md:mb-4 md:text-sm">{server.description}</Text>
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
          <ScrollText speed="normal"><div className="text-center mb-8"><Text variant="h3">Why Choose MCP Marketplace?</Text></div></ScrollText>
          <div className="grid grid-cols-2 gap-2 md:grid-cols-2 md:gap-6 lg:grid-cols-3">
            {features.map((feature, index) => {
              const Icon = feature.icon
              return (
                <BoldCard
                  key={feature.title}
                  interactive
                  className={`min-w-0 h-full ${featureContainerColors[index % featureContainerColors.length]}`}
                >
                  <BoldCardHeader className="p-3 md:p-6">
                    <BoldCardTitle className="break-words text-[11px] leading-3.5 md:text-xl">{feature.title}</BoldCardTitle>
                    <BoldCardDescription className="text-[10px] leading-4 md:text-sm">{feature.description}</BoldCardDescription>
                  </BoldCardHeader>
                  <BoldCardContent className="p-3 md:p-6">
                    <div className={`flex h-10 w-10 items-center justify-center border-[3px] border-foreground shadow-[3px_3px_0px_hsl(var(--shadow-color))] md:h-14 md:w-14 ${featureIconColors[index % featureIconColors.length]}`}>
                      <Icon className="h-4 w-4 md:h-6 md:w-6" />
                    </div>
                  </BoldCardContent>
                  <BoldCardFooter className="p-3 md:p-6">
                    <Text variant="caption" className="text-[9px] uppercase tracking-wide md:text-xs">MCP Capability</Text>
                  </BoldCardFooter>
                </BoldCard>
              )
            })}
          </div>
        </section>

        <section id="categories" className="px-4 sm:px-6 lg:px-8 py-16 max-w-7xl mx-auto w-full border-t border-border">
          <ScrollText speed="normal"><div className="mb-8"><Text variant="h3">Browse by Category</Text></div></ScrollText>
          <LightModeOnly>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:gap-4">{categories.map(category => <Link key={category.name} href={`/marketplace?category=${category.name.toLowerCase()}`}><RetroCard className="h-full p-4 md:p-6"><ScrollText speed="normal"><div className="flex items-center justify-between gap-2"><div><Text variant="h6" className="break-words text-sm leading-5 md:text-lg">{category.name}</Text><Text variant="small" className="text-xs md:text-sm">{category.count} servers</Text></div><ArrowRight className="hidden md:block md:h-4 md:w-4" /></div></ScrollText></RetroCard></Link>)}</div>
          </LightModeOnly>
          <DarkModeOnly>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:gap-4">{categories.map(category => <Link key={category.name} href={`/marketplace?category=${category.name.toLowerCase()}`}><Card className="h-full cursor-pointer p-4 transition-colors hover:border-primary md:p-6"><div className="flex items-center justify-between gap-2"><div><Text variant="h6" className="break-words text-sm leading-5 md:text-lg">{category.name}</Text><Text variant="small" className="text-xs opacity-75 md:text-sm">{category.count} servers</Text></div><ArrowRight className="hidden opacity-50 md:block md:h-4 md:w-4" /></div></Card></Link>)}</div>
          </DarkModeOnly>
        </section>

        <section className="px-4 sm:px-6 lg:px-8 py-16 max-w-5xl mx-auto w-full border-t border-border">
          <div className="grid gap-6 lg:grid-cols-[1.25fr_0.75fr]">
            <Card className="border-2 border-foreground p-8 shadow-[4px_4px_0px_hsl(var(--shadow-color))]">
              <p className="text-sm font-black uppercase tracking-wide text-muted-foreground">Seamless Install</p>
              <h2 className="mt-3 text-3xl font-black uppercase">How we make MCP installation feel simple</h2>
              <p className="mt-4 max-w-3xl text-sm leading-7 text-muted-foreground">
                MCP Marketplace makes installation feel seamless by turning protocol complexity into
                a guided flow: choose a client, verify install metadata, approve scopes, settle
                payment only when required, and launch the exact action that client expects.
              </p>
              <div className="mt-6 grid gap-4 sm:grid-cols-2">
                <div className="rounded-lg border-2 border-foreground p-4">
                  <h3 className="font-black uppercase">What buyers see</h3>
                  <p className="mt-2 text-sm text-muted-foreground">
                    A short sequence with one decision per step and clear recovery if auth, payment,
                    or client readiness blocks the install.
                  </p>
                </div>
                <div className="rounded-lg border-2 border-foreground p-4">
                  <h3 className="font-black uppercase">What the platform handles</h3>
                  <p className="mt-2 text-sm text-muted-foreground">
                    CIMD checks, OAuth metadata, scope validation, entitlement handling, buyer hub
                    connection setup, and the final one-click launch action.
                  </p>
                </div>
              </div>
            </Card>

            <Card className="border-2 border-foreground p-8 shadow-[4px_4px_0px_hsl(var(--shadow-color))]">
              <p className="text-sm font-black uppercase tracking-wide text-muted-foreground">Install Path</p>
              <ol className="mt-4 space-y-4 text-sm text-muted-foreground">
                <li><span className="font-black text-foreground">1.</span> Pick Codex, Claude, Cursor, VS Code, or ChatGPT.</li>
                <li><span className="font-black text-foreground">2.</span> Check metadata before asking the buyer to trust anything.</li>
                <li><span className="font-black text-foreground">3.</span> Review the scopes the server needs.</li>
                <li><span className="font-black text-foreground">4.</span> Handle payment separately when the server is not free.</li>
                <li><span className="font-black text-foreground">5.</span> Launch the generated install action or local bridge flow.</li>
              </ol>
              <Button size="sm" asChild className={`mt-6 ${orangeBrutalButtonClass}`}>
                <Link href="/guides/mcp-server-installation">Read the install guide</Link>
              </Button>
            </Card>
          </div>
        </section>

        <section className="px-4 sm:px-6 lg:px-8 py-16 max-w-7xl mx-auto w-full border-t border-border">
          <div className="mb-8">
            <h2 className="text-3xl font-black uppercase">Canonical Guides</h2>
            <p className="mt-2 text-sm text-muted-foreground">
              Use these guides to understand MCP installation, client setup, pricing, and auth without leaving the marketplace context.
            </p>
          </div>
          <div className="grid grid-cols-2 gap-2 md:grid-cols-2 md:gap-4 lg:grid-cols-4">
            {[
              { href: '/guides', label: 'Guide Library' },
              { href: '/guides/mcp-server-installation', label: 'Install an MCP Server' },
              { href: '/guides/install-mcp-in-codex', label: 'Install MCP in Codex' },
              { href: '/guides/x402-pricing-explained', label: 'x402 Pricing Explained' },
            ].map(item => (
              <Link key={item.href} href={item.href}>
                <Card className="h-full border-2 border-foreground p-3 shadow-[4px_4px_0px_hsl(var(--shadow-color))] transition-all hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-[6px_6px_0px_hsl(var(--shadow-color))] md:p-6">
                  <p className="break-words text-[10px] font-black uppercase leading-4 md:text-base">{item.label}</p>
                </Card>
              </Link>
            ))}
          </div>
        </section>

        <FAQSection.Accordion
          title="Frequently Asked Questions"
          subtitle="FAQ"
          items={homeFaqItems}
        />

        <section className="px-4 sm:px-6 lg:px-8 py-16 max-w-7xl mx-auto w-full border-t border-border">
          <Card className="relative overflow-hidden border-border bg-gradient-to-br from-accent to-background p-6 text-center sm:p-12">
            <div className="pointer-events-none absolute -right-8 top-4 hidden md:block">
              <LightningShape size={80} className="text-tertiary" />
            </div>
            <ScrollText speed="normal">
              <div>
                <h2 className="mb-4 text-2xl font-black uppercase sm:text-3xl">Ready to get started?</h2>
                <p className="mx-auto mb-8 max-w-xl text-sm leading-6 text-muted-foreground sm:text-base">Explore thousands of verified MCP servers and integrate them into your workflow in minutes.</p>
              </div>
            </ScrollText>
            <Button size="lg" asChild className={orangeBrutalButtonClass}>
              <Link href="/marketplace">Explore Marketplace</Link>
            </Button>
          </Card>
        </section>

        <section id="contact" className="px-4 sm:px-6 lg:px-8 py-16 max-w-7xl mx-auto w-full border-t border-border">
          <div className="grid gap-6 lg:grid-cols-2">
            <Card className="relative overflow-hidden border-2 border-foreground bg-accent p-6 sm:p-8">
              <Sticker className="mb-4 bg-[hsl(25_100%_55%)] text-black border-2 border-black shadow-[3px_3px_0px_#000]">Contact Team</Sticker>
              <h3 className="mb-3 pr-10 text-2xl font-black uppercase sm:pr-0 sm:text-3xl">Need a custom deployment?</h3>
              <p className="text-sm text-foreground/80 max-w-prose">
                Tell us your tenant size, throughput goals, and compliance requirements. We can provision hosted or hybrid MCP infrastructure with role-based access and billing controls.
              </p>
              <div className="mt-6 min-w-0 space-y-2 text-sm font-semibold">
                <p className="break-all">Sales: {SALES_EMAIL}</p>
                <p className="break-all">Security: {SECURITY_EMAIL}</p>
                <p>Support SLA: 24/7 enterprise response</p>
              </div>
              <BurstShape size={56} className="absolute bottom-3 right-3 text-tertiary sm:-bottom-5 sm:-right-5 sm:h-auto sm:w-auto" />
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
                { label: SUPPORT_EMAIL, href: `mailto:${SUPPORT_EMAIL}` },
                { label: SECURITY_EMAIL, href: `mailto:${SECURITY_EMAIL}` },
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

