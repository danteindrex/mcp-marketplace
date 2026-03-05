'use client'

import { useRef, useEffect, useState } from 'react'

interface MouseEffectCardProps {
  children: React.ReactNode
  className?: string
  intensity?: number
}

export function MouseEffectCard({ 
  children, 
  className = '',
  intensity = 1
}: MouseEffectCardProps) {
  const ref = useRef<HTMLDivElement>(null)
  const [isHovering, setIsHovering] = useState(false)
  const [mousePosition, setMousePosition] = useState({ x: 0, y: 0 })

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
      className={`relative overflow-hidden transition-all duration-300 ${
        isHovering ? 'shadow-xl' : 'shadow-lg'
      } ${className}`}
      style={{
        perspective: '1000px',
        transform: isHovering 
          ? `rotateX(${(mousePosition.y - 100) / 50 * intensity}deg) rotateY(${(mousePosition.x - 100) / 50 * intensity}deg)`
          : 'rotateX(0deg) rotateY(0deg)',
      }}
    >
      {/* Gradient light effect following mouse */}
      {isHovering && (
        <div
          className="absolute inset-0 opacity-30 pointer-events-none"
          style={{
            background: `radial-gradient(circle at ${mousePosition.x}px ${mousePosition.y}px, rgba(255,255,255,0.4) 0%, transparent 50%)`,
          }}
        />
      )}
      {children}
    </div>
  )
}
