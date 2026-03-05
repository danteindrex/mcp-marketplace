'use client'

import { useState } from 'react'
import Link from 'next/link'
import { use } from 'react'
import { ArrowLeft, Check, ChevronRight, Copy, Check as CheckIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Label } from '@/components/ui/label'
import { toast } from 'sonner'
import { mockServers } from '@/lib/mock-data'

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

  // Find server by slug
  const server = mockServers.find(s => s.slug === serverId)

  if (!server) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen p-4">
        <h1 className="text-2xl font-bold mb-4">Server not found</h1>
        <Button asChild>
          <Link href="/marketplace">Back to Marketplace</Link>
        </Button>
      </div>
    )
  }

  const currentStepIndex = steps.findIndex(s => s.id === currentStep)
  const currentStepData = steps[currentStepIndex]

  const handleNext = () => {
    const nextStep = steps[currentStepIndex + 1]?.id
    if (nextStep) {
      setCurrentStep(nextStep)
    } else {
      // Installation complete, redirect
      window.location.href = '/buyer/connections'
    }
  }

  const handleBack = () => {
    const prevStep = steps[currentStepIndex - 1]?.id
    if (prevStep) {
      setCurrentStep(prevStep)
    }
  }

  const canProceed = () => {
    switch (currentStep) {
      case 'client':
        return selectedClient !== ''
      case 'auth':
        return authReady
      case 'scopes':
        return acceptedScopes
      case 'connect':
        return true
      case 'complete':
        return true
      default:
        return false
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
      {/* Header */}
      <div className="border-b border-border bg-background/50 backdrop-blur sticky top-0 z-40">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-6 flex items-center justify-between">
          <Link
            href={`/marketplace/${server.slug}`}
            className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            Back
          </Link>
          <h1 className="text-xl font-bold">{server.name}</h1>
          <div />
        </div>
      </div>

      {/* Main Content */}
      <main className="flex-1 max-w-3xl mx-auto w-full px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          {/* Sidebar - Step Progress */}
          <div className="lg:col-span-1">
            <div className="space-y-3">
              {steps.map((step, index) => {
                const isCompleted = steps.findIndex(s => s.id === currentStep) > index
                const isCurrent = step.id === currentStep
                return (
                  <button
                    key={step.id}
                    onClick={() => {
                      if (isCompleted) {
                        setCurrentStep(step.id)
                      }
                    }}
                    className={`w-full text-left px-4 py-3 rounded-lg transition-colors flex items-center gap-3 ${
                      isCurrent
                        ? 'bg-primary text-primary-foreground'
                        : isCompleted
                          ? 'bg-green-500/20 text-green-700 dark:text-green-400 hover:bg-green-500/30 cursor-pointer'
                          : 'bg-muted text-muted-foreground'
                    }`}
                  >
                    <div
                      className={`w-6 h-6 rounded-full flex items-center justify-center text-sm font-medium ${
                        isCurrent || isCompleted ? 'bg-current' : 'border border-current'
                      }`}
                    >
                      {isCompleted ? <Check className="w-4 h-4" /> : index + 1}
                    </div>
                    <div>
                      <p className="text-xs font-medium">{step.label}</p>
                      {(isCurrent || isCompleted) && <p className="text-xs opacity-75">{step.title}</p>}
                    </div>
                  </button>
                )
              })}
            </div>
          </div>

          {/* Main Content - Steps */}
          <div className="lg:col-span-3">
            <Card className="p-8">
              <h2 className="text-2xl font-bold mb-6">{currentStepData.title}</h2>

              {/* Step: Client Selection */}
              {currentStep === 'client' && (
                <div className="space-y-4">
                  <p className="text-muted-foreground mb-6">Which client will you use to access this server?</p>
                  <RadioGroup value={selectedClient} onValueChange={setSelectedClient}>
                    {clientOptions.map(client => (
                      <div key={client.value} className="flex items-center space-x-3 p-4 border border-border rounded-lg hover:bg-muted/50 cursor-pointer transition-colors">
                        <RadioGroupItem value={client.value} id={client.value} />
                        <Label htmlFor={client.value} className="flex-1 cursor-pointer">
                          <p className="font-medium">{client.label}</p>
                          <p className="text-sm text-muted-foreground">{client.description}</p>
                        </Label>
                      </div>
                    ))}
                  </RadioGroup>
                </div>
              )}

              {/* Step: Auth Check */}
              {currentStep === 'auth' && (
                <div className="space-y-4">
                  <p className="text-muted-foreground mb-6">
                    This server uses OAuth2 authentication. Do you have the necessary credentials ready?
                  </p>
                  <div className="bg-blue-500/10 border border-blue-200 dark:border-blue-800 rounded-lg p-4 mb-6">
                    <p className="text-sm text-blue-900 dark:text-blue-200">
                      You'll need to authorize {server.name} to access your account. We'll redirect you to the OAuth
                      provider to complete this.
                    </p>
                  </div>

                  <div className="flex items-center space-x-3 p-4 border border-border rounded-lg">
                    <Checkbox
                      id="auth-ready"
                      checked={authReady}
                      onCheckedChange={state => setAuthReady(state === true)}
                    />
                    <Label htmlFor="auth-ready" className="flex-1 cursor-pointer">
                      <p className="font-medium">I have OAuth credentials ready</p>
                      <p className="text-sm text-muted-foreground">Or I can authorize during installation</p>
                    </Label>
                  </div>
                </div>
              )}

              {/* Step: Scopes Review */}
              {currentStep === 'scopes' && (
                <div className="space-y-6">
                  <p className="text-muted-foreground">
                    {server.name} requires the following permissions to function:
                  </p>

                  <div className="space-y-3 bg-muted/50 p-4 rounded-lg">
                    {server.requiredScopes.map(scope => (
                      <div key={scope} className="flex items-center gap-3">
                        <CheckIcon className="w-5 h-5 text-green-500 flex-shrink-0" />
                        <div>
                          <p className="font-medium text-sm">{scope}</p>
                          <p className="text-xs text-muted-foreground">
                            {scope === 'db:read' && 'Read access to database'}
                            {scope === 'db:write' && 'Write access to database'}
                            {scope === 'db:admin' && 'Administrative database access'}
                            {scope === 'github:repos' && 'Repository management'}
                            {scope === 'github:issues' && 'Issue and pull request access'}
                            {scope === 'github:pulls' && 'Pull request operations'}
                            {scope === 'github:admin' && 'GitHub admin access'}
                            {scope === 'documents:read' && 'Document read access'}
                            {scope === 'documents:write' && 'Document write access'}
                            {scope === 'ai:inference' && 'AI model inference'}
                            {scope === 'api:manage' && 'API management'}
                            {scope === 'api:observe' && 'API monitoring and logging'}
                            {scope === 'email:send' && 'Send email messages'}
                            {scope === 'email:manage' && 'Manage email campaigns'}
                            {scope === 'analytics:read' && 'Read analytics data'}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>

                  <div className="bg-amber-500/10 border border-amber-200 dark:border-amber-800 rounded-lg p-4">
                    <p className="text-sm text-amber-900 dark:text-amber-200">
                      You can revoke these permissions at any time from your Connections page.
                    </p>
                  </div>

                  <div className="flex items-center space-x-3 p-4 border border-border rounded-lg">
                    <Checkbox
                      id="scopes-accepted"
                      checked={acceptedScopes}
                      onCheckedChange={state => setAcceptedScopes(state === true)}
                    />
                    <Label htmlFor="scopes-accepted" className="flex-1 cursor-pointer">
                      <p className="font-medium">I accept these permissions</p>
                    </Label>
                  </div>
                </div>
              )}

              {/* Step: Connect */}
              {currentStep === 'connect' && (
                <div className="space-y-6">
                  <div className="bg-green-500/10 border border-green-200 dark:border-green-800 rounded-lg p-4 mb-6">
                    <p className="text-sm text-green-900 dark:text-green-200 font-medium">Almost there!</p>
                    <p className="text-sm text-green-900 dark:text-green-200 mt-1">
                      Next, we'll connect to the OAuth provider and set up your access token.
                    </p>
                  </div>

                  <Card className="bg-muted p-6 border-border">
                    <h3 className="font-semibold mb-4">Configuration Snippet</h3>
                    <div className="bg-background rounded p-4 mb-4 font-mono text-xs overflow-x-auto">
                      <pre className="text-foreground/70">
{`// Add to your client configuration
const mcpConfig = {
  server: "${server.name}",
  version: "${server.version}",
  client: "${selectedClient}",
  scopes: [
${server.requiredScopes.map(s => `    "${s}",`).join('\n')}
  ],
  auth: {
    type: "oauth2",
    provider: "standard"
  }
}`}
                      </pre>
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => copyToClipboard(JSON.stringify({ server: server.name, version: server.version, client: selectedClient }, null, 2))}
                    >
                      <Copy className="w-4 h-4 mr-2" />
                      {copiedCode ? 'Copied!' : 'Copy Config'}
                    </Button>
                  </Card>

                  <div className="space-y-4">
                    <h3 className="font-semibold">Next Steps:</h3>
                    <ol className="space-y-3">
                      {[
                        'Click "Complete Setup" below',
                        'Authorize the OAuth request in your browser',
                        'Configure your client settings',
                        'Start using the server!',
                      ].map((step, i) => (
                        <li key={i} className="flex items-start gap-3">
                          <div className="w-6 h-6 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-xs font-medium flex-shrink-0">
                            {i + 1}
                          </div>
                          <p className="pt-0.5">{step}</p>
                        </li>
                      ))}
                    </ol>
                  </div>
                </div>
              )}

              {/* Step: Complete */}
              {currentStep === 'complete' && (
                <div className="text-center space-y-6">
                  <div className="w-16 h-16 rounded-full bg-green-500/20 flex items-center justify-center mx-auto">
                    <Check className="w-8 h-8 text-green-600 dark:text-green-400" />
                  </div>

                  <div>
                    <h3 className="text-2xl font-bold mb-2">Installation Complete!</h3>
                    <p className="text-muted-foreground">
                      {server.name} is now connected to your account and ready to use.
                    </p>
                  </div>

                  <div className="bg-muted/50 rounded-lg p-6 text-left space-y-3">
                    <div>
                      <p className="text-xs text-muted-foreground uppercase font-medium mb-1">Server</p>
                      <p className="font-medium">{server.name} v{server.version}</p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground uppercase font-medium mb-1">Client</p>
                      <p className="font-medium capitalize">{selectedClient}</p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground uppercase font-medium mb-1">Authorized Scopes</p>
                      <div className="flex flex-wrap gap-2 mt-2">
                        {server.requiredScopes.slice(0, 3).map(scope => (
                          <Badge key={scope} variant="outline" className="text-xs">
                            {scope}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  </div>

                  <div className="flex flex-col sm:flex-row gap-3">
                    <Button variant="outline" asChild className="w-full sm:w-auto">
                      <Link href="/marketplace">Browse More Servers</Link>
                    </Button>
                    <Button asChild className="w-full sm:w-auto">
                      <Link href="/buyer/connections">View All Connections</Link>
                    </Button>
                  </div>
                </div>
              )}

              {/* Navigation Buttons */}
              <div className="flex items-center justify-between gap-4 mt-8 pt-6 border-t border-border">
                <Button
                  variant="outline"
                  onClick={handleBack}
                  disabled={currentStepIndex === 0}
                >
                  Back
                </Button>

                <div className="text-sm text-muted-foreground">
                  Step {currentStepIndex + 1} of {steps.length}
                </div>

                <Button
                  onClick={handleNext}
                  disabled={!canProceed()}
                >
                  {currentStep === 'complete' ? 'Done' : 'Next'}
                  {currentStep !== 'complete' && <ChevronRight className="ml-2 w-4 h-4" />}
                </Button>
              </div>
            </Card>
          </div>
        </div>
      </main>
    </div>
  )
}
