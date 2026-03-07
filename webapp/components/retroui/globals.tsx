'use client'

import React from 'react'
import { useTheme } from 'next-themes'
import { useEffect, useState } from 'react'

/**
 * Global RetroUI wrapper that applies RetroUI styling in light mode
 * Use this at the app level to ensure all pages respect the theme
 */
export function RetroUIGlobalProvider({ children }: { children: React.ReactNode }) {
  const { theme, systemTheme } = useTheme()
  const [mounted, setMounted] = useState(false)
  const currentTheme = theme === 'system' ? systemTheme : theme

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) return <>{children}</>

  // Apply RetroUI styling globally in light mode
  if (currentTheme === 'light') {
    return (
      <div className="retroui-light-theme">
        <style>{`
          .retroui-light-theme {
            --retro-primary: #000;
            --retro-secondary: #f5f5f5;
            --retro-text: #000;
            --retro-border: 2px solid #000;
            --retro-shadow: 4px 4px 0px 0px rgba(0, 0, 0, 0.3);
            --retro-shadow-sm: 2px 2px 0px 0px rgba(0, 0, 0, 0.2);
          }

          /* Keep typographic weight without forcing color overrides */
          .retroui-light-theme h1,
          .retroui-light-theme h2,
          .retroui-light-theme h3,
          .retroui-light-theme h4,
          .retroui-light-theme h5,
          .retroui-light-theme h6 {
            font-family: inherit;
            font-weight: 900;
            letter-spacing: -0.02em;
          }

          /* Button styling */
          .retroui-light-theme button {
            font-weight: bold;
            border-radius: 0;
            box-shadow: var(--retro-shadow);
            transition: all 0.2s;
          }

          .retroui-light-theme button:hover {
            transform: translate(2px, 2px);
            box-shadow: var(--retro-shadow-sm);
          }

          /* Card styling */
          .retroui-light-theme [class*="card"],
          .retroui-light-theme [class*="Card"] {
            border-width: 3px;
            border-style: solid;
            border-color: #000;
            border-radius: 0;
            box-shadow: var(--retro-shadow);
          }

          /* Input styling */
          .retroui-light-theme input,
          .retroui-light-theme textarea,
          .retroui-light-theme select {
            border: 2px solid #000;
            border-radius: 0;
            box-shadow: var(--retro-shadow-sm);
            font-weight: bold;
          }

          .retroui-light-theme input:focus,
          .retroui-light-theme textarea:focus,
          .retroui-light-theme select:focus {
            outline: none;
            box-shadow: var(--retro-shadow);
          }
        `}</style>
        {children}
      </div>
    )
  }

  return <>{children}</>
}
