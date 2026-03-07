import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'

type FAQItem = {
  question: string
  answer: string
}

function AccordionSection({
  title,
  subtitle,
  items,
}: {
  title: string
  subtitle?: string
  items: FAQItem[]
}) {
  return (
    <section className="px-4 sm:px-6 lg:px-8 py-20 max-w-4xl mx-auto w-full border-t border-border">
      <div className="text-center mb-10">
        {subtitle ? <Badge variant="outline" className="mb-4">{subtitle}</Badge> : null}
        <h2 className="text-4xl font-bold mb-4">{title}</h2>
      </div>
      <Card className="border-[3px] border-foreground shadow-[4px_4px_0px_hsl(var(--shadow-color))]">
        <CardContent className="p-6">
          <Accordion type="single" collapsible className="w-full">
            {items.map((item, index) => (
              <AccordionItem key={item.question} value={`faq-${index}`}>
                <AccordionTrigger className="text-base font-semibold">
                  {item.question}
                </AccordionTrigger>
                <AccordionContent className="text-sm text-muted-foreground">
                  {item.answer}
                </AccordionContent>
              </AccordionItem>
            ))}
          </Accordion>
        </CardContent>
      </Card>
    </section>
  )
}

export const FAQSection = {
  Accordion: AccordionSection,
}
