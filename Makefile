.PHONY: swagger swagger-check build test run migrate-up migrate-down-to-zero migrate-reset

include .env
export

DB_DSN=user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_DB) host=$(POSTGRES_HOST) port=$(POSTGRES_PORT) sslmode=disable

migrate-up:
	GOOSE_DRIVER=postgres \
	GOOSE_DBSTRING="$(DB_DSN)" \
	goose -dir migrations up

migrate-reset:
	@echo "⚠️  RESET DATABASE"
	GOOSE_DRIVER=postgres 
	GOOSE_DBSTRING="$(DB_DSN)" \
	goose -dir migrations reset

# Генерация Swagger документации
swagger:
	@echo "📚 Generating Swagger documentation..."
	@cd backend && swag init -g cmd/main.go -o docs --parseDependency --parseInternal
	@echo "✅ Swagger docs generated in backend/docs/"

# Проверка актуальности документации
swagger-check:
	@echo "🔍 Checking Swagger documentation..."
	@cd backend && swag init -g cmd/main.go -o docs --parseDependency --parseInternal
	@if [ -n "$$(git status --porcelain backend/docs/)" ]; then \
		echo "❌ Swagger documentation is out of date!"; \
		exit 1; \
	else \
		echo "✅ Swagger documentation is up to date"; \
	fi

# Сборка backend
build:
	@cd backend && go build -o bin/server ./cmd/main.go

# Запуск тестов
test:
	@cd backend && go test ./... -v

# Запуск сервера
run:
	@cd backend && go run cmd/main.go

# Установка swag CLI
install-swag:
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ swag installed"

# Помощь
help:
	@echo "Available commands:"
	@echo "  make swagger       - Generate Swagger documentation"
	@echo "  make swagger-check - Check if Swagger docs are up to date"
	@echo "  make build         - Build the backend"
	@echo "  make test          - Run tests"
	@echo "  make run           - Run the server"
	@echo "  make install-swag  - Install swag CLI tool"
