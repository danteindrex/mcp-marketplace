'use client'

import { useState } from 'react'
import { AlertTriangle, AlertCircle, CheckCircle2, Shield } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { TableToolbar } from '@/components/table-toolbar'
import { toast } from 'sonner'

const securityEvents = [
  {
    id: 'evt_001',
    type: 'ssrf_attempt',
    severity: 'high' as const,
    description: 'SSRF attempt detected on PostgreSQL server internal endpoint',
    actor: 'tenant_unknown_ip_192.168.1.100',
    targetServer: 'PostgreSQL Assistant',
    timestamp: new Date('2025-03-03T14:30:00'),
    resolved: true,
    action: 'Blocked and logged',
  },
  {
    id: 'evt_002',
    type: 'token_reuse',
    severity: 'critical' as const,
    description: 'Auth token used from multiple geographic locations in 1 minute',
    actor: 'user_001@acmecorp.com',
    targetServer: 'GitHub Integration Suite',
    timestamp: new Date('2025-03-02T10:15:00'),
    resolved: true,
    action: 'Token revoked, user notified',
  },
  {
    id: 'evt_003',
    type: 'auth_failure',
    severity: 'medium' as const,
    description: '5 failed authentication attempts in 10 minutes',
    actor: 'user_unknown_ip_203.0.113.5',
    targetServer: 'PostgreSQL Assistant',
    timestamp: new Date('2025-03-01T09:45:00'),
    resolved: false,
    action: 'Pending review',
  },
  {
    id: 'evt_004',
    type: 'unusual_activity',
    severity: 'medium' as const,
    description: 'Excessive API calls from single user (1000+ per minute)',
    actor: 'user_002@dataflow.io',
    targetServer: 'Document Analyzer',
    timestamp: new Date('2025-02-28T16:20:00'),
    resolved: true,
    action: 'Rate limit applied',
  },
  {
    id: 'evt_005',
    type: 'ssrf_attempt',
    severity: 'high' as const,
    description: 'SSRF scanning for internal services detected',
    actor: 'tenant_003_unknown_actor',
    targetServer: 'API Rate Limiter',
    timestamp: new Date('2025-02-27T11:10:00'),
    resolved: false,
    action: 'Tenant suspended',
  },
]

