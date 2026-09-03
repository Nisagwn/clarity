// migrate komutu, veritabanı şema göçlerini uygular.
//
// Kullanım:
//
//	go run ./cmd/migrate up        tüm bekleyen göçleri uygular
//	go run ./cmd/migrate down      son göçü geri alır
//	go run ./cmd/migrate down-all  tüm göçleri geri alır
//	go run ./cmd/migrate version   mevcut sürümü yazar
//	go run ./cmd/migrate force 1   kirli durumu verilen sürüme sabitler
//
// Göçler sunucu başlangıcında otomatik çalışmaz: birden fazla örnek aynı anda
// açıldığında yarış oluşur. Dağıtımda ayrı bir adım olarak çalıştırılır.
package main

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

const migrationsPath = "file://db/migrations"

func main() {
	log.SetFlags(0)

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf(".env yüklenmedi (%v), ortam değişkenlerine düşülüyor", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL tanımlı değil")
	}

	if len(os.Args) < 2 {
		log.Fatal("kullanım: migrate up|down|down-all|version|force <n>")
	}

	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		log.Fatalf("göç başlatılamadı: %v", err)
	}
	defer m.Close()

	switch cmd := os.Args[1]; cmd {
	case "up":
		run(m.Up, "şema güncel")
	case "down":
		run(func() error { return m.Steps(-1) }, "geri alınacak göç yok")
	case "down-all":
		run(m.Down, "geri alınacak göç yok")
	case "version":
		v, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			log.Println("henüz göç uygulanmamış")
			return
		}
		if err != nil {
			log.Fatalf("sürüm okunamadı: %v", err)
		}
		log.Printf("sürüm %d (kirli: %t)", v, dirty)
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("force bir sürüm numarası gerektirir")
		}
		v, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("geçersiz sürüm: %v", err)
		}
		if err := m.Force(v); err != nil {
			log.Fatalf("force başarısız: %v", err)
		}
		log.Printf("sürüm %d olarak sabitlendi", v)
	default:
		log.Fatalf("bilinmeyen komut: %s", cmd)
	}
}

// run, bir göç işlemini çalıştırır. ErrNoChange hata değildir.
func run(fn func() error, noChangeMsg string) {
	err := fn()
	if errors.Is(err, migrate.ErrNoChange) {
		log.Println(noChangeMsg)
		return
	}
	if err != nil {
		log.Fatalf("göç başarısız: %v", err)
	}
	log.Println("tamam")
}
