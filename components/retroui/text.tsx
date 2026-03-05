'use client'

import * as React from 'react'
import { cn } from '@/lib/utils'

interface TextProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6' | 'body' | 'small' | 'caption'
  bold?: boolean
}

const variantStyles = {
  h1: 'text-5xl font-black leading-tight',
  h2: 'text-4xl font-bold leading-tight',
  h3: 'text-3xl font-bold leading-tight',
  h4: 'text-2xl font-bold leading-tight',
  h5: 'text-xl font-bold',
  h6: 'text-lg font-bold',
  body: 'text-base font-semibold',
  small: 'text-sm font-semibold',
  caption: 'text-xs font-semibold'
}

const Text = React.forwardRef<
  HTMLDivElement,
  TextProps
>(({ className, variant = 'body', bold = false, ...props }, ref) => (
  <div
    ref={ref}
    className={cn(
      'text-black dark:text-white',
      variantStyles[variant],
      bold && 'font-black',
      className
    )}
    {...props}
  />
))
Text.displayName = 'Text'

export { Text }
