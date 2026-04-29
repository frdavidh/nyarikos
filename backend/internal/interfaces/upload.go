package interfaces

import (
	"context"
	"mime/multipart"
)

type UploadProvider interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, path string) (string, error)
	DeleteFile(ctx context.Context, path string) error
}
