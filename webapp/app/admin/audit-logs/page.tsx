'use client'

import { useState } from 'react'
import { Download, Filter } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { AppShell } from '@/components/app-shell'
import { TableToolbar } from '@/components/table-toolbar'

const auditLogs = [
  {
    id: 'log_001',
    timestamp: new Date('2025-03-04T15:30:45'),
    actor: 'user_001@acmecorp.com',
    action: 'INSTALL_SERVER',
    target: 'PostgreSQL Assistant v2.1.0',
    result: 'success' as const,
    details: 'Installation completed successfully',
  },
  {
    id: 'log_002',
    timestamp: new Date('2025-03-04T14:15:22'),
    actor: 'admin@marketplace.local',
    action: 'SUSPEND_TENANT',
    target: 'Bad Actors Inc (tenant_004)',
    result: 'success' as const,
    details: 'Suspended due to SSRF attack attempts',
  },
  {
    id: 'log_003',
    timestamp: new Date('2025-03-04T13:42:10'),
    actor: 'user_002@dataflow.io',
    action: 'PUBLISH_SERVER',
    target: 'GitHub Integration Suite v3.0.1',
    result: 'success' as const,
    details: 'Server published to marketplace',
  },
  {
    id: 'log_004',
    timestamp: new Date('2025-03-04T12:18:33'),
    actor: 'user_003@example.com',
    action: 'REVOKE_TOKEN',
    target: 'Connection conn_003',
    result: 'success' as const,
    details: 'User manually revoked auth token',
  },
  {
    id: 'log_005',
    timestamp: new Date('2025-03-04T11:05:12'),
    actor: 'system@marketplace.local',
    action: 'AUTO_SUSPEND_ACCOUNT',
    target: 'TechStartup Labs',
    result: 'success' as const,
    details: 'Suspended after 3 SSRF attempts detected',
  },
  {
    id: 'log_006',
    timestamp: new Date('2025-03-04T10:30:48'),
    actor: 'user_001@acmecorp.com',
    action: 'UPDATE_BILLING',
    target: 'Payment method ending in 4242',
    result: 'success' as const,
    details: 'Billing information updated',
  },
  {
    id: 'log_007',
    timestamp: new Date('2025-03-04T09:15:20'),
    actor: 'admin@marketplace.local',
    action: 'LOGIN_ADMIN',
    target: 'Admin Portal',
    result: 'success' as const,
    details: 'Successful admin portal login',
  },
  {
    id: 'log_008',
    timestamp: new Date('2025-03-03T16:45:33'),
    actor: 'user_002@dataflow.io',
    action: 'UPDATE_SERVER',
    target: 'PostgreSQL Assistant',
    result: 'success' as const,
    details: 'Updated server metadata and configuration',
  },
]

const actionTypes = [
  'INSTALL_SERVER',
  'PUBLISH_SERVER',
  'REVOKE_TOKEN',
  'SUSPEND_TENANT',
  'UPDATE_BILLING',
  'LOGIN_ADMIN',
  'UPDATE_SERVER',
  'AUTO_SUSPEND_ACCOUNT',
]

