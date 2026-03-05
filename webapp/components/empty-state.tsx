'use client'

import { LucideIcon } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useTheme } from 'next-themes'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/retroui'

export interface EmptyStateProps {
  icon?: LucideIcon
  title: string
  description?: string
  action?: {
    label: string
    onClick: () => void
  }
  illustration?: React.ReactNode
}

export function EmptyState({ icon: Icon, title, description, action, illustration }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center min-h-96 p-8 text-center">
      {illustration ? (
        <div className="mb-6 w-32 h-32 flex items-center justify-center">{illustration}</div>
      ) : Icon ? (
        <div className="mb-6 w-24 h-24 rounded-lg bg-muted flex items-center justify-center">
          <Icon className="w-12 h-12 text-muted-foreground" />
        </div>
      ) : null}

      <h3 className="text-xl font-semibold text-foreground mb-2">{title}</h3>

      {description && <p className="text-muted-foreground mb-6 max-w-sm">{description}</p>}

      {action && (
        <Button onClick={action.onClick} size="sm">
          {action.label}
        </Button>
      )}
    </div>
  )
}

// Loading State variant
export interface LoadingStateProps {
  message?: string
}

export function LoadingState({ message = 'Loading...' }: LoadingStateProps) {
  const { resolvedTheme } = useTheme()
  const [progress, setProgress] = useState(13)

  useEffect(() => {
    const timer = setTimeout(() => setProgress(66), 500)
    return () => clearTimeout(timer)
  }, [])

  return (
    <div className="flex flex-col items-center justify-center min-h-96 gap-4">
      {resolvedTheme === 'light' ? (
        <Progress value={progress} className="w-[60%] max-w-xs" />
      ) : (
        <div className="w-8 h-8 border-4 border-muted border-t-primary rounded-full animate-spin" />
      )}
      <p className="text-muted-foreground">{message}</p>
    </div>
  )
}

// Error State variant
export interface ErrorStateProps {
  title?: string
  message: string
  action?: {
    label: string
    onClick: () => void
  }
}

export function ErrorState({
  title = 'Something went wrong',
  message,
  action,
}: ErrorStateProps) {
  return (
    <div className="flex flex-col items-center justify-center min-h-96 p-8 text-center">
      <div className="mb-6 w-24 h-24 rounded-lg bg-destructive/10 flex items-center justify-center">
        <div className="w-12 h-12 rounded-full bg-destructive/20 flex items-center justify-center">
          <span className="text-2xl">⚠️</span>
        </div>
      </div>

      <h3 className="text-xl font-semibold text-foreground mb-2">{title}</h3>
      <p className="text-muted-foreground mb-6 max-w-sm">{message}</p>

      {action && (
        <Button onClick={action.onClick} size="sm" variant="outline">
          {action.label}
        </Button>
      )}
    </div>
  )
}
