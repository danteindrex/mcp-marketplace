'use client'

import * as React from 'react'
import { cn } from '@/lib/utils'

interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'secondary' | 'destructive' | 'outline'
}

const Badge = React.forwardRef<HTMLDivElement, BadgeProps>(
  ({ className, variant = 'default', ...props }, ref) => {
    const variants = {
      default: 'bg-black text-white border-2 border-black',
      secondary: 'bg-gray-300 text-black border-2 border-black',
      destructive: 'bg-red-500 text-white border-2 border-black',
      outline: 'bg-white text-black border-2 border-black'
    }

    return (
      <div
        ref={ref}
        className={cn(
          'inline-flex items-center rounded-none px-3 py-1 text-sm font-bold',
          'shadow-[2px_2px_0px_0px_rgba(0,0,0,0.3)]',
          variants[variant],
          className
        )}
        {...props}
      />
    )
  }
)
Badge.displayName = 'Badge'

export { Badge }
