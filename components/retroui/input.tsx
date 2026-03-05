'use client'

import * as React from 'react'
import { cn } from '@/lib/utils'

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, ...props }, ref) => (
    <input
      type={type}
      className={cn(
        'flex h-10 w-full rounded-none border-4 border-black bg-white px-3 py-2 text-black',
        'text-base font-bold placeholder:text-gray-600',
        'focus:outline-none focus:ring-0 focus:border-black',
        'transition-all duration-200',
        'hover:shadow-[4px_4px_0px_0px_rgba(0,0,0,0.3)]',
        'focus:shadow-[4px_4px_0px_0px_rgba(0,0,0,0.3)]',
        'disabled:cursor-not-allowed disabled:opacity-50',
        className
      )}
      ref={ref}
      {...props}
    />
  )
)
Input.displayName = 'Input'

export { Input }
