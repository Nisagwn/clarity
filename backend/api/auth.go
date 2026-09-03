package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/nisa/beauty-ingredient/middleware"
	"github.com/nisa/beauty-ingredient/models"
)

// Rıza metinlerinin sürümü. Metin değişirse bu artar ve kullanıcıdan yeniden
// rıza istenir; eski rızanın hangi metne verildiği consent_log'da kalır.
const PolicyVersion = "2026-09-03"

// Rıza türleri. health_data ve marketing AYRI tutulur ve asla birlikte
// sorulmaz: bir sağlık uygulaması, sağlık verisi rızasını pazarlama rızasıyla
// paketlediği için 1,5M € ceza aldı.
const (
	ConsentHealthData = "health_data"
	ConsentMarketing  = "marketing"
)

// secureCookies, cookie'nin yalnızca HTTPS üzerinden gönderilip
// gönderilmeyeceğini belirler. Yerelde http kullanıldığı için kapalı.
func secureCookies() bool {
	return strings.EqualFold(os.Getenv("COOKIE_SECURE"), "true")
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	SkinType string `json:"skin_type"`
	// Alerjen verisi işlenmesine açık rıza. Zorunlu ve ayrı: bu olmadan
	// alerjen kaydedilmez. Kayıt sözleşmesine gömülmez.
	HealthDataConsent bool     `json:"health_data_consent"`
	MarketingConsent  bool     `json:"marketing_consent"`
	Allergens         []string `json:"allergens"`
}

// Register, POST /auth/register isteğini işler.
func (s *Server) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		badRequest(c, "Geçerli bir e-posta adresi girin")
		return
	}
	if len(req.Password) < 10 {
		badRequest(c, "Parola en az 10 karakter olmalı")
		return
	}

	req.SkinType = strings.ToLower(strings.TrimSpace(req.SkinType))
	if req.SkinType != "" && !models.IsValidSkinType(req.SkinType) {
		badRequest(c, "Geçersiz skin_type. Geçerli değerler: "+models.SkinTypeHint())
		return
	}

	// Alerjen verisi sağlık verisidir: açık rıza olmadan kaydedilemez.
	allergens := normalizeAllergens(req.Allergens)
	if len(allergens) > 0 && !req.HealthDataConsent {
		badRequest(c, "Alerjen bilgisi kaydedebilmemiz için açık rızanız gerekiyor")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		serverError(c, err)
		return
	}

	ctx := c.Request.Context()
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		serverError(c, err)
		return
	}
	defer tx.Rollback()

	var userID int
	err = tx.QueryRowContext(ctx,
		`INSERT INTO user_profiles (email, password_hash, skin_type)
		 VALUES ($1, $2, NULLIF($3, '')) RETURNING id`,
		email, string(hash), req.SkinType,
	).Scan(&userID)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation" {
		c.JSON(http.StatusConflict, gin.H{"error": "Bu e-postayla bir hesap zaten var"})
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}

	// Rızalar ayrı ayrı kaydedilir, verilmese bile: "hayır" da bir kayıttır.
	if err := logConsent(ctx, tx, c, userID, ConsentHealthData, req.HealthDataConsent); err != nil {
		serverError(c, err)
		return
	}
	if err := logConsent(ctx, tx, c, userID, ConsentMarketing, req.MarketingConsent); err != nil {
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

	if err := middleware.IssueSession(c, userID, secureCookies()); err != nil {
		serverError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": userID, "email": email})
}

// Login, POST /auth/login isteğini işler.
func (s *Server) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	var (
		userID int
		hash   sql.NullString
	)
	err := s.DB.QueryRowContext(c.Request.Context(),
		`SELECT id, password_hash FROM user_profiles
		 WHERE LOWER(email) = $1 AND deleted_at IS NULL`, email,
	).Scan(&userID, &hash)

	// Kullanıcı yok ile parola yanlış aynı yanıtı verir: aksi halde bu uç
	// nokta hangi e-postaların kayıtlı olduğunu sızdıran bir araca dönüşür.
	if errors.Is(err, sql.ErrNoRows) || !hash.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "E-posta veya parola hatalı"})
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hash.String), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "E-posta veya parola hatalı"})
		return
	}

	if err := middleware.IssueSession(c, userID, secureCookies()); err != nil {
		serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": userID, "email": email})
}

