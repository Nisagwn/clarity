package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"

	"github.com/nisa/beauty-ingredient/models"
)

// dupeQuery, diğer ürünleri içerik kümeleri üzerinden Jaccard benzerliğiyle
// sıralar: ortak içerik sayısı bölü birleşim kümesinin büyüklüğü. Aynı
// kategorideki ürünler daha yüksek puan alır; profil verildiğinde kullanıcının
// alerjenlerini taşıyan adaylar tamamen elenir.
const dupeQuery = `
	WITH target AS (
	    SELECT ingredient_id FROM product_ingredients WHERE product_id = $1
	),
	target_meta AS (
	    SELECT COALESCE(category, '') AS category, COALESCE(price, 0) AS price
	    FROM products WHERE id = $1
	),
	candidate AS (
	    SELECT p.id,
	           p.name,
	           p.brand,
	           p.price,
	           COALESCE(p.currency, 'USD') AS currency,
	           COALESCE(p.image_url, '')  AS image_url,
	           COALESCE(p.category, '')   AS category,
	           COUNT(*) FILTER (WHERE pi.ingredient_id IN (SELECT ingredient_id FROM target)) AS shared,
	           COUNT(*) AS total_ingredients
	    FROM products p
	    JOIN product_ingredients pi ON pi.product_id = p.id
	    WHERE p.id <> $1
	      -- Eksik içerik listeli ürünler aday olamaz: iki kısa listenin
	      -- Jaccard benzerliği yüksek çıkar ve bu benzerlik değil, veri
	      -- eksikliğidir.
	      AND p.data_quality <> 'incomplete'
	    GROUP BY p.id
	)
	SELECT c.id, c.name, c.brand, c.price, c.currency, c.image_url, c.category,
	       c.shared::float / NULLIF(
	           (SELECT COUNT(*) FROM target) + c.total_ingredients - c.shared, 0) AS similarity,
	       (SELECT price FROM target_meta) AS target_price,
	       (SELECT category FROM target_meta) AS target_category
	FROM candidate c
	WHERE c.shared > 0
	  -- Alerjen elemesi kanonik sözlükten TAM eşleşmeyle yapılır. Alt dize
	  -- karşılaştırması "alkol" arayan kullanıcıya alerjeni "yün alkolü" olan
	  -- Lanolin'i eledirirdi; Faz 1'de alerjen kontrolünde düzeltilen hata
	  -- muadil sorgusunda kalmıştı.
	  AND ($3::text[] IS NULL OR NOT EXISTS (
	      SELECT 1
	      FROM product_ingredients pi2
	      JOIN ingredient_allergens ia ON ia.ingredient_id = pi2.ingredient_id
	      JOIN allergen_alias ing_alias ON LOWER(ing_alias.alias) = LOWER(ia.allergen_name)
	      JOIN allergen_alias user_alias ON user_alias.canonical_id = ing_alias.canonical_id
	      WHERE pi2.product_id = c.id
	        AND LOWER(user_alias.alias) = ANY($3::text[])
	  ))
	ORDER BY similarity DESC,
	         ABS(c.price - (SELECT price FROM target_meta)) ASC NULLS LAST
	LIMIT $2`

// recommend, muadil sorgusunu çalıştırır ve her sonucu etiketler.
func (s *Server) recommend(ctx context.Context, productID, topN int, allergens []string) ([]models.Recommendation, error) {
	var allergenArg any
	if len(allergens) > 0 {
		allergenArg = pq.Array(allergens)
	}

	rows, err := s.DB.QueryContext(ctx, dupeQuery, productID, topN, allergenArg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recs := []models.Recommendation{}
	for rows.Next() {
		var (
			r              models.Recommendation
			category       string
			similarity     sql.NullFloat64
			targetPrice    *float64
			targetCategory string
		)
		if err := rows.Scan(
			&r.ID, &r.Name, &r.Brand, &r.Price, &r.Currency, &r.ImageURL, &category,
			&similarity, &targetPrice, &targetCategory,
		); err != nil {
			return nil, err
		}

		r.SimilarityScore = similarity.Float64

		// Muadil, aynı kategoride yakın içerik eşleşmesidir; örtüşen diğer her
		// şey alternatif olarak sunulur.
		switch {
		case r.SimilarityScore >= 0.5 && (targetCategory == "" || category == targetCategory):
			r.Type = "dupe"
		default:
			r.Type = "alternative"
		}

		// Fiyat karşılaştırması yalnızca iki fiyat da biliniyorsa yapılır:
		// bilinmeyen fiyatı 0 sayıp "daha ucuz" demek uydurma olurdu.
		cheaper := r.Price != nil && targetPrice != nil && *r.Price < *targetPrice

		switch {
		case r.Type == "dupe" && cheaper:
			r.Reason = fmt.Sprintf("%%%.0f içerik örtüşmesi, %.2f %s daha ucuz",
				r.SimilarityScore*100, *targetPrice-*r.Price, r.Currency)
		case r.Type == "dupe":
			r.Reason = fmt.Sprintf("%%%.0f içerik örtüşmesi", r.SimilarityScore*100)
		default:
			r.Reason = fmt.Sprintf("İçerik profilinin %%%.0f kadarını paylaşıyor", r.SimilarityScore*100)
		}
		if len(allergens) > 0 {
			r.Reason += "; listelediğiniz alerjenleri içermiyor"
		}

		recs = append(recs, r)
	}
	return recs, rows.Err()
}

// GetRecommendations, POST /recommendations isteğini işler.
func (s *Server) GetRecommendations(c *gin.Context) {
	var req struct {
		ProductID     int `json:"product_id"`
		UserProfileID int `json:"user_profile_id"`
		TopN          int `json:"top_n"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}
	if req.ProductID <= 0 {
		badRequest(c, "product_id zorunludur")
		return
	}
	if req.TopN <= 0 || req.TopN > 50 {
		req.TopN = 5
	}

	ctx := c.Request.Context()

	var exists bool
	if err := s.DB.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)", req.ProductID,
	).Scan(&exists); err != nil {
		serverError(c, err)
		return
	}
	if !exists {
		notFound(c, "Ürün bulunamadı")
		return
	}

	// Profil verildiğinde, onun alerjenlerini taşıyan her şey elenir.
	allergens := []string{}
	if req.UserProfileID > 0 {
		rows, err := s.DB.QueryContext(ctx,
			"SELECT LOWER(allergen_name) FROM user_allergens WHERE user_id = $1", req.UserProfileID)
		if err != nil {
			serverError(c, err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var a string
			if err := rows.Scan(&a); err != nil {
				serverError(c, err)
				return
			}
			allergens = append(allergens, a)
		}
		if err := rows.Err(); err != nil {
			serverError(c, err)
			return
		}
	}

	recs, err := s.recommend(ctx, req.ProductID, req.TopN, allergens)
	if err != nil {
		serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"product_id":      req.ProductID,
		"recommendations": recs,
	})
}

// GetDupes, GET /products/:id/dupes isteğini işler; ön yüzün ürün sayfası için
// pratik bir sarmalayıcıdır.
func (s *Server) GetDupes(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}

	ctx := c.Request.Context()

	var name string
	err := s.DB.QueryRowContext(ctx, "SELECT name FROM products WHERE id = $1", id).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		notFound(c, "Ürün bulunamadı")
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}

	recs, err := s.recommend(ctx, id, intQuery(c, "limit", 5, 50), nil)
	if err != nil {
		serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"product_id": id, "product_name": name, "recommendations": recs})
}
