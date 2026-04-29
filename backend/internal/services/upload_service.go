package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/frdavidh/nyarikos/internal/interfaces"
)

type UploadService struct {
	uploadProvider interfaces.UploadProvider
}

func NewUploadService(uploadProvider interfaces.UploadProvider) *UploadService {
	return &UploadService{uploadProvider: uploadProvider}
}

func (s *UploadService) UploadFile(ctx context.Context, file *multipart.FileHeader, path string) (string, error) {
	return s.uploadProvider.UploadFile(ctx, file, path)
}

func (s *UploadService) DeleteFile(ctx context.Context, path string) error {
	return s.uploadProvider.DeleteFile(ctx, path)
}

var unsafeFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = unsafeFilenameChars.ReplaceAllString(name, "_")
	return name
}

func (s *UploadService) UploadKostImage(ctx context.Context, kostID uint, file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isValidImageExt(ext) {
		return "", fmt.Errorf("%w: %s", ErrInvalidFileType, ext)
	}

	safeName := sanitizeFilename(file.Filename)
	path := fmt.Sprintf("kost/%d/%s", kostID, safeName)

	return s.uploadProvider.UploadFile(ctx, file, path)
}

func isValidImageExt(ext string) bool {
	validExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	return slices.Contains(validExts, ext)
}