// Logout, POST /auth/logout isteğini işler.
func (s *Server) Logout(c *gin.Context) {
	middleware.ClearSession(c, secureCookies())
	c.JSON(http.StatusOK, gin.H{"status": "çıkış yapıldı"})
}

// UpdateConsent, POST /auth/consent isteğini işler: rıza verme veya geri alma.
//
// Rıza geri alındığında işleme derhal durmalı — sağlık verisi rızası geri
// alınırsa alerjen kayıtları da silinir.
func (s *Server) UpdateConsent(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Giriş yapmalısınız"})
		return
	}

	var req struct {
		ConsentType string `json:"consent_type"`
		Granted     bool   `json:"granted"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}
	if req.ConsentType != ConsentHealthData && req.ConsentType != ConsentMarketing {
		badRequest(c, "Geçersiz consent_type")
		return
	}

	ctx := c.Request.Context()
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		serverError(c, err)
		return
	}
	defer tx.Rollback()

	if err := logConsent(ctx, tx, c, userID, req.ConsentType, req.Granted); err != nil {
		serverError(c, err)
		return
	}

	// Sağlık verisi rızası geri alındıysa veri derhal silinir.
	if req.ConsentType == ConsentHealthData && !req.Granted {
		if _, err := tx.ExecContext(ctx,
			"DELETE FROM user_allergens WHERE user_id = $1", userID); err != nil {
			serverError(c, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"consent_type":   req.ConsentType,
		"granted":        req.Granted,
		"policy_version": PolicyVersion,
	})
}

// DeleteAccount, DELETE /auth/account isteğini işler (silme hakkı).
func (s *Server) DeleteAccount(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Giriş yapmalısınız"})
		return
	}

	// Gerçek silme, işaretleme değil: KVKK ve GDPR silme hakkı verinin
	// gerçekten yok edilmesini istiyor. İlişkili satırlar cascade ile gider.
	res, err := s.DB.ExecContext(c.Request.Context(),
		"DELETE FROM user_profiles WHERE id = $1", userID)
	if err != nil {
		serverError(c, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		notFound(c, "Hesap bulunamadı")
		return
	}

	middleware.ClearSession(c, secureCookies())
	c.JSON(http.StatusOK, gin.H{"status": "hesap ve ilişkili tüm veri silindi"})
}

// ExportData, GET /profiles/me/export isteğini işler (veri taşınabilirliği).
func (s *Server) ExportData(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Giriş yapmalısınız"})
		return
	}

	ctx := c.Request.Context()

	profile, err := s.loadProfile(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		notFound(c, "Profil bulunamadı")
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}

	type consentRow struct {
		Type          string `json:"consent_type"`
		Granted       bool   `json:"granted"`
		PolicyVersion string `json:"policy_version"`
		CreatedAt     string `json:"created_at"`
	}

	rows, err := s.DB.QueryContext(ctx,
		`SELECT consent_type, granted, policy_version, created_at
		 FROM consent_log WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		serverError(c, err)
		return
	}
	defer rows.Close()

	consents := []consentRow{}
	for rows.Next() {
		var r consentRow
		if err := rows.Scan(&r.Type, &r.Granted, &r.PolicyVersion, &r.CreatedAt); err != nil {
			serverError(c, err)
			return
		}
		consents = append(consents, r)
	}
	if err := rows.Err(); err != nil {
		serverError(c, err)
		return
	}

	c.Header("Content-Disposition", `attachment; filename="clarity-verilerim.json"`)
	c.JSON(http.StatusOK, gin.H{
		"profile":        profile,
		"consent_log":    consents,
		"exported_at":    nowRFC3339(),
		"policy_version": PolicyVersion,
	})
}

// logConsent, bir rıza kararını kanıtlanabilir biçimde kaydeder.
func logConsent(ctx context.Context, tx *sql.Tx, c *gin.Context, userID int, consentType string, granted bool) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO consent_log (user_id, consent_type, granted, policy_version, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, NULLIF($5, '')::inet, NULLIF($6, ''))`,
		userID, consentType, granted, PolicyVersion, c.ClientIP(), c.Request.UserAgent(),
	)
	return err
}
