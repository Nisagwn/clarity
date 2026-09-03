// app/upload/page.tsx
'use client'

import { useState, useRef, useEffect } from 'react'
import Image from 'next/image'
import Link from 'next/link'
import { listProducts, getMyProfile, concernColor, type Product } from '@/lib/api'

interface DetectedProduct {
  id: number
  name: string
  brand: string
  confidence: number
  image_url: string
}

interface IngredientsResult {
  product: DetectedProduct
  ingredients: Array<{
    id: number
    name: string
    inci_name: string
    concern_level: number
    allergens: string[]
    benefits: string[]
  }>
  userMatches: {
    flaggedAllergens: string[]
    safeAlternatives: string[]
    unmatchedTerms: string[]
    suggestions: Record<string, string[]>
  }
}

export default function UploadPage() {
  const [image, setImage] = useState<File | null>(null)
  const [preview, setPreview] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<IngredientsResult | null>(null)
  const [error, setError] = useState<string>('')
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [userAllergens, setUserAllergens] = useState<string[]>([])
  const [allergenText, setAllergenText] = useState('')

  // Faz 1'de eşleştirme elle yapılır, bu yüzden kullanıcı katalogdan seçer.
  const [products, setProducts] = useState<Product[]>([])
  const [productId, setProductId] = useState('')
  const [catalogueError, setCatalogueError] = useState('')

  useEffect(() => {
    listProducts({ limit: 200 })
      .then((data) => setProducts(data.products))
      .catch(() =>
        setCatalogueError(
          'Ürün kataloğu yüklenemedi. Backend çalışıyor mu?',
        ),
      )

    // Oturumdaki profilden doldur; alerjenler sayfalar arasında taşınsın.
    // Oturum yoksa 401 döner, bu beklenen durum: giriş yapmadan da
    // alerjen yazıp analiz yapılabilir, sadece kaydedilmez.
    getMyProfile()
      .then((profile) => {
        setUserAllergens(profile.allergens)
        setAllergenText(profile.allergens.join(', '))
      })
      .catch(() => {
        /* Giriş yapılmamış; elle giriş alanı zaten var. */
      })
  }, [])

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    setImage(file)
    setError('')
    setResult(null)

    const reader = new FileReader()
    reader.onload = (event) => {
      setPreview(event.target?.result as string)
    }
    reader.readAsDataURL(file)
  }

  const handleAnalyze = async () => {
    if (!image) {
      setError('Lütfen bir görsel seçin')
      return
    }
    if (!productId) {
      setError('Lütfen bu fotoğraftaki ürünü seçin')
      return
    }

    setLoading(true)
    setError('')

    try {
      const formData = new FormData()
      formData.append('image', image)
      formData.append('product_id', productId)
      formData.append('user_allergens', JSON.stringify(userAllergens))

      const response = await fetch('/api/analyze-makeup', {
        method: 'POST',
        body: formData,
      })

      const data = await response.json()
      if (!response.ok) {
        throw new Error(data.error ?? 'Analiz başarısız oldu')
      }

      setResult(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Bir hata oluştu')
    } finally {
      setLoading(false)
    }
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    e.currentTarget.classList.add('border-brand-400')
  }

  const handleDragLeave = (e: React.DragEvent) => {
    e.currentTarget.classList.remove('border-brand-400')
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    e.currentTarget.classList.remove('border-brand-400')
    const files = e.dataTransfer.files
    if (files.length > 0) {
      const fakeEvent = {
        target: { files },
      } as unknown as React.ChangeEvent<HTMLInputElement>
      handleImageUpload(fakeEvent)
    }
  }

  return (
    <div className="max-w-3xl mx-auto py-12">
      <div className="mb-8">
        <h1 className="text-4xl font-bold mb-2">Makyaj Fotoğrafını Yükle</h1>
        <p className="text-espresso/70">
          Ürününün ya da ambalajının fotoğrafını çek, hangi ürün olduğunu seç;
          içerikleri çözüp alerjenlerini işaretleyelim.
        </p>
        <p className="text-sm text-mist mt-2">
          Otomatik ürün tanıma Faz 2&apos;de geliyor — şimdilik eşleştirme elle
          yapılıyor.
        </p>
      </div>

      {/* Alerjen girişi */}
      <div className="mb-8 p-6 bg-brand-100/60 rounded-xl border border-brand-200">
        <h2 className="font-semibold mb-3">Bilinen alerjenlerin (isteğe bağlı)</h2>
        <p className="text-sm text-espresso/70 mb-4">
          Alerjin veya hassasiyetin olan içerikleri gir; sonuçlarda
          işaretleyelim.{' '}
          <Link href="/profile" className="text-brand-500 hover:underline">
            Bir profil kaydet
          </Link>{' '}
          ve tekrar tekrar yazma.
        </p>
        <input
          type="text"
          value={allergenText}
          placeholder="örn. formaldehit, nikel, parfüm (virgülle ayır)"
          onChange={(e) => {
            setAllergenText(e.target.value)
            setUserAllergens(
              e.target.value
                .split(',')
                .map((s) => s.trim())
                .filter(Boolean),
            )
          }}
          className="w-full px-4 py-2 border border-brand-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-400"
        />
      </div>

      {/* Ürün seçici */}
      <div className="mb-8">
        <label htmlFor="product" className="block font-semibold mb-2">
          Bu hangi ürün?
        </label>
        {catalogueError ? (
          <div className="p-4 bg-clay/10 border border-clay/40 text-cocoa rounded-lg text-sm">
            {catalogueError}
          </div>
        ) : (
          <select
            id="product"
            value={productId}
            onChange={(e) => setProductId(e.target.value)}
            className="w-full px-4 py-2 border border-brand-200 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-brand-400"
          >
            <option value="">Bir ürün seçin…</option>
            {products.map((p) => (
              <option key={p.id} value={p.id}>
                {p.brand} — {p.name}
              </option>
            ))}
          </select>
        )}
      </div>

      {/* Yükleme alanı */}
      <div
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={() => fileInputRef.current?.click()}
        className="mb-8 border-2 border-dashed border-brand-200 rounded-xl p-12 text-center cursor-pointer hover:border-brand-400 transition"
      >
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          onChange={handleImageUpload}
          className="hidden"
        />
        {preview ? (
          <div className="relative w-full h-64">
            <Image
              src={preview}
              alt="Önizleme"
              fill
              unoptimized
              className="object-contain"
            />
          </div>
        ) : (
          <div>
            <svg className="mx-auto h-12 w-12 text-mist mb-4" stroke="currentColor" fill="none" viewBox="0 0 48 48">
              <path d="M28 8H12a4 4 0 00-4 4v20a4 4 0 004 4h24a4 4 0 004-4V20m-14-8l-4-4m0 0l-4 4m4-4v12m8-8h8m0 0v8" strokeWidth={2} />
            </svg>
            <p className="text-espresso/70 mb-2">
              Görselini buraya sürükle ya da seçmek için tıkla
            </p>
            <p className="text-sm text-mist">PNG, JPG, WEBP — en fazla 10MB</p>
          </div>
        )}
      </div>

      {error && (
        <div className="mb-6 p-4 bg-brand-100 border border-brand-300 text-brand-500 rounded-lg">
          {error}
        </div>
      )}

      {/* Analiz düğmesi */}
      <button
        onClick={handleAnalyze}
        disabled={!preview || !productId || loading}
        className="w-full py-3 bg-brand-500 text-white font-semibold rounded-lg hover:bg-brand-500 disabled:opacity-50 disabled:cursor-not-allowed transition mb-12"
      >
        {loading ? 'Analiz ediliyor...' : 'İçerikleri analiz et'}
      </button>

      {/* Sonuçlar */}
      {result && (
        <div className="space-y-8">
          {/* Eşleşen ürün */}
          <section className="border border-brand-100 rounded-xl p-6 bg-white">
            <h2 className="text-2xl font-bold mb-4">Eşleşen ürün</h2>
            <div className="flex gap-6">
              {result.product.image_url && (
                <div className="relative w-40 h-40">
                  <Image
                    src={result.product.image_url}
                    alt={result.product.name}
                    fill
                    className="object-cover rounded"
                  />
                </div>
              )}
              <div>
                <h3 className="text-xl font-semibold">{result.product.name}</h3>
                <p className="text-mauve">{result.product.brand}</p>
                <p className="text-sm text-mist mt-2">
                  {result.product.confidence === 1
                    ? 'Elle eşleştirildi'
                    : `Güven: %${(result.product.confidence * 100).toFixed(1)}`}
                </p>
                <Link
                  href={`/products/${result.product.id}`}
                  className="text-brand-400 hover:text-brand-500 hover:underline text-sm mt-4 inline-block"
                >
                  Tüm ürün detaylarını gör →
                </Link>
              </div>
            </div>
          </section>

          {/* İçerik dökümü */}
          <section className="border border-brand-100 rounded-xl p-6 bg-white">
            <h2 className="text-2xl font-bold mb-4">İçerik dökümü</h2>
            <div className="space-y-3">
              {result.ingredients.map((ing) => (
                <div
                  key={ing.id}
                  className="p-4 bg-brand-50 rounded-lg border border-brand-100"
                >
                  <div className="flex items-start justify-between mb-2">
                    <div>
                      <h4 className="font-semibold">{ing.name}</h4>
                      {ing.inci_name && ing.inci_name !== ing.name && (
                        <p className="text-sm text-ink-muted">{ing.inci_name}</p>
                      )}
                    </div>
                    <div className="text-right">
                      <div className="text-sm font-semibold">Güvenlik puanı</div>
                      <div
                        className={`text-xl font-bold ${concernColor(
                          ing.concern_level,
                        )}`}
                      >
                        {ing.concern_level}/10
                      </div>
                    </div>
                  </div>
                  {ing.benefits.length > 0 && (
                    <p className="text-sm text-sage mb-2">
                      ✓ {ing.benefits.join(', ')}
                    </p>
                  )}
                  {ing.allergens.length > 0 && (
                    <p className="text-sm text-cocoa">
                      ⚠️ Alerjenler: {ing.allergens.join(', ')}
                    </p>
                  )}
                </div>
              ))}
            </div>
          </section>

          {/* Profil eşleşmeleri */}
          <section className="border border-brand-100 rounded-xl p-6 bg-white">
            <h2 className="text-2xl font-bold mb-4">Senin profiline göre</h2>
            {result.userMatches.flaggedAllergens.length > 0 ? (
              <div className="mb-4 p-4 bg-brand-100 border border-brand-300 rounded-lg">
                <h3 className="font-semibold text-brand-500 mb-2">
                  ⚠️ İşaretlenen içerikler
                </h3>
                <p className="text-sm text-brand-500">
                  Bu ürün şunları içeriyor:{' '}
                  {result.userMatches.flaggedAllergens.join(', ')}
                </p>
              </div>
            ) : (
              <div className="mb-4 p-4 bg-sage/10 border border-sage/30 rounded-lg">
                <h3 className="font-semibold text-sage mb-2">
                  ✓ İşaretlenen alerjen yok
                </h3>
                <p className="text-sm text-sage">
                  {userAllergens.length === 0
                    ? 'Bu ürünü kendi alerjenlerine karşı kontrol etmek için yukarıya ekle.'
                    : 'Listelediğin alerjenlerin hiçbiri bu üründe bulunmuyor.'}
                </p>
              </div>
            )}
            {result.userMatches.unmatchedTerms.length > 0 && (
              <div
                role="alert"
                className="mb-4 p-4 bg-clay/10 border border-clay/40 rounded-lg"
              >
                <h3 className="font-semibold text-cocoa mb-2">
                  Bu terimleri tanıyamadık
                </h3>
                <p className="text-sm text-cocoa mb-2">
                  Şunlar alerjen listemizde bulunamadı, dolayısıyla bu ürüne
                  karşı <strong>kontrol edilmediler</strong>:{' '}
                  {result.userMatches.unmatchedTerms.join(', ')}
                </p>
                {Object.entries(result.userMatches.suggestions).map(
                  ([term, options]) => (
                    <p key={term} className="text-sm text-cocoa">
                      “{term}” için şunu mu demek istediniz:{' '}
                      <strong>{options.join(', ')}</strong>?
                    </p>
                  ),
                )}
              </div>
            )}
            {result.userMatches.safeAlternatives.length > 0 && (
              <div className="p-4 bg-sage/10 border border-sage/30 rounded-lg">
                <h3 className="font-semibold text-sage mb-2">
                  Değerlendirebileceğin benzer ürünler
                </h3>
                <ul className="text-sm text-sage list-disc list-inside space-y-1">
                  {result.userMatches.safeAlternatives.map((alt) => (
                    <li key={alt}>{alt}</li>
                  ))}
                </ul>
              </div>
            )}
          </section>
        </div>
      )}
    </div>
  )
}
