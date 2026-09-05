package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func sampleProduct() obfProduct {
	return obfProduct{
		Code:        "0018787788059",
		ProductName: "Süper Mat Ruj",
		Brands:      "Maybelline New York, Maybelline",
		Quantity:    "5 ml",
		Lang:        "en",
		CategoriesTags: []string{
			"en:beauty", "en:makeup", "en:lip-makeup", "en:lipsticks",
		},
		Ingredients: []obfIngredient{
			{ID: "en:dimethicone", Text: "DIMETHICONE"},
			{ID: "en:e900", Text: "ISODODECANE"},
			{ID: "en:tocopherol", Text: "TOCOPHEROL"},
			{ID: "en:linalool", Text: "LINALOOL"},
		},
	}
}

func TestConvertKeepsLabelOrderAndIdentity(t *testing.T) {
	p, reason, ok := convert(sampleProduct())
	if !ok {
		t.Fatalf("ürün alınmadı: %s", reason)
	}

	if p.SourceID != "0018787788059" || p.GTIN != "0018787788059" {
		t.Errorf("kaynak kimliği yanlış: %+v", p)
	}
	if p.Brand != "Maybelline New York" {
		t.Errorf("marka %q, ilk marka alınmalıydı", p.Brand)
	}
	if p.Category != "ruj" {
		t.Errorf("kategori %q, beklenen ruj", p.Category)
	}
	if p.SourceURL != productURL+"0018787788059" {
		t.Errorf("kaynak adresi yanlış: %s", p.SourceURL)
	}

	want := []string{"DIMETHICONE", "ISODODECANE", "TOCOPHEROL", "LINALOOL"}
	if !reflect.DeepEqual(p.Ingredients, want) {
		t.Errorf("içerik listesi %v, beklenen %v", p.Ingredients, want)
	}
	if p.Incomplete {
		t.Error("dört içerikli ürün eksik veri sayıldı")
	}
}

// Kimliği eksik kayıtlar sessizce alınmaz; nedeni özette sayılabilsin diye
// sebep döner.
func TestConvertRejectsIncompleteRecords(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*obfProduct)
		reason string
	}{
		{"barkodsuz", func(p *obfProduct) { p.Code = "" }, "barkod yok"},
		{"adsız", func(p *obfProduct) { p.ProductName = ""; p.ProductNameEN = "" }, "ürün adı yok"},
		{"markasız", func(p *obfProduct) { p.Brands = "" }, "marka yok"},
		{"içeriksiz", func(p *obfProduct) { p.Ingredients = nil }, "içerik listesi yok"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := sampleProduct()
			tc.mutate(&raw)

			_, reason, ok := convert(raw)
			if ok {
				t.Fatal("eksik kayıt alındı")
			}
			if reason != tc.reason {
				t.Errorf("sebep %q, beklenen %q", reason, tc.reason)
			}
		})
	}
}

// Üç içerikten az listelenmiş ürün "eksik veri" sayılır: iki içerikli iki
// ürünün Jaccard benzerliği %100 çıkar ve bu bilgi değil, gürültüdür.
func TestConvertFlagsThinIngredientLists(t *testing.T) {
	raw := sampleProduct()
	raw.Ingredients = raw.Ingredients[:2]

	p, _, ok := convert(raw)
	if !ok {
		t.Fatal("iki içerikli ürün büsbütün elendi; katalogda görünmeli")
	}
	if !p.Incomplete {
		t.Error("iki içerikli ürün eksik veri olarak işaretlenmedi")
	}
}

func TestIngredientListFallsBackToText(t *testing.T) {
	raw := sampleProduct()
	raw.Ingredients = nil
	raw.IngredientsText = "Aqua, Glycerin, Tocopherol. May contain (+/-): CI 77891"

	got := ingredientList(raw)
	want := []string{"Aqua", "Glycerin", "Tocopherol"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("liste %v, beklenen %v", got, want)
	}
}

// OBF'nin taksonomi kimliği güvenilmez (dimetikon "en:e900" olarak geliyor);
// eşleştirme her zaman metin üzerinden yapılmalı.
func TestIngredientListUsesTextNotTaxonomyID(t *testing.T) {
	raw := sampleProduct()

	got := ingredientList(raw)
	for _, name := range got {
		if name == "en:e900" {
			t.Fatalf("taksonomi kimliği içerik adı olarak alındı: %v", got)
		}
	}
	if got[1] != "ISODODECANE" {
		t.Errorf("ikinci içerik %q, beklenen ISODODECANE", got[1])
	}
}

func TestIngredientListDeduplicates(t *testing.T) {
	raw := sampleProduct()
	raw.Ingredients = []obfIngredient{
		{Text: "Aqua"}, {Text: "AQUA"}, {Text: "Glycerin"},
	}

	if got := ingredientList(raw); len(got) != 2 {
		t.Errorf("liste %v, yinelenen ayıklanmadı", got)
	}
}

