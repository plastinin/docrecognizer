.PHONY: build run-api run-worker test lint migrate-up migrate-down \
        docker-build docker-up docker-down infra-up infra-down ollama-pull

# Go параметры
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

# Бинарники
API_BINARY=bin/api
WORKER_BINARY=bin/worker

# =============================================================================
# Build
# =============================================================================

## build: Собрать все бинарники
build:
	$(GOBUILD) -o $(API_BINARY) ./cmd/api
	$(GOBUILD) -o $(WORKER_BINARY) ./cmd/worker

## run-api: Запустить API сервер
run-api:
	$(GOCMD) run ./cmd/api

## run-worker: Запустить воркер
run-worker:
	$(GOCMD) run ./cmd/worker

## test: Запустить тесты
test:
	$(GOTEST) -v ./...

## tidy: Обновить зависимости
tidy:
	$(GOMOD) tidy

# =============================================================================
# Database
# =============================================================================

## migrate-up: Применить миграции
migrate-up:
	migrate -path migrations -database "postgres://docrecognizer:secret@localhost:5432/docrecognizer?sslmode=disable" up

## migrate-down: Откатить миграции
migrate-down:
	migrate -path migrations -database "postgres://docrecognizer:secret@localhost:5432/docrecognizer?sslmode=disable" down

## migrate-create: Создать новую миграцию (usage: make migrate-create name=migration_name)
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# =============================================================================
# Docker - Full stack
# =============================================================================

## docker-build: Собрать Docker образ
docker-build:
	docker build -f deployments/Dockerfile -t docrecognizer:latest .

## docker-up: Запустить весь стек через docker-compose
docker-up:
	docker compose -f deployments/docker-compose.yml up -d

## docker-down: Остановить docker-compose
docker-down:
	docker compose -f deployments/docker-compose.yml down

## docker-logs: Показать логи
docker-logs:
	docker compose -f deployments/docker-compose.yml logs -f

## docker-logs-api: Показать логи API
docker-logs-api:
	docker compose -f deployments/docker-compose.yml logs -f api

## docker-logs-worker: Показать логи воркера
docker-logs-worker:
	docker compose -f deployments/docker-compose.yml logs -f worker

# =============================================================================
# Docker - Infrastructure only (for local development)
# =============================================================================

## infra-up: Запустить только инфраструктуру (postgres, redis, minio, ollama)
infra-up:
	docker compose -f deployments/docker-compose.infra.yml up -d

## infra-down: Остановить инфраструктуру
infra-down:
	docker compose -f deployments/docker-compose.infra.yml down

## infra-logs: Показать логи инфраструктуры
infra-logs:
	docker compose -f deployments/docker-compose.infra.yml logs -f

# =============================================================================
# Ollama
# =============================================================================

## ollama-pull: Скачать модель 
ollama-pull:
	docker exec -it docrecognizer-ollama ollama pull qwen3-vl

## ollama-list: Показать список моделей
ollama-list:
	docker exec -it docrecognizer-ollama ollama list

# =============================================================================
# Development
# =============================================================================

## dev: Запустить инфраструктуру + миграции + API + Worker
dev: infra-up
	@echo "Waiting for services to start..."
	@sleep 5
	@make migrate-up
	@echo "Starting API and Worker..."
	@make run-api &
	@make run-worker

## help: Показать справку
help:
	@echo "Доступные команды:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'