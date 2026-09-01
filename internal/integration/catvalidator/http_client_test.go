package catvalidator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/validate" {
			t.Fatalf("expected /validate, got %s", r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parse multipart form: %v", err)
		}
		if _, ok := r.MultipartForm.File["image"]; !ok {
			t.Fatal("expected image field in multipart form")
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"is_cat":     true,
			"confidence": 0.95,
		})
	}))
	t.Cleanup(server.Close)

	client := NewHTTPClient(server.URL, server.Client())
	result, err := client.Validate(context.Background(), newMockJPEGFile(), "cat.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsCat {
		t.Fatal("expected is_cat=true")
	}
	if result.Confidence != 0.95 {
		t.Fatalf("expected confidence 0.95, got %f", result.Confidence)
	}
}

func TestValidate_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)

	client := NewHTTPClient(server.URL, server.Client())
	_, err := client.Validate(context.Background(), newMockJPEGFile(), "cat.jpg")
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}
