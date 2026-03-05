'use client'

import React, { useState } from 'react'
import { ChevronDown } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'

interface AccordionItemProps {
  value: string
  trigger: React.ReactNode
  content: React.ReactNode
}

interface AccordionProps {
  items: AccordionItemProps[]
  type?: 'single' | 'multiple'
  collapsible?: boolean
}

export const Accordion: React.FC<AccordionProps> = ({
  items,
  type = 'single',
  collapsible = true
}) => {
  const [openItems, setOpenItems] = useState<string[]>([])

  const toggleItem = (value: string) => {
    if (type === 'single') {
      if (collapsible && openItems.includes(value)) {
        setOpenItems([])
      } else {
        setOpenItems([value])
      }
    } else {
      setOpenItems(prev =>
        prev.includes(value)
          ? prev.filter(item => item !== value)
          : [...prev, value]
      )
    }
  }

  return (
    <div className="space-y-2 w-full">
      {items.map(item => (
        <div key={item.value} className="border border-border rounded-lg overflow-hidden">
          <button
            onClick={() => toggleItem(item.value)}
            className="w-full px-4 py-3 text-left flex items-center justify-between hover:bg-muted/50 transition-colors duration-200"
          >
            <span className="font-medium text-sm md:text-base">{item.trigger}</span>
            <motion.div
              animate={{ rotate: openItems.includes(item.value) ? 180 : 0 }}
              transition={{ duration: 0.3 }}
              className="flex-shrink-0 ml-4"
            >
              <ChevronDown className="w-4 h-4 md:w-5 md:h-5" />
            </motion.div>
          </button>

          <AnimatePresence>
            {openItems.includes(item.value) && (
              <motion.div
                initial={{ height: 0, opacity: 0 }}
                animate={{ height: 'auto', opacity: 1 }}
                exit={{ height: 0, opacity: 0 }}
                transition={{ duration: 0.3 }}
                className="overflow-hidden"
              >
                <div className="px-4 py-3 bg-muted/30 border-t border-border text-sm md:text-base text-muted-foreground">
                  {item.content}
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      ))}
    </div>
  )
}
