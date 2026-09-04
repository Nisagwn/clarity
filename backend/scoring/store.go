package scoring

import (
	"context"
	"database/sql"
	"fmt"
)

// Querier, *sql.DB ve *sql.Tx'in ortak yüzeyi; rubrik hem istek işlerken hem
// göç/komut içinde bir işlemin ortasında okunabilsin diye.
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// LoadRubric, verilen sürümün kurallarını veritabanından okur.
//
// Kurallar kodda sabit değil: puanın gerekçesi (rationale) kullanıcıya
// gösterilen metindir ve mevzuat değiştiğinde yeni bir sürüm satırı eklenerek
// güncellenir; eski puanlar hangi sürümle hesaplandıklarını taşımaya devam eder.
func LoadRubric(ctx context.Context, q Querier, version int) (*Rubric, error) {
	const query = `SELECT rule_key, score, rationale FROM scoring_rule WHERE version = $1`

	rows, err := q.QueryContext(ctx, query, version)
	if err != nil {
		return nil, fmt.Errorf("rubrik v%d okunamadı: %w", version, err)
	}
	defer rows.Close()

	rules := []Rule{}
	for rows.Next() {
		var r Rule
		if err := rows.Scan(&r.Key, &r.Score, &r.Rationale); err != nil {
			return nil, fmt.Errorf("rubrik v%d okunamadı: %w", version, err)
		}
		rules = append(rules, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rubrik v%d okunamadı: %w", version, err)
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("rubrik v%d veritabanında yok", version)
	}

	return NewRubric(version, rules)
}
