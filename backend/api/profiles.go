package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"

	"github.com/nisa/beauty-ingredient/models"
)

type profileRequest struct {
	Email     string   `json:"email"`
	SkinType  string   `json:"skin_type"`
	Allergens []string `json:"allergens"`
}

// normalizeAllergens, alerjen listesini kırpar, küçük harfe çevirir ve tekilleştirir.
func normalizeAllergens(in []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, a := range in {
		a = strings.ToLower(strings.TrimSpace(a))
		if a == "" || seen[a] {
			continue
		}
		seen[a] = true
		out = append(out, a)
	}
	return out
}

// CreateProfile, POST /profiles isteğini işler.
func (s *Server) CreateProfile(c *gin.Context) {
	var req profileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}

	req.SkinType = strings.ToLower(strings.TrimSpace(req.SkinType))
	if req.SkinType != "" && !models.IsValidSkinType(req.SkinType) {
		badRequest(c, "Geçersiz skin_type. Geçerli değerler: "+models.SkinTypeHint())
		return
	}

	ctx := c.Request.Context()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		serverError(c, err)
		return
	}
	defer tx.Rollback()

	profile := models.UserProfile{
		Email:     strings.TrimSpace(req.Email),
		SkinType:  req.SkinType,
		Allergens: normalizeAllergens(req.Allergens),
	}

	err = tx.QueryRowContext(ctx,
		`INSERT INTO user_profiles (email, skin_type) VALUES (NULLIF($1, ''), NULLIF($2, ''))
		 RETURNING id, created_at`,
		profile.Email, profile.SkinType,
	).Scan(&profile.ID, &profile.CreatedAt)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation" {
		c.JSON(http.StatusConflict, gin.H{"error": "Bu e-postayla bir profil zaten var"})
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}

	for _, allergen := range profile.Allergens {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO user_allergens (user_id, allergen_name) VALUES ($1, $2)
			 ON CONFLICT (user_id, allergen_name) DO NOTHING`,
			profile.ID, allergen,
		); err != nil {
			serverError(c, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		serverError(c, err)
		return
	}

	c.JSON(http.StatusCreated, profile)
}

// GetProfile, GET /profiles/:id isteğini işler.
func (s *Server) GetProfile(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}

	const query = `
		SELECT p.id, COALESCE(p.email, ''), COALESCE(p.skin_type, ''), p.created_at,
		       COALESCE(ARRAY_AGG(ua.allergen_name) FILTER (WHERE ua.allergen_name IS NOT NULL), '{}')
		FROM user_profiles p
		LEFT JOIN user_allergens ua ON ua.user_id = p.id
		WHERE p.id = $1
		GROUP BY p.id`

	var profile models.UserProfile
	err := s.DB.QueryRowContext(c.Request.Context(), query, id).Scan(
		&profile.ID, &profile.Email, &profile.SkinType, &profile.CreatedAt,
		pq.Array(&profile.Allergens),
	)
	if errors.Is(err, sql.ErrNoRows) {
		notFound(c, "Profil bulunamadı")
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile, PUT /profiles/:id isteğini işler; cilt tipini ve kayıtlı
// alerjen listesini değiştirir.
func (s *Server) UpdateProfile(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}

	var req profileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}

	req.SkinType = strings.ToLower(strings.TrimSpace(req.SkinType))
	if req.SkinType != "" && !models.IsValidSkinType(req.SkinType) {
		badRequest(c, "Geçersiz skin_type. Geçerli değerler: "+models.SkinTypeHint())
		return
	}

	ctx := c.Request.Context()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		serverError(c, err)
		return
	}
	defer tx.Rollback()

	profile := models.UserProfile{
		ID:        id,
		SkinType:  req.SkinType,
		Allergens: normalizeAllergens(req.Allergens),
	}

	err = tx.QueryRowContext(ctx,
		`UPDATE user_profiles
		 SET skin_type = COALESCE(NULLIF($2, ''), skin_type), updated_at = CURRENT_TIMESTAMP
		 WHERE id = $1
		 RETURNING COALESCE(email, ''), COALESCE(skin_type, ''), created_at`,
		id, profile.SkinType,
	).Scan(&profile.Email, &profile.SkinType, &profile.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		notFound(c, "Profil bulunamadı")
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM user_allergens WHERE user_id = $1", id); err != nil {
		serverError(c, err)
		return
	}
	for _, allergen := range profile.Allergens {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO user_allergens (user_id, allergen_name) VALUES ($1, $2)
			 ON CONFLICT (user_id, allergen_name) DO NOTHING`,
			id, allergen,
		); err != nil {
			serverError(c, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, profile)
}
