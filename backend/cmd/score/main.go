// score komutu, içerik puanlarını mevzuat verisinden yeniden türetir.
//
// Kullanım:
//
//	go run ./cmd/score             puanları yazar
//	go run ./cmd/score -dry-run    yalnızca ne değişeceğini raporlar
//	go run ./cmd/score -version 2  belirli bir rubrik sürümüyle puanlar
//
// CosIng içe aktarımından sonra çalıştırılır:
//
//	go run ./scripts/import-cosing -file COSING_Annex_III_v2.csv -annex III
//	go run ./cmd/score
//
// Puanın kendisi burada hesaplanmaz; mantık scoring paketindedir ve API ile
// aynı kodu paylaşır. Bu komut yalnızca onu tüm katalog üzerinde koşturur.
package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/nisa/beauty-ingredient/scoring"
)

func main() {
	log.SetFlags(0)

	dryRun := flag.Bool("dry-run", false, "veritabanına yazma, yalnızca özeti göster")
	version := flag.Int("version", scoring.CurrentVersion, "kullanılacak rubrik sürümü")
	flag.Parse()

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf(".env yüklenmedi (%v), ortam değişkenlerine düşülüyor", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL tanımlı değil")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("veritabanına bağlanılamadı: %v", err)
	}
	defer db.Close()

	sum, err := scoring.Recompute(context.Background(), db, *version, *dryRun)
	if err != nil {
		log.Fatalf("puanlama başarısız: %v", err)
	}

	report(sum, *version, *dryRun)
}

// report, ne yapıldığını ve NEYİN YAPILMADIĞINI yazar. Puansız kalan
// içeriklerin sayısı görünmezse katalog puanlanmış sanılır; oysa mevzuat
// verisi gelene kadar bu içeriklerin risk seviyesi bilinmiyor.
func report(sum scoring.Summary, version int, dryRun bool) {
	mode := ""
	if dryRun {
		mode = " (deneme — hiçbir şey yazılmadı)"
	}

	log.Printf("rubrik v%d%s", version, mode)
	log.Printf("  katalog       : %d içerik", sum.Total)
	log.Printf("  puanlandı     : %d (%d tanesinin puanı değişti)", sum.Scored, sum.Changed)
	log.Printf("  puanı silindi : %d (mevzuat kaydı yok)", sum.Cleared)
	log.Printf("  puansız kalan : %d", len(sum.Unscored))

	if len(sum.Unscored) == 0 {
		return
	}

	preview := sum.Unscored
	if len(preview) > 10 {
		preview = preview[:10]
	}
	log.Printf("      %s", strings.Join(preview, ", "))
	if len(sum.Unscored) > len(preview) {
		log.Printf("      ... ve %d tane daha", len(sum.Unscored)-len(preview))
	}
	log.Println(`  bunlar arayüzde "henüz puanlanmadı" görünür; CosIng içe aktarımı`)
	log.Println("  tamamlandığında puanlanacaklar (bkz. scripts/import-cosing)")
}
