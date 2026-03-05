// Mock data for MCP Marketplace
// All data is seeded for realistic UI development

export interface Server {
  id: string
  name: string
  slug: string
  description: string
  category: 'data' | 'automation' | 'ai' | 'integration' | 'other'
  version: string
  author: string
  homepage?: string
  license: string
  dockerImage: string
  verified: boolean
  rating: number
  installCount: number
  toolCount: number
  requiredScopes: string[]
  pricingType: 'free' | 'subscription' | 'flat' | 'x402'
  pricingAmount?: number
  currency?: string
  featured: boolean
  lastUpdated: Date
}

export interface User {
  id: string
  email: string
  name: string
  avatar?: string
  role: 'buyer' | 'merchant' | 'admin'
  organization?: string
  createdAt: Date
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
  plan: 'free' | 'starter' | 'professional' | 'enterprise'
  monthlySpend: number
  currentBalance: number
  nextBillingDate: Date
  paymentMethod?: string
  status: 'active' | 'past_due' | 'suspended'
}

export interface SecurityEvent {
  id: string
  type: 'ssrf_attempt' | 'token_reuse' | 'unusual_activity' | 'auth_failure'
  severity: 'low' | 'medium' | 'high' | 'critical'
  description: string
  actor: string
  targetServer?: string
  timestamp: Date
  resolved: boolean
}

// Mock Servers
export const mockServers: Server[] = [
  {
    id: 'srv_001',
    name: 'PostgreSQL Assistant',
    slug: 'postgresql-assistant',
    description: 'Query and manage PostgreSQL databases with natural language',
    category: 'data',
    version: '2.1.0',
    author: 'DataFlow Inc',
    homepage: 'https://dataflow.io/postgresql',
    license: 'MIT',
    dockerImage: 'dataflow/postgresql-assistant:2.1.0',
    verified: true,
    rating: 4.8,
    installCount: 2340,
    toolCount: 8,
    requiredScopes: ['db:read', 'db:write', 'db:admin'],
    pricingType: 'subscription',
    pricingAmount: 29,
    currency: 'USD',
    featured: true,
    lastUpdated: new Date('2025-02-15'),
  },
  {
    id: 'srv_002',
    name: 'GitHub Integration Suite',
    slug: 'github-integration-suite',
    description: 'Complete GitHub automation tools for repos, issues, and PRs',
    category: 'integration',
    version: '3.0.1',
    author: 'DevTools Labs',
    homepage: 'https://devtools-labs.com/github',
    license: 'Apache-2.0',
    dockerImage: 'devtoolslabs/github-suite:3.0.1',
    verified: true,
    rating: 4.9,
    installCount: 3240,
    toolCount: 12,
    requiredScopes: ['github:repos', 'github:issues', 'github:pulls', 'github:admin'],
    pricingType: 'free',
    featured: true,
    lastUpdated: new Date('2025-03-01'),
  },
  {
    id: 'srv_003',
    name: 'Document Analyzer',
    slug: 'document-analyzer',
    description: 'Extract, analyze, and transform documents with AI',
    category: 'ai',
    version: '1.5.2',
    author: 'DocAI Corp',
    homepage: 'https://docai.io',
    license: 'MIT',
    dockerImage: 'docai/document-analyzer:1.5.2',
    verified: true,
    rating: 4.6,
    installCount: 1820,
    toolCount: 6,
    requiredScopes: ['documents:read', 'documents:write', 'ai:inference'],
    pricingType: 'x402',
    featured: true,
    lastUpdated: new Date('2025-02-20'),
  },
  {
    id: 'srv_004',
    name: 'API Rate Limiter',
    slug: 'api-rate-limiter',
    description: 'Advanced rate limiting and quota management for APIs',
    category: 'automation',
    version: '1.2.0',
    author: 'CloudOps',
    license: 'MIT',
    dockerImage: 'cloudops/rate-limiter:1.2.0',
    verified: false,
    rating: 4.3,
    installCount: 540,
    toolCount: 4,
    requiredScopes: ['api:manage', 'api:observe'],
    pricingType: 'flat',
    pricingAmount: 49,
    currency: 'USD',
    featured: false,
    lastUpdated: new Date('2025-01-10'),
  },
  {
    id: 'srv_005',
    name: 'Email Campaign Manager',
    slug: 'email-campaign-manager',
    description: 'Design, send, and track email campaigns at scale',
    category: 'automation',
    version: '2.0.0',
    author: 'MarketingAI',
    homepage: 'https://marketingai.io',
    license: 'MIT',
    dockerImage: 'marketingai/email-campaign:2.0.0',
    verified: true,
    rating: 4.7,
    installCount: 1260,
    toolCount: 9,
    requiredScopes: ['email:send', 'email:manage', 'analytics:read'],
    pricingType: 'subscription',
    pricingAmount: 59,
    currency: 'USD',
    featured: false,
    lastUpdated: new Date('2025-02-28'),
  },
]

