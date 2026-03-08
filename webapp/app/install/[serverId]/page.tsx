'use client'

import { useState, useEffect, use } from 'react'
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
  fetchServerBySlug,
  installMarketplaceServer,
  settleX402Intent,
  type InstallPaymentRequired,
  type InstallAction,
  type InstallSession,
  type Server,
} from '@/lib/api-client'

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

export default function InstallWizardPage({ params }: PageProps) {
  const { serverId } = use(params)
  const [currentStep, setCurrentStep] = useState<Step>('client')
  const [selectedClient, setSelectedClient] = useState<string>('')
  const [authReady, setAuthReady] = useState(false)
  const [acceptedScopes, setAcceptedScopes] = useState(false)
  const [copiedCode, setCopiedCode] = useState(false)
  const [showBridgeHelp, setShowBridgeHelp] = useState(false)
  const [installing, setInstalling] = useState(false)
  const [installSession, setInstallSession] = useState<InstallSession | null>(null)
  const [paymentRequired, setPaymentRequired] = useState<InstallPaymentRequired | null>(null)
  const [selectedPaymentMethod, setSelectedPaymentMethod] = useState<'wallet_balance' | 'x402_wallet'>('wallet_balance')
  const [externalPaymentID, setExternalPaymentID] = useState('')
  const [server, setServer] = useState<Server | null>(null)

  useEffect(() => {
    fetchServerBySlug(serverId).then(setServer)
  }, [serverId])

  if (!server) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen p-4">
        <h1 className="text-2xl font-bold mb-4">Server not found</h1>
        <Button asChild><Link href="/marketplace">Back to Marketplace</Link></Button>
      </div>
    )
  }

  const currentStepIndex = steps.findIndex(s => s.id === currentStep)
  const currentStepData = steps[currentStepIndex]
  const selectedAction: InstallAction | null = installSession?.install?.selected || null
  const bridgeInstallCommand = 'powershell -ExecutionPolicy Bypass -File backend\\scripts\\install-local-bridge.ps1'
  const installActionLabel = (() => {
    if (selectedAction?.requiresLocalExec && selectedAction?.launchUrl) return 'Run One-Click Install'
    if (selectedAction?.requiresLocalExec) return 'Run Install Command'
    return 'Open Installer'
  })()

  const installToolName = `install_${server.slug.replace(/-/g, '_')}`

  const runInstallAttempt = async (autoSettle: boolean) => {
    try {
      setInstalling(true)
      const out = await installMarketplaceServer(server.slug, {
        client: selectedClient,
        grantedScopes: server.requiredScopes,
        paymentMethod: selectedPaymentMethod,
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
      case 'client': return selectedClient !== ''
      case 'auth': return authReady
      case 'scopes': return acceptedScopes
      default: return true
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
                <div className="space-y-4">
                  <div className="flex items-center space-x-3 p-4 border border-border rounded-lg">
                    <Checkbox id="auth-ready" checked={authReady} onCheckedChange={state => setAuthReady(state === true)} />
                    <Label htmlFor="auth-ready" className="flex-1 cursor-pointer"><p className="font-medium">I can complete OAuth authorization</p></Label>
                  </div>
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
                      <p className="text-xs text-muted-foreground">Payment Method For Paid Servers</p>
                      <RadioGroup
                        value={selectedPaymentMethod}
                        onValueChange={value => setSelectedPaymentMethod(value as 'wallet_balance' | 'x402_wallet')}
                      >
                        <div className="flex items-center gap-2">
                          <RadioGroupItem value="wallet_balance" id="wallet_balance" />
                          <Label htmlFor="wallet_balance">wallet_balance (auto-pay from prepaid wallet)</Label>
                        </div>
                        <div className="flex items-center gap-2">
                          <RadioGroupItem value="x402_wallet" id="x402_wallet" />
                          <Label htmlFor="x402_wallet">x402_wallet (external wallet payment-response)</Label>
                        </div>
                      </RadioGroup>
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
