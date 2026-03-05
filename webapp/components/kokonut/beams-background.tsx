'use client'

import React from 'react'

interface BeamsBackgroundProps {
  children?: React.ReactNode
  className?: string
  colorFrom?: string
  colorTo?: string
  intensity?: number
}

export const BeamsBackground: React.FC<BeamsBackgroundProps> = ({
  children,
  className = '',
  colorFrom = '#3b82f6',
  colorTo = '#8b5cf6',
  intensity = 0.5,
}) => {
  return (
    <div className={`relative w-full overflow-hidden ${className}`}>
      {/* Animated beams background */}
      <div className="absolute inset-0 overflow-hidden">
        <svg
          className="w-full h-full"
          viewBox="0 0 1200 800"
          preserveAspectRatio="xMidYMid slice"
          xmlns="http://www.w3.org/2000/svg"
        >
          <defs>
            <linearGradient id="beam-gradient-1" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stopColor={colorFrom} stopOpacity={intensity * 0.6} />
              <stop offset="100%" stopColor={colorTo} stopOpacity={intensity * 0.2} />
            </linearGradient>
            <linearGradient id="beam-gradient-2" x1="100%" y1="0%" x2="0%" y2="100%">
              <stop offset="0%" stopColor={colorTo} stopOpacity={intensity * 0.5} />
              <stop offset="100%" stopColor={colorFrom} stopOpacity={intensity * 0.1} />
            </linearGradient>
            <filter id="beam-blur">
              <feGaussianBlur in="SourceGraphic" stdDeviation="2" />
            </filter>
          </defs>

          {/* Animated beam lines */}
          <g opacity={intensity}>
            <line
              x1="0"
              y1="0"
              x2="600"
              y2="800"
              stroke="url(#beam-gradient-1)"
              strokeWidth="2"
              filter="url(#beam-blur)"
              className="animate-pulse"
            />
            <line
              x1="200"
              y1="0"
              x2="800"
              y2="800"
              stroke="url(#beam-gradient-1)"
              strokeWidth="2"
              filter="url(#beam-blur)"
              className="animate-pulse"
              style={{ animationDelay: '0.5s' }}
            />
            <line
              x1="400"
              y1="0"
              x2="1000"
              y2="800"
              stroke="url(#beam-gradient-2)"
              strokeWidth="2"
              filter="url(#beam-blur)"
              className="animate-pulse"
              style={{ animationDelay: '1s' }}
            />
            <line
              x1="600"
              y1="0"
              x2="1200"
              y2="800"
              stroke="url(#beam-gradient-2)"
              strokeWidth="2"
              filter="url(#beam-blur)"
              className="animate-pulse"
              style={{ animationDelay: '1.5s' }}
            />
            <line
              x1="800"
              y1="0"
              x2="200"
              y2="800"
              stroke="url(#beam-gradient-1)"
              strokeWidth="2"
              filter="url(#beam-blur)"
              className="animate-pulse"
              style={{ animationDelay: '0.7s' }}
            />
          </g>
        </svg>

        {/* Gradient overlay */}
        <div className="absolute inset-0 bg-gradient-to-b from-transparent via-transparent to-background" />
      </div>

      {/* Content */}
      <div className="relative z-10">{children}</div>
    </div>
  )
}
