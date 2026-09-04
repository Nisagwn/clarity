# Clarity — Kurulum Rehberi

## Projeye Genel Bakış
- **Amaç**: Fotoğraf yükleme → içerik dökümü + alerjen eşleştirme + muadil önerisi
- **Süre**: 6 haftalık MVP
- **Teknoloji**: Go + Next.js + PostgreSQL
- **Mevcut faz**: Faz 1 (MVP — elle ürün eşleştirme)

Yol haritası ve öncelikler için [DEVELOPMENT_PLAN.md](DEVELOPMENT_PLAN.md).

---

## Adım 1: Veritabanı

Veritabanı Docker'da çalışır; yerel bir PostgreSQL kurmanıza gerek yoktur.

```bash
docker compose up -d
```

Bu komut PostgreSQL 16'yı başlatır. Şema **göçlerle** yüklenir:

```bash
make migrate          # bekleyen göçleri uygular
make seed             # örnek veriyi yükler (idempotent)
make migrate-version  # mevcut şema sürümü
```

`make up` zaten göçleri de çalıştırır.

```bash
# Bağlantıyı test et
docker compose exec postgres psql -U beauty -d beauty_ingredient \
  -c "SELECT COUNT(*) FROM ingredients;"

# psql kabuğu
make db-shell
```

Şema değiştirmek için artık veri silmeye gerek yok: yeni bir göç dosyası
ekleyip `make migrate` çalıştırın. `make db-reset` yalnızca tamamen temiz bir
başlangıç istediğinizde.

---

## Adım 2: Backend (Go)

```bash
cd backend
cp .env.example .env
go mod download
go run .          # http://localhost:8090
```

### Paket yapısı

```
backend/
├── main.go          Yapılandırma, DB havuzu, router, düzgün kapanış
├── api/             İstek işleyicileri
│   ├── server.go        Rota kaydı ve ortak yardımcılar
│   ├── products.go
│   ├── ingredients.go
│   ├── profiles.go
│   └── recommendations.go
├── models/          Paylaşılan alan tipleri
├── middleware/      CORS
├── cmd/migrate/     Göç komutu
└── db/
    ├── migrations/  Sıralı şema göçleri
    └── seed.sql     Örnek veri
```

### Ortam değişkenleri

```bash
DATABASE_URL=postgres://beauty:beauty@127.0.0.1:5433/beauty_ingredient?sslmode=disable
PORT=8090
GIN_MODE=debug
CORS_ALLOWED_ORIGINS=http://localhost:3001,http://localhost:3000
```

`.env` isteğe bağlıdır; üretimde ortam değişkenlerini platform belirler.

---

## Adım 3: Ön yüz (Next.js)

```bash
cd web
npm install
npm run dev       # http://localhost:3001
```

