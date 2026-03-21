import { getSiteUrl } from './site'

const INDEXNOW_ENDPOINT = 'https://api.indexnow.org/indexnow'

export function getIndexNowKey(): string | null {
  const key = process.env.INDEXNOW_KEY?.trim()
  return key ? key : null
}

export function getIndexNowKeyLocation(): string | null {
  const key = getIndexNowKey()
  if (!key) {
    return null
  }
  return `${getSiteUrl()}/indexnow-key`
}

export function buildIndexNowPayload(urlList: string[]) {
  const host = new URL(getSiteUrl()).host
  const key = getIndexNowKey()
  const keyLocation = getIndexNowKeyLocation()
  if (!key || !keyLocation) {
    return null
  }

  return {
    host,
    key,
    keyLocation,
    urlList,
  }
}

export async function submitIndexNow(urlList: string[]) {
  const payload = buildIndexNowPayload(urlList)
  if (!payload) {
    return {
      ok: false,
      status: 500,
      error: 'INDEXNOW_KEY is not configured',
    }
  }

  const response = await fetch(INDEXNOW_ENDPOINT, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json; charset=utf-8',
    },
    body: JSON.stringify(payload),
  })

  return {
    ok: response.ok,
    status: response.status,
    error: response.ok ? null : await response.text(),
  }
}
