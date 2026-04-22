package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cat-uploader-go/internal/app"
	"cat-uploader-go/internal/domain"
	"cat-uploader-go/internal/http/router"
	"cat-uploader-go/internal/integration/catvalidator"
	"cat-uploader-go/internal/storage/postgres"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
    addr := ":8080"
    
    // Tehlikesiz olanlara varsayılan (fallback) değer bırakabilirsin
    validatorURL := getenvOrDefault("CAT_VALIDATOR_URL", "http://host.docker.internal:8000")    
    minioEndpoint := getenvOrDefault("MINIO_ENDPOINT", "localhost:9000")
    minioBucket := getenvOrDefault("MINIO_BUCKET", "cats")

    // ŞİFRELER! Fallback kısımlarını tamamen siliyoruz (Boş bırakıyoruz)
    dsn := getenvOrDefault("DATABASE_URL", "") 
    minioAccessKey := getenvOrDefault("MINIO_ACCESS_KEY", "")
    minioSecretKey := getenvOrDefault("MINIO_SECRET_KEY", "")

    adminUser := getenvOrDefault("ADMIN_USER", "")
    adminPass := getenvOrDefault("ADMIN_PASS", "")

    // GÜVENLİK DUVARI (Fail-Fast Koruması)
    // Eğer env dosyasından şifreler gelmezse, uydurma şifreyle çalışmaya kalkmasın, direkt sistemi durdursun!
    if dsn == "" || minioSecretKey == "" || adminUser == "" || adminPass == "" {
        log.Fatal("KRITIK SIZINTI ÖNLENDİ: Veritabanı, MinIO veya Admin şifresi eksik! Sunucu durduruluyor. Lütfen .env dosyanı kontrol et.")
    }

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	if err := db.AutoMigrate(&domain.Photo{}); err != nil {
		log.Fatalf("failed to migrate photos table: %v", err)
	}

	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("failed to create minio client: %v", err)
	}

	if err := ensureBucketReady(context.Background(), minioClient, minioBucket); err != nil {
		log.Fatalf("failed to prepare minio bucket: %v", err)
	}

	repo := postgres.NewPhotoRepository(db, minioClient, minioBucket)
	validator := catvalidator.NewHTTPClient(
		validatorURL,
		&http.Client{Timeout: 5 * time.Second},
	)
	service := app.NewPhotoService(repo, validator)
	handler := router.NewHandlerSet(service)
	r := router.New(handler)

	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}

func getenvOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func ensureBucketReady(ctx context.Context, client *minio.Client, bucket string) error {
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("check bucket existence: %w", err)
	}

	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
	}

	policy := fmt.Sprintf(`{
		"Version":"2012-10-17",
		"Statement":[
			{
				"Effect":"Allow",
				"Principal":{"AWS":["*"]},
				"Action":["s3:GetObject"],
				"Resource":["arn:aws:s3:::%s/*"]
			}
		]
	}`, bucket)

	if err := client.SetBucketPolicy(ctx, bucket, policy); err != nil {
		return fmt.Errorf("set bucket public policy: %w", err)
	}

	return nil
}
