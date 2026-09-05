package main

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/nisa/beauty-ingredient/inci"
)

// obfProduct, Open Beauty Facts kaydının ilgilendiğimiz alanları.
//
// Dump satırlarında 140'tan fazla alan var ve bazılarının tipi kayıttan
// kayda değişiyor (aynı alan bir yerde sayı, başka yerde metin). Yalnızca
// kullandığımız alanları çözüyoruz: tanımadığımız bir alanın tipi tutmadığı
// için sağlam bir ürünü atlamak istemiyoruz.
type obfProduct struct {
	Code              string              `json:"code"`
	ProductName       string              `json:"product_name"`
	ProductNameEN     string              `json:"product_name_en"`
	Brands            string              `json:"brands"`
	Quantity          string              `json:"quantity"`
	Lang              string              `json:"lang"`
	CategoriesTags    []string            `json:"categories_tags"`
	IngredientsText   string              `json:"ingredients_text"`
	IngredientsTextEN string              `json:"ingredients_text_en"`
	Ingredients       []obfIngredient     `json:"ingredients"`
	Images            map[string]obfImage `json:"images"`
	ImageURL          string              `json:"image_url"`
	LastModified      int64               `json:"last_modified_t"`
}

// obfIngredient, OBF'nin kendi ayrıştırdığı içerik girdisi.
//
// id alanı taksonomi kimliğidir ama güvenilmez: dimetikon "en:e900" (gıda katkı
// numarası) olarak da gelebiliyor. Eşleştirme her zaman text üzerinden yapılır.
type obfIngredient struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// obfImage, yalnızca sürüm numarasını taşır; görsel adresi ondan kurulur.
// Rev bazı kayıtlarda metin ("17"), bazılarında sayı (17) geliyor.
type obfImage struct {
	Rev json.RawMessage `json:"rev"`
}

// product, içe aktarılmaya hazır ürün.
type product struct {
	SourceID    string
	Name        string
	Brand       string
	GTIN        string
	Category    string
	Description string
	ImageURL    string
	SourceURL   string
	Ingredients []string // INCI listesi, etiketteki sırayla
	Incomplete  bool
}

// minIngredients, bir içerik listesinin muadil hesabına girebilmesi için
// gereken en az içerik sayısı.
//
// Crowdsource veri eksik olabiliyor. İki içeriği listelenmiş iki ürünün
// Jaccard benzerliği %100 çıkar; bu "aynı ürün" demek değil, "ikisinin de
// listesi eksik" demektir. Eksik listeli ürünler katalogda görünür ama
// muadil hesabına sokulmaz.
const minIngredients = 3

// productURL, ürünün Open Beauty Facts sayfası — ODbL atfı buraya bağlanır.
const productURL = "https://world.openbeautyfacts.org/product/"

// imageBase, OBF görsel sunucusu. Görseller CC-BY-SA lisanslı.
const imageBase = "https://images.openbeautyfacts.org/images/products/"

// convert, bir OBF kaydını içe aktarılabilir ürüne çevirir.
// İkinci dönüş değeri false ise kayıt kullanılamaz; neden ilk değerin
// yerine dönen sebep metnindedir.
func convert(p obfProduct) (product, string, bool) {
	code := strings.TrimSpace(p.Code)
	if code == "" {
		return product{}, "barkod yok", false
	}

	name := firstNonEmpty(p.ProductName, p.ProductNameEN)
	if name == "" {
		return product{}, "ürün adı yok", false
	}

	brand := firstBrand(p.Brands)
	if brand == "" {
		// Marka arayüzde ürünün kimliğinin yarısı; "bilinmeyen marka" diye
		// bir şey uydurmaktansa kaydı almıyoruz.
		return product{}, "marka yok", false
	}

	ingredients := ingredientList(p)
	if len(ingredients) == 0 {
		return product{}, "içerik listesi yok", false
	}

	return product{
		SourceID:    code,
		Name:        trimTo(name, 255),
		Brand:       trimTo(brand, 255),
		GTIN:        trimTo(code, 50),
		Category:    category(p.CategoriesTags),
		Description: strings.TrimSpace(p.Quantity),
		ImageURL:    imageURL(p),
		SourceURL:   productURL + code,
		Ingredients: ingredients,
		Incomplete:  len(ingredients) < minIngredients,
	}, "", true
}

