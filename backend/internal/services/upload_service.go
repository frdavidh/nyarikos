package services

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
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

func (s *UploadService) UploadFile(file *multipart.FileHeader, path string) (string, error) {
	return s.uploadProvider.UploadFile(file, path)
}

func (s *UploadService) DeleteFile(path string) error {
	return s.uploadProvider.DeleteFile(path)
}

func (s *UploadService) UploadKostImage(kostID uint, file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isValidImageExt(ext) {
		return "", fmt.Errorf("%w: %s", ErrInvalidFileType, ext)
	}

	path := fmt.Sprintf("kost/%d/%s", kostID, file.Filename)

	return s.uploadProvider.UploadFile(file, path)
}

func isValidImageExt(ext string) bool {
	validExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	return slices.Contains(validExts, ext)
}
