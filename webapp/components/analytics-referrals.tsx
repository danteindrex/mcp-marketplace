'use client'

import { useEffect } from 'react'
import { usePathname } from 'next/navigation'
import { track } from '@vercel/analytics/react'

function classifyReferrer(referrer: string): string | null {
  if (!referrer) return null

  const value = referrer.toLowerCase()
  if (value.includes('chat.openai.com') || value.includes('chatgpt.com')) return 'openai'
  if (value.includes('perplexity.ai')) return 'perplexity'
  if (value.includes('claude.ai')) return 'anthropic'
  if (value.includes('gemini.google.com') || value.includes('bard.google.com')) return 'gemini'
  if (value.includes('google.com')) return 'google'
  return null
}

export function AnalyticsReferrals() {
  const pathname = usePathname()

  useEffect(() => {
    if (typeof window === 'undefined') return

    const referrer = document.referrer || ''
    const source = classifyReferrer(referrer)
    if (!source) return

    const key = `ai-referral:${source}:${pathname}`
    if (window.sessionStorage.getItem(key)) return

    window.sessionStorage.setItem(key, '1')
    track('AI Referral Landing', {
      source,
      landing_path: pathname,
      referrer_host: (() => {
        try {
          return new URL(referrer).host
        } catch {
          return 'unknown'
        }
      })(),
    })
  }, [pathname])

  return null
}
