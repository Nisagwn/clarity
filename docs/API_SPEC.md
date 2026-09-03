# Clarity — API Spesifikasyonu

## Temel adres
```
http://localhost:8090
```

**Not:** JSON alan adları, rota yolları ve `skin_type` değerleri sözleşmenin
parçasıdır ve İngilizce kalır. Arayüz bunları Türkçe etiketlerle gösterir.
Hata mesajları son kullanıcıya gittiği için Türkçedir.

---

## Uç noktalar (Faz 1)

### Sağlık kontrolü
```
GET /health
Yanıt: { "status": "ok", "database": "ok" }
```
Veritabanına ulaşılamıyorsa `503` ve `{"status":"degraded","database":"unreachable"}`.

---

## Kullanıcı profilleri

### Profil oluştur
```
POST /profiles
Content-Type: application/json

{
  "email": "kullanici@ornek.com",
  "skin_type": "sensitive",
  "allergens": ["nikel", "formaldehit"]
}

Yanıt 201:
{
  "id": 1,
  "email": "kullanici@ornek.com",
  "skin_type": "sensitive",
  "allergens": ["nikel", "formaldehit"],
  "created_at": "2026-09-03T10:30:00Z"
}
```
Alerjenler kırpılır, küçük harfe çevrilir ve tekilleştirilir.
Aynı e-posta ikinci kez gönderilirse `409`.

### Profil getir
```
GET /profiles/:id

Yanıt 200:
{
  "id": 1,
  "email": "kullanici@ornek.com",
  "skin_type": "sensitive",
  "allergens": ["nikel", "formaldehit"],
  "created_at": "2026-09-03T10:30:00Z"
}
```

### Profil güncelle
```
PUT /profiles/:id
{ "skin_type": "dry", "allergens": ["parfüm"] }
```
Alerjen listesinin tamamını değiştirir.

---

## Ürünler

### Ürün oluştur
```
POST /products
Content-Type: application/json

{
  "name": "Işıltılı İpek Fondöten SPF 20",
  "brand": "Maison Lumiere",
  "gtin": "8690000000011",
  "price": 52.00,
  "currency": "USD",
  "category": "fondöten",
  "image_url": "https://ornek.com/gorsel.jpg"
}

Yanıt 201: oluşturulan ürün, id ve created_at ile
```
`name` ve `brand` zorunludur. Aynı markada aynı GTIN varsa `409`.

### Ürünleri listele
```
GET /products?q=&brand=&category=&limit=50&offset=0

Yanıt 200:
{ "total": 14, "limit": 50, "offset": 0, "products": [...] }
```

### İçerikleriyle ürün getir
```
GET /products/:id

Yanıt 200:
{
  "id": 1,
  "name": "Işıltılı İpek Fondöten SPF 20",
  "brand": "Maison Lumiere",
  "price": 52,
  "currency": "USD",
  "ingredients": [
    {
      "id": 1,
      "name": "Su",
      "inci_name": "Aqua",
      "concern_level": 1,
      "benefits": ["çözücü"],
      "allergens": [],
      "order_index": 1
    },
    {
      "id": 34,
      "name": "Parfüm",
      "inci_name": "Parfum (Fragrance)",
      "concern_level": 8,
      "benefits": [],
      "allergens": ["parfüm"],
      "order_index": 11
    }
  ],
  "created_at": "..."
}
```
İçerikler `order_index`, yani INCI listesindeki sıraya göre döner.

### Ürün ara
```
GET /products/search?q=retinol

Yanıt 200:
{ "query": "retinol", "count": 1, "results": [...] }
```
Ürün adı, marka **ve içerik adları** üzerinde arar.

### Markaya göre ürünler
```
GET /products/brand/Rue%20Belle

Yanıt 200:
{ "brand": "Rue Belle", "count": 5, "products": [...] }
```

### Ürüne içerik bağla
```
POST /products/:id/ingredients
{ "ingredient_ids": [1, 21, 12, 2] }

Yanıt 200: { "product_id": 1, "linked": 4 }
```
Dizideki sıra `order_index` olarak kaydedilir.

---

## İçerikler

### İçerikleri listele
```
GET /ingredients?skin_type=sensitive&avoid_allergens=nikel,parfüm&max_concern=5&limit=50&offset=0

Yanıt 200:
{
  "skin_type": "sensitive",
  "avoid_allergens": ["nikel", "parfüm"],
  "total": 11,
  "limit": 50,
  "offset": 0,
  "ingredients": [
    {
      "id": 1,
      "name": "Su",
      "inci_name": "Aqua",
      "concern_level": 1,
      "skin_types": ["all"],
      "benefits": ["çözücü"],
      "allergens": []
    }
  ]
}
```

| Parametre | Etkisi |
|-----------|--------|
| `q` | Ad ve INCI adı üzerinde serbest metin araması |
| `skin_type` | O cilt tipine uygun (veya `all` etiketli) içerikleri tutar |
| `avoid_allergens` | Virgülle ayrılır; bu alerjenleri taşıyan içerikleri eler |
| `max_concern` | Risk seviyesi üst sınırı |
| `limit` / `offset` | Sayfalama (varsayılan 50, en fazla 200) |

`skin_type` geçersizse `400` ve geçerli değerleri Türkçe karşılıklarıyla
listeleyen bir mesaj döner.

### Tek içerik getir
```
GET /ingredients/:id

Yanıt 200:
{
  "id": 34,
  "name": "Parfüm",
  "inci_name": "Parfum (Fragrance)",
  "description": "İçeriği açıklanmayan koku karışımı; en yaygın kozmetik alerjeni.",
  "concern_level": 8,
  "skin_types": [],
  "benefits": [],
  "allergens": ["parfüm"],
  "products_count": 5
}
```

