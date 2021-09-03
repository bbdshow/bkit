package icrypto

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func Test_AESEncryptCBC(t *testing.T) {
	secretKey := "347f36057c5373fab0d69158f345bf8d"
	iv := "7ca646d6dbf731aa0af9e77b"
	origin := "Hello aes"

	sign, err := AESEncryptCBC([]byte(secretKey), []byte(iv), []byte(origin))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(base64.StdEncoding.EncodeToString(sign))
}

func Test_AESDecryptCBC(t *testing.T) {
	//sign, _ := base64.StdEncoding.DecodeString("WRmcoDMyQIJqZUljZI4/lwzSwbgSNGLT/AJVit8awhDB6HsCWMWTnOZDalUkOCHYflBF56fOSBozKkfvlL46IAwWXEf2x9XSRjv7hp3Isxin2CW99D/iezcNxkH6+pBQLPs9bMdX5KutA8tPuE2ioYQ/7Kt2hYPyQOMKJgBYx0hiWZYisUfsIXF1xG0BImbvYpxV1gbllfQYVaLwk6Zt+w==")
	//secretKey, _ := base64.StdEncoding.DecodeString("+F6smyqEdKVVq1AaFl4yRQ==")
	//iv, _ := base64.StdEncoding.DecodeString("c8y6cCgDjenL7IFAeI3BIw==")
	//iv := "/bQg2aJUGkOmBXzBXaFVJA=="
	sign, _ := base64.StdEncoding.DecodeString("ClnJjHTpMasT3FnELakVQQ==")
	secretKey := "347f36057c5373fab0d69158f345bf8d"
	iv := "7ca646d6dbf731aa0af9e77b963f1d41"
	value, err := AESDecryptCBC([]byte(secretKey), []byte(iv), sign)
	fmt.Println(string(value), err)

}
