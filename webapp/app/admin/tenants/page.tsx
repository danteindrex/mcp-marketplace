'use client'

import { useEffect, useMemo, useState } from 'react'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { TableToolbar } from '@/components/table-toolbar'
import { BarChart } from '@/components/retroui/charts/BarChart'
import { PieChart } from '@/components/retroui/charts/PieChart'
import { fetchTenants, type TenantRecord } from '@/lib/api-client'

export default function AdminTenantsPage() {
  const [tenants, setTenants] = useState<TenantRecord[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedStatus, setSelectedStatus] = useState('')

  useEffect(() => {
    fetchTenants()
      .then(setTenants)
      .catch(() => setTenants([]))
  }, [])

  const filtered = useMemo(() => {
    return tenants.filter(tenant => {
      const matchesSearch =
        !searchQuery ||
        tenant.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        tenant.slug.toLowerCase().includes(searchQuery.toLowerCase())
      const matchesStatus = !selectedStatus || tenant.status === selectedStatus
      return matchesSearch && matchesStatus
    })
  }, [tenants, searchQuery, selectedStatus])

  const planDistribution = useMemo(() => {
    const counts = new Map<string, number>()
    for (const tenant of tenants) {
      counts.set(tenant.planTier, (counts.get(tenant.planTier) || 0) + 1)
    }
    return Array.from(counts.entries()).map(([name, value]) => ({ name, value }))
  }, [tenants])

  const statusDistribution = useMemo(() => {
    const counts = new Map<string, number>()
    for (const tenant of tenants) {
      counts.set(tenant.status, (counts.get(tenant.status) || 0) + 1)
    }
    return Array.from(counts.entries()).map(([name, value]) => ({ name, value }))
  }, [tenants])

  return (
    <AppShell role="admin">
      <div className="p-6 space-y-6">
        <div>
          <h1 className="text-3xl font-bold mb-2">Tenant Management</h1>
          <p className="text-muted-foreground">Operational view of real tenant data and status.</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Total Tenants</p>
            <p className="text-3xl font-bold">{tenants.length}</p>
          </Card>
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Active Tenants</p>
            <p className="text-3xl font-bold text-green-600 dark:text-green-400">
              {tenants.filter(tenant => tenant.status === 'active').length}
            </p>
          </Card>
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Suspended Tenants</p>
            <p className="text-3xl font-bold text-red-600 dark:text-red-400">
              {tenants.filter(tenant => tenant.status === 'suspended').length}
            </p>
          </Card>
        </div>

        <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
          <Card className="p-6">
            <h2 className="text-lg font-semibold mb-4">Plan Distribution</h2>
            <BarChart data={planDistribution} index="name" categories={['value']} />
          </Card>
          <Card className="p-6">
            <h2 className="text-lg font-semibold mb-4">Status Distribution</h2>
            <PieChart data={statusDistribution} dataKey="value" nameKey="name" />
          </Card>
        </div>

        <TableToolbar
          searchPlaceholder="Search by tenant name or slug..."
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

        <div className="space-y-3">
          {filtered.length === 0 ? (
            <Card className="p-8 text-center text-muted-foreground">No tenants found.</Card>
          ) : (
            filtered.map(tenant => (
              <Card key={tenant.id} className="p-5">
                <div className="grid grid-cols-1 md:grid-cols-5 gap-4 items-center">
                  <div>
                    <p className="text-xs text-muted-foreground">Name</p>
                    <p className="font-semibold">{tenant.name}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Slug</p>
                    <p className="font-semibold">{tenant.slug}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Plan</p>
                    <p className="font-semibold">{tenant.planTier}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Status</p>
                    <p className="font-semibold">{tenant.status}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Created</p>
                    <p className="font-semibold">{new Date(tenant.createdAt).toLocaleDateString()}</p>
                  </div>
                </div>
              </Card>
            ))
          )}
        </div>
      </div>
    </AppShell>
  )
}

