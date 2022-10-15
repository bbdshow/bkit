package oss

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
)

type QiNiuOSS struct {
	domain    string
	bucket    string
	accessKey string
	secretKey string
}

func NewQiNiuOSS(ak, sk, domain, bucket string) *QiNiuOSS {
	oss := &QiNiuOSS{
		domain:    domain,
		bucket:    bucket,
		accessKey: ak,
		secretKey: sk,
	}
	return oss
}

func (oss *QiNiuOSS) URL(key string) string {
	return storage.MakePublicURL(oss.domain, key)
}
func (oss *QiNiuOSS) Domain() string {
	return oss.domain
}
func (oss *QiNiuOSS) Bucket() string {
	return oss.bucket
}

func (oss *QiNiuOSS) RegionId() string {
	return ""
}

func (oss *QiNiuOSS) Put(ctx context.Context, key string, data io.Reader, size int64, mimeType string) error {
	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", oss.bucket, key),
	}
	upToken := putPolicy.UploadToken(auth.New(oss.accessKey, oss.secretKey))

	cfg := storage.Config{
		ApiHost: storage.DefaultAPIHost,
	}
	up := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	err := up.Put(ctx, &ret, upToken, key, data, size, &storage.PutExtra{
		MimeType: mimeType,
	})
	return err
}

func (oss *QiNiuOSS) Base64Put(ctx context.Context, key string, raw []byte, mimeType string) error {
	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", oss.bucket, key),
	}
	upToken := putPolicy.UploadToken(auth.New(oss.accessKey, oss.secretKey))

	cfg := storage.Config{
		ApiHost: storage.DefaultAPIHost,
	}
	up := storage.NewBase64Uploader(&cfg)
	ret := storage.PutRet{}
	err := up.Put(ctx, &ret, upToken, key, raw, &storage.Base64PutExtra{
		MimeType: mimeType,
	})
	return err
}

func (oss *QiNiuOSS) Delete(key string) error {
	cfg := storage.Config{
		ApiHost: storage.DefaultAPIHost,
	}
	mgr := storage.NewBucketManager(auth.New(oss.accessKey, oss.secretKey), &cfg)
	return mgr.Delete(oss.bucket, key)
}

func (oss *QiNiuOSS) PutToken(expiredSec int, dir string) (token, accessKey, secretKey string, err error) {
	scope := oss.bucket
	if dir != "" {
		scope = fmt.Sprintf("%s:%s", scope, dir)
	}
	putPolicy := storage.PutPolicy{
		Scope:   scope,
		Expires: uint64(expiredSec),
	}
	upToken := putPolicy.UploadToken(auth.New(oss.accessKey, oss.secretKey))
	return upToken, "", "", nil
}

func (oss *QiNiuOSS) GetModifyTime(ctx context.Context, key string) (time.Time, error) {
	cfg := storage.Config{
		ApiHost: storage.DefaultAPIHost,
	}
	mgr := storage.NewBucketManager(auth.New(oss.accessKey, oss.secretKey), &cfg)
	fInfo, err := mgr.Stat(oss.bucket, key)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, fInfo.PutTime*100), nil
}

func (oss *QiNiuOSS) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	_url := oss.URL(key)
	return oss.GetWithURL(ctx, _url)
}

func (oss *QiNiuOSS) GetWithURL(ctx context.Context, url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(b)), nil
}
