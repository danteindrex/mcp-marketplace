'use client'

import React, { useEffect, useState } from 'react'
import { motion } from 'framer-motion'

interface CursorProps {
  color?: string
  size?: number
  followDelay?: number
}

interface CursorFollowProps {
  children?: React.ReactNode
  color?: string
  size?: number
}

export const CursorProvider: React.FC<{ children: React.ReactNode; global?: boolean }> = ({ 
  children, 
  global = false 
}) => {
  const [_position, setPosition] = useState({ x: 0, y: 0 })
  const [_followPosition, setFollowPosition] = useState({ x: 0, y: 0 })
  const [_isVisible, setIsVisible] = useState(false)

  useEffect(() => {
    if (!global) return

    const handleMouseMove = (e: MouseEvent) => {
      setPosition({ x: e.clientX, y: e.clientY })
      setIsVisible(true)

      setTimeout(() => {
        setFollowPosition({ x: e.clientX, y: e.clientY })
      }, 50)
    }

    const handleMouseLeave = () => {
      setIsVisible(false)
    }

    window.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseleave', handleMouseLeave)

    return () => {
      window.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseleave', handleMouseLeave)
    }
  }, [global])

  return (
    <div data-cursor-provider>
      {children}
    </div>
  )
}

export const Cursor: React.FC<CursorProps> = ({
  color = 'rgba(59, 130, 246, 0.8)',
  size = 12,
  followDelay: _followDelay = 0.1
}) => {
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const [isVisible, setIsVisible] = useState(false)

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      setPosition({ x: e.clientX - size / 2, y: e.clientY - size / 2 })
      setIsVisible(true)
    }

    const handleMouseLeave = () => {
      setIsVisible(false)
    }

    window.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseleave', handleMouseLeave)

    return () => {
      window.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseleave', handleMouseLeave)
    }
  }, [size])

  return (
    <>
      <style>{`
        * {
          cursor: none;
        }
      `}</style>
      <motion.div
        className="pointer-events-none fixed z-50 rounded-full"
        style={{
          backgroundColor: color,
          width: size,
          height: size,
          opacity: isVisible ? 1 : 0,
        }}
        animate={{
          x: position.x,
          y: position.y,
        }}
        transition={{
          type: 'spring',
          stiffness: 500,
          damping: 50,
          mass: 1,
        }}
      />
    </>
  )
}

export const CursorFollow: React.FC<CursorFollowProps> = ({
  children,
  color = 'rgba(59, 130, 246, 0.3)',
  size = 40
}) => {
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const [isVisible, setIsVisible] = useState(false)

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      setPosition({ x: e.clientX - size / 2, y: e.clientY - size / 2 })
      setIsVisible(true)
    }

    const handleMouseLeave = () => {
      setIsVisible(false)
    }

    window.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseleave', handleMouseLeave)

    return () => {
      window.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseleave', handleMouseLeave)
    }
  }, [size])

  return (
    <motion.div
      className="pointer-events-none fixed z-40 rounded-full flex items-center justify-center font-medium text-xs text-white"
      style={{
        backgroundColor: color,
        width: size,
        height: size,
        opacity: isVisible ? 1 : 0,
      }}
      animate={{
        x: position.x,
        y: position.y,
      }}
      transition={{
        type: 'spring',
        stiffness: 400,
        damping: 80,
        mass: 2,
      }}
    >
      {children}
    </motion.div>
  )
}
