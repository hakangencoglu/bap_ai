# Multi-stage build
# Stage 1: Builder
FROM golang:1.25-alpine AS builder

# Gerekli araçları yükle
RUN apk add --no-cache git

# Çalışma dizini oluştur
WORKDIR /app

# Bağımlılıkları kopyala ve yükle
COPY go.mod go.sum ./
RUN go mod download

# Kaynak kodları kopyala
COPY . .

# Uygulamayı derle (CGO_ENABLED=0 statik binary için önemlidir)
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/main.go

# Stage 2: Runner
FROM alpine:latest

# SSL sertifikaları ve health check için wget yükle
RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

# Builder aşamasından kopyalamalar
COPY --from=builder /app/main .
COPY --from=builder /app/frontend/templates ./frontend/templates
COPY --from=builder /app/frontend/static ./frontend/static
COPY --from=builder /app/migrations ./migrations

# Uygulamanın çalışacağı port
EXPOSE 8080

# Konteyner sağlık kontrolü
HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=15s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Uygulamayı başlat
CMD ["./main"]
