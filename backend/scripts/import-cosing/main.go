// import-cosing komutu, Avrupa Komisyonu'nun CosIng dışa aktarımını
// ingredient_regulatory tablosuna alır.
//
// CosIng, kozmetik içeriklerin resmî AB veritabanıdır: INCI adları, CAS/EC
// numaraları, AB Tüzüğü 1223/2009 Eklerindeki yerleri ve SCCS görüş
// referansları. Dosyalar kayıt gerektirmeden indirilir:
//
//	https://ec.europa.eu/growth/tools-databases/cosing/
//
// Kullanım:
//
//	go run ./scripts/import-cosing -file COSING_Annex_III_v2.csv -annex III
//	go run ./scripts/import-cosing -file fragrance_allergens.csv -annex III -declarable
//	go run ./scripts/import-cosing -file COSING_Annex_II_v2.csv -annex II -dry-run
//
// Sonrasında puanların yeniden türetilmesi gerekir:
//
//	go run ./cmd/score
//
// İki kural:
//
//   - Eşleşmeyen satır SESSİZCE ATLANMAZ. Sayısı özette görünür, -report ile
//     tamamı CSV olarak yazılır. Sessiz atlama, kataloğun mevzuata bağlandığı
//     yanılgısını üretir.
//   - sccs_adverse otomatik doldurulmaz. CosIng, görüşün olumlu mu olumsuz mu
//     olduğunu makine okunur biçimde vermiyor; "görüş var" ile "olumsuz görüş
//     var" arasındaki farkı tahmin etmek puanı uydurmak olurdu. Görüşü olan
//     içerikler özette elle inceleme için listelenir.
package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// defaultSourceURL, mevzuatın kendisine işaret eder; her puanın yanında
// gösterilen atıf budur.
const defaultSourceURL = "https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223"

// bom, bazı CosIng dosyalarının başındaki UTF-8 işareti.
const bom = "\uFEFF"

// columnAliases, CosIng dışa aktarımlarında karşılaşılan sütun adları.
// Dosya sürümleri arasında başlıklar değişiyor; INCI sütunu tanınmazsa komut
// tahmin etmek yerine durur ve dosyadaki başlıkları yazar.
var columnAliases = map[string][]string{
	"inci":        {"inci name", "inci_name", "inci", "ingredient", "substance name", "chemical name", "name"},
	"cas":         {"cas no", "cas no.", "cas number", "cas", "cas registry number"},
	"ec":          {"ec no", "ec no.", "ec number", "ec", "einecs/elincs no", "einecs elincs no"},
	"entry":       {"reference number", "reference no", "ref no", "ref. no", "annex ref", "regulation reference"},
	"restriction": {"wording of conditions of use and warnings", "conditions of use and warnings", "restrictions", "restriction", "other"},
	"max_conc":    {"maximum concentration in ready for use preparation", "maximum concentration", "max concentration"},
	"sccs":        {"sccs opinions", "sccs opinion", "scientific committee opinions", "opinions"},
	"annex":       {"annex", "annex number", "annex no"},
}

type options struct {
	file       string
	annex      string
	declarable bool
	sourceURL  string
	report     string
	dryRun     bool
}

func main() {
	log.SetFlags(0)

	var opt options
	flag.StringVar(&opt.file, "file", "", "CosIng CSV dosyası (zorunlu)")
	flag.StringVar(&opt.annex, "annex", "", "dosyanın ait olduğu Ek: II, III, IV, V, VI")
	flag.BoolVar(&opt.declarable, "declarable", false, "satırları bildirimli koku alerjeni olarak işaretle")
	flag.StringVar(&opt.sourceURL, "source-url", defaultSourceURL, "puanların yanında gösterilecek atıf")
	flag.StringVar(&opt.report, "report", "", "eşleşmeyen satırların yazılacağı CSV")
	flag.BoolVar(&opt.dryRun, "dry-run", false, "veritabanına yazma, yalnızca özeti göster")
	flag.Parse()

	if opt.file == "" {
		flag.Usage()
		log.Fatal("\n-file zorunludur")
	}
	if opt.annex != "" {
		annex := normalizeAnnex(opt.annex)
		if annex == "" {
			log.Fatalf("bilinmeyen Ek: %q (II, III, IV, V, VI)", opt.annex)
		}
		opt.annex = annex
	}

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

	if err := run(context.Background(), db, opt); err != nil {
		log.Fatalf("içe aktarım başarısız: %v", err)
	}
}

