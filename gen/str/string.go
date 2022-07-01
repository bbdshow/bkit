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

// RandAlphaNumString rand x len string
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

// PasswordSlatMD5 md5 pwd & slat
func PasswordSlatMD5(password, slat string) string {
	return Md5String(password, ":", slat)
}

// Md5String md5 strings
func Md5String(s string, multi ...string) string {
	for _, v := range multi {
		s += v
	}
	val := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", val)
}

// Substring if start < 0, desc, if arg invalid, return str   [start,end)
func Substring(str string, start, end int) string {
	s := []rune(str)
	if start > len(s) || end > len(s) || end < start {
		return str
	}
	if start < 0 {
		s = s[len(s)-end:]
		return string(s)
	}
	return string(s[start:end])
}

// Substr if start < 0, desc, if arg invalid, return str  [start,start+length)
func Substr(str string, start, length int) string {
	s := []rune(str)
	if start > len(s) || length > len(s) || length <= 0 {
		return str
	}
	if start < 0 {
		s = s[len(s)-length:]
		return string(s)
	}
	maxLength := len(s) - start
	if length > maxLength {
		length = maxLength
	}
	return string(s[start : start+length])
}
