// app/layout.tsx
import type { Metadata } from 'next'
import './globals.css'
import Navigation from '@/components/Navigation'

export const metadata: Metadata = {
  title: 'Clarity · Makyaj İçerik Analizi',
  description:
    'Makyaj ürünlerinin içeriklerini çöz, alerjenlerini işaretle ve cildine uygun daha uygun fiyatlı muadilleri bul.',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="tr">
      <body>
        <Navigation />
        <main className="container mx-auto px-4 py-8">{children}</main>
        <footer className="border-t border-brand-100 mt-12 py-6 text-center text-mauve text-sm">
          <p>Clarity • Hassas ciltler için özenle hazırlandı</p>
        </footer>
      </body>
    </html>
  )
}
