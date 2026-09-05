// api paketi HTTP rotalarını bağlar ve istek işleyicilerini barındırır.
//
// Hata mesajları son kullanıcıya gösterildiği için Türkçedir; rota yolları ve
// JSON alan adları docs/API_SPEC.md'deki sözleşme gereği İngilizce kalır.
package api

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/nisa/beauty-ingredient/middleware"
	"github.com/nisa/beauty-ingredient/scoring"
)

// Server, tüm işleyicilerin paylaştığı bağımlılıkları taşır.
type Server struct {
	DB *sql.DB

	// Puanlama rubrikleri sürüm başına bir kez okunur: kural kümesi yalnızca
	// göçle değişir, her istekte sorgulamak boşuna tur olurdu.
	rubricMu sync.Mutex
	rubrics  map[int]*scoring.Rubric
}

// New, db ile desteklenen bir Server oluşturur.
func New(db *sql.DB) *Server {
	return &Server{DB: db, rubrics: map[int]*scoring.Rubric{}}
}

// rubric, verilen sürümün puanlama rubriğini döndürür ve önbelleğe alır.
func (s *Server) rubric(ctx context.Context, version int) (*scoring.Rubric, error) {
	s.rubricMu.Lock()
	defer s.rubricMu.Unlock()

	if r, ok := s.rubrics[version]; ok {
		return r, nil
	}

	r, err := scoring.LoadRubric(ctx, s.DB, version)
	if err != nil {
		return nil, err
	}
	s.rubrics[version] = r
	return r, nil
}

// RegisterRoutes, docs/API_SPEC.md'de tanımlanan tüm uç noktaları bağlar.
func (s *Server) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", s.Health)

	// Ürünler — sabit segmentler :id jokerından önce kaydedilir.
	r.POST("/products", s.CreateProduct)
	r.GET("/products", s.ListProducts)
	r.GET("/products/search", s.SearchProducts)
	r.GET("/products/categories", s.ListCategories)
	r.GET("/products/brand/:brand", s.GetProductsByBrand)
	r.GET("/products/:id", s.GetProduct)
	r.GET("/products/:id/dupes", s.GetDupes)
	r.POST("/products/:id/ingredients", s.AttachIngredients)

	// İçerikler
	r.GET("/ingredients", s.ListIngredients)
	r.POST("/ingredients/allergen-check", s.AllergenCheck)
	r.GET("/ingredients/:id", s.GetIngredient)

	// Öneriler
	r.POST("/recommendations", s.GetRecommendations)

	// Kimlik doğrulama
	r.POST("/auth/register", s.Register)
	r.POST("/auth/login", s.Login)
	r.POST("/auth/logout", s.Logout)

	// Kullanıcı profilleri — hepsi oturum gerektirir.
	//
	// /profiles/:id bilinçli olarak yok: çıplak bir tam sayı alan uç nokta,
	// sayıyı artıran herkese başkasının alerjen listesini açıyordu. Kimliği
	// istemciden almak yerine oturumdan almak, sahiplik kontrolünü unutmayı
	// yapısal olarak imkânsız kılar.
	authed := r.Group("/", middleware.RequireAuth())
	{
		authed.GET("profiles/me", s.GetMyProfile)
		authed.PUT("profiles/me", s.UpdateMyProfile)
		authed.GET("profiles/me/export", s.ExportData)
		authed.POST("auth/consent", s.UpdateConsent)
		authed.DELETE("auth/account", s.DeleteAccount)
	}
}

// Health, servisin ve veritabanının ayakta olup olmadığını bildirir.
func (s *Server) Health(c *gin.Context) {
	if err := s.DB.PingContext(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "database": "unreachable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "database": "ok"})
}

// ===== yardımcılar =====

func badRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": msg})
}

func notFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, gin.H{"error": msg})
}

func serverError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

// whereClause, koşulları tek bir WHERE cümlesine bağlar; koşul yoksa boş döner.
func whereClause(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(conditions, " AND ")
}

// idParam, yoldan pozitif tam sayı bir parametre okur.
func idParam(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id <= 0 {
		badRequest(c, "Geçersiz "+name+": pozitif bir tam sayı olmalı")
		return 0, false
	}
	return id, true
}

// intQuery, tam sayı sorgu parametresi okur; yoksa def'e düşer ve max ile sınırlar.
func intQuery(c *gin.Context, name string, def, max int) int {
	raw := c.Query(name)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return def
	}
	if max > 0 && v > max {
		return max
	}
	return v
}
