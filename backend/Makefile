.PHONY: migrate-up migrate-down-to-zero migrate-reset

include .env
export

DB_DSN=user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_DB) host=$(POSTGRES_HOST) port=$(POSTGRES_PORT) sslmode=disable

migrate-up:
	GOOSE_DRIVER=postgres \
	GOOSE_DBSTRING="$(DB_DSN)" \
	goose -dir migrations up
	
migrate-reset:
	@echo "⚠️  RESET DATABASE"
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$(DB_DSN)" goose -dir migrations reset
