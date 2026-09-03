// app/ingredients/page.tsx — filtrelenebilir içerik kaşifi
import Link from 'next/link'
import IngredientCard from '@/components/IngredientCard'
import { listIngredients, SKIN_TYPES, skinTypeLabel, ApiError } from '@/lib/api'
import type { Ingredient } from '@/lib/api'

export const metadata = {
  title: 'İçerikleri Keşfet · Clarity',
}

interface SearchParams {
  q?: string
  skin_type?: string
  avoid_allergens?: string
  max_concern?: string
}

export default async function IngredientsPage({
  searchParams,
}: {
  searchParams: SearchParams
}) {
  const { q = '', skin_type = '', avoid_allergens = '', max_concern = '' } =
    searchParams

  let ingredients: Ingredient[] = []
  let total = 0
  let error = ''

  try {
    const data = await listIngredients({
      q,
      skin_type,
      avoid_allergens,
      max_concern,
      limit: 100,
    })
    ingredients = data.ingredients
    total = data.total
  } catch (err) {
    error =
      err instanceof ApiError
        ? err.message
        : 'İçerikler yüklenirken bir sorun oluştu.'
  }

  const hasFilters = Boolean(q || skin_type || avoid_allergens || max_concern)

  return (
    <div className="max-w-4xl mx-auto py-8">
      <h1 className="text-4xl font-bold mb-2">İçerikleri Keşfet</h1>
      <p className="text-espresso/70 mb-8">
        Katalogda ara, cilt tipine göre süz ve tepki verdiğin alerjenleri
        taşıyan her şeyi gizle.
      </p>

      {/* Filtreler — düz bir GET formu, böylece sonuçlar URL ile paylaşılabilir. */}
      <form
        method="GET"
        className="mb-8 p-6 bg-white border border-brand-100 rounded-xl space-y-4"
      >
        <div>
          <label htmlFor="q" className="block text-sm font-semibold mb-1">
            Ara
          </label>
          <input
            id="q"
            name="q"
            defaultValue={q}
            placeholder="örn. niasinamid, aqua, retinol"
            className="w-full px-4 py-2 border border-brand-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-400"
          />
        </div>

        <div className="grid gap-4 sm:grid-cols-3">
          <div>
            <label htmlFor="skin_type" className="block text-sm font-semibold mb-1">
              Cilt tipi
            </label>
            <select
              id="skin_type"
              name="skin_type"
              defaultValue={skin_type}
              className="w-full px-4 py-2 border border-brand-200 rounded-lg bg-white"
            >
              <option value="">Farketmez</option>
              {SKIN_TYPES.map((t) => (
                <option key={t} value={t}>
                  {skinTypeLabel(t)}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label
              htmlFor="avoid_allergens"
              className="block text-sm font-semibold mb-1"
            >
              Kaçınılacak alerjenler
            </label>
            <input
              id="avoid_allergens"
              name="avoid_allergens"
              defaultValue={avoid_allergens}
              placeholder="parfüm, nikel"
              className="w-full px-4 py-2 border border-brand-200 rounded-lg"
            />
          </div>

          <div>
            <label
              htmlFor="max_concern"
              className="block text-sm font-semibold mb-1"
            >
              En yüksek risk seviyesi
            </label>
            <select
              id="max_concern"
              name="max_concern"
              defaultValue={max_concern}
              className="w-full px-4 py-2 border border-brand-200 rounded-lg bg-white"
            >
              <option value="">Farketmez</option>
              {[3, 5, 7].map((n) => (
                <option key={n} value={n}>
                  {n} ve altı
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="flex gap-3">
          <button
            type="submit"
            className="px-5 py-2 bg-brand-500 text-white font-semibold rounded-lg hover:bg-brand-600 transition"
          >
            Filtreleri uygula
          </button>
          {hasFilters && (
            <Link
              href="/ingredients"
              className="px-5 py-2 border border-brand-200 rounded-lg hover:bg-brand-50 transition"
            >
              Temizle
            </Link>
          )}
        </div>
      </form>

      {error ? (
        <div className="p-6 bg-brand-100 border border-brand-300 text-brand-500 rounded-xl">
          {error}
        </div>
      ) : (
        <>
          <p className="text-sm text-ink-muted mb-4">
            Filtrelerine {total} içerik uyuyor
            {ingredients.length < total &&
              ` (ilk ${ingredients.length} tanesi gösteriliyor)`}
          </p>

          {ingredients.length === 0 ? (
            <div className="p-8 text-center border border-dashed border-brand-200 rounded-xl text-mauve">
              Hiçbir içerik eşleşmedi. Bir filtreyi gevşetmeyi deneyin.
            </div>
          ) : (
            <div className="space-y-3">
              {ingredients.map((ing) => (
                <IngredientCard key={ing.id} ingredient={ing} />
              ))}
            </div>
          )}
        </>
      )}
    </div>
  )
}
