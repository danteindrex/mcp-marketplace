import type { Metadata } from 'next'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import { JsonLd } from '@/components/json-ld'
import { Text } from '@/components/retroui/text'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { fetchPublicServers } from '@/lib/public-marketplace'
import { createBreadcrumbJsonLd, toAbsoluteUrl } from '@/lib/seo'
import { slugifyText } from '@/lib/slugs'

interface PageProps {
  params: Promise<{ authorSlug: string }>
}

async function loadAuthorProfile(authorSlug: string) {
  const servers = await fetchPublicServers()
  const authorServers = servers.filter(server => slugifyText(server.author) === authorSlug)
  if (authorServers.length === 0) {
    return null
  }

  const author = authorServers[0].author
  const categories = Array.from(new Set(authorServers.map(server => server.category)))
  return { author, authorServers, categories }
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { authorSlug } = await params
  const profile = await loadAuthorProfile(authorSlug)

  if (!profile) {
    return {
      title: 'Author Not Found',
      robots: { index: false, follow: false },
    }
  }

  return {
    title: `${profile.author} Author Profile`,
    description: `Browse published MCP servers by ${profile.author} on MCP Marketplace.`,
    alternates: {
      canonical: `/authors/${authorSlug}`,
    },
    openGraph: {
      title: `${profile.author} on MCP Marketplace`,
      description: `Browse published MCP servers by ${profile.author}.`,
      url: `/authors/${authorSlug}`,
      type: 'profile',
    },
  }
}

export default async function AuthorProfilePage({ params }: PageProps) {
  const { authorSlug } = await params
  const profile = await loadAuthorProfile(authorSlug)

  if (!profile) {
    notFound()
  }

  const profileUrl = toAbsoluteUrl(`/authors/${authorSlug}`)
  const jsonLd = [
    createBreadcrumbJsonLd([
      { name: 'Home', path: '/' },
      { name: 'Marketplace', path: '/marketplace' },
      { name: profile.author, path: `/authors/${authorSlug}` },
    ]),
    {
      '@context': 'https://schema.org',
      '@type': 'ProfilePage',
      url: profileUrl,
      mainEntity: {
        '@type': 'Person',
        name: profile.author,
        description: `${profile.author} publishes MCP servers on MCP Marketplace.`,
      },
    },
    {
      '@context': 'https://schema.org',
      '@type': 'ItemList',
      itemListElement: profile.authorServers.map((server, index) => ({
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
          <Badge variant="outline">Author Profile</Badge>
          <Text variant="h1" className="uppercase sm:text-5xl">{profile.author}</Text>
          <Text variant="h6" className="max-w-3xl text-muted-foreground">
            {profile.author} is represented here through published MCP Marketplace listings. This
            profile aggregates their public server catalog, categories, and listing history.
          </Text>
        </div>

        <Card className="border-2 border-foreground p-6">
          <Text variant="h4" className="uppercase">Answer Capsule</Text>
          <Text variant="small" className="mt-3 leading-7 text-muted-foreground">
            {profile.author} currently has {profile.authorServers.length} public MCP server
            {profile.authorServers.length === 1 ? '' : 's'} listed on MCP Marketplace across{' '}
            {profile.categories.length} categor{profile.categories.length === 1 ? 'y' : 'ies'}.
          </Text>
        </Card>

        <div className="grid gap-4 md:grid-cols-3">
          <Card className="border-2 border-foreground p-6">
            <Text variant="caption" className="uppercase text-muted-foreground">Listings</Text>
            <Text variant="h3" className="mt-2">{profile.authorServers.length}</Text>
          </Card>
          <Card className="border-2 border-foreground p-6">
            <Text variant="caption" className="uppercase text-muted-foreground">Categories</Text>
            <Text variant="h3" className="mt-2">{profile.categories.length}</Text>
          </Card>
          <Card className="border-2 border-foreground p-6">
            <Text variant="caption" className="uppercase text-muted-foreground">Latest Update</Text>
            <Text variant="h6" className="mt-2">
              {new Date(
                profile.authorServers
                  .map(server => server.updatedAt)
                  .sort((a, b) => new Date(b).getTime() - new Date(a).getTime())[0]
              ).toLocaleDateString()}
            </Text>
          </Card>
        </div>

        <section className="space-y-4">
          <Text variant="h4" className="uppercase">Published Servers</Text>
          <div className="grid gap-4 md:grid-cols-2">
            {profile.authorServers.map(server => (
              <Link key={server.id} href={`/marketplace/${server.slug}`}>
                <Card className="border-2 border-foreground p-6 transition-all hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-[6px_6px_0px_hsl(var(--shadow-color))]">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <Text variant="h6" className="uppercase">{server.name}</Text>
                      <Text variant="small" className="mt-2 text-muted-foreground">{server.description}</Text>
                    </div>
                    {server.verified ? <Badge>Verified</Badge> : null}
                  </div>
                  <div className="mt-4 flex flex-wrap gap-2">
                    <Badge variant="secondary">{server.category}</Badge>
                    <Badge variant="outline">v{server.version}</Badge>
                  </div>
                </Card>
              </Link>
            ))}
          </div>
        </section>
      </div>
    </main>
  )
}
