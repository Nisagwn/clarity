package api

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"testing"

	"github.com/nisa/beauty-ingredient/models"
)

// Bu dosya Faz 4'ün sözlerini koruyor: gerçek katalog kaynağını ve lisansını
// taşır, eksik veri muadil hesabına girmez, fiyat bilinmiyorsa uydurulmaz.

type productDetailResponse struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Price       *float64 `json:"price"`
	Source      string   `json:"source"`
	SourceID    string   `json:"source_id"`
	License     string   `json:"license"`
	SourceURL   string   `json:"source_url"`
	DataQuality string   `json:"data_quality"`
}

// addImportedProduct, içe aktarımdan gelmiş gibi bir ürün ekler.
func addImportedProduct(t *testing.T, db *sql.DB, name, brand, sourceID, quality string) int {
	t.Helper()

	const q = `
		INSERT INTO products (name, brand, category, currency, source, source_id,
		                      license, source_url, data_quality, verified_at)
		VALUES ($1, $2, 'ruj', 'USD', 'openbeautyfacts', $3, 'ODbL-1.0',
		        $5, $4, CURRENT_TIMESTAMP)
		RETURNING id`

	var id int
	sourceURL := "https://world.openbeautyfacts.org/product/" + sourceID
	if err := db.QueryRow(q, name, brand, sourceID, quality, sourceURL).Scan(&id); err != nil {
		t.Fatalf("ürün eklenemedi (%s): %v", name, err)
	}
	return id
}

// Lisans yükümlülüğü: ODbL atıf gerektiriyor. Atıf ancak kaynağı API
// döndürürse gösterilebilir.
func TestProductCarriesSourceAndLicense(t *testing.T) {
	_, r, db := newTestServer(t)

	id := addImportedProduct(t, db, "Süper Mat Ruj", "Maybelline", "0018787788059", "ok")

	var resp productDetailResponse
	code := doJSON(t, r, http.MethodGet, "/products/"+strconv.Itoa(id), nil, &resp)
	mustStatus(t, code, statusOK, "ürün detayı")

	if resp.Source != "openbeautyfacts" || resp.License != "ODbL-1.0" {
		t.Errorf("kaynak/lisans eksik döndü: %+v", resp)
	}
	if resp.SourceID != "0018787788059" {
		t.Errorf("kaynak kimliği %q", resp.SourceID)
	}
	if resp.SourceURL == "" {
		t.Error("kaynak adresi boş; atıf bağlantı veremez")
	}
}

// Open Beauty Facts fiyat taşımıyor. 0 döndürmek ürünü bedava sanmaya davet
// ederdi; bilinmeyen fiyat null olmalı.
func TestUnknownPriceIsNull(t *testing.T) {
	_, r, db := newTestServer(t)

	id := addImportedProduct(t, db, "Fiyatsız Ruj", "Maybelline", "111", "ok")

	var resp productDetailResponse
	code := doJSON(t, r, http.MethodGet, "/products/"+strconv.Itoa(id), nil, &resp)
	mustStatus(t, code, statusOK, "ürün detayı")

	if resp.Price != nil {
		t.Errorf("fiyat %v döndü, null beklenirdi", *resp.Price)
	}
}

// Eksik içerik listeli ürün katalogda görünür ama muadil hesabına GİRMEZ:
// iki kısa listenin Jaccard benzerliği yüksek çıkar ve bu benzerlik değil,
// veri eksikliğidir.
func TestIncompleteProductsExcludedFromDupes(t *testing.T) {
	s, r, db := newTestServer(t)

	su := addIngredient(t, db, "Su", "Aqua", 0)
	gliserin := addIngredient(t, db, "Gliserin", "Glycerin", 0)
	skualan := addIngredient(t, db, "Skualan", "Squalane", 0)
	pantenol := addIngredient(t, db, "Pantenol", "Panthenol", 0)

	target := addImportedProduct(t, db, "Hedef Ruj", "Marka A", "t-1", "ok")
	linkIngredients(t, db, target, su, gliserin, skualan, pantenol)

	complete := addImportedProduct(t, db, "Tam Listeli Ruj", "Marka B", "c-1", "ok")
	linkIngredients(t, db, complete, su, gliserin, skualan)

	// Aynı iki içerikle "kusursuz" eşleşen ama listesi eksik ürün.
	thin := addImportedProduct(t, db, "Eksik Listeli Ruj", "Marka C", "i-1", "incomplete")
	linkIngredients(t, db, thin, su, gliserin)

	recs, err := s.recommend(context.Background(), target, 10, nil)
	if err != nil {
		t.Fatalf("muadil sorgusu başarısız: %v", err)
	}

	for _, rec := range recs {
		if rec.ID == thin {
			t.Fatalf("eksik veri etiketli ürün öneriye girdi: %+v", rec)
		}
	}
	if !containsID(recs, complete) {
		t.Error("tam listeli aday öneride yok")
	}

	// Katalogdan gizlenmiyor: eksik olan veri, ürünün kendisi değil.
	var list struct {
		Products []struct {
			ID          int    `json:"id"`
			DataQuality string `json:"data_quality"`
		} `json:"products"`
	}
	code := doJSON(t, r, http.MethodGet, "/products?limit=50", nil, &list)
	mustStatus(t, code, statusOK, "ürün listesi")

	found := false
	for _, p := range list.Products {
		if p.ID == thin {
			found = true
			if p.DataQuality != "incomplete" {
				t.Errorf("veri kalitesi %q, incomplete beklenirdi", p.DataQuality)
			}
		}
	}
	if !found {
		t.Error("eksik veri etiketli ürün listeden gizlendi; etiketlenip gösterilmeli")
	}
}

