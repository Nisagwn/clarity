// models paketi, API katmanının paylaştığı alan tiplerini barındırır.
//
// JSON alan adları docs/API_SPEC.md'de belgelenen sözleşmenin parçasıdır ve
// bu yüzden İngilizce kalır; arayüz bunları Türkçe etiketlerle gösterir.
package models

import (
	"time"

	"github.com/nisa/beauty-ingredient/scoring"
)

// Ingredient, güvenlik bilgileriyle birlikte tek bir kozmetik içeriktir.
// Name Türkçe yaygın adı, INCIName ise çevrilmeyen uluslararası INCI adını tutar.
type Ingredient struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	INCIName    string `json:"inci_name"`
	Description string `json:"description,omitempty"`

	// ConcernLevel 1-10 ölçeğinde bir endişe seviyesidir ve mevzuattan
	// TÜRETİLİR (bkz. scoring paketi). Mevzuat kaydı olmayan içerikte null
	// kalır: 0 göstermek "güvenli" diye okunurdu, oysa bilinen bir şey yok.
	ConcernLevel *int `json:"concern_level"`
	// ScoreVersion, puanı üreten rubrik sürümü; ScoreSources ise puanın
	// dayandığı mevzuat atıflarıdır.
	ScoreVersion *int     `json:"score_version"`
	ScoreSources []string `json:"score_sources"`
	// Scoring, "neden bu puan?" açıklamasıdır ve yalnızca detay uç
	// noktasında doldurulur.
	Scoring *ScoreExplanation `json:"scoring,omitempty"`

	SkinTypes     []string `json:"skin_types"`
	Allergens     []string `json:"allergens"`
	Benefits      []string `json:"benefits"`
	ProductsCount int      `json:"products_count,omitempty"`
	OrderIndex    int      `json:"order_index,omitempty"` // ürünün INCI listesindeki sırası
}

// ScoreExplanation, bir puanın tam gerekçesi: hangi kurallar uygulandı ve
// hangi mevzuata dayanıyor. Puan hiçbir yerde gerekçesiz gösterilmez.
type ScoreExplanation struct {
	scoring.Score
	Regulatory Regulatory `json:"regulatory"`
}

// Regulatory, bir içeriğin ingredient_regulatory kaydıdır: AB Tüzüğü
// 1223/2009 Eklerindeki yeri ve varsa SCCS görüşü.
type Regulatory struct {
	CASNumber          string    `json:"cas_number,omitempty"`
	ECNumber           string    `json:"ec_number,omitempty"`
	Annex              string    `json:"annex,omitempty"`
	AnnexEntry         string    `json:"annex_entry,omitempty"`
	Restriction        string    `json:"restriction,omitempty"`
	MaxConcentration   *float64  `json:"max_concentration,omitempty"`
	DeclarableAllergen bool      `json:"declarable_allergen"`
	SCCSOpinion        string    `json:"sccs_opinion,omitempty"`
	SCCSAdverse        bool      `json:"sccs_adverse"`
	SourceURL          string    `json:"source_url"`
	FetchedAt          time.Time `json:"fetched_at"`
}

// Facts, kaydın puanlamayı ilgilendiren alanlarını scoring paketine taşır.
func (r Regulatory) Facts() scoring.Facts {
	return scoring.Facts{
		Annex:              r.Annex,
		AnnexEntry:         r.AnnexEntry,
		DeclarableAllergen: r.DeclarableAllergen,
		SCCSAdverse:        r.SCCSAdverse,
		SCCSOpinion:        r.SCCSOpinion,
		SourceURL:          r.SourceURL,
	}
}

// Product, bir makyaj ürünü ve istendiğinde içerik listesidir.
type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Brand string `json:"brand"`
	GTIN  string `json:"gtin,omitempty"`
	// Fiyat bilinmiyorsa null. Open Beauty Facts fiyat taşımıyor; 0 göstermek
	// ürünü bedava sanmaya davet ederdi. Fiyat takibi Faz 9'da.
	Price       *float64     `json:"price"`
	Currency    string       `json:"currency"`
	ImageURL    string       `json:"image_url,omitempty"`
	Category    string       `json:"category,omitempty"`
	Description string       `json:"description,omitempty"`
	SourceURL   string       `json:"source_url,omitempty"`
	Ingredients []Ingredient `json:"ingredients,omitempty"`

	// Kaynak ve lisans. ODbL atıf gerektiriyor: veriyi gösteren her yer
	// nereden geldiğini söylemek zorunda, bu yüzden alanlar API sözleşmesinin
	// parçası.
	Source     string     `json:"source,omitempty"`
	SourceID   string     `json:"source_id,omitempty"`
	License    string     `json:"license,omitempty"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	// DataQuality "ok" veya "incomplete". Eksik içerik listeli ürünler
	// katalogda görünür ama muadil hesabına girmez.
	DataQuality string `json:"data_quality"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// UserProfile, kullanıcının cilt profili ve kaçındığı alerjenlerdir.
type UserProfile struct {
	ID        int       `json:"id"`
	Email     string    `json:"email,omitempty"`
	SkinType  string    `json:"skin_type"`
	Allergens []string  `json:"allergens"`
	CreatedAt time.Time `json:"created_at"`
}

// AllergenMatch, kullanıcının alerjen listesine takılan bir ürün içeriğidir.
type AllergenMatch struct {
	Allergen   string `json:"allergen"`
	Ingredient string `json:"ingredient"`
	Severity   int    `json:"severity"`
	// Puanlanmamış içerikte null; uyarının kendisi puandan bağımsızdır.
	ConcernLevel *int `json:"concern_level"`
}

// Recommendation, bir ürün için önerilen muadil ya da alternatiftir.
type Recommendation struct {
	ID              int      `json:"id"`
	Type            string   `json:"type"` // dupe | alternative
	Name            string   `json:"name"`
	Brand           string   `json:"brand"`
	Price           *float64 `json:"price"`
	Currency        string   `json:"currency"`
	ImageURL        string   `json:"image_url,omitempty"`
	SimilarityScore float64  `json:"similarity_score"`
	Reason          string   `json:"reason"`
}

// ValidSkinTypes, profillerde ve filtrelerde kabul edilen cilt tipleridir.
// Değerler API sözleşmesi gereği İngilizcedir; SkinTypeLabels karşılıklarını verir.
var ValidSkinTypes = []string{"oily", "dry", "combination", "sensitive", "normal"}

// SkinTypeLabels, API değerlerinin Türkçe karşılıklarıdır. Hata mesajlarında
// kullanıcıya bu etiketler gösterilir.
var SkinTypeLabels = map[string]string{
	"oily":        "yağlı",
	"dry":         "kuru",
	"combination": "karma",
	"sensitive":   "hassas",
	"normal":      "normal",
}

// IsValidSkinType, s'nin desteklenen bir cilt tipi olup olmadığını bildirir.
func IsValidSkinType(s string) bool {
	_, ok := SkinTypeLabels[s]
	return ok
}

// SkinTypeHint, hata mesajlarında kullanılmak üzere geçerli cilt tiplerini
// "oily (yağlı), dry (kuru), ..." biçiminde döndürür.
func SkinTypeHint() string {
	out := ""
	for i, v := range ValidSkinTypes {
		if i > 0 {
			out += ", "
		}
		out += v + " (" + SkinTypeLabels[v] + ")"
	}
	return out
}
