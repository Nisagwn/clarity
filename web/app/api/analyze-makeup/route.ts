// app/api/analyze-makeup/route.ts
//
// Faz 1 ürün analizi. Henüz bir görüntü modeli yok; bu yüzden fotoğraftaki
// ürünü istemci bildiriyor (docs/SETUP.md'de planlandığı gibi elle eşleştirme)
// ve bu rota dökümü Go backend'den derliyor. Sunucu tarafında çalışması
// tarayıcıyı API kaynağından tamamen uzak tutuyor.
//
// Faz 2'de product_id argümanının yerini yüklenen görsel üzerinde çalışacak
// bir Vision API çağrısı alacak; aşağıdaki yanıt biçimi zaten o biçim.

import { NextResponse } from 'next/server'
import {
  allergenCheck,
  getProduct,
  getDupes,
  ApiError,
  type Ingredient,
} from '@/lib/api'

const MAX_IMAGE_BYTES = 10 * 1024 * 1024 // 10MB; yükleme sayfasındaki metinle aynı

export async function POST(request: Request) {
  let form: FormData
  try {
    form = await request.formData()
  } catch {
    return NextResponse.json(
      { error: 'Multipart form yüklemesi bekleniyordu' },
      { status: 400 },
    )
  }

  const image = form.get('image')
  if (!(image instanceof File)) {
    return NextResponse.json({ error: 'Lütfen bir görsel seçin' }, { status: 400 })
  }
  if (image.size > MAX_IMAGE_BYTES) {
    return NextResponse.json(
      { error: 'Bu görsel 10MB’dan büyük' },
      { status: 400 },
    )
  }
  if (!image.type.startsWith('image/')) {
    return NextResponse.json(
      { error: 'Bu dosya bir görsel değil' },
      { status: 400 },
    )
  }

  const productId = Number(form.get('product_id'))
  if (!Number.isInteger(productId) || productId <= 0) {
    return NextResponse.json(
      { error: 'Bu fotoğraftaki ürünü seçin' },
      { status: 400 },
    )
  }

  let userAllergens: string[] = []
  const rawAllergens = form.get('user_allergens')
  if (typeof rawAllergens === 'string' && rawAllergens.trim() !== '') {
    try {
      const parsed: unknown = JSON.parse(rawAllergens)
      if (Array.isArray(parsed)) {
        userAllergens = parsed.filter((a): a is string => typeof a === 'string')
      }
    } catch {
      return NextResponse.json(
        { error: 'user_allergens bir JSON dizisi olmalı' },
        { status: 400 },
      )
    }
  }

  try {
    const product = await getProduct(productId)
    const check = await allergenCheck(productId, userAllergens)

    // Yalnızca gerçekten bir işaretleme olduysa daha temiz seçenekler öner.
    let safeAlternatives: string[] = []
    if (!check.safe) {
      try {
        const { recommendations } = await getDupes(productId, 4)
        safeAlternatives = recommendations.map(
          (r) => `${r.brand} ${r.name} (%${Math.round(r.similarity_score * 100)} eşleşme)`,
        )
      } catch {
        safeAlternatives = []
      }
    }

    return NextResponse.json({
      product: {
        id: product.id,
        name: product.name,
        brand: product.brand,
        // 1 = model değil, elle eşleştirildi.
        confidence: 1,
        image_url: product.image_url ?? '',
      },
      ingredients: (product.ingredients ?? []).map((ing: Ingredient) => ({
        id: ing.id,
        name: ing.name,
        inci_name: ing.inci_name,
        concern_level: ing.concern_level,
        allergens: ing.allergens,
        benefits: ing.benefits,
      })),
      userMatches: {
        flaggedAllergens: check.matches.map(
          (m) => `${m.ingredient} (${m.allergen})`,
        ),
        safeAlternatives,
        // Tanınmayan terimler kullanıcıya gösterilmek zorunda: bir alerjeni
        // yanlış yazan kişi hiç uyarı almazsa ürünü güvenli sanır.
        unmatchedTerms: check.unmatched_terms ?? [],
        suggestions: check.suggestions ?? {},
      },
    })
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json(
        { error: err.message },
        { status: err.status === 0 ? 503 : err.status },
      )
    }
    return NextResponse.json({ error: 'Analiz başarısız oldu' }, { status: 500 })
  }
}
