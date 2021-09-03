package icrypto

import (
	"crypto/md5"
	"fmt"
)

// PasswordSlatMD5 密码MD5加盐
func PasswordSlatMD5(password, slat string) string {
	return Md5String(password, ":", slat)
}

// Md5String 字符串md5
func Md5String(s string, multi ...string) string {
	for _, v := range multi {
		s += v
	}
	val := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", val)
}
