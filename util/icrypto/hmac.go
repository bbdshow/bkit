package icrypto

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// HmacSha1ToBase64  HMAC-SHA1
func HmacSha1ToBase64(rawStr, key string) string {
	hmacHash := hmac.New(sha1.New, []byte(key))
	hmacHash.Write([]byte(rawStr))
	return base64.StdEncoding.EncodeToString(hmacHash.Sum(nil))
}

// HmacSha256ToBase64  HMAC-SHA256
func HmacSha256ToBase64(rawStr, key string) string {
	hmacHash := hmac.New(sha256.New, []byte(key))
	hmacHash.Write([]byte(rawStr))
	return base64.StdEncoding.EncodeToString(hmacHash.Sum(nil))
}

// HmacSha1ToHex
func HmacSha1ToHex(rawStr, key string) string {
	hmacHash := hmac.New(sha1.New, []byte(key))
	hmacHash.Write([]byte(rawStr))
	return hex.EncodeToString(hmacHash.Sum(nil))
}

// HmacSha256ToHex
func HmacSha256ToHex(rawStr, key string) string {
	hmacHash := hmac.New(sha256.New, []byte(key))
	hmacHash.Write([]byte(rawStr))
	return hex.EncodeToString(hmacHash.Sum(nil))
}
