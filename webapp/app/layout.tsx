import type { Metadata } from 'next'
import { Analytics } from '@vercel/analytics/next'
import { ThemeProvider } from 'next-themes'
import { Toaster } from 'sonner'
import { AnalyticsReferrals } from '@/components/analytics-referrals'
import { RetroUIGlobalProvider } from '@/components/retroui/globals'
import { getSiteUrlObject } from '@/lib/site'
import './globals.css'

export const metadata: Metadata = {
  metadataBase: getSiteUrlObject(),
  title: {
    default: 'MCP Marketplace',
    template: '%s | MCP Marketplace',
  },
  description: 'Discover, install, and manage Model Context Protocol servers across Codex, Claude, Cursor, VS Code, and ChatGPT.',
  generator: 'v0.app',
  applicationName: 'MCP Marketplace',
  keywords: [
    'MCP marketplace',
    'Model Context Protocol',
    'MCP server install',
    'Codex MCP',
    'Claude MCP',
    'Cursor MCP',
    'VS Code MCP',
  ],
  alternates: {
    canonical: '/',
  },
  openGraph: {
    title: 'MCP Marketplace',
    description:
      'Discover, install, and manage Model Context Protocol servers across Codex, Claude, Cursor, VS Code, and ChatGPT.',
    url: '/',
    siteName: 'MCP Marketplace',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'MCP Marketplace',
    description:
      'Discover, install, and manage Model Context Protocol servers across leading AI clients.',
  },
  icons: {
    icon: [
      {
        url: '/icon-light-32x32.png',
        media: '(prefers-color-scheme: light)',
      },
      {
        url: '/icon-dark-32x32.png',
        media: '(prefers-color-scheme: dark)',
      },
      {
        url: '/icon.svg',
        type: 'image/svg+xml',
      },
    ],
    apple: '/apple-icon.png',
  },
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="font-sans">
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <RetroUIGlobalProvider>
            {children}
            <Toaster />
            <AnalyticsReferrals />
          </RetroUIGlobalProvider>
        </ThemeProvider>
        <Analytics />
      </body>
    </html>
  )
}