// catalogEntry, katalogdaki bir içeriğin eşleştirmeye giren hâli.
type catalogEntry struct {
	id   int
	name string
}

// summary, içe aktarımın sonucu.
type summary struct {
	rows        int
	matched     int
	created     int
	updated     int
	unmatched   [][]string // INCI adı + neden
	ambiguous   []string
	needsReview []string // SCCS görüşü olan, elle bakılması gereken içerikler
}

func run(ctx context.Context, db *sql.DB, opt options) error {
	rows, header, err := readCSV(opt.file)
	if err != nil {
		return err
	}

	cols, err := mapColumns(header)
	if err != nil {
		return err
	}

	catalog, ambiguous, err := loadCatalog(ctx, db)
	if err != nil {
		return err
	}

	existing, err := loadRegulatory(ctx, db)
	if err != nil {
		return err
	}

	sum := summary{rows: len(rows), ambiguous: ambiguous}
	merged := map[int]record{}

	for _, row := range rows {
		inci := strings.TrimSpace(field(row, cols, "inci"))
		if inci == "" {
			continue
		}

		entry, ok := catalog[normalizeINCI(inci)]
		if !ok {
			sum.unmatched = append(sum.unmatched, []string{inci, "katalogda karşılığı yok"})
			continue
		}

		annex := opt.annex
		if fromRow := normalizeAnnex(field(row, cols, "annex")); fromRow != "" {
			annex = fromRow
		}

		incoming := record{
			ingredientID:     entry.id,
			cas:              strings.TrimSpace(field(row, cols, "cas")),
			ec:               strings.TrimSpace(field(row, cols, "ec")),
			annex:            annex,
			annexEntry:       strings.TrimSpace(field(row, cols, "entry")),
			restriction:      strings.TrimSpace(field(row, cols, "restriction")),
			maxConcentration: parseConcentration(field(row, cols, "max_conc")),
			declarable:       opt.declarable,
			sccsOpinion:      strings.TrimSpace(field(row, cols, "sccs")),
			sourceURL:        opt.sourceURL,
		}

		// Aynı madde dosyada birden fazla satırda geçebilir (farklı kullanım
		// koşulları); hepsi tek kayıtta birleşir.
		base, seen := merged[entry.id]
		if !seen {
			base = existing[entry.id]
		}
		merged[entry.id] = merge(base, incoming)

		sum.matched++
		if incoming.sccsOpinion != "" {
			sum.needsReview = append(sum.needsReview, entry.name)
		}
	}

	for id := range merged {
		if _, ok := existing[id]; ok {
			sum.updated++
		} else {
			sum.created++
		}
	}

	if !opt.dryRun {
		if err := writeRecords(ctx, db, merged); err != nil {
			return err
		}
	}

	if opt.report != "" {
		if err := writeReport(opt.report, sum.unmatched); err != nil {
			return err
		}
	}

	unlinked, err := ingredientsWithoutRegulatory(ctx, db)
	if err != nil {
		return err
	}

	printSummary(opt, sum, unlinked)
	return nil
}

// readCSV, dosyayı okur ve başlık satırını ayırır.
//
// CosIng dışa aktarımları hem virgül hem noktalı virgül ayraçlı gelebiliyor
// ve başında BOM taşıyabiliyor; ikisi de sessizce ele alınır.
func readCSV(path string) ([][]string, []string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	text := strings.TrimPrefix(string(raw), bom)
	firstLine, _, _ := strings.Cut(text, "\n")

	r := csv.NewReader(strings.NewReader(text))
	r.FieldsPerRecord = -1 // satır uzunlukları dosya içinde değişebiliyor
	r.LazyQuotes = true
	if strings.Count(firstLine, ";") > strings.Count(firstLine, ",") {
		r.Comma = ';'
	}

	header, err := r.Read()
	if errors.Is(err, io.EOF) {
		return nil, nil, fmt.Errorf("%s boş", path)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("%s okunamadı: %w", path, err)
	}

	rows, err := r.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("%s okunamadı: %w", path, err)
	}
	return rows, header, nil
}

