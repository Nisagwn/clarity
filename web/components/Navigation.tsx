// components/Navigation.tsx
'use client'

import Link from 'next/link'
import Image from 'next/image'
import { useState } from 'react'

const links = [
  { href: '/upload', label: 'Fotoğraf Yükle' },
  { href: '/ingredients', label: 'İçerikleri Keşfet' },
  { href: '/products', label: 'Ürünler' },
  { href: '/about', label: 'Hakkında' },
]

export default function Navigation() {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <nav className="border-b border-brand-100 sticky top-0 bg-brand-50/90 backdrop-blur z-50">
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2.5">
          <Image
            src="/logo.svg"
            alt=""
            width={36}
            height={36}
            className="rounded-xl"
          />
          <span className="font-bold text-lg text-brand-500">Clarity</span>
        </Link>

        {/* Masaüstü menü */}
        <div className="hidden md:flex items-center gap-7">
          {links.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className="text-espresso/80 hover:text-brand-500 transition"
            >
              {link.label}
            </Link>
          ))}
          <Link
            href="/profile"
            className="px-4 py-2 bg-brand-500 text-white rounded-lg hover:bg-brand-600 transition"
          >
            Profilim
          </Link>
        </div>

        {/* Mobil menü düğmesi */}
        <button
          onClick={() => setIsOpen(!isOpen)}
          aria-expanded={isOpen}
          aria-label="Menüyü aç/kapat"
          className="md:hidden p-2 hover:bg-brand-100 rounded"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
      </div>

      {/* Mobil menü */}
      {isOpen && (
        <div className="md:hidden border-t border-brand-100">
          <div className="container mx-auto px-4 py-4 space-y-1">
            {links.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className="block text-espresso/80 hover:text-brand-500 py-2"
                onClick={() => setIsOpen(false)}
              >
                {link.label}
              </Link>
            ))}
            <Link
              href="/profile"
              className="block mt-2 px-4 py-2 bg-brand-500 text-white rounded-lg text-center"
              onClick={() => setIsOpen(false)}
            >
              Profilim
            </Link>
          </div>
        </div>
      )}
    </nav>
  )
}
