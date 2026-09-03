# Clarity — Geliştirme Planı

3 Eylül 2026 itibarıyla durum. Faz 1 (MVP) yerelde uçtan uca çalışıyor.

---

## 1. Proje şu anda nerede

Yerel yığına karşı doğrulandı, varsayılmadı:

| Alan | Durum |
|------|-------|
| Veritabanı | Docker'da PostgreSQL 16; şema ve örnek veri ilk açılışta otomatik yükleniyor |
| Örnek veri | 37 içerik, 26 alerjen bağı, 40 fayda, 14 ürün, 113 ürün-içerik satırı |
| Backend | 15 rotanın tamamı gerçek SQL üzerinde; `go build` ve `go vet` temiz |
| Ön yüz | 11 rota; `next build`, `tsc --noEmit` ve `next lint` temiz |
| Yükleme akışı | Fotoğraf + elle ürün seçimi → içerik dökümü ve alerjen işaretleri |
| Muadil motoru | İçerik kümeleri üzerinde Jaccard benzerliği; %83'lük eşleşmeyi doğru buluyor |

Buraya gelmek için değişenler: `main.go` içindeki taslak işleyicilerin yerini
gerçek sorgular aldı ve kod `api/`, `models/`, `middleware/` olarak bölündü;
5 nanosaniyelik veritabanı ping zaman aşımı düzeltildi; eksik olan `/`,
`/ingredients`, `/products`, `/about` ve `/profile` sayfaları yazıldı; yükleme
sayfasının çağırdığı ama hiç var olmayan `/api/analyze-makeup` uygulandı.
Ardından arayüz, dokümanlar, kod yorumları ve örnek veri Türkçeleştirildi ve
proje sahibinin verdiği palet ile ikon üzerine bir marka kimliği kuruldu.

---

## 2. Bilinen eksikler ve riskler

Efora göre değil, ne kadar önemli olduğuna göre sıralandı.

### P0 — doğruluk ve güvenlik

1. **Alerjen eşleştirmesi yanlış pozitif üretiyor.**
   `api/ingredients.go` alt dizeleri iki yönlü eşleştiriyor. Bu yüzden
   `alkol` yazan bir kullanıcı **Lanolin** için uyarı alıyor: lanolinin
   alerjeni "yün alkolü" olarak kayıtlı. Yün alkolleri sterol yapıdadır,
   etanolle ilgisi yoktur. İnsanların reaksiyondan kaçınmak için güvendiği bir
   özellikte, yanlış bir uyarı kaçırılan bir uyarı kadar hızlı güven kaybettirir.

   > Not: Aynı hatanın İngilizce sürümdeki örneği (`nut` → **Coconut Oil**)
   > Türkçe sözlükte kayboldu, çünkü "kuruyemiş" ile "hindistan cevizi" ortak
   > alt dize taşımıyor. Yani sözlük değişimi o tek örneği tesadüfen düzeltti;
   > altta yatan hata duruyor.

   *Çözüm:* Alt dize eşleştirmesini bırakıp küratörlü bir alerjen sözlüğü ve
   açık bir eşanlam/üst-küme tablosu (`allergen_aliases`) ile tam eşleşmeye
   geçmek.

2. **Hiçbir yerde test yok.** `go test ./...` hiç test dosyası bulmuyor, ön
   yüzde de yok. Yukarıdaki her davranış elle, birer kez doğrulandı.
   *Çözüm:* Tek kullanımlık bir veritabanına karşı işleyici testleri ve
   asıl mantığın yaşadığı yer olan alerjen/muadil SQL'i için tablo temelli
   testler.

3. **Herhangi bir istemci herhangi bir profili okuyabilir veya değiştirebilir.**
   `GET /profiles/:id` ve `PUT /profiles/:id` çıplak bir tam sayı alıyor; web
   uygulaması bu kimliği `localStorage`'da tutuyor. Sayıyı artıran biri
   başkasının alerjen listesini görebilir — bu sağlıkla ilişkili kişisel veri.
   *Çözüm:* Herkese açık dağıtımdan önce gerçek oturumlar. Pazarlık konusu değil.

