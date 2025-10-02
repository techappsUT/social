// path: frontend/src/app/layout.tsx

import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import { ThemeProvider } from '@/components/providers/theme-provider';
import { QueryProvider } from '@/components/providers/query-provider';
import { Toaster } from 'sonner';
import './globals.css';

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-sans',
});

export const metadata: Metadata = {
  title: {
    default: 'SocialQueue - Social Media Management',
    template: '%s | SocialQueue',
  },
  description: 'Schedule and manage your social media posts across multiple platforms',
  keywords: ['social media', 'scheduling', 'marketing', 'automation'],
  authors: [{ name: 'SocialQueue Team' }],
  creator: 'SocialQueue',
  metadataBase: new URL(process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000'),
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: process.env.NEXT_PUBLIC_APP_URL,
    title: 'SocialQueue',
    description: 'Schedule and manage your social media posts',
    siteName: 'SocialQueue',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'SocialQueue',
    description: 'Schedule and manage your social media posts',
    creator: '@socialqueue',
  },
  robots: {
    index: true,
    follow: true,
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${inter.variable} font-sans antialiased`}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <QueryProvider>
            {children}
            <Toaster richColors position="top-right" />
          </QueryProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}