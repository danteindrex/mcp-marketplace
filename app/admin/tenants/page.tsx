'use client'

import { useState } from 'react'
import { Shield, Ban, AlertTriangle, CheckCircle2, Lock, Unlock } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { TableToolbar } from '@/components/table-toolbar'
import { toast } from 'sonner'

const tenantData = [
  {
    id: 'tenant_001',
    name: 'Acme Corp',
    email: 'admin@acmecorp.com',
    status: 'active' as const,
    riskScore: 12,
    suspensions: 0,
    installs: 1240,
    createdAt: new Date('2024-06-15'),
    lastActivity: new Date('2025-03-04'),
  },
  {
    id: 'tenant_002',
    name: 'DataFlow Inc',
    email: 'admin@dataflow.io',
    status: 'active' as const,
    riskScore: 5,
    suspensions: 0,
    installs: 3240,
    createdAt: new Date('2024-03-20'),
    lastActivity: new Date('2025-03-03'),
  },
  {
    id: 'tenant_003',
    name: 'TechStartup Labs',
    email: 'legal@techstartuplabs.com',
    status: 'active' as const,
    riskScore: 45,
    suspensions: 1,
    installs: 340,
    createdAt: new Date('2025-01-10'),
    lastActivity: new Date('2025-02-28'),
  },
  {
    id: 'tenant_004',
    name: 'Bad Actors Inc',
    email: 'contact@badactors.local',
    status: 'suspended' as const,
    riskScore: 92,
    suspensions: 3,
    installs: 0,
    createdAt: new Date('2025-02-01'),
    lastActivity: new Date('2025-02-15'),
  },
]

