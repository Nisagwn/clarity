package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"

	"github.com/nisa/beauty-ingredient/models"
)

// productColumns, ürün satırları için ortak izdüşümdür.
const productColumns = `
	SELECT p.id, p.name, p.brand,
	       COALESCE(p.gtin, ''),
	       COALESCE(p.price, 0),
	       COALESCE(p.currency, 'USD'),
	       COALESCE(p.image_url, ''),
	       COALESCE(p.category, ''),
	       COALESCE(p.description, ''),
	       COALESCE(p.source_url, ''),
	       p.created_at,
	       COALESCE(p.updated_at, p.created_at)
	FROM products p`

func scanProduct(scan func(dest ...any) error) (models.Product, error) {
	var p models.Product
	err := scan(
		&p.ID, &p.Name, &p.Brand, &p.GTIN, &p.Price, &p.Currency,
		&p.ImageURL, &p.Category, &p.Description, &p.SourceURL,
		&p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

func (s *Server) queryProducts(c *gin.Context, query string, args ...any) ([]models.Product, bool) {
	rows, err := s.DB.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		serverError(c, err)
		return nil, false
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		p, err := scanProduct(rows.Scan)
		if err != nil {
			serverError(c, err)
			return nil, false
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		serverError(c, err)
		return nil, false
	}
	return products, true
}

// CreateProduct, POST /products isteğini işler.
func (s *Server) CreateProduct(c *gin.Context) {
	var p models.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		badRequest(c, err.Error())
		return
	}
	if strings.TrimSpace(p.Name) == "" || strings.TrimSpace(p.Brand) == "" {
		badRequest(c, "name ve brand zorunludur")
		return
	}
	if p.Currency == "" {
		p.Currency = "USD"
	}

	const query = `
		INSERT INTO products (name, brand, gtin, price, currency, image_url, category, description, source_url)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''))
		RETURNING id, created_at, COALESCE(updated_at, created_at)`

	err := s.DB.QueryRowContext(c.Request.Context(), query,
		p.Name, p.Brand, p.GTIN, p.Price, p.Currency,
		p.ImageURL, p.Category, p.Description, p.SourceURL,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation" {
		c.JSON(http.StatusConflict, gin.H{"error": "Bu markada bu GTIN'e sahip bir ürün zaten var"})
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}

	c.JSON(http.StatusCreated, p)
}

// GetProduct, GET /products/:id isteğini işler ve tam içerik listesini ekler.
func (s *Server) GetProduct(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}

	ctx := c.Request.Context()

	p, err := scanProduct(s.DB.QueryRowContext(ctx, productColumns+" WHERE p.id = $1", id).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		notFound(c, "Ürün bulunamadı")
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}

	ingredients, err := s.productIngredients(ctx, id)
	if err != nil {
		serverError(c, err)
		return
	}
	p.Ingredients = ingredients

	c.JSON(http.StatusOK, p)
}

// productIngredients, bir ürünün içeriklerini INCI listesi sırasıyla yükler.
func (s *Server) productIngredients(ctx context.Context, productID int) ([]models.Ingredient, error) {
	const query = `
		SELECT i.id,
		       i.name,
		       COALESCE(i.inci_name, ''),
		       COALESCE(i.description, ''),
		       COALESCE(i.concern_level, 0),
		       COALESCE(ARRAY_AGG(DISTINCT st.skin_type) FILTER (WHERE st.skin_type IS NOT NULL), '{}'),
		       COALESCE(ARRAY_AGG(DISTINCT al.allergen_name) FILTER (WHERE al.allergen_name IS NOT NULL), '{}'),
		       COALESCE(ARRAY_AGG(DISTINCT be.benefit) FILTER (WHERE be.benefit IS NOT NULL), '{}'),
		       COALESCE(pi.order_index, 0)
		FROM product_ingredients pi
		JOIN ingredients i ON i.id = pi.ingredient_id
		LEFT JOIN ingredient_skin_types st ON st.ingredient_id = i.id
		LEFT JOIN ingredient_allergens al ON al.ingredient_id = i.id
		LEFT JOIN ingredient_benefits be ON be.ingredient_id = i.id
		WHERE pi.product_id = $1
		GROUP BY i.id, pi.order_index
		ORDER BY COALESCE(pi.order_index, 0), i.name`

	rows, err := s.DB.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ingredients := []models.Ingredient{}
	for rows.Next() {
		var ing models.Ingredient
		if err := rows.Scan(
			&ing.ID, &ing.Name, &ing.INCIName, &ing.Description, &ing.ConcernLevel,
			pq.Array(&ing.SkinTypes), pq.Array(&ing.Allergens), pq.Array(&ing.Benefits),
			&ing.OrderIndex,
		); err != nil {
			return nil, err
		}
		ingredients = append(ingredients, ing)
	}
	return ingredients, rows.Err()
}

