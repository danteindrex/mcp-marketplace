'use client'

import { useEffect, useMemo, useState } from 'react'
import Script from 'next/script'

declare global {
  interface Window {
    StripeOnramp?: (publishableKey: string) => {
      createSession: (config: { clientSecret: string }) => {
        addEventListener?: (event: string, cb: (ev: any) => void) => void
        removeEventListener?: (event: string, cb: (ev: any) => void) => void
        mount?: (target: string | HTMLElement) => void
        unmount?: () => void
        destroy?: () => void
      }
    }
  }
}

export function StripeOnrampWidget(props: {
  publishableKey: string
  clientSecret: string
  onSessionUpdate?: (session: any) => void
}) {
  const { publishableKey, clientSecret, onSessionUpdate } = props
  const [stripeLoaded, setStripeLoaded] = useState(false)
  const [onrampLoaded, setOnrampLoaded] = useState(false)
  const [error, setError] = useState('')
  const containerId = useMemo(
    () => `stripe-onramp-${Math.random().toString(36).slice(2)}`,
    [],
  )

  useEffect(() => {
    if (!stripeLoaded || !onrampLoaded || !publishableKey || !clientSecret) {
      return
    }
    if (typeof window === 'undefined' || typeof window.StripeOnramp !== 'function') {
      setError('Stripe Onramp SDK failed to load.')
      return
    }

    let session: any
    const onUpdated = (event: any) => {
      const payloadSession = event?.payload?.session || event?.session || null
      if (payloadSession && onSessionUpdate) {
        onSessionUpdate(payloadSession)
      }
    }

    try {
      const sdk = window.StripeOnramp(publishableKey)
      session = sdk.createSession({ clientSecret })
      if (session?.addEventListener) {
        session.addEventListener('onramp_session_updated', onUpdated)
      }
      if (session?.mount) {
        session.mount(`#${containerId}`)
      } else {
        setError('Stripe Onramp session mount API unavailable.')
      }
    } catch (e: any) {
      setError(e?.message || 'Unable to initialize Stripe Onramp.')
    }

    return () => {
      try {
        if (session?.removeEventListener) {
          session.removeEventListener('onramp_session_updated', onUpdated)
        }
        if (session?.unmount) {
          session.unmount()
        }
        if (session?.destroy) {
          session.destroy()
        }
      } catch {
        // no-op cleanup fallback for SDK shape differences
      }
    }
  }, [stripeLoaded, onrampLoaded, publishableKey, clientSecret, onSessionUpdate, containerId])

  return (
    <div className="space-y-2">
      <Script
        src="https://js.stripe.com/clover/stripe.js"
        strategy="afterInteractive"
        onLoad={() => setStripeLoaded(true)}
      />
      <Script
        src="https://js.stripe.com/crypto-onramp-outer.js"
        strategy="afterInteractive"
        onLoad={() => setOnrampLoaded(true)}
      />
      {error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : (
        <p className="text-xs text-muted-foreground">Loading secure Stripe Onramp...</p>
      )}
      <div id={containerId} className="min-h-[640px] rounded-md border border-border bg-background" />
    </div>
  )
}
