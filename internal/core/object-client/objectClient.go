package objectclient

import (
	"context"
	"io"
)

// ObjectClient defines interactions with S3 or any object storage.
// It’s abstract so you can replace AWS with MinIO, GCP, etc. easily.
type ObjectClient interface {
	UploadFile(ctx context.Context, bucket, key string, data io.Reader, contentType string) (url string, err error)
	DeleteFile(ctx context.Context, bucket, key string) error
	GetFile(ctx context.Context, bucket, key string) ([]byte, error)

	GetObjectReader(ctx context.Context, bucket, key string) (io.ReadCloser, error)
}