// ListProducts, isteğe bağlı marka/kategori filtreleriyle GET /products'i işler.
func (s *Server) ListProducts(c *gin.Context) {
	var (
		where []string
		args  []any
	)
	arg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if brand := strings.TrimSpace(c.Query("brand")); brand != "" {
		where = append(where, fmt.Sprintf("p.brand ILIKE %s", arg(brand)))
	}
	if category := strings.TrimSpace(c.Query("category")); category != "" {
		where = append(where, fmt.Sprintf("p.category ILIKE %s", arg(category)))
	}
	if q := strings.TrimSpace(c.Query("q")); q != "" {
		p := arg("%" + q + "%")
		where = append(where, fmt.Sprintf("(p.name ILIKE %s OR p.brand ILIKE %s)", p, p))
	}

	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	if err := s.DB.QueryRowContext(c.Request.Context(), "SELECT COUNT(*) FROM products p"+clause, args...).Scan(&total); err != nil {
		serverError(c, err)
		return
	}

	limit := intQuery(c, "limit", 50, 200)
	if limit == 0 {
		limit = 50
	}
	offset := intQuery(c, "offset", 0, 0)

	query := productColumns + clause + fmt.Sprintf(" ORDER BY p.brand, p.name LIMIT %s OFFSET %s", arg(limit), arg(offset))

	products, ok := s.queryProducts(c, query, args...)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"limit":    limit,
		"offset":   offset,
		"products": products,
	})
}

// SearchProducts, GET /products/search?q= isteğini ürün adları, markalar ve
// içerik adları üzerinde işler.
func (s *Server) SearchProducts(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		badRequest(c, "q sorgu parametresi zorunludur")
		return
	}

	query := productColumns + `
		WHERE p.name ILIKE $1
		   OR p.brand ILIKE $1
		   OR EXISTS (
		       SELECT 1 FROM product_ingredients pi
		       JOIN ingredients i ON i.id = pi.ingredient_id
		       WHERE pi.product_id = p.id AND (i.name ILIKE $1 OR i.inci_name ILIKE $1)
		   )
		ORDER BY p.brand, p.name
		LIMIT $2`

	products, ok := s.queryProducts(c, query, "%"+q+"%", intQuery(c, "limit", 50, 200))
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{"query": q, "count": len(products), "results": products})
}

// GetProductsByBrand, GET /products/brand/:brand isteğini işler.
func (s *Server) GetProductsByBrand(c *gin.Context) {
	brand := c.Param("brand")

	products, ok := s.queryProducts(c, productColumns+" WHERE p.brand ILIKE $1 ORDER BY p.name", brand)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{"brand": brand, "count": len(products), "products": products})
}

// AttachIngredients, POST /products/:id/ingredients isteğini işler; mevcut
// içerikleri INCI sırasıyla bir ürüne bağlar.
func (s *Server) AttachIngredients(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}

	var req struct {
		IngredientIDs []int `json:"ingredient_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}
	if len(req.IngredientIDs) == 0 {
		badRequest(c, "ingredient_ids boş olamaz")
		return
	}

	ctx := c.Request.Context()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		serverError(c, err)
		return
	}
	defer tx.Rollback()

	const insert = `
		INSERT INTO product_ingredients (product_id, ingredient_id, order_index)
		VALUES ($1, $2, $3)
		ON CONFLICT (product_id, ingredient_id) DO UPDATE SET order_index = EXCLUDED.order_index`

	for i, ingID := range req.IngredientIDs {
		if _, err := tx.ExecContext(ctx, insert, id, ingID, i+1); err != nil {
			serverError(c, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"product_id": id, "linked": len(req.IngredientIDs)})
}
