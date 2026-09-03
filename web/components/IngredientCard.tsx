// components/IngredientCard.tsx
import Link from 'next/link'
import { concernBadge, concernLabel, skinTypeLabel } from '@/lib/api'
import type { Ingredient } from '@/lib/api'

/**
 * Tek bir içerik satırı. `flagged`, kullanıcının bu içerikle eşleşen
 * alerjenlerini taşır ve kartı uyarı durumuna geçirir.
 */
export default function IngredientCard({
  ingredient,
  flagged = [],
}: {
  ingredient: Ingredient
  flagged?: string[]
}) {
  const isFlagged = flagged.length > 0

  return (
    <div
      className={`p-4 rounded-xl border ${
        isFlagged
          ? 'bg-brand-100 border-brand-400'
          : 'bg-white border-brand-100'
      }`}
    >
      <div className="flex items-start justify-between gap-4 mb-2">
        <div className="min-w-0">
          <Link
            href={`/ingredients/${ingredient.id}`}
            className="font-semibold hover:text-brand-500 hover:underline"
          >
            {ingredient.name}
          </Link>
          {ingredient.inci_name && ingredient.inci_name !== ingredient.name && (
            <p className="text-sm text-ink-muted truncate">
              INCI: {ingredient.inci_name}
            </p>
          )}
        </div>
        <span
          className={`shrink-0 px-2 py-1 rounded border text-xs font-semibold ${concernBadge(
            ingredient.concern_level,
          )}`}
          title={concernLabel(ingredient.concern_level)}
        >
          {ingredient.concern_level}/10
        </span>
      </div>

      {ingredient.description && (
        <p className="text-sm text-espresso/70 mb-2">{ingredient.description}</p>
      )}

      {ingredient.benefits.length > 0 && (
        <p className="text-sm text-sage">✓ {ingredient.benefits.join(', ')}</p>
      )}

      {ingredient.allergens.length > 0 && (
        <p className="text-sm text-cocoa mt-1">
          ⚠️ Alerjenler: {ingredient.allergens.join(', ')}
        </p>
      )}

      {ingredient.skin_types.length > 0 && (
        <p className="text-xs text-mist mt-1">
          Uygun: {ingredient.skin_types.map(skinTypeLabel).join(', ')}
        </p>
      )}

      {isFlagged && (
        <p className="text-sm font-semibold text-brand-500 mt-2">
          🚫 Profilinle eşleşiyor: {flagged.join(', ')}
        </p>
      )}
    </div>
  )
}
