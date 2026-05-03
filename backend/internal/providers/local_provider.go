package providers

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

type LocalUploadProvider struct {
	basePath string
}

func NewLocalUploadProvider(basePath string) *LocalUploadProvider {
	return &LocalUploadProvider{basePath: basePath}
}

func (p *LocalUploadProvider) UploadFile(_ context.Context, file *multipart.FileHeader, path string) (string, error) {
	fullPath := filepath.Join(p.basePath, path)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer func() {
		if err := src.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close source")
		}
	}()

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := dst.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close destination")
		}
	}()

	if _, err := dst.ReadFrom(src); err != nil {
		return "", err
	}

	return fmt.Sprintf("/uploads/%s", path), nil
}

func (p *LocalUploadProvider) DeleteFile(_ context.Context, path string) error {
	relPath := filepath.Join(".", filepath.Clean("/"+path))
	fullPath := filepath.Join(p.basePath, relPath)
	resolvedPath, err := filepath.EvalSymlinks(fullPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	basePathResolved, err := filepath.EvalSymlinks(p.basePath)
	if err != nil {
		return fmt.Errorf("invalid base path: %w", err)
	}
	if !strings.HasPrefix(resolvedPath, basePathResolved+string(filepath.Separator)) && resolvedPath != basePathResolved {
		return fmt.Errorf("invalid path")
	}
	return os.Remove(resolvedPath)
}