// Muadil sorgusundaki alerjen elemesi de kanonik sözlükten çözülmeli.
// Alt dize karşılaştırması "alkol" arayan kullanıcıya alerjeni "yün alkolü"
// olan Lanolin'i eledirirdi — Faz 1'de alerjen kontrolünde düzeltilen hata
// bu sorguda kalmıştı.
func TestDupeAllergenFilterUsesCanonicalVocabulary(t *testing.T) {
	s, _, db := newTestServer(t)

	su := addIngredient(t, db, "Su", "Aqua", 0)
	gliserin := addIngredient(t, db, "Gliserin", "Glycerin", 0)
	skualan := addIngredient(t, db, "Skualan", "Squalane", 0)

	lanolin := addIngredient(t, db, "Lanolin", "Lanolin", 0)
	addAllergen(t, db, lanolin, "yün alkolü", 6)

	denature := addIngredient(t, db, "Denatüre Alkol", "Alcohol Denat.", 0)
	addAllergen(t, db, denature, "alkol", 4)

	target := addImportedProduct(t, db, "Hedef", "Marka A", "t-2", "ok")
	linkIngredients(t, db, target, su, gliserin, skualan)

	withLanolin := addImportedProduct(t, db, "Lanolinli", "Marka B", "b-2", "ok")
	linkIngredients(t, db, withLanolin, su, gliserin, skualan, lanolin)

	withAlcohol := addImportedProduct(t, db, "Alkollü", "Marka C", "c-2", "ok")
	linkIngredients(t, db, withAlcohol, su, gliserin, skualan, denature)

	recs, err := s.recommend(context.Background(), target, 10, []string{"alkol"})
	if err != nil {
		t.Fatalf("muadil sorgusu başarısız: %v", err)
	}

	if containsID(recs, withAlcohol) {
		t.Error("alerjeni gerçekten taşıyan ürün elenmedi")
	}
	if !containsID(recs, withLanolin) {
		t.Error(`"alkol" araması "yün alkolü" taşıyan ürünü eledi — yanlış pozitif`)
	}
}

func containsID(recs []models.Recommendation, id int) bool {
	for _, r := range recs {
		if r.ID == id {
			return true
		}
	}
	return false
}

// Katalog binlerce ürüne çıktı; kategori süzgeci olmadan gezilemez.
// Sayılar kataloğun nerede kalın, nerede ince olduğunu söylüyor.
func TestListCategoriesCountsProducts(t *testing.T) {
	_, r, db := newTestServer(t)

	addImportedProduct(t, db, "Ruj 1", "Marka A", "k-1", "ok")
	addImportedProduct(t, db, "Ruj 2", "Marka B", "k-2", "ok")

	// addImportedProduct 'ruj' kategorisiyle ekliyor; ikinci bir kategori
	// ekleyip sıralamayı da doğruluyoruz.
	if _, err := db.Exec(
		`INSERT INTO products (name, brand, category, source, source_id)
		 VALUES ('Maskara', 'Marka C', 'maskara', 'openbeautyfacts', 'k-3')`); err != nil {
		t.Fatalf("ürün eklenemedi: %v", err)
	}

	var resp struct {
		Categories []struct {
			Name         string `json:"name"`
			ProductCount int    `json:"product_count"`
		} `json:"categories"`
	}
	code := doJSON(t, r, http.MethodGet, "/products/categories", nil, &resp)
	mustStatus(t, code, statusOK, "kategori listesi")

	if len(resp.Categories) != 2 {
		t.Fatalf("kategori sayısı %d, beklenen 2: %+v", len(resp.Categories), resp.Categories)
	}
	if resp.Categories[0].Name != "ruj" || resp.Categories[0].ProductCount != 2 {
		t.Errorf("en kalabalık kategori yanlış: %+v", resp.Categories[0])
	}
	if resp.Categories[1].Name != "maskara" || resp.Categories[1].ProductCount != 1 {
		t.Errorf("ikinci kategori yanlış: %+v", resp.Categories[1])
	}
}