`web/.env.local`:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8090
```

### Sayfalar

| Rota | İşi |
|------|-----|
| `/` | Açılış, katalog önizlemesi |
| `/upload` | Fotoğraf yükleme + elle ürün eşleştirme |
| `/ingredients` | Filtreli içerik kaşifi |
| `/ingredients/[id]` | Tek içerik detayı |
| `/products` | Ürün kataloğu |
| `/products/[id]` | İçerikler ve muadillerle ürün detayı |
| `/profile` | Hesap, cilt profili ve rıza yönetimi |
| `/about` | Yöntem ve sorumluluk reddi |
| `/api/analyze-makeup` | Yükleme akışını besleyen rota |

---

## Dil kararları

Arayüz, dokümanlar, kod yorumları ve örnek veri Türkçedir. Bilinçli olarak
İngilizce kalanlar:

- **JSON alan adları ve DB kolonları** — `API_SPEC.md`'deki sözleşmenin parçası.
  Değiştirmek belgelenmiş API'yi kırardı.
- **`skin_type` değerleri** — `oily`, `dry`, `combination`, `sensitive`,
  `normal`. Arayüz `skinTypeLabel()` ile Türkçe gösterir; sunucu hata
  mesajlarında iki dili birlikte verir.
- **INCI adları** — uluslararası standart, çevrilmez. Her içeriğin ayrıca
  Türkçe yaygın adı (`name`) vardır.

---

## Adım 4: MVP özellik listesi

### Faz 1A: Çekirdek backend — ✅ tamamlandı
- [x] Veritabanı şeması
- [x] Kullanıcı profili uç noktaları (oluştur, getir, güncelle)
- [x] İçerik uç noktaları (liste, detay, filtreler)
- [x] Ürün uç noktaları (oluştur, liste, detay, arama, markaya göre)
- [x] Ürün-içerik ilişkilendirmesi
- [x] Alerjen kontrolü uç noktası

### Faz 1B: Ön yüz MVP — ✅ tamamlandı
- [x] Fotoğraf yükleme sayfası (sürükle-bırak)
- [x] Kullanıcı profili oluşturma
- [x] Kullanıcı alerjen girişi
- [x] İçerik kaşifi (arama + filtre)
- [x] Ürün detay sayfası
- [x] Alerjen eşleşme gösterimi

### Faz 1C: Ürün verisi — sırada
- [ ] En popüler 50–100 ürünü elle ekle
- [ ] İlk 200 içeriğin küratörlüğü
- [ ] İçerikleri ürünlerle eşle
- [ ] Her içerik için EWG güvenlik puanı
- [ ] Temel alerjen ilişkilendirmeleri

### Faz 1D: Test ve cila
- [ ] API entegrasyon testleri
- [ ] Mobilde arayüz testi
- [ ] Performans testi (görsel yükleme)
- [ ] Hata yönetimi gözden geçirmesi
- [ ] Oturumlar (herkese açık dağıtımdan önce zorunlu)

---

## Yerel çalıştırma

### 1. Terminal: veritabanı
```bash
docker compose up -d
```

### 2. Terminal: backend
```bash
make backend      # http://localhost:8090
```

### 3. Terminal: ön yüz
```bash
make frontend     # http://localhost:3001
```

---

## MVP kabul kriterleri

- ✅ Kullanıcı makyaj fotoğrafı yükleyebiliyor
- ✅ Sistem eşleşen ürünü gösteriyor (şimdilik elle)
- ✅ İçerik dökümü güvenlik puanlarıyla gösteriliyor
- ✅ Kullanıcının alerjenleri sonuçlarda işaretleniyor
- ✅ Alerjen kontrolü çalışıyor
- ✅ Mobil uyumlu tasarım
- ✅ Tüm uç noktalar belgelenmiş
- ⬜ Sorgular optimize edildi (<100 ms) — bkz. DEVELOPMENT_PLAN P1-5

---

## Sorun giderme

**`API'ye ulaşılamıyor` uyarısı görüyorum.**
Backend çalışmıyor. `make backend` ile başlatın ve
`curl http://localhost:8090/health` ile doğrulayın.

**Backend `veritabanı: ... connection refused` diyor.**
PostgreSQL ayakta değil. `docker compose up -d` çalıştırın; `docker ps` ile
`beauty-postgres` konteynerinin `healthy` olduğunu görün.

**Konteyner `healthy` ama bağlantı yine de kopuyor (`wsarecv: An existing
connection was forcibly closed`).**
Windows'ta `localhost` IPv6 `::1` adresine çözülüyor, Docker'ın port eşlemesi
ise IPv4'te dinliyor. Bu yüzden DSN'de **`127.0.0.1` kullanılıyor, `localhost`
değil**. `.env` dosyanızda `@localhost:5433` yazıyorsa `@127.0.0.1:5433` yapın.

Belirtisi kafa karıştırıcı: backend başlarken ölür ama port bir süre
"LISTENING" görünmeye devam eder, yani site açılmıyor ama port doluymuş gibi
durur. `netstat -ano | grep 8090` ile kalan ölü dinleyiciyi bulup
sonlandırmanız gerekebilir.

**Port zaten kullanımda.**
3001 ve 8090 seçildi çünkü 3000 ve 8080 sık sık başka servisler tarafından
tutuluyor. Değiştirmek için `backend/.env`, `web/.env.local` ve
`web/package.json` dosyalarını güncelleyin.

**Yeni göç yazdım ama uygulanmadı.**
`make migrate` çalıştırın. Göçler sunucu başlangıcında otomatik çalışmaz:
birden fazla örnek aynı anda açıldığında yarış oluşurdu.
