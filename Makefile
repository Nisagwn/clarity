.PHONY: help setup up down backend frontend db db-reset db-shell seed test lint build clean

# Veritabanı Docker'da çalışır; yerel PostgreSQL kurulumuna gerek yoktur.
COMPOSE := docker compose
PSQL    := $(COMPOSE) exec -T postgres psql -U beauty -d beauty_ingredient

help:
	@echo "Clarity - Komutlar"
	@echo "=================="
	@echo "make setup      - Veritabanını başlat ve tüm bağımlılıkları kur"
	@echo "make up         - PostgreSQL'i başlat (docker compose)"
	@echo "make down       - PostgreSQL'i durdur"
	@echo "make backend    - Go API'yi çalıştır  (http://localhost:8090)"
	@echo "make frontend   - Next.js uygulamasını çalıştır (http://localhost:3001)"
	@echo "make db-shell   - Konteynerde psql kabuğu aç"
	@echo "make db-reset   - Volume'ü sil, şemayı ve örnek veriyi yeniden yükle (DİKKAT)"
	@echo "make seed       - Örnek veriyi yeniden uygula (idempotent)"
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

down:
	$(COMPOSE) down

backend:
	@echo "API başlatılıyor: http://localhost:8090 ..."
	cd backend && go run .

frontend:
	@echo "Web uygulaması başlatılıyor: http://localhost:3001 ..."
	cd web && npm run dev

db-shell:
	$(COMPOSE) exec postgres psql -U beauty -d beauty_ingredient

# Şema ve örnek veri yalnızca volume ilk oluşturulduğunda çalışır; bu yüzden
# sıfırlamak volume'ü silmek demektir.
db-reset:
	@echo "DİKKAT: veritabanı volume'ü ve içindeki tüm veri siliniyor..."
	$(COMPOSE) down -v
	$(MAKE) up

seed:
	$(PSQL) < backend/db/seed.sql
	@echo "Örnek veri uygulandı"

test:
	cd backend && go test ./...

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
