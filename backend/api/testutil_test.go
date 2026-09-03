package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Testler TEST_DATABASE_URL ile ayrı bir veritabanına bağlanır. Geliştirme
// veritabanına bağlanmak veri kaybı olurdu: her test tabloları temizliyor.
//
//	make test    (test veritabanını da ayağa kaldırır)
//
// Değişken tanımlı değilse testler atlanır; böylece veritabanı olmayan bir
// makinede `go test ./...` kırmızı olmaz, sessizce atlar.

var (
	migrateOnce sync.Once
	migrateErr  error
)

// testDB, göçleri bir kez uygulayıp temiz bir bağlantı döndürür.
func testDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL tanımlı değil — testler atlanıyor (bkz. make test)")
	}

	migrateOnce.Do(func() {
		m, err := migrate.New("file://../db/migrations", dsn)
		if err != nil {
			migrateErr = err
			return
		}
		defer m.Close()
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			migrateErr = err
		}
	})
	if migrateErr != nil {
		t.Fatalf("test veritabanı göçleri başarısız: %v", migrateErr)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("test veritabanına bağlanılamadı: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	truncateAll(t, db)
	return db
}

// truncateAll, her testin temiz bir tablodan başlamasını sağlar.
// RESTART IDENTITY, id'lerin testler arasında öngörülebilir kalması için.
func truncateAll(t *testing.T, db *sql.DB) {
	t.Helper()

	const q = `
		TRUNCATE ingredients, ingredient_allergens, ingredient_benefits,
		         ingredient_skin_types, products, product_ingredients,
		         user_profiles, user_allergens, user_favorites,
		         product_reviews, price_history
		RESTART IDENTITY CASCADE`

	if _, err := db.Exec(q); err != nil {
		t.Fatalf("tablolar temizlenemedi: %v", err)
	}
}

// newTestServer, gerçek rotalarla bir test sunucusu kurar.
func newTestServer(t *testing.T) (*Server, *gin.Engine, *sql.DB) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	db := testDB(t)
	s := New(db)
	r := gin.New()
	s.RegisterRoutes(r)
	return s, r, db
}

// ===== İstek yardımcıları =====

// doJSON, bir istek çalıştırıp yanıtı çözer. body nil ise gövde gönderilmez.
func doJSON(t *testing.T, r *gin.Engine, method, path string, body any, out any) int {
	t.Helper()

	var reader *strings.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("istek gövdesi kodlanamadı: %v", err)
		}
		reader = strings.NewReader(string(raw))
	} else {
		reader = strings.NewReader("")
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if out != nil && w.Body.Len() > 0 {
		if err := json.Unmarshal(w.Body.Bytes(), out); err != nil {
			t.Fatalf("yanıt çözülemedi (%d): %v — gövde: %s", w.Code, err, w.Body.String())
		}
	}
	return w.Code
}

// mustStatus, beklenen durum kodunu doğrular ve aksi halde gövdeyi yazar.
func mustStatus(t *testing.T, got, want int, ctx string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: durum kodu %d, beklenen %d", ctx, got, want)
	}
}

// ===== Fixture yardımcıları =====
//
// Testler seed.sql'e bağlı değildir: örnek veri bir gösterim aracı ve
// değişecek. Testler kendi asgari verisini kurar.

// addIngredient, bir içerik ekler ve id'sini döndürür.
func addIngredient(t *testing.T, db *sql.DB, name, inci string, concern int) int {
	t.Helper()

	var id int
	err := db.QueryRow(
		`INSERT INTO ingredients (name, inci_name, concern_level) VALUES ($1, $2, $3) RETURNING id`,
		name, inci, concern,
	).Scan(&id)
	if err != nil {
		t.Fatalf("içerik eklenemedi (%s): %v", name, err)
	}
	return id
}

// addAllergen, bir içeriğe alerjen bağlar.
func addAllergen(t *testing.T, db *sql.DB, ingredientID int, allergen string, severity int) {
	t.Helper()

	_, err := db.Exec(
		`INSERT INTO ingredient_allergens (ingredient_id, allergen_name, severity) VALUES ($1, $2, $3)`,
		ingredientID, allergen, severity,
	)
	if err != nil {
		t.Fatalf("alerjen eklenemedi (%s): %v", allergen, err)
	}
}

// addSkinType, bir içeriğe cilt tipi etiketi ekler.
func addSkinType(t *testing.T, db *sql.DB, ingredientID int, skinType string) {
	t.Helper()

	_, err := db.Exec(
		`INSERT INTO ingredient_skin_types (ingredient_id, skin_type) VALUES ($1, $2)`,
		ingredientID, skinType,
	)
	if err != nil {
		t.Fatalf("cilt tipi eklenemedi (%s): %v", skinType, err)
	}
}

// addProduct, bir ürün ekler ve id'sini döndürür.
func addProduct(t *testing.T, db *sql.DB, name, brand, category string, price float64) int {
	t.Helper()

	var id int
	err := db.QueryRow(
		`INSERT INTO products (name, brand, category, price, currency)
		 VALUES ($1, $2, $3, $4, 'USD') RETURNING id`,
		name, brand, category, price,
	).Scan(&id)
	if err != nil {
		t.Fatalf("ürün eklenemedi (%s): %v", name, err)
	}
	return id
}

// linkIngredients, içerikleri verilen sırayla bir ürüne bağlar.
func linkIngredients(t *testing.T, db *sql.DB, productID int, ingredientIDs ...int) {
	t.Helper()

	for i, ingID := range ingredientIDs {
		_, err := db.Exec(
			`INSERT INTO product_ingredients (product_id, ingredient_id, order_index)
			 VALUES ($1, $2, $3)`,
			productID, ingID, i+1,
		)
		if err != nil {
			t.Fatalf("ürün-içerik bağlanamadı: %v", err)
		}
	}
}

// statusOK, okunabilirlik için.
const statusOK = http.StatusOK