export default function AdminTenantsPage() {
  const [tenants, setTenants] = useState(tenantData)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedStatus, setSelectedStatus] = useState('')
  const [expandedId, setExpandedId] = useState<string | null>(null)

  const filteredTenants = tenants
    .filter(t => {
      if (searchQuery) {
        const query = searchQuery.toLowerCase()
        return (
          t.name.toLowerCase().includes(query) ||
          t.email.toLowerCase().includes(query)
        )
      }
      return true
    })
    .filter(t => {
      if (selectedStatus) {
        return t.status === selectedStatus
      }
      return true
    })

  const handleSuspend = (tenantId: string) => {
    setTenants(
      tenants.map(t =>
        t.id === tenantId
          ? { ...t, status: 'suspended' as const, suspensions: t.suspensions + 1 }
          : t
      )
    )
    toast.success('Tenant suspended')
  }

  const handleUnsuspend = (tenantId: string) => {
    setTenants(
      tenants.map(t =>
        t.id === tenantId ? { ...t, status: 'active' as const } : t
      )
    )
    toast.success('Tenant unsuspended')
  }

  const getRiskColor = (score: number) => {
    if (score < 20) return 'text-green-600 dark:text-green-400 bg-green-500/10'
    if (score < 50) return 'text-amber-600 dark:text-amber-400 bg-amber-500/10'
    return 'text-red-600 dark:text-red-400 bg-red-500/10'
  }

  return (
    <AppShell role="admin">
      <div className="p-6 space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold mb-2">Tenant Management</h1>
          <p className="text-muted-foreground">
            Oversee organizations, risk assessments, and account status
          </p>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Total Tenants</p>
            <p className="text-3xl font-bold">{tenants.length}</p>
          </Card>
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Active</p>
            <p className="text-3xl font-bold text-green-600 dark:text-green-400">
              {tenants.filter(t => t.status === 'active').length}
            </p>
          </Card>
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Suspended</p>
            <p className="text-3xl font-bold text-red-600 dark:text-red-400">
              {tenants.filter(t => t.status === 'suspended').length}
            </p>
          </Card>
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">High Risk</p>
            <p className="text-3xl font-bold text-amber-600 dark:text-amber-400">
              {tenants.filter(t => t.riskScore > 50).length}
            </p>
          </Card>
        </div>

        {/* Toolbar */}
        <TableToolbar
          searchPlaceholder="Search by name or email..."
          onSearch={setSearchQuery}
          filters={[
            {
              name: 'status',
              label: 'Status',
              options: [
                { value: 'active', label: 'Active' },
                { value: 'suspended', label: 'Suspended' },
              ],
              onFilter: setSelectedStatus,
            },
          ]}
        />

        {/* Tenants List */}
        <div className="space-y-3">
          {filteredTenants.length === 0 ? (
            <Card className="p-8 text-center text-muted-foreground">
              <p>No tenants found</p>
            </Card>
          ) : (
            filteredTenants.map(tenant => (
              <Card
                key={tenant.id}
                className="overflow-hidden"
              >
                <button
                  onClick={() =>
                    setExpandedId(expandedId === tenant.id ? null : tenant.id)
                  }
                  className="w-full text-left p-6 hover:bg-muted/50 transition-colors"
                >
                  <div className="grid grid-cols-1 md:grid-cols-6 gap-4 items-center">
                    {/* Name */}
                    <div>
                      <h3 className="font-semibold mb-1">{tenant.name}</h3>
                      <p className="text-sm text-muted-foreground">{tenant.email}</p>
                    </div>

                    {/* Status */}
                    <div>
                      <div
                        className={`inline-flex items-center gap-1 text-xs font-medium px-2 py-1 rounded ${
                          tenant.status === 'active'
                            ? 'bg-green-500/20 text-green-700 dark:text-green-400'
                            : 'bg-red-500/20 text-red-700 dark:text-red-400'
                        }`}
                      >
                        {tenant.status === 'active' ? (
                          <CheckCircle2 className="w-3 h-3" />
                        ) : (
                          <Ban className="w-3 h-3" />
                        )}
                        {tenant.status === 'active' ? 'Active' : 'Suspended'}
                      </div>
                    </div>

                    {/* Risk Score */}
                    <div>
                      <p className="text-xs text-muted-foreground mb-1">Risk Score</p>
                      <p className={`text-lg font-bold px-2 py-1 rounded text-center ${getRiskColor(tenant.riskScore)}`}>
                        {tenant.riskScore}
                      </p>
                    </div>

                    {/* Suspensions */}
                    <div>
                      <p className="text-xs text-muted-foreground mb-1">Suspensions</p>
                      <p className="font-semibold">{tenant.suspensions}</p>
                    </div>

                    {/* Installs */}
                    <div>
                      <p className="text-xs text-muted-foreground mb-1">Installs</p>
                      <p className="font-semibold">{tenant.installs.toLocaleString()}</p>
                    </div>

                    {/* Expand Indicator */}
                    <div className="text-right">
                      <span className="text-muted-foreground">
                        {expandedId === tenant.id ? '▼' : '▶'}
                      </span>
                    </div>
                  </div>
                </button>

                {/* Expanded Details */}
                {expandedId === tenant.id && (
                  <div className="border-t border-border bg-muted/30 p-6 space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div>
                        <h4 className="font-semibold mb-3">Account Information</h4>
                        <div className="space-y-2 text-sm">
                          <div>
                            <p className="text-muted-foreground">Created</p>
                            <p>{tenant.createdAt.toLocaleDateString()}</p>
                          </div>
                          <div>
                            <p className="text-muted-foreground">Last Activity</p>
                            <p>{tenant.lastActivity.toLocaleDateString()}</p>
                          </div>
                          <div>
                            <p className="text-muted-foreground">Total Installs</p>
                            <p>{tenant.installs.toLocaleString()}</p>
                          </div>
                        </div>
                      </div>

                      <div>
                        <h4 className="font-semibold mb-3">Risk Assessment</h4>
                        <div className="space-y-2 text-sm">
                          <div className="flex items-center justify-between">
                            <span className="text-muted-foreground">Risk Score</span>
                            <span className={`font-bold px-2 py-1 rounded ${getRiskColor(tenant.riskScore)}`}>
                              {tenant.riskScore}%
                            </span>
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="text-muted-foreground">Total Suspensions</span>
                            <span className="font-bold">{tenant.suspensions}</span>
                          </div>
                          {tenant.riskScore > 50 && (
                            <div className="mt-2 p-3 bg-amber-500/10 border border-amber-200 dark:border-amber-800 rounded text-amber-900 dark:text-amber-100 text-xs">
                              <div className="flex items-start gap-2">
                                <AlertTriangle className="w-4 h-4 flex-shrink-0 mt-0.5" />
                                <p>High-risk tenant. Consider suspension or investigation.</p>
                              </div>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>

                    {/* Actions */}
                    <div className="flex flex-wrap gap-2 pt-4 border-t border-border">
                      {tenant.status === 'active' ? (
                        <Button
                          size="sm"
                          variant="destructive"
                          onClick={() => handleSuspend(tenant.id)}
                        >
                          <Ban className="w-4 h-4 mr-2" />
                          Suspend Tenant
                        </Button>
                      ) : (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleUnsuspend(tenant.id)}
                        >
                          <Unlock className="w-4 h-4 mr-2" />
                          Unsuspend
                        </Button>
                      )}
                      <Button size="sm" variant="outline">
                        <Shield className="w-4 h-4 mr-2" />
                        Review Risk
                      </Button>
                      <Button size="sm" variant="outline">
                        <Lock className="w-4 h-4 mr-2" />
                        Force Password Reset
                      </Button>
                    </div>
                  </div>
                )}
              </Card>
            ))
          )}
        </div>
      </div>
    </AppShell>
  )
}
