import type { MetadataRoute } from 'next'
import { fetchPublicServers } from '@/lib/public-marketplace'
import { getSiteUrl } from '@/lib/site'

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const siteUrl = getSiteUrl()
  const servers = await fetchPublicServers()
  const authors = Array.from(new Set(servers.map(server => server.author.trim()).filter(Boolean)))
    .map(author => author.toLowerCase().trim().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, ''))

  const staticRoutes: MetadataRoute.Sitemap = [
    {
      url: `${siteUrl}/`,
      lastModified: new Date(),
      changeFrequency: 'daily',
      priority: 1,
    },
    {
      url: `${siteUrl}/marketplace`,
      lastModified: new Date(),
      changeFrequency: 'daily',
      priority: 0.9,
    },
    {
      url: `${siteUrl}/pricing`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.6,
    },
    {
      url: `${siteUrl}/guides/mcp-server-installation`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.8,
    },
    {
      url: `${siteUrl}/guides`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.8,
    },
    {
      url: `${siteUrl}/guides/install-mcp-in-codex`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.7,
    },
    {
      url: `${siteUrl}/guides/install-mcp-in-claude-desktop`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.7,
    },
    {
      url: `${siteUrl}/guides/install-mcp-in-cursor`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.7,
    },
    {
      url: `${siteUrl}/guides/install-mcp-in-vs-code`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.7,
    },
    {
      url: `${siteUrl}/guides/install-mcp-in-chatgpt`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.7,
    },
    {
      url: `${siteUrl}/guides/what-is-mcp`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.7,
    },
    {
      url: `${siteUrl}/guides/mcp-cimd-explained`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.7,
    },
    {
      url: `${siteUrl}/guides/mcp-dcr-explained`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.7,
    },
    {
      url: `${siteUrl}/guides/mcp-oauth-explained`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.7,
    },
    {
      url: `${siteUrl}/guides/x402-pricing-explained`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.7,
    },
    {
      url: `${siteUrl}/guides/hosted-vs-local-mcp-servers`,
      lastModified: new Date(),
      changeFrequency: 'weekly',
      priority: 0.7,
    },
  ]

  const marketplaceRoutes: MetadataRoute.Sitemap = servers.map(server => ({
    url: `${siteUrl}/marketplace/${server.slug}`,
    lastModified: server.updatedAt ? new Date(server.updatedAt) : new Date(),
    changeFrequency: 'daily',
    priority: 0.8,
  }))

  const authorRoutes: MetadataRoute.Sitemap = authors.map(authorSlug => ({
    url: `${siteUrl}/authors/${authorSlug}`,
    lastModified: new Date(),
    changeFrequency: 'weekly',
    priority: 0.6,
  }))

  const publisherRoutes: MetadataRoute.Sitemap = authors.map(publisherSlug => ({
    url: `${siteUrl}/publishers/${publisherSlug}`,
    lastModified: new Date(),
    changeFrequency: 'weekly',
    priority: 0.6,
  }))

  return [...staticRoutes, ...marketplaceRoutes, ...authorRoutes, ...publisherRoutes]
}
