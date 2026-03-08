'use client'

import { useState, useEffect, use, useCallback, useMemo } from 'react'
import Link from 'next/link'
import { ArrowLeft, Check, ChevronRight, Copy, ExternalLink, Check as CheckIcon, Terminal } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Label } from '@/components/ui/label'
import { toast } from 'sonner'
import {
  fetchBuyerPaymentControls,
  fetchServerBySlug,
  installMarketplaceServer,
  settleX402Intent,
  checkInstallScopes,
  type BuyerPaymentControls,
  type InstallPaymentRequired,
  type InstallAction,
  type InstallSession,
  type ScopeCheckResult,
  type Server,
} from '@/lib/api-client'
import {
  getPaymentMethodDescription,
  getPaymentMethodStatus,
  getPaymentMethodTitle,
  isInstallPaymentMethod,
} from '@/lib/payment-methods'

interface PageProps {
  params: Promise<{ serverId: string }>
}

type Step = 'client' | 'auth' | 'scopes' | 'connect' | 'complete'

const steps: Array<{ id: Step; label: string; title: string }> = [
  { id: 'client', label: 'Client', title: 'Select Your Client' },
  { id: 'auth', label: 'Auth', title: 'Authentication Ready?' },
  { id: 'scopes', label: 'Scopes', title: 'Review Permissions' },
  { id: 'connect', label: 'Connect', title: 'Complete Setup' },
  { id: 'complete', label: 'Done', title: 'Installation Complete' },
]

const clientOptions = [
  { value: 'vscode', label: 'VS Code', description: 'Full support with native extension' },
  { value: 'cursor', label: 'Cursor', description: 'Built-in MCP support' },
  { value: 'claude', label: 'Claude', description: 'Desktop application' },
  { value: 'codex', label: 'OpenAI Codex', description: 'CLI MCP install flow' },
  { value: 'chatgpt', label: 'ChatGPT', description: 'Connector setup for remote MCP' },
]

type MetadataState =
  | { status: 'idle' | 'loading'; data?: undefined; error?: undefined }
  | { status: 'ready'; data: InstallReadinessResponse; error?: undefined }
  | { status: 'error'; data?: InstallReadinessResponse; error: string }

interface InstallReadinessResponse {
  baseUrl: string
  timestamp: string
  cimd: {
    status: 'ok' | 'error'
    httpStatus: number
    url: string
    error?: string
    json?: Record<string, any> | null
  }
  oauth: {
    status: 'ok' | 'error'
    httpStatus: number
    url: string
    error?: string
    json?: Record<string, any> | null
  }
  jwks: null | {
    status: 'ok' | 'error'
    httpStatus: number
    url: string
    error?: string
    json?: Record<string, any> | null
  }
  links: {
    cimdUrl: string
    oauthMetadataUrl: string
    jwksUrl: string | null
  }
}

type ScopeCheckState =
  | { status: 'idle' | 'loading'; result?: undefined; error?: undefined }
  | { status: 'success'; result: ScopeCheckResult; error?: undefined }
  | { status: 'error'; result?: ScopeCheckResult; error: string }

