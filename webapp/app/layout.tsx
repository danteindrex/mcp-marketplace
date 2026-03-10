import type { Metadata } from 'next'
import { Analytics } from '@vercel/analytics/next'
import { ThemeProvider } from 'next-themes'
import { Toaster } from 'sonner'
import { RetroUIGlobalProvider } from '@/components/retroui/globals'
import './globals.css'

export const metadata: Metadata = {
  title: 'MCP Marketplace',
  description: 'Discover, install, and manage Model Context Protocol servers',
  generator: 'v0.app',
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
      <body className="font-sans antialiased">
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <RetroUIGlobalProvider>
            {children}
            <Toaster />
          </RetroUIGlobalProvider>
        </ThemeProvider>
        <Analytics />
      </body>
    </html>
  )
}
