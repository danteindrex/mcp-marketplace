'use client'

import * as React from 'react'
import { cn } from '@/lib/utils'

type TextElement = React.ElementType

interface TextProps extends React.HTMLAttributes<HTMLElement> {
  as?: TextElement
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
  caption: 'text-xs font-semibold',
}

function inferVariant(as?: TextElement): NonNullable<TextProps['variant']> {
  if (as === 'h1' || as === 'h2' || as === 'h3' || as === 'h4' || as === 'h5' || as === 'h6') {
    return as
  }
  if (as === 'span') {
    return 'caption'
  }
  return 'body'
}

const defaultTagByVariant: Record<NonNullable<TextProps['variant']>, TextElement> = {
  h1: 'h1',
  h2: 'h2',
  h3: 'h3',
  h4: 'h4',
  h5: 'h5',
  h6: 'h6',
  body: 'p',
  small: 'p',
  caption: 'span',
}

const Text = React.forwardRef<HTMLElement, TextProps>(
  ({ as, className, variant, bold = false, ...props }, ref) => {
    const resolvedVariant = variant || inferVariant(as)
    const Component = as || defaultTagByVariant[resolvedVariant]

    return React.createElement(Component as React.ElementType, {
      ...props,
      ref,
      className: cn(
        'font-sans text-black dark:text-white',
        variantStyles[resolvedVariant],
        bold && 'font-black',
        className
      ),
    })
  }
)

Text.displayName = 'Text'

export { Text }
