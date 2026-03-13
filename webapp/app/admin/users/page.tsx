'use client'

import { useEffect, useMemo, useState } from 'react'
import { AppShell } from '@/components/app-shell'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Text } from '@/components/retroui'
import { fetchAdminUsers, type AdminUserRecord } from '@/lib/api-client'

export default function AdminUsersPage() {
  const [users, setUsers] = useState<AdminUserRecord[]>([])
  const [query, setQuery] = useState('')

  useEffect(() => {
    fetchAdminUsers().then(setUsers).catch(() => setUsers([]))
  }, [])

  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase()
    if (!needle) return users
    return users.filter(user =>
      user.email.toLowerCase().includes(needle) ||
      user.name.toLowerCase().includes(needle) ||
      String(user.tenantName || '').toLowerCase().includes(needle) ||
      user.role.toLowerCase().includes(needle),
    )
  }, [query, users])

  return (
    <AppShell role="admin">
      <div className="p-6 space-y-6">
        <div>
          <Text variant="h3" className="mb-2">User Directory</Text>
          <Text variant="body" className="text-muted-foreground">Platform-wide user visibility across buyers, merchants, and admins.</Text>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="p-5"><Text variant="caption" className="text-muted-foreground">Total Users</Text><Text variant="h3">{users.length}</Text></Card>
          <Card className="p-5"><Text variant="caption" className="text-muted-foreground">Merchants</Text><Text variant="h3">{users.filter(user => user.role === 'merchant').length}</Text></Card>
          <Card className="p-5"><Text variant="caption" className="text-muted-foreground">Buyers</Text><Text variant="h3">{users.filter(user => user.role === 'buyer').length}</Text></Card>
        </div>

        <Input
          placeholder="Search users by name, email, tenant, or role..."
          value={query}
          onChange={event => setQuery(event.target.value)}
        />

        <div className="space-y-3">
          {filtered.length === 0 ? <Card className="p-8 text-center"><Text variant="body" className="text-muted-foreground">No users found.</Text></Card> : null}
          {filtered.map(user => (
            <Card key={user.id} className="p-5">
              <div className="grid grid-cols-1 md:grid-cols-5 gap-4 text-sm">
                <div>
                  <Text variant="caption" className="text-muted-foreground">Name</Text>
                  <Text variant="small">{user.name}</Text>
                </div>
                <div>
                  <Text variant="caption" className="text-muted-foreground">Email</Text>
                  <Text variant="small" className="break-all">{user.email}</Text>
                </div>
                <div>
                  <Text variant="caption" className="text-muted-foreground">Role</Text>
                  <Text variant="small">{user.role}</Text>
                </div>
                <div>
                  <Text variant="caption" className="text-muted-foreground">Tenant</Text>
                  <Text variant="small">{user.tenantName || user.tenantId}</Text>
                </div>
                <div>
                  <Text variant="caption" className="text-muted-foreground">Created</Text>
                  <Text variant="small">{new Date(user.createdAt).toLocaleDateString()}</Text>
                </div>
              </div>
            </Card>
          ))}
        </div>
      </div>
    </AppShell>
  )
}
