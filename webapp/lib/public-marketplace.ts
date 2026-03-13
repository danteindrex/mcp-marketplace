import { getServerApiBase } from '@/lib/api-base'
import type { Server } from '@/lib/api-client'

export async function fetchPublicServers(): Promise<Server[]> {
  try {
    const res = await fetch(`${getServerApiBase()}/v1/marketplace/servers`, {
      next: { revalidate: 300 },
    })
    if (!res.ok) {
      return []
    }
    const data = (await res.json()) as { items?: Server[] }
    return Array.isArray(data.items) ? data.items : []
  } catch {
    return []
  }
}

export async function fetchPublicServerBySlug(slug: string): Promise<Server | null> {
  try {
    const res = await fetch(`${getServerApiBase()}/v1/marketplace/servers/${slug}`, {
      next: { revalidate: 300 },
    })
    if (!res.ok) {
      return null
    }
    const data = (await res.json()) as { server?: Server }
    return data.server ?? null
  } catch {
    return null
  }
}
