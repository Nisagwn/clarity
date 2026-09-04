package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/nisa/beauty-ingredient/scoring"
)

// Bu dosya Faz 3'ün sözünü koruyor: hiçbir puan elle atanmaz, her puan
// AB Tüzüğü 1223/2009 Eklerinden türetilir ve kaynağını gösterir.

type ingredientDetailResponse struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	ConcernLevel *int     `json:"concern_level"`
	ScoreVersion *int     `json:"score_version"`
	ScoreSources []string `json:"score_sources"`
	Scoring      *struct {
		Version int `json:"version"`
		Value   int `json:"value"`
		Rules   []struct {
			Key       string `json:"key"`
			Score     int    `json:"score"`
			Rationale string `json:"rationale"`
		} `json:"rules"`
		Sources    []string `json:"sources"`
		Regulatory struct {
			Annex              string `json:"annex"`
			AnnexEntry         string `json:"annex_entry"`
			DeclarableAllergen bool   `json:"declarable_allergen"`
			SourceURL          string `json:"source_url"`
		} `json:"regulatory"`
	} `json:"scoring"`
}

// TestRubricV1MatchesMigration, koddaki kural anahtarlarının göçle gelen
// rubrikle örtüştüğünü doğrular. Rubrik veritabanında yaşıyor; kod bir
// anahtarı yanlış yazarsa puan sessizce yanlış çıkardı.
func TestRubricV1MatchesMigration(t *testing.T) {
	db := testDB(t)

	rubric, err := scoring.LoadRubric(context.Background(), db, scoring.CurrentVersion)
	if err != nil {
		t.Fatalf("rubrik yüklenemedi: %v", err)
	}

	// Plandaki rubrik v1 tablosu.
	want := map[string]int{
		scoring.RuleAnnexIIBanned:      10,
		scoring.RuleDeclarableAllergen: 7,
		scoring.RuleAnnexIIIRestricted: 5,
		scoring.RuleAnnexVPreservative: 4,
		scoring.RuleAnnexVIUVFilter:    4,
		scoring.RuleAnnexIVColorant:    3,
		scoring.RuleUnrestricted:       2,
	}

	for key, score := range want {
		facts, ok := factsFor(key)
		if !ok {
			continue
		}
		if got := rubric.Apply(facts); got.Value != score {
			t.Errorf("%s: puan %d, göçte tanımlı olan %d", key, got.Value, score)
		}
	}
}

// factsFor, bir kuralı tetikleyen asgari mevzuat olgularını üretir.
func factsFor(rule string) (scoring.Facts, bool) {
	switch rule {
	case scoring.RuleAnnexIIBanned:
		return scoring.Facts{Annex: "II"}, true
	case scoring.RuleDeclarableAllergen:
		return scoring.Facts{Annex: "III", DeclarableAllergen: true}, true
	case scoring.RuleAnnexIIIRestricted:
		return scoring.Facts{Annex: "III"}, true
	case scoring.RuleAnnexVPreservative:
		return scoring.Facts{Annex: "V"}, true
	case scoring.RuleAnnexVIUVFilter:
		return scoring.Facts{Annex: "VI"}, true
	case scoring.RuleAnnexIVColorant:
		return scoring.Facts{Annex: "IV"}, true
	case scoring.RuleUnrestricted:
		return scoring.Facts{}, true
	}
	return scoring.Facts{}, false
}

// scoringFixture, mevzuat kaydı olan ve olmayan içerikleri birlikte kurar.
type scoringFixture struct {
	linaloolID      int // Ek III, bildirimli koku alerjeni
	fenoksietanolID int // Ek V, koruyucu
	niasinamidID    int // mevzuat kaydı yok
}

