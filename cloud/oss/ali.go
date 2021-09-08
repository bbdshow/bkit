package oss

import (
	"context"
	"fmt"
	"io"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	aliOss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type AliOSSConfig struct {
	AccessKey string
	SecretKey string
	Domain    string
	RegionId  string
	Bucket    string

	StsDomain      string
	StsRoleArn     string
	StsSessionName string
}

type AliOSS struct {
	domain     string
	regionId   string
	bucket     string
	prefixPath string
	accessKey  string
	secretKey  string

	cliBucket   *aliOss.Bucket
	sts         *sts.Client
	stsDomain   string
	roleArn     string
	sessionName string
}

// roleArn, sessionName  aliCloud sts service
func NewAliOSS(cfg AliOSSConfig) (*AliOSS, error) {
	oss := &AliOSS{
		domain:    cfg.Domain,
		bucket:    cfg.Bucket,
		accessKey: cfg.AccessKey,
		secretKey: cfg.SecretKey,
		regionId:  cfg.RegionId,

		stsDomain:   cfg.StsDomain,
		roleArn:     cfg.StsRoleArn,
		sessionName: cfg.StsSessionName,
	}
	cli, err := aliOss.New(oss.domain, oss.accessKey, oss.secretKey)
	if err != nil {
		return nil, err
	}
	oss.cliBucket, err = cli.Bucket(oss.bucket)
	if err != nil {
		return nil, err
	}

	oss.sts, err = sts.NewClientWithAccessKey("", oss.accessKey, oss.secretKey)
	if err != nil {
		return nil, err
	}
	oss.sts.Domain = oss.stsDomain

	return oss, nil
}

func (oss *AliOSS) URL(key string) string {
	return ""
}
func (oss *AliOSS) Domain() string {
	return oss.domain
}
func (oss *AliOSS) Bucket() string {
	return oss.bucket
}

func (oss *AliOSS) RegionId() string {
	return oss.regionId
}

func (oss *AliOSS) Put(ctx context.Context, key string, data io.Reader, size int64, mimeType string) error {
	return oss.cliBucket.PutObject(key, data)
}

func (oss *AliOSS) Base64Put(ctx context.Context, key string, raw []byte, mimeType string) error {
	return nil
}

func (oss *AliOSS) Delete(key string) error {
	return nil
}

func (oss *AliOSS) PutToken(expiredSec int, dir string) (token, accessKey, secretKey string, err error) {
	req := sts.CreateAssumeRoleRequest()
	req.SetScheme("https")
	req.Domain = oss.stsDomain
	req.RoleArn = oss.roleArn
	req.RoleSessionName = oss.sessionName
	req.DurationSeconds = requests.NewInteger(expiredSec)
	req.Policy = fmt.Sprintf(`{"Version": "1","Statement": [{"Effect": "Allow","Action": ["oss:PutObject","oss:GetObject"],"Resource": ["acs:oss:*:*:%s/%s","acs:oss:*:*:%s/%s*"]}]}`,
		oss.bucket, dir, oss.bucket, dir)
	resp, err := oss.sts.AssumeRole(req)
	if err != nil {
		return "", "", "", err
	}
	if resp != nil && resp.GetHttpStatus() == 200 {
		return resp.Credentials.SecurityToken, resp.Credentials.AccessKeyId, resp.Credentials.AccessKeySecret, nil
	}
	return "", ":", "", fmt.Errorf("temp token invalid")
}
