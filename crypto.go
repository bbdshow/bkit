package bkit

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

var (
	AES *AESUtil
)

// InitAES  初始化默认AES
func InitAES(key []byte) {
	AES = NewAESUtil(key)
}

func EncryptSecretKey(secretKey string) (string, error) {
	if AES == nil {
		return "", fmt.Errorf("default AES not init")
	}
	return AES.EncryptSecretKey(secretKey)
}
func DecryptSecretKey(encryptSecretKey string) (string, error) {
	if AES == nil {
		return "", fmt.Errorf("default AES not init")
	}
	return AES.DecryptSecretKey(encryptSecretKey)
}

// AESUtil - 与统一调度平台实现一致，可能涉及到密钥授权
type AESUtil struct {
	key []byte
}

func NewAESUtil(key []byte) *AESUtil {
	return &AESUtil{
		key: key,
	}
}

// EncryptSecretKey -
func (a *AESUtil) EncryptSecretKey(secretKey string) (string, error) {
	plaintext := []byte(secretKey)

	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", err
	}
	cipherText := make([]byte, aes.BlockSize+len(plaintext))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// DecryptSecretKey -
func (a *AESUtil) DecryptSecretKey(encryptSecretKey string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(encryptSecretKey)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", fmt.Errorf("cipher text too short")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}
