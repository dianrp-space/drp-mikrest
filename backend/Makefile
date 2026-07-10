DB_DSN ?= postgres://postgres:34267793@localhost:5432/drp_mikrest?sslmode=disable

.PHONY: tidy run dev build vet test migrate-new

tidy:
	go mod tidy

run:
	go run ./cmd/api

dev:
	@go run github.com/air-verse/air@latest

build:
	go build -o bin/api ./cmd/api

vet:
	go vet ./...

test:
	go test ./... -v

# Buat file migrasi baru: make migrate-new name=add_index
migrate-new:
	@test -n "$(name)" || (echo "usage: make migrate-new name=add_index" && exit 1)
	@ts=$$(date +%Y%m%d%H%M%S); \
	echo "creating migrations/$${ts}_$(name).up.sql & .down.sql"; \
	touch migrations/$${ts}_$(name).up.sql; \
	touch migrations/$${ts}_$(name).down.sql
