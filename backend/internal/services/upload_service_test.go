package services

import (
	"fmt"
	"mime/multipart"
	"testing"

	"github.com/frdavidh/nyarikos/internal/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUploadProvider struct {
	mock.Mock
}

func (m *mockUploadProvider) UploadFile(file *multipart.FileHeader, path string) (string, error) {
	args := m.Called(file, path)
	return args.String(0), args.Error(1)
}

func (m *mockUploadProvider) DeleteFile(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func TestUploadService_UploadFile(t *testing.T) {
	mockProvider := new(mockUploadProvider)
	service := NewUploadService(mockProvider)

	file := &multipart.FileHeader{Filename: "test.jpg"}
	mockProvider.On("UploadFile", file, "uploads/test.jpg").Return("https://cdn.example.com/test.jpg", nil)

	url, err := service.UploadFile(file, "uploads/test.jpg")

	assert.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/test.jpg", url)
	mockProvider.AssertExpectations(t)
}

func TestUploadService_DeleteFile(t *testing.T) {
	mockProvider := new(mockUploadProvider)
	service := NewUploadService(mockProvider)

	mockProvider.On("DeleteFile", "uploads/test.jpg").Return(nil)

	err := service.DeleteFile("uploads/test.jpg")

	assert.NoError(t, err)
	mockProvider.AssertExpectations(t)
}

func TestUploadService_UploadKostImage_Success(t *testing.T) {
	mockProvider := new(mockUploadProvider)
	service := NewUploadService(mockProvider)

	file := &multipart.FileHeader{Filename: "kost.jpg"}
	mockProvider.On("UploadFile", file, "kost/1/kost.jpg").Return("https://cdn.example.com/kost/1/kost.jpg", nil)

	url, err := service.UploadKostImage(1, file)

	assert.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/kost/1/kost.jpg", url)
	mockProvider.AssertExpectations(t)
}

func TestUploadService_UploadKostImage_InvalidExtension(t *testing.T) {
	mockProvider := new(mockUploadProvider)
	service := NewUploadService(mockProvider)

	file := &multipart.FileHeader{Filename: "kost.pdf"}

	url, err := service.UploadKostImage(1, file)

	assert.Error(t, err)
	assert.Empty(t, url)
	assert.ErrorIs(t, err, ErrInvalidFileType)
}

func TestUploadService_UploadKostImage_InvalidExtension_BMP(t *testing.T) {
	mockProvider := new(mockUploadProvider)
	service := NewUploadService(mockProvider)

	file := &multipart.FileHeader{Filename: "kost.bmp"}

	url, err := service.UploadKostImage(1, file)

	assert.Error(t, err)
	assert.Empty(t, url)
}

func TestIsValidImageExt(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{".jpg", true},
		{".jpeg", true},
		{".png", true},
		{".gif", true},
		{".webp", true},
		{".pdf", false},
		{".bmp", false},
		{".txt", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("ext=%s", tt.ext), func(t *testing.T) {
			result := isValidImageExt(tt.ext)
			assert.Equal(t, tt.expected, result)
		})
	}
}

var _ interfaces.UploadProvider = (*mockUploadProvider)(nil)
