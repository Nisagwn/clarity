// import-obf komutu, Open Beauty Facts kataloğunu içe aktarır.
//
// Open Beauty Facts, kozmetik ürünlerin topluluk tarafından derlenen açık
// veritabanıdır: barkod, marka, kategori ve etiketten okunmuş INCI listeleri.
// Anahtar gerektirmez.
//
// Kullanım:
//
//	# Toplu içe aktarım (önerilen) — dump indirilip verilir:
//	#   https://static.openbeautyfacts.org/data/openbeautyfacts-products.jsonl.gz
//	go run ./scripts/import-obf -file openbeautyfacts-products.jsonl.gz -limit 5000
//
//	# Tek kategori, API üzerinden:
//	go run ./scripts/import-obf -category lipsticks -limit 200
//
//	go run ./scripts/import-obf -file ... -dry-run
//
// Sonrasında puanlar tazelenir (yeni aday içerikler geldi):
//
//	go run ./cmd/score
//
// LİSANS — opsiyonel değil:
// Veri ODbL-1.0 altındadır; atıf ve aynı lisansla paylaşma gerektirir. Ürün
// görselleri CC-BY-SA'dır. Bu yüzden her ürün kaynağını (source, source_id,
// license, source_url) taşır ve arayüz her ürün sayfasında atfı gösterir.
// Atıfsız gösterim lisans ihlalidir; kaynağı kaydetmeyen bir içe aktarım da
// atfı imkânsız kılar.
package main

import (
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/lib/pq"

	"github.com/nisa/beauty-ingredient/inci"
)

const (
	sourceName    = "openbeautyfacts"
	licenseName   = "ODbL-1.0"
	userAgent     = "Clarity/0.1 (github.com/Nisagwn/clarity)"
	apiSearchURL  = "https://world.openbeautyfacts.org/api/v2/search"
	apiFields     = "code,product_name,product_name_en,brands,quantity,lang,categories_tags,ingredients_text,ingredients_text_en,ingredients,image_url,last_modified_t"
	apiPageSize   = 100
	apiPauseEvery = 6 * time.Second // OBF arama uç noktası için nazik hız sınırı
	batchSize     = 500             // bu kadar üründe bir işlem kapatılır
)

type options struct {
	file     string
	category string
	limit    int
	dryRun   bool
}

func main() {
	log.SetFlags(0)

	var opt options
	flag.StringVar(&opt.file, "file", "", "JSONL dump dosyası (.jsonl veya .jsonl.gz)")
	flag.StringVar(&opt.category, "category", "", "API modu: OBF kategori etiketi (örn. lipsticks)")
	flag.IntVar(&opt.limit, "limit", 1000, "en fazla kaç ürün ALINACAK (okunan satır değil; 0 = sınırsız)")
	flag.BoolVar(&opt.dryRun, "dry-run", false, "veritabanına yazma, yalnızca özeti göster")
	flag.Parse()

	if opt.file == "" && opt.category == "" {
		flag.Usage()
		log.Fatal("\n-file veya -category gerekli")
	}

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf(".env yüklenmedi (%v), ortam değişkenlerine düşülüyor", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL tanımlı değil")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("veritabanına bağlanılamadı: %v", err)
	}
	defer db.Close()

	if err := run(context.Background(), db, opt); err != nil {
		log.Fatalf("içe aktarım başarısız: %v", err)
	}
}

// summary, içe aktarımın sonucu. Atlanan kayıtlar nedenleriyle sayılır:
// "5.000 üründen 3.000'i alındı" cümlesinin geri kalanı olmadan anlamı yok.
type summary struct {
	read           int
	created        int
	updated        int
	incomplete     int
	skipped        map[string]int
	newIngredients []string
}