// ingredientList, ürünün INCI listesini etiketteki sırayla döndürür.
//
// Önce OBF'nin kendi ayrıştırdığı dizi denenir; yoksa serbest metin
// ayrıştırılır. İç içe geçmiş bileşenler (bir maddenin kendi alt içerikleri)
// bilinçli olarak açılmıyor: etikette yazan üst seviye listedir ve
// order_index onu yansıtmalı.
func ingredientList(p obfProduct) []string {
	if len(p.Ingredients) > 0 {
		out := make([]string, 0, len(p.Ingredients))
		seen := map[string]bool{}

		for _, ing := range p.Ingredients {
			// Kaynağın kendi ayrıştırması da her zaman temiz değil: pazarlama
			// cümleleri ve tırnak içinde kalmış parçalar içerik girdisi gibi
			// gelebiliyor.
			text := strings.TrimSpace(ing.Text)
			if !inci.IsPlausibleName(text) {
				continue
			}
			key := inci.Normalize(text)
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, text)
		}
		if len(out) > 0 {
			return out
		}
	}

	return inci.ParseList(firstNonEmpty(p.IngredientsText, p.IngredientsTextEN))
}

// categoryLabels, OBF kategori etiketlerinin Türkçe karşılıkları.
// Katalog global; etiketler İngilizce geliyor, arayüz Türkçe gösteriyor.
var categoryLabels = map[string]string{
	"en:lipsticks":        "ruj",
	"en:lip-glosses":      "dudak parlatıcısı",
	"en:lip-balms":        "dudak balmı",
	"en:foundations":      "fondöten",
	"en:concealers":       "kapatıcı",
	"en:face-powders":     "pudra",
	"en:blushes":          "allık",
	"en:mascaras":         "maskara",
	"en:eyeshadows":       "far",
	"en:eyeliners":        "eyeliner",
	"en:nail-polish":      "oje",
	"en:nail-polishes":    "oje",
	"en:makeup":           "makyaj",
	"en:make-up":          "makyaj",
	"en:shampoos":         "şampuan",
	"en:conditioners":     "saç kremi",
	"en:hair-care":        "saç bakımı",
	"en:soaps":            "sabun",
	"en:shower-gels":      "duş jeli",
	"en:deodorants":       "deodorant",
	"en:toothpastes":      "diş macunu",
	"en:sunscreens":       "güneş koruyucu",
	"en:sun-protection":   "güneş koruyucu",
	"en:moisturizers":     "nemlendirici",
	"en:face-creams":      "yüz kremi",
	"en:hand-creams":      "el kremi",
	"en:body-lotions":     "vücut losyonu",
	"en:serums":           "serum",
	"en:cleansers":        "temizleyici",
	"en:face-cleansers":   "yüz temizleyici",
	"en:micellar-waters":  "misel su",
	"en:perfumes":         "parfüm",
	"en:fragrances":       "parfüm",
	"en:skincare":         "cilt bakımı",
	"en:beauty":           "kozmetik",
	"en:hygiene":          "hijyen",
	"en:body-care":        "vücut bakımı",
	"en:eau-de-toilettes": "eau de toilette",

	// Kaynakta tekil/çoğul ve eşanlamlı etiketler bir arada kullanılıyor.
	"en:sunscreen":              "güneş koruyucu",
	"en:toothpaste":             "diş macunu",
	"en:shampoo":                "şampuan",
	"en:soap":                   "sabun",
	"en:anti-dandruff-shampoos": "kepek şampuanı",
	"en:hair-conditioners":      "saç kremi",
	"en:hair-dyes":              "saç boyası",
	"en:hair":                   "saç bakımı",
	"en:body-milks":             "vücut sütü",
	"en:facial-creams":          "yüz kremi",
	"en:lip-cosmetics":          "dudak ürünü",
	"en:mouthwash":              "ağız gargarası",
	"en:liquid-soaps":           "sıvı sabun",
}

// ignoredCategories, kategori olmayan etiketler. Veri kümesinin kendi adı
// ürünü tarif etmiyor; kategori diye gösterilirse süzgeç listesini kirletir.
var ignoredCategories = map[string]bool{
	"en:open-beauty-facts": true,
	"en:products":          true,
	"en:beauty-products":   true,
	"en:cosmetics":         true,
}

var tagPrefix = regexp.MustCompile(`^[a-z]{2}:`)

// genericCategories, ürünü tarif etmeyen üst başlıklar. Bunlar yalnızca
// başka hiçbir şey bulunamazsa kullanılır.
var genericCategories = map[string]bool{
	"en:beauty":    true,
	"en:makeup":    true,
	"en:make-up":   true,
	"en:hygiene":   true,
	"en:skincare":  true,
	"en:body-care": true,
	"en:hair-care": true,
}

