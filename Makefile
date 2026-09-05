.PHONY: help setup up down backend frontend migrate migrate-down migrate-version db-reset db-shell seed score import-cosing import-obf test test-db-up test-db-down lint build clean

# Veritabanı Docker'da çalışır; yerel PostgreSQL kurulumuna gerek yoktur.
COMPOSE      := docker compose
COMPOSE_TEST := docker compose -f docker-compose.test.yml
TEST_DSN     := postgres://beauty:beauty@127.0.0.1:5434/beauty_test?sslmode=disable
PSQL    := $(COMPOSE) exec -T postgres psql -U beauty -d beauty_ingredient

help:
	@echo "Clarity - Komutlar"
	@echo "=================="
	@echo "make setup      - Veritabanını başlat ve tüm bağımlılıkları kur"
	@echo "make up         - PostgreSQL'i başlat (docker compose)"
	@echo "make down       - PostgreSQL'i durdur"
	@echo "make backend    - Go API'yi çalıştır  (http://localhost:8090)"
	@echo "make frontend   - Next.js uygulamasını çalıştır (http://localhost:3001)"
	@echo "make migrate    - Bekleyen şema göçlerini uygula"
	@echo "make migrate-down - Son göçü geri al"
	@echo "make db-shell   - Konteynerde psql kabuğu aç"
	@echo "make db-reset   - Volume'ü sil, şemayı ve örnek veriyi yeniden yükle (DİKKAT)"
	@echo "make seed       - Örnek veriyi yeniden uygula ve puanları türet (idempotent)"
	@echo "make score      - Puanları mevzuat verisinden yeniden türet"
	@echo "make import-cosing FILE=... ANNEX=III - CosIng dışa aktarımını içe al"
	@echo "make import-obf FILE=... LIMIT=5000 - Open Beauty Facts kataloğunu içe al"
	@echo "make test       - Backend testlerini çalıştır"
	@echo "make lint       - Go kodunu vet'le ve ön yüzü lint'le"
	@echo "make build      - API ikilisini ve Next.js paketini derle"
	@echo "make clean      - Derleme çıktılarını ve node_modules'ü sil"

setup: up
	@echo "Backend bağımlılıkları kuruluyor..."
	cd backend && go mod download
	@echo "Ön yüz bağımlılıkları kuruluyor..."
	cd web && npm install
	@echo ""
	@echo "Kurulum tamam. Ayrı terminallerde 'make backend' ve 'make frontend' çalıştırın."

up:
	$(COMPOSE) up -d
	@echo "PostgreSQL'in bağlantı kabul etmesi bekleniyor..."
	@$(COMPOSE) exec -T postgres sh -c 'until pg_isready -U beauty -d beauty_ingredient >/dev/null 2>&1; do sleep 1; done'
	@echo "Veritabanı hazır: localhost:5433"
	$(MAKE) migrate

down:
	$(COMPOSE) down

backend:
	@echo "API başlatılıyor: http://localhost:8090 ..."
	cd backend && go run .

frontend:
	@echo "Web uygulaması başlatılıyor: http://localhost:3001 ..."
	cd web && npm run dev

migrate:
	cd backend && go run ./cmd/migrate up

migrate-down:
	cd backend && go run ./cmd/migrate down

migrate-version:
	cd backend && go run ./cmd/migrate version

db-shell:
	$(COMPOSE) exec postgres psql -U beauty -d beauty_ingredient

# Göçler sayesinde şema değiştirmek için artık sıfırlama gerekmiyor;
# bu hedef yalnızca temiz bir başlangıç istendiğinde kullanılır.
db-reset:
	@echo "DİKKAT: veritabanı volume'ü ve içindeki tüm veri siliniyor..."
	$(COMPOSE) down -v
	$(MAKE) up
	$(MAKE) seed

# Seed puan yazmaz; yalnızca puanın dayanağını (ingredient_regulatory) kurar.
# Puanlar hemen ardından mevzuat verisinden türetilir.
seed:
	$(PSQL) < backend/db/seed.sql
	@echo "Örnek veri uygulandı"
	$(MAKE) score

score:
	cd backend && go run ./cmd/score

# CosIng dosyaları elle indirilir: https://ec.europa.eu/growth/tools-databases/cosing/
#   make import-cosing FILE=~/COSING_Annex_III_v2.csv ANNEX=III
import-cosing:
	@test -n "$(FILE)" || (echo "FILE gerekli: make import-cosing FILE=... ANNEX=III"; exit 1)
	cd backend && go run ./scripts/import-cosing -file "$(FILE)" -annex "$(ANNEX)" -report unmatched.csv
	$(MAKE) score

# Katalog verisi Open Beauty Facts'ten gelir (ODbL). Toplu dosya:
#   https://static.openbeautyfacts.org/data/openbeautyfacts-products.jsonl.gz
#   make import-obf FILE=~/openbeautyfacts-products.jsonl.gz LIMIT=5000
import-obf:
	@test -n "$(FILE)" || (echo "FILE gerekli: make import-obf FILE=... LIMIT=5000"; exit 1)
	cd backend && go run ./scripts/import-obf -file "$(FILE)" -limit $(or $(LIMIT),1000)
	$(MAKE) score

test: test-db-up
	cd backend && TEST_DATABASE_URL=$(TEST_DSN) go test ./... -count=1

test-db-up:
	$(COMPOSE_TEST) up -d
	@$(COMPOSE_TEST) exec -T postgres-test sh -c 'until pg_isready -U beauty -d beauty_test >/dev/null 2>&1; do sleep 1; done'

test-db-down:
	$(COMPOSE_TEST) down

lint:
	cd backend && go vet ./...
	cd web && npm run lint

build:
	cd backend && go build -o bin/api .
	cd web && npm run build

clean:
	rm -rf backend/bin backend/*.out backend/main
	rm -rf web/.next web/out web/dist web/node_modules
	@echo "Temizlik tamam"

.DEFAULT_GOAL := help