### P1 — veri ve ölçekte doğruluk

4. **Örnek ürünler kurgusal.** Geliştirme için faydalı, ürün olarak
   kullanışsız. Gerçek katalog Faz 1C'nin işi.
5. **Arama ölçeklenmez.** `ILIKE '%terim%'` indeks kullanamaz; API
   spesifikasyonundaki 100 ms hedefi, katalog ilginçleşmeden çok önce kırılır.
   *Çözüm:* GIN indeksli bir `tsvector` kolonu ya da `pg_trgm`.
6. **`UNIQUE(gtin, brand)` tekrarları engellemiyor.** PostgreSQL'de NULL'lar
   birbirinden farklı sayılır; GTIN'i boş olan istediğiniz kadar satır bir arada
   durabilir. *Çözüm:* GTIN'siz satırlar için `(brand, lower(name))` üzerinde
   kısmi tekil indeks.
7. **Migration aracı yok.** `schema.sql` yalnızca Docker volume'ü ilk
   oluşturulduğunda çalışıyor; şema değişikliği `make db-reset` ve veri kaybı
   demek. *Çözüm:* Şema oturmadan önce `golang-migrate` veya `goose`.

### P2 — yarım kalan yüzey

8. `price_history` ve `product_reviews` tabloları var ama hiçbir şey okumuyor
   veya yazmıyor.
9. Yüklenen görseller doğrulanıp atılıyor — hiçbir yerde saklanmıyor.
10. API `limit`/`offset` destekliyor ama arayüzde sayfalama yok.
11. API spesifikasyonunda söz verilen hız sınırlama ve API anahtarı yok.
12. Backend için `Dockerfile`, dağıtım yapılandırması ve CI yok.

### Türkçeye özgü notlar

- Go'nun `strings.ToLower`'ı Türkçe farkında değildir, ancak PostgreSQL'in
  `LOWER`'ı da aynı şekilde davrandığı için ikisi tutarlı çalışıyor. Noktalı
  **İ** ve noktasız **I** ile yapılan aramalar test edildi, ikisi de doğru
  eşleşiyor. Yine de veritabanına Türkçe bir collation gelirse bu varsayım
  bozulur — o noktada eşleştirmeyi tek bir yerde normalize etmek gerekir.
- Fiyatlar `tr-TR` biçimiyle gösteriliyor ama örnek veri USD tutuyor. Gerçek
  katalogla birlikte para birimi de kararlaştırılmalı.

---

## 3. Faz 1'in tamamlanması (önümüzdeki 2–3 hafta)

**1. hafta — güvenilir hale getir**
- [ ] Alt dize alerjen eşleştirmesini eşanlam tablosu ve tam eşleşmeyle değiştir
- [ ] Backend test paketi: işleyiciler, alerjen eşleştirme, muadil puanlama
- [ ] `golang-migrate` ekle; `schema.sql`'i ilk migration'a dönüştür
- [ ] GitHub Actions: `go vet`, `go test`, `next build`, `next lint`

**2. hafta — veriyi gerçek yap**
- [ ] Kaynağı belli INCI adları ve risk seviyeleriyle 200 içerik küratörlüğü
- [ ] Doğrulanmış içerik listeleriyle 50–100 gerçek ürün
- [ ] Ürün başına köken bilgisi: `source_url`, `verified_at`
- [ ] Ürün ve içerik adları üzerinde tam metin arama indeksi

**3. hafta — kullanılabilir yap**
- [ ] Oturumlar ve gerçek profil sahipliği
- [ ] İçerik ve ürün listelerinde sayfalama ve yükleme durumları
- [ ] Ürün detayında kayıtlı profile karşı satır içi alerjen kontrolü
- [ ] Her sayfada boş, hatalı ve çevrimdışı durumlarının gözden geçirilmesi

