import type { Metadata } from 'next'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import { JsonLd } from '@/components/json-ld'
import { Text } from '@/components/retroui/Text'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { fetchPublicServers } from '@/lib/public-marketplace'
import { createBreadcrumbJsonLd, toAbsoluteUrl } from '@/lib/seo'
import { slugifyText } from '@/lib/slugs'

interface PageProps {
  params: Promise<{ publisherSlug: string }>
}

async function loadPublisherProfile(publisherSlug: string) {
  const servers = await fetchPublicServers()
  const publisherServers = servers.filter(server => slugifyText(server.author) === publisherSlug)
  if (publisherServers.length === 0) {
    return null
  }

  const publisher = publisherServers[0].author
  const categories = Array.from(new Set(publisherServers.map(server => server.category)))
  return { publisher, publisherServers, categories }
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { publisherSlug } = await params
  const profile = await loadPublisherProfile(publisherSlug)

  if (!profile) {
    return {
      title: 'Publisher Not Found',
      robots: { index: false, follow: false },
    }
  }

  return {
    title: `${profile.publisher} Publisher Profile`,
    description: `Browse MCP Marketplace listings published by ${profile.publisher}.`,
    alternates: {
      canonical: `/publishers/${publisherSlug}`,
    },
    openGraph: {
      title: `${profile.publisher} on MCP Marketplace`,
      description: `Browse MCP server listings published by ${profile.publisher}.`,
      url: `/publishers/${publisherSlug}`,
      type: 'profile',
    },
  }
}

export default async function PublisherProfilePage({ params }: PageProps) {
  const { publisherSlug } = await params
  const profile = await loadPublisherProfile(publisherSlug)

  if (!profile) {
    notFound()
  }

  const profileUrl = toAbsoluteUrl(`/publishers/${publisherSlug}`)
  const jsonLd = [
    createBreadcrumbJsonLd([
      { name: 'Home', path: '/' },
      { name: 'Marketplace', path: '/marketplace' },
      { name: profile.publisher, path: `/publishers/${publisherSlug}` },
    ]),
    {
      '@context': 'https://schema.org',
      '@type': 'ProfilePage',
      url: profileUrl,
      mainEntity: {
        '@type': 'Organization',
        name: profile.publisher,
        description: `${profile.publisher} publishes MCP servers on MCP Marketplace.`,
      },
    },
    {
      '@context': 'https://schema.org',
      '@type': 'ItemList',
      itemListElement: profile.publisherServers.map((server, index) => ({
        '@type': 'ListItem',
        position: index + 1,
        name: server.name,
        url: toAbsoluteUrl(`/marketplace/${server.slug}`),
      })),
    },
  ]

  return (
    <main className="min-h-screen bg-background px-4 py-16 sm:px-6 lg:px-8">
      <JsonLd data={jsonLd} />
      <div className="mx-auto flex max-w-5xl flex-col gap-8">
        <div className="space-y-4">
          <Badge variant="outline">Publisher Profile</Badge>
          <Text variant="h1" className="uppercase sm:text-5xl">{profile.publisher}</Text>
          <Text variant="h6" className="max-w-3xl text-muted-foreground">
            This publisher profile groups public MCP Marketplace listings under the visible author
            identity used on published server pages.
          </Text>
        </div>

        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Answer Capsule</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            {profile.publisher} currently publishes {profile.publisherServers.length} public MCP
            server{profile.publisherServers.length === 1 ? '' : 's'} across {profile.categories.length}{' '}
            categor{profile.categories.length === 1 ? 'y' : 'ies'} on MCP Marketplace.
          </Text>
        </Card>

        <div className="grid gap-4 md:grid-cols-2">
          {profile.publisherServers.map(server => (
            <Link key={server.id} href={`/marketplace/${server.slug}`}>
              <Card className="border-2 border-foreground p-6 transition-all hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-[6px_6px_0px_hsl(var(--shadow-color))]">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <Text variant="h6" className="uppercase">{server.name}</Text>
                    <Text variant="small" className="mt-2 text-muted-foreground">{server.description}</Text>
                  </div>
                  {server.verified ? <Badge>Verified</Badge> : null}
                </div>
              </Card>
            </Link>
          ))}
        </div>
      </div>
    </main>
  )
}
