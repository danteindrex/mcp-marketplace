'use client'

import { useTheme } from 'next-themes'
import { ReactNode } from 'react'

interface ThemeAwareProps {
  children?: ReactNode
  light: ReactNode
  dark: ReactNode
}

export function ThemeAware({ light, dark }: ThemeAwareProps) {
  const { theme, systemTheme } = useTheme()
  const currentTheme = theme === 'system' ? systemTheme : theme

  return currentTheme === 'dark' ? dark : light
}

interface ThemeWrapperProps {
  children: ReactNode
  className?: string
}

export function DarkModeOnly({ children, className }: ThemeWrapperProps) {
  const { theme, systemTheme } = useTheme()
  const currentTheme = theme === 'system' ? systemTheme : theme

  if (currentTheme !== 'dark') return null
  
  return <div className={className}>{children}</div>
}

export function LightModeOnly({ children, className }: ThemeWrapperProps) {
  const { theme, systemTheme } = useTheme()
  const currentTheme = theme === 'system' ? systemTheme : theme

  if (currentTheme === 'dark') return null
  
  return <div className={className}>{children}</div>
}
