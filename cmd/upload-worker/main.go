package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cat-uploader-go/internal/app"
	"cat-uploader-go/internal/integration/catvalidator"
	"cat-uploader-go/internal/storage/postgres"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	kafkago "github.com/segmentio/kafka-go"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn, brokers := os.Getenv("DATABASE_URL"), splitEnv("KAFKA_BROKERS", "kafka:9092")
	if dsn == "" { log.Fatal("DATABASE_URL is required") }
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil { log.Fatalf("connect postgres: %v", err) }
	mc, err := minio.New(env("MINIO_ENDPOINT", "minio:9000"), &minio.Options{Creds: credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), ""), Secure: os.Getenv("MINIO_SECURE") == "true", Region:"us-east-1", BucketLookup:minio.BucketLookupPath})
	if err != nil { log.Fatalf("create minio client: %v", err) }
	repo := postgres.NewPhotoRepository(db, mc, env("MINIO_BUCKET", "cats"))
	validator := catvalidator.NewHTTPClient(env("CAT_VALIDATOR_URL", "http://validator:8000"), &http.Client{Timeout: 7*time.Second})
	service := app.NewPhotoService(repo, validator)
	reader := kafkago.NewReader(kafkago.ReaderConfig{Brokers: brokers, Topic:env("KAFKA_TOPIC", "photo-uploads"), GroupID:"sadecekedi-upload-workers", MinBytes:1, MaxBytes:10e6, CommitInterval:0})
	defer reader.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil { return }
			log.Printf("read Kafka message: %v", err); time.Sleep(time.Second); continue
		}
		jobID := string(msg.Value)
		for {
			err := service.ProcessQueuedUpload(ctx, jobID)
			if err == nil { break }
			log.Printf("process upload job %s; retrying: %v", jobID, err)
			select { case <-ctx.Done(): return; case <-time.After(time.Second): }
		}
		if err := reader.CommitMessages(ctx, msg); err != nil { log.Printf("commit upload job %s: %v", jobID, err) }
	}
}

func env(key, fallback string) string { if value:=os.Getenv(key); value!="" { return value }; return fallback }
func splitEnv(key, fallback string) []string { value:=env(key, fallback); parts:=strings.Split(value, ","); out:=parts[:0]; for _, p:=range parts { if p=strings.TrimSpace(p); p!="" { out=append(out,p) } }; return out }

