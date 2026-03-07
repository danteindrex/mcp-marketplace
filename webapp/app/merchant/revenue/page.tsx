'use client'

import { useState, useEffect } from 'react'
import { TrendingUp, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { AppShell } from '@/components/app-shell'
import { TableToolbar } from '@/components/table-toolbar'
import {
  createMerchantStripeOnboardingLink,
  fetchMerchantPaymentsOverview,
  fetchMerchantPayoutProfile,
  fetchMerchantRevenue,
  fetchMerchantServerPaymentConfig,
  fetchMerchantServers,
  refreshMerchantStripeKYC,
  updateMerchantPayoutProfile,
  updateMerchantServerPaymentConfig,
} from '@/lib/api-client'
import { BarChart } from '@/components/retroui/charts/BarChart'
import { AreaChart } from '@/components/retroui/charts/AreaChart'
import { LineChart } from '@/components/retroui/charts/LineChart'

export default function RevenuePage() {
  const [dateRange, setDateRange] = useState('6m')
  const [searchQuery, setSearchQuery] = useState('')
  const [data, setData] = useState<any>({ totalRevenue: 0, totalCustomers: 0, servers: [], trend: [] })
  const [payments, setPayments] = useState<any>({ methodBreakdown: {}, byServer: [] })
  const [selectedServerId, setSelectedServerId] = useState('')
  const [serverConfig, setServerConfig] = useState<any>(null)
  const [isSavingConfig, setIsSavingConfig] = useState(false)
  const [payoutData, setPayoutData] = useState<any>(null)
  const [isSavingPayout, setIsSavingPayout] = useState(false)
  const [isRefreshingKYC, setIsRefreshingKYC] = useState(false)

  useEffect(() => {
    Promise.all([fetchMerchantRevenue(), fetchMerchantPaymentsOverview(), fetchMerchantServers(), fetchMerchantPayoutProfile()]).then(
      ([rev, pay, servers, payout]) => {
        setData(rev as any)
        setPayments(pay as any)
        setPayoutData(payout as any)
        if (servers.length > 0) {
          setSelectedServerId(servers[0].id)
        }
      },
    )
  }, [])

  useEffect(() => {
    if (!selectedServerId) return
    fetchMerchantServerPaymentConfig(selectedServerId).then(setServerConfig)
  }, [selectedServerId])

  const filteredServers = (data.servers || []).filter((s: any) => s.name.toLowerCase().includes(searchQuery.toLowerCase()))

  return (
    <AppShell role="merchant">
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <div><h1 className="text-3xl font-bold mb-2">Revenue Dashboard</h1><p className="text-muted-foreground">Track earnings from your MCP servers</p></div>
          <div className="flex gap-2">
            <select value={dateRange} onChange={e => setDateRange(e.target.value)} className="px-3 py-2 rounded-md border border-input bg-background text-sm"><option value="1m">Last Month</option><option value="3m">Last 3 Months</option><option value="6m">Last 6 Months</option></select>
            <Button variant="outline" size="sm"><Download className="w-4 h-4 mr-2" />Export</Button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">Total Revenue</p><p className="text-3xl font-bold">${Number(data.totalRevenue).toLocaleString()}</p><p className="text-xs text-green-600 dark:text-green-400 mt-2 flex items-center gap-1"><TrendingUp className="w-3 h-3" />+22% from last period</p></Card>
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">Active Customers</p><p className="text-3xl font-bold">{data.totalCustomers}</p></Card>
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">Avg Revenue / Server</p><p className="text-3xl font-bold">${data.servers?.length ? (Number(data.totalRevenue) / data.servers.length).toLocaleString() : 0}</p></Card>
          <Card className="p-6"><p className="text-sm text-muted-foreground mb-2">Payout Status</p><p className="text-3xl font-bold text-green-600 dark:text-green-400">Processed</p></Card>
        </div>

        <Card className="p-6">
          <h2 className="text-lg font-semibold mb-4">Revenue Trend</h2>
          <BarChart
            data={data.trend || []}
            index="month"
            categories={['subscriptions', 'perCall']}
            stacked
          />
        </Card>

        <Card className="p-6 space-y-4">
          <h2 className="text-lg font-semibold">x402 Payment Monitoring</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div><p className="text-xs text-muted-foreground">Settled Volume</p><p className="text-2xl font-bold">${Number(payments.totalSettledUsdc || 0).toFixed(2)}</p></div>
            <div><p className="text-xs text-muted-foreground">Platform Fee</p><p className="text-2xl font-bold">${Number(payments.totalPlatformFeeUsdc || 0).toFixed(2)}</p></div>
            <div><p className="text-xs text-muted-foreground">Net To Seller</p><p className="text-2xl font-bold">${Number(payments.totalNetUsdc || 0).toFixed(2)}</p></div>
            <div><p className="text-xs text-muted-foreground">Settled Intents</p><p className="text-2xl font-bold">{payments.settledCount || 0}</p></div>
            <div><p className="text-xs text-muted-foreground">Pending Intents</p><p className="text-2xl font-bold">{payments.pendingCount || 0}</p></div>
            <div><p className="text-xs text-muted-foreground">Seller Payable</p><p className="text-2xl font-bold">${Number(payments.sellerPayableUsdc || 0).toFixed(2)}</p></div>
          </div>
          <div>
            <p className="text-sm font-semibold mb-2">Method Breakdown</p>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-2">
              {Object.entries(payments.methodBreakdown || {}).map(([method, count]: [string, any]) => (
                <div key={method} className="border border-border rounded-md p-3">
                  <p className="font-semibold text-sm">{method}</p>
                  <p className="text-xs text-muted-foreground">{count} intents</p>
                </div>
              ))}
            </div>
          </div>
        </Card>

        {payoutData && (
          <Card className="p-6 space-y-4">
            <h2 className="text-lg font-semibold">Payout + KYC</h2>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
              <div>
                <p className="text-xs text-muted-foreground mb-1">Preferred Method</p>
                <select
                  value={payoutData.profile?.preferredMethod || 'stablecoin'}
                  onChange={e => setPayoutData((prev: any) => ({ ...prev, profile: { ...prev.profile, preferredMethod: e.target.value } }))}
                  className="px-3 py-2 rounded-md border border-input bg-background text-sm w-full"
                >
                  <option value="stablecoin">stablecoin</option>
                  <option value="stripe_connect">stripe_connect</option>
                </select>
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-1">USDC Wallet Address</p>
                <Input
                  value={payoutData.profile?.stablecoinAddress || ''}
                  onChange={e => setPayoutData((prev: any) => ({ ...prev, profile: { ...prev.profile, stablecoinAddress: e.target.value } }))}
                />
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-1">Min Payout (USDC)</p>
                <Input
                  type="number"
                  value={payoutData.profile?.minPayoutUsdc || 10}
                  onChange={e => setPayoutData((prev: any) => ({ ...prev, profile: { ...prev.profile, minPayoutUsdc: Number(e.target.value || 0) } }))}
                />
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-1">Hold Days</p>
                <Input
                  type="number"
                  value={payoutData.profile?.holdDays || 0}
                  onChange={e => setPayoutData((prev: any) => ({ ...prev, profile: { ...prev.profile, holdDays: Number(e.target.value || 0) } }))}
                />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div className="border border-border rounded-md p-3">
                <p className="text-xs text-muted-foreground">KYC Status</p>
                <p className="text-lg font-semibold">{payoutData.profile?.kycStatus || 'pending'}</p>
                <p className="text-xs text-muted-foreground mt-1">charges={String(Boolean(payoutData.profile?.kycChargesEnabled))} payouts={String(Boolean(payoutData.profile?.kycPayoutsEnabled))}</p>
              </div>
              <div className="border border-border rounded-md p-3">
                <p className="text-xs text-muted-foreground">Stripe Account</p>
                <p className="text-sm font-mono">{payoutData.profile?.stripeAccountId || 'not connected'}</p>
              </div>
              <div className="border border-border rounded-md p-3">
                <p className="text-xs text-muted-foreground">Current Payable</p>
                <p className="text-lg font-semibold">${Number(payoutData.payableUsdc || 0).toFixed(2)}</p>
              </div>
            </div>

            {Array.isArray(payoutData.profile?.kycCurrentlyDue) && payoutData.profile.kycCurrentlyDue.length > 0 ? (
              <div className="border border-border rounded-md p-3">
                <p className="text-xs text-muted-foreground mb-1">KYC Currently Due</p>
                <p className="text-xs">{payoutData.profile.kycCurrentlyDue.join(', ')}</p>
              </div>
            ) : null}

            <div className="flex flex-wrap gap-2">
              <Button
                disabled={isSavingPayout}
                onClick={async () => {
                  setIsSavingPayout(true)
                  try {
                    const updated = await updateMerchantPayoutProfile({
                      preferredMethod: payoutData.profile?.preferredMethod,
                      stablecoinAddress: payoutData.profile?.stablecoinAddress || '',
                      minPayoutUsdc: Number(payoutData.profile?.minPayoutUsdc || 0),
                      holdDays: Number(payoutData.profile?.holdDays || 0),
                    })
                    setPayoutData(updated)
                    const refreshed = await fetchMerchantPaymentsOverview()
                    setPayments(refreshed as any)
                  } finally {
                    setIsSavingPayout(false)
                  }
                }}
              >
                {isSavingPayout ? 'Saving...' : 'Save Payout Profile'}
              </Button>
              <Button
                variant="outline"
                onClick={async () => {
                  const out = await createMerchantStripeOnboardingLink()
                  if ((out as any)?.onboardingUrl) {
                    window.open((out as any).onboardingUrl, '_blank', 'noopener,noreferrer')
                  }
                  const latest = await fetchMerchantPayoutProfile()
                  setPayoutData(latest)
                }}
              >
                Connect Stripe
              </Button>
              <Button
                variant="outline"
                disabled={isRefreshingKYC}
                onClick={async () => {
                  setIsRefreshingKYC(true)
                  try {
                    await refreshMerchantStripeKYC()
                    const latest = await fetchMerchantPayoutProfile()
                    setPayoutData(latest)
                  } finally {
                    setIsRefreshingKYC(false)
                  }
                }}
              >
                {isRefreshingKYC ? 'Refreshing...' : 'Refresh KYC'}
              </Button>
            </div>

            <div>
              <p className="text-sm font-semibold mb-2">Recent Payouts</p>
              <div className="space-y-2">
                {(payoutData.recentPayouts || []).slice(0, 5).map((p: any) => (
                  <div key={p.id} className="border border-border rounded-md p-3 text-xs">
                    {p.id} | {p.method} | {p.status} | {Number(p.amountUsdc || 0).toFixed(2)} USDC
                  </div>
                ))}
                {(payoutData.recentPayouts || []).length === 0 ? <p className="text-sm text-muted-foreground">No payouts yet.</p> : null}
              </div>
            </div>
          </Card>
        )}

        {serverConfig && (
          <Card className="p-6 space-y-4">
            <div className="flex flex-col md:flex-row md:items-end gap-3">
              <div className="min-w-48">
                <p className="text-xs text-muted-foreground mb-1">Server</p>
                <select
                  value={selectedServerId}
                  onChange={e => setSelectedServerId(e.target.value)}
                  className="px-3 py-2 rounded-md border border-input bg-background text-sm w-full"
                >
                  {(data.servers || []).map((s: any) => (
                    <option key={s.id} value={s.id}>{s.name}</option>
                  ))}
                </select>
              </div>
              <div className="grid grid-cols-3 gap-2 w-full">
                <div><p className="text-xs text-muted-foreground mb-1">Per Call Cap</p><Input type="number" value={serverConfig.config?.perCallCapUsdc || 0} onChange={e => setServerConfig((prev: any) => ({ ...prev, config: { ...prev.config, perCallCapUsdc: Number(e.target.value || 0) } }))} /></div>
                <div><p className="text-xs text-muted-foreground mb-1">Daily Cap</p><Input type="number" value={serverConfig.config?.dailyCapUsdc || 0} onChange={e => setServerConfig((prev: any) => ({ ...prev, config: { ...prev.config, dailyCapUsdc: Number(e.target.value || 0) } }))} /></div>
                <div><p className="text-xs text-muted-foreground mb-1">Monthly Cap</p><Input type="number" value={serverConfig.config?.monthlyCapUsdc || 0} onChange={e => setServerConfig((prev: any) => ({ ...prev, config: { ...prev.config, monthlyCapUsdc: Number(e.target.value || 0) } }))} /></div>
              </div>
            </div>
            <div>
              <p className="text-xs text-muted-foreground mb-1">Payment Address</p>
              <Input value={serverConfig.config?.paymentAddress || ''} onChange={e => setServerConfig((prev: any) => ({ ...prev, config: { ...prev.config, paymentAddress: e.target.value } }))} />
            </div>
            <Button
              disabled={isSavingConfig}
              onClick={async () => {
                setIsSavingConfig(true)
                try {
                  const updated = await updateMerchantServerPaymentConfig(selectedServerId, {
                    paymentAddress: serverConfig.config?.paymentAddress || '',
                    perCallCapUsdc: Number(serverConfig.config?.perCallCapUsdc || 0),
                    dailyCapUsdc: Number(serverConfig.config?.dailyCapUsdc || 0),
                    monthlyCapUsdc: Number(serverConfig.config?.monthlyCapUsdc || 0),
                    paymentMethods: serverConfig.config?.paymentMethods || [],
                  })
                  setServerConfig(updated)
                } finally {
                  setIsSavingConfig(false)
                }
              }}
            >
              {isSavingConfig ? 'Saving...' : 'Save Payment Config'}
            </Button>
          </Card>
        )}

        <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
          <Card className="p-6">
            <h2 className="text-lg font-semibold mb-4">Cumulative Revenue</h2>
            <AreaChart
              data={(data.trend || []).map((row: any, idx: number) => ({
                ...row,
                total: Number(row.subscriptions || 0) + Number(row.perCall || 0) + idx * 40,
              }))}
              index="month"
              categories={['total']}
            />
          </Card>

          <Card className="p-6">
            <h2 className="text-lg font-semibold mb-4">Subscriptions vs Per-Call</h2>
            <LineChart
              data={data.trend || []}
              index="month"
              categories={['subscriptions', 'perCall']}
            />
          </Card>
        </div>

        <div className="space-y-4">
          <h2 className="text-lg font-semibold mb-4">Revenue by Server</h2>
          <TableToolbar searchPlaceholder="Search by server name..." onSearch={setSearchQuery} />
          <div className="space-y-2">
            {filteredServers.map((server: any) => (
              <Card key={server.id} className="p-6">
                <div className="grid grid-cols-1 md:grid-cols-5 gap-4 items-center">
                  <div><h3 className="font-semibold">{server.name}</h3><p className="text-sm text-muted-foreground">{server.customers} customers</p></div>
                  <div><p className="text-sm text-muted-foreground mb-1">Revenue</p><p className="text-2xl font-bold">${Number(server.revenue).toLocaleString()}</p></div>
                  <div><p className="text-sm text-muted-foreground mb-1">Trend</p><p className="text-lg font-semibold text-green-600 dark:text-green-400">{server.trend}</p></div>
                  <div><p className="text-sm text-muted-foreground mb-1">Avg / Customer</p><p className="text-lg font-semibold">${server.customers ? (Number(server.revenue) / server.customers).toFixed(2) : '0.00'}</p></div>
                  <Button variant="outline" size="sm" className="w-full md:w-auto">View Details</Button>
                </div>
              </Card>
            ))}
          </div>
        </div>
      </div>
    </AppShell>
  )
}
