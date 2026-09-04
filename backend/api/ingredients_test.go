package api

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

// allergenFixture, alerjen eşleştirmesinin zor vakalarını içeren bir kurulum.
//
// Kritik olan iki çift:
//   - Lanolin'in alerjeni "yün alkolü", Denatüre Alkol'ünki "alkol".
//     Alt dize eşleştirmesi "alkol" arayan kullanıcıyı Lanolin'e takıyordu.
//   - Shea Yağı "kuruyemiş", Hindistan Cevizi Yağı "hindistan cevizi" taşır.
//     Bunlar farklı alerjenlerdir; hindistan cevizi botanik olarak ağaç
//     kuruyemişi değildir.
type allergenFixture struct {
	rujID, fondotenID          int
	lanolinID, denatureAlkolID int
	sheaID, hindistanCeviziID  int
	parfumID, linaloolID       int
}

func setupAllergenFixture(t *testing.T, db *sql.DB) allergenFixture {
	t.Helper()

	f := allergenFixture{}

	f.lanolinID = addIngredient(t, db, "Lanolin", "Lanolin", 3)
	addAllergen(t, db, f.lanolinID, "lanolin", 6)
	addAllergen(t, db, f.lanolinID, "yün alkolü", 6)

	f.denatureAlkolID = addIngredient(t, db, "Denatüre Alkol", "Alcohol Denat.", 4)
	addAllergen(t, db, f.denatureAlkolID, "alkol", 4)

	f.sheaID = addIngredient(t, db, "Shea Yağı", "Butyrospermum Parkii Butter", 1)
	addAllergen(t, db, f.sheaID, "kuruyemiş", 3)

	f.hindistanCeviziID = addIngredient(t, db, "Hindistan Cevizi Yağı", "Cocos Nucifera Oil", 2)
	addAllergen(t, db, f.hindistanCeviziID, "hindistan cevizi", 4)

	f.parfumID = addIngredient(t, db, "Parfüm", "Parfum (Fragrance)", 8)
	addAllergen(t, db, f.parfumID, "parfüm", 8)

	f.linaloolID = addIngredient(t, db, "Linalool", "Linalool", 5)
	addAllergen(t, db, f.linaloolID, "parfüm", 6)
	addAllergen(t, db, f.linaloolID, "linalool", 6)

	// Ruj: lanolin + kuruyemiş + hindistan cevizi taşır, alkol taşımaz.
	f.rujID = addProduct(t, db, "Gül Yaprağı Ruj", "Atelier Noir", "ruj", 34.00)
	linkIngredients(t, db, f.rujID, f.sheaID, f.hindistanCeviziID, f.lanolinID)

	// Fondöten: denatüre alkol + parfüm taşır, lanolin taşımaz.
	f.fondotenID = addProduct(t, db, "Kadife Mat Fondöten", "Atelier Noir", "fondöten", 46.00)
	linkIngredients(t, db, f.fondotenID, f.denatureAlkolID, f.parfumID, f.linaloolID)

	return f
}

type allergenCheckResponse struct {
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name"`
	Matches     []struct {
		Allergen     string `json:"allergen"`
		Ingredient   string `json:"ingredient"`
		Severity     int    `json:"severity"`
		ConcernLevel int    `json:"concern_level"`
	} `json:"matches"`
	Safe           bool     `json:"safe"`
	Flags          []string `json:"flags"`
	UnmatchedTerms []string `json:"unmatched_terms"`
}

// matchedIngredients, yanıttaki eşleşen içerik adlarını döndürür.
func (r allergenCheckResponse) matchedIngredients() []string {
	out := make([]string, 0, len(r.Matches))
	for _, m := range r.Matches {
		out = append(out, m.Ingredient)
	}
	return out
}

func contains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func checkAllergens(t *testing.T, r *gin.Engine, productID int, terms []string) allergenCheckResponse {
	t.Helper()

	var resp allergenCheckResponse
	code := doJSON(t, r, http.MethodPost, "/ingredients/allergen-check", map[string]any{
		"product_id":     productID,
		"user_allergens": terms,
	}, &resp)
	mustStatus(t, code, statusOK, "allergen-check")
	return resp
}

