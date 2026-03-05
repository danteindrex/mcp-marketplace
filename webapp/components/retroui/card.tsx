'use client'

import * as React from 'react'
import { cn } from '@/lib/utils'

interface RetroCardProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'outline' | 'neon'
}

const Card = React.forwardRef<HTMLDivElement, RetroCardProps>(
  ({ className, variant = 'default', ...props }, ref) => {
    const variants = {
      default: 'bg-white border-2 border-black shadow-[6px_6px_0px_0px_rgba(0,0,0,1)]',
      outline: 'bg-transparent border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,0.5)]',
      neon: 'bg-black border-2 border-gray-300 shadow-[0_0_20px_rgba(100,200,255,0.5)]'
    }

    return (
      <div
        ref={ref}
        className={cn(
          'rounded-lg p-6 transition-all duration-300 hover:shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]',
          variants[variant],
          className
        )}
        {...props}
      />
    )
  }
)

Card.displayName = 'RetroCard'

export { Card }
