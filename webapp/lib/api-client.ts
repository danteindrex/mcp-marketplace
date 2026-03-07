import { apiGet, apiPost, apiPut, getActiveRole, type AppRole } from './api'

export interface Server {
  id: string
  tenantId: string
  author: string
  name: string
  slug: string
  description: string
  category: string
  version: string
  dockerImage: string
  canonicalResourceUri: string
  requiredScopes: string[]
  pricingType: string
  pricingAmount: number
  verified: boolean
  featured: boolean
  installCount: number
  rating: number
  status: string
  supportsLocal: boolean
  supportsCloud: boolean
  updatedAt: string
  createdAt: string
}

export interface MarketplaceInstallMetadata {
  oneClick: boolean
  hubStrategy: string
  clients: string[]
  installEndpoint?: string
  supportsCommands?: boolean
}

export interface InstallAction {
  client: string
  label: string
  launchUrl?: string
  openUrl?: string
  command?: string
  fallbackCopy?: string
  description?: string
  requiresLocalExec?: boolean
}

export interface InstallSession {
  server: Server
  hub: {
    id: string
    tenantId: string
    userId: string
    hubUrl: string
    catalogVersion: number
  }
  resource: string
  connection: {
    id: string
    serverId?: string
    serverName?: string
    client: string
    status: string
    tokenExpiresAt?: string
  }
  install: {
    selected: InstallAction
    actions: InstallAction[]
  }
}

export interface Connection {
  id: string
  userId: string
  serverId: string
  serverName: string
  status: 'active' | 'expired' | 'revoked' | 'pending'
  scopes: string[]
  tokenExpiresAt?: Date
  createdAt: Date
  lastUsed?: Date
}

export interface Billing {
  id: string
  userId: string
  plan: string
  monthlySpend: number
  dailySpend?: number
  currentBalance: number
  nextBillingDate: Date
  paymentMethod?: string
  allowedMethods?: string[]
  caps?: {
    perCallCapUsdc?: number
    dailySpendCapUsdc?: number
    monthlySpendCapUsdc?: number
    minimumBalanceUsdc?: number
  }
  wallet?: {
    balanceUsdc?: number
    minimumBalanceUsdc?: number
    hardStopOnLowFunds?: boolean
    fundingMethod?: string
    walletAddress?: string
    lastTopUpAt?: string
  }
  status: string
}

export interface PaymentMethodCatalogItem {
  id: string
  displayName: string
  enabled: boolean
  configured: boolean
  integration: string
  docs?: string
  notes?: string
  network?: string
  asset?: string
}

export interface BuyerPaymentControls {
  policy: {
    monthlySpendCapUsdc: number
    dailySpendCapUsdc: number
    perCallCapUsdc: number
    allowedMethods: string[]
    siwxWallet?: string
    walletBalanceUsdc?: number
    minimumBalanceUsdc?: number
    hardStopOnLowFunds?: boolean
    autoTopUpEnabled?: boolean
    autoTopUpAmountUsd?: number
    autoTopUpTriggerUsd?: number
    fundingMethod?: string
    walletAddress?: string
    lastTopUpAt?: string
  }
  methods: PaymentMethodCatalogItem[]
  dailySpendUsdc: number
  monthlySpendUsdc: number
  dailyRemaining: number
  monthlyRemaining: number
  facilitatorMode: string
  facilitatorTarget?: string
  wallet?: {
    balanceUsdc: number
    minimumBalanceUsdc: number
    hardStopOnLowFunds: boolean
    autoTopUpEnabled: boolean
    autoTopUpAmountUsd: number
    autoTopUpTriggerUsd: number
    fundingMethod: string
    walletAddress?: string
    lastTopUpAt?: string
  }
  topups?: WalletTopUp[]
}

export interface WalletTopUp {
  id: string
  provider: string
  providerSessionId?: string
  status: string
  sourceCurrency?: string
  sourceAmount?: number
  destinationAsset?: string
  destinationNetwork?: string
  destinationAmount?: number
  paymentMethod?: string
  walletAddress?: string
  hostedUrl?: string
  failureReason?: string
  createdAt: string
  updatedAt?: string
  fulfilledAt?: string
}

