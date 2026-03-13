'use client'

import { useEffect, useMemo, useState } from 'react'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { TableToolbar } from '@/components/table-toolbar'
import { BarChart } from '@/components/retroui/charts/BarChart'
import { PieChart } from '@/components/retroui/charts/PieChart'
import { Text } from '@/components/retroui'
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
          <Text variant="h3" className="mb-2">Tenant Management</Text>
          <Text variant="body" className="text-muted-foreground">Operational view of real tenant data and status.</Text>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="p-6">
            <Text variant="small" className="text-muted-foreground mb-2">Total Tenants</Text>
            <Text variant="h3">{tenants.length}</Text>
          </Card>
          <Card className="p-6">
            <Text variant="small" className="text-muted-foreground mb-2">Active Tenants</Text>
            <Text variant="h3" className="text-green-600 dark:text-green-400">
              {tenants.filter(tenant => tenant.status === 'active').length}
            </Text>
          </Card>
          <Card className="p-6">
            <Text variant="small" className="text-muted-foreground mb-2">Suspended Tenants</Text>
            <Text variant="h3" className="text-red-600 dark:text-red-400">
              {tenants.filter(tenant => tenant.status === 'suspended').length}
            </Text>
          </Card>
        </div>

        <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
          <Card className="p-6">
            <Text variant="h6" className="mb-4">Plan Distribution</Text>
            <BarChart data={planDistribution} index="name" categories={['value']} />
          </Card>
          <Card className="p-6">
            <Text variant="h6" className="mb-4">Status Distribution</Text>
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
            <Card className="p-8 text-center"><Text variant="body" className="text-muted-foreground">No tenants found.</Text></Card>
          ) : (
            filtered.map(tenant => (
              <Card key={tenant.id} className="p-5">
                <div className="grid grid-cols-1 md:grid-cols-5 gap-4 items-center">
                  <div>
                    <Text variant="caption" className="text-muted-foreground">Name</Text>
                    <Text variant="small">{tenant.name}</Text>
                  </div>
                  <div>
                    <Text variant="caption" className="text-muted-foreground">Slug</Text>
                    <Text variant="small">{tenant.slug}</Text>
                  </div>
                  <div>
                    <Text variant="caption" className="text-muted-foreground">Plan</Text>
                    <Text variant="small">{tenant.planTier}</Text>
                  </div>
                  <div>
                    <Text variant="caption" className="text-muted-foreground">Status</Text>
                    <Text variant="small">{tenant.status}</Text>
                  </div>
                  <div>
                    <Text variant="caption" className="text-muted-foreground">Created</Text>
                    <Text variant="small">{new Date(tenant.createdAt).toLocaleDateString()}</Text>
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

