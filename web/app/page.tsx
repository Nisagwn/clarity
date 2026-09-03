// app/page.tsx — açılış sayfası
import Link from 'next/link'
import Image from 'next/image'
import { listProducts, formatPrice, ApiError } from '@/lib/api'
import type { Product } from '@/lib/api'

export default async function HomePage() {
  let products: Product[] = []
  let apiDown = false

  try {
    const data = await listProducts({ limit: 6 })
    products = data.products
  } catch (err) {
    apiDown = err instanceof ApiError
  }

  return (
    <div className="space-y-16">
      {/* Hero */}
      <section className="grid gap-8 md:grid-cols-2 md:items-center py-6">
        <div>
          <h1 className="text-4xl sm:text-5xl font-bold leading-tight">
            Makyajının <span className="gradient-text">içinde ne var</span>, bil
          </h1>
          <p className="text-lg text-espresso/70 mt-5 max-w-prose">
            Herhangi bir ürünü içeriklerine ayır, cildinin tepki verdiklerini
            işaretle ve daha uygun fiyatlı, daha nazik muadillerini bul.
          </p>
          <div className="flex flex-wrap gap-3 mt-7">
            <Link
              href="/upload"
              className="px-6 py-3 bg-brand-500 text-white font-semibold rounded-lg hover:bg-brand-600 transition"
            >
              Ürün analiz et
            </Link>
            <Link
              href="/ingredients"
              className="px-6 py-3 border border-brand-300 text-brand-500 font-semibold rounded-lg hover:bg-brand-100 transition"
            >
              İçerikleri keşfet
            </Link>
          </div>
        </div>

        <div className="relative aspect-[4/3] rounded-2xl overflow-hidden shadow-sm border border-brand-100">
          <Image
            src="/hero-lilies.jpg"
            alt="Pembe zambaklar"
            fill
            priority
            sizes="(max-width: 768px) 100vw, 50vw"
            className="object-cover"
          />
        </div>
      </section>

      {/* Nasıl çalışır */}
      <section>
        <h2 className="text-2xl font-bold mb-6 text-center">Nasıl çalışır?</h2>
        <div className="grid gap-6 md:grid-cols-3">
          {[
            {
              step: '1',
              title: 'Alerjenlerini söyle',
              body: 'Tepki verdiğin içeriklerle bir cilt profili kaydet.',
            },
            {
              step: '2',
              title: 'Ürünü seç',
              body: 'Fotoğraf yükle ya da kataloğu ada veya markaya göre ara.',
            },
            {
              step: '3',
              title: 'Dökümü gör',
              body: 'Her içerik puanlanır, alerjenlerin işaretlenir, muadiller önerilir.',
            },
          ].map((item) => (
            <div
              key={item.step}
              className="border border-brand-100 rounded-xl p-6 bg-white"
            >
              <div className="w-8 h-8 rounded-full gradient-primary text-white flex items-center justify-center font-bold mb-3">
                {item.step}
              </div>
              <h3 className="font-semibold mb-2">{item.title}</h3>
              <p className="text-sm text-espresso/70">{item.body}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Katalog önizlemesi */}
      <section>
        <div className="flex items-baseline justify-between mb-6">
          <h2 className="text-2xl font-bold">Katalogdan</h2>
          <Link
            href="/products"
            className="text-brand-400 hover:text-brand-500 hover:underline text-sm"
          >
            Tümünü gör →
          </Link>
        </div>

        {apiDown ? (
          <div className="p-6 bg-clay/10 border border-clay/40 rounded-xl text-cocoa">
            <p className="font-semibold mb-1">API&apos;ye ulaşılamıyor.</p>
            <p className="text-sm">
              PostgreSQL&apos;i <code>docker compose up -d</code>, backend&apos;i{' '}
              <code>make backend</code> ile başlatıp bu sayfayı yenileyin.
            </p>
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {products.map((p) => (
              <Link
                key={p.id}
                href={`/products/${p.id}`}
                className="border border-brand-100 rounded-xl p-5 bg-white hover:border-brand-300 hover:shadow-sm transition"
              >
                <p className="text-xs uppercase tracking-wide text-mauve">
                  {p.brand}
                </p>
                <h3 className="font-semibold mt-1">{p.name}</h3>
                <p className="text-sm text-espresso/70 mt-2">
                  {formatPrice(p.price, p.currency)}
                  {p.category ? ` · ${p.category}` : ''}
                </p>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
