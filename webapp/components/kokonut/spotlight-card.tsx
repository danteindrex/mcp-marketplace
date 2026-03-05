'use client'

import { useRef, useEffect, useState } from 'react'

interface SpotlightCardProps {
  children: React.ReactNode
  className?: string
  spotlightColor?: string
}

export function SpotlightCard({ 
  children, 
  className = '',
  spotlightColor = 'rgba(59, 130, 246, 0.5)' // blue by default
}: SpotlightCardProps) {
  const ref = useRef<HTMLDivElement>(null)
  const [mousePosition, setMousePosition] = useState({ x: 0, y: 0 })
  const [isHovering, setIsHovering] = useState(false)

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!ref.current) return

      const rect = ref.current.getBoundingClientRect()
      const x = e.clientX - rect.left
      const y = e.clientY - rect.top

      setMousePosition({ x, y })
    }

    const handleMouseEnter = () => setIsHovering(true)
    const handleMouseLeave = () => setIsHovering(false)

    const element = ref.current
    if (element) {
      element.addEventListener('mousemove', handleMouseMove)
      element.addEventListener('mouseenter', handleMouseEnter)
      element.addEventListener('mouseleave', handleMouseLeave)
    }

    return () => {
      if (element) {
        element.removeEventListener('mousemove', handleMouseMove)
        element.removeEventListener('mouseenter', handleMouseEnter)
        element.removeEventListener('mouseleave', handleMouseLeave)
      }
    }
  }, [])

  return (
    <div
      ref={ref}
      className={`relative group overflow-hidden rounded-xl border border-border/50 transition-all duration-300 ${
        isHovering ? 'border-border shadow-lg' : ''
      } ${className}`}
    >
      {/* Spotlight effect */}
      {isHovering && (
        <div
          className="absolute inset-0 pointer-events-none"
          style={{
            background: `radial-gradient(circle 150px at ${mousePosition.x}px ${mousePosition.y}px, ${spotlightColor} 0%, transparent 80%)`,
          }}
        />
      )}
      {/* Content */}
      <div className="relative z-10">
        {children}
      </div>
    </div>
  )
}
