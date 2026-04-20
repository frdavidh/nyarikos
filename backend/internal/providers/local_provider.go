package providers

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

type LocalUploadProvider struct {
	basePath string
}

func NewLocalUploadProvider(basePath string) *LocalUploadProvider {
	return &LocalUploadProvider{basePath: basePath}
}

func (p *LocalUploadProvider) UploadFile(file *multipart.FileHeader, path string) (string, error) {
	fullPath := filepath.Join(p.basePath, path)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(src); err != nil {
		return "", err
	}

	return fmt.Sprintf("/uploads/%s", path), nil
}

func (p *LocalUploadProvider) DeleteFile(path string) error {
	fullPath := filepath.Join(p.basePath, filepath.Clean("/"+path))
	if !strings.HasPrefix(fullPath, filepath.Clean(p.basePath)) {
		return fmt.Errorf("invalid path")
	}
	return os.Remove(fullPath)
}
