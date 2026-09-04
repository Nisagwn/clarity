package main

import (
	"regexp"
	"strconv"
	"strings"
)

// record, tek bir içeriğin ingredient_regulatory satırı.
type record struct {
	ingredientID     int
	cas              string
	ec               string
	annex            string
	annexEntry       string
	restriction      string
	maxConcentration *float64
	declarable       bool
	sccsOpinion      string
	sccsAdverse      bool
	sourceURL        string
}

// annexRank, Eklerin kısıtlama ağırlığı. Bir madde birden fazla Ek'te
// geçebilir (örneğin hem Ek III'te kısıtlı hem Ek V'te koruyucu); o zaman
// puanı belirleyen daha ağır olan kısıtlamadır.
var annexRank = map[string]int{
	"II":  5, // yasaklı
	"III": 4, // kısıtlı
	"V":   3, // koruyucu
	"VI":  2, // UV filtresi
	"IV":  1, // renklendirici
}

// merge, mevcut kaydı yeni içe aktarımla birleştirir.
//
// İçe aktarımlar üst üste binebilir: Ek III dosyası ile bildirimli koku
// alerjenleri listesi aynı maddeye dokunur. Yeni dosya eskisini körlemesine
// ezerse, önceki dosyadan gelen bilgi sessizce kaybolur.
func merge(existing, incoming record) record {
	out := existing
	out.ingredientID = incoming.ingredientID

	// Daha ağır kısıtlama kazanır; Ek'e bağlı alanlar onunla birlikte gelir.
	if annexRank[incoming.annex] > annexRank[out.annex] {
		out.annex = incoming.annex
		out.annexEntry = incoming.annexEntry
		out.restriction = incoming.restriction
		out.maxConcentration = incoming.maxConcentration
	} else if incoming.annex == out.annex {
		out.annexEntry = firstNonEmpty(incoming.annexEntry, out.annexEntry)
		out.restriction = firstNonEmpty(incoming.restriction, out.restriction)
		if incoming.maxConcentration != nil {
			out.maxConcentration = incoming.maxConcentration
		}
	}

	out.cas = firstNonEmpty(incoming.cas, out.cas)
	out.ec = firstNonEmpty(incoming.ec, out.ec)
	out.sccsOpinion = firstNonEmpty(incoming.sccsOpinion, out.sccsOpinion)
	out.sourceURL = firstNonEmpty(incoming.sourceURL, out.sourceURL)

	// Bayraklar yalnızca eklenir: bir dosya "bildirimli alerjen" dediyse,
	// bunu söylemeyen başka bir dosya o bilgiyi geri alamaz.
	out.declarable = out.declarable || incoming.declarable
	out.sccsAdverse = out.sccsAdverse || incoming.sccsAdverse

	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	return ""
}

var (
	parenthetical = regexp.MustCompile(`\([^)]*\)`)
	nonAlphaNum   = regexp.MustCompile(`[^\p{L}\p{N}]+`)
	leadingNumber = regexp.MustCompile(`^[0-9]+(?:[.,][0-9]+)?`)
)

// normalizeINCI, iki INCI adını karşılaştırılabilir hale getirir.
//
// Katalogdaki "Titanium Dioxide (CI 77891)" ile CosIng'deki "Titanium dioxide"
// aynı maddedir; parantez içi ekler, noktalama ve büyük/küçük harf farkı
// eşleşmeyi engellememelidir. Eşleştirme yine de TAM eşleşmedir: alt dize
// karşılaştırması "Alcohol" ile "Alcohol denat."ı birbirine karıştırırdı.
func normalizeINCI(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = parenthetical.ReplaceAllString(s, " ")
	s = nonAlphaNum.ReplaceAllString(s, " ")
	return strings.Join(strings.Fields(s), " ")
}

// normalizeAnnex, "Annex III", "ek iii", "III" yazımlarını tek biçime getirir.
func normalizeAnnex(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "ANNEX ")
	s = strings.TrimPrefix(s, "EK ")
	s = strings.TrimSpace(s)
	if _, known := annexRank[s]; !known {
		return ""
	}
	return s
}

// parseConcentration, "1 %", "0,5 %" gibi değerlerden sayıyı çıkarır.
//
// Serbest metin ("5 % asit olarak, yalnızca profesyonel kullanım") sayıya
// indirgenmez: nil döner ve metnin tamamı restriction alanında durur. Yarım
// anlaşılmış bir sınır, sınırsız görünmekten daha yanıltıcıdır.
func parseConcentration(s string) *float64 {
	s = strings.TrimSpace(s)
	match := leadingNumber.FindString(s)
	if match == "" {
		return nil
	}

	rest := strings.TrimSpace(strings.TrimPrefix(s, match))
	rest = strings.TrimPrefix(rest, "%")
	if strings.TrimSpace(rest) != "" {
		return nil // sayının yanında açıklama var; metni bozmayalım
	}

	v, err := strconv.ParseFloat(strings.Replace(match, ",", ".", 1), 64)
	if err != nil {
		return nil
	}
	return &v
}
