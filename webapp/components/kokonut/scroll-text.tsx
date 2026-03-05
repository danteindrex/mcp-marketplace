'use client'

import { motion } from 'framer-motion'

interface ScrollTextProps {
  children: string | React.ReactNode
  className?: string
  speed?: 'slow' | 'normal' | 'fast'
}

export function ScrollText({ children, className = '', speed = 'normal' }: ScrollTextProps) {
  const config = {
    slow: { duration: 0.9, y: 32 },
    normal: { duration: 0.65, y: 24 },
    fast: { duration: 0.45, y: 16 },
  }[speed]

  return (
    <motion.div
      className={className}
      initial={{ opacity: 0, y: config.y }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: false, amount: 0.2 }}
      transition={{ duration: config.duration, ease: 'easeOut' }}
    >
      {children}
    </motion.div>
  )
}
