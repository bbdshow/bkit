package oss

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"testing"
)

const (
	bucket = ""
	domain = "http://qbavo0jyg.bkt.clouddn.com"
	ak     = ""
	sk     = ""
)

func TestQiNiuOSS_Base64Put(t *testing.T) {
	str, err := encode("/Users/hzq/Desktop/test.docx")
	if err != nil {
		t.Fatal(err)
	}
	oss := NewQiNiuOSS(ak, sk, domain, bucket)
	err = oss.Base64Put(context.Background(), "test.docx", str, "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewQiNiuOSS_Put(t *testing.T) {
	buff, err := ioutil.ReadFile("/Users/hzq/Desktop/images/IMG_20170423_085501.jpg")
	if err != nil {
		t.Fatal(err)
	}
	reader := bytes.NewReader(buff)
	fmt.Println(len(buff), reader.Size())
	oss := NewQiNiuOSS(ak, sk, domain, bucket)
	err = oss.Put(context.Background(), "images/i.jpg", reader, reader.Size(), "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestQiNiuOSS_TempToken(t *testing.T) {
	oss := NewQiNiuOSS(ak, sk, domain, bucket)
	t.Log(oss.PutToken(60, "test"))
}

func TestQiNiuOSS_URL(t *testing.T) {
	oss := NewQiNiuOSS(ak, sk, domain, bucket)
	t.Log(oss.URL("test.docx"))
}

func encode(filename string) ([]byte, error) {
	buff, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	ret := make([]byte, base64.StdEncoding.EncodedLen(len(buff)))
	base64.StdEncoding.Encode(ret, buff)
	return ret, nil
}
