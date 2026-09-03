// app/products/[id]/page.tsx — içerikler ve muadillerle ürün detayı
import Link from 'next/link'
import { notFound } from 'next/navigation'
import IngredientCard from '@/components/IngredientCard'
import { getProduct, getDupes, formatPrice, ApiError } from '@/lib/api'
import type { Product, Recommendation } from '@/lib/api'

export default async function ProductDetailPage({
  params,
}: {
  params: { id: string }
}) {
  let product: Product
  try {
    product = await getProduct(params.id)
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound()
    throw err
  }

  // Muadiller güzel bir ek; buradaki bir hata sayfayı bozmamalı.
  let dupes: Recommendation[] = []
  try {
    dupes = (await getDupes(params.id, 4)).recommendations
  } catch {
    dupes = []
  }

  const ingredients = product.ingredients ?? []
  const riskiest = ingredients.filter((i) => i.concern_level >= 7)
  const withAllergens = ingredients.filter((i) => i.allergens.length > 0)

  return (
    <div className="max-w-4xl mx-auto py-8">
      <Link
        href="/products"
        className="text-brand-400 hover:text-brand-500 hover:underline text-sm"
      >
        ← Ürünlere dön
      </Link>

      <header className="mt-4 mb-8">
        <p className="text-sm uppercase tracking-wide text-mauve">
          {product.brand}
        </p>
        <h1 className="text-4xl font-bold mt-1">{product.name}</h1>
        <p className="text-lg font-semibold mt-2">
          {formatPrice(product.price, product.currency)}
          {product.category && (
            <span className="font-normal text-mist"> · {product.category}</span>
          )}
        </p>
        {product.description && (
          <p className="text-espresso/80 mt-3">{product.description}</p>
        )}
      </header>

      {/* Bir bakışta özet */}
      <div className="grid gap-4 sm:grid-cols-3 mb-10">
        <Stat label="İçerik sayısı" value={String(ingredients.length)} />
        <Stat
          label="Alerjen taşıyan"
          value={String(withAllergens.length)}
          tone={withAllergens.length > 0 ? 'clay' : 'sage'}
        />
        <Stat
          label="Yüksek riskli (7+)"
          value={String(riskiest.length)}
          tone={riskiest.length > 0 ? 'brand' : 'sage'}
        />
      </div>

      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">İçerik dökümü</h2>
        <p className="text-sm text-mauve mb-4">
          INCI sırasına göre listelenir — üstteki içerikler en yüksek oranda
          bulunanlardır.
        </p>
        {ingredients.length === 0 ? (
          <p className="text-mauve">
            Bu ürüne henüz içerik eşlenmemiş.
          </p>
        ) : (
          <div className="space-y-3">
            {ingredients.map((ing) => (
              <IngredientCard key={ing.id} ingredient={ing} />
            ))}
          </div>
        )}
      </section>

      {dupes.length > 0 && (
        <section>
          <h2 className="text-2xl font-bold mb-2">Muadiller ve alternatifler</h2>
          <p className="text-sm text-mauve mb-4">
            Bu ürünle içerik listesini ne kadar paylaştıklarına göre sıralandı.
          </p>
          <div className="grid gap-4 sm:grid-cols-2">
            {dupes.map((rec) => (
              <Link
                key={rec.id}
                href={`/products/${rec.id}`}
                className="border border-brand-100 rounded-xl p-5 bg-white hover:border-brand-300 transition"
              >
                <div className="flex items-center justify-between mb-2">
                  <span
                    className={`text-xs font-semibold px-2 py-1 rounded ${
                      rec.type === 'dupe'
                        ? 'bg-brand-100 text-brand-500'
                        : 'bg-mist/25 text-mauve'
                    }`}
                  >
                    {rec.type === 'dupe' ? 'muadil' : 'alternatif'}
                  </span>
                  <span className="text-sm font-semibold">
                    %{Math.round(rec.similarity_score * 100)} eşleşme
                  </span>
                </div>
                <p className="text-xs uppercase tracking-wide text-mauve">
                  {rec.brand}
                </p>
                <h3 className="font-semibold">{rec.name}</h3>
                <p className="text-sm font-semibold mt-1">
                  {formatPrice(rec.price, rec.currency)}
                  {rec.price < product.price && (
                    <span className="text-sage font-normal">
                      {' '}
                      · {formatPrice(product.price - rec.price, rec.currency)}{' '}
                      tasarruf
                    </span>
                  )}
                </p>
                <p className="text-sm text-espresso/70 mt-2">{rec.reason}</p>
              </Link>
            ))}
          </div>
        </section>
      )}
    </div>
  )
}

function Stat({
  label,
  value,
  tone = 'neutral',
}: {
  label: string
  value: string
  tone?: 'neutral' | 'sage' | 'clay' | 'brand'
}) {
  const tones = {
    neutral: 'bg-white border-brand-100',
    sage: 'bg-sage/10 border-sage/30',
    clay: 'bg-clay/15 border-clay/40',
    brand: 'bg-brand-100 border-brand-300',
  }

  return (
    <div className={`border rounded-xl p-4 ${tones[tone]}`}>
      <p className="text-2xl font-bold">{value}</p>
      <p className="text-sm text-espresso/70">{label}</p>
    </div>
  )
}