export default function AdminAuditLogsPage() {
  const [logs, setLogs] = useState(auditLogs)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedAction, setSelectedAction] = useState('')
  const [selectedResult, setSelectedResult] = useState('')

  const filteredLogs = logs
    .filter(log => {
      if (searchQuery) {
        const query = searchQuery.toLowerCase()
        return (
          log.actor.toLowerCase().includes(query) ||
          log.action.toLowerCase().includes(query) ||
          log.target.toLowerCase().includes(query) ||
          log.details.toLowerCase().includes(query)
        )
      }
      return true
    })
    .filter(log => {
      if (selectedAction) {
        return log.action === selectedAction
      }
      return true
    })
    .filter(log => {
      if (selectedResult) {
        return log.result === selectedResult
      }
      return true
    })

  return (
    <AppShell role="admin">
      <div className="p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold mb-2">Audit Logs</h1>
            <p className="text-muted-foreground">
              Append-only record of all platform actions and changes
            </p>
          </div>

          <Button variant="outline">
            <Download className="w-4 h-4 mr-2" />
            Export Logs
          </Button>
        </div>

        {/* Info */}
        <Card className="bg-blue-500/10 border-blue-200 dark:border-blue-800 p-4">
          <p className="text-sm text-blue-900 dark:text-blue-100">
            Audit logs are immutable and retained for 90 days for compliance purposes. All user actions,
            administrative decisions, and system events are logged here.
          </p>
        </Card>

        {/* Filters */}
        <TableToolbar
          searchPlaceholder="Search by actor, action, target, or details..."
          onSearch={setSearchQuery}
          filters={[
            {
              name: 'action',
              label: 'Action',
              options: actionTypes.map(action => ({
                value: action,
                label: action.replace(/_/g, ' '),
              })),
              onFilter: setSelectedAction,
            },
            {
              name: 'result',
              label: 'Result',
              options: [
                { value: 'success', label: 'Success' },
                { value: 'failure', label: 'Failure' },
              ],
              onFilter: setSelectedResult,
            },
          ]}
        />

        {/* Stats */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="p-4">
            <p className="text-sm text-muted-foreground mb-1">Total Logs</p>
            <p className="text-2xl font-bold">{logs.length}</p>
          </Card>
          <Card className="p-4">
            <p className="text-sm text-muted-foreground mb-1">Successful Actions</p>
            <p className="text-2xl font-bold text-green-600 dark:text-green-400">
              {logs.filter(l => l.result === 'success').length}
            </p>
          </Card>
          <Card className="p-4">
            <p className="text-sm text-muted-foreground mb-1">Failed Actions</p>
            <p className="text-2xl font-bold text-red-600 dark:text-red-400">
              {logs.filter(l => l.result === 'failure').length}
            </p>
          </Card>
        </div>

        {/* Logs Table */}
        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 border-b border-border">
                <tr>
                  <th className="text-left py-3 px-6 font-semibold text-muted-foreground whitespace-nowrap">
                    Timestamp
                  </th>
                  <th className="text-left py-3 px-6 font-semibold text-muted-foreground whitespace-nowrap">
                    Actor
                  </th>
                  <th className="text-left py-3 px-6 font-semibold text-muted-foreground whitespace-nowrap">
                    Action
                  </th>
                  <th className="text-left py-3 px-6 font-semibold text-muted-foreground whitespace-nowrap">
                    Target
                  </th>
                  <th className="text-left py-3 px-6 font-semibold text-muted-foreground whitespace-nowrap">
                    Result
                  </th>
                  <th className="text-left py-3 px-6 font-semibold text-muted-foreground whitespace-nowrap">
                    Details
                  </th>
                </tr>
              </thead>
              <tbody>
                {filteredLogs.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="text-center py-8 text-muted-foreground">
                      No logs found
                    </td>
                  </tr>
                ) : (
                  filteredLogs.map(log => (
                    <tr key={log.id} className="border-b border-border hover:bg-muted/50 transition-colors">
                      <td className="py-3 px-6 font-mono text-xs whitespace-nowrap">
                        {log.timestamp.toLocaleString()}
                      </td>
                      <td className="py-3 px-6">
                        <code className="text-xs bg-muted px-2 py-1 rounded">
                          {log.actor}
                        </code>
                      </td>
                      <td className="py-3 px-6">
                        <span className="text-xs font-medium bg-primary/10 text-primary px-2 py-1 rounded">
                          {log.action.replace(/_/g, ' ')}
                        </span>
                      </td>
                      <td className="py-3 px-6 truncate max-w-sm">{log.target}</td>
                      <td className="py-3 px-6">
                        {log.result === 'success' ? (
                          <span className="inline-flex items-center gap-1 text-xs bg-green-500/20 text-green-700 dark:text-green-400 px-2 py-1 rounded">
                            ✓ Success
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-xs bg-red-500/20 text-red-700 dark:text-red-400 px-2 py-1 rounded">
                            ✗ Failure
                          </span>
                        )}
                      </td>
                      <td className="py-3 px-6 text-muted-foreground truncate max-w-md">
                        {log.details}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </Card>

        {/* Pagination Info */}
        <div className="text-sm text-muted-foreground text-center">
          Showing {filteredLogs.length} of {logs.length} entries
        </div>

        {/* Help Section */}
        <Card className="bg-green-500/10 border-green-200 dark:border-green-800 p-6 space-y-3">
          <h3 className="font-semibold text-sm">Audit Log Features</h3>
          <ul className="text-sm space-y-1 text-foreground/80">
            <li>• Immutable records ensure data integrity</li>
            <li>• 90-day retention for compliance and forensics</li>
            <li>• Searchable by actor, action, target, or time</li>
            <li>• All administrative actions are tracked</li>
            <li>• Real-time logging of all platform events</li>
          </ul>
        </Card>
      </div>
    </AppShell>
  )
}