export interface SecurityEvent {
  id: string
  type: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  description: string
  actor: string
  targetServer?: string
  timestamp: Date
  resolved: boolean
}

export interface UserProfileSettings {
  id: string
  tenantId: string
  email: string
  name: string
  phone?: string
  avatarUrl?: string
  locale?: string
  timezone?: string
  role: AppRole
}

export interface UserPreferencesSettings {
  theme: 'light' | 'dark' | 'system'
  language: string
  timezone: string
  defaultLanding: string
  compactMode: boolean
}

export interface UserNotificationSettings {
  productUpdates: boolean
  securityAlerts: boolean
  billingAlerts: boolean
  marketingEmail: boolean
  weeklyDigest: boolean
}

export interface MFAStatus {
  mfaEnabled: boolean
  method: '' | 'totp'
}

export interface TenantRecord {
  id: string
  name: string
  slug: string
  ownerUserId: string
  planTier: string
  status: string
  createdAt: string
}

export interface CurrentUser {
  id: string
  tenantId: string
  email: string
  name: string
  role: AppRole
  mfaEnabled?: boolean
}

function requireActiveRole(): AppRole {
  const role = getActiveRole()
  if (!role) throw new Error('Not authenticated')
  return role
}

export async function fetchServers(): Promise<Server[]> {
  const data = await apiGet<{ items: Server[] }>('/v1/marketplace/servers')
  return data.items
}

export async function fetchCurrentUser(): Promise<CurrentUser | null> {
  try {
    return await apiGet<CurrentUser>('/v1/me')
  } catch {
    return null
  }
}

export async function fetchServerBySlug(slug: string): Promise<Server | null> {
  try {
    const data = await apiGet<{ server: Server }>(`/v1/marketplace/servers/${slug}`)
    return data.server
  } catch {
    return null
  }
}

export async function fetchServerDetailBySlug(
  slug: string,
): Promise<{ server: Server; install?: MarketplaceInstallMetadata } | null> {
  try {
    return await apiGet<{ server: Server; install?: MarketplaceInstallMetadata }>(
      `/v1/marketplace/servers/${slug}`,
    )
  } catch {
    return null
  }
}

export async function fetchFeaturedServers(): Promise<Server[]> {
  const all = await fetchServers()
  return all.filter(s => s.featured)
}

export async function fetchConnections(): Promise<Connection[]> {
  const data = await apiGet<{ items: any[] }>('/v1/buyer/connections', 'buyer')
  return data.items.map(c => ({
    id: c.id,
    userId: c.userId || 'buyer',
    serverId: c.serverId || c.id,
    serverName: c.serverName || c.client || 'Connected Server',
    status: c.status || 'active',
    scopes: c.grantedScopes || [],
    tokenExpiresAt: c.tokenExpiresAt ? new Date(c.tokenExpiresAt) : undefined,
    createdAt: c.createdAt ? new Date(c.createdAt) : new Date(),
    lastUsed: c.lastUsedAt ? new Date(c.lastUsedAt) : undefined,
  }))
}

export async function createConnection(payload: { client: string; resource: string; grantedScopes: string[] }) {
  return apiPost('/v1/buyer/connections', payload, 'buyer')
}

export async function installMarketplaceServer(
  slug: string,
  payload: { client: string; grantedScopes?: string[] },
): Promise<InstallSession> {
  return apiPost(`/v1/marketplace/servers/${slug}/install`, payload, 'buyer')
}

export async function fetchBuyerHub() {
  return apiGet<{ hub: any; routes: any[]; strategy: any }>('/v1/buyer/hub', 'buyer')
}

export async function fetchLocalAgents() {
  return apiGet<{ items: any[]; count: number }>('/v1/buyer/local-agents', 'buyer')
}

export async function rotateToken(connectionId: string): Promise<void> {
  await apiPost(`/v1/buyer/connections/${connectionId}/rotate`, {}, 'buyer')
}

export async function revokeConnection(connectionId: string): Promise<void> {
  await apiPost(`/v1/buyer/connections/${connectionId}/revoke`, {}, 'buyer')
}

