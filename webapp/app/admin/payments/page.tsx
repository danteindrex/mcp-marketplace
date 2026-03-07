'use client'

import { useEffect, useState } from 'react'
import { AppShell } from '@/components/app-shell'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  fetchAdminFeePolicies,
  fetchAdminPaymentsOverview,
  fetchAdminPayoutProfiles,
  fetchAdminPayouts,
  fetchAdminReconciliation,
  runAdminPayout,
  updateAdminPayoutBlock,
  upsertAdminFeePolicy,
} from '@/lib/api-client'

export default function AdminPaymentsPage() {
  const [overview, setOverview] = useState<any>({
    intentCount: 0,
    settledCount: 0,
    pendingCount: 0,
    failedCount: 0,
    settledVolumeUsdc: 0,
    platformRevenueUsdc: 0,
    byMethod: {},
    x402: {},
    methods: [],
    payouts: { count: 0, byStatus: {} },
  })
  const [policyData, setPolicyData] = useState<any>({ default: null, items: [] })
  const [payoutProfiles, setPayoutProfiles] = useState<any[]>([])
  const [recon, setRecon] = useState<any>({ imbalances: [] })
  const [payoutResult, setPayoutResult] = useState<any>(null)
  const [isSavingPolicy, setIsSavingPolicy] = useState(false)
  const [isRunningPayout, setIsRunningPayout] = useState(false)
  const [policyForm, setPolicyForm] = useState({
    scope: 'global',
    tenantId: '',
    serverId: '',
    platformFeeBps: 1000,
    minFeeUsdc: 0,
    maxFeeUsdc: 0,
    holdDays: 0,
    payoutCadence: 'manual',
    enabled: true,
  })
  const [payoutRun, setPayoutRun] = useState({
    tenantId: '',
    amountUsdc: 0,
    method: 'stablecoin',
    dryRun: false,
    force: false,
  })

  async function loadData() {
    const [o, p, pp, po, r] = await Promise.all([
      fetchAdminPaymentsOverview(),
      fetchAdminFeePolicies(),
      fetchAdminPayoutProfiles(),
      fetchAdminPayouts(),
      fetchAdminReconciliation(),
    ])
    setOverview({ ...(o as any), payouts: (po as any) || { count: 0, byStatus: {} } })
    setPolicyData(p as any)
    setPayoutProfiles((pp as any).items || [])
    setRecon(r as any)
    const globalPolicy = ((p as any).items || []).find((x: any) => x.scope === 'global') || (p as any).default
    if (globalPolicy) {
      setPolicyForm(prev => ({
        ...prev,
        platformFeeBps: Number(globalPolicy.platformFeeBps || 0),
        minFeeUsdc: Number(globalPolicy.minFeeUsdc || 0),
        maxFeeUsdc: Number(globalPolicy.maxFeeUsdc || 0),
        holdDays: Number(globalPolicy.holdDays || 0),
        payoutCadence: globalPolicy.payoutCadence || 'manual',
        enabled: globalPolicy.enabled !== false,
      }))
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  return (
    <AppShell role="admin">
      <div className="p-6 space-y-6">
        <div>
          <h1 className="text-3xl font-bold mb-2">Payments Control Center</h1>
          <p className="text-muted-foreground">Manage fee policy, payout execution, KYC readiness, and reconciliation.</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
          <Card className="p-5"><p className="text-xs text-muted-foreground">Intents</p><p className="text-3xl font-bold">{overview.intentCount}</p></Card>
          <Card className="p-5"><p className="text-xs text-muted-foreground">Settled Volume (USDC)</p><p className="text-3xl font-bold">{Number(overview.settledVolumeUsdc || 0).toFixed(2)}</p></Card>
          <Card className="p-5"><p className="text-xs text-muted-foreground">Platform Revenue (USDC)</p><p className="text-3xl font-bold">{Number(overview.platformRevenueUsdc || 0).toFixed(2)}</p></Card>
          <Card className="p-5"><p className="text-xs text-muted-foreground">Payouts</p><p className="text-3xl font-bold">{overview.payouts?.count || 0}</p></Card>
          <Card className="p-5"><p className="text-xs text-muted-foreground">Reconciliation Alerts</p><p className="text-3xl font-bold">{(recon.imbalances || []).length}</p></Card>
        </div>

        <Card className="p-6 space-y-2">
          <h2 className="text-xl font-bold">x402 + Stripe Status</h2>
          <p className="text-sm">x402 mode: <span className="font-semibold">{overview.x402?.mode || 'mock'}</span></p>
          <p className="text-sm">Facilitator: <span className="font-mono text-xs">{overview.x402?.facilitatorUrl || 'not configured'}</span></p>
          <p className="text-sm">Stripe Connect configured: <span className="font-semibold">{overview.stripeConnect?.configured ? 'yes' : 'no'}</span></p>
        </Card>

        <Card className="p-6 space-y-4">
          <h2 className="text-xl font-bold">Global Fee Policy</h2>
          <div className="grid grid-cols-1 md:grid-cols-6 gap-3">
            <div><p className="text-xs text-muted-foreground mb-1">Platform Fee (bps)</p><Input type="number" value={policyForm.platformFeeBps} onChange={e => setPolicyForm(prev => ({ ...prev, platformFeeBps: Number(e.target.value || 0) }))} /></div>
            <div><p className="text-xs text-muted-foreground mb-1">Min Fee (USDC)</p><Input type="number" value={policyForm.minFeeUsdc} onChange={e => setPolicyForm(prev => ({ ...prev, minFeeUsdc: Number(e.target.value || 0) }))} /></div>
            <div><p className="text-xs text-muted-foreground mb-1">Max Fee (USDC)</p><Input type="number" value={policyForm.maxFeeUsdc} onChange={e => setPolicyForm(prev => ({ ...prev, maxFeeUsdc: Number(e.target.value || 0) }))} /></div>
            <div><p className="text-xs text-muted-foreground mb-1">Hold Days</p><Input type="number" value={policyForm.holdDays} onChange={e => setPolicyForm(prev => ({ ...prev, holdDays: Number(e.target.value || 0) }))} /></div>
            <div>
              <p className="text-xs text-muted-foreground mb-1">Cadence</p>
              <select value={policyForm.payoutCadence} onChange={e => setPolicyForm(prev => ({ ...prev, payoutCadence: e.target.value }))} className="px-3 py-2 rounded-md border border-input bg-background text-sm w-full">
                <option value="manual">manual</option>
                <option value="daily">daily</option>
                <option value="weekly">weekly</option>
              </select>
            </div>
            <div className="flex items-end">
              <label className="text-sm flex items-center gap-2">
                <input type="checkbox" checked={policyForm.enabled} onChange={e => setPolicyForm(prev => ({ ...prev, enabled: e.target.checked }))} />
                Enabled
              </label>
            </div>
          </div>
          <Button
            disabled={isSavingPolicy}
            onClick={async () => {
              setIsSavingPolicy(true)
              try {
                await upsertAdminFeePolicy(policyForm as any)
                await loadData()
              } finally {
                setIsSavingPolicy(false)
              }
            }}
          >
            {isSavingPolicy ? 'Saving...' : 'Save Policy'}
          </Button>
          <div className="text-xs text-muted-foreground">Policies configured: {(policyData.items || []).length}</div>
        </Card>

        <Card className="p-6 space-y-3">
          <h2 className="text-xl font-bold">Run Payout</h2>
          <div className="grid grid-cols-1 md:grid-cols-5 gap-3">
            <Input placeholder="tenant id" value={payoutRun.tenantId} onChange={e => setPayoutRun(prev => ({ ...prev, tenantId: e.target.value }))} />
            <Input type="number" placeholder="amount (optional)" value={payoutRun.amountUsdc} onChange={e => setPayoutRun(prev => ({ ...prev, amountUsdc: Number(e.target.value || 0) }))} />
            <select value={payoutRun.method} onChange={e => setPayoutRun(prev => ({ ...prev, method: e.target.value }))} className="px-3 py-2 rounded-md border border-input bg-background text-sm w-full">
              <option value="stablecoin">stablecoin</option>
              <option value="stripe_connect">stripe_connect</option>
            </select>
            <label className="text-sm flex items-center gap-2"><input type="checkbox" checked={payoutRun.dryRun} onChange={e => setPayoutRun(prev => ({ ...prev, dryRun: e.target.checked }))} />Dry Run</label>
            <label className="text-sm flex items-center gap-2"><input type="checkbox" checked={payoutRun.force} onChange={e => setPayoutRun(prev => ({ ...prev, force: e.target.checked }))} />Force</label>
          </div>
          <Button
            disabled={isRunningPayout || !payoutRun.tenantId}
            onClick={async () => {
              setIsRunningPayout(true)
              try {
                const out = await runAdminPayout({
                  tenantId: payoutRun.tenantId,
                  amountUsdc: payoutRun.amountUsdc > 0 ? payoutRun.amountUsdc : undefined,
                  method: payoutRun.method,
                  dryRun: payoutRun.dryRun,
                  force: payoutRun.force,
                })
                setPayoutResult(out)
                await loadData()
              } finally {
                setIsRunningPayout(false)
              }
            }}
          >
            {isRunningPayout ? 'Running...' : 'Execute'}
          </Button>
          {payoutResult ? <pre className="text-xs bg-muted p-3 rounded-md overflow-auto">{JSON.stringify(payoutResult, null, 2)}</pre> : null}
        </Card>

        <Card className="p-6">
          <h2 className="text-xl font-bold mb-3">Seller Payout Profiles</h2>
          <div className="space-y-3">
            {payoutProfiles.length === 0 ? <p className="text-sm text-muted-foreground">No payout profiles found.</p> : null}
            {payoutProfiles.map((item: any) => (
              <div key={item.profile?.tenantId} className="border border-border rounded-md p-3 flex flex-col md:flex-row md:items-center md:justify-between gap-3">
                <div>
                  <p className="font-semibold">{item.profile?.tenantId}</p>
                  <p className="text-xs text-muted-foreground">
                    method={item.profile?.preferredMethod} | kyc={item.profile?.kycStatus || 'pending'} | payable={Number(item.payableUsdc || 0).toFixed(2)} USDC
                  </p>
                  {item.profile?.payoutBlocked ? <p className="text-xs text-red-600">blocked: {item.profile?.payoutBlockReason || 'n/a'}</p> : null}
                </div>
                <Button
                  variant="outline"
                  onClick={async () => {
                    await updateAdminPayoutBlock(item.profile?.tenantId, !item.profile?.payoutBlocked, item.profile?.payoutBlocked ? '' : 'manual_admin_block')
                    await loadData()
                  }}
                >
                  {item.profile?.payoutBlocked ? 'Unblock' : 'Block'}
                </Button>
              </div>
            ))}
          </div>
        </Card>

        <Card className="p-6">
          <h2 className="text-xl font-bold mb-3">Reconciliation</h2>
          {(recon.imbalances || []).length === 0 ? (
            <p className="text-sm text-green-600">No ledger imbalances detected.</p>
          ) : (
            <div className="space-y-2">
              {(recon.imbalances || []).map((item: any) => (
                <div key={item.transactionId} className="border border-border rounded-md p-3 text-xs">
                  tx={item.transactionId} | debits={item.debits} | credits={item.credits} | delta={item.delta}
                </div>
              ))}
            </div>
          )}
        </Card>

        <Card className="p-6">
          <h2 className="text-xl font-bold mb-3">Supported Payment Methods</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {(overview.methods || []).map((method: any) => (
              <div key={method.id} className="border border-border rounded-md p-3">
                <p className="font-semibold">{method.displayName}</p>
                <p className="text-xs text-muted-foreground mt-1">{method.integration}</p>
                <p className="text-xs mt-1">{method.configured ? 'Configured' : 'Not configured'}</p>
              </div>
            ))}
          </div>
        </Card>
      </div>
    </AppShell>
  )
}

