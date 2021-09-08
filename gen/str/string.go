package str

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// RandNumCode rand number verify code, max len 20
func RandNumCode(strLen int) string {
	if strLen > 20 {
		strLen = 20
	}
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	max := int64(1)
	for i := 0; i < strLen; i++ {
		max *= 10
	}
	val := strconv.FormatInt(rnd.Int63n(max), 10)
	bit := strLen - len(val)
	s := ""
	for bit > 0 {
		// zero fill
		bit--
		s += "0"
	}
	return s + val
}

// RandAlphaNumString
func RandAlphaNumString(strLen int, lower ...bool) string {
	str := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := make([]byte, strLen)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < strLen; i++ {
		result[i] = bytes[r.Intn(len(bytes))]
	}
	if len(lower) > 0 && lower[0] {
		return strings.ToLower(string(result))
	}
	return string(result)
}

// PasswordSlatMD5
func PasswordSlatMD5(password, slat string) string {
	return Md5String(password, ":", slat)
}

// Md5String
func Md5String(s string, multi ...string) string {
	for _, v := range multi {
		s += v
	}
	val := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", val)
}