export default function InstallWizardPage({ params }: PageProps) {
  const { serverId } = use(params)
  const [currentStep, setCurrentStep] = useState<Step>('client')
  const [selectedClient, setSelectedClient] = useState<string>('')
  const [authConfirmed, setAuthConfirmed] = useState(false)
  const [manualAuthOverride, setManualAuthOverride] = useState(false)
  const [acceptedScopes, setAcceptedScopes] = useState(false)
  const [copiedCode, setCopiedCode] = useState(false)
  const [showBridgeHelp, setShowBridgeHelp] = useState(false)
  const [installing, setInstalling] = useState(false)
  const [installSession, setInstallSession] = useState<InstallSession | null>(null)
  const [paymentRequired, setPaymentRequired] = useState<InstallPaymentRequired | null>(null)
  const [selectedPaymentMethod, setSelectedPaymentMethod] = useState<string>('')
  const [externalPaymentID, setExternalPaymentID] = useState('')
  const [server, setServer] = useState<Server | null>(null)
  const [paymentControls, setPaymentControls] = useState<BuyerPaymentControls | null>(null)
  const [paymentControlsLoading, setPaymentControlsLoading] = useState(true)
  const [paymentControlsError, setPaymentControlsError] = useState('')
  const [metadataState, setMetadataState] = useState<MetadataState>({ status: 'idle' })
  const [scopeCheckState, setScopeCheckState] = useState<ScopeCheckState>({ status: 'idle' })
  const parsedPaymentChallenge = useMemo(() => {
    if (!paymentRequired?.paymentChallenge) return null
    try {
      return JSON.parse(paymentRequired.paymentChallenge)
    } catch {
      return paymentRequired.paymentChallenge
    }
  }, [paymentRequired?.paymentChallenge])

  useEffect(() => {
    fetchServerBySlug(serverId).then(setServer)
  }, [serverId])

  useEffect(() => {
    const loadPaymentControls = async () => {
      try {
        setPaymentControlsLoading(true)
        setPaymentControlsError('')
        const controls = await fetchBuyerPaymentControls()
        setPaymentControls(controls)
      } catch (error) {
        setPaymentControlsError(error instanceof Error ? error.message : 'Unable to load buyer payment controls')
      } finally {
        setPaymentControlsLoading(false)
      }
    }

    loadPaymentControls()
  }, [])

  useEffect(() => {
    setScopeCheckState({ status: 'idle' })
  }, [selectedClient, server?.id])

  const refreshInstallReadiness = useCallback(async () => {
    if (!server) {
      setMetadataState({ status: 'idle' })
      return
    }
    if (!server.canonicalResourceUri) {
      setMetadataState({ status: 'error', error: 'Server is missing a canonical resource URI.' })
      return
    }
    try {
      setMetadataState({ status: 'loading' })
      const params = new URLSearchParams({ resource: server.canonicalResourceUri })
      const res = await fetch(`/api/install/readiness?${params.toString()}`, { cache: 'no-store' })
      const data = await res.json()
      if (!res.ok) {
        throw new Error(data?.error || 'Unable to fetch install metadata')
      }
      setMetadataState({ status: 'ready', data })
      setManualAuthOverride(false)
    } catch (error) {
      setMetadataState({
        status: 'error',
        error: error instanceof Error ? error.message : 'Failed to verify install metadata',
      })
    }
  }, [server])

  useEffect(() => {
    refreshInstallReadiness()
  }, [refreshInstallReadiness])

  const validateScopes = useCallback(async () => {
    if (!server) {
      toast.error('Server data unavailable')
      return false
    }
    if (!selectedClient) {
      toast.error('Select a client before validating scopes')
      return false
    }
    try {
      setScopeCheckState({ status: 'loading' })
      const result = await checkInstallScopes(server.slug, {
        client: selectedClient,
        grantedScopes: server.requiredScopes,
      })
      setScopeCheckState({ status: 'success', result })
      return true
    } catch (error: any) {
      const message = error?.message || 'Scope validation failed'
      setScopeCheckState({ status: 'error', error: message })
      toast.error(message)
      return false
    }
  }, [server, selectedClient])

  const currentStepIndex = steps.findIndex(s => s.id === currentStep)
  const currentStepData = steps[currentStepIndex]
  const selectedAction: InstallAction | null = installSession?.install?.selected || null
  const allowedMethods = paymentControls?.policy?.allowedMethods || []
  const installPaymentMethods = (paymentControls?.methods || []).filter(method => isInstallPaymentMethod(method.id))
  const readyInstallMethods = installPaymentMethods.filter(
    method => getPaymentMethodStatus(method, allowedMethods).selectable,
  )
  const paymentRequiredAheadOfInstall = Boolean(scopeCheckState.result?.paymentRequired)
  const bridgeInstallCommand = 'powershell -ExecutionPolicy Bypass -File backend\\scripts\\install-local-bridge.ps1'
  const installActionLabel = (() => {
    if (selectedAction?.requiresLocalExec && selectedAction?.launchUrl) return 'Run One-Click Install'
    if (selectedAction?.requiresLocalExec) return 'Run Install Command'
    return 'Open Installer'
  })()

  const installToolName = `install_${server?.slug?.replace(/-/g, '_') || 'server'}`

  useEffect(() => {
    if (!paymentControls) return
    const defaultMethod = readyInstallMethods[0]?.id || ''
    if (!selectedPaymentMethod || !readyInstallMethods.some(method => method.id === selectedPaymentMethod)) {
      setSelectedPaymentMethod(defaultMethod)
    }
  }, [paymentControls, readyInstallMethods, selectedPaymentMethod])

  if (!server) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen p-4">
        <h1 className="text-2xl font-bold mb-4">Server not found</h1>
        <Button asChild><Link href="/marketplace">Back to Marketplace</Link></Button>
      </div>
    )
  }

  const runInstallAttempt = async (autoSettle: boolean) => {
    try {
      setInstalling(true)
      const grantedScopesToUse = scopeCheckState.result?.grantedScopes?.length
        ? scopeCheckState.result.grantedScopes
        : server.requiredScopes
      const out = await installMarketplaceServer(server.slug, {
        client: selectedClient,
        grantedScopes: grantedScopesToUse,
        paymentMethod: selectedPaymentMethod || undefined,
        autoSettle,
        toolName: installToolName,
      })
      if (out.type === 'payment_required') {
        setPaymentRequired(out.payment)
        return false
      }
      setPaymentRequired(null)
      setInstallSession(out.session)
      setCurrentStep('complete')
      return true
    } catch (error: any) {
      toast.error(error?.message || 'Failed to prepare install')
      return false
    } finally {
      setInstalling(false)
    }
  }

  const handleNext = async () => {
    if (currentStep === 'connect') {
      const paidAndInstalled = await runInstallAttempt(selectedPaymentMethod === 'wallet_balance')
      if (!paidAndInstalled && selectedPaymentMethod === 'wallet_balance') {
        toast.info('Payment required. You can fund wallet or retry with external x402 payment.')
      }
      return
    }
    if (currentStep === 'scopes') {
      const ok = await validateScopes()
      if (!ok) return
    }
    if (currentStep === 'complete') {
      window.location.href = '/buyer/connections'
      return
    }
    const nextStep = steps[currentStepIndex + 1]?.id
    if (nextStep) setCurrentStep(nextStep)
  }

  const handleBack = () => {
    const prevStep = steps[currentStepIndex - 1]?.id
    if (prevStep) setCurrentStep(prevStep)
  }

  const canProceed = () => {
    switch (currentStep) {
      case 'client':
        return selectedClient !== ''
      case 'auth':
        return (metadataState.status === 'ready' || manualAuthOverride) && authConfirmed
      case 'scopes':
        return acceptedScopes
      case 'connect':
        return !paymentRequiredAheadOfInstall || Boolean(selectedPaymentMethod)
      default:
        return true
    }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    setCopiedCode(true)
    toast.success('Copied to clipboard')
    setTimeout(() => setCopiedCode(false), 2000)
  }

  const runInstallAction = (action: InstallAction) => {
    setShowBridgeHelp(false)
    if (action.launchUrl) {
      window.location.href = action.launchUrl
      if (action.requiresLocalExec && action.launchUrl.startsWith('mcp-marketplace://')) {
        window.setTimeout(() => {
          if (!document.hidden) {
            setShowBridgeHelp(true)
            toast.info('If nothing opened, install MCP Local Bridge once')
          }
        }, 1600)
      }
      return
    }
    if (action.openUrl) {
      window.open(action.openUrl, '_blank', 'noopener,noreferrer')
      return
    }
    if (action.command) {
      copyToClipboard(action.command)
      toast.info('Run the copied command in your local terminal')
      return
    }
    if (action.fallbackCopy) {
      copyToClipboard(action.fallbackCopy)
    }
  }

  return (
    <div className="flex flex-col min-h-screen bg-background">
      <div className="border-b border-border bg-background/50 backdrop-blur sticky top-0 z-40">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-6 flex items-center justify-between">
          <Link href={`/marketplace/${server.slug}`} className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors"><ArrowLeft className="w-4 h-4" />Back</Link>
          <h1 className="text-xl font-bold">{server.name}</h1><div />
        </div>
      </div>

      <main className="flex-1 max-w-3xl mx-auto w-full px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          <div className="lg:col-span-1">
            <div className="space-y-3">
              {steps.map((step, index) => {
                const isCompleted = currentStepIndex > index
                const isCurrent = step.id === currentStep
                return (
                  <button key={step.id} className={`w-full text-left px-4 py-3 rounded-lg transition-colors flex items-center gap-3 ${isCurrent ? 'bg-primary text-primary-foreground' : isCompleted ? 'bg-green-500/20 text-green-700 dark:text-green-400' : 'bg-muted text-muted-foreground'}`}>
                    <div className={`w-6 h-6 rounded-full flex items-center justify-center text-sm font-medium ${isCurrent || isCompleted ? 'bg-current' : 'border border-current'}`}>{isCompleted ? <Check className="w-4 h-4" /> : index + 1}</div>
                    <div><p className="text-xs font-medium">{step.label}</p></div>
                  </button>
                )
              })}
            </div>
          </div>

          <div className="lg:col-span-3">
            <Card className="p-8">
              <h2 className="text-2xl font-bold mb-6">{currentStepData.title}</h2>

              {currentStep === 'client' && (
                <RadioGroup value={selectedClient} onValueChange={setSelectedClient}>
                  {clientOptions.map(client => (
                    <div key={client.value} className="flex items-center space-x-3 p-4 border border-border rounded-lg hover:bg-muted/50">
                      <RadioGroupItem value={client.value} id={client.value} />
                      <Label htmlFor={client.value} className="flex-1 cursor-pointer"><p className="font-medium">{client.label}</p><p className="text-sm text-muted-foreground">{client.description}</p></Label>
                    </div>
                  ))}
                </RadioGroup>
              )}

              {currentStep === 'auth' && (
                <div className="space-y-5">
                  <Card className="p-4 space-y-4 border border-border">
                    <div className="flex items-center justify-between gap-3">
                      <div>
                        <p className="font-semibold">CIMD Metadata</p>
                        <p className="text-xs text-muted-foreground break-all">
                          {metadataState.status === 'ready' ? metadataState.data.links.cimdUrl : server.canonicalResourceUri || 'unknown resource'}
                        </p>
                      </div>
                      <Badge variant={metadataState.status === 'ready' && metadataState.data.cimd.status === 'ok' ? 'default' : metadataState.status === 'loading' ? 'outline' : 'destructive'}>
                        {metadataState.status === 'ready' && metadataState.data.cimd.status === 'ok'
                          ? 'Ready'
                          : metadataState.status === 'loading'
                            ? 'Checking'
                            : 'Error'}
                      </Badge>
                    </div>
                    {metadataState.status === 'ready' && metadataState.data.cimd.json?.scopes_supported && (
                      <p className="text-xs text-muted-foreground">
                        Scopes exposed: {metadataState.data.cimd.json.scopes_supported.join(', ')}
                      </p>
                    )}
                    {metadataState.status === 'error' && (
                      <p className="text-sm text-destructive">
                        {metadataState.error}
                      </p>
                    )}
                    <div className="grid gap-3 md:grid-cols-2">
                      <div className="p-3 rounded-md bg-muted/60">
                        <p className="text-xs uppercase text-muted-foreground">OAuth Metadata</p>
                        <p className="text-sm font-medium break-all">
                          {metadataState.status === 'ready' ? metadataState.data.links.oauthMetadataUrl : 'Pending'}
                        </p>
                        <Badge className="mt-2" variant={metadataState.status === 'ready' && metadataState.data.oauth.status === 'ok' ? 'default' : metadataState.status === 'loading' ? 'outline' : 'destructive'}>
                          {metadataState.status === 'ready' && metadataState.data.oauth.status === 'ok'
                            ? 'Reachable'
                            : metadataState.status === 'loading'
                              ? 'Checking'
                              : 'Error'}
                        </Badge>
                      </div>
                      <div className="p-3 rounded-md bg-muted/60">
                        <p className="text-xs uppercase text-muted-foreground">JWKS</p>
                        <p className="text-sm font-medium break-all">
                          {metadataState.status === 'ready' && metadataState.data.links.jwksUrl ? metadataState.data.links.jwksUrl : 'Awaiting metadata'}
                        </p>
                        <Badge className="mt-2" variant={metadataState.status === 'ready' && metadataState.data.jwks && metadataState.data.jwks.status === 'ok' ? 'default' : metadataState.status === 'loading' ? 'outline' : 'destructive'}>
                          {metadataState.status === 'ready' && metadataState.data.jwks?.status === 'ok'
                            ? `${Array.isArray(metadataState.data.jwks?.json?.keys) ? metadataState.data.jwks?.json?.keys.length : 0} keys`
                            : metadataState.status === 'loading'
                              ? 'Checking'
                              : 'Missing'}
                        </Badge>
                      </div>
                    </div>
                    <div className="flex flex-wrap gap-3">
                      <Button variant="outline" size="sm" onClick={refreshInstallReadiness} disabled={metadataState.status === 'loading'}>
                        {metadataState.status === 'loading' ? 'Checking...' : 'Check Metadata'}
                      </Button>
                      {metadataState.status === 'ready' && (
                        <p className="text-xs text-muted-foreground">
                          Last checked {new Date(metadataState.data.timestamp).toLocaleTimeString()}
                        </p>
                      )}
                    </div>
                  </Card>

                  <div className="flex items-center space-x-3 p-4 border border-border rounded-lg">
                    <Checkbox id="auth-ready" checked={authConfirmed} onCheckedChange={state => setAuthConfirmed(state === true)} />
                    <Label htmlFor="auth-ready" className="flex-1 cursor-pointer">
                      <p className="font-medium">I can complete OAuth authorization using the metadata above.</p>
                    </Label>
                  </div>

                  {metadataState.status === 'error' && (
                    <div className="flex items-center space-x-3 p-4 border border-border rounded-lg bg-amber-500/10">
                      <Checkbox id="auth-override" checked={manualAuthOverride} onCheckedChange={state => setManualAuthOverride(state === true)} />
                      <Label htmlFor="auth-override" className="flex-1 cursor-pointer">
                        <p className="font-medium">Allow manual override if CIMD metadata blocks automated checks.</p>
                        <p className="text-xs text-muted-foreground">Only enable this if you verified the metadata manually.</p>
                      </Label>
                    </div>
                  )}
                </div>
              )}

              {currentStep === 'scopes' && (
                <div className="space-y-6">
                  <div className="space-y-3 bg-muted/50 p-4 rounded-lg">
                    {server.requiredScopes.map(scope => (
                      <div key={scope} className="flex items-center gap-3"><CheckIcon className="w-5 h-5 text-green-500 flex-shrink-0" /><p className="font-medium text-sm">{scope}</p></div>
                    ))}
                  </div>
                  <div className="flex items-center space-x-3 p-4 border border-border rounded-lg">
                    <Checkbox id="scopes-accepted" checked={acceptedScopes} onCheckedChange={state => setAcceptedScopes(state === true)} />
                    <Label htmlFor="scopes-accepted" className="flex-1 cursor-pointer"><p className="font-medium">I accept these permissions</p></Label>
                  </div>
                  <div className="flex flex-wrap gap-3 items-center">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={validateScopes}
                      disabled={scopeCheckState.status === 'loading'}
                    >
                      {scopeCheckState.status === 'loading' ? 'Validating…' : 'Validate With Server'}
                    </Button>
                    {scopeCheckState.status === 'success' && (
                      <Badge variant={scopeCheckState.result.paymentRequired ? 'destructive' : 'default'}>
                        {scopeCheckState.result.paymentRequired ? 'Payment required' : 'Scopes confirmed'}
                      </Badge>
                    )}
                  </div>
                  {scopeCheckState.status === 'success' && (
                    <Card className="p-4 space-y-2">
                      <p className="text-sm text-muted-foreground">
                        Entitlement status: <span className="font-medium text-foreground">{scopeCheckState.result.entitlementStatus}</span>
                      </p>
                      {scopeCheckState.result.autoGrantAvailable && (
                        <p className="text-sm text-muted-foreground">
                          This free server will auto-grant an entitlement during install.
                        </p>
                      )}
                      {scopeCheckState.result.paymentRequired && (
                        <p className="text-sm text-amber-600 dark:text-amber-400">
                          You will be prompted to settle payment before install completes.
                        </p>
                      )}
                      <p className="text-xs text-muted-foreground">Validated scopes: {scopeCheckState.result.grantedScopes.join(', ')}</p>
                    </Card>
                  )}
                  {scopeCheckState.status === 'error' && (
                    <p className="text-sm text-destructive">{scopeCheckState.error}</p>
                  )}
                </div>
              )}

              {currentStep === 'connect' && (
                <div className="space-y-6">
                  <Card className="bg-muted p-6 border-border space-y-4">
                    <h3 className="font-semibold">One-Click Install Ready</h3>
                    <p className="text-sm text-muted-foreground">
                      Next will create your server connection and generate a client-specific install action.
                    </p>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                      <div>
                        <p className="text-muted-foreground">Client</p>
                        <p className="font-medium uppercase">{selectedClient}</p>
                      </div>
                      <div>
                        <p className="text-muted-foreground">Permissions</p>
                        <p className="font-medium">{server.requiredScopes.length} scopes</p>
                      </div>
                    </div>
                    <div className="space-y-2 pt-2 border-t border-border">
                      <p className="text-xs text-muted-foreground">Install payment methods from buyer payment controls</p>
                      {paymentControlsLoading ? (
                        <p className="text-sm text-muted-foreground">Loading payment readiness...</p>
                      ) : paymentControlsError ? (
                        <p className="text-sm text-destructive">{paymentControlsError}</p>
                      ) : installPaymentMethods.length === 0 ? (
                        <p className="text-sm text-muted-foreground">No install-capable payment methods were returned by the API.</p>
                      ) : (
                        <RadioGroup value={selectedPaymentMethod} onValueChange={setSelectedPaymentMethod}>
                          {installPaymentMethods.map(method => {
                            const readiness = getPaymentMethodStatus(method, allowedMethods)
                            const disabled = !readiness.selectable

                            return (
                              <label
                                key={method.id}
                                htmlFor={method.id}
                                className={`flex items-start gap-3 rounded-lg border border-border p-4 ${disabled ? 'opacity-70' : 'cursor-pointer'}`}
                              >
                                <RadioGroupItem value={method.id} id={method.id} disabled={disabled} />
                                <div className="flex-1 space-y-2">
                                  <div className="flex items-center gap-2 flex-wrap">
                                    <p className="font-medium">{getPaymentMethodTitle(method)}</p>
                                    <Badge variant={readiness.tone}>{readiness.label}</Badge>
                                  </div>
                                  <p className="text-sm text-muted-foreground">{getPaymentMethodDescription(method)}</p>
                                  <p className="text-xs text-muted-foreground">{readiness.description}</p>
                                </div>
                              </label>
                            )
                          })}
                        </RadioGroup>
                      )}
                      {paymentRequiredAheadOfInstall && readyInstallMethods.length === 0 && !paymentControlsLoading && !paymentControlsError && (
                        <p className="text-xs text-amber-600 dark:text-amber-400">
                          Scope validation says payment will be required, but no allowed method is ready. Open billing to enable or fund one first.
                        </p>
                      )}
                    </div>
                  </Card>
                  {installing && (
                    <p className="text-sm text-muted-foreground">Preparing install session...</p>
                  )}
                  {paymentRequired && (
                    <Card className="p-6 space-y-4 border-2 border-amber-500/50">
                      <h3 className="font-semibold">Payment Required</h3>
                      <p className="text-sm text-muted-foreground">{paymentRequired.error || 'Settle payment to continue install.'}</p>
                      <p className="text-xs text-muted-foreground">
                        Intent: {paymentRequired.intent?.id} | Amount: {Number(paymentRequired.intent?.amountUsdc || 0).toFixed(2)} USDC
                      </p>
                      {selectedPaymentMethod === 'x402_wallet' && (
                        <div className="space-y-2">
                          <Label htmlFor="external-payment-id">External Payment Identifier</Label>
                          <input
                            id="external-payment-id"
                            value={externalPaymentID}
                            onChange={e => setExternalPaymentID(e.target.value)}
                            className="w-full px-3 py-2 rounded-md border border-input bg-background text-sm"
                            placeholder="tx hash or provider payment id"
                          />
                        </div>
                      )}
                      {paymentRequired.wwwAuthenticate && (
                        <div className="bg-background rounded-md p-3 font-mono text-xs overflow-x-auto">
                          <p className="text-[11px] uppercase text-muted-foreground mb-1">WWW-Authenticate</p>
                          <pre className="whitespace-pre-wrap break-all">{paymentRequired.wwwAuthenticate}</pre>
                        </div>
                      )}
                      {parsedPaymentChallenge && (
                        <div className="bg-background rounded-md p-3 font-mono text-xs overflow-x-auto space-y-1">
                          <p className="text-[11px] uppercase text-muted-foreground">PAYMENT-REQUIRED</p>
                          <pre className="whitespace-pre-wrap break-all">
                            {typeof parsedPaymentChallenge === 'string'
                              ? parsedPaymentChallenge
                              : JSON.stringify(parsedPaymentChallenge, null, 2)}
                          </pre>
                        </div>
                      )}
                      <div className="flex flex-wrap gap-2">
                        <Button
                          disabled={installing || (selectedPaymentMethod === 'x402_wallet' && !externalPaymentID)}
                          onClick={async () => {
                            if (!paymentRequired?.intent?.id) return
                            try {
                              setInstalling(true)
                              if (selectedPaymentMethod === 'wallet_balance') {
                                await settleX402Intent(paymentRequired.intent.id)
                              } else {
                                await settleX402Intent(paymentRequired.intent.id, {
                                  paymentResponse: {
                                    paymentIdentifier: externalPaymentID,
                                    method: 'x402_wallet',
                                  },
                                })
                              }
                              await runInstallAttempt(false)
                            } catch (error: any) {
                              toast.error(error?.message || 'Payment settlement failed')
                            } finally {
                              setInstalling(false)
                            }
                          }}
                        >
                          Pay And Continue
                        </Button>
                        <Button asChild variant="outline">
                          <Link href="/buyer/billing">Open Wallet Top-Up</Link>
                        </Button>
                      </div>
                    </Card>
                  )}
                </div>
              )}

              {currentStep === 'complete' && (
                <div className="text-center space-y-6">
                  <div className="w-16 h-16 rounded-full bg-green-500/20 flex items-center justify-center mx-auto"><Check className="w-8 h-8 text-green-600 dark:text-green-400" /></div>
                  <h3 className="text-2xl font-bold">Installation Session Created</h3>
                  <div className="flex flex-wrap gap-2 justify-center">{server.requiredScopes.map(scope => <Badge key={scope} variant="outline">{scope}</Badge>)}</div>

                  {selectedAction && (
                    <Card className="text-left p-6 space-y-4">
                      <p className="font-semibold">{selectedAction.label}</p>
                      {selectedAction.description && (
                        <p className="text-sm text-muted-foreground">{selectedAction.description}</p>
                      )}
                      <div className="flex flex-wrap gap-2">
                        <Button onClick={() => runInstallAction(selectedAction)}>
                          {selectedAction.requiresLocalExec ? (
                            <Terminal className="w-4 h-4 mr-2" />
                          ) : (
                            <ExternalLink className="w-4 h-4 mr-2" />
                          )}
                          {installActionLabel}
                        </Button>
                        {selectedAction.fallbackCopy && (
                          <Button
                            variant="outline"
                            onClick={() => copyToClipboard(selectedAction.fallbackCopy || '')}
                          >
                            <Copy className="w-4 h-4 mr-2" />
                            {copiedCode ? 'Copied!' : 'Copy Fallback'}
                          </Button>
                        )}
                      </div>
                      {selectedAction.command && (
                        <div className="bg-background rounded p-4 font-mono text-xs overflow-x-auto">
                          <pre className="text-foreground/80">{selectedAction.command}</pre>
                        </div>
                      )}
                    </Card>
                  )}

                  {showBridgeHelp && selectedAction?.requiresLocalExec && selectedAction.launchUrl?.startsWith('mcp-marketplace://') && (
                    <Card className="text-left p-6 space-y-4 border-dashed">
                      <p className="font-semibold">Install MCP Local Bridge (one-time)</p>
                      <p className="text-sm text-muted-foreground">
                        One-click local command execution needs the local bridge registered once on this machine.
                      </p>
                      <div className="bg-background rounded p-4 font-mono text-xs overflow-x-auto">
                        <pre className="text-foreground/80">{bridgeInstallCommand}</pre>
                      </div>
                      <div className="flex flex-wrap gap-2">
                        <Button variant="outline" onClick={() => copyToClipboard(bridgeInstallCommand)}>
                          <Copy className="w-4 h-4 mr-2" />
                          Copy Bridge Install Command
                        </Button>
                        <Button onClick={() => runInstallAction(selectedAction)}>
                          Retry One-Click Install
                        </Button>
                      </div>
                    </Card>
                  )}

                  {installSession?.connection?.id && (
                    <p className="text-sm text-muted-foreground">
                      Connection ID: {installSession.connection.id}
                    </p>
                  )}
                  <Button asChild>
                    <Link href="/buyer/connections">Go to Connections</Link>
                  </Button>
                </div>
              )}

              <div className="flex items-center justify-between gap-4 mt-8 pt-6 border-t border-border">
                <Button variant="outline" onClick={handleBack} disabled={currentStepIndex === 0}>Back</Button>
                <div className="text-sm text-muted-foreground">Step {currentStepIndex + 1} of {steps.length}</div>
                <Button onClick={handleNext} disabled={!canProceed() || installing}>{currentStep === 'complete' ? 'Done' : installing ? 'Preparing...' : 'Next'}{currentStep !== 'complete' && !installing && <ChevronRight className="ml-2 w-4 h-4" />}</Button>
              </div>
            </Card>
          </div>
        </div>
      </main>
    </div>
  )
}
