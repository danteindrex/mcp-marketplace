'use client'

import { useEffect, useState } from 'react'

interface DynamicTextProps {
  texts: string[]
  className?: string
  interval?: number
  transition?: boolean
}

export function DynamicText({ 
  texts, 
  className = '', 
  interval = 3000,
  transition = true
}: DynamicTextProps) {
  const [currentIndex, setCurrentIndex] = useState(0)
  const [isVisible, setIsVisible] = useState(true)

  useEffect(() => {
    const timer = setInterval(() => {
      if (transition) {
        setIsVisible(false)
        setTimeout(() => {
          setCurrentIndex((prev) => (prev + 1) % texts.length)
          setIsVisible(true)
        }, 300)
      } else {
        setCurrentIndex((prev) => (prev + 1) % texts.length)
      }
    }, interval)

    return () => clearInterval(timer)
  }, [texts.length, interval, transition])

  return (
    <span 
      className={`inline-block transition-opacity duration-300 ${
        isVisible ? 'opacity-100' : 'opacity-0'
      } ${className}`}
    >
      {texts[currentIndex]}
    </span>
  )
}
