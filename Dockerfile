# 1. Derleme Aşaması (Builder)
FROM golang:alpine AS builder
WORKDIR /app

# CA sertifikaları (Go binary'sinin MinIO/HTTPS istekleri için)
RUN apk add --no-cache ca-certificates

# Önce bağımlılıkları indir (Cache katmanı)
COPY go.mod go.sum ./
RUN go mod download

# Kodları kopyala ve statik olarak derle
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/web

# 2. Çalıştırma Aşaması (Final Image)
FROM alpine:latest
WORKDIR /root/

# HTTPS çağrıları için root sertifikalarını builder'dan aktar
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Binary ve web varlıklarını al
COPY --from=builder /app/main .
COPY --from=builder /app/web ./web

RUN mkdir -p uploads

EXPOSE 8080
CMD ["./main"]