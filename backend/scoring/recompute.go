package scoring

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

// Summary, bir yeniden puanlamanın sonucu.
//
// Unscored alanı ayrı tutulur çünkü asıl bilgi orada: kaç içeriğin puanı
// olduğu değil, kaç içeriğin puanının HÂLÂ OLMADIĞI. Bu sayı görünmezse
// katalog puanlanmış sanılır.
type Summary struct {
	Total    int
	Scored   int
	Changed  int
	Cleared  int
	Unscored []string
}

// Recompute, katalogdaki tüm puanları mevzuat verisinden yeniden türetir.
//
// İdempotenttir ve elle atanmış puan bırakmaz: mevzuat kaydı olmayan içeriğin
// puanı NULL'a çekilir. Uydurulmuş bir puan, eksik bir puandan daha
// zararlıdır — kullanıcı ona güvenir ve neye dayandığını soramaz.
//
// dryRun true iken hiçbir şey yazılmaz; özet yine de ne olacağını söyler.
func Recompute(ctx context.Context, db *sql.DB, version int, dryRun bool) (Summary, error) {
	var sum Summary

	rubric, err := LoadRubric(ctx, db, version)
	if err != nil {
		return sum, err
	}

	rows, err := loadIngredients(ctx, db)
	if err != nil {
		return sum, err
	}
	sum.Total = len(rows)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return sum, err
	}
	defer tx.Rollback()

	for _, r := range rows {
		if r.facts == nil {
			// Mevzuat kaydı olmayan içerik puansız kalır. Bu, göç öncesinden
			// kalan elle atanmış puanları da temizler.
			sum.Unscored = append(sum.Unscored, r.name)
			if r.current != nil {
				sum.Cleared++
				if !dryRun {
					if err := clearScore(ctx, tx, r.id); err != nil {
						return sum, err
					}
				}
			}
			continue
		}

		score := rubric.Apply(*r.facts)
		sum.Scored++
		if r.current == nil || *r.current != score.Value {
			sum.Changed++
		}
		if !dryRun {
			if err := writeScore(ctx, tx, r.id, score); err != nil {
				return sum, err
			}
		}
	}

	if dryRun {
		return sum, nil
	}
	return sum, tx.Commit()
}

// ingredientRow, tek bir içeriğin puanlamaya giren hâli.
type ingredientRow struct {
	id      int
	name    string
	current *int
	facts   *Facts // mevzuat kaydı yoksa nil
}

// loadIngredients, katalogdaki her içeriği varsa mevzuat kaydıyla okur.
func loadIngredients(ctx context.Context, db *sql.DB) ([]ingredientRow, error) {
	const query = `
		SELECT i.id, i.name, i.concern_level,
		       r.ingredient_id IS NOT NULL,
		       COALESCE(r.annex, ''), COALESCE(r.annex_entry, ''),
		       COALESCE(r.declarable_allergen, FALSE),
		       COALESCE(r.sccs_adverse, FALSE), COALESCE(r.sccs_opinion, ''),
		       COALESCE(r.source_url, '')
		FROM ingredients i
		LEFT JOIN ingredient_regulatory r ON r.ingredient_id = i.id
		ORDER BY i.name`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ingredientRow{}
	for rows.Next() {
		var (
			r         ingredientRow
			hasRecord bool
			f         Facts
		)
		if err := rows.Scan(&r.id, &r.name, &r.current, &hasRecord,
			&f.Annex, &f.AnnexEntry, &f.DeclarableAllergen,
			&f.SCCSAdverse, &f.SCCSOpinion, &f.SourceURL); err != nil {
			return nil, err
		}
		if hasRecord {
			r.facts = &f
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func writeScore(ctx context.Context, tx *sql.Tx, id int, score Score) error {
	const query = `
		UPDATE ingredients
		SET concern_level = $2, score_version = $3, score_sources = $4
		WHERE id = $1`

	_, err := tx.ExecContext(ctx, query, id, score.Value, score.Version, pq.Array(score.Sources))
	return err
}

func clearScore(ctx context.Context, tx *sql.Tx, id int) error {
	const query = `
		UPDATE ingredients
		SET concern_level = NULL, score_version = NULL, score_sources = NULL
		WHERE id = $1`

	_, err := tx.ExecContext(ctx, query, id)
	return err
}
