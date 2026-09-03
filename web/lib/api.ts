// lib/api.ts — Go backend için tipli istemci.
//
// Alan adları API sözleşmesi gereği İngilizce kalır; arayüze giden etiketler
// bu dosyanın altındaki yardımcılarda Türkçeleştirilir.

export const API_URL =
  process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8090'

export interface Ingredient {
  id: number
  name: string
  inci_name: string
  description?: string
  concern_level: number
  skin_types: string[]
  allergens: string[]
  benefits: string[]
  products_count?: number
  order_index?: number
}

export interface Product {
  id: number
  name: string
  brand: string
  gtin?: string
  price: number
  currency: string
  image_url?: string
  category?: string
  description?: string
  source_url?: string
  ingredients?: Ingredient[]
  created_at: string
}

export interface UserProfile {
  id: number
  email?: string
  skin_type: string
  allergens: string[]
  created_at: string
}

export interface AllergenMatch {
  allergen: string
  ingredient: string
  severity: number
  concern_level: number
}

export interface AllergenCheckResult {
  product_id: number
  product_name: string
  matches: AllergenMatch[]
  safe: boolean
  flags: string[]
}

export interface Recommendation {
  id: number
  type: 'dupe' | 'alternative'
  name: string
  brand: string
  price: number
  currency: string
  image_url?: string
  similarity_score: number
  reason: string
}

// API değerleri (sözleşme gereği İngilizce) ve Türkçe arayüz etiketleri.
export const SKIN_TYPES = [
  'oily',
  'dry',
  'combination',
  'sensitive',
  'normal',
] as const

export const SKIN_TYPE_LABELS: Record<string, string> = {
  oily: 'yağlı',
  dry: 'kuru',
  combination: 'karma',
  sensitive: 'hassas',
  normal: 'normal',
  all: 'tüm cilt tipleri',
}

/** Bir cilt tipi API değerini Türkçe etikete çevirir. */
export function skinTypeLabel(value: string): string {
  return SKIN_TYPE_LABELS[value] ?? value
}

/** Backend 2xx dışı bir durum kodu döndürdüğünde fırlatılır. */
export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

/**
 * Backend'i çağırır ve JSON gövdesini açar. Sunucu bileşenleri `init` ile
 * `cache: 'no-store'` geçer; böylece sayfalar her zaman veritabanını yansıtır.
 */
export async function apiFetch<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  let res: Response
  try {
    res = await fetch(`${API_URL}${path}`, {
      ...init,
      headers: { Accept: 'application/json', ...init?.headers },
    })
  } catch {
    throw new ApiError(
      `API'ye ulaşılamıyor (${API_URL}). Go backend çalışıyor mu?`,
      0,
    )
  }

  if (!res.ok) {
    let message = `İstek ${res.status} durum koduyla başarısız oldu`
    try {
      const body = (await res.json()) as { error?: string }
      if (body.error) message = body.error
    } catch {
      // Gövde JSON değildi; durum koduna dayalı mesaj kalsın.
    }
    throw new ApiError(message, res.status)
  }

  return (await res.json()) as T
}

// ===== İçerikler =====

export interface IngredientListResponse {
  total: number
  limit: number
  offset: number
  ingredients: Ingredient[]
}

export function listIngredients(params: {
  q?: string
  skin_type?: string
  avoid_allergens?: string
  max_concern?: string
  limit?: number
}) {
  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') search.set(key, String(value))
  }
  return apiFetch<IngredientListResponse>(`/ingredients?${search}`, {
    cache: 'no-store',
  })
}

export function getIngredient(id: number | string) {
  return apiFetch<Ingredient>(`/ingredients/${id}`, { cache: 'no-store' })
}

export function allergenCheck(productId: number, userAllergens: string[]) {
  return apiFetch<AllergenCheckResult>('/ingredients/allergen-check', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ product_id: productId, user_allergens: userAllergens }),
  })
}

// ===== Ürünler =====

export interface ProductListResponse {
  total: number
  limit: number
  offset: number
  products: Product[]
}

export function listProducts(params: { q?: string; category?: string; limit?: number } = {}) {
  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') search.set(key, String(value))
  }
  return apiFetch<ProductListResponse>(`/products?${search}`, { cache: 'no-store' })
}

export function getProduct(id: number | string) {
  return apiFetch<Product>(`/products/${id}`, { cache: 'no-store' })
}

export interface DupesResponse {
  product_id: number
  product_name: string
  recommendations: Recommendation[]
}

export function getDupes(id: number | string, limit = 5) {
  return apiFetch<DupesResponse>(`/products/${id}/dupes?limit=${limit}`, {
    cache: 'no-store',
  })
}

// ===== Profiller =====

export function createProfile(body: {
  email?: string
  skin_type: string
  allergens: string[]
}) {
  return apiFetch<UserProfile>('/profiles', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
}

export function getProfile(id: number | string) {
  return apiFetch<UserProfile>(`/profiles/${id}`, { cache: 'no-store' })
}

// ===== Görüntüleme yardımcıları =====

/**
 * EWG endişe seviyesini marka paletinden seçilmiş sıralı bir ölçeğe eşler:
 * adaçayı (güvenli) → terrakota (orta) → şarap (yüksek). Renkler paletin
 * içinden geliyor ama sıcaklık arttıkça uyarı şiddeti de artıyor.
 */
export function concernColor(level: number): string {
  if (level <= 3) return 'text-sage'
  if (level <= 6) return 'text-clay'
  return 'text-brand-500'
}

/** EWG endişe seviyesini rozet stiline eşler. */
export function concernBadge(level: number): string {
  if (level <= 3) return 'bg-sage/10 text-sage border-sage/30'
  if (level <= 6) return 'bg-clay/15 text-cocoa border-clay/40'
  return 'bg-brand-100 text-brand-500 border-brand-300'
}

export function concernLabel(level: number): string {
  if (level <= 3) return 'Düşük risk'
  if (level <= 6) return 'Orta risk'
  return 'Yüksek risk'
}

export function formatPrice(price: number, currency: string): string {
  try {
    return new Intl.NumberFormat('tr-TR', {
      style: 'currency',
      currency: currency || 'USD',
    }).format(price)
  } catch {
    return `${price.toFixed(2)} ${currency}`
  }
}
