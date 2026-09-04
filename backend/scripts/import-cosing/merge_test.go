package main

import "testing"

// Katalogdaki INCI adı ile CosIng'deki yazım birebir aynı olmuyor; parantez
// içi CI numaraları, noktalama ve büyük/küçük harf farkı eşleşmeyi
// engellememeli.
func TestNormalizeINCIMatchesCatalogSpelling(t *testing.T) {
	same := [][2]string{
		{"Titanium Dioxide (CI 77891)", "Titanium dioxide"},
		{"Parfum (Fragrance)", "PARFUM"},
		{"Butyrospermum Parkii Butter", "butyrospermum parkii  butter"},
		{"Cetearyl Alcohol", "Cetearyl alcohol."},
	}
	for _, pair := range same {
		if normalizeINCI(pair[0]) != normalizeINCI(pair[1]) {
			t.Errorf("%q ve %q aynı maddeye çözülmedi: %q / %q",
				pair[0], pair[1], normalizeINCI(pair[0]), normalizeINCI(pair[1]))
		}
	}
}

// Eşleştirme normalleştirmeden sonra TAM eşleşmedir. "Alcohol" ile
// "Alcohol Denat." farklı maddelerdir; alt dize karşılaştırması bunları
// birbirine karıştırıp yanlış içeriğe mevzuat bağlardı.
func TestNormalizeINCIKeepsDistinctSubstancesApart(t *testing.T) {
	different := [][2]string{
		{"Alcohol", "Alcohol Denat."},
		{"Benzyl Alcohol", "Benzyl Salicylate"},
		{"Iron Oxides", "Zinc Oxide"},
	}
	for _, pair := range different {
		if normalizeINCI(pair[0]) == normalizeINCI(pair[1]) {
			t.Errorf("%q ve %q aynı maddeye çözüldü: %q", pair[0], pair[1], normalizeINCI(pair[0]))
		}
	}
}

func TestNormalizeAnnex(t *testing.T) {
	cases := map[string]string{
		"III":       "III",
		"annex iii": "III",
		" Ek VI ":   "VI",
		"VII":       "", // böyle bir Ek puanlamada kullanılmıyor
		"":          "",
	}
	for in, want := range cases {
		if got := normalizeAnnex(in); got != want {
			t.Errorf("normalizeAnnex(%q) = %q, beklenen %q", in, got, want)
		}
	}
}

// Bir madde hem Ek III'te kısıtlı hem Ek V'te koruyucu olabilir. İkinci dosya
// birinciyi ezerse, daha ağır kısıtlama sessizce kaybolur.
func TestMergeKeepsStricterAnnex(t *testing.T) {
	restricted := record{ingredientID: 1, annex: "III", annexEntry: "84", restriction: "%0,01 üstünde bildirilir"}
	preservative := record{ingredientID: 1, annex: "V", annexEntry: "29", restriction: "en fazla %1"}

	got := merge(restricted, preservative)
	if got.annex != "III" {
		t.Errorf("Ek %q korundu, beklenen III", got.annex)
	}
	if got.annexEntry != "84" || got.restriction != "%0,01 üstünde bildirilir" {
		t.Errorf("Ek'e bağlı alanlar ayrıştı: %+v", got)
	}

	// Ters sırada da sonuç aynı olmalı: dosyaların işlenme sırası puanı
	// değiştirmemeli.
	if reverse := merge(preservative, restricted); reverse.annex != "III" || reverse.annexEntry != "84" {
		t.Errorf("sıraya bağlı sonuç: %+v", reverse)
	}
}

// Bildirimli alerjen bilgisi ayrı bir listeden geliyor; onu taşımayan bir
// dosya bu bilgiyi geri alamamalı.
func TestMergeOnlyAddsFlags(t *testing.T) {
	flagged := record{ingredientID: 1, annex: "III", declarable: true, sccsAdverse: true}
	plain := record{ingredientID: 1, annex: "III"}

	got := merge(flagged, plain)
	if !got.declarable {
		t.Error("declarable bayrağı silindi")
	}
	if !got.sccsAdverse {
		t.Error("sccsAdverse bayrağı silindi")
	}
}

func TestMergeFillsEmptyFields(t *testing.T) {
	existing := record{ingredientID: 1, annex: "V", annexEntry: "29", sourceURL: "https://eur-lex.europa.eu/x"}
	incoming := record{ingredientID: 1, annex: "V", cas: "122-99-6", ec: "204-589-7"}

	got := merge(existing, incoming)
	if got.cas != "122-99-6" || got.ec != "204-589-7" {
		t.Errorf("yeni alanlar yazılmadı: %+v", got)
	}
	if got.annexEntry != "29" || got.sourceURL != "https://eur-lex.europa.eu/x" {
		t.Errorf("var olan alanlar boşla ezildi: %+v", got)
	}
}

// Sayıya indirgenemeyen bir sınır, sınırsız görünmektense hiç gösterilmez;
// metnin tamamı restriction alanında durmaya devam eder.
func TestParseConcentration(t *testing.T) {
	cases := []struct {
		in   string
		want *float64
	}{
		{"1 %", ptr(1)},
		{"0,5 %", ptr(0.5)},
		{"25%", ptr(25)},
		{"  10 ", ptr(10)},
		{"5 % asit olarak", nil},
		{"yıkanan ürünlerde %0,001", nil},
		{"", nil},
	}

	for _, tc := range cases {
		got := parseConcentration(tc.in)
		switch {
		case tc.want == nil && got != nil:
			t.Errorf("parseConcentration(%q) = %v, beklenen nil", tc.in, *got)
		case tc.want != nil && got == nil:
			t.Errorf("parseConcentration(%q) = nil, beklenen %v", tc.in, *tc.want)
		case tc.want != nil && *got != *tc.want:
			t.Errorf("parseConcentration(%q) = %v, beklenen %v", tc.in, *got, *tc.want)
		}
	}
}

func ptr(v float64) *float64 { return &v }
