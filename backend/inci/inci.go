// inci paketi, INCI adlarını karşılaştırılabilir hale getirir ve etiket
// üzerindeki içerik listelerini ayrıştırır.
//
// İki içe aktarım da (CosIng mevzuat verisi ve Open Beauty Facts ürün verisi)
// aynı katalog satırlarına bağlanmak zorunda. Normalleştirme iki yerde ayrı
// yazılırsa, aynı madde bir dosyada eşleşip diğerinde eşleşmez ve bunu kimse
// fark etmez — bu yüzden tek yerde duruyor.
package inci

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	parenthetical = regexp.MustCompile(`\([^)]*\)`)
	nonAlphaNum   = regexp.MustCompile(`[^\p{L}\p{N}]+`)

	// Etiketin başındaki "INGREDIENTS:" türü başlıklar.
	// Türkçe "İ" (U+0130) Go'nun (?i) katlamasıyla "i"ye eşlenmediği için
	// ayrıca yazılıyor.
	listPrefix = regexp.MustCompile(`(?i)^\s*(ingredient[sé]?s?|ingr[ée]dients?|inci|composition|bile[şs]enler|[İIi][çc]indekiler)\s*[:\-–]\s*`)

	// "May contain (+/-): CI 77891" — koşullu renklendiriciler. Bunlar ürünün
	// içinde OLMAYABİLİR; varmış gibi saymak yanlış uyarı üretir.
	mayContain = regexp.MustCompile(`(?i)(may\s+contain|peut\s+contenir|kann\s+enthalten|i[çc]erebilir)\b|\[?\+/-\]?|±`)
)

// Normalize, iki INCI adını karşılaştırılabilir hale getirir.
//
// Katalogdaki "Titanium Dioxide (CI 77891)" ile kaynaktaki "Titanium dioxide"
// aynı maddedir; parantez içi ekler, noktalama ve büyük/küçük harf farkı
// eşleşmeyi engellememelidir.
//
// Eşleştirme yine de TAM eşleşmedir: alt dize karşılaştırması "Alcohol" ile
// "Alcohol Denat."ı birbirine karıştırırdı.
func Normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = parenthetical.ReplaceAllString(s, " ")
	s = nonAlphaNum.ReplaceAllString(s, " ")
	return strings.Join(strings.Fields(s), " ")
}

// maxTokenLength, bir INCI adının makul üst sınırı. Bunu aşan parçalar
// ayrıştırma artığıdır (ayraçsız yazılmış bir paragraf gibi).
const maxTokenLength = 120

// ParseList, etiketteki serbest metin içerik listesini tek tek adlara böler.
//
// "May contain (+/-)" bölümü DIŞARIDA bırakılır: oradaki renklendiriciler
// üründe olmayabilir. Varmış gibi kaydetmek, kullanıcıya taşımadığı bir
// alerjen için uyarı vermek olurdu.
//
// Sıra korunur — INCI listesi azalan konsantrasyon sırasıdır ve order_index
// bu sıradan gelir.
func ParseList(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	if loc := mayContain.FindStringIndex(text); loc != nil {
		text = text[:loc[0]]
	}
	text = listPrefix.ReplaceAllString(text, "")

	fields := strings.FieldsFunc(text, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '•' || r == '|'
	})

	out := make([]string, 0, len(fields))
	seen := map[string]bool{}

	for _, f := range fields {
		token := cleanToken(f)
		if token == "" || len(token) > maxTokenLength {
			continue
		}

		if !IsPlausibleName(token) {
			continue
		}

		key := Normalize(token)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, token)
	}
	return out
}

// cleanToken, tek bir içerik adının kenarlarını temizler.
func cleanToken(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, ".*_-–—•·\"'`[]{}")
	return strings.Join(strings.Fields(s), " ")
}

// INCI adı için yapısal sınırlar. Etiket ayrıştırması her zaman temiz
// çıkmıyor: kaynaklarda pazarlama cümleleri, sertifika notları ve tırnak
// içinde kalmış parçalar içerik adı gibi görünebiliyor.
const (
	maxNameLength = 80 // en uzun INCI adları bunun altında kalır
	maxNameWords  = 8
)

// rejectedInName, INCI adlarında yeri olmayan karakterler. Yüzde işareti ve
// cümle noktalaması, ayrıştırma artığının en güvenilir işareti.
const rejectedInName = `!?<>"“”„%*=|&`

// colourIndex, "CI 77891" biçimindeki renklendirici numaraları.
var colourIndex = regexp.MustCompile(`(?i)^ci\s?\d{4,6}$`)

// IsPlausibleName, bir metnin INCI adı OLABİLECEĞİNİ söyler.
//
// Kesin bir doğrulama değil — kanonik bir INCI listesi olmadan mümkün de
// değil. Amaç, katalogun "içerik" diye ayrıştırma artığıyla dolmasını
// engellemek: aday içerikler küratörlükten geçmeden puanlanmıyor ama yine de
// arayüzde görünüyorlar.
func IsPlausibleName(s string) bool {
	s = strings.TrimSpace(s)

	runes := []rune(s)
	if len(runes) < 2 || len(runes) > maxNameLength {
		return false
	}
	if strings.ContainsAny(s, rejectedInName) {
		return false
	}
	// INCI adları harfle ya da rakamla başlar ("1,2-Hexanediol"); tırnak veya
	// noktalamayla başlayan şey bir addan kalan parçadır.
	if !unicode.IsLetter(runes[0]) && !unicode.IsDigit(runes[0]) {
		return false
	}
	if len(strings.Fields(s)) > maxNameWords {
		return false
	}

	// Renk indeksi numaraları çoğunlukla rakamdan oluşur ama geçerli
	// adlardır; oran kuralının dışında tutulurlar.
	if colourIndex.MatchString(s) {
		return true
	}

	letters := 0
	for _, r := range runes {
		if unicode.IsLetter(r) {
			letters++
		}
	}
	// Ağırlıklı olarak rakam ve simgeden oluşan bir dizi ("8RN01-3") ad değil.
	return letters > 0 && float64(letters)/float64(len(runes)) >= 0.5
}
