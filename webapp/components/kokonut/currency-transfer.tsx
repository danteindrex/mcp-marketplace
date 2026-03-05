'use client'

import { CheckCircle2, ArrowRight } from 'lucide-react'

interface CurrencyTransferProps {
  fromAmount: number
  fromCurrency: string
  toCurrency: string
  toAmount: number
  conversionRate: number
  isComplete?: boolean
}

export function CurrencyTransfer({
  fromAmount,
  fromCurrency,
  toCurrency,
  toAmount,
  conversionRate,
  isComplete = false
}: CurrencyTransferProps) {
  return (
    <div className="relative w-full max-w-md mx-auto">
      <div className="bg-gradient-to-br from-background to-muted rounded-2xl border border-border p-8">
        {/* From Amount */}
        <div className="mb-8">
          <p className="text-sm text-muted-foreground mb-2">You're sending</p>
          <div className="flex items-baseline gap-2">
            <span className="text-4xl font-bold">{fromAmount.toFixed(2)}</span>
            <span className="text-xl text-muted-foreground">{fromCurrency}</span>
          </div>
        </div>

        {/* Arrow with rate */}
        <div className="relative my-6 flex items-center justify-center">
          <div className="absolute inset-x-0 h-px bg-gradient-to-r from-transparent via-border to-transparent" />
          <div className="relative bg-background px-4 flex items-center gap-2">
            <ArrowRight className="w-5 h-5 text-muted-foreground" />
            <span className="text-xs text-muted-foreground font-medium">
              {conversionRate.toFixed(4)} rate
            </span>
          </div>
        </div>

        {/* To Amount */}
        <div className="mb-8">
          <p className="text-sm text-muted-foreground mb-2">Recipient receives</p>
          <div className="flex items-baseline gap-2">
            <span className="text-4xl font-bold text-green-600 dark:text-green-400">
              {toAmount.toFixed(2)}
            </span>
            <span className="text-xl text-muted-foreground">{toCurrency}</span>
          </div>
        </div>

        {/* Completion Status */}
        {isComplete && (
          <div className="flex items-center justify-center gap-2 pt-6 border-t border-border">
            <CheckCircle2 className="w-5 h-5 text-green-600 dark:text-green-400" />
            <span className="text-sm font-medium text-green-600 dark:text-green-400">
              Payment complete
            </span>
          </div>
        )}
      </div>
    </div>
  )
}
