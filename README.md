# Clarity 🌸

Makyaj içeriklerini analiz eder, kişisel cilt profiline göre alerjenleri
işaretler ve benzer içerik listesine sahip daha uygun fiyatlı ürünleri bulur.
Go + Next.js + PostgreSQL.

## Hızlı Başlangıç

### Gereksinimler
- Go 1.21+
- Node.js 18+
- Docker (PostgreSQL'i çalıştırır — yerel kuruluma gerek yok)

### Çalıştırma

```bash
# 1. Veritabanı + şema göçleri + örnek veri
make up && make seed

# 2. Backend
cd backend
cp .env.example .env      # temiz bir kopyada zaten mevcut
go mod download
go run .                  # http://localhost:8090

# 3. Ön yüz (ayrı bir terminalde)
cd web
npm install
npm run dev               # http://localhost:3001
```

Make ile: önce `make setup`, sonra iki ayrı terminalde `make backend` ve
`make frontend`.

### Portlar

| Servis     | Adres / Port                 |
|------------|------------------------------|
| Web        | http://localhost:3001        |
| API        | http://localhost:8090        |
| PostgreSQL | `127.0.0.1:5433` (localhost değil — bkz. SETUP) |

3000 ve 8080 portları bilinçli olarak kullanılmadı; böylece proje makinedeki
diğer servislerle birlikte çalışabiliyor. Değiştirmek isterseniz
`backend/.env`, `web/.env.local` ve `web/package.json` dosyalarına bakın.

### Doğrulama

```bash
curl http://localhost:8090/health
# {"database":"ok","status":"ok"}
```

## Proje Yapısı

```
beauty-ingredient-mvp/
├── docker-compose.yml       PostgreSQL 16 + otomatik yüklenen şema ve veri
├── backend/
│   ├── main.go              Yapılandırma, DB havuzu, router, düzgün kapanış
│   ├── api/                 İşleyiciler: ürün, içerik, profil, muadil
│   ├── models/              Paylaşılan alan tipleri
│   ├── middleware/          CORS
│   ├── cmd/migrate/         Göç komutu
│   └── db/
│       ├── migrations/      Sıralı şema göçleri
│       └── seed.sql         37 içerik, 14 örnek ürün
├── web/
│   ├── app/                 App Router sayfaları
│   │   ├── page.tsx         Açılış
│   │   ├── upload/          Fotoğraf yükleme + elle ürün eşleştirme
│   │   ├── ingredients/     Filtreli kaşif ve detay sayfaları
│   │   ├── products/        Katalog ve muadilli detay
│   │   ├── profile/         Hesap, cilt profili ve rıza
│   │   └── api/             Yükleme akışını besleyen rota
│   ├── components/          Navigation, IngredientCard
│   ├── public/              logo.svg, hero-lilies.jpg
│   └── lib/api.ts           Tipli backend istemcisi
└── docs/
    ├── SETUP.md
    └── API_SPEC.md
```

## Özellikler (Faz 1 MVP)

- ✅ Mevzuata bağlı risk seviyeleri (AB Tüzüğü 1223/2009 Ekleri)
- ✅ Cilt tipi, kaçınılacak alerjenler ve en yüksek risk seviyesine göre filtre
- ✅ INCI sırasına göre tam içerik dökümlü ürün kataloğu
- ✅ Kayıtlı profile karşı alerjen kontrolü
- ✅ İçerik kümeleri üzerinde Jaccard benzerliğiyle muadil tespiti
- ✅ Elle ürün eşleştirmeli fotoğraf yükleme
- ✅ Mobil uyumlu arayüz

Fotoğraftan otomatik ürün tanıma Faz 2'de —
[docs/DEVELOPMENT_PLAN.md](docs/DEVELOPMENT_PLAN.md) dosyasına bakın.

## Dil ve Yerelleştirme

Arayüz, dokümanlar ve kod yorumları Türkçedir. Bilinçli olarak İngilizce
bırakılanlar:

- **JSON alan adları ve DB kolonları** — `docs/API_SPEC.md`'de belgelenen
  sözleşmenin parçası.
- **`skin_type` değerleri** (`oily`, `dry`, `combination`, `sensitive`,
  `normal`) — API sözleşmesi; arayüz `skinTypeLabel()` ile Türkçe gösterir.
- **INCI adları** — uluslararası standart. Her içeriğin ayrıca Türkçe yaygın
  adı vardır (`name`), INCI adı `inci_name` alanında durur.

Fiyatlar `tr-TR` yerel ayarıyla biçimlendirilir. Örnek veri USD tutuyor;
gerçek katalog geldiğinde para birimi de gözden geçirilmeli.

## Marka

Renk paleti ve ikon proje sahibi tarafından sağlandı. Risk ölçeği paletin
içinden sıralı seçildi: **adaçayı** (düşük) → **terrakota** (orta) →
**şarap** (yüksek). Böylece hem markayla uyumlu hem de sıcaklık arttıkça uyarı
şiddeti artıyor. Renkler `web/tailwind.config.ts` içinde tanımlı.

## Örnek Veri

`backend/db/seed.sql` **gerçek içerikler** (doğru INCI adları) ama
**kurgusal ürünler** barındırır — uydurma marka adları
altında, böylece gerçekten satın alabileceğiniz hiçbir ürün yanlış tanıtılmaz.
Bunu küratörlüğü yapılmış gerçek ürün verisiyle değiştirmek Faz 1C'nin ilk işi.

## Kullanışlı Komutlar

```bash
make db-shell     # konteyner içinde psql
make db-reset     # volume'ü sil, şema + veriyi yeniden yükle
make lint         # go vet + next lint
make build        # API ikilisi + üretim Next.js paketi
```

## Dokümantasyon
- [Geliştirme Planı](docs/DEVELOPMENT_PLAN.md) — yol haritası ve öncelikler
- [Kurulum Rehberi](docs/SETUP.md) — ayrıntılı kurulum notları
- [API Spesifikasyonu](docs/API_SPEC.md) — uç nokta dokümantasyonu

## Teknoloji
- **Backend**: Go 1.21 + Gin + PostgreSQL 16 (`lib/pq`)
- **Ön yüz**: Next.js 14 (App Router) + React 18 + Tailwind CSS 3
- **Faz 2**: Vision API (ürün tanıma), Kafka (fiyat takibi)

---

Hassas ciltler için özenle hazırlandı. 🌸

*Risk seviyeleri yol göstericidir, tıbbi tavsiye değildir. Yeni ürünleri küçük
bir alanda deneyin; geçmeyen reaksiyonlar için dermatoloğa danışın.*