export default function AdminSecurityPage() {
  const [events, setEvents] = useState(securityEvents)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedSeverity, setSelectedSeverity] = useState('')
  const [selectedStatus, setSelectedStatus] = useState('')

  const filteredEvents = events
    .filter(e => {
      if (searchQuery) {
        const query = searchQuery.toLowerCase()
        return (
          e.description.toLowerCase().includes(query) ||
          e.actor.toLowerCase().includes(query) ||
          e.targetServer.toLowerCase().includes(query)
        )
      }
      return true
    })
    .filter(e => {
      if (selectedSeverity) {
        return e.severity === selectedSeverity
      }
      return true
    })
    .filter(e => {
      if (selectedStatus) {
        return e.resolved ? selectedStatus === 'resolved' : selectedStatus === 'unresolved'
      }
      return true
    })

  const handleResolve = (eventId: string) => {
    setEvents(events.map(e => (e.id === eventId ? { ...e, resolved: true } : e)))
    toast.success('Event marked as resolved')
  }

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical':
        return <AlertTriangle className="w-4 h-4" />
      case 'high':
        return <AlertCircle className="w-4 h-4" />
      default:
        return <Shield className="w-4 h-4" />
    }
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-500/20 text-red-700 dark:text-red-400 border-red-200 dark:border-red-800'
      case 'high':
        return 'bg-orange-500/20 text-orange-700 dark:text-orange-400 border-orange-200 dark:border-orange-800'
      default:
        return 'bg-amber-500/20 text-amber-700 dark:text-amber-400 border-amber-200 dark:border-amber-800'
    }
  }

  return (
    <AppShell role="admin">
      <div className="p-6 space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold mb-2">Security Events</h1>
          <p className="text-muted-foreground">Monitor suspicious activity and security threats</p>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Total Events</p>
            <p className="text-3xl font-bold">{events.length}</p>
          </Card>
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Unresolved</p>
            <p className="text-3xl font-bold text-red-600 dark:text-red-400">
              {events.filter(e => !e.resolved).length}
            </p>
          </Card>
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">Critical</p>
            <p className="text-3xl font-bold text-red-600 dark:text-red-400">
              {events.filter(e => e.severity === 'critical').length}
            </p>
          </Card>
          <Card className="p-6">
            <p className="text-sm text-muted-foreground mb-2">This Week</p>
            <p className="text-3xl font-bold">{events.length}</p>
          </Card>
        </div>

        {/* Toolbar */}
        <TableToolbar
          searchPlaceholder="Search by description, actor, or server..."
          onSearch={setSearchQuery}
          filters={[
            {
              name: 'severity',
              label: 'Severity',
              options: [
                { value: 'critical', label: 'Critical' },
                { value: 'high', label: 'High' },
                { value: 'medium', label: 'Medium' },
              ],
              onFilter: setSelectedSeverity,
            },
            {
              name: 'status',
              label: 'Status',
              options: [
                { value: 'unresolved', label: 'Unresolved' },
                { value: 'resolved', label: 'Resolved' },
              ],
              onFilter: setSelectedStatus,
            },
          ]}
        />

        {/* Events List */}
        <div className="space-y-3">
          {filteredEvents.length === 0 ? (
            <Card className="p-8 text-center text-muted-foreground">
              <p>No security events found</p>
            </Card>
          ) : (
            filteredEvents.map(event => (
              <Card
                key={event.id}
                className={`p-6 border-l-4 ${
                  event.severity === 'critical'
                    ? 'border-l-red-500'
                    : event.severity === 'high'
                      ? 'border-l-orange-500'
                      : 'border-l-amber-500'
                }`}
              >
                <div className="space-y-4">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-3 mb-2">
                        <div
                          className={`inline-flex items-center gap-1 text-xs font-medium px-2 py-1 rounded border ${getSeverityColor(event.severity)}`}
                        >
                          {getSeverityIcon(event.severity)}
                          {event.severity.toUpperCase()}
                        </div>
                        {!event.resolved ? (
                          <span className="inline-flex items-center gap-1 text-xs font-medium bg-red-500/20 text-red-700 dark:text-red-400 px-2 py-1 rounded">
                            Unresolved
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-xs font-medium bg-green-500/20 text-green-700 dark:text-green-400 px-2 py-1 rounded">
                            <CheckCircle2 className="w-3 h-3" />
                            Resolved
                          </span>
                        )}
                      </div>
                      <h3 className="font-semibold mb-1">{event.description}</h3>
                      <p className="text-sm text-muted-foreground mb-3">{event.action}</p>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-4 gap-4 text-sm pt-3 border-t border-border">
                    <div>
                      <p className="text-muted-foreground text-xs mb-1">Actor</p>
                      <p className="font-mono text-xs truncate">{event.actor}</p>
                    </div>
                    <div>
                      <p className="text-muted-foreground text-xs mb-1">Target Server</p>
                      <p className="text-sm">{event.targetServer}</p>
                    </div>
                    <div>
                      <p className="text-muted-foreground text-xs mb-1">Timestamp</p>
                      <p className="text-sm">{event.timestamp.toLocaleString()}</p>
                    </div>
                    <div className="flex gap-2 justify-end">
                      {!event.resolved && (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleResolve(event.id)}
                        >
                          Mark Resolved
                        </Button>
                      )}
                      <Button size="sm" variant="outline">
                        View Details
                      </Button>
                    </div>
                  </div>
                </div>
              </Card>
            ))
          )}
        </div>

        {/* Help Section */}
        <Card className="bg-blue-500/10 border-blue-200 dark:border-blue-800 p-6 space-y-3">
          <h3 className="font-semibold">Event Types</h3>
          <ul className="text-sm space-y-2 text-foreground/80">
            <li>
              <strong>SSRF Attempts:</strong> Server-side request forgery attempts on internal
              endpoints
            </li>
            <li>
              <strong>Token Reuse:</strong> Authentication tokens used from multiple locations
            </li>
            <li>
              <strong>Auth Failures:</strong> Repeated failed authentication attempts
            </li>
            <li>
              <strong>Unusual Activity:</strong> Abnormal usage patterns or rate limit violations
            </li>
          </ul>
        </Card>
      </div>
    </AppShell>
  )
}
