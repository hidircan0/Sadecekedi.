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
	"github.com/google/uuid"

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
	return r.save(ctx, source, originalName, nil)
}

func (r *PhotoRepository) SaveForJob(ctx context.Context, source io.Reader, originalName, jobID string) (domain.Photo, error) {
	return r.save(ctx, source, originalName, &jobID)
}

func (r *PhotoRepository) save(ctx context.Context, source io.Reader, originalName string, jobID *string) (domain.Photo, error) {
	if jobID != nil {
		var existing domain.Photo
		if err := r.db.WithContext(ctx).Where("upload_job_id = ?", *jobID).First(&existing).Error; err == nil { return existing, nil } else if !errors.Is(err, gorm.ErrRecordNotFound) { return domain.Photo{}, fmt.Errorf("find photo for upload job: %w", err) }
	}
	safeName := sanitizeFileName(originalName)
	if safeName == "" {
		return domain.Photo{}, app.ErrInvalidFileType
	}

	if _, err := r.minio.StatObject(ctx, r.bucket, safeName, minio.StatObjectOptions{}); err == nil {
		if jobID == nil {
			return domain.Photo{}, app.ErrFileNameCollision
		}
		var existing domain.Photo
		if dbErr := r.db.WithContext(ctx).Where("filename = ?", safeName).First(&existing).Error; dbErr == nil {
			return domain.Photo{}, app.ErrFileNameCollision
		} else if !errors.Is(dbErr, gorm.ErrRecordNotFound) {
			return domain.Photo{}, fmt.Errorf("check photo filename collision: %w", dbErr)
		}
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
		UploadJobID: jobID,
	}

	if err := r.db.WithContext(ctx).Create(&photo).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.Photo{}, app.ErrFileNameCollision
		}
		return domain.Photo{}, fmt.Errorf("create photo record: %w", err)
	}

	return photo, nil
}

func (r *PhotoRepository) EnqueueUpload(ctx context.Context, source io.Reader, originalName, jobID string) (domain.UploadJob, error) {
	if _, err := uuid.Parse(jobID); err != nil { return domain.UploadJob{}, fmt.Errorf("invalid upload job id: %w", err) }
	content, err := io.ReadAll(source)
	if err != nil { return domain.UploadJob{}, fmt.Errorf("read staged upload: %w", err) }
	key := stagingKey(jobID)
	filename := sanitizeFileName(originalName)
	if filename == "" { return domain.UploadJob{}, app.ErrInvalidFileType }
	if _, err := r.minio.PutObject(ctx, r.bucket, key, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{ContentType: "application/octet-stream"}); err != nil {
		return domain.UploadJob{}, fmt.Errorf("stage upload in object storage: %w", err)
	}
	job := domain.UploadJob{ID: jobID, Filename: filename, Status: "queued"}
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&job).Error; err != nil { return err }
		return tx.Create(&domain.OutboxEvent{JobID: jobID}).Error
	})
	if err != nil {
		_ = r.minio.RemoveObject(context.Background(), r.bucket, key, minio.RemoveObjectOptions{})
		return domain.UploadJob{}, fmt.Errorf("create upload job: %w", err)
	}
	return job, nil
}

func stagingKey(id string) string { return "staging/" + id }

func (r *PhotoRepository) GetUploadJob(ctx context.Context, id string) (domain.UploadJob, error) {
	var job domain.UploadJob
	if err := r.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil { return job, fmt.Errorf("get upload job: %w", err) }
	return job, nil
}

func (r *PhotoRepository) GetStagedUpload(ctx context.Context, id string) ([]byte, string, error) {
	job, err := r.GetUploadJob(ctx, id); if err != nil { return nil, "", err }
	obj, err := r.minio.GetObject(ctx, r.bucket, stagingKey(id), minio.GetObjectOptions{}); if err != nil { return nil, "", err }
	defer obj.Close()
	content, err := io.ReadAll(obj); if err != nil { return nil, "", fmt.Errorf("read staged object: %w", err) }
	return content, job.Filename, nil
}

func (r *PhotoRepository) SetUploadJobResult(ctx context.Context, id string, photo *domain.Photo, jobErr error) error {
	updates := map[string]interface{}{"status": "succeeded", "error": "", "photo_id": nil}
	if jobErr != nil {
		message := jobErr.Error()
		if len(message) > 500 { message = message[:500] }
		updates["status"] = "failed"
		updates["error"] = message
	}
	if photo != nil { updates["photo_id"] = photo.ID }
	return r.db.WithContext(ctx).Model(&domain.UploadJob{}).Where("id = ?", id).Updates(updates).Error
}

func (r *PhotoRepository) MarkUploadJobProcessing(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.UploadJob{}).Where("id = ? AND status <> ?", id, "succeeded").Update("status", "processing").Error
}

func (r *PhotoRepository) DeleteStagedUpload(ctx context.Context, id string) error {
	return r.minio.RemoveObject(ctx, r.bucket, stagingKey(id), minio.RemoveObjectOptions{})
}

func (r *PhotoRepository) PendingOutbox(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	var events []domain.OutboxEvent
	err := r.db.WithContext(ctx).Where("published_at IS NULL").Order("id ASC").Limit(limit).Find(&events).Error
	return events, err
}

func (r *PhotoRepository) MarkOutboxPublished(ctx context.Context, id uint) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&domain.OutboxEvent{}).Where("id = ? AND published_at IS NULL", id).Update("published_at", now).Error
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
