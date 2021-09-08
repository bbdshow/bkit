package icrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
)

func AESEncryptCBC(key, iv []byte, origData []byte) (encrypted []byte, err error) {
	if len(iv) < 16 {
		return nil, fmt.Errorf("iv length must lte 16")
	}
	defer func() {
		e := recover()
		if e != nil {
			err = fmt.Errorf("EncryptCBC %v", e)
		}
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	iv = iv[:blockSize]
	origData = PKCS7Padding(origData, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, iv)
	encrypted = make([]byte, len(origData))

	blockMode.CryptBlocks(encrypted, origData)

	return encrypted, nil

}

func AESDecryptCBC(key, iv, encrypted []byte) (origData []byte, err error) {
	if len(iv) < 16 {
		return nil, fmt.Errorf("iv length must lte 16")
	}

	defer func() {
		e := recover()
		if e != nil {
			err = fmt.Errorf("DecryptCBC %v", e)
		}
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	iv = iv[:blockSize]

	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData = make([]byte, len(encrypted))

	blockMode.CryptBlocks(origData, encrypted)
	origData, err = UnPadding(origData)

	return origData, err
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

func UnPadding(origData []byte) ([]byte, error) {
	if len(origData) == 0 {
		return origData, nil
	}
	length := len(origData)
	// reduce last byte unPadding
	unPadding := int(origData[length-1])
	if len(origData) >= (length-unPadding) && (length-unPadding) >= 0 {
		return origData[:(length - unPadding)], nil
	}
	return origData, errors.New("UnPadding error, please check key")
}
