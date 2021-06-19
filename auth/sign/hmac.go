package sign

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

// HmacSha1ToBase64  HMAC-SHA1 签名并BASE64 编码
func HmacSha1ToBase64(rawStr, key string) string {
	hmacHash := hmac.New(sha1.New, []byte(key))
	hmacHash.Write([]byte(rawStr))
	return base64.StdEncoding.EncodeToString(hmacHash.Sum(nil))
}