// Mock Users
export const mockUsers: User[] = [
  {
    id: 'user_001',
    email: 'alice@acmecorp.com',
    name: 'Alice Johnson',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Alice',
    role: 'buyer',
    organization: 'Acme Corp',
    createdAt: new Date('2024-06-15'),
  },
  {
    id: 'user_002',
    email: 'bob@dataflow.io',
    name: 'Bob Smith',
    role: 'merchant',
    organization: 'DataFlow Inc',
    createdAt: new Date('2024-03-20'),
  },
  {
    id: 'user_003',
    email: 'charlie@admin.local',
    name: 'Charlie Root',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Charlie',
    role: 'admin',
    createdAt: new Date('2024-01-01'),
  },
]

// Mock Connections
export const mockConnections: Connection[] = [
  {
    id: 'conn_001',
    userId: 'user_001',
    serverId: 'srv_001',
    serverName: 'PostgreSQL Assistant',
    status: 'active',
    scopes: ['db:read', 'db:write'],
    tokenExpiresAt: new Date('2025-09-15'),
    createdAt: new Date('2024-12-15'),
    lastUsed: new Date('2025-03-04'),
  },
  {
    id: 'conn_002',
    userId: 'user_001',
    serverId: 'srv_002',
    serverName: 'GitHub Integration Suite',
    status: 'active',
    scopes: ['github:repos', 'github:issues'],
    tokenExpiresAt: new Date('2025-12-01'),
    createdAt: new Date('2025-01-20'),
    lastUsed: new Date('2025-03-03'),
  },
  {
    id: 'conn_003',
    userId: 'user_001',
    serverId: 'srv_003',
    serverName: 'Document Analyzer',
    status: 'expired',
    scopes: ['documents:read'],
    tokenExpiresAt: new Date('2025-02-15'),
    createdAt: new Date('2024-08-15'),
  },
]

// Mock Billing
export const mockBilling: Billing[] = [
  {
    id: 'bill_001',
    userId: 'user_001',
    plan: 'professional',
    monthlySpend: 245.50,
    currentBalance: 1245.50,
    nextBillingDate: new Date('2025-04-01'),
    paymentMethod: '**** **** **** 4242',
    status: 'active',
  },
]

// Mock Security Events
export const mockSecurityEvents: SecurityEvent[] = [
  {
    id: 'evt_001',
    type: 'ssrf_attempt',
    severity: 'high',
    description: 'SSRF attempt detected on PostgreSQL server internal endpoint',
    actor: 'user_unknown_ip_192.168.1.100',
    targetServer: 'srv_001',
    timestamp: new Date('2025-03-03T14:30:00'),
    resolved: true,
  },
  {
    id: 'evt_002',
    type: 'token_reuse',
    severity: 'critical',
    description: 'Auth token used from multiple geographic locations in 1 minute',
    actor: 'user_001',
    targetServer: 'srv_002',
    timestamp: new Date('2025-03-02T10:15:00'),
    resolved: true,
  },
  {
    id: 'evt_003',
    type: 'auth_failure',
    severity: 'medium',
    description: '5 failed authentication attempts in 10 minutes',
    actor: 'user_unknown_ip_203.0.113.5',
    targetServer: 'srv_001',
    timestamp: new Date('2025-03-01T09:45:00'),
    resolved: false,
  },
]

// Helper functions
export function getServerById(id: string): Server | undefined {
  return mockServers.find(s => s.id === id)
}

export function getUserById(id: string): User | undefined {
  return mockUsers.find(u => u.id === id)
}

export function getConnectionsByUserId(userId: string): Connection[] {
  return mockConnections.filter(c => c.userId === userId)
}

export function getBillingByUserId(userId: string): Billing | undefined {
  return mockBilling.find(b => b.userId === userId)
}

export function getSecurityEventsBySeverity(severity: string): SecurityEvent[] {
  return mockSecurityEvents.filter(e => e.severity === severity)
}

export function searchServers(query: string): Server[] {
  const lower = query.toLowerCase()
  return mockServers.filter(
    s =>
      s.name.toLowerCase().includes(lower) ||
      s.description.toLowerCase().includes(lower) ||
      s.author.toLowerCase().includes(lower)
  )
}

export function filterServersByCategory(category: string): Server[] {
  return mockServers.filter(s => s.category === category)
}

export function getFeaturedServers(): Server[] {
  return mockServers.filter(s => s.featured)
}
