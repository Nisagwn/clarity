// scoring paketi, bir içeriğin endişe puanını mevzuat verisinden TÜRETİR.
//
// İş bölümü bilinçli:
//
//   - Hangi koşula kaç puan verildiği veritabanındaki scoring_rule
//     tablosunda ve versiyonludur. Rubrik değişince eski puanların hangi
//     kurala göre verildiği izlenebilir kalır.
//   - Hangi koşulun geçerli olduğuna karar veren sınıflandırma burada,
//     tek yerde durur; hem API hem cmd/score aynı kodu çalıştırır.
//
// Puanın kaynağı AB Tüzüğü 1223/2009 Ekleri ve SCCS görüşleridir. EWG puanı
// ne kopyalanır ne de referans alınır: lisansı buna izin vermiyor ve
// dayanağını gösteremeyen bir puan kullanıcıya da bir şey anlatmıyor.
package scoring

import (
	"fmt"
	"strings"
)

// CurrentVersion, yeni puanların yazıldığı rubrik sürümü.
const CurrentVersion = 1

// maxScore, ölçeğin üst sınırı. Değiştirici kural puanı bunun üstüne çıkaramaz.
const maxScore = 10

// Rubrikteki kural anahtarları; 000005_regulatory_scoring göçünde
// scoring_rule tablosuna yazılan satırlarla birebir eşleşir.
const (
	RuleAnnexIIBanned      = "annex_ii_banned"
	RuleDeclarableAllergen = "declarable_allergen"
	RuleAnnexIIIRestricted = "annex_iii_restricted"
	RuleAnnexVPreservative = "annex_v_preservative"
	RuleAnnexVIUVFilter    = "annex_vi_uv_filter"
	RuleAnnexIVColorant    = "annex_iv_colorant"
	RuleUnrestricted       = "unrestricted"
	RuleSCCSAdverse        = "sccs_adverse_modifier"
)

// requiredRules, bir rubriğin eksiksiz sayılması için gereken anahtarlar.
// Eksik kuralla puanlamaya başlamak, sessizce yanlış puan üretmek olurdu.
var requiredRules = []string{
	RuleAnnexIIBanned,
	RuleDeclarableAllergen,
	RuleAnnexIIIRestricted,
	RuleAnnexVPreservative,
	RuleAnnexVIUVFilter,
	RuleAnnexIVColorant,
	RuleUnrestricted,
	RuleSCCSAdverse,
}

// citationBase, her atfın önüne gelen mevzuat adı.
const citationBase = "AB Tüzüğü 1223/2009"

// annexLabels, Ek numarasının kullanıcıya gösterilen karşılığı.
var annexLabels = map[string]string{
	"II":  "Ek II — kozmetikte kullanımı yasak maddeler",
	"III": "Ek III — kısıtlı maddeler",
	"IV":  "Ek IV — izin verilen renklendiriciler",
	"V":   "Ek V — izin verilen koruyucular",
	"VI":  "Ek VI — izin verilen UV filtreleri",
}

// Facts, puanın dayandığı mevzuat olguları: ingredient_regulatory satırının
// puanlamayı ilgilendiren alanları.
type Facts struct {
	Annex              string
	AnnexEntry         string
	DeclarableAllergen bool
	SCCSAdverse        bool
	SCCSOpinion        string
	SourceURL          string
}

// Rule, rubrikteki tek bir kural.
type Rule struct {
	Key       string `json:"key"`
	Score     int    `json:"score"`
	Rationale string `json:"rationale"`
}

// Score, türetilmiş puan ve onu üreten kurallar. Rules alanı arayüzdeki
// "neden bu puan?" açıklamasının tamamıdır: puan hiçbir zaman gerekçesiz
// gösterilmez.
type Score struct {
	Version int      `json:"version"`
	Value   int      `json:"value"`
	Rules   []Rule   `json:"rules"`
	Sources []string `json:"sources"`
}

// Rubric, tek bir sürümün kural kümesi.
type Rubric struct {
	version int
	rules   map[string]Rule
}