func TestCategoryPrefersMostSpecific(t *testing.T) {
	cases := []struct {
		tags []string
		want string
	}{
		{[]string{"en:beauty", "en:makeup", "en:lipsticks"}, "ruj"},
		{[]string{"en:beauty", "en:makeup"}, "makyaj"},
		{[]string{"en:hygiene", "en:toothpastes"}, "diş macunu"},
		// Karşılığı olmayan etiket okunabilir hale getirilip kullanılır:
		// uydurma bir kategori atamaktansa kaynaktakini göstermek dürüst.
		{[]string{"en:beauty", "en:cuticle-oils"}, "cuticle oils"},
		{nil, ""},
	}

	for _, tc := range cases {
		if got := category(tc.tags); got != tc.want {
			t.Errorf("category(%v) = %q, beklenen %q", tc.tags, got, tc.want)
		}
	}
}

func TestImageURLFromDumpFields(t *testing.T) {
	raw := sampleProduct()
	raw.Images = map[string]obfImage{
		"1":        {Rev: json.RawMessage(`"3"`)},
		"front_en": {Rev: json.RawMessage(`"17"`)},
	}

	want := imageBase + "001/878/778/8059/front_en.17.400.jpg"
	if got := imageURL(raw); got != want {
		t.Errorf("görsel adresi %q, beklenen %q", got, want)
	}
}

// Sürüm numarası kayıtlara göre hem metin hem sayı geliyor.
func TestImageURLAcceptsNumericRevision(t *testing.T) {
	raw := sampleProduct()
	raw.Images = map[string]obfImage{"front_en": {Rev: json.RawMessage(`4`)}}

	if got := imageURL(raw); got != imageBase+"001/878/778/8059/front_en.4.400.jpg" {
		t.Errorf("sayısal sürümde adres kurulamadı: %q", got)
	}
}

func TestImageURLEmptyWhenUnknown(t *testing.T) {
	raw := sampleProduct()
	raw.Images = map[string]obfImage{"1": {Rev: json.RawMessage(`"3"`)}}

	if got := imageURL(raw); got != "" {
		t.Errorf("ön yüz görseli yokken adres üretildi: %q", got)
	}
}

func TestCodePath(t *testing.T) {
	cases := map[string]string{
		"0018787788059": "001/878/778/8059",
		"00032831":      "00032831",
		"123456789":     "123/456/789/",
	}
	for code, want := range cases {
		if got := codePath(code); got != want {
			t.Errorf("codePath(%q) = %q, beklenen %q", code, got, want)
		}
	}
}

// Kaynağın kendi ayrıştırdığı dizi de artık içerebiliyor; bunlar katalogda
// "içerik" olarak görünmemeli.
func TestIngredientListDropsParsingDebris(t *testing.T) {
	raw := sampleProduct()
	raw.Ingredients = []obfIngredient{
		{Text: "Aqua"},
		{Text: `"8RN01-3"`},
		{Text: "! esté dermatologique"},
		{Text: "Glycerin"},
		{Text: "25% du total des ingrédients sont issus de l'agriculture"},
		{Text: "CI 77891"},
	}

	got := ingredientList(raw)
	want := []string{"Aqua", "Glycerin", "CI 77891"}

	if len(got) != len(want) {
		t.Fatalf("liste %v, beklenen %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("liste %v, beklenen %v", got, want)
		}
	}
}

// Veri kümesinin kendi adı kategori değil; süzgeç listesini kirletmemeli.
func TestCategoryIgnoresDatasetTags(t *testing.T) {
	got := category([]string{"en:open-beauty-facts", "en:shower-gels"})
	if got != "duş jeli" {
		t.Errorf("kategori %q, beklenen duş jeli", got)
	}

	// Yalnızca kirli etiket varsa kategori boş kalır: uydurmaktansa boş.
	if got := category([]string{"en:open-beauty-facts"}); got != "" {
		t.Errorf("kategori %q, boş beklenirdi", got)
	}
}

// Aynı ürün birden fazla dilde etiketleniyor. Türkçe arayüzde Fransızca bir
// kategori göstermenin faydası yok; karşılığı olmayan etiketlerde İngilizce
// olan tercih edilir.
func TestCategoryPrefersEnglishFallback(t *testing.T) {
	got := category([]string{"fr:gels-douche-hydratants", "en:moisturizing-shower-gels"})
	if got != "moisturizing shower gels" {
		t.Errorf("kategori %q, İngilizce etiket beklenirdi", got)
	}

	// Başka dil yoksa eldeki kullanılır: kategori hiç göstermemekten iyi.
	if got := category([]string{"fr:gels-douche-hydratants"}); got != "gels douche hydratants" {
		t.Errorf("kategori %q", got)
	}
}