func setupScoringFixture(t *testing.T, db *sql.DB) scoringFixture {
	t.Helper()

	f := scoringFixture{}

	f.linaloolID = addIngredient(t, db, "Linalool", "Linalool", 0)
	addRegulatory(t, db, f.linaloolID, "III", "84", true)

	f.fenoksietanolID = addIngredient(t, db, "Fenoksietanol", "Phenoxyethanol", 0)
	addRegulatory(t, db, f.fenoksietanolID, "V", "29", false)

	// Mevzuat kaydı yok: bilinçli olarak puansız kalmalı.
	f.niasinamidID = addIngredient(t, db, "Niasinamid", "Niacinamide", 0)

	rescore(t, db)
	return f
}

func TestScoresAreDerivedFromRegulation(t *testing.T) {
	_, r, db := newTestServer(t)
	f := setupScoringFixture(t, db)

	cases := []struct {
		name string
		id   int
		want *int
	}{
		{"bildirimli koku alerjeni", f.linaloolID, intPtr(7)},
		{"Ek V koruyucu", f.fenoksietanolID, intPtr(4)},
		{"mevzuat kaydı yok", f.niasinamidID, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var resp ingredientDetailResponse
			code := doJSON(t, r, http.MethodGet, fmt.Sprintf("/ingredients/%d", tc.id), nil, &resp)
			mustStatus(t, code, statusOK, "içerik detayı")

			switch {
			case tc.want == nil && resp.ConcernLevel != nil:
				t.Fatalf("puansız içerik %d puan aldı", *resp.ConcernLevel)
			case tc.want != nil && resp.ConcernLevel == nil:
				t.Fatalf("puan null döndü, beklenen %d", *tc.want)
			case tc.want != nil && *resp.ConcernLevel != *tc.want:
				t.Fatalf("puan %d, beklenen %d", *resp.ConcernLevel, *tc.want)
			}
		})
	}
}

// Puansız içerik "0" değil "bilinmiyor" demektir: 0 göstermek onu ölçeğin
// en güvenli ucuna yerleştirirdi.
func TestUnscoredIngredientHasNoScoreOrExplanation(t *testing.T) {
	_, r, db := newTestServer(t)
	f := setupScoringFixture(t, db)

	var resp ingredientDetailResponse
	code := doJSON(t, r, http.MethodGet, fmt.Sprintf("/ingredients/%d", f.niasinamidID), nil, &resp)
	mustStatus(t, code, statusOK, "puansız içerik")

	if resp.ScoreVersion != nil {
		t.Errorf("puansız içerikte score_version %d", *resp.ScoreVersion)
	}
	if len(resp.ScoreSources) != 0 {
		t.Errorf("puansız içerikte kaynak var: %v", resp.ScoreSources)
	}
	if resp.Scoring != nil {
		t.Errorf("puansız içerikte açıklama var: %+v", resp.Scoring)
	}
}

// Her puanın yanında "neden bu puan?" yanıtı gelmeli: kural, gerekçe ve
// mevzuat atfı. Faz 3'ün tamamı bunun için.
func TestScoreExplanationCitesRegulation(t *testing.T) {
	_, r, db := newTestServer(t)
	f := setupScoringFixture(t, db)

	var resp ingredientDetailResponse
	code := doJSON(t, r, http.MethodGet, fmt.Sprintf("/ingredients/%d", f.linaloolID), nil, &resp)
	mustStatus(t, code, statusOK, "içerik detayı")

	if resp.Scoring == nil {
		t.Fatal("puan açıklamasız döndü")
	}
	if resp.Scoring.Version != scoring.CurrentVersion {
		t.Errorf("açıklama sürümü %d, beklenen %d", resp.Scoring.Version, scoring.CurrentVersion)
	}
	if len(resp.Scoring.Rules) == 0 {
		t.Fatal("puanı üreten kural listelenmedi")
	}

	rule := resp.Scoring.Rules[0]
	if rule.Key != scoring.RuleDeclarableAllergen {
		t.Errorf("uygulanan kural %q, beklenen %q", rule.Key, scoring.RuleDeclarableAllergen)
	}
	if strings.TrimSpace(rule.Rationale) == "" {
		t.Error("kural gerekçesiz döndü")
	}

	if len(resp.Scoring.Sources) == 0 {
		t.Fatal("puan kaynaksız döndü")
	}
	src := resp.Scoring.Sources[0]
	for _, want := range []string{"1223/2009", "Ek III", "giriş 84", testSourceURL} {
		if !strings.Contains(src, want) {
			t.Errorf("atıfta %q yok: %s", want, src)
		}
	}

	if resp.Scoring.Regulatory.Annex != "III" || !resp.Scoring.Regulatory.DeclarableAllergen {
		t.Errorf("mevzuat kaydı eksik döndü: %+v", resp.Scoring.Regulatory)
	}

	// Kaydedilen kaynaklar da aynı atfı taşımalı: liste uç noktası açıklamayı
	// dönmüyor, yalnızca bunları dönüyor.
	if len(resp.ScoreSources) == 0 || !strings.Contains(resp.ScoreSources[0], "1223/2009") {
		t.Errorf("score_sources mevzuata bağlanmıyor: %v", resp.ScoreSources)
	}
}

