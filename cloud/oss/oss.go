package oss

import (
	"context"
	"io"
)

type Operation interface {
	Put(ctx context.Context, key string, data io.Reader, size int64, mimeType string) error
	Base64Put(ctx context.Context, key string, raw []byte, mimeType string) error
	Delete(key string) error
	// 仓库
	Bucket() string
	RegionId() string
	// 域名
	Domain() string
	// 通过Key 生成访问URL
	URL(key string) string
	// 临时put Token
	PutToken(expiredSec int, dir string) (token, accessKey, secretKey string, err error)
}