---

## Alerjen kontrolü

```
POST /ingredients/allergen-check
Content-Type: application/json

{
  "product_id": 1,
  "user_allergens": ["parfüm", "nikel"]
}

Yanıt 200:
{
  "product_id": 1,
  "product_name": "Işıltılı İpek Fondöten SPF 20",
  "matches": [
    { "allergen": "parfüm", "ingredient": "Parfüm",         "severity": 8, "concern_level": 8 },
    { "allergen": "parfüm", "ingredient": "Linalool",       "severity": 6, "concern_level": 5 },
    { "allergen": "nikel",  "ingredient": "Mika",           "severity": 4, "concern_level": 2 },
    { "allergen": "nikel",  "ingredient": "Demir Oksitler", "severity": 3, "concern_level": 1 }
  ],
  "safe": false,
  "flags": ["parfüm içeriyor", "nikel içeriyor"]
}
```

> **Bilinen sorun.** Eşleştirme şu an iki yönlü alt dize karşılaştırmasıyla
> yapılıyor ve yanlış pozitif üretebiliyor: `alkol` arayan bir kullanıcı
> "yün alkolü" yüzünden Lanolin için uyarı alıyor. Bkz.
> [DEVELOPMENT_PLAN.md](DEVELOPMENT_PLAN.md) P0-1.

---

## Öneriler

### Muadil ve alternatifler
```
POST /recommendations
{ "product_id": 1, "user_profile_id": 1, "top_n": 5 }

Yanıt 200:
{
  "product_id": 1,
  "recommendations": [
    {
      "id": 2,
      "type": "dupe",
      "name": "Günlük Işıltı Fondöten SPF 15",
      "brand": "Rue Belle",
      "price": 18.5,
      "currency": "USD",
      "similarity_score": 0.8333,
      "reason": "%83 içerik örtüşmesi, 33.50 USD daha ucuz"
    }
  ]
}
```

Puanlama, içerik kümeleri üzerinde **Jaccard benzerliğidir**: ortak içerik
sayısı bölü birleşim kümesinin büyüklüğü. Aynı kategoride 0,5 ve üzeri
eşleşmeler `dupe`, diğer örtüşmeler `alternative` olarak etiketlenir.
`user_profile_id` verildiğinde, o profilin alerjenlerini taşıyan adaylar
tamamen elenir.

### Ürün sayfası için kısayol
```
GET /products/:id/dupes?limit=5

Yanıt 200:
{ "product_id": 1, "product_name": "...", "recommendations": [...] }
```

---

## Veri modelleri

Kolon adları veritabanı şemasıyla birebir aynıdır; ayrıntı için
`backend/db/schema.sql`.

### Product
```
id, name, brand, gtin, price, currency, image_url,
category, description, source_url, created_at, updated_at
```

### Ingredient
```
id, name (Türkçe yaygın ad), inci_name (uluslararası standart),
description, concern_level (1-10 EWG), created_at, updated_at
```

### ProductIngredient
```
id, product_id, ingredient_id, percentage, order_index, created_at
unique(product_id, ingredient_id)
```

### UserProfile / UserAllergen / IngredientAllergen
```
UserProfile:        id, email, skin_type, created_at, updated_at
UserAllergen:       id, user_id, allergen_name, severity  unique(user_id, allergen_name)
IngredientAllergen: id, ingredient_id, allergen_name, severity
                    unique(ingredient_id, allergen_name)
```

`skin_type` değerleri: `oily`, `dry`, `combination`, `sensitive`, `normal`.

---

## Hata yanıtları

### 400 Bad Request
```json
{ "error": "Geçersiz skin_type. Geçerli değerler: oily (yağlı), dry (kuru), combination (karma), sensitive (hassas), normal (normal)" }
```

### 404 Not Found
```json
{ "error": "Ürün bulunamadı" }
```

### 409 Conflict
```json
{ "error": "Bu e-postayla bir profil zaten var" }
```

### 500 Internal Server Error
```json
{ "error": "..." }
```

---

## cURL ile deneme

```bash
# Profil oluştur
curl -X POST http://localhost:8090/profiles \
  -H "Content-Type: application/json; charset=utf-8" \
  -d '{"email":"test@ornek.com","skin_type":"sensitive","allergens":["nikel"]}'

# Alerjen kontrolü
curl -X POST http://localhost:8090/ingredients/allergen-check \
  -H "Content-Type: application/json; charset=utf-8" \
  -d '{"product_id":1,"user_allergens":["parfüm","nikel"]}'

# Muadiller
curl "http://localhost:8090/products/1/dupes?limit=3"
```

Türkçe karakter gönderirken gövdenin UTF-8 olduğundan emin olun; bazı
kabuklar farklı bir kod sayfasıyla gönderip eşleşmeyi sessizce bozar.

---

## Faz 2

- **Kimlik doğrulama.** Şu anda yok. `GET/PUT /profiles/:id` çıplak bir tam
  sayı aldığı için herkese açık dağıtımdan önce oturum zorunlu.
- **Hız sınırlama.** IP başına 100 istek/dakika, anahtar başına 1000/saat.
- **API anahtarları.** `X-API-Key` başlığı; web için JWT.

## Performans hedefleri
- `GET /products/:id`: < 50 ms
- `GET /ingredients`: < 100 ms (filtrelerle)
- `POST /recommendations`: < 500 ms
- `POST /ingredients/allergen-check`: < 100 ms

Arama şu anda `ILIKE '%terim%'` kullanıyor ve indeks kullanamıyor; katalog
büyüdükçe bu hedefler tutmaz. Bkz. DEVELOPMENT_PLAN P1-5.