// TestAllergenCheckFalsePositives, alt dize eşleştirmesinin ürettiği yanlış
// uyarıları yakalar. Bu testler, uygulamanın varlık sebebini koruyor:
// yanlış bir uyarı, kaçırılan bir uyarı kadar hızlı güven kaybettirir.
func TestAllergenCheckFalsePositives(t *testing.T) {
	_, r, db := newTestServer(t)
	f := setupAllergenFixture(t, db)

	tests := []struct {
		name        string
		productID   int
		terms       []string
		wantPresent []string
		wantAbsent  []string
	}{
		{
			name:       "alkol, yün alkolü yüzünden Lanolin'i işaretlememeli",
			productID:  f.rujID,
			terms:      []string{"alkol"},
			wantAbsent: []string{"Lanolin"},
		},
		{
			name:        "alkol, Denatüre Alkol'ü işaretlemeli",
			productID:   f.fondotenID,
			terms:       []string{"alkol"},
			wantPresent: []string{"Denatüre Alkol"},
		},
		{
			name:        "kuruyemiş, Hindistan Cevizi Yağı'nı işaretlememeli",
			productID:   f.rujID,
			terms:       []string{"kuruyemiş"},
			wantPresent: []string{"Shea Yağı"},
			wantAbsent:  []string{"Hindistan Cevizi Yağı"},
		},
		{
			name:        "lanolin, doğrudan Lanolin'i işaretlemeli",
			productID:   f.rujID,
			terms:       []string{"lanolin"},
			wantPresent: []string{"Lanolin"},
		},
		{
			name:        "parfüm, hem Parfüm hem Linalool'ü işaretlemeli",
			productID:   f.fondotenID,
			terms:       []string{"parfüm"},
			wantPresent: []string{"Parfüm", "Linalool"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := checkAllergens(t, r, tc.productID, tc.terms).matchedIngredients()

			for _, want := range tc.wantPresent {
				if !contains(got, want) {
					t.Errorf("%q işaretlenmeliydi, işaretlenmedi. Eşleşenler: %v", want, got)
				}
			}
			for _, notWant := range tc.wantAbsent {
				if contains(got, notWant) {
					t.Errorf("%q YANLIŞ işaretlendi. Eşleşenler: %v", notWant, got)
				}
			}
		})
	}
}

// TestAllergenCheckCaseAndDiacritics, Türkçe büyük/küçük harf ve aksan
// varyantlarının aynı sonucu vermesini doğrular.
func TestAllergenCheckCaseAndDiacritics(t *testing.T) {
	_, r, db := newTestServer(t)
	f := setupAllergenFixture(t, db)

	variants := []string{"parfüm", "PARFÜM", "Parfüm", "  parfüm  "}

	var baseline []string
	for i, v := range variants {
		got := checkAllergens(t, r, f.fondotenID, []string{v}).matchedIngredients()
		if i == 0 {
			baseline = got
			if len(baseline) == 0 {
				t.Fatalf("temel durum boş: %q hiçbir şey işaretlemedi", v)
			}
			continue
		}
		if len(got) != len(baseline) {
			t.Errorf("%q varyantı farklı sonuç verdi: %v (beklenen %v)", v, got, baseline)
		}
	}
}

// TestAllergenCheckUnmatchedTerms, tanınmayan girdinin sessizce yutulmadığını
// doğrular. Sessiz sıfır eşleşme bu uygulamada yanlış güven demektir:
// kullanıcı korunduğunu sanır.
func TestAllergenCheckUnmatchedTerms(t *testing.T) {
	_, r, db := newTestServer(t)
	f := setupAllergenFixture(t, db)

	resp := checkAllergens(t, r, f.fondotenID, []string{"parfüm", "zzzbilinmeyen"})

	if !contains(resp.UnmatchedTerms, "zzzbilinmeyen") {
		t.Errorf("tanınmayan terim bildirilmedi. unmatched_terms: %v", resp.UnmatchedTerms)
	}
	if contains(resp.UnmatchedTerms, "parfüm") {
		t.Errorf("eşleşen terim yanlışlıkla tanınmayan olarak bildirildi: %v", resp.UnmatchedTerms)
	}
}

// TestAllergenCheckSafeProduct, hiç alerjen taşımayan ürünün güvenli
// işaretlenmesini doğrular.
func TestAllergenCheckSafeProduct(t *testing.T) {
	_, r, db := newTestServer(t)

	su := addIngredient(t, db, "Su", "Aqua", 1)
	temiz := addProduct(t, db, "Sade Tonik", "Derma Basics", "tonik", 10.00)
	linkIngredients(t, db, temiz, su)

	resp := checkAllergens(t, r, temiz, []string{"parfüm"})

	if !resp.Safe {
		t.Errorf("alerjen taşımayan ürün güvenli işaretlenmeliydi: %+v", resp.Matches)
	}
	if len(resp.Matches) != 0 {
		t.Errorf("eşleşme olmamalıydı: %v", resp.matchedIngredients())
	}
}