export async function fetchBilling(): Promise<Billing> {
  const data = await apiGet<any>('/v1/buyer/billing', 'buyer')
  return {
    id: data.id,
    userId: data.userId,
    plan: data.plan,
    monthlySpend: data.monthlySpend,
    dailySpend: data.dailySpend,
    currentBalance: data.currentBalance,
    nextBillingDate: new Date(data.nextBillingDate),
    paymentMethod: data.paymentMethod,
    allowedMethods: data.allowedMethods || [],
    caps: data.caps,
    wallet: data.wallet,
    status: data.status,
  }
}

export async function fetchInvoices(): Promise<Array<{ id: string; date: Date; amount: number; status: string }>> {
  const data = await apiGet<{ items: any[] }>('/v1/buyer/invoices', 'buyer')
  return data.items.map(inv => ({
    id: inv.id,
    date: new Date(inv.date),
    amount: Number(inv.amount || 0),
    status: inv.status || 'paid',
  }))
}

export async function fetchBuyerPaymentControls(): Promise<BuyerPaymentControls> {
  return apiGet('/v1/buyer/payments/controls', 'buyer')
}

export async function updateBuyerPaymentControls(payload: {
  monthlySpendCapUsdc?: number
  dailySpendCapUsdc?: number
  perCallCapUsdc?: number
  allowedMethods?: string[]
  siwxWallet?: string
  minimumBalanceUsdc?: number
  hardStopOnLowFunds?: boolean
  autoTopUpEnabled?: boolean
  autoTopUpAmountUsd?: number
  autoTopUpTriggerUsd?: number
  fundingMethod?: string
  walletAddress?: string
}) {
  return apiPut('/v1/buyer/payments/controls', payload, 'buyer')
}

export async function fetchBuyerWalletTopUps(): Promise<{
  items: WalletTopUp[]
  count: number
  minimumTopUpUsd: number
  defaultTopUpUsd: number
  stripeConfigured: boolean
}> {
  return apiGet('/v1/buyer/payments/topups', 'buyer')
}

export async function createStripeTopUpSession(payload: {
  amountUsd: number
  walletAddress?: string
  paymentMethod?: string
}) {
  return apiPost('/v1/buyer/payments/topups/stripe/session', payload, 'buyer')
}

export async function fetchSecurityEvents(severity?: string): Promise<SecurityEvent[]> {
  const data = await apiGet<{ items: any[] }>('/v1/admin/security-events', 'admin')
  const events = data.items.map(e => ({
    id: e.id,
    type: e.type,
    severity: e.severity,
    description: e.description,
    actor: e.actor,
    targetServer: e.targetId,
    timestamp: new Date(e.createdAt),
    resolved: Boolean(e.resolved),
  })) as SecurityEvent[]
  return severity ? events.filter(e => e.severity === severity) : events
}

export async function fetchAuditLogs(): Promise<Array<{ id: string; timestamp: Date; actor: string; action: string; target: string; result: 'success' | 'failure'; details: string }>> {
  const data = await apiGet<{ items: any[] }>('/v1/admin/audit-logs', 'admin')
  return data.items.map(log => ({
    id: log.id,
    timestamp: new Date(log.createdAt),
    actor: log.actorId,
    action: log.action,
    target: log.targetId,
    result: log.outcome === 'success' ? 'success' : 'failure',
    details: log.targetType,
  }))
}

export async function fetchTenants(): Promise<TenantRecord[]> {
  const data = await apiGet<{ items: TenantRecord[] }>('/v1/admin/tenants', 'admin')
  return data.items
}

export async function fetchClientCompatibility(): Promise<Array<{ client: string; supportsDCR: boolean; supportsCIMD: boolean; supportsInteractive: boolean; notes: string }>> {
  const data = await apiGet<{ items: any[] }>('/v1/admin/client-compatibility', 'admin')
  return data.items
}

export async function fetchAdminPaymentsOverview() {
  return apiGet('/v1/admin/payments/overview', 'admin')
}

export async function fetchMerchantServers(): Promise<Server[]> {
  const data = await apiGet<{ items: Server[] }>('/v1/merchant/servers', 'merchant')
  return data.items
}

export async function fetchMerchantRevenue(): Promise<{ totalRevenue: number; totalCustomers: number; servers: Array<{ id: string; name: string; revenue: number; customers: number; trend: string }>; trend: Array<{ month: string; revenue: number; subscriptions: number; perCall: number }> }> {
  return apiGet('/v1/merchant/revenue', 'merchant')
}

