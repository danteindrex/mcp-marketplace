'use client'

import { useState } from 'react'
import Link from 'next/link'
import { ArrowLeft, Plus, Trash2, AlertCircle, Copy } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { AppShell } from '@/components/app-shell'
import { toast } from 'sonner'

export default function AuthConfigPage({ params }: any) {
  const [authServers, setAuthServers] = useState<any[]>([
    {
      id: '1',
      name: 'Primary OAuth Provider',
      url: 'https://auth.example.com',
      clientId: 'client_123',
      metadataUrl: 'https://auth.example.com/.well-known/openid-configuration',
      registrationMode: 'pre-registered',
      scopes: ['profile', 'email'],
    },
  ])

  const [editingId, setEditingId] = useState<string | null>(null)
  const [showNewForm, setShowNewForm] = useState(false)
  const [newServer, setNewServer] = useState({
    name: '',
    url: '',
    clientId: '',
    clientSecret: '',
    metadataUrl: '',
    registrationMode: 'pre-registered' as const,
    scopes: [] as string[],
  })

  const handleAddServer = () => {
    if (!newServer.name || !newServer.url) {
      toast.error('Please fill in required fields')
      return
    }

    const server = {
      id: Date.now().toString(),
      ...newServer,
    }
    setAuthServers([...authServers, server])
    setNewServer({
      name: '',
      url: '',
      clientId: '',
      clientSecret: '',
      metadataUrl: '',
      registrationMode: 'pre-registered',
      scopes: [],
    })
    setShowNewForm(false)
    toast.success('Auth server added')
  }

  const handleDeleteServer = (id: string) => {
    setAuthServers(authServers.filter(s => s.id !== id))
    toast.success('Auth server removed')
  }

  const handleUpdateServer = (id: string, field: string, value: any) => {
    setAuthServers(
      authServers.map(s => (s.id === id ? { ...s, [field]: value } : s))
    )
  }

  const toggleScope = (id: string, scope: string) => {
    handleUpdateServer(
      id,
      'scopes',
      (authServers.find(s => s.id === id)?.scopes || []).includes(scope)
        ? authServers.find(s => s.id === id)?.scopes.filter((s: string) => s !== scope)
        : [...(authServers.find(s => s.id === id)?.scopes || []), scope]
    )
  }

  const availableScopes = [
    { value: 'profile', label: 'User Profile' },
    { value: 'email', label: 'Email Address' },
    { value: 'openid', label: 'OpenID' },
    { value: 'offline_access', label: 'Offline Access' },
    { value: 'custom:read', label: 'Custom Read' },
    { value: 'custom:write', label: 'Custom Write' },
  ]

  return (
    <AppShell role="merchant">
      <div className="max-w-4xl mx-auto p-6 space-y-6">
        {/* Header */}
        <Link
          href="/merchant/servers"
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors w-fit"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Servers
        </Link>

        <div>
          <h1 className="text-3xl font-bold mb-2">Authentication Configuration</h1>
          <p className="text-muted-foreground">
            Set up OAuth2 and scope requirements for your server
          </p>
        </div>

        {/* Required Scopes */}
        <Card className="p-8">
          <h2 className="text-xl font-bold mb-4">Required Scopes</h2>
          <p className="text-sm text-muted-foreground mb-6">
            Define which permissions users must grant to use your server
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {availableScopes.map(scope => (
              <div
                key={scope.value}
                className="flex items-center space-x-3 p-3 border border-border rounded-lg hover:bg-muted/50 transition-colors"
              >
                <Checkbox
                  id={`scope-${scope.value}`}
                  defaultChecked={['profile', 'email'].includes(scope.value)}
                />
                <Label
                  htmlFor={`scope-${scope.value}`}
                  className="flex-1 cursor-pointer"
                >
                  <p className="font-medium text-sm">{scope.label}</p>
                  <p className="text-xs text-muted-foreground">{scope.value}</p>
                </Label>
              </div>
            ))}
          </div>

          <Button variant="outline" className="mt-4">
            Add Custom Scope
          </Button>
        </Card>

        {/* OAuth Servers */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-bold">OAuth Servers</h2>
            <Button
              size="sm"
              onClick={() => setShowNewForm(!showNewForm)}
            >
              <Plus className="w-4 h-4 mr-2" />
              Add Server
            </Button>
          </div>

          {/* New Server Form */}
          {showNewForm && (
            <Card className="p-6 space-y-4">
              <h3 className="font-semibold">New OAuth Server</h3>

              <div>
                <Label className="text-xs mb-2 block">Server Name</Label>
                <Input
                  value={newServer.name}
                  onChange={e => setNewServer({ ...newServer, name: e.target.value })}
                  placeholder="e.g., Production Auth"
                />
              </div>

              <div>
                <Label className="text-xs mb-2 block">Authorization Server URL</Label>
                <Input
                  value={newServer.url}
                  onChange={e => setNewServer({ ...newServer, url: e.target.value })}
                  placeholder="https://auth.example.com"
                  className="font-mono text-sm"
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <Label className="text-xs mb-2 block">Client ID</Label>
                  <Input
                    value={newServer.clientId}
                    onChange={e =>
                      setNewServer({ ...newServer, clientId: e.target.value })
                    }
                    placeholder="client_..."
                    className="font-mono text-sm"
                  />
                </div>

                <div>
                  <Label className="text-xs mb-2 block">Client Secret</Label>
                  <Input
                    type="password"
                    value={newServer.clientSecret}
                    onChange={e =>
                      setNewServer({ ...newServer, clientSecret: e.target.value })
                    }
                    placeholder="••••••••••••••••"
                  />
                </div>
              </div>

              <div>
                <Label className="text-xs mb-2 block">Metadata URL (Optional)</Label>
                <Input
                  value={newServer.metadataUrl}
                  onChange={e =>
                    setNewServer({ ...newServer, metadataUrl: e.target.value })
                  }
                  placeholder="https://auth.example.com/.well-known/openid-configuration"
                  className="font-mono text-sm"
                />
              </div>

              <div>
                <Label className="text-xs mb-2 block">Registration Mode</Label>
                <Select
                  value={newServer.registrationMode}
                  onValueChange={value =>
                    setNewServer({
                      ...newServer,
                      registrationMode: value as any,
                    })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="pre-registered">Pre-Registered Clients</SelectItem>
                    <SelectItem value="cimd">Client Initiated MetaData Discovery</SelectItem>
                    <SelectItem value="dcr">Dynamic Client Registration</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="flex gap-2">
                <Button variant="outline" onClick={() => setShowNewForm(false)}>
                  Cancel
                </Button>
                <Button onClick={handleAddServer}>Add Server</Button>
              </div>
            </Card>
          )}

          {/* Auth Servers List */}
          {authServers.length === 0 && !showNewForm ? (
            <Card className="p-8 text-center text-muted-foreground">
              <p>No auth servers configured yet.</p>
            </Card>
          ) : (
            <div className="space-y-4">
              {authServers.map(server => (
                <Card
                  key={server.id}
                  className="p-6 space-y-4"
                  onMouseEnter={() => setEditingId(server.id)}
                  onMouseLeave={() => setEditingId(null)}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <h3 className="font-semibold">{server.name}</h3>
                      <p className="text-sm text-muted-foreground font-mono">
                        {server.url}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => {
                          navigator.clipboard.writeText(server.url)
                          toast.success('URL copied')
                        }}
                      >
                        <Copy className="w-4 h-4" />
                      </Button>
                      <Button
                        size="sm"
                        variant="destructive"
                        onClick={() => handleDeleteServer(server.id)}
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>

                  <div className="border-t border-border pt-4 space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div>
                        <p className="text-xs text-muted-foreground mb-1">Client ID</p>
                        <p className="text-sm font-mono">{server.clientId}</p>
                      </div>
                      <div>
                        <p className="text-xs text-muted-foreground mb-1">
                          Registration Mode
                        </p>
                        <p className="text-sm capitalize">{server.registrationMode}</p>
                      </div>
                    </div>

                    {server.metadataUrl && (
                      <div>
                        <p className="text-xs text-muted-foreground mb-1">
                          Metadata URL
                        </p>
                        <p className="text-sm font-mono truncate">{server.metadataUrl}</p>
                      </div>
                    )}

                    {server.scopes && server.scopes.length > 0 && (
                      <div>
                        <p className="text-xs text-muted-foreground mb-2">
                          Required Scopes
                        </p>
                        <div className="flex flex-wrap gap-2">
                          {server.scopes.map((scope: string) => (
                            <span
                              key={scope}
                              className="text-xs bg-muted px-2 py-1 rounded border border-border"
                            >
                              {scope}
                            </span>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                </Card>
              ))}
            </div>
          )}
        </div>

        {/* Help Section */}
        <Card className="bg-blue-500/10 border-blue-200 dark:border-blue-800 p-6 space-y-3">
          <h3 className="font-semibold text-sm flex items-center gap-2">
            <AlertCircle className="w-4 h-4" />
            OAuth Registration Modes
          </h3>
          <ul className="text-sm space-y-2 text-foreground/80">
            <li>
              <strong>Pre-Registered:</strong> Clients are pre-registered in your auth server
            </li>
            <li>
              <strong>CIMD:</strong> Client uses metadata discovery to find your auth endpoints
            </li>
            <li>
              <strong>DCR:</strong> Dynamic registration allowing clients to self-register
            </li>
          </ul>
        </Card>

        {/* Actions */}
        <div className="flex gap-3">
          <Button variant="outline" asChild>
            <Link href="/merchant/servers">Cancel</Link>
          </Button>
          <Button className="flex-1">Save Configuration</Button>
        </div>
      </div>
    </AppShell>
  )
}
