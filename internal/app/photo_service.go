package app

import (
	"bytes"
	"context"
	"errors"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"cat-uploader-go/internal/domain"
	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp"
)

var (
	ErrEmptyFile         = errors.New("file is required")
	ErrInvalidFileType   = errors.New("only image files are accepted")
	ErrFileNameCollision = errors.New("file name already exists")
	ErrNotCat            = errors.New("Bu bir kedi değil!")
)

// 1. KABLO BAĞLANTISI: Veritabanından neleri isteyebileceğimizi güncelledik
type PhotoRepository interface {
	Save(ctx context.Context, source io.Reader, originalName string) (domain.Photo, error)
	List(ctx context.Context) ([]domain.Photo, error)
	GetByID(ctx context.Context, id uint) (domain.Photo, error)
	Delete(ctx context.Context, id uint) error
	DeleteFromMinIO(ctx context.Context, filename string) error
	// SADECEKEDİ: Yeni yetenekler eklendi
	GetRandom(ctx context.Context) (domain.Photo, error)
	GetNext(ctx context.Context, id uint) (domain.Photo, error)
	GetPrevious(ctx context.Context, id uint) (domain.Photo, error)
	GetPhotoStream(ctx context.Context, filename string) (io.ReadCloser, string, error)
}

type ValidationResult struct {
	IsCat      bool
	Confidence float64
}

type CatValidator interface {
	Validate(ctx context.Context, source multipart.File, originalName string) (ValidationResult, error)
}

type PhotoService struct {
	repo      PhotoRepository
	validator CatValidator
}

func NewPhotoService(repo PhotoRepository, validator CatValidator) *PhotoService {
	return &PhotoService{repo: repo, validator: validator}
}

func (s *PhotoService) Upload(ctx context.Context, source multipart.File, originalName string) (domain.Photo, error) {
	if source == nil || strings.TrimSpace(originalName) == "" {
		return domain.Photo{}, ErrEmptyFile
	}

	if !isAllowedImageExtension(originalName) {
		return domain.Photo{}, ErrInvalidFileType
	}

	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return domain.Photo{}, err
	}

	validation, err := s.validator.Validate(ctx, source, originalName)
	if err != nil {
		return domain.Photo{}, err
	}
	if !validation.IsCat {
		return domain.Photo{}, ErrNotCat
	}

	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return domain.Photo{}, err
	}

	optimizedImage, optimizedName, err := optimizeImageForStorage(source, originalName)
	if err != nil {
		return domain.Photo{}, err
	}

	return s.repo.Save(ctx, bytes.NewReader(optimizedImage), optimizedName)
}

func (s *PhotoService) List(ctx context.Context) ([]domain.Photo, error) {
	return s.repo.List(ctx)
}

// 2. KABLO BAĞLANTISI: Handler'dan gelen istekleri doğrudan veritabanına iletiyoruz
func (s *PhotoService) GetRandom(ctx context.Context) (domain.Photo, error) {
	return s.repo.GetRandom(ctx)
}

func (s *PhotoService) GetNext(ctx context.Context, id uint) (domain.Photo, error) {
	return s.repo.GetNext(ctx, id)
}

func (s *PhotoService) GetPrevious(ctx context.Context, id uint) (domain.Photo, error) {
	return s.repo.GetPrevious(ctx, id)
}

func (s *PhotoService) Delete(ctx context.Context, id uint) error {
	photo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteFromMinIO(ctx, photo.Filename); err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

func isAllowedImageExtension(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func optimizeImageForStorage(source multipart.File, originalName string) ([]byte, string, error) {
	decodedImage, _, err := image.Decode(source)
	if err != nil {
		return nil, "", ErrInvalidFileType
	}

	resizedImage := resize.Resize(800, 0, decodedImage, resize.Lanczos3)

	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, resizedImage, &jpeg.Options{Quality: 80}); err != nil {
		return nil, "", err
	}

	outputName := strings.TrimSuffix(originalName, filepath.Ext(originalName)) + ".jpg"
	return buffer.Bytes(), outputName, nil
}
func (s *PhotoService) GetPhotoStream(ctx context.Context, filename string) (io.ReadCloser, string, error) {
    return s.repo.GetPhotoStream(ctx, filename)
}