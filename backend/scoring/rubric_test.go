package scoring

import (
	"strings"
	"testing"
)

// v1Rules, 000005_regulatory_scoring göçündeki rubrik v1'in aynısıdır.
// Testler veritabanına bağlanmaz: sınıflandırma saf mantıktır, kural
// puanlarının göçle tutarlılığını api paketindeki TestRubricV1MatchesMigration korur.
func v1Rules() []Rule {
	return []Rule{
		{RuleAnnexIIBanned, 10, "AB'de kozmetik ürünlerde kullanımı yasak (Ek II)"},
		{RuleDeclarableAllergen, 7, "Etikette beyanı zorunlu tutulan temas alerjeni (Ek III)"},
		{RuleAnnexIIIRestricted, 5, "Belirli bir oranın üstünde güvenli kabul edilmiyor (Ek III)"},
		{RuleAnnexVPreservative, 4, "Koruyucu olarak koşullu izinli (Ek V)"},
		{RuleAnnexVIUVFilter, 4, "UV filtresi olarak koşullu izinli (Ek VI)"},
		{RuleAnnexIVColorant, 3, "Renklendirici olarak izinli (Ek IV)"},
		{RuleUnrestricted, 2, "Eklerde kısıtlama kaydı yok"},
		{RuleSCCSAdverse, 2, "SCCS olumsuz görüş bildirdi"},
	}
}

func testRubric(t *testing.T) *Rubric {
	t.Helper()

	r, err := NewRubric(CurrentVersion, v1Rules())
	if err != nil {
		t.Fatalf("rubrik kurulamadı: %v", err)
	}
	return r
}

func TestApplyScoresByAnnex(t *testing.T) {
	r := testRubric(t)

	cases := []struct {
		name  string
		facts Facts
		want  int
		rule  string
	}{
		{"Ek II yasaklı", Facts{Annex: "II"}, 10, RuleAnnexIIBanned},
		{"bildirimli alerjen", Facts{Annex: "III", DeclarableAllergen: true}, 7, RuleDeclarableAllergen},
		{"Ek III kısıtlı", Facts{Annex: "III"}, 5, RuleAnnexIIIRestricted},
		{"Ek V koruyucu", Facts{Annex: "V"}, 4, RuleAnnexVPreservative},
		{"Ek VI UV filtresi", Facts{Annex: "VI"}, 4, RuleAnnexVIUVFilter},
		{"Ek IV renklendirici", Facts{Annex: "IV"}, 3, RuleAnnexIVColorant},
		{"kayıt var, kısıt yok", Facts{}, 2, RuleUnrestricted},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := r.Apply(tc.facts)
			if got.Value != tc.want {
				t.Errorf("puan %d, beklenen %d", got.Value, tc.want)
			}
			if len(got.Rules) != 1 || got.Rules[0].Key != tc.rule {
				t.Errorf("uygulanan kural %+v, beklenen %s", got.Rules, tc.rule)
			}
			if got.Version != CurrentVersion {
				t.Errorf("sürüm %d, beklenen %d", got.Version, CurrentVersion)
			}
		})
	}
}

// Bildirimli alerjen olmak, Ek III'ün genel kısıt kuralından ağır basmalı:
// kullanıcının cildinin tepki verdiği şey tam olarak budur.
func TestDeclarableAllergenOutranksAnnexIII(t *testing.T) {
	r := testRubric(t)

	plain := r.Apply(Facts{Annex: "III"})
	declarable := r.Apply(Facts{Annex: "III", DeclarableAllergen: true})

	if declarable.Value <= plain.Value {
		t.Errorf("bildirimli alerjen %d puan aldı, Ek III kısıtı %d — daha yüksek olmalıydı",
			declarable.Value, plain.Value)
	}
}

func TestSCCSAdverseIsAdditiveAndCapped(t *testing.T) {
	r := testRubric(t)

	got := r.Apply(Facts{Annex: "V", SCCSAdverse: true})
	if got.Value != 6 { // 4 + 2
		t.Errorf("puan %d, beklenen 6", got.Value)
	}
	if len(got.Rules) != 2 || got.Rules[1].Key != RuleSCCSAdverse {
		t.Errorf("değiştirici kural açıklamaya eklenmedi: %+v", got.Rules)
	}

	// Ek II zaten tavanda; değiştirici ölçeğin dışına taşıramaz.
	capped := r.Apply(Facts{Annex: "II", SCCSAdverse: true})
	if capped.Value != maxScore {
		t.Errorf("puan %d, beklenen %d (tavan)", capped.Value, maxScore)
	}
}

// Her puan bir mevzuat atfıyla birlikte gelmeli: atıfsız puan, elle atanmış
// puandan farksızdır. Faz 3'ün tüm gerekçesi bu.
func TestScoreAlwaysCitesSource(t *testing.T) {
	r := testRubric(t)

	const url = "https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223"
	got := r.Apply(Facts{Annex: "III", AnnexEntry: "84", DeclarableAllergen: true, SourceURL: url})

	if len(got.Sources) == 0 {
		t.Fatal("puan kaynaksız döndü")
	}
	src := got.Sources[0]
	for _, want := range []string{"1223/2009", "Ek III", "giriş 84", url} {
		if !strings.Contains(src, want) {
			t.Errorf("atıfta %q yok: %s", want, src)
		}
	}
}

func TestSCCSOpinionAppearsInSources(t *testing.T) {
	r := testRubric(t)

	got := r.Apply(Facts{Annex: "V", SCCSAdverse: true, SCCSOpinion: "SCCS/1670/24"})
	if len(got.Sources) != 2 {
		t.Fatalf("kaynak sayısı %d, beklenen 2: %v", len(got.Sources), got.Sources)
	}
	if !strings.Contains(got.Sources[1], "SCCS/1670/24") {
		t.Errorf("SCCS görüşü kaynaklarda yok: %s", got.Sources[1])
	}
}

func TestAnnexNormalization(t *testing.T) {
	r := testRubric(t)

	for _, annex := range []string{"iii", " III ", "Ek III", "Annex III"} {
		if got := r.Apply(Facts{Annex: annex}); got.Value != 5 {
			t.Errorf("%q için puan %d, beklenen 5", annex, got.Value)
		}
	}
}

// Eksik kuralla puanlamaya başlamak, sessizce yanlış puan üretmek olurdu.
func TestNewRubricRejectsIncompleteRuleSet(t *testing.T) {
	partial := v1Rules()[:3]

	_, err := NewRubric(1, partial)
	if err == nil {
		t.Fatal("eksik rubrik kabul edildi")
	}
	if !strings.Contains(err.Error(), RuleUnrestricted) {
		t.Errorf("hata eksik kuralı adlandırmıyor: %v", err)
	}
}