// mapColumns, başlıkları bildiğimiz alanlara bağlar.
func mapColumns(header []string) (map[string]int, error) {
	index := map[string]int{}
	for i, h := range header {
		index[strings.ToLower(strings.TrimSpace(strings.TrimPrefix(h, bom)))] = i
	}

	cols := map[string]int{}
	for name, aliases := range columnAliases {
		for _, alias := range aliases {
			if i, ok := index[alias]; ok {
				cols[name] = i
				break
			}
		}
	}

	if _, ok := cols["inci"]; !ok {
		return nil, fmt.Errorf(
			"INCI adı sütunu bulunamadı.\n  Dosyadaki başlıklar: %s\n  Beklenen adlardan biri: %s",
			strings.Join(header, ", "), strings.Join(columnAliases["inci"], ", "))
	}
	return cols, nil
}

func field(row []string, cols map[string]int, name string) string {
	i, ok := cols[name]
	if !ok || i >= len(row) {
		return ""
	}
	return row[i]
}

// loadCatalog, içerikleri normalleştirilmiş INCI adına göre indeksler.
//
// Aynı ada çözülen birden fazla içerik varsa hiçbiri eşleştirilmez ve durum
// raporlanır: yanlış içeriğe mevzuat bağlamak, hiç bağlamamaktan kötüdür.
func loadCatalog(ctx context.Context, db *sql.DB) (map[string]catalogEntry, []string, error) {
	const query = `SELECT id, name, COALESCE(inci_name, '') FROM ingredients ORDER BY id`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	catalog := map[string]catalogEntry{}
	collisions := map[string][]string{}

	for rows.Next() {
		var (
			id         int
			name, inci string
		)
		if err := rows.Scan(&id, &name, &inci); err != nil {
			return nil, nil, err
		}

		key := normalizeINCI(inci)
		if key == "" {
			key = normalizeINCI(name)
		}
		if key == "" {
			continue
		}

		if prev, exists := catalog[key]; exists {
			collisions[key] = append(collisions[key], prev.name, name)
			continue
		}
		catalog[key] = catalogEntry{id: id, name: name}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	ambiguous := []string{}
	for key, names := range collisions {
		delete(catalog, key)
		ambiguous = append(ambiguous, strings.Join(unique(names), " / "))
	}
	sort.Strings(ambiguous)

	return catalog, ambiguous, nil
}

func loadRegulatory(ctx context.Context, db *sql.DB) (map[int]record, error) {
	const query = `
		SELECT ingredient_id, COALESCE(cas_number, ''), COALESCE(ec_number, ''),
		       COALESCE(annex, ''), COALESCE(annex_entry, ''),
		       COALESCE(restriction, ''), max_concentration,
		       declarable_allergen, COALESCE(sccs_opinion, ''), sccs_adverse,
		       source_url
		FROM ingredient_regulatory`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int]record{}
	for rows.Next() {
		var r record
		if err := rows.Scan(&r.ingredientID, &r.cas, &r.ec, &r.annex, &r.annexEntry,
			&r.restriction, &r.maxConcentration, &r.declarable, &r.sccsOpinion,
			&r.sccsAdverse, &r.sourceURL); err != nil {
			return nil, err
		}
		out[r.ingredientID] = r
	}
	return out, rows.Err()
}

