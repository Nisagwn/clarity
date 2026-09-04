// app/ingredients/[id]/page.tsx — tek içerik detayı
import Link from 'next/link'
import { notFound } from 'next/navigation'
import {
  getIngredient,
  concernBadge,
  concernLabel,
  concernScore,
  skinTypeLabel,
  ApiError,
} from '@/lib/api'
import type { Ingredient, ScoreExplanation } from '@/lib/api'

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

  const scored = ingredient.concern_level !== null

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
            <p className="text-ink-muted mt-1">INCI: {ingredient.inci_name}</p>
          )}
        </div>
        <span
          className={`shrink-0 px-3 py-2 rounded-lg border font-bold ${concernBadge(
            ingredient.concern_level,
          )}`}
        >
          <span className="sr-only">
            {concernLabel(ingredient.concern_level)}:{' '}
          </span>
          {concernScore(ingredient.concern_level)}
        </span>
      </div>

      <p className="text-sm text-ink-muted mt-2">
        {scored
          ? `1–10 ölçeğinde ${concernLabel(ingredient.concern_level).toLowerCase()}`
          : 'Bu içerik henüz puanlanmadı'}
        {ingredient.products_count !== undefined &&
          ` · katalogdaki ${ingredient.products_count} üründe bulunuyor`}
      </p>

      {ingredient.scoring ? (
        <ScoreBreakdown scoring={ingredient.scoring} />
      ) : (
        <Unscored />
      )}

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

      <p className="mt-10 text-xs text-ink-muted border-t border-brand-100 pt-4">
        Puanlar AB Tüzüğü 1223/2009 Eklerinden türetilir; tıbbi tavsiye değil,
        yol göstericidir. Yüksek bir puan, içeriğin bitmiş üründeki oranında
        tehlikeli olduğu anlamına gelmez — mevzuatın onu kısıtladığı anlamına
        gelir. Yeni ürünleri küçük bir alanda deneyin ve geçmeyen reaksiyonlar
        için bir dermatoloğa danışın.
      </p>
    </div>
  )
}

/**
 * "Neden bu puan?" — puanı üreten kurallar ve mevzuat atfı.
 *
 * Puanı gerekçesiz göstermek, Faz 3'te terk edilen elle atanmış puanlardan
 * farksız olurdu: kullanıcı sayıya bakıp neye dayandığını soramaz.
 */
function ScoreBreakdown({ scoring }: { scoring: ScoreExplanation }) {
  const { regulatory } = scoring

  return (
    <details className="mt-4 rounded-xl border border-brand-100 bg-white">
      <summary className="px-4 py-3 cursor-pointer font-semibold text-brand-500 hover:text-brand-600">
        Neden bu puan?
      </summary>

      <div className="px-4 pb-4 space-y-4 text-sm">
        <ul className="space-y-2">
          {scoring.rules.map((rule) => (
            <li key={rule.key} className="flex gap-3">
              <span className="shrink-0 font-bold tabular-nums text-espresso">
                {rule.key === 'sccs_adverse_modifier' ? `+${rule.score}` : rule.score}
              </span>
              <span className="text-espresso/80">{rule.rationale}</span>
            </li>
          ))}
          <li className="flex gap-3 border-t border-brand-100 pt-2">
            <span className="shrink-0 font-bold tabular-nums text-espresso">
              {scoring.value}
            </span>
            <span className="text-espresso/80">toplam (rubrik v{scoring.version})</span>
          </li>
        </ul>

        {regulatory.restriction && (
          <p className="text-espresso/80">
            <span className="font-semibold">Kısıtlama: </span>
            {regulatory.restriction}
            {regulatory.max_concentration !== undefined &&
              ` (en fazla %${regulatory.max_concentration})`}
          </p>
        )}

        {regulatory.sccs_opinion && (
          <p className="text-espresso/80">
            <span className="font-semibold">SCCS görüşü: </span>
            {regulatory.sccs_opinion}
          </p>
        )}

        <div>
          <h3 className="font-semibold mb-1">Kaynaklar</h3>
          <ul className="space-y-1">
            {scoring.sources.map((source) => (
              <li key={source} className="text-ink-muted break-words">
                {source}
              </li>
            ))}
          </ul>
          <a
            href={regulatory.source_url}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-block mt-2 text-brand-400 hover:text-brand-500 hover:underline"
          >
            Mevzuat metnini aç ↗
          </a>
        </div>

        {(regulatory.cas_number || regulatory.ec_number) && (
          <p className="text-ink-muted">
            {regulatory.cas_number && `CAS ${regulatory.cas_number}`}
            {regulatory.cas_number && regulatory.ec_number && ' · '}
            {regulatory.ec_number && `EC ${regulatory.ec_number}`}
          </p>
        )}
      </div>
    </details>
  )
}

/**
 * Puanı olmayan içerik. Boş bırakmak yerine nedenini söylüyoruz: kullanıcı
 * puanın eksik olduğunu bilmezse, olmamasını "sorun yok" diye okur.
 */
function Unscored() {
  return (
    <div className="mt-4 rounded-xl border border-brand-100 bg-brand-50 px-4 py-3 text-sm text-espresso/80">
      <p>
        <span className="font-semibold">Henüz puanlanmadı. </span>
        Bu içeriğin AB Tüzüğü 1223/2009 Eklerinde kayıtlı bir kısıtlaması
        bulunamadı; puan üretmek için dayanağımız yok. Puan uydurmaktansa boş
        bırakıyoruz.
      </p>
      <Link
        href="/about"
        className="inline-block mt-2 text-brand-400 hover:text-brand-500 hover:underline"
      >
        Puanlar nasıl hesaplanıyor? →
      </Link>
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
        <p className="text-sm text-ink-muted">Kayıtlı bilgi yok</p>
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