// TestAllergenCheckNotFound, olmayan ürün için 404 döndüğünü doğrular.
func TestAllergenCheckNotFound(t *testing.T) {
	_, r, _ := newTestServer(t)

	code := doJSON(t, r, http.MethodPost, "/ingredients/allergen-check", map[string]any{
		"product_id":     999999,
		"user_allergens": []string{"parfüm"},
	}, nil)
	mustStatus(t, code, http.StatusNotFound, "olmayan ürün")
}

// ===== Filtre testleri =====

type ingredientListResponse struct {
	Total int `json:"total"`
	// Puanı olmadığı için max_concern süzgecinin dışında kalanların sayısı.
	UnscoredExcluded int `json:"unscored_excluded"`
	Ingredients      []struct {
		ID           int      `json:"id"`
		Name         string   `json:"name"`
		ConcernLevel *int     `json:"concern_level"`
		ScoreVersion *int     `json:"score_version"`
		ScoreSources []string `json:"score_sources"`
		Allergens    []string `json:"allergens"`
		SkinTypes    []string `json:"skin_types"`
	} `json:"ingredients"`
}

func (r ingredientListResponse) names() []string {
	out := make([]string, 0, len(r.Ingredients))
	for _, i := range r.Ingredients {
		out = append(out, i.Name)
	}
	return out
}

// TestListIngredientsAvoidAllergens, kaçınılacak alerjen filtresinin
// doğru içerikleri elediğini doğrular.
func TestListIngredientsAvoidAllergens(t *testing.T) {
	_, r, db := newTestServer(t)
	setupAllergenFixture(t, db)

	var resp ingredientListResponse
	code := doJSON(t, r, http.MethodGet, "/ingredients?avoid_allergens=parf%C3%BCm", nil, &resp)
	mustStatus(t, code, statusOK, "avoid_allergens filtresi")

	got := resp.names()
	for _, banned := range []string{"Parfüm", "Linalool"} {
		if contains(got, banned) {
			t.Errorf("%q elenmeliydi ama listede: %v", banned, got)
		}
	}
	if !contains(got, "Shea Yağı") {
		t.Errorf("alerjen taşımayan Shea Yağı listede olmalıydı: %v", got)
	}
}

// TestListIngredientsSkinTypeFilter, cilt tipi filtresini ve 'all' jokerini
// doğrular.
func TestListIngredientsSkinTypeFilter(t *testing.T) {
	_, r, db := newTestServer(t)

	su := addIngredient(t, db, "Su", "Aqua", 1)
	addSkinType(t, db, su, "all")

	salisilik := addIngredient(t, db, "Salisilik Asit", "Salicylic Acid", 4)
	addSkinType(t, db, salisilik, "oily")

	seramid := addIngredient(t, db, "Seramid NP", "Ceramide NP", 1)
	addSkinType(t, db, seramid, "dry")

	var resp ingredientListResponse
	code := doJSON(t, r, http.MethodGet, "/ingredients?skin_type=oily", nil, &resp)
	mustStatus(t, code, statusOK, "skin_type filtresi")

	got := resp.names()
	if !contains(got, "Salisilik Asit") {
		t.Errorf("yağlı cilde uygun içerik eksik: %v", got)
	}
	if !contains(got, "Su") {
		t.Errorf("'all' etiketli içerik her cilt tipinde dönmeliydi: %v", got)
	}
	if contains(got, "Seramid NP") {
		t.Errorf("kuru cilde özel içerik yağlı filtresinde dönmemeliydi: %v", got)
	}
}

// TestListIngredientsInvalidSkinType, geçersiz cilt tipinin 400 döndürdüğünü
// doğrular.
func TestListIngredientsInvalidSkinType(t *testing.T) {
	_, r, _ := newTestServer(t)

	code := doJSON(t, r, http.MethodGet, "/ingredients?skin_type=uzayli", nil, nil)
	mustStatus(t, code, http.StatusBadRequest, "geçersiz skin_type")
}

// TestListIngredientsMaxConcern, risk seviyesi üst sınırını doğrular.
func TestListIngredientsMaxConcern(t *testing.T) {
	_, r, db := newTestServer(t)

	addIngredient(t, db, "Su", "Aqua", 1)
	addIngredient(t, db, "Parfüm", "Parfum (Fragrance)", 8)

	var resp ingredientListResponse
	code := doJSON(t, r, http.MethodGet, "/ingredients?max_concern=3", nil, &resp)
	mustStatus(t, code, statusOK, "max_concern filtresi")

	got := resp.names()
	if !contains(got, "Su") {
		t.Errorf("düşük riskli içerik eksik: %v", got)
	}
	if contains(got, "Parfüm") {
		t.Errorf("yüksek riskli içerik elenmeliydi: %v", got)
	}
}
