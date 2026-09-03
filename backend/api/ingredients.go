package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"

	"github.com/nisa/beauty-ingredient/models"
)

// ingredientSelect, içerik satırını cilt tipleri, alerjenleri ve faydaları
// dizilere toplanmış halde döndürür; böylece tek sorgu yeter.
const ingredientSelect = `
	SELECT i.id,
	       i.name,
	       COALESCE(i.inci_name, ''),
	       COALESCE(i.description, ''),
	       COALESCE(i.concern_level, 0),
	       COALESCE(ARRAY_AGG(DISTINCT st.skin_type) FILTER (WHERE st.skin_type IS NOT NULL), '{}'),
	       COALESCE(ARRAY_AGG(DISTINCT al.allergen_name) FILTER (WHERE al.allergen_name IS NOT NULL), '{}'),
	       COALESCE(ARRAY_AGG(DISTINCT be.benefit) FILTER (WHERE be.benefit IS NOT NULL), '{}')
	FROM ingredients i
	LEFT JOIN ingredient_skin_types st ON st.ingredient_id = i.id
	LEFT JOIN ingredient_allergens al ON al.ingredient_id = i.id
	LEFT JOIN ingredient_benefits be ON be.ingredient_id = i.id`

func scanIngredient(scan func(dest ...any) error) (models.Ingredient, error) {
	var ing models.Ingredient
	err := scan(
		&ing.ID, &ing.Name, &ing.INCIName, &ing.Description, &ing.ConcernLevel,
		pq.Array(&ing.SkinTypes), pq.Array(&ing.Allergens), pq.Array(&ing.Benefits),
	)
	return ing, err
}

// ListIngredients, isteğe bağlı arama ve filtrelerle GET /ingredients'i işler.
//
//	?q=              ad / INCI adı üzerinde serbest metin araması
//	?skin_type=      o cilt tipine uygun (veya "all" etiketli) içerikleri tutar
//	?avoid_allergens=virgülle ayrılır; bu alerjenleri taşıyan içerikleri eler
//	?max_concern=    EWG endişe seviyesi üst sınırı
//	?limit= &offset= sayfalama (limit varsayılan 50, en fazla 200)
func (s *Server) ListIngredients(c *gin.Context) {
	skinType := strings.ToLower(strings.TrimSpace(c.Query("skin_type")))
	if skinType != "" && !models.IsValidSkinType(skinType) {
		badRequest(c, "Geçersiz skin_type. Geçerli değerler: "+models.SkinTypeHint())
		return
	}

	avoid := []string{}
	if raw := c.Query("avoid_allergens"); raw != "" {
		for _, a := range strings.Split(raw, ",") {
			if a = strings.ToLower(strings.TrimSpace(a)); a != "" {
				avoid = append(avoid, a)
			}
		}
	}

	var (
		where []string
		args  []any
	)
	arg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if q := strings.TrimSpace(c.Query("q")); q != "" {
		p := arg("%" + q + "%")
		where = append(where, fmt.Sprintf("(i.name ILIKE %s OR i.inci_name ILIKE %s)", p, p))
	}
	if skinType != "" {
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM ingredient_skin_types s2 "+
				"WHERE s2.ingredient_id = i.id AND s2.skin_type IN (%s, 'all'))", arg(skinType)))
	}
	if len(avoid) > 0 {
		where = append(where, fmt.Sprintf(
			"NOT EXISTS (SELECT 1 FROM ingredient_allergens a2 "+
				"WHERE a2.ingredient_id = i.id AND LOWER(a2.allergen_name) = ANY(%s))", arg(pq.Array(avoid))))
	}
	if c.Query("max_concern") != "" {
		where = append(where, fmt.Sprintf("COALESCE(i.concern_level, 0) <= %s", arg(intQuery(c, "max_concern", 10, 10))))
	}

	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}

	ctx := c.Request.Context()

	// Sayfalamadan önceki toplam; istemci filtrelenmiş kümede gezinebilsin diye.
	var total int
	if err := s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM ingredients i"+clause, args...).Scan(&total); err != nil {
		serverError(c, err)
		return
	}

	limit := intQuery(c, "limit", 50, 200)
	if limit == 0 {
		limit = 50
	}
	offset := intQuery(c, "offset", 0, 0)

	query := ingredientSelect + clause +
		fmt.Sprintf(" GROUP BY i.id ORDER BY i.name LIMIT %s OFFSET %s", arg(limit), arg(offset))

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		serverError(c, err)
		return
	}
	defer rows.Close()

	ingredients := []models.Ingredient{}
	for rows.Next() {
		ing, err := scanIngredient(rows.Scan)
		if err != nil {
			serverError(c, err)
			return
		}
		ingredients = append(ingredients, ing)
	}
	if err := rows.Err(); err != nil {
		serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"skin_type":       skinType,
		"avoid_allergens": avoid,
		"total":           total,
		"limit":           limit,
		"offset":          offset,
		"ingredients":     ingredients,
	})
}

