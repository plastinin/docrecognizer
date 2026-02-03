# =============================================================================
# Build stage
# =============================================================================
FROM golang:1.25-alpine AS builder

# Устанавливаем зависимости для сборки (включая CGO для go-fitz)
RUN apk add --no-cache \
    gcc \
    musl-dev \
    mupdf-dev \
    mupdf-tools

WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарники
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/api ./cmd/api
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/worker ./cmd/worker

# =============================================================================
# Runtime stage
# =============================================================================
FROM alpine:3.19

# Устанавливаем runtime зависимости
RUN apk add --no-cache \
    ca-certificates \
    mupdf \
    tzdata

WORKDIR /app

# Копируем бинарники из builder
COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/bin/worker /app/worker

# Копируем миграции
COPY --from=builder /app/migrations /app/migrations

# Создаём непривилегированного пользователя
RUN adduser -D -g '' appuser
USER appuser

# По умолчанию запускаем API
CMD ["/app/api"]