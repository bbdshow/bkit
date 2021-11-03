package oss

import (
	"context"
	"fmt"
	"testing"
)

var aliOSS *AliOSS

func init() {
	// oss-ap-northeast-1.aliyuncs.com
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

func TestAliOSS_GetModifyTime(t *testing.T) {
	mt, err := aliOSS.GetModifyTime(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(mt)
}
