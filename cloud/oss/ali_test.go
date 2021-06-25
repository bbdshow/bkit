package oss

import (
	"fmt"
	"testing"
)

var aliOSS *AliOSS

func init() {
	cfg := AliOSSConfig{
		AccessKey:      "",
		SecretKey:      "",
		Domain:         "oss-ap-southeast-1.aliyuncs.com",
		Bucket:         "",
		StsDomain:      "sts.ap-southeast-1.aliyuncs.com",
		StsRoleArn:     "",
		StsSessionName: "",
	}
	oss, err := NewAliOSS(cfg)
	if err != nil {
		panic(err)
	}
	aliOSS = oss
}

func TestAliOSS_TempToken(t *testing.T) {
	fmt.Println(aliOSS.PutToken(3600, ""))
}
