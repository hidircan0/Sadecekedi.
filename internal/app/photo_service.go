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
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp"
)

var (
	ErrEmptyFile         = errors.New("Bir fotoğraf yükleyin")
	ErrInvalidFileType   = errors.New("Sadece fotoğraf yükleyebilirsiniz")
	ErrFileNameCollision = errors.New("Bu kei hali hazırda sitede kayıtlı!")
	ErrNotCat            = errors.New("Bu bir kedi değil!")
)

// 1. KABLO BAĞLANTISI: Veritabanından neleri isteyebileceğimizi güncelledik
type PhotoRepository interface {
	EnqueueUpload(ctx context.Context, source io.Reader, originalName string, jobID string) (domain.UploadJob, error)
	GetUploadJob(ctx context.Context, jobID string) (domain.UploadJob, error)
	GetStagedUpload(ctx context.Context, jobID string) ([]byte, string, error)
	SetUploadJobResult(ctx context.Context, jobID string, photo *domain.Photo, jobErr error) error
	MarkUploadJobProcessing(ctx context.Context, jobID string) error
	DeleteStagedUpload(ctx context.Context, jobID string) error
	SaveForJob(ctx context.Context, source io.Reader, originalName string, jobID string) (domain.Photo, error)
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
	return s.upload(ctx, source, originalName, "")
}

func (s *PhotoService) upload(ctx context.Context, source multipart.File, originalName, jobID string) (domain.Photo, error) {
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

	if jobID != "" { return s.repo.SaveForJob(ctx, bytes.NewReader(optimizedImage), optimizedName, jobID) }
	return s.repo.Save(ctx, bytes.NewReader(optimizedImage), optimizedName)
}

func (s *PhotoService) QueueUpload(ctx context.Context, source multipart.File, originalName string) (domain.UploadJob, error) {
	if source == nil || strings.TrimSpace(originalName) == "" {
		return domain.UploadJob{}, ErrEmptyFile
	}
	if !isAllowedImageExtension(originalName) {
		return domain.UploadJob{}, ErrInvalidFileType
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return domain.UploadJob{}, err
	}
	return s.repo.EnqueueUpload(ctx, source, filepath.Base(originalName), uuid.NewString())
}

func (s *PhotoService) UploadJob(ctx context.Context, id string) (domain.UploadJob, error) {
	return s.repo.GetUploadJob(ctx, id)
}

// ProcessQueuedUpload validates and persists a staged upload. The repository
// associates the photo with the job so Kafka redelivery is idempotent.
func (s *PhotoService) ProcessQueuedUpload(ctx context.Context, id string) error {
	job, err := s.repo.GetUploadJob(ctx, id)
	if err != nil { return err }
	if job.Status == "succeeded" || job.Status == "failed" { return nil }
	if err := s.repo.MarkUploadJobProcessing(ctx, id); err != nil { return err }
	content, name, err := s.repo.GetStagedUpload(ctx, id)
	if err != nil {
		if updateErr := s.repo.SetUploadJobResult(ctx, id, nil, err); updateErr != nil { return errors.Join(err, updateErr) }
		return nil
	}
	file := &memoryMultipartFile{Reader: bytes.NewReader(content)}
	photo, err := s.upload(ctx, file, name, id)
	if err != nil {
		if updateErr := s.repo.SetUploadJobResult(ctx, id, nil, err); updateErr != nil { return errors.Join(err, updateErr) }
		_ = s.repo.DeleteStagedUpload(ctx, id)
		return nil
	}
	if err := s.repo.SetUploadJobResult(ctx, id, &photo, nil); err != nil { return err }
	_ = s.repo.DeleteStagedUpload(ctx, id)
	return nil
}

type memoryMultipartFile struct {
	*bytes.Reader
}

func (*memoryMultipartFile) Close() error { return nil }

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
