package catvalidator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"cat-uploader-go/internal/app"
)

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

type validateResponse struct {
	IsCat      bool    `json:"is_cat"`
	Confidence float64 `json:"confidence"`
}

func NewHTTPClient(baseURL string, httpClient *http.Client) *HTTPClient {
	return &HTTPClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *HTTPClient) Validate(ctx context.Context, source multipart.File, originalName string) (app.ValidationResult, error) {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("image", filepath.Base(originalName))
	if err != nil {
		return app.ValidationResult{}, fmt.Errorf("create multipart field: %w", err)
	}
	if _, err := io.Copy(part, source); err != nil {
		return app.ValidationResult{}, fmt.Errorf("copy image payload: %w", err)
	}
	if err := writer.Close(); err != nil {
		return app.ValidationResult{}, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/validate", &requestBody)
	if err != nil {
		return app.ValidationResult{}, fmt.Errorf("build validation request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return app.ValidationResult{}, fmt.Errorf("request validator service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return app.ValidationResult{}, fmt.Errorf("validator service returned status %d", resp.StatusCode)
	}

	var parsed validateResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return app.ValidationResult{}, fmt.Errorf("decode validator response: %w", err)
	}

	return app.ValidationResult{
		IsCat:      parsed.IsCat,
		Confidence: parsed.Confidence,
	}, nil
}
