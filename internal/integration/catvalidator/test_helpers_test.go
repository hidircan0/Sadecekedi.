package catvalidator

import (
	"bytes"
	"image"
	"image/jpeg"
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
