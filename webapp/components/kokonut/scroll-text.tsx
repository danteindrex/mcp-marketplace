'use client'

import { useEffect, useRef } from 'react'

interface ScrollTextProps {
  children: string | React.ReactNode
  className?: string
  speed?: 'slow' | 'normal' | 'fast'
}

export function ScrollText({ children, className = '', speed = 'normal' }: ScrollTextProps) {
  const ref = useRef<HTMLDivElement>(null)

  const speedClass = {
    slow: 'animate-scroll-slow',
    normal: 'animate-scroll',
    fast: 'animate-scroll-fast'
  }[speed]

  return (
    <div className={`overflow-hidden ${className}`}>
      <style>{`
        @keyframes scroll {
          0% { transform: translateY(100%); opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { transform: translateY(-100%); opacity: 0; }
        }
        @keyframes scroll-slow {
          0% { transform: translateY(100%); opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { transform: translateY(-100%); opacity: 0; }
        }
        @keyframes scroll-fast {
          0% { transform: translateY(100%); opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { transform: translateY(-100%); opacity: 0; }
        }
        .animate-scroll {
          animation: scroll 4s ease-in-out infinite;
        }
        .animate-scroll-slow {
          animation: scroll-slow 6s ease-in-out infinite;
        }
        .animate-scroll-fast {
          animation: scroll-fast 2s ease-in-out infinite;
        }
      `}</style>
      <div ref={ref} className={`${speedClass}`}>
        {children}
      </div>
    </div>
  )
}
