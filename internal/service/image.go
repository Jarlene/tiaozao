package service

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type ImageService struct {
	uploadDir string
	baseURL   string
}

func NewImageService(uploadDir, baseURL string) *ImageService {
	os.MkdirAll(uploadDir, 0755)
	return &ImageService{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

func (s *ImageService) Save(ctx context.Context, filename string, reader io.Reader) (string, error) {
	subDir := filepath.Dir(filename)
	if subDir != "." {
		os.MkdirAll(filepath.Join(s.uploadDir, subDir), 0755)
	}

	dst := filepath.Join(s.uploadDir, filename)
	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return "", err
	}

	return s.baseURL + "/" + filename, nil
}