// Faz 3'ün asıl işi: seed'den gelen elle atanmış puanlar temizlenir.
func TestRecomputeClearsHandAssignedScores(t *testing.T) {
	_, _, db := newTestServer(t)

	// Göç öncesinden kalmış gibi: puanı var, dayanağı yok.
	id := addIngredient(t, db, "Talk", "Talc", 5)

	sum := rescore(t, db)
	if sum.Cleared != 1 {
		t.Errorf("silinen puan sayısı %d, beklenen 1", sum.Cleared)
	}

	var (
		concern *int
		version *int
	)
	err := db.QueryRow("SELECT concern_level, score_version FROM ingredients WHERE id = $1", id).
		Scan(&concern, &version)
	if err != nil {
		t.Fatalf("içerik okunamadı: %v", err)
	}
	if concern != nil {
		t.Errorf("dayanaksız puan duruyor: %d", *concern)
	}
	if version != nil {
		t.Errorf("dayanaksız score_version duruyor: %d", *version)
	}
}

// Puanı olmayan içerik max_concern süzgecinden geçmez — bilinmeyen bir puanı
// "yeterince düşük" saymak, olmayan bir güvence vermek olurdu. Kaç tanesinin
// elendiği yanıtta görünür.
func TestMaxConcernExcludesUnscoredAndReportsCount(t *testing.T) {
	_, r, db := newTestServer(t)
	setupScoringFixture(t, db)

	var resp ingredientListResponse
	code := doJSON(t, r, http.MethodGet, "/ingredients?max_concern=10", nil, &resp)
	mustStatus(t, code, statusOK, "max_concern filtresi")

	got := resp.names()
	if contains(got, "Niasinamid") {
		t.Errorf("puansız içerik süzgeçten geçti: %v", got)
	}
	if !contains(got, "Linalool") || !contains(got, "Fenoksietanol") {
		t.Errorf("puanlı içerikler eksik: %v", got)
	}
	if resp.UnscoredExcluded != 1 {
		t.Errorf("unscored_excluded %d, beklenen 1", resp.UnscoredExcluded)
	}
}

// Süzgeç yokken puansız içerik listede kalır: gizlemek, katalogda olmadığı
// izlenimi verirdi.
func TestUnscoredIngredientsStillListed(t *testing.T) {
	_, r, db := newTestServer(t)
	setupScoringFixture(t, db)

	var resp ingredientListResponse
	code := doJSON(t, r, http.MethodGet, "/ingredients", nil, &resp)
	mustStatus(t, code, statusOK, "içerik listesi")

	if !contains(resp.names(), "Niasinamid") {
		t.Errorf("puansız içerik listeden düştü: %v", resp.names())
	}
	if resp.UnscoredExcluded != 0 {
		t.Errorf("süzgeç yokken unscored_excluded %d", resp.UnscoredExcluded)
	}
}

func intPtr(v int) *int { return &v }
