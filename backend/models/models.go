// models paketi, API katmanının paylaştığı alan tiplerini barındırır.
//
// JSON alan adları docs/API_SPEC.md'de belgelenen sözleşmenin parçasıdır ve
// bu yüzden İngilizce kalır; arayüz bunları Türkçe etiketlerle gösterir.
package models

import "time"

// Ingredient, güvenlik bilgileriyle birlikte tek bir kozmetik içeriktir.
// Name Türkçe yaygın adı, INCIName ise çevrilmeyen uluslararası INCI adını tutar.
type Ingredient struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	INCIName      string   `json:"inci_name"`
	Description   string   `json:"description,omitempty"`
	ConcernLevel  int      `json:"concern_level"` // EWG ölçeği 1-10
	SkinTypes     []string `json:"skin_types"`
	Allergens     []string `json:"allergens"`
	Benefits      []string `json:"benefits"`
	ProductsCount int      `json:"products_count,omitempty"`
	OrderIndex    int      `json:"order_index,omitempty"` // ürünün INCI listesindeki sırası
}

// Product, bir makyaj ürünü ve istendiğinde içerik listesidir.
type Product struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Brand       string       `json:"brand"`
	GTIN        string       `json:"gtin,omitempty"`
	Price       float64      `json:"price"`
	Currency    string       `json:"currency"`
	ImageURL    string       `json:"image_url,omitempty"`
	Category    string       `json:"category,omitempty"`
	Description string       `json:"description,omitempty"`
	SourceURL   string       `json:"source_url,omitempty"`
	Ingredients []Ingredient `json:"ingredients,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at,omitempty"`
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
	Allergen     string `json:"allergen"`
	Ingredient   string `json:"ingredient"`
	Severity     int    `json:"severity"`
	ConcernLevel int    `json:"concern_level"`
}

// Recommendation, bir ürün için önerilen muadil ya da alternatiftir.
type Recommendation struct {
	ID              int     `json:"id"`
	Type            string  `json:"type"` // dupe | alternative
	Name            string  `json:"name"`
	Brand           string  `json:"brand"`
	Price           float64 `json:"price"`
	Currency        string  `json:"currency"`
	ImageURL        string  `json:"image_url,omitempty"`
	SimilarityScore float64 `json:"similarity_score"`
	Reason          string  `json:"reason"`
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
