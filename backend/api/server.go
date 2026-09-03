// api paketi HTTP rotalarını bağlar ve istek işleyicilerini barındırır.
//
// Hata mesajları son kullanıcıya gösterildiği için Türkçedir; rota yolları ve
// JSON alan adları docs/API_SPEC.md'deki sözleşme gereği İngilizce kalır.
package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Server, tüm işleyicilerin paylaştığı bağımlılıkları taşır.
type Server struct {
	DB *sql.DB
}

// New, db ile desteklenen bir Server oluşturur.
func New(db *sql.DB) *Server {
	return &Server{DB: db}
}

// RegisterRoutes, docs/API_SPEC.md'de tanımlanan tüm uç noktaları bağlar.
func (s *Server) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", s.Health)

	// Ürünler — sabit segmentler :id jokerından önce kaydedilir.
	r.POST("/products", s.CreateProduct)
	r.GET("/products", s.ListProducts)
	r.GET("/products/search", s.SearchProducts)
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

	// Kullanıcı profilleri
	r.POST("/profiles", s.CreateProfile)
	r.GET("/profiles/:id", s.GetProfile)
	r.PUT("/profiles/:id", s.UpdateProfile)
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
