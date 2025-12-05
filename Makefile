.PHONY: help build run test clean docker-build docker-up docker-down docker-logs docker-clean

# Variables
APP_NAME=kero-kero
DOCKER_COMPOSE=docker-compose
GO=go

help: ## Mostrar esta ayuda
	@echo "Comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Compilar la aplicación
	$(GO) build -o server cmd/server/main.go

obfuscate: ## Compilar con ofuscación (requiere garble)
	go install mvdan.cc/garble@latest
	garble -literals -tiny build -o server cmd/server/main.go

run: ## Ejecutar la aplicación localmente
	$(GO) run cmd/server/main.go

test: ## Ejecutar tests
	$(GO) test -v ./...

test-coverage: ## Ejecutar tests con cobertura
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

clean: ## Limpiar archivos compilados
	rm -f server
	rm -f coverage.out coverage.html

deps: ## Descargar dependencias
	$(GO) mod download
	$(GO) mod tidy

# Docker commands
docker-build: ## Construir imagen Docker
	$(DOCKER_COMPOSE) build

docker-up: ## Iniciar servicios con Docker
	$(DOCKER_COMPOSE) up -d

docker-down: ## Detener servicios Docker
	$(DOCKER_COMPOSE) down

docker-restart: ## Reiniciar servicios Docker
	$(DOCKER_COMPOSE) restart

docker-logs: ## Ver logs de Docker
	$(DOCKER_COMPOSE) logs -f

docker-logs-api: ## Ver logs solo de la API
	$(DOCKER_COMPOSE) logs -f api

docker-ps: ## Ver estado de contenedores
	$(DOCKER_COMPOSE) ps

docker-clean: ## Limpiar contenedores y volúmenes (⚠️ BORRA DATOS)
	$(DOCKER_COMPOSE) down -v
	docker system prune -f

docker-rebuild: ## Reconstruir y reiniciar servicios
	$(DOCKER_COMPOSE) down
	$(DOCKER_COMPOSE) build --no-cache
	$(DOCKER_COMPOSE) up -d

# Database commands
db-migrate: ## Ejecutar migraciones (cuando se implementen)
	@echo "Migraciones pendientes de implementar"

db-backup: ## Backup de PostgreSQL
	$(DOCKER_COMPOSE) exec postgres pg_dump -U kerokero kerokero > backup_$$(date +%Y%m%d_%H%M%S).sql

db-restore: ## Restaurar backup (uso: make db-restore FILE=backup.sql)
	cat $(FILE) | $(DOCKER_COMPOSE) exec -T postgres psql -U kerokero kerokero

# Development
dev: ## Modo desarrollo con hot reload (requiere air)
	air

install-dev-tools: ## Instalar herramientas de desarrollo
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint: ## Ejecutar linter
	golangci-lint run

fmt: ## Formatear código
	$(GO) fmt ./...

# Production
deploy: docker-build docker-up ## Desplegar en producción

health-check: ## Verificar salud de la API
	@curl -s http://localhost:8080/health | jq .

# Información
version: ## Mostrar versión
	@echo "Kero-Kero WhatsApp API v2.0.0"

info: ## Mostrar información del sistema
	@echo "Go version: $$($(GO) version)"
	@echo "Docker version: $$(docker --version)"
	@echo "Docker Compose version: $$($(DOCKER_COMPOSE) version)"
