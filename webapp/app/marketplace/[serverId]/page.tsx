import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { JsonLd } from '@/components/json-ld'
import { fetchPublicServerBySlug } from '@/lib/public-marketplace'
import { createBreadcrumbJsonLd, toAbsoluteUrl } from '@/lib/seo'
import { ServerDetailClientPage } from './client-page'

interface PageProps {
  params: Promise<{ serverId: string }>
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { serverId } = await params
  const server = await fetchPublicServerBySlug(serverId)

  if (!server) {
    return {
      title: 'Server Not Found',
      robots: {
        index: false,
        follow: false,
      },
    }
  }

  return {
    title: server.name,
    description: server.description,
    alternates: {
      canonical: `/marketplace/${server.slug}`,
    },
    openGraph: {
      title: server.name,
      description: server.description,
      url: `/marketplace/${server.slug}`,
      type: 'website',
    },
  }
}

export default async function ServerDetailPage({ params }: PageProps) {
  const { serverId } = await params
  const server = await fetchPublicServerBySlug(serverId)

  if (!server) {
    notFound()
  }

  const canonicalUrl = toAbsoluteUrl(`/marketplace/${server.slug}`)
  const jsonLd = [
    createBreadcrumbJsonLd([
      { name: 'Home', path: '/' },
      { name: 'Marketplace', path: '/marketplace' },
      { name: server.name, path: `/marketplace/${server.slug}` },
    ]),
    {
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      name: server.name,
      description: server.description,
      applicationCategory: 'DeveloperApplication',
      applicationSubCategory: server.category,
      operatingSystem: 'Web',
      creator: {
        '@type': 'Person',
        name: server.author,
      },
      publisher: {
        '@type': 'Organization',
        name: 'MCP Marketplace',
      },
      featureList: server.requiredScopes,
      keywords: [server.category, server.pricingType, server.supportsCloud ? 'cloud' : null, server.supportsLocal ? 'local' : null]
        .filter((value): value is string => Boolean(value))
        .join(', '),
      isAccessibleForFree: server.pricingType === 'free',
      offers: {
        '@type': 'Offer',
        price: Number(server.pricingAmount || 0),
        priceCurrency: 'USD',
        availability: 'https://schema.org/InStock',
        category: server.pricingType,
      },
      aggregateRating: server.rating > 0
        ? {
            '@type': 'AggregateRating',
            ratingValue: server.rating,
            ratingCount: Math.max(server.installCount, 1),
          }
        : undefined,
      url: canonicalUrl,
    },
    {
      '@context': 'https://schema.org',
      '@type': 'Product',
      name: server.name,
      description: server.description,
      brand: {
        '@type': 'Brand',
        name: server.author,
      },
      category: server.category,
      sku: server.id,
      releaseDate: server.publishedAt || server.createdAt,
      url: canonicalUrl,
      offers: {
        '@type': 'Offer',
        price: Number(server.pricingAmount || 0),
        priceCurrency: 'USD',
        availability: 'https://schema.org/InStock',
      },
      aggregateRating: server.rating > 0
        ? {
            '@type': 'AggregateRating',
            ratingValue: server.rating,
            ratingCount: Math.max(server.installCount, 1),
          }
        : undefined,
    },
  ]

  return (
    <>
      <JsonLd data={jsonLd} />
      <ServerDetailClientPage serverId={serverId} initialServer={server} />
    </>
  )
}
