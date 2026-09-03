package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"

	"github.com/nisa/beauty-ingredient/middleware"
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

// loadProfile, bir kullanıcının profilini alerjenleriyle birlikte yükler.
func (s *Server) loadProfile(ctx context.Context, userID int) (models.UserProfile, error) {
	const query = `
		SELECT p.id, COALESCE(p.email, ''), COALESCE(p.skin_type, ''), p.created_at,
		       COALESCE(ARRAY_AGG(ua.allergen_name) FILTER (WHERE ua.allergen_name IS NOT NULL), '{}')
		FROM user_profiles p
		LEFT JOIN user_allergens ua ON ua.user_id = p.id
		WHERE p.id = $1 AND p.deleted_at IS NULL
		GROUP BY p.id`

	var profile models.UserProfile
	err := s.DB.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID, &profile.Email, &profile.SkinType, &profile.CreatedAt,
		pq.Array(&profile.Allergens),
	)
	return profile, err
}

// GetMyProfile, GET /profiles/me isteğini işler.
//
// /profiles/:id bilinçli olarak kaldırıldı: sahiplik kontrolünü unutmanın
// yapısal olarak imkânsız olduğu tasarım, her uç noktaya kontrol eklemekten
// güvenlidir.
func (s *Server) GetMyProfile(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Giriş yapmalısınız"})
		return
	}

	profile, err := s.loadProfile(c.Request.Context(), userID)
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

// UpdateMyProfile, PUT /profiles/me isteğini işler.
func (s *Server) UpdateMyProfile(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Giriş yapmalısınız"})
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

	allergens := normalizeAllergens(req.Allergens)
	ctx := c.Request.Context()

	// Alerjen yazmak sağlık verisi işlemektir: geçerli bir açık rıza şart.
	if len(allergens) > 0 {
		granted, err := s.hasConsent(ctx, userID, ConsentHealthData)
		if err != nil {
			serverError(c, err)
			return
		}
		if !granted {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Alerjen bilgisi kaydedebilmemiz için açık rızanız gerekiyor",
			})
			return
		}
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		serverError(c, err)
		return
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE user_profiles
		 SET skin_type = COALESCE(NULLIF($2, ''), skin_type), updated_at = CURRENT_TIMESTAMP
		 WHERE id = $1 AND deleted_at IS NULL`,
		userID, req.SkinType)
	if err != nil {
		serverError(c, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		notFound(c, "Profil bulunamadı")
		return
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM user_allergens WHERE user_id = $1", userID); err != nil {
		serverError(c, err)
		return
	}
	for _, a := range allergens {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO user_allergens (user_id, allergen_name) VALUES ($1, $2)
			 ON CONFLICT (user_id, allergen_name) DO NOTHING`, userID, a); err != nil {
			serverError(c, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		serverError(c, err)
		return
	}

	profile, err := s.loadProfile(ctx, userID)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, profile)
}

// hasConsent, kullanıcının verilen tür için EN SON kararının olumlu olup
// olmadığını söyler. Geri alınmış rıza geçmişte kalır ama geçerli değildir.
func (s *Server) hasConsent(ctx context.Context, userID int, consentType string) (bool, error) {
	var granted bool
	err := s.DB.QueryRowContext(ctx,
		`SELECT granted FROM consent_log
		 WHERE user_id = $1 AND consent_type = $2
		 ORDER BY created_at DESC, id DESC LIMIT 1`,
		userID, consentType).Scan(&granted)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return granted, err
}

// nowRFC3339, dışa aktarım damgası için.
func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}