// category, kategori etiketlerinden ürünü en iyi tarif edeni seçer.
//
// OBF etiketleri sıralı gelmiyor — bir rujda "en:beauty" listenin ortasında
// olabiliyor — bu yüzden sıraya değil, etiketin kendisine bakılır:
//
//  1. Türkçe karşılığı olan özgül bir etiket ("en:lipsticks" → ruj)
//  2. Karşılığı olmayan ama özgül bir etiket, okunabilir hâliyle. Uydurma bir
//     kategori atamaktansa kaynaktakini göstermek dürüst.
//  3. Son çare olarak üst başlık ("en:makeup" → makyaj)
func category(tags []string) string {
	var generic, fallback string

	for i := len(tags) - 1; i >= 0; i-- {
		tag := strings.ToLower(strings.TrimSpace(tags[i]))
		if tag == "" || ignoredCategories[tag] {
			continue
		}

		if label, mapped := categoryLabels[tag]; mapped {
			if !genericCategories[tag] {
				return label
			}
			if generic == "" {
				generic = label
			}
			continue
		}

		// Karşılığı olmayan etiket olduğu gibi kullanılabilir ama yalnızca
		// İngilizcesi: kaynakta aynı ürün "en:shower-gels" ve
		// "fr:gels-douche" olarak birlikte etiketlenebiliyor ve Türkçe
		// arayüzde Fransızca bir kategori göstermenin kimseye faydası yok.
		if strings.HasPrefix(tag, "en:") {
			return readableTag(tag)
		}
		if fallback == "" {
			fallback = readableTag(tag)
		}
	}

	if generic != "" {
		return generic
	}
	return fallback
}

// readableTag, "en:shower-gels" etiketini "shower gels" hâline getirir.
func readableTag(tag string) string {
	return trimTo(strings.ReplaceAll(tagPrefix.ReplaceAllString(tag, ""), "-", " "), 100)
}

// imageURL, ürünün ön yüz görselinin adresini kurar.
//
// Dump'ta hazır adres yok, yalnızca images sözlüğü var. Adres biçimi:
//
//	.../products/001/878/778/8059/front_en.17.400.jpg
//
// Görseller CC-BY-SA; ürün sayfasındaki atıf bu yüzden zorunlu.
func imageURL(p obfProduct) string {
	if url := strings.TrimSpace(p.ImageURL); url != "" {
		return url // API modunda hazır geliyor
	}

	key, rev := frontImage(p)
	if key == "" || rev == "" {
		return ""
	}
	return imageBase + codePath(p.Code) + "/" + key + "." + rev + ".400.jpg"
}

// frontImage, ürün dilini tercih ederek bir ön yüz görseli seçer.
func frontImage(p obfProduct) (string, string) {
	preferred := "front_" + strings.ToLower(strings.TrimSpace(p.Lang))

	if img, ok := p.Images[preferred]; ok {
		if rev := revString(img.Rev); rev != "" {
			return preferred, rev
		}
	}

	// Ürün dilinde görsel yoksa herhangi bir ön yüz görseli iş görür.
	for key, img := range p.Images {
		if !strings.HasPrefix(key, "front") {
			continue
		}
		if rev := revString(img.Rev); rev != "" {
			return key, rev
		}
	}
	return "", ""
}

// revString, sürüm numarasını metne çevirir; alan hem "17" hem 17 gelebiliyor.
func revString(raw json.RawMessage) string {
	s := strings.Trim(strings.TrimSpace(string(raw)), `"`)
	if s == "" || s == "null" {
		return ""
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return ""
		}
	}
	return s
}

// codePath, barkodu görsel dizin yoluna çevirir: 8 karakterden uzun kodlar
// üçerli gruplara ayrılır.
func codePath(code string) string {
	code = strings.TrimSpace(code)
	if len(code) <= 8 {
		return code
	}
	return code[0:3] + "/" + code[3:6] + "/" + code[6:9] + "/" + code[9:]
}

// firstBrand, virgülle ayrılmış marka listesinden ilkini alır.
func firstBrand(brands string) string {
	first, _, _ := strings.Cut(brands, ",")
	return strings.TrimSpace(first)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	return ""
}

// trimTo, metni kolon sınırına kırpar (rune sınırında).
func trimTo(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return strings.TrimSpace(string(runes[:max]))
}
