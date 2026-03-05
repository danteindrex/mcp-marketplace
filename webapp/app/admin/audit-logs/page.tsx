'use client'

import { useState, useEffect } from 'react'
import { Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { AppShell } from '@/components/app-shell'
import { TableToolbar } from '@/components/table-toolbar'
import { fetchAuditLogs } from '@/lib/api-client'

export default function AdminAuditLogsPage() {
  const [logs, setLogs] = useState<any[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedAction, setSelectedAction] = useState('')
  const [selectedResult, setSelectedResult] = useState('')

  useEffect(() => {
    fetchAuditLogs().then(setLogs)
  }, [])

  const actionTypes = Array.from(new Set(logs.map(l => l.action)))

  const filteredLogs = logs
    .filter(log => {
      if (!searchQuery) return true
      const query = searchQuery.toLowerCase()
      return log.actor.toLowerCase().includes(query) || log.action.toLowerCase().includes(query) || log.target.toLowerCase().includes(query) || log.details.toLowerCase().includes(query)
    })
    .filter(log => (selectedAction ? log.action === selectedAction : true))
    .filter(log => (selectedResult ? log.result === selectedResult : true))

  return (
    <AppShell role="admin">
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold mb-2">Audit Logs</h1>
            <p className="text-muted-foreground">Append-only record of all platform actions and changes</p>
          </div>
          <Button variant="outline"><Download className="w-4 h-4 mr-2" />Export Logs</Button>
        </div>

        <TableToolbar
          searchPlaceholder="Search by actor, action, target, or details..."
          onSearch={setSearchQuery}
          filters={[
            { name: 'action', label: 'Action', options: actionTypes.map(action => ({ value: action, label: action.replace(/_/g, ' ') })), onFilter: setSelectedAction },
            { name: 'result', label: 'Result', options: [{ value: 'success', label: 'Success' }, { value: 'failure', label: 'Failure' }], onFilter: setSelectedResult },
          ]}
        />

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="p-4"><p className="text-sm text-muted-foreground mb-1">Total Logs</p><p className="text-2xl font-bold">{logs.length}</p></Card>
          <Card className="p-4"><p className="text-sm text-muted-foreground mb-1">Successful Actions</p><p className="text-2xl font-bold text-green-600 dark:text-green-400">{logs.filter(l => l.result === 'success').length}</p></Card>
          <Card className="p-4"><p className="text-sm text-muted-foreground mb-1">Failed Actions</p><p className="text-2xl font-bold text-red-600 dark:text-red-400">{logs.filter(l => l.result === 'failure').length}</p></Card>
        </div>

        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 border-b border-border"><tr><th className="text-left py-3 px-6 font-semibold text-muted-foreground whitespace-nowrap">Timestamp</th><th className="text-left py-3 px-6 font-semibold text-muted-foreground whitespace-nowrap">Actor</th><th className="text-left py-3 px-6 font-semibold text-muted-foreground whitespace-nowrap">Action</th><th className="text-left py-3 px-6 font-semibold text-muted-foreground whitespace-nowrap">Target</th><th className="text-left py-3 px-6 font-semibold text-muted-foreground whitespace-nowrap">Result</th><th className="text-left py-3 px-6 font-semibold text-muted-foreground whitespace-nowrap">Details</th></tr></thead>
              <tbody>
                {filteredLogs.length === 0 ? <tr><td colSpan={6} className="text-center py-8 text-muted-foreground">No logs found</td></tr> : filteredLogs.map(log => (
                  <tr key={log.id} className="border-b border-border hover:bg-muted/50 transition-colors">
                    <td className="py-3 px-6 font-mono text-xs whitespace-nowrap">{new Date(log.timestamp).toLocaleString()}</td>
                    <td className="py-3 px-6"><code className="text-xs bg-muted px-2 py-1 rounded">{log.actor}</code></td>
                    <td className="py-3 px-6"><span className="text-xs font-medium bg-primary/10 text-primary px-2 py-1 rounded">{log.action.replace(/_/g, ' ')}</span></td>
                    <td className="py-3 px-6 truncate max-w-sm">{log.target}</td>
                    <td className="py-3 px-6">{log.result === 'success' ? <span className="inline-flex items-center gap-1 text-xs bg-green-500/20 text-green-700 dark:text-green-400 px-2 py-1 rounded">Success</span> : <span className="inline-flex items-center gap-1 text-xs bg-red-500/20 text-red-700 dark:text-red-400 px-2 py-1 rounded">Failure</span>}</td>
                    <td className="py-3 px-6 text-muted-foreground truncate max-w-md">{log.details}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      </div>
    </AppShell>
  )
}