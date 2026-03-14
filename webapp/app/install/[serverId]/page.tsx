'use client'

import { useState, useEffect, use, useCallback, useMemo } from 'react'
import Link from 'next/link'
import { ArrowLeft, Check, ChevronRight, ExternalLink, Check as CheckIcon, Terminal } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Label } from '@/components/ui/label'
import { toast } from 'sonner'
import { Text } from '@/components/retroui/Text'
import {
  fetchBuyerHub,
  fetchBuyerPaymentControls,
  fetchServerDetailBySlug,
  installMarketplaceServer,
  settleX402Intent,
  checkInstallScopes,
  type BuyerPaymentControls,
  type InstallPaymentRequired,
  type InstallAction,
  type InstallSession,
  type MarketplaceInstallMetadata,
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
  const [showBridgeHelp, setShowBridgeHelp] = useState(false)
  const [autoLaunchAttempted, setAutoLaunchAttempted] = useState(false)
  const [installing, setInstalling] = useState(false)
  const [installSession, setInstallSession] = useState<InstallSession | null>(null)
  const [paymentRequired, setPaymentRequired] = useState<InstallPaymentRequired | null>(null)
  const [selectedPaymentMethod, setSelectedPaymentMethod] = useState<string>('')
  const [externalPaymentInput, setExternalPaymentInput] = useState('')
  const [server, setServer] = useState<Server | null>(null)
  const [installMetadata, setInstallMetadata] = useState<MarketplaceInstallMetadata | null>(null)
  const [hubResource, setHubResource] = useState<string>('')
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
    fetchServerDetailBySlug(serverId).then(detail => {
      setServer(detail?.server || null)
      setInstallMetadata(detail?.install || null)
    })
    fetchBuyerHub()
      .then(hub => {
        setHubResource(hub?.hub?.hubUrl || '')
      })
      .catch(() => {
        setHubResource('')
      })
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
    const metadataBaseUrl = installMetadata?.discoveryBaseUrl
    if (!metadataBaseUrl) {
      setMetadataState({ status: 'error', error: 'Marketplace install metadata is unavailable.' })
      return
    }
    if (!hubResource) {
      setMetadataState({ status: 'error', error: 'Buyer hub metadata is unavailable.' })
      return
    }
    try {
      setMetadataState({ status: 'loading' })
      const params = new URLSearchParams({ metadataBaseUrl })
      params.set('resource', hubResource)
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
  }, [hubResource, installMetadata, server])

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

  useEffect(() => {
    if (currentStep !== 'complete' || !selectedAction || autoLaunchAttempted) {
      return
    }

    setAutoLaunchAttempted(true)
    runInstallAction(selectedAction)
  }, [autoLaunchAttempted, currentStep, selectedAction])

  if (!server) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen p-4">
        <Text variant="h4" className="mb-4">Server not found</Text>
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
      setAutoLaunchAttempted(false)
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

  const parseExternalPaymentResponse = () => {
    const raw = externalPaymentInput.trim()
    if (!raw) return null
    try {
      return JSON.parse(raw)
    } catch {
      return {
        paymentIdentifier: raw,
        method: 'x402_wallet',
      }
    }
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
    toast.error('No one-click install action is available for this client.')
  }

  return (
    <div className="flex flex-col min-h-screen bg-background">
      <div className="border-b border-border bg-background/50 backdrop-blur sticky top-0 z-40">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-6 flex items-center justify-between">
          <Link href={`/marketplace/${server.slug}`} className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors"><ArrowLeft className="w-4 h-4" />Back</Link>
          <Text variant="h6">{server.name}</Text><div />
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
                    <div><Text variant="caption">{step.label}</Text></div>
                  </button>
                )
              })}
            </div>
          </div>

          <div className="lg:col-span-3">
            <Card className="p-8">
              <Text variant="h4" className="mb-6">{currentStepData.title}</Text>

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
                        <Text variant="small">CIMD Metadata</Text>
                        <Text variant="caption" className="break-all text-muted-foreground">
                          {metadataState.status === 'ready'
                            ? metadataState.data.links.cimdUrl
                            : installMetadata?.cimdUrl || 'unknown resource'}
                        </Text>
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
                      <Text variant="caption" className="text-muted-foreground">
                        Scopes exposed: {metadataState.data.cimd.json.scopes_supported.join(', ')}
                      </Text>
                    )}
                    {server.canonicalResourceUri && (
                      <Text variant="caption" className="text-muted-foreground">
                        Upstream runtime target: {server.canonicalResourceUri}
                      </Text>
                    )}
                    {metadataState.status === 'error' && (
                      <Text variant="small" className="text-destructive">
                        {metadataState.error}
                      </Text>
                    )}
                    <div className="grid gap-3 md:grid-cols-2">
                      <div className="p-3 rounded-md bg-muted/60">
                        <Text variant="caption" className="uppercase text-muted-foreground">OAuth Metadata</Text>
                        <Text variant="small" className="break-all">
                          {metadataState.status === 'ready' ? metadataState.data.links.oauthMetadataUrl : 'Pending'}
                        </Text>
                        <Badge className="mt-2" variant={metadataState.status === 'ready' && metadataState.data.oauth.status === 'ok' ? 'default' : metadataState.status === 'loading' ? 'outline' : 'destructive'}>
                          {metadataState.status === 'ready' && metadataState.data.oauth.status === 'ok'
                            ? 'Reachable'
                            : metadataState.status === 'loading'
                              ? 'Checking'
                              : 'Error'}
                        </Badge>
                      </div>
                      <div className="p-3 rounded-md bg-muted/60">
                        <Text variant="caption" className="uppercase text-muted-foreground">JWKS</Text>
                        <Text variant="small" className="break-all">
                          {metadataState.status === 'ready' && metadataState.data.links.jwksUrl ? metadataState.data.links.jwksUrl : 'Awaiting metadata'}
                        </Text>
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
                        <Text variant="caption" className="text-muted-foreground">
                          Last checked {new Date(metadataState.data.timestamp).toLocaleTimeString()}
                        </Text>
                      )}
                    </div>
                  </Card>

                  <div className="flex items-center space-x-3 p-4 border border-border rounded-lg">
                    <Checkbox id="auth-ready" checked={authConfirmed} onCheckedChange={state => setAuthConfirmed(state === true)} />
                    <Label htmlFor="auth-ready" className="flex-1 cursor-pointer">
                      <Text variant="small">I can complete OAuth authorization using the metadata above.</Text>
                    </Label>
                  </div>

                  {metadataState.status === 'error' && (
                    <div className="flex items-center space-x-3 p-4 border border-border rounded-lg bg-amber-500/10">
                      <Checkbox id="auth-override" checked={manualAuthOverride} onCheckedChange={state => setManualAuthOverride(state === true)} />
                      <Label htmlFor="auth-override" className="flex-1 cursor-pointer">
                        <Text variant="small">Allow manual override if CIMD metadata blocks automated checks.</Text>
                        <Text variant="caption" className="text-muted-foreground">Only enable this if you verified the metadata manually.</Text>
                      </Label>
                    </div>
                  )}
                </div>
              )}

              {currentStep === 'scopes' && (
                <div className="space-y-6">
                  <div className="space-y-3 bg-muted/50 p-4 rounded-lg">
                    {server.requiredScopes.map(scope => (
                      <div key={scope} className="flex items-center gap-3"><CheckIcon className="w-5 h-5 text-green-500 flex-shrink-0" /><Text variant="small">{scope}</Text></div>
                    ))}
                  </div>
                  <div className="flex items-center space-x-3 p-4 border border-border rounded-lg">
                    <Checkbox id="scopes-accepted" checked={acceptedScopes} onCheckedChange={state => setAcceptedScopes(state === true)} />
                    <Label htmlFor="scopes-accepted" className="flex-1 cursor-pointer"><Text variant="small">I accept these permissions</Text></Label>
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
                      <Text variant="small" className="text-muted-foreground">
                        Entitlement status: <Text as="span" variant="small" className="text-foreground">{scopeCheckState.result.entitlementStatus}</Text>
                      </Text>
                      {scopeCheckState.result.autoGrantAvailable && (
                        <Text variant="small" className="text-muted-foreground">
                          This free server will auto-grant an entitlement during install.
                        </Text>
                      )}
                      {scopeCheckState.result.paymentRequired && (
                        <Text variant="small" className="text-amber-600 dark:text-amber-400">
                          You will be prompted to settle payment before install completes.
                        </Text>
                      )}
                      <Text variant="caption" className="text-muted-foreground">Validated scopes: {scopeCheckState.result.grantedScopes.join(', ')}</Text>
                    </Card>
                  )}
                  {scopeCheckState.status === 'error' && (
                    <Text variant="small" className="text-destructive">{scopeCheckState.error}</Text>
                  )}
                </div>
              )}

              {currentStep === 'connect' && (
                <div className="space-y-6">
                  <Card className="bg-muted p-6 border-border space-y-4">
                    <Text variant="h6">One-Click Install Ready</Text>
                    <Text variant="small" className="text-muted-foreground">
                      Next will create your server connection and generate a client-specific install action.
                    </Text>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                      <div>
                        <Text variant="small" className="text-muted-foreground">Client</Text>
                        <Text variant="small" className="uppercase">{selectedClient}</Text>
                      </div>
                      <div>
                        <Text variant="small" className="text-muted-foreground">Permissions</Text>
                        <Text variant="small">{server.requiredScopes.length} scopes</Text>
                      </div>
                    </div>
                    <div className="space-y-2 pt-2 border-t border-border">
                      <Text variant="caption" className="text-muted-foreground">Install payment methods from buyer payment controls</Text>
                      {paymentControlsLoading ? (
                        <Text variant="small" className="text-muted-foreground">Loading payment readiness...</Text>
                      ) : paymentControlsError ? (
                        <Text variant="small" className="text-destructive">{paymentControlsError}</Text>
                      ) : installPaymentMethods.length === 0 ? (
                        <Text variant="small" className="text-muted-foreground">No install-capable payment methods were returned by the API.</Text>
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
                                    <Text variant="small">{getPaymentMethodTitle(method)}</Text>
                                    <Badge variant={readiness.tone}>{readiness.label}</Badge>
                                  </div>
                                  <Text variant="small" className="text-muted-foreground">{getPaymentMethodDescription(method)}</Text>
                                  <Text variant="caption" className="text-muted-foreground">{readiness.description}</Text>
                                </div>
                              </label>
                            )
                          })}
                        </RadioGroup>
                      )}
                      {paymentRequiredAheadOfInstall && readyInstallMethods.length === 0 && !paymentControlsLoading && !paymentControlsError && (
                        <Text variant="caption" className="text-amber-600 dark:text-amber-400">
                          Scope validation says payment will be required, but no allowed method is ready. Open billing to enable or fund one first.
                        </Text>
                      )}
                    </div>
                  </Card>
                  {installing && (
                    <Text variant="small" className="text-muted-foreground">Preparing install session...</Text>
                  )}
                  {paymentRequired && (
                    <Card className="p-6 space-y-4 border-2 border-amber-500/50">
                      <Text variant="h6">Payment Required</Text>
                      <Text variant="small" className="text-muted-foreground">{paymentRequired.error || 'Settle payment to continue install.'}</Text>
                      <Text variant="caption" className="text-muted-foreground">
                        Intent: {paymentRequired.intent?.id} | Amount: {Number(paymentRequired.intent?.amountUsdc || 0).toFixed(2)} USDC
                      </Text>
                      {selectedPaymentMethod === 'x402_wallet' && (
                          <div className="space-y-2">
                            <Label htmlFor="external-payment-id">External Payment Payload</Label>
                            <textarea
                              id="external-payment-id"
                              value={externalPaymentInput}
                              onChange={e => setExternalPaymentInput(e.target.value)}
                              className="w-full min-h-28 px-3 py-2 rounded-md border border-input bg-background text-sm font-mono"
                              placeholder='Paste the x402 payment JSON payload, or a payment identifier / tx hash'
                            />
                            <Text variant="caption" className="text-muted-foreground">
                              x402 clients normally retry the request with a `PAYMENT-SIGNATURE` payload. Paste that JSON here for facilitator-backed settlement testing.
                            </Text>
                          </div>
                        )}
                      {paymentRequired.wwwAuthenticate && (
                        <div className="bg-background rounded-md p-3 font-mono text-xs overflow-x-auto">
                          <Text variant="caption" className="mb-1 uppercase text-muted-foreground">WWW-Authenticate</Text>
                          <pre className="whitespace-pre-wrap break-all">{paymentRequired.wwwAuthenticate}</pre>
                        </div>
                      )}
                      {parsedPaymentChallenge && (
                        <div className="bg-background rounded-md p-3 font-mono text-xs overflow-x-auto space-y-1">
                          <Text variant="caption" className="uppercase text-muted-foreground">PAYMENT-REQUIRED</Text>
                          <pre className="whitespace-pre-wrap break-all">
                            {typeof parsedPaymentChallenge === 'string'
                              ? parsedPaymentChallenge
                              : JSON.stringify(parsedPaymentChallenge, null, 2)}
                          </pre>
                        </div>
                      )}
                      <div className="flex flex-wrap gap-2">
                          <Button
                            disabled={installing || (selectedPaymentMethod === 'x402_wallet' && !externalPaymentInput.trim())}
                            onClick={async () => {
                              if (!paymentRequired?.intent?.id) return
                              try {
                                setInstalling(true)
                                if (selectedPaymentMethod === 'wallet_balance') {
                                  await settleX402Intent(paymentRequired.intent.id)
                                } else {
                                  const parsed = parseExternalPaymentResponse()
                                  if (!parsed) {
                                    throw new Error('Provide an x402 payment payload or payment identifier before settling.')
                                  }
                                  await settleX402Intent(paymentRequired.intent.id, {
                                    paymentResponse: parsed,
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
                  <Text variant="h4">Installation Session Created</Text>
                  <div className="flex flex-wrap gap-2 justify-center">{server.requiredScopes.map(scope => <Badge key={scope} variant="outline">{scope}</Badge>)}</div>

                  {selectedAction && (
                    <Card className="text-left p-6 space-y-4">
                      <Text variant="small">{selectedAction.label}</Text>
                      {selectedAction.description && (
                        <Text variant="small" className="text-muted-foreground">{selectedAction.description}</Text>
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
                      </div>
                    </Card>
                  )}

                  {showBridgeHelp && selectedAction?.requiresLocalExec && selectedAction.launchUrl?.startsWith('mcp-marketplace://') && (
                    <Card className="text-left p-6 space-y-4 border-dashed">
                      <Text variant="small">Install MCP Local Bridge (one-time)</Text>
                      <Text variant="small" className="text-muted-foreground">
                        One-click local command execution needs the local bridge registered once on this machine.
                      </Text>
                      <div className="bg-background rounded p-4 font-mono text-xs overflow-x-auto">
                        <pre className="text-foreground/80">{bridgeInstallCommand}</pre>
                      </div>
                      <div className="flex flex-wrap gap-2">
                        <Button onClick={() => runInstallAction(selectedAction)}>
                          Retry One-Click Install
                        </Button>
                      </div>
                    </Card>
                  )}

                  {installSession?.connection?.id && (
                    <Text variant="small" className="text-muted-foreground">
                      Connection ID: {installSession.connection.id}
                    </Text>
                  )}
                  <Button asChild>
                    <Link href="/buyer/connections">Go to Connections</Link>
                  </Button>
                </div>
              )}

              <div className="flex items-center justify-between gap-4 mt-8 pt-6 border-t border-border">
                <Button variant="outline" onClick={handleBack} disabled={currentStepIndex === 0}>Back</Button>
                <Text variant="small" className="text-muted-foreground">Step {currentStepIndex + 1} of {steps.length}</Text>
                <Button onClick={handleNext} disabled={!canProceed() || installing}>{currentStep === 'complete' ? 'Done' : installing ? 'Preparing...' : 'Next'}{currentStep !== 'complete' && !installing && <ChevronRight className="ml-2 w-4 h-4" />}</Button>
              </div>
            </Card>
          </div>
        </div>
      </main>
    </div>
  )
}
