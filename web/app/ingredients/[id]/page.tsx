// app/ingredients/[id]/page.tsx — tek içerik detayı
import Link from 'next/link'
import { notFound } from 'next/navigation'
import {
  getIngredient,
  concernBadge,
  concernLabel,
  skinTypeLabel,
  ApiError,
} from '@/lib/api'
import type { Ingredient } from '@/lib/api'

export default async function IngredientDetailPage({
  params,
}: {
  params: { id: string }
}) {
  let ingredient: Ingredient
  try {
    ingredient = await getIngredient(params.id)
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound()
    throw err
  }

  return (
    <div className="max-w-3xl mx-auto py-8">
      <Link
        href="/ingredients"
        className="text-brand-400 hover:text-brand-500 hover:underline text-sm"
      >
        ← İçeriklere dön
      </Link>

      <div className="mt-4 flex items-start justify-between gap-6">
        <div>
          <h1 className="text-4xl font-bold">{ingredient.name}</h1>
          {ingredient.inci_name && (
            <p className="text-mauve mt-1">INCI: {ingredient.inci_name}</p>
          )}
        </div>
        <span
          className={`shrink-0 px-3 py-2 rounded-lg border font-bold ${concernBadge(
            ingredient.concern_level,
          )}`}
        >
          {ingredient.concern_level}/10
        </span>
      </div>

      <p className="text-sm text-mist mt-2">
        EWG 1–10 ölçeğinde {concernLabel(ingredient.concern_level).toLowerCase()}
        {ingredient.products_count !== undefined &&
          ` · katalogdaki ${ingredient.products_count} üründe bulunuyor`}
      </p>

      {ingredient.description && (
        <p className="mt-6 text-espresso/80 leading-relaxed">
          {ingredient.description}
        </p>
      )}

      <div className="grid gap-6 sm:grid-cols-3 mt-8">
        <Section title="Faydaları" items={ingredient.benefits} tone="sage" />
        <Section title="Alerjenler" items={ingredient.allergens} tone="brand" />
        <Section
          title="Uygun cilt tipleri"
          items={ingredient.skin_types.map(skinTypeLabel)}
          tone="clay"
        />
      </div>

      <p className="mt-10 text-xs text-mist border-t border-brand-100 pt-4">
        Risk seviyeleri EWG geleneğini izler; tıbbi tavsiye değil, yol
        göstericidir. Yeni ürünleri küçük bir alanda deneyin ve geçmeyen
        reaksiyonlar için bir dermatoloğa danışın.
      </p>
    </div>
  )
}

function Section({
  title,
  items,
  tone,
}: {
  title: string
  items: string[]
  tone: 'sage' | 'brand' | 'clay'
}) {
  const tones = {
    sage: 'bg-sage/10 text-sage',
    brand: 'bg-brand-100 text-brand-500',
    clay: 'bg-clay/15 text-cocoa',
  }

  return (
    <div>
      <h2 className="font-semibold mb-2">{title}</h2>
      {items.length === 0 ? (
        <p className="text-sm text-mist">Kayıtlı bilgi yok</p>
      ) : (
        <ul className="flex flex-wrap gap-2">
          {items.map((item) => (
            <li key={item} className={`px-2 py-1 rounded text-sm ${tones[tone]}`}>
              {item}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
