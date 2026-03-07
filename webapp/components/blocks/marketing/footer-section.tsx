import Link from 'next/link'
import { type ReactNode } from 'react'
import { Facebook, Github, Globe, Instagram, Linkedin, Twitter, Youtube } from 'lucide-react'

type FooterLink = {
  label: string
  href: string
}

type FooterColumn = {
  title: string
  links: FooterLink[]
}

type SocialLink = {
  platform: string
  href: string
}

function socialIcon(platform: string) {
  const key = platform.toLowerCase()
  if (key === 'twitter' || key === 'x') return <Twitter className="h-4 w-4" />
  if (key === 'github') return <Github className="h-4 w-4" />
  if (key === 'linkedin') return <Linkedin className="h-4 w-4" />
  if (key === 'youtube') return <Youtube className="h-4 w-4" />
  if (key === 'instagram') return <Instagram className="h-4 w-4" />
  if (key === 'facebook') return <Facebook className="h-4 w-4" />
  return <Globe className="h-4 w-4" />
}

function MultiColumn({
  logo,
  description,
  columns,
  socialLinks = [],
  copyright,
}: {
  logo: ReactNode
  description?: string
  columns: FooterColumn[]
  socialLinks?: SocialLink[]
  copyright?: string
}) {
  return (
    <footer className="border-t border-border mt-auto bg-foreground text-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="grid gap-8 md:grid-cols-4">
          <div>
            <div className="mb-4">{logo}</div>
            {description ? <p className="text-sm text-background/75">{description}</p> : null}
            {socialLinks.length > 0 ? (
              <div className="mt-4 flex flex-wrap gap-2">
                {socialLinks.map((item) => (
                  <a
                    key={`${item.platform}-${item.href}`}
                    href={item.href}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center justify-center h-8 w-8 border border-background/30 hover:bg-background/10 transition-colors"
                    aria-label={item.platform}
                  >
                    {socialIcon(item.platform)}
                  </a>
                ))}
              </div>
            ) : null}
          </div>
          {columns.map((column) => (
            <div key={column.title}>
              <h4 className="mb-3 text-sm font-black uppercase">{column.title}</h4>
              <ul className="space-y-2 text-sm text-background/80">
                {column.links.map((link) => (
                  <li key={`${column.title}-${link.href}`}>
                    {link.href.startsWith('http') ? (
                      <a href={link.href} target="_blank" rel="noreferrer" className="hover:text-background">
                        {link.label}
                      </a>
                    ) : (
                      <Link href={link.href} className="hover:text-background">
                        {link.label}
                      </Link>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
        {copyright ? (
          <div className="mt-10 border-t border-background/25 pt-6 text-center text-sm text-background/70">
            {copyright}
          </div>
        ) : null}
      </div>
    </footer>
  )
}

export const FooterSection = {
  MultiColumn,
}
