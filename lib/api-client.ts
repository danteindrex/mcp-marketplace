// API Client - Service layer for MCP Marketplace
// Currently returns mock data, ready for real API integration

import {
  mockServers,
  mockConnections,
  mockBilling,
  mockSecurityEvents,
  getServerById,
  searchServers,
  filterServersByCategory,
  getConnectionsByUserId,
  getBillingByUserId,
  getSecurityEventsBySeverity,
  Server,
  Connection,
  Billing,
  SecurityEvent,
} from './mock-data'

// Simulated API delay for realistic UX
const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

// Server APIs
export async function fetchServers(): Promise<Server[]> {
  await delay(300)
  return mockServers
}

export async function fetchServerById(id: string): Promise<Server | null> {
  await delay(200)
  return getServerById(id) || null
}

export async function fetchFeaturedServers(): Promise<Server[]> {
  await delay(250)
  return mockServers.filter(s => s.featured)
}

export async function fetchServersByCategory(category: string): Promise<Server[]> {
  await delay(300)
  return filterServersByCategory(category)
}

export async function searchServersAPI(query: string): Promise<Server[]> {
  await delay(300)
  return searchServers(query)
}

export async function installServer(serverId: string, userId: string): Promise<Connection> {
  await delay(500)
  const server = getServerById(serverId)
  if (!server) throw new Error('Server not found')

  const newConnection: Connection = {
    id: `conn_${Date.now()}`,
    userId,
    serverId,
    serverName: server.name,
    status: 'active',
    scopes: server.requiredScopes,
    tokenExpiresAt: new Date(Date.now() + 365 * 24 * 60 * 60 * 1000), // 1 year
    createdAt: new Date(),
  }
  return newConnection
}

export async function uninstallServer(connectionId: string): Promise<void> {
  await delay(400)
  // In real implementation, would update database
}

// Connection APIs
export async function fetchConnections(userId: string): Promise<Connection[]> {
  await delay(300)
  return getConnectionsByUserId(userId)
}

export async function rotateToken(connectionId: string): Promise<Connection> {
  await delay(600)
  const conn = mockConnections.find(c => c.id === connectionId)
  if (!conn) throw new Error('Connection not found')

  return {
    ...conn,
    tokenExpiresAt: new Date(Date.now() + 365 * 24 * 60 * 60 * 1000),
  }
}

export async function revokeConnection(connectionId: string): Promise<void> {
  await delay(400)
  // In real implementation, would update database
}

// Billing APIs
export async function fetchBilling(userId: string): Promise<Billing | null> {
  await delay(300)
  return getBillingByUserId(userId) || null
}

export async function fetchInvoices(userId: string): Promise<Array<{
  id: string
  date: Date
  amount: number
  status: string
}>> {
  await delay(400)
  return [
    { id: 'inv_001', date: new Date('2025-02-01'), amount: 245.50, status: 'paid' },
    { id: 'inv_002', date: new Date('2025-01-01'), amount: 245.50, status: 'paid' },
    { id: 'inv_003', date: new Date('2024-12-01'), amount: 189.99, status: 'paid' },
  ]
}

export async function updatePaymentMethod(userId: string, token: string): Promise<void> {
  await delay(500)
  // In real implementation, would call Stripe or payment provider
}

// Merchant APIs
export async function publishServer(serverId: string, merchantId: string): Promise<void> {
  await delay(800)
  // In real implementation, would update database and trigger deployment
}

export async function updateServerMetadata(serverId: string, metadata: Partial<Server>): Promise<Server> {
  await delay(400)
  const server = getServerById(serverId)
  if (!server) throw new Error('Server not found')
  return { ...server, ...metadata }
}

export async function deployServer(
  serverId: string,
  environment: 'development' | 'staging' | 'production'
): Promise<void> {
  await delay(1200)
  // In real implementation, would trigger Docker deployment
}