func run(ctx context.Context, db *sql.DB, opt options) error {
	st, err := newStore(ctx, db)
	if err != nil {
		return err
	}

	sum := summary{skipped: map[string]int{}}

	// Sınır, okunan satır değil ALINAN ürün sayısıdır: dump'ta her üç kayıttan
	// ikisinin içerik listesi yok, "1000 satır oku" demek işe yaramıyor.
	handle := func(raw obfProduct) error {
		sum.read++

		p, reason, ok := convert(raw)
		if !ok {
			sum.skipped[reason]++
			return nil
		}
		if p.Incomplete {
			sum.incomplete++
		}

		if opt.dryRun {
			sum.created++
			return limitReached(opt, sum)
		}

		created, err := st.save(ctx, p)
		if err != nil {
			// Tek bir bozuk kayıt tüm içe aktarımı düşürmemeli; ama sessizce
			// de geçilmez, nedeni özette görünür.
			sum.skipped[dbSkipReason(err)]++
			return nil
		}
		if created {
			sum.created++
		} else {
			sum.updated++
		}
		return limitReached(opt, sum)
	}

	if opt.file != "" {
		err = readFile(ctx, opt.file, handle)
	} else {
		err = readAPI(ctx, opt.category, handle)
	}
	if err != nil && !errors.Is(err, errLimitReached) {
		return err
	}

	if !opt.dryRun {
		if err := st.flush(ctx); err != nil {
			return err
		}
	}

	sum.newIngredients = st.added
	report(opt, sum)
	return catalogState(ctx, db)
}

// errLimitReached, -limit karşılandığında okumayı durdurur.
var errLimitReached = errors.New("istenen ürün sayısına ulaşıldı")

func limitReached(opt options, sum summary) error {
	if opt.limit > 0 && sum.created+sum.updated >= opt.limit {
		return errLimitReached
	}
	return nil
}

// ===== Okuma =====

// readFile, JSONL dump'ı satır satır akıtır. Dosya 100 MB'ın üstünde;
// tamamı belleğe alınmaz.
func readFile(ctx context.Context, path string, handle func(obfProduct) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var reader io.Reader = f
	if strings.HasSuffix(strings.ToLower(path), ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("%s açılamadı: %w", path, err)
		}
		defer gz.Close()
		reader = gz
	}

	dec := json.NewDecoder(reader)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		var p obfProduct
		err := dec.Decode(&p)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			// Tek bir bozuk satır dump'ın tamamını çöpe atmamalı.
			var syntaxErr *json.SyntaxError
			var typeErr *json.UnmarshalTypeError
			if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) {
				continue
			}
			return err
		}

		if err := handle(p); err != nil {
			return err
		}
	}
}

