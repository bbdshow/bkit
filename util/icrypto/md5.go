package icrypto

import (
	"crypto/md5"
	"fmt"
)

// PasswordSlatMD5 md5(password+slat)
func PasswordSlatMD5(password, slat string) string {
	return Md5String(password, slat)
}

// Md5String md5(str...)
func Md5String(s string, multi ...string) string {
	for _, v := range multi {
		s += v
	}
	val := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", val)
}
