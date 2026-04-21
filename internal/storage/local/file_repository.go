package local

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"cat-uploader-go/internal/app"
	"cat-uploader-go/internal/domain"
)

type FileRepository struct {
	uploadDir string
}

func NewFileRepository(uploadDir string) *FileRepository {
	return &FileRepository{uploadDir: uploadDir}
}

func (r *FileRepository) Save(_ context.Context, source multipart.File, originalName string) (domain.Photo, error) {
	safeName := sanitizeFileName(originalName)
	if safeName == "" {
		return domain.Photo{}, app.ErrInvalidFileType
	}

	targetPath := filepath.Join(r.uploadDir, safeName)
	if _, err := os.Stat(targetPath); err == nil {
		return domain.Photo{}, app.ErrFileNameCollision
	}

	dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return domain.Photo{}, fmt.Errorf("create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, source); err != nil {
		return domain.Photo{}, fmt.Errorf("copy uploaded file: %w", err)
	}

	return domain.Photo{
		Filename:   safeName,
		UploadedAt: time.Now(),
		IsCat:      true,
	}, nil
}

func (r *FileRepository) List(_ context.Context) ([]domain.Photo, error) {
	entries, err := os.ReadDir(r.uploadDir)
	if err != nil {
		return nil, fmt.Errorf("read upload directory: %w", err)
	}

	photos := make([]domain.Photo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !isAllowedImage(name) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("read file info: %w", err)
		}

		photos = append(photos, domain.Photo{
			Filename:   name,
			UploadedAt: info.ModTime(),
			IsCat:      true,
		})
	}

	sort.Slice(photos, func(i, j int) bool {
		return photos[i].UploadedAt.After(photos[j].UploadedAt)
	})

	return photos, nil
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

func isAllowedImage(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func IsDomainError(err error) bool {
	return errors.Is(err, app.ErrInvalidFileType) || errors.Is(err, app.ErrFileNameCollision)
}
