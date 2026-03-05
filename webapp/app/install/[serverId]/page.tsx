'use client'

import { useState, useEffect, use } from 'react'
import Link from 'next/link'
import { ArrowLeft, Check, ChevronRight, Copy, Check as CheckIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Label } from '@/components/ui/label'
import { toast } from 'sonner'
import { createConnection, fetchServerBySlug, type Server } from '@/lib/api-client'

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
  { value: 'codex', label: 'OpenAI Codex', description: 'Legacy support' },
]

export default function InstallWizardPage({ params }: PageProps) {
  const { serverId } = use(params)
  const [currentStep, setCurrentStep] = useState<Step>('client')
  const [selectedClient, setSelectedClient] = useState<string>('')
  const [authReady, setAuthReady] = useState(false)
  const [acceptedScopes, setAcceptedScopes] = useState(false)
  const [copiedCode, setCopiedCode] = useState(false)
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

  const handleNext = async () => {
    if (currentStep === 'connect') {
      await createConnection({
        client: selectedClient,
        resource: `https://mcp.marketplace.local/hub/tenant_acme/user_buyer`,
        grantedScopes: server.requiredScopes,
      })
      setCurrentStep('complete')
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
    toast.success('Configuration copied to clipboard')
    setTimeout(() => setCopiedCode(false), 2000)
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
                  <Card className="bg-muted p-6 border-border">
                    <h3 className="font-semibold mb-4">Configuration Snippet</h3>
                    <div className="bg-background rounded p-4 mb-4 font-mono text-xs overflow-x-auto">
                      <pre className="text-foreground/70">{`{"server":"${server.name}","client":"${selectedClient}"}`}</pre>
                    </div>
                    <Button variant="outline" size="sm" onClick={() => copyToClipboard(JSON.stringify({ server: server.name, client: selectedClient }))}><Copy className="w-4 h-4 mr-2" />{copiedCode ? 'Copied!' : 'Copy Config'}</Button>
                  </Card>
                </div>
              )}

              {currentStep === 'complete' && (
                <div className="text-center space-y-6">
                  <div className="w-16 h-16 rounded-full bg-green-500/20 flex items-center justify-center mx-auto"><Check className="w-8 h-8 text-green-600 dark:text-green-400" /></div>
                  <h3 className="text-2xl font-bold">Installation Complete!</h3>
                  <div className="flex flex-wrap gap-2 justify-center">{server.requiredScopes.map(scope => <Badge key={scope} variant="outline">{scope}</Badge>)}</div>
                </div>
              )}

              <div className="flex items-center justify-between gap-4 mt-8 pt-6 border-t border-border">
                <Button variant="outline" onClick={handleBack} disabled={currentStepIndex === 0}>Back</Button>
                <div className="text-sm text-muted-foreground">Step {currentStepIndex + 1} of {steps.length}</div>
                <Button onClick={handleNext} disabled={!canProceed()}>{currentStep === 'complete' ? 'Done' : 'Next'}{currentStep !== 'complete' && <ChevronRight className="ml-2 w-4 h-4" />}</Button>
              </div>
            </Card>
          </div>
        </div>
      </main>
    </div>
  )
}