// readAPI, kategori araması yapar. OBF'nin arama uç noktası hız sınırlı;
// sayfalar arasında bekliyoruz.
func readAPI(ctx context.Context, category string, handle func(obfProduct) error) error {
	client := &http.Client{Timeout: 60 * time.Second}
	read := 0

	for page := 1; ; page++ {
		q := url.Values{}
		q.Set("categories_tags_en", category)
		q.Set("fields", apiFields)
		q.Set("page", fmt.Sprint(page))
		q.Set("page_size", fmt.Sprint(apiPageSize))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiSearchURL+"?"+q.Encode(), nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", userAgent)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		var body struct {
			Count    int          `json:"count"`
			Products []obfProduct `json:"products"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("API yanıtı çözülemedi (sayfa %d): %w", page, err)
		}
		if len(body.Products) == 0 {
			return nil
		}

		for _, p := range body.Products {
			if err := handle(p); err != nil {
				return err
			}
			read++
		}

		log.Printf("  sayfa %d alındı (%d/%d kayıt okundu)", page, read, body.Count)
		time.Sleep(apiPauseEvery)
	}
}

// ===== Yazma =====

// store, içe aktarımı toplu işlemlerle veritabanına yazar.
type store struct {
	db *sql.DB
	tx *sql.Tx

	// ingredients, normalleştirilmiş INCI adından katalog id'sine eşler.
	// Katalog başta bir kez okunur; içe aktarım sırasında eklenen adaylar da
	// buraya girer, böylece aynı madde iki kez eklenmez.
	ingredients map[string]int
	added       []string
	pending     int
}

func newStore(ctx context.Context, db *sql.DB) (*store, error) {
	const query = `SELECT id, name, COALESCE(inci_name, '') FROM ingredients`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cache := map[string]int{}
	for rows.Next() {
		var (
			id             int
			name, inciName string
		)
		if err := rows.Scan(&id, &name, &inciName); err != nil {
			return nil, err
		}
		for _, key := range []string{inci.Normalize(inciName), inci.Normalize(name)} {
			if key != "" {
				if _, exists := cache[key]; !exists {
					cache[key] = id
				}
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &store{db: db, ingredients: cache}, nil
}

// save, ürünü ve içerik bağlarını yazar; ürün yeni açıldıysa true döner.
func (s *store) save(ctx context.Context, p product) (bool, error) {
	if err := s.begin(ctx); err != nil {
		return false, err
	}

	quality := "ok"
	if p.Incomplete {
		quality = "incomplete"
	}

	const upsert = `
		INSERT INTO products
		    (name, brand, gtin, category, description, image_url, source_url,
		     source, source_id, license, data_quality, verified_at, updated_at)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7,
		        $8, $9, $10, $11, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (source, source_id) WHERE source IS NOT NULL AND source_id IS NOT NULL
		DO UPDATE SET
		    name         = EXCLUDED.name,
		    brand        = EXCLUDED.brand,
		    category     = EXCLUDED.category,
		    description  = EXCLUDED.description,
		    image_url    = EXCLUDED.image_url,
		    source_url   = EXCLUDED.source_url,
		    license      = EXCLUDED.license,
		    data_quality = EXCLUDED.data_quality,
		    verified_at  = CURRENT_TIMESTAMP,
		    updated_at   = CURRENT_TIMESTAMP
		RETURNING id, (xmax = 0) AS created`

	var (
		productID int
		created   bool
	)
	err := s.tx.QueryRowContext(ctx, upsert,
		p.Name, p.Brand, p.GTIN, p.Category, p.Description, p.ImageURL, p.SourceURL,
		sourceName, p.SourceID, licenseName, quality,
	).Scan(&productID, &created)
	if err != nil {
		return false, s.abort(err)
	}

	if err := s.writeIngredients(ctx, productID, p.Ingredients); err != nil {
		return false, s.abort(err)
	}

	s.pending++
	if s.pending >= batchSize {
		if err := s.flush(ctx); err != nil {
			return false, err
		}
	}
	return created, nil
}

// writeIngredients, ürünün INCI listesini sırasıyla yazar. Önce eski bağlar
// silinir: kaynaktaki liste değiştiyse eskisi kalmamalı.
func (s *store) writeIngredients(ctx context.Context, productID int, names []string) error {
	if _, err := s.tx.ExecContext(ctx,
		`DELETE FROM product_ingredients WHERE product_id = $1`, productID); err != nil {
		return err
	}

	const link = `
		INSERT INTO product_ingredients (product_id, ingredient_id, order_index)
		VALUES ($1, $2, $3)
		ON CONFLICT (product_id, ingredient_id) DO UPDATE SET order_index = EXCLUDED.order_index`

	for i, name := range names {
		id, err := s.ingredientID(ctx, name)
		if err != nil {
			return err
		}
		if _, err := s.tx.ExecContext(ctx, link, productID, id, i+1); err != nil {
			return err
		}
	}
	return nil
}

// ingredientID, INCI adını katalog kimliğine çözer; yoksa ADAY olarak ekler.
//
// Aday içerikler küratörlükten geçmemiştir: Türkçe adı, açıklaması ve cilt
// tipi etiketi yoktur, mevzuat eşleşmesi bulunana kadar da puansız kalırlar.
// Bunları eklemek yerine atlamak, ürünün içerik listesini eksik göstermek
// olurdu — muadil hesabı da yanlış çıkardı.
func (s *store) ingredientID(ctx context.Context, name string) (int, error) {
	key := inci.Normalize(name)
	if id, ok := s.ingredients[key]; ok {
		return id, nil
	}

	const insert = `
		INSERT INTO ingredients (name, inci_name, source)
		VALUES ($1, $1, $2)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`

	var id int
	if err := s.tx.QueryRowContext(ctx, insert, trimTo(name, 255), sourceName).Scan(&id); err != nil {
		return 0, err
	}

	s.ingredients[key] = id
	s.added = append(s.added, name)
	return id, nil
}

func (s *store) begin(ctx context.Context) error {
	if s.tx != nil {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	s.tx = tx
	return nil
}

// flush, açık işlemi kapatır.
func (s *store) flush(ctx context.Context) error {
	if s.tx == nil {
		return nil
	}
	err := s.tx.Commit()
	s.tx = nil
	s.pending = 0
	return err
}

// abort, işlemi geri alır ve hatayı olduğu gibi döndürür. Postgres bir hatadan
// sonra işlemin geri kalanını reddeder; devam edebilmek için yeni bir işlem
// gerekiyor.
func (s *store) abort(cause error) error {
	if s.tx != nil {
		s.tx.Rollback()
		s.tx = nil
		s.pending = 0
	}
	return cause
}

// dbSkipReason, veritabanı hatasını özette görünecek bir sebebe çevirir.
func dbSkipReason(err error) string {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return "aynı barkod ve marka katalogda zaten var"
	}
	return "veritabanı hatası: " + err.Error()
}

// ===== Raporlama =====

func report(opt options, sum summary) {
	mode := "dump: " + opt.file
	if opt.file == "" {
		mode = "API kategorisi: " + opt.category
	}
	if opt.dryRun {
		mode += " (deneme — hiçbir şey yazılmadı)"
	}

	log.Printf("%s", mode)
	log.Printf("  okunan kayıt      : %d", sum.read)
	log.Printf("  yeni ürün         : %d", sum.created)
	log.Printf("  güncellenen ürün  : %d", sum.updated)
	log.Printf("  eksik veri etiketi: %d (üçten az içerik — muadil hesabına girmez)", sum.incomplete)

	if len(sum.skipped) > 0 {
		reasons := make([]string, 0, len(sum.skipped))
		for reason := range sum.skipped {
			reasons = append(reasons, reason)
		}
		sort.Slice(reasons, func(i, j int) bool { return sum.skipped[reasons[i]] > sum.skipped[reasons[j]] })

		total := 0
		for _, n := range sum.skipped {
			total += n
		}
		log.Printf("  alınmayan kayıt   : %d", total)
		for _, reason := range reasons {
			log.Printf("      %-42s %d", trimTo(reason, 42), sum.skipped[reason])
		}
	}

	if n := len(sum.newIngredients); n > 0 {
		log.Printf("  aday içerik eklendi: %d (küratörlük ve mevzuat eşleşmesi bekliyor)", n)
		log.Printf("      %s", strings.Join(preview(sum.newIngredients, 8), ", "))
	}

	log.Println("  lisans: ODbL-1.0 — atıf her ürün sayfasında gösterilmeli")
}

// catalogState, içe aktarımdan sonra kataloğun durumunu yazar: asıl soru
// "kaç ürün var" değil, "kaçının verisi kullanılabilir".
func catalogState(ctx context.Context, db *sql.DB) error {
	const query = `
		SELECT (SELECT COUNT(*) FROM products),
		       (SELECT COUNT(*) FROM products WHERE data_quality = 'incomplete'),
		       (SELECT COUNT(*) FROM ingredients),
		       (SELECT COUNT(*) FROM ingredients WHERE score_version IS NOT NULL)`

	var products, incomplete, ingredients, scored int
	if err := db.QueryRowContext(ctx, query).Scan(&products, &incomplete, &ingredients, &scored); err != nil {
		return err
	}

	log.Println("katalog durumu")
	log.Printf("  ürün        : %d (%d eksik veri)", products, incomplete)
	log.Printf("  içerik      : %d (%d tanesi mevzuata bağlı ve puanlı)", ingredients, scored)
	log.Println("  puanları tazelemek için: go run ./cmd/score")
	return nil
}

func preview(values []string, n int) []string {
	if len(values) <= n {
		return values
	}
	return append(append([]string{}, values[:n]...), fmt.Sprintf("... ve %d tane daha", len(values)-n))
}
