import type { PaymentMethodCatalogItem } from './api-client'

export const INSTALL_PAYMENT_METHOD_IDS = ['wallet_balance', 'x402_wallet'] as const

const PAYMENT_METHOD_COPY: Record<string, { title: string; description: string }> = {
  wallet_balance: {
    title: 'Prepaid wallet balance',
    description: 'Auto-settle paid installs from your buyer wallet when balance and policy allow it.',
  },
  x402_wallet: {
    title: 'Marketplace wallet',
    description: 'Use the managed marketplace wallet to sign and settle x402 payments automatically.',
  },
  stripe_onramp: {
    title: 'Stripe onramp top-up',
    description: 'Fund the buyer wallet with a one-time Stripe onramp payment before retrying paid installs.',
  },
}

export type PaymentMethodStatus = 'ready' | 'needs_setup' | 'blocked' | 'disabled'

export function getPaymentMethodTitle(method: Pick<PaymentMethodCatalogItem, 'id' | 'displayName'>) {
  return PAYMENT_METHOD_COPY[method.id]?.title || method.displayName
}

export function getPaymentMethodDescription(method: Pick<PaymentMethodCatalogItem, 'id' | 'notes'>) {
  return method.notes || PAYMENT_METHOD_COPY[method.id]?.description || 'Payment method status is managed by backend controls.'
}

export function getPaymentMethodStatus(
  method: Pick<PaymentMethodCatalogItem, 'id' | 'enabled' | 'configured'>,
  allowedMethods: string[],
): {
  tone: 'default' | 'secondary' | 'outline' | 'destructive'
  label: string
  description: string
  selectable: boolean
  status: PaymentMethodStatus
} {
  if (!method.enabled) {
    return {
      tone: 'destructive',
      label: 'Disabled',
      description: 'Backend controls currently disable this method.',
      selectable: false,
      status: 'disabled',
    }
  }

  if (!method.configured) {
    return {
      tone: 'secondary',
      label: 'Needs setup',
      description: 'The backend exposes this method, but it is not configured yet.',
      selectable: false,
      status: 'needs_setup',
    }
  }

  if (!allowedMethods.includes(method.id)) {
    return {
      tone: 'outline',
      label: 'Blocked by policy',
      description: 'Available in the backend, but not currently allowed for this buyer.',
      selectable: false,
      status: 'blocked',
    }
  }

  return {
    tone: 'default',
    label: 'Ready',
    description: 'Enabled, configured, and allowed for buyer payment flows.',
    selectable: true,
    status: 'ready',
  }
}

export function isInstallPaymentMethod(methodId: string) {
  return INSTALL_PAYMENT_METHOD_IDS.includes(methodId as (typeof INSTALL_PAYMENT_METHOD_IDS)[number])
}