export async function fetchDeployments(serverId: string): Promise<Array<{
  id: string
  environment: string
  status: string
  timestamp: Date
  version: string
}>> {
  await delay(300)
  return [
    { id: 'd_001', environment: 'production', status: 'active', timestamp: new Date('2025-03-01'), version: '2.1.0' },
    { id: 'd_002', environment: 'staging', status: 'active', timestamp: new Date('2025-02-28'), version: '2.0.9' },
    { id: 'd_003', environment: 'development', status: 'active', timestamp: new Date('2025-02-20'), version: '2.1.0-beta' },
  ]
}

// Observability APIs
export async function fetchMetrics(
  serverId: string,
  range: '24h' | '7d' | '30d'
): Promise<Array<{
  timestamp: Date
  toolName: string
  invocations: number
  avgLatency: number
  errorRate: number
}>> {
  await delay(400)
  const baseDate = new Date()
  return [
    {
      timestamp: new Date(baseDate.getTime() - 3600000),
      toolName: 'query',
      invocations: 245,
      avgLatency: 342,
      errorRate: 0.02,
    },
    {
      timestamp: new Date(baseDate.getTime() - 7200000),
      toolName: 'query',
      invocations: 189,
      avgLatency: 298,
      errorRate: 0.01,
    },
  ]
}

// Admin APIs
export async function fetchSecurityEvents(severity?: string): Promise<SecurityEvent[]> {
  await delay(300)
  if (severity) {
    return getSecurityEventsBySeverity(severity)
  }
  return mockSecurityEvents
}

export async function resolveSecurityEvent(eventId: string): Promise<void> {
  await delay(400)
  // In real implementation, would update database
}

export async function fetchAuditLogs(filters?: {
  actor?: string
  action?: string
  dateRange?: [Date, Date]
}): Promise<Array<{
  id: string
  timestamp: Date
  actor: string
  action: string
  target: string
  result: 'success' | 'failure'
  details: string
}>> {
  await delay(400)
  return [
    {
      id: 'log_001',
      timestamp: new Date('2025-03-03T15:30:00'),
      actor: 'user_001',
      action: 'INSTALL_SERVER',
      target: 'srv_001',
      result: 'success',
      details: 'PostgreSQL Assistant v2.1.0',
    },
    {
      id: 'log_002',
      timestamp: new Date('2025-03-03T14:15:00'),
      actor: 'user_001',
      action: 'REVOKE_CONNECTION',
      target: 'conn_003',
      result: 'success',
      details: 'Manually revoked by user',
    },
  ]
}

export async function fetchTenants(): Promise<Array<{
  id: string
  name: string
  email: string
  status: string
  riskScore: number
  suspensions: number
  createdAt: Date
}>> {
  await delay(400)
  return [
    {
      id: 'tenant_001',
      name: 'Acme Corp',
      email: 'admin@acmecorp.com',
      status: 'active',
      riskScore: 12,
      suspensions: 0,
      createdAt: new Date('2024-06-15'),
    },
    {
      id: 'tenant_002',
      name: 'DataFlow Inc',
      email: 'admin@dataflow.io',
      status: 'active',
      riskScore: 5,
      suspensions: 0,
      createdAt: new Date('2024-03-20'),
    },
  ]
}

export async function fetchClientCompatibility(): Promise<Array<{
  client: string
  codexSupport: boolean
  vsCodeSupport: boolean
  cursorSupport: boolean
  claudeSupport: boolean
  notes: string
}>> {
  await delay(300)
  return [
    {
      client: 'VS Code',
      codexSupport: true,
      vsCodeSupport: true,
      cursorSupport: false,
      claudeSupport: false,
      notes: 'Full support via native MCP extension',
    },
    {
      client: 'Cursor',
      codexSupport: true,
      vsCodeSupport: false,
      cursorSupport: true,
      claudeSupport: false,
      notes: 'Beta support, some limitations',
    },
  ]
}