**Faz 1 için bitti tanımı:** Tanımadığınız biri bir profil oluşturabiliyor,
gerçek bir ürünü arayabiliyor, hangi içeriklerine tepki verdiğini görebiliyor ve
o alerjenleri içermeyen daha ucuz bir alternatif alabiliyor — ve bu yolun
tamamı testlerle kaplı.

---

## 4. Faz 2 (4–8. haftalar)

**Görüntüyle ürün tanıma.** Yükleme sayfası ve `/api/analyze-makeup` buna göre
şekillendirildi: rota bugün zaten bir görsel alıyor, sadece hangi ürün olduğunu
kullanıcıya soruyor. Elle verilen `product_id`'yi bir model çağrısıyla
değiştirmek sınırlı bir değişiklik.
- [ ] Önce içerik panelini OCR ile oku — ambalaj tanımaktan hem daha
      ulaşılabilir hem de katalogda olmayan ürünlerde de çalışıyor
- [ ] Güven düşük olduğunda elle seçime geri düş
- [ ] Değerlendirme kümesi oluşturmak için yüklenen görseli ve modelin
      cevabını sakla

**Fiyat takibi.** `price_history` zaten modellenmiş durumda.
- [ ] Kafka'dan önce tek bir kazıyıcıyı uçtan uca çalıştır
- [ ] Birden fazla kaynak olunca dağıtım için Kafka
- [ ] Favorilenen ürünlerde fiyat düşüşü bildirimleri

**Daha iyi öneriler.** Jaccard benzerliği hâlihazırda makul muadiller üreten
sağlam bir temel.
- [ ] İçerik sırasına göre ağırlıklandır — bir formülü ilk beş içerik belirler,
      mevcut puanlama ise her içeriği eşit sayıyor
- [ ] Her yerde bulunan içerikleri (su, gliserin) IDF benzeri bir yaklaşımla
      geri plana at
- [ ] Ancak ondan sonra, kullanıcı gerektiren ortak filtrelemeyi düşün

---

## 5. Faz 3 — topluluk ve ölçek

- Kullanıcı yorumları ve "bu bende alerji yaptı" bildirimleri, moderasyonla
- Türkiye pazarı kataloğu (yerelleştirme büyük ölçüde tamamlandı)
- Anahtar ve hız sınırlarıyla herkese açık API
- Mobil uygulama ya da aradaki farkı daha ucuza kapatıyorsa bir PWA

---

## 6. Enine kesen konular

**Test.** Kapsamı riskin olduğu yere yoğunlaştırın: `api/` içindeki SQL ve
alerjen mantığı. Arayüz anlık görüntü testlerinin değeri burada düşük; yükleme
ve profil oluşturma üzerinden geçen birkaç Playwright senaryosunun değeri yüksek.

**Dağıtım.** Backend Render veya Fly.io'ya, ön yüz Vercel'e, veritabanı yönetilen
bir PostgreSQL'e. Backend için bir `Dockerfile` ve hazırlık sondası ekleyin —
`/health` zaten veritabanı canlılığını bildiriyor.

**Veri kökeni.** Kullanıcıya gösterilen her içerik iddiası bir kaynağa kadar
izlenebilmeli. Bu, etik olduğu kadar hukuki bir mesele de: uygulama insanların
cildine sürdüğü ürünler hakkında güvenlik ifadeleri kuruyor. "Tıbbi tavsiye
değildir" uyarısını içerik puanlayan her yüzeyde tutun.

---

## 7. Önerilen ilk üç iş

1. Alerjen yanlış pozitiflerini düzelt (P0-1) — kullanıcıyı kendi güvenliği
   konusunda aktif olarak yanıltabilecek tek hata bu.
2. Backend test paketini ekle (P0-2) — ondan sonraki her şeyi değiştirmek
   güvenli hale gelir.
3. Oturumları ekle (P0-3) — herkese açık bir yere dağıtmadan önce zorunlu.
