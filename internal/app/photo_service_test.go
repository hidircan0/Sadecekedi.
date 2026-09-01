package app

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"testing"

	"cat-uploader-go/internal/domain"
)

type mockFile struct {
	*bytes.Reader
}

func (m *mockFile) Close() error { return nil }

func newMockJPEGFile() *mockFile {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		panic(err)
	}
	return &mockFile{Reader: bytes.NewReader(buf.Bytes())}
}

type mockRepo struct {
	saved bool
}

func (m *mockRepo) Save(_ context.Context, _ io.Reader, originalName string) (domain.Photo, error) {
	m.saved = true
	return domain.Photo{ID: 1, Filename: originalName}, nil
}

func (m *mockRepo) List(_ context.Context) ([]domain.Photo, error)         { return nil, nil }
func (m *mockRepo) GetByID(_ context.Context, _ uint) (domain.Photo, error) { return domain.Photo{}, nil }
func (m *mockRepo) Delete(_ context.Context, _ uint) error                  { return nil }
func (m *mockRepo) DeleteFromMinIO(_ context.Context, _ string) error        { return nil }
func (m *mockRepo) GetRandom(_ context.Context) (domain.Photo, error)        { return domain.Photo{}, nil }
func (m *mockRepo) GetNext(_ context.Context, _ uint) (domain.Photo, error)   { return domain.Photo{}, nil }
func (m *mockRepo) GetPrevious(_ context.Context, _ uint) (domain.Photo, error) {
	return domain.Photo{}, nil
}
func (m *mockRepo) GetPhotoStream(_ context.Context, _ string) (io.ReadCloser, string, error) {
	return nil, "", nil
}

type mockValidator struct {
	isCat bool
}

func (m *mockValidator) Validate(_ context.Context, _ multipart.File, _ string) (ValidationResult, error) {
	return ValidationResult{IsCat: m.isCat, Confidence: 0.99}, nil
}

func TestIsAllowedImageExtension(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     bool
	}{
		{name: "jpeg", fileName: "cat.jpeg", want: true},
		{name: "jpg", fileName: "cat.JPG", want: true},
		{name: "png", fileName: "cat.png", want: true},
		{name: "txt rejected", fileName: "payload.txt", want: false},
		{name: "empty extension", fileName: "cat", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAllowedImageExtension(tt.fileName); got != tt.want {
				t.Fatalf("isAllowedImageExtension(%q) = %v, want %v", tt.fileName, got, tt.want)
			}
		})
	}
}

func TestUpload_RejectsInvalidExtension(t *testing.T) {
	service := NewPhotoService(&mockRepo{}, &mockValidator{isCat: true})
	_, err := service.Upload(context.Background(), newMockJPEGFile(), "virus.exe")
	if err != ErrInvalidFileType {
		t.Fatalf("expected ErrInvalidFileType, got %v", err)
	}
}

func TestUpload_RejectsNonCat(t *testing.T) {
	service := NewPhotoService(&mockRepo{}, &mockValidator{isCat: false})
	_, err := service.Upload(context.Background(), newMockJPEGFile(), "cat.jpg")
	if err != ErrNotCat {
		t.Fatalf("expected ErrNotCat, got %v", err)
	}
}

func TestUpload_SavesValidCat(t *testing.T) {
	repo := &mockRepo{}
	service := NewPhotoService(repo, &mockValidator{isCat: true})

	photo, err := service.Upload(context.Background(), newMockJPEGFile(), "cat.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.saved {
		t.Fatal("expected repository Save to be called")
	}
	if photo.ID != 1 {
		t.Fatalf("expected photo ID 1, got %d", photo.ID)
	}
}
