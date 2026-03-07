'use client'

import { useEffect, useState } from 'react'

interface TypeWriterProps {
  text: string
  className?: string
  speed?: number
  deleteSpeed?: number
  delay?: number
  cursor?: boolean
  loop?: boolean
  holdDuration?: number
  loopDelay?: number
}

export function TypeWriter({ 
  text, 
  className = '', 
  speed = 50, 
  deleteSpeed = 34,
  delay = 0,
  cursor = true,
  loop = false,
  holdDuration = 2000,
  loopDelay = 2000
}: TypeWriterProps) {
  const [displayedText, setDisplayedText] = useState('')
  const [isComplete, setIsComplete] = useState(false)
  const [phase, setPhase] = useState<'waiting' | 'typing' | 'holding' | 'deleting' | 'stopped'>(
    delay > 0 ? 'waiting' : 'typing',
  )

  useEffect(() => {
    setDisplayedText('')
    setIsComplete(false)
    setPhase(delay > 0 ? 'waiting' : 'typing')
  }, [text, delay])

  useEffect(() => {
    let timer: ReturnType<typeof setTimeout> | undefined

    if (phase === 'waiting') {
      timer = setTimeout(() => setPhase('typing'), delay)
    } else if (phase === 'typing') {
      if (displayedText.length < text.length) {
        timer = setTimeout(() => {
          const nextLength = displayedText.length + 1
          setDisplayedText(text.slice(0, nextLength))
        }, speed)
      } else {
        setIsComplete(true)
        if (loop) {
          setPhase('holding')
        } else {
          setPhase('stopped')
        }
      }
    } else if (phase === 'holding') {
      timer = setTimeout(() => setPhase('deleting'), holdDuration)
    } else if (phase === 'deleting') {
      if (displayedText.length > 0) {
        timer = setTimeout(() => {
          const nextLength = displayedText.length - 1
          setDisplayedText(text.slice(0, nextLength))
        }, deleteSpeed)
      } else {
        setIsComplete(false)
        timer = setTimeout(() => setPhase('typing'), loopDelay)
      }
    }

    return () => {
      if (timer) clearTimeout(timer)
    }
  }, [phase, displayedText, text, speed, deleteSpeed, delay, loop, holdDuration, loopDelay])

  return (
    <span className={className}>
      {displayedText}
      {cursor && (loop || !isComplete) && (
        <span className="inline-block w-1 h-6 ml-1 bg-current animate-pulse" />
      )}
    </span>
  )
}
