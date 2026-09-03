// app/products/page.tsx — ürün kataloğu
import Link from 'next/link'
import { listProducts, formatPrice, ApiError } from '@/lib/api'
import type { Product } from '@/lib/api'

export const metadata = {
  title: 'Ürünler · Clarity',
}

export default async function ProductsPage({
  searchParams,
}: {
  searchParams: { q?: string }
}) {
  const q = searchParams.q ?? ''

  let products: Product[] = []
  let total = 0
  let error = ''

  try {
    const data = await listProducts({ q, limit: 100 })
    products = data.products
    total = data.total
  } catch (err) {
    error = err instanceof ApiError ? err.message : 'Katalog yüklenemedi.'
  }

  return (
    <div className="max-w-4xl mx-auto py-8">
      <h1 className="text-4xl font-bold mb-2">Ürünler</h1>
      <p className="text-espresso/70 mb-8">
        Katalogdaki tüm ürünler ve tam içerik dökümleri.
      </p>

      <form method="GET" className="mb-8 flex gap-3">
        <input
          name="q"
          defaultValue={q}
          placeholder="Ürün veya marka ara"
          className="flex-1 px-4 py-2 border border-brand-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-400"
        />
        <button
          type="submit"
          className="px-5 py-2 bg-brand-500 text-white font-semibold rounded-lg hover:bg-brand-600 transition"
        >
          Ara
        </button>
      </form>

      {error ? (
        <div className="p-6 bg-brand-100 border border-brand-300 text-brand-500 rounded-xl">
          {error}
        </div>
      ) : products.length === 0 ? (
        <div className="p-8 text-center border border-dashed border-brand-200 rounded-xl text-mauve">
          “{q}” için ürün bulunamadı.
        </div>
      ) : (
        <>
          <p className="text-sm text-ink-muted mb-4">{total} ürün</p>
          <div className="grid gap-4 sm:grid-cols-2">
            {products.map((p) => (
              <Link
                key={p.id}
                href={`/products/${p.id}`}
                className="border border-brand-100 rounded-xl p-5 bg-white hover:border-brand-300 hover:shadow-sm transition"
              >
                <p className="text-xs uppercase tracking-wide text-mauve">
                  {p.brand}
                </p>
                <h2 className="font-semibold mt-1">{p.name}</h2>
                {p.description && (
                  <p className="text-sm text-espresso/70 mt-2">{p.description}</p>
                )}
                <p className="text-sm font-semibold mt-3">
                  {formatPrice(p.price, p.currency)}
                  {p.category && (
                    <span className="font-normal text-mist"> · {p.category}</span>
                  )}
                </p>
              </Link>
            ))}
          </div>
        </>
      )}
    </div>
  )
}