func writeRecords(ctx context.Context, db *sql.DB, records map[int]record) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const query = `
		INSERT INTO ingredient_regulatory
		    (ingredient_id, cas_number, ec_number, annex, annex_entry, restriction,
		     max_concentration, declarable_allergen, sccs_opinion, sccs_adverse,
		     source_url, fetched_at)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''),
		        NULLIF($6, ''), $7, $8, NULLIF($9, ''), $10, $11, CURRENT_TIMESTAMP)
		ON CONFLICT (ingredient_id) DO UPDATE SET
		    cas_number          = EXCLUDED.cas_number,
		    ec_number           = EXCLUDED.ec_number,
		    annex               = EXCLUDED.annex,
		    annex_entry         = EXCLUDED.annex_entry,
		    restriction         = EXCLUDED.restriction,
		    max_concentration   = EXCLUDED.max_concentration,
		    declarable_allergen = EXCLUDED.declarable_allergen,
		    sccs_opinion        = EXCLUDED.sccs_opinion,
		    sccs_adverse        = EXCLUDED.sccs_adverse,
		    source_url          = EXCLUDED.source_url,
		    fetched_at          = CURRENT_TIMESTAMP`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	ids := make([]int, 0, len(records))
	for id := range records {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		r := records[id]
		if _, err := stmt.ExecContext(ctx, r.ingredientID, r.cas, r.ec, r.annex,
			r.annexEntry, r.restriction, r.maxConcentration, r.declarable,
			r.sccsOpinion, r.sccsAdverse, r.sourceURL); err != nil {
			return fmt.Errorf("içerik %d yazılamadı: %w", id, err)
		}
	}

	return tx.Commit()
}

// ingredientsWithoutRegulatory, mevzuat kaydı hâlâ olmayan içeriklerin adları.
func ingredientsWithoutRegulatory(ctx context.Context, db *sql.DB) ([]string, error) {
	const query = `
		SELECT i.name
		FROM ingredients i
		LEFT JOIN ingredient_regulatory r ON r.ingredient_id = i.id
		WHERE r.ingredient_id IS NULL
		ORDER BY i.name`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

func writeReport(path string, unmatched [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"inci_name", "reason"}); err != nil {
		return err
	}
	return w.WriteAll(unmatched)
}

// printSummary, ne aktarıldığını ve neyin eksik kaldığını yazar.
func printSummary(opt options, sum summary, unlinked []string) {
	mode := ""
	if opt.dryRun {
		mode = " (deneme — hiçbir şey yazılmadı)"
	}

	log.Printf("%s%s", opt.file, mode)
	log.Printf("  okunan satır      : %d", sum.rows)
	log.Printf("  eşleşen           : %d (%d yeni kayıt, %d güncelleme)", sum.matched, sum.created, sum.updated)
	log.Printf("  eşleşmeyen        : %d", len(sum.unmatched))

	if len(sum.unmatched) > 0 && opt.report == "" {
		log.Println("      tamamını görmek için: -report unmatched.csv")
	}
	if opt.report != "" {
		log.Printf("      rapor: %s", opt.report)
	}

	if len(sum.ambiguous) > 0 {
		log.Printf("  belirsiz katalog kaydı: %d — aynı INCI adına çözülüyor, eşleştirilmedi", len(sum.ambiguous))
		for _, a := range sum.ambiguous {
			log.Printf("      %s", a)
		}
	}

	if names := unique(sum.needsReview); len(names) > 0 {
		log.Printf("  SCCS görüşü olan  : %d — görüşün olumsuz olup olmadığı ELLE işaretlenir", len(names))
		log.Printf("      %s", strings.Join(preview(names, 10), ", "))
		log.Println("      UPDATE ingredient_regulatory SET sccs_adverse = TRUE WHERE ingredient_id = ...")
	}

	log.Printf("  mevzuat kaydı olmayan içerik: %d", len(unlinked))
	if len(unlinked) > 0 {
		log.Printf("      %s", strings.Join(preview(unlinked, 10), ", "))
	}
	log.Println("  puanları yenilemek için: go run ./cmd/score")
}

func unique(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func preview(values []string, n int) []string {
	if len(values) <= n {
		return values
	}
	return append(append([]string{}, values[:n]...), fmt.Sprintf("... ve %d tane daha", len(values)-n))
}
