'use client'

import { useState, useEffect } from 'react'
import { AlertTriangle, AlertCircle, CheckCircle2, Shield } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { TableToolbar } from '@/components/table-toolbar'
import { toast } from 'sonner'
import { fetchSecurityEvents } from '@/lib/api-client'

export default function AdminSecurityPage() {
  const [events, setEvents] = useState<any[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedSeverity, setSelectedSeverity] = useState('')
  const [selectedStatus, setSelectedStatus] = useState('')

  useEffect(() => {
    fetchSecurityEvents().then(setEvents)
  }, [])

  const filteredEvents = events
    .filter(e => {
      if (!searchQuery) return true
      const query = searchQuery.toLowerCase()
      return e.description.toLowerCase().includes(query) || e.actor.toLowerCase().includes(query) || (e.targetServer || '').toLowerCase().includes(query)
    })
    .filter(e => (selectedSeverity ? e.severity === selectedSeverity : true))
    .filter(e => (selectedStatus ? (e.resolved ? selectedStatus === 'resolved' : selectedStatus === 'unresolved') : true))

  const handleResolve = (eventId: string) => {
    setEvents(events.map(e => (e.id === eventId ? { ...e, resolved: true } : e)))
    toast.success('Event marked as resolved')
  }

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical': return <AlertTriangle className="w-4 h-4" />
      case 'high': return <AlertCircle className="w-4 h-4" />
      default: return <Shield className="w-4 h-4" />
    }
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'bg-red-500/20 text-red-700 dark:text-red-400 border-red-200 dark:border-red-800'
      case 'high': return 'bg-orange-500/20 text-orange-700 dark:text-orange-400 border-orange-200 dark:border-orange-800'
      default: return 'bg-amber-500/20 text-amber-700 dark:text-amber-400 border-amber-200 dark:border-amber-800'
    }
  }

  return (
    <AppShell role="admin">
      <div className="p-6 space-y-6">
        <div>
          <h1 className="text-3xl font-bold mb-2">Security Events</h1>
          <p className="text-muted-foreground">Monitor suspicious activity and security threats</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">Total Events</p><p className="text-3xl font-bold">{events.length}</p></Card>
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">Unresolved</p><p className="text-3xl font-bold text-red-600 dark:text-red-400">{events.filter(e => !e.resolved).length}</p></Card>
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">Critical</p><p className="text-3xl font-bold text-red-600 dark:text-red-400">{events.filter(e => e.severity === 'critical').length}</p></Card>
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">This Week</p><p className="text-3xl font-bold">{events.length}</p></Card>
        </div>

        <TableToolbar
          searchPlaceholder="Search by description, actor, or server..."
          onSearch={setSearchQuery}
          filters={[
            { name: 'severity', label: 'Severity', options: [{ value: 'critical', label: 'Critical' }, { value: 'high', label: 'High' }, { value: 'medium', label: 'Medium' }], onFilter: setSelectedSeverity },
            { name: 'status', label: 'Status', options: [{ value: 'unresolved', label: 'Unresolved' }, { value: 'resolved', label: 'Resolved' }], onFilter: setSelectedStatus },
          ]}
        />

        <div className="space-y-3">
          {filteredEvents.length === 0 ? <Card className="p-8 text-center text-muted-foreground"><p>No security events found</p></Card> : filteredEvents.map(event => (
            <Card key={event.id} className={`p-6 border-l-4 ${event.severity === 'critical' ? 'border-l-red-500' : event.severity === 'high' ? 'border-l-orange-500' : 'border-l-amber-500'}`}>
              <div className="space-y-4">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-3 mb-2">
                      <div className={`inline-flex items-center gap-1 text-xs font-medium px-2 py-1 rounded border ${getSeverityColor(event.severity)}`}>{getSeverityIcon(event.severity)}{event.severity.toUpperCase()}</div>
                      {!event.resolved ? <span className="inline-flex items-center gap-1 text-xs font-medium bg-red-500/20 text-red-700 dark:text-red-400 px-2 py-1 rounded">Unresolved</span> : <span className="inline-flex items-center gap-1 text-xs font-medium bg-green-500/20 text-green-700 dark:text-green-400 px-2 py-1 rounded"><CheckCircle2 className="w-3 h-3" />Resolved</span>}
                    </div>
                    <h3 className="font-semibold mb-1">{event.description}</h3>
                  </div>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-4 gap-4 text-sm pt-3 border-t border-border">
                  <div><p className="text-muted-foreground text-xs mb-1">Actor</p><p className="font-mono text-xs truncate">{event.actor}</p></div>
                  <div><p className="text-muted-foreground text-xs mb-1">Target Server</p><p className="text-sm">{event.targetServer}</p></div>
                  <div><p className="text-muted-foreground text-xs mb-1">Timestamp</p><p className="text-sm">{new Date(event.timestamp).toLocaleString()}</p></div>
                  <div className="flex gap-2 justify-end">{!event.resolved && <Button size="sm" variant="outline" onClick={() => handleResolve(event.id)}>Mark Resolved</Button>}<Button size="sm" variant="outline">View Details</Button></div>
                </div>
              </div>
            </Card>
          ))}
        </div>
      </div>
    </AppShell>
  )
}