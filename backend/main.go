// beauty-ingredient komutu, Beauty Ingredient Explorer HTTP API'sini sunar.
package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/nisa/beauty-ingredient/api"
	"github.com/nisa/beauty-ingredient/middleware"
)

func main() {
	// .env isteğe bağlıdır: üretimde ortam değişkenlerini platform belirler.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf(".env dosyası yüklenmedi (%v), ortam değişkenlerine düşülüyor", err)
	}

	db, err := openDB(env("DATABASE_URL", "postgres://beauty:beauty@localhost:5433/beauty_ingredient?sslmode=disable"))
	if err != nil {
		log.Fatalf("veritabanı: %v", err)
	}
	defer db.Close()

	if mode := env("GIN_MODE", ""); mode != "" {
		gin.SetMode(mode)
	}

	router := gin.Default()
	router.Use(middleware.CORS(env("CORS_ALLOWED_ORIGINS", "http://localhost:3000")))
	router.MaxMultipartMemory = 10 << 20 // 10 MB; yükleme sayfasının sınırıyla aynı

	api.New(db).RegisterRoutes(router)

	srv := &http.Server{
		Addr:              ":" + env("PORT", "8080"),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Arka planda sun; ana goroutine sinyal bekleyebilsin.
	go func() {
		log.Printf("%s adresinde dinleniyor", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("sunucu: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("kapatılıyor...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("zorla kapatıldı: %v", err)
	}
}

// openDB, PostgreSQL'e bağlanır ve bağlantının kullanılabilir olduğunu doğrular.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	log.Println("postgres bağlantısı kuruldu")
	return db, nil
}

// env, ortam değişkenini okur; tanımsız veya boşsa def değerine düşer.
func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
