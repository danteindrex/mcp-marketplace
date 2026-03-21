'use client'

import { useEffect, useState } from 'react'
import { AppShell } from '@/components/app-shell'
import { Card } from '@/components/ui/card'
import { Text } from '@/components/retroui/text'
import { fetchClientCompatibility } from '@/lib/api-client'

export default function ClientCompatibilityPage() {
  const [items, setItems] = useState<any[]>([])

  useEffect(() => {
    fetchClientCompatibility().then(setItems)
  }, [])

  return (
    <AppShell role="admin">
      <div className="p-6 space-y-6">
        <div>
          <Text variant="h3" className="mb-2">Client Compatibility</Text>
          <Text variant="body" className="text-muted-foreground">Cross-client MCP auth/install capability matrix.</Text>
        </div>

        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 border-b border-border"><tr><th className="text-left py-3 px-4">Client</th><th className="text-left py-3 px-4">DCR</th><th className="text-left py-3 px-4">CIMD</th><th className="text-left py-3 px-4">Interactive</th><th className="text-left py-3 px-4">Notes</th></tr></thead>
              <tbody>
                {items.map(item => (
                  <tr key={item.client} className="border-b border-border">
                    <td className="py-3 px-4 font-medium uppercase">{item.client}</td>
                    <td className="py-3 px-4">{item.supportsDCR ? 'Yes' : 'No'}</td>
                    <td className="py-3 px-4">{item.supportsCIMD ? 'Yes' : 'No'}</td>
                    <td className="py-3 px-4">{item.supportsInteractive ? 'Yes' : 'No'}</td>
                    <td className="py-3 px-4 text-muted-foreground">{item.notes}</td>
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
