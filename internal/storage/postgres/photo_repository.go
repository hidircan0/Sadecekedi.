package postgres

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"cat-uploader-go/internal/app"
	"cat-uploader-go/internal/domain"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type PhotoRepository struct {
	db     *gorm.DB
	minio  *minio.Client
	bucket string
}

func NewPhotoRepository(db *gorm.DB, minioClient *minio.Client, bucket string) *PhotoRepository {
	return &PhotoRepository{
		db:     db,
		minio:  minioClient,
		bucket: bucket,
	}
}

func (r *PhotoRepository) Save(ctx context.Context, source io.Reader, originalName string) (domain.Photo, error) {
	safeName := sanitizeFileName(originalName)
	if safeName == "" {
		return domain.Photo{}, app.ErrInvalidFileType
	}

	if _, err := r.minio.StatObject(ctx, r.bucket, safeName, minio.StatObjectOptions{}); err == nil {
		return domain.Photo{}, app.ErrFileNameCollision
	} else {
		var minioErr minio.ErrorResponse
		if errors.As(err, &minioErr) && minioErr.Code != "NoSuchKey" && minioErr.Code != "NoSuchObject" {
			return domain.Photo{}, fmt.Errorf("check minio object collision: %w", err)
		}
	}

	content, err := io.ReadAll(source)
	if err != nil {
		return domain.Photo{}, fmt.Errorf("read upload payload: %w", err)
	}

	if _, err := r.minio.PutObject(
		ctx,
		r.bucket,
		safeName,
		bytes.NewReader(content),
		int64(len(content)),
		minio.PutObjectOptions{ContentType: "image/jpeg"},
	); err != nil {
		return domain.Photo{}, fmt.Errorf("upload object to minio: %w", err)
	}

	photo := domain.Photo{
		Filename:   safeName,
		UploadedAt: time.Now().UTC(),
		IsCat:      true,
	}

	if err := r.db.WithContext(ctx).Create(&photo).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.Photo{}, app.ErrFileNameCollision
		}
		return domain.Photo{}, fmt.Errorf("create photo record: %w", err)
	}

	return photo, nil
}

func (r *PhotoRepository) List(ctx context.Context) ([]domain.Photo, error) {
	photos := make([]domain.Photo, 0, 64)
	if err := r.db.WithContext(ctx).
		Order("id ASC").
		Find(&photos).Error; err != nil {
		return nil, fmt.Errorf("list photos from database: %w", err)
	}

	return photos, nil
}

func (r *PhotoRepository) GetByID(ctx context.Context, id uint) (domain.Photo, error) {
	var photo domain.Photo
	if err := r.db.WithContext(ctx).First(&photo, id).Error; err != nil {
		return domain.Photo{}, fmt.Errorf("get photo by id: %w", err)
	}
	return photo, nil
}

func (r *PhotoRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Photo{}, id).Error; err != nil {
		return fmt.Errorf("delete photo record: %w", err)
	}
	return nil
}

func (r *PhotoRepository) DeleteFromMinIO(ctx context.Context, filename string) error {
	if strings.TrimSpace(filename) == "" {
		return fmt.Errorf("delete minio object: empty filename")
	}
	if err := r.minio.RemoveObject(ctx, r.bucket, filename, minio.RemoveObjectOptions{}); err != nil {
		var minioErr minio.ErrorResponse
		if errors.As(err, &minioErr) && (minioErr.Code == "NoSuchKey" || minioErr.Code == "NoSuchObject") {
			return nil
		}
		return fmt.Errorf("remove object from minio: %w", err)
	}
	return nil
}

// 1. RASTGELE KEDİ BUTONU İÇİN
// Veritabanına "Bana şans eseri 1 tane kedi ver" der.
func (r *PhotoRepository) GetRandom(ctx context.Context) (domain.Photo, error) {
	var photo domain.Photo
	// Postgres'te verileri karıştırmak için RANDOM() kullanılır
	if err := r.db.WithContext(ctx).Order("RANDOM()").First(&photo).Error; err != nil {
		return domain.Photo{}, fmt.Errorf("get random photo: %w", err)
	}
	return photo, nil
}

// 2. SONRAKİ KEDİ BUTONU İÇİN (Geçmişe doğru kaydırma)
// Mevcut ID'den daha küçük olan ilk fotoğrafı getirir.
func (r *PhotoRepository) GetNext(ctx context.Context, currentID uint) (domain.Photo, error) {
	var photo domain.Photo
	if err := r.db.WithContext(ctx).
		Where("id < ?", currentID).
		Order("id DESC").
		First(&photo).Error; err != nil {
		return domain.Photo{}, fmt.Errorf("get next photo: %w", err)
	}
	return photo, nil
}

func (r *PhotoRepository) GetPrevious(ctx context.Context, currentID uint) (domain.Photo, error) {
	var photo domain.Photo
	if err := r.db.WithContext(ctx).
		Where("id > ?", currentID).
		Order("id ASC").
		First(&photo).Error; err != nil {
		return domain.Photo{}, fmt.Errorf("get previous photo: %w", err)
	}
	return photo, nil
}
func sanitizeFileName(fileName string) string {
	base := filepath.Base(fileName)
	base = strings.TrimSpace(base)
	base = strings.ReplaceAll(base, " ", "_")
	base = strings.ToLower(base)

	var b strings.Builder
	b.Grow(len(base))
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}

	clean := b.String()
	if clean == "" || strings.HasPrefix(clean, ".") {
		return ""
	}

	return clean
}
func (r *PhotoRepository) GetPhotoStream(ctx context.Context, filename string) (io.ReadCloser, string, error) {
    // MinIO'dan nesneyi al
    object, err := r.minio.GetObject(ctx, r.bucket, filename, minio.GetObjectOptions{})
    if err != nil {
        return nil, "", fmt.Errorf("get object from minio: %w", err)
    }

    // Dosyanın tipini (jpeg, png) öğren
    stat, err := object.Stat()
    if err != nil {
        object.Close()
        return nil, "", fmt.Errorf("stat minio object: %w", err)
    }

    return object, stat.ContentType, nil
}