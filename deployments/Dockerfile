# =============================================================================
# Build stage
# =============================================================================
FROM golang:1.25-bookworm AS builder

WORKDIR /app

# Зависимости для CGO и go-fitz
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарники
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/api ./cmd/api
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/worker ./cmd/worker

# =============================================================================
# API Runtime
# =============================================================================
FROM debian:bookworm-slim AS api

RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/migrations /app/migrations

# Непривилегированный пользователь
RUN useradd -r -s /bin/false appuser
USER appuser

EXPOSE 8080
CMD ["/app/api"]

# =============================================================================
# Worker Runtime
# =============================================================================
FROM debian:bookworm-slim AS worker

RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/bin/worker /app/worker

# Непривилегированный пользователь
RUN useradd -r -s /bin/false appuser
USER appuser

CMD ["/app/worker"]