// NewRubric, kuralları doğrulayıp bir rubrik oluşturur. Gerekli kurallardan
// biri eksikse hata döner.
func NewRubric(version int, rules []Rule) (*Rubric, error) {
	if version <= 0 {
		return nil, fmt.Errorf("rubrik sürümü pozitif olmalı, %d verildi", version)
	}

	byKey := make(map[string]Rule, len(rules))
	for _, r := range rules {
		byKey[r.Key] = r
	}

	missing := []string{}
	for _, key := range requiredRules {
		if _, ok := byKey[key]; !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("rubrik v%d eksik: %s kuralları tanımlı değil",
			version, strings.Join(missing, ", "))
	}

	return &Rubric{version: version, rules: byKey}, nil
}

// Version, rubriğin sürümünü döndürür.
func (r *Rubric) Version() int { return r.version }

// Apply, mevzuat olgularından puanı türetir.
//
// Sıralama en kısıtlayıcıdan gevşeğe doğrudur: bir madde hem Ek III'te hem
// Ek V'te geçebiliyor; o zaman geçerli olan daha ağır kısıtlamadır.
func (r *Rubric) Apply(f Facts) Score {
	f.Annex = normalizeAnnex(f.Annex)

	base := r.rules[baseRuleKey(f)]
	value := base.Score
	rules := []Rule{base}

	// SCCS olumsuz görüşü bir değiştirici: temel kuralın yerine geçmez,
	// üstüne eklenir.
	if f.SCCSAdverse {
		modifier := r.rules[RuleSCCSAdverse]
		value += modifier.Score
		rules = append(rules, modifier)
	}
	if value > maxScore {
		value = maxScore
	}

	return Score{
		Version: r.version,
		Value:   value,
		Rules:   rules,
		Sources: sources(f),
	}
}

// baseRuleKey, olgulara uyan temel kuralı seçer.
func baseRuleKey(f Facts) string {
	switch {
	case f.Annex == "II":
		return RuleAnnexIIBanned
	case f.DeclarableAllergen:
		// Bildirimli koku alerjeni olmak Ek III kısıtından ağır basar:
		// kullanıcının tepki verdiği şey tam olarak budur.
		return RuleDeclarableAllergen
	case f.Annex == "III":
		return RuleAnnexIIIRestricted
	case f.Annex == "V":
		return RuleAnnexVPreservative
	case f.Annex == "VI":
		return RuleAnnexVIUVFilter
	case f.Annex == "IV":
		return RuleAnnexIVColorant
	default:
		return RuleUnrestricted
	}
}

// sources, puanın mevzuat atıflarını üretir. Her puanın yanında bunlar
// gösterilir; atıfsız puan, elle atanmış puandan farksız olurdu.
func sources(f Facts) []string {
	out := []string{withURL(annexCitation(f), f.SourceURL)}

	if f.SCCSAdverse {
		opinion := strings.TrimSpace(f.SCCSOpinion)
		if opinion == "" {
			opinion = "olumsuz görüş"
		}
		out = append(out, withURL("SCCS görüşü: "+opinion, f.SourceURL))
	}
	return out
}

// annexCitation, Ek numarasını okunabilir bir atfa çevirir.
func annexCitation(f Facts) string {
	label, known := annexLabels[f.Annex]
	if !known {
		return citationBase + " Ekleri: kayıtlı bir kısıtlama yok"
	}

	citation := citationBase + " " + label
	if entry := strings.TrimSpace(f.AnnexEntry); entry != "" {
		citation += ", giriş " + entry
	}
	return citation
}

func withURL(citation, url string) string {
	if url = strings.TrimSpace(url); url == "" {
		return citation
	}
	return citation + " — " + url
}

// normalizeAnnex, "ek iii", "III " gibi yazımları tek biçime getirir.
func normalizeAnnex(annex string) string {
	annex = strings.ToUpper(strings.TrimSpace(annex))
	annex = strings.TrimPrefix(annex, "EK ")
	annex = strings.TrimPrefix(annex, "ANNEX ")
	return strings.TrimSpace(annex)
}
