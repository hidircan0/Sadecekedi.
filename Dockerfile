# 1. Derleme Aşaması (Builder)
FROM golang:alpine AS builder
WORKDIR /app

# Önce sadece bağımlılıkları kopyala ve indir (Docker önbelleğini efektif kullanmak için)
COPY go.mod go.sum ./
RUN go mod download

# Tüm kodları ve klasörleri kopyala
COPY . .

# cmd/web içindeki ana uygulamayı derle
RUN go build -o main ./cmd/web

# 2. Çalıştırma Aşaması (Final Image)
FROM alpine:latest
WORKDIR /root/

# Derlenen çalıştırılabilir ana dosyayı al
COPY --from=builder /app/main .

# İŞTE BÜYÜK ÇÖZÜM: Bütün "web" klasörünü (static, templates ve partials ile birlikte) olduğu gibi kopyala
COPY --from=builder /app/web ./web

# Fotoğrafların kaydedileceği klasörü oluştur
RUN mkdir -p uploads

EXPOSE 8080
CMD ["./main"]