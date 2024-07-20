package bkit

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var Str = StrUtil{}

// StrUtil 字符串工具函数
type StrUtil struct{}

// RandNumCode rand number verify code, max len 20
func (StrUtil) RandNumCode(strLen int) string {
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
func (StrUtil) RandAlphaNumString(strLen int, lower ...bool) string {
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

// MD5
func (StrUtil) MD5(s string, multi ...string) string {
	for _, v := range multi {
		s += v
	}
	val := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", val)
}

// Substring if start < 0, desc, if arg invalid, return str   [start,end)
func (StrUtil) Substring(str string, start, end int) string {
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

// SubstringLength if start < 0, desc, if arg invalid, return str  [start,start+length)
func (StrUtil) SubstringLength(str string, start, length int) string {
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

// StrRangeN 取字符串 前后部分，总计N个字符
// 这样写主要是处理中文乱码问题
// 注意，返回非严格N字符
func (StrUtil) StrRangeN(str string, n int) string {
	// 获取字符串的字符长度
	charLen := utf8.RuneCountInString(str)

	if n <= 0 || charLen <= n {
		return str
	}
	half := n / 2
	beginByteLen := 0
	for i := 0; i < half; {
		_, size := utf8.DecodeRuneInString(str[beginByteLen:])
		beginByteLen += size
		i++
	}
	begin := str[:beginByteLen]

	endIndex := charLen - half
	endByteLen := endIndex
	for i := endIndex; i < charLen; {
		if endByteLen >= len(str) {
			break
		}
		_, size := utf8.DecodeRuneInString(str[endByteLen:])
		endByteLen += size
		i++
	}
	end := str[endByteLen:]
	return begin + "..." + end
}

// IsLetterOrDigit 是否字母或数字
func (StrUtil) IsLetterOrDigit(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return true
}

// ContainsChinese 是否存在中文字符
func (StrUtil) ContainsChinese(text string) bool {
	re := regexp.MustCompile(`\p{Han}`)
	return re.MatchString(text)
}

func (StrUtil) ReverseString(s string) string {
	runes := []rune(s)
	n := len(runes)
	for i := 0; i < n/2; i++ {
		runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
	}
	return string(runes)
}

// RandNumPadding 随机数填充, 生成指定长度的随机数字符串,且首位不为0
func (StrUtil) RandNumPadding(length int) string {
	first := 9
	if length > 18 {
		length = 18
		first = 8
	}
	if length <= 0 {
		return ""
	}
	if length == 1 {
		return strconv.Itoa(1 + rand.Intn(first))
	}
	n := rand.Int63n(int64(math.Pow10(length - 1)))
	return fmt.Sprintf("%d%0*d", 1+rand.Intn(first), length-1, n)
}

// HidePhoneMiddle eg：86-187****0987
func (StrUtil) HidePhoneMiddle(phone string) string {
	hidden := "****"
	if phone == "" {
		return hidden
	}
	arr := strings.Split(phone, "-")
	if len(arr) != 2 {
		return hidden
	}
	number := arr[1]
	if len(number) < 7 {
		return hidden
	}
	prefix := number[0:3]
	suffix := number[len(number)-4:]

	var buf bytes.Buffer
	buf.WriteString(arr[0])
	buf.WriteString("-")
	buf.WriteString(prefix)
	buf.WriteString(hidden)
	buf.WriteString(suffix)

	return buf.String()
}
