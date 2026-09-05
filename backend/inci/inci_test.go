package inci

import (
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeMatchesSameSubstance(t *testing.T) {
	same := [][2]string{
		{"Titanium Dioxide (CI 77891)", "Titanium dioxide"},
		{"Parfum (Fragrance)", "PARFUM"},
		{"Butyrospermum Parkii Butter", "butyrospermum parkii  butter"},
		{"Cetearyl Alcohol", "Cetearyl alcohol."},
		{"PEG-10 Dimethicone", "peg 10 dimethicone"},
	}
	for _, pair := range same {
		if Normalize(pair[0]) != Normalize(pair[1]) {
			t.Errorf("%q ve %q aynı maddeye çözülmedi: %q / %q",
				pair[0], pair[1], Normalize(pair[0]), Normalize(pair[1]))
		}
	}
}

// Eşleştirme normalleştirmeden sonra TAM eşleşmedir; farklı maddeler
// birbirine karışmamalı.
func TestNormalizeKeepsDistinctSubstancesApart(t *testing.T) {
	different := [][2]string{
		{"Alcohol", "Alcohol Denat."},
		{"Benzyl Alcohol", "Benzyl Salicylate"},
		{"Iron Oxides", "Zinc Oxide"},
		{"Sodium Laureth Sulfate", "Sodium Lauryl Sulfate"},
	}
	for _, pair := range different {
		if Normalize(pair[0]) == Normalize(pair[1]) {
			t.Errorf("%q ve %q aynı maddeye çözüldü: %q", pair[0], pair[1], Normalize(pair[0]))
		}
	}
}

func TestParseListKeepsLabelOrder(t *testing.T) {
	got := ParseList("Aqua, Glycerin, Niacinamide, Tocopherol")
	want := []string{"Aqua", "Glycerin", "Niacinamide", "Tocopherol"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("liste %v, beklenen %v", got, want)
	}
}

// "May contain" bölümündeki renklendiriciler üründe OLMAYABİLİR. Varmış gibi
// kaydetmek, kullanıcıya taşımadığı bir madde için uyarı vermek olurdu.
func TestParseListDropsMayContainSection(t *testing.T) {
	const label = "DIMETHICONE, TOCOPHEROL, LINALOOL. May contain (+/-): CI 77891, CI 77491, CI 15850"

	got := ParseList(label)
	want := []string{"DIMETHICONE", "TOCOPHEROL", "LINALOOL"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("liste %v, beklenen %v", got, want)
	}

	for _, variant := range []string{
		"Aqua, Glycerin [+/- CI 77891, CI 77492]",
		"Aqua, Glycerin ± CI 77891",
		"Aqua, Glycerin, peut contenir: CI 77891",
	} {
		got := ParseList(variant)
		if len(got) != 2 {
			t.Errorf("%q → %v; koşullu bölüm ayıklanmadı", variant, got)
		}
	}
}

func TestParseListStripsLabelHeading(t *testing.T) {
	for _, label := range []string{
		"INGREDIENTS: Aqua, Glycerin",
		"Ingrédients : Aqua, Glycerin",
		"İçindekiler: Aqua, Glycerin",
	} {
		got := ParseList(label)
		if len(got) == 0 || Normalize(got[0]) != "aqua" {
			t.Errorf("%q → %v; başlık ayıklanmadı", label, got)
		}
	}
}

func TestParseListDeduplicatesAndTrims(t *testing.T) {
	got := ParseList("  Aqua*, aqua , GLYCERIN.,  ,  Glycerin (vegetable) ")
	want := []string{"Aqua", "GLYCERIN"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("liste %v, beklenen %v", got, want)
	}
}

// Ayraçsız yazılmış bir paragraf içerik adı değildir; öyleymiş gibi
// kaydedilirse katalog çöple dolar.
func TestParseListRejectsOverlongTokens(t *testing.T) {
	junk := strings.Repeat("bu bir cümle ", 20)

	if got := ParseList("Aqua, " + junk); len(got) != 1 {
		t.Errorf("liste %v, yalnızca Aqua beklenirdi", got)
	}
}

func TestParseListEmpty(t *testing.T) {
	for _, in := range []string{"", "   ", "May contain (+/-): CI 77891"} {
		if got := ParseList(in); len(got) != 0 {
			t.Errorf("ParseList(%q) = %v, boş beklenirdi", in, got)
		}
	}
}

// Kaynak veride etiket ayrıştırması her zaman temiz çıkmıyor. Bu artıkların
// "içerik" olarak kataloğa girmesi, kullanıcının gördüğü listeyi çöple
// doldurur.
func TestIsPlausibleNameRejectsParsingDebris(t *testing.T) {
	junk := []string{
		`! esté dermatologique`,
		`" 20% du total des 99% du total est`,
		`""ingredie entfied organic agriculture`,
		`"8RN01-3"`,
		`25% do plastico utilizado para fabricar este frasco e reciclado`,
		`&gt`,
		`-`,
		`1`,
		``,
	}
	for _, s := range junk {
		if IsPlausibleName(s) {
			t.Errorf("ayrıştırma artığı kabul edildi: %q", s)
		}
	}
}

// Süzgeç gerçek INCI adlarını elememeli; rakamla başlayanlar, eğik çizgili
// kopolimerler ve parantezli botanik adlar hepsi geçerli.
func TestIsPlausibleNameAcceptsRealINCINames(t *testing.T) {
	names := []string{
		"Aqua",
		"Butyrospermum Parkii (Shea) Butter",
		"1,2-Hexanediol",
		"PEG-10 Dimethicone",
		"CI 77891",
		"Sodium C14-16 Olefin Sulfonate",
		"Acrylates/Vinyl Isodecanoate Crosspolymer",
		"Parfum (Fragrance)",
	}
	for _, s := range names {
		if !IsPlausibleName(s) {
			t.Errorf("gerçek INCI adı elendi: %q", s)
		}
	}
}

// Etiket ayrıştırması bir cümleyi tek parça olarak verdiğinde liste onu
// içerik saymamalı.
func TestParseListDropsImplausibleTokens(t *testing.T) {
	got := ParseList(`Aqua, "8RN01-3", Glycerin, 25% du total, Tocopherol`)
	want := []string{"Aqua", "Glycerin", "Tocopherol"}

	if len(got) != len(want) {
		t.Fatalf("liste %v, beklenen %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("liste %v, beklenen %v", got, want)
			break
		}
	}
}
