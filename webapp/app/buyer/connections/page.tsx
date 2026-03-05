'use client'

import { useState, useEffect } from 'react'
import { Copy, RefreshCw, Trash2, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { LoadingState } from '@/components/empty-state'
import { TableToolbar } from '@/components/table-toolbar'
import { toast } from 'sonner'
import { fetchConnections, rotateToken, revokeConnection } from '@/lib/api-client'

export default function ConnectionsPage() {
  const [isLoading, setIsLoading] = useState(true)
  const [connections, setConnections] = useState<any[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedStatus, setSelectedStatus] = useState('')
  const [revoking, setRevoking] = useState<string | null>(null)
  const [rotating, setRotating] = useState<string | null>(null)

  useEffect(() => {
    const loadData = async () => {
      setIsLoading(true)
      try {
        const conns = await fetchConnections()
        setConnections(conns as any)
      } finally {
        setIsLoading(false)
      }
    }

    loadData()
  }, [])

  const filteredConnections = connections
    .filter(c => {
      if (searchQuery) {
        const query = searchQuery.toLowerCase()
        return c.serverName.toLowerCase().includes(query)
      }
      return true
    })
    .filter(c => {
      if (selectedStatus) {
        return c.status === selectedStatus
      }
      return true
    })

  const handleRotateToken = async (connId: string) => {
    setRotating(connId)
    try {
      await rotateToken(connId)
      const updated = connections.map(c =>
        c.id === connId
          ? {
              ...c,
              tokenExpiresAt: new Date(Date.now() + 365 * 24 * 60 * 60 * 1000),
            }
          : c
      )
      setConnections(updated)
      toast.success('Token rotated successfully')
    } catch (error) {
      toast.error('Failed to rotate token')
    } finally {
      setRotating(null)
    }
  }

  const handleRevokeConnection = async (connId: string) => {
    if (!window.confirm('Are you sure? This will revoke access to this server.')) {
      return
    }

    setRevoking(connId)
    try {
      await revokeConnection(connId)
      setConnections(connections.map(c => (c.id === connId ? { ...c, status: 'revoked' } : c)))
      toast.success('Connection revoked')
    } catch (error) {
      toast.error('Failed to revoke connection')
    } finally {
      setRevoking(null)
    }
  }

  if (isLoading) {
    return (
      <AppShell role="buyer">
        <LoadingState />
      </AppShell>
    )
  }

  return (
    <AppShell role="buyer">
      <div className="p-6 space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold mb-2">Connections</h1>
          <p className="text-muted-foreground">Manage your MCP server connections and API tokens</p>
        </div>

        {/* Toolbar */}
        <TableToolbar
          searchPlaceholder="Search by server name..."
          onSearch={setSearchQuery}
          filters={[
            {
              name: 'status',
              label: 'Status',
              options: [
                { value: 'active', label: 'Active' },
                { value: 'expired', label: 'Expired' },
                { value: 'revoked', label: 'Revoked' },
              ],
              onFilter: setSelectedStatus,
            },
          ]}
        />

        {/* Connection Cards */}
        <div className="space-y-4">
          {filteredConnections.length === 0 ? (
            <Card className="p-8 text-center">
              <p className="text-muted-foreground mb-4">No connections found</p>
              <Button asChild>
                <a href="/marketplace">Browse Marketplace</a>
              </Button>
            </Card>
          ) : (
            filteredConnections.map(conn => {
              const isExpiringSoon =
                conn.tokenExpiresAt &&
                new Date(conn.tokenExpiresAt).getTime() - Date.now() < 7 * 24 * 60 * 60 * 1000

              return (
                <Card
                  key={conn.id}
                  className={`p-6 ${
                    conn.status === 'expired' || conn.status === 'revoked'
                      ? 'bg-muted/50 border-muted'
                      : 'hover:border-primary'
                  } transition-colors`}
                >
                  <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                    {/* Server Info */}
                    <div>
                      <h3 className="font-semibold mb-2">{conn.serverName}</h3>
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          <div
                            className={`w-2 h-2 rounded-full ${
                              conn.status === 'active'
                                ? 'bg-green-500'
                                : conn.status === 'expired'
                                  ? 'bg-red-500'
                                  : 'bg-gray-500'
                            }`}
                          />
                          <span className="text-sm capitalize font-medium">{conn.status}</span>
                        </div>
                        <p className="text-xs text-muted-foreground">
                          Connected{' '}
                          {Math.floor(
                            (new Date().getTime() - conn.createdAt.getTime()) / (1000 * 60 * 60 * 24)
                          )}{' '}
                          days ago
                        </p>
                      </div>
                    </div>

                    {/* Scopes */}
                    <div>
                      <h4 className="text-sm text-muted-foreground mb-2">Scopes</h4>
                      <div className="flex flex-wrap gap-1">
                        {conn.scopes.map(scope => (
                          <span
                            key={scope}
                            className="text-xs bg-muted px-2 py-1 rounded border border-border"
                          >
                            {scope}
                          </span>
                        ))}
                      </div>
                    </div>

                    {/* Expiration */}
                    <div>
                      <h4 className="text-sm text-muted-foreground mb-2">Token Expires</h4>
                      {conn.tokenExpiresAt ? (
                        <div>
                          <p className="text-sm font-medium">{conn.tokenExpiresAt.toLocaleDateString()}</p>
                          {isExpiringSoon && (
                            <p className="text-xs text-amber-600 dark:text-amber-400 flex items-center gap-1 mt-1">
                              <AlertCircle className="w-3 h-3" />
                              Expiring soon
                            </p>
                          )}
                        </div>
                      ) : (
                        <p className="text-sm text-muted-foreground">No expiration</p>
                      )}
                    </div>

                    {/* Last Used */}
                    <div>
                      <h4 className="text-sm text-muted-foreground mb-2">Last Used</h4>
                      <p className="text-sm font-medium">
                        {conn.lastUsed ? conn.lastUsed.toLocaleDateString() : 'Never'}
                      </p>
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex flex-wrap gap-2 mt-4 pt-4 border-t border-border">
                    {conn.status === 'active' && (
                      <>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleRotateToken(conn.id)}
                          disabled={rotating === conn.id}
                        >
                          <RefreshCw className="w-4 h-4 mr-2" />
                          {rotating === conn.id ? 'Rotating...' : 'Rotate Token'}
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => {
                            navigator.clipboard.writeText(conn.id)
                            toast.success('Connection ID copied')
                          }}
                        >
                          <Copy className="w-4 h-4 mr-2" />
                          Copy ID
                        </Button>
                      </>
                    )}

                    <Button
                      size="sm"
                      variant="destructive"
                      onClick={() => handleRevokeConnection(conn.id)}
                      disabled={revoking === conn.id || conn.status === 'revoked'}
                    >
                      <Trash2 className="w-4 h-4 mr-2" />
                      {revoking === conn.id ? 'Revoking...' : 'Revoke'}
                    </Button>
                  </div>
                </Card>
              )
            })
          )}
        </div>

        {/* Help Section */}
        <Card className="bg-blue-500/10 border-blue-200 dark:border-blue-800 p-6">
          <h3 className="font-semibold mb-2 flex items-center gap-2">
            <AlertCircle className="w-4 h-4" />
            Need Help?
          </h3>
          <p className="text-sm text-foreground/80 mb-4">
            A connection allows a server to access specific resources on your behalf. You can rotate tokens anytime for
            security, or revoke access entirely.
          </p>
          <ul className="text-sm space-y-2 text-foreground/70">
            <li>• Rotate tokens regularly to maintain security</li>
            <li>• Revoked connections can be re-authorized from the marketplace</li>
            <li>• Monitor token expiration dates to avoid service interruptions</li>
          </ul>
        </Card>
      </div>
    </AppShell>
  )
}