export async function fetchMerchantPaymentsOverview() {
  return apiGet('/v1/merchant/payments/overview', 'merchant')
}

export async function fetchMerchantServerPaymentConfig(id: string) {
  return apiGet(`/v1/merchant/servers/${id}/payments/config`, 'merchant')
}

export async function updateMerchantServerPaymentConfig(id: string, payload: {
  paymentMethods?: string[]
  paymentAddress?: string
  perCallCapUsdc?: number
  dailyCapUsdc?: number
  monthlyCapUsdc?: number
}) {
  return apiPut(`/v1/merchant/servers/${id}/payments/config`, payload, 'merchant')
}

export async function fetchServerAuth(id: string) {
  return apiGet(`/v1/merchant/servers/${id}/auth`, 'merchant')
}

export async function fetchServerPricing(id: string) {
  return apiGet(`/v1/merchant/servers/${id}/pricing`, 'merchant')
}

export async function fetchServerObservability(id: string) {
  return apiGet(`/v1/merchant/servers/${id}/observability`, 'merchant')
}

export async function fetchServerDeployments(id: string) {
  return apiGet<{ items: any[] }>(`/v1/merchant/servers/${id}/deployments`, 'merchant')
}

export async function fetchServerBuilder(id: string) {
  return apiGet(`/v1/merchant/servers/${id}/builder`, 'merchant')
}

export async function fetchUserProfileSettings(): Promise<UserProfileSettings> {
  const role = requireActiveRole()
  const data = await apiGet<{ profile: UserProfileSettings }>('/v1/settings/profile', role)
  return data.profile
}

export async function updateUserProfileSettings(payload: {
  name: string
  email: string
  phone?: string
  avatarUrl?: string
  locale?: string
  timezone?: string
}): Promise<UserProfileSettings> {
  const role = requireActiveRole()
  const data = await apiPut<{ profile: UserProfileSettings }>('/v1/settings/profile', payload, role)
  return data.profile
}

export async function changeUserPassword(payload: {
  currentPassword: string
  newPassword: string
  confirmPassword?: string
}): Promise<void> {
  const role = requireActiveRole()
  await apiPut('/v1/settings/security/password', payload, role)
}

export async function fetchMFAStatus(): Promise<MFAStatus> {
  const role = requireActiveRole()
  return apiGet<MFAStatus>('/v1/settings/security/mfa', role)
}

export async function setupMFATOTP(): Promise<{
  secret: string
  otpauthURL: string
  issuer: string
  accountName: string
}> {
  const role = requireActiveRole()
  return apiPost('/v1/settings/security/mfa/totp/setup', {}, role)
}

export async function verifyMFATOTP(code: string): Promise<MFAStatus> {
  const role = requireActiveRole()
  return apiPost('/v1/settings/security/mfa/totp/verify', { code }, role)
}

export async function disableMFATOTP(payload: {
  currentPassword: string
  code: string
}): Promise<MFAStatus> {
  const role = requireActiveRole()
  return apiPost('/v1/settings/security/mfa/totp/disable', payload, role)
}

export async function fetchUserPreferencesSettings(): Promise<UserPreferencesSettings> {
  const role = requireActiveRole()
  const data = await apiGet<{ preferences: UserPreferencesSettings }>('/v1/settings/preferences', role)
  return data.preferences
}

export async function updateUserPreferencesSettings(
  payload: Partial<UserPreferencesSettings>,
): Promise<UserPreferencesSettings> {
  const role = requireActiveRole()
  const data = await apiPut<{ preferences: UserPreferencesSettings }>('/v1/settings/preferences', payload, role)
  return data.preferences
}

export async function fetchUserNotificationSettings(): Promise<UserNotificationSettings> {
  const role = requireActiveRole()
  const data = await apiGet<{ notifications: UserNotificationSettings }>('/v1/settings/notifications', role)
  return data.notifications
}

export async function updateUserNotificationSettings(
  payload: Partial<UserNotificationSettings>,
): Promise<UserNotificationSettings> {
  const role = requireActiveRole()
  const data = await apiPut<{ notifications: UserNotificationSettings }>('/v1/settings/notifications', payload, role)
  return data.notifications
}