// GetIngredient, GET /ingredients/:id isteğini işler.
func (s *Server) GetIngredient(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}

	ctx := c.Request.Context()
	query := ingredientSelect + " WHERE i.id = $1 GROUP BY i.id"

	ing, err := scanIngredient(s.DB.QueryRowContext(ctx, query, id).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		notFound(c, "İçerik bulunamadı")
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}

	if err := s.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM product_ingredients WHERE ingredient_id = $1", id,
	).Scan(&ing.ProductsCount); err != nil {
		serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, ing)
}

// AllergenCheck, POST /ingredients/allergen-check isteğini işler: bir ürünün
// hangi içeriklerinin kullanıcının listelediği alerjenleri taşıdığını bildirir.
func (s *Server) AllergenCheck(c *gin.Context) {
	var req struct {
		ProductID     int      `json:"product_id"`
		UserAllergens []string `json:"user_allergens"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}
	if req.ProductID <= 0 {
		badRequest(c, "product_id zorunludur")
		return
	}

	ctx := c.Request.Context()

	var productName string
	err := s.DB.QueryRowContext(ctx, "SELECT name FROM products WHERE id = $1", req.ProductID).Scan(&productName)
	if errors.Is(err, sql.ErrNoRows) {
		notFound(c, "Ürün bulunamadı")
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}

	normalized := []string{}
	for _, a := range req.UserAllergens {
		if a = strings.ToLower(strings.TrimSpace(a)); a != "" {
			normalized = append(normalized, a)
		}
	}

	matches := []models.AllergenMatch{}
	if len(normalized) > 0 {
		// İki yönlü alt dize eşleşmesi: "parfüm" arayan "parfüm karışımı"nı da,
		// tersi de yakalansın diye. NOT: bu yaklaşım yanlış pozitif üretebilir
		// (örn. "alkol" arayanın "yün alkolü"ne takılması);
		// docs/DEVELOPMENT_PLAN.md P0-1 bunu ele alır.
		const query = `
			SELECT DISTINCT ia.allergen_name,
			                i.name,
			                COALESCE(ia.severity, 0),
			                COALESCE(i.concern_level, 0)
			FROM product_ingredients pi
			JOIN ingredients i ON i.id = pi.ingredient_id
			JOIN ingredient_allergens ia ON ia.ingredient_id = i.id
			WHERE pi.product_id = $1
			  AND EXISTS (
			      SELECT 1 FROM UNNEST($2::text[]) AS ua(term)
			      WHERE LOWER(ia.allergen_name) LIKE '%' || ua.term || '%'
			         OR ua.term LIKE '%' || LOWER(ia.allergen_name) || '%'
			  )
			ORDER BY 3 DESC`

		rows, err := s.DB.QueryContext(ctx, query, req.ProductID, pq.Array(normalized))
		if err != nil {
			serverError(c, err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var m models.AllergenMatch
			if err := rows.Scan(&m.Allergen, &m.Ingredient, &m.Severity, &m.ConcernLevel); err != nil {
				serverError(c, err)
				return
			}
			matches = append(matches, m)
		}
		if err := rows.Err(); err != nil {
			serverError(c, err)
			return
		}
	}

	flags := []string{}
	seen := map[string]bool{}
	for _, m := range matches {
		if !seen[m.Allergen] {
			seen[m.Allergen] = true
			flags = append(flags, m.Allergen+" içeriyor")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"product_id":   req.ProductID,
		"product_name": productName,
		"matches":      matches,
		"safe":         len(matches) == 0,
		"flags":        flags,
	})
}
