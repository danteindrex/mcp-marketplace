'use client'

import { useEffect, useState } from 'react'

interface TypeWriterProps {
  text: string
  className?: string
  speed?: number
  delay?: number
  cursor?: boolean
  loop?: boolean
  loopDelay?: number
}

export function TypeWriter({ 
  text, 
  className = '', 
  speed = 50, 
  delay = 0,
  cursor = true,
  loop = false,
  loopDelay = 2000
}: TypeWriterProps) {
  const [displayedText, setDisplayedText] = useState('')
  const [isComplete, setIsComplete] = useState(false)

  useEffect(() => {
    const startTime = setTimeout(() => {
      let currentIndex = 0
      
      const interval = setInterval(() => {
        if (currentIndex <= text.length) {
          setDisplayedText(text.slice(0, currentIndex))
          currentIndex++
        } else {
          setIsComplete(true)
          
          if (loop) {
            // Wait before restarting
            const resetTimeout = setTimeout(() => {
              currentIndex = 0
              setDisplayedText('')
              setIsComplete(false)
            }, loopDelay)
            
            return () => clearTimeout(resetTimeout)
          }
          
          clearInterval(interval)
        }
      }, speed)

      return () => clearInterval(interval)
    }, delay)

    return () => clearTimeout(startTime)
  }, [text, speed, delay, loop, loopDelay])

  return (
    <span className={className}>
      {displayedText}
      {cursor && !isComplete && (
        <span className="inline-block w-1 h-6 ml-1 bg-current animate-pulse" />
      )}
    </span>
  )
}
