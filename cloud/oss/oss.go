package oss

import (
	"context"
	"io"
)

type Operation interface {
	Put(ctx context.Context, key string, data io.Reader, size int64, mimeType string) error
	Base64Put(ctx context.Context, key string, raw []byte, mimeType string) error
	Delete(key string) error
	Bucket() string
	RegionId() string
	Domain() string
	// gen source URL by key
	URL(key string) string
	// temp Put action Token
	PutToken(expiredSec int, dir string) (token, accessKey, secretKey string, err error)
}
