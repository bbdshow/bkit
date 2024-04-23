package typ

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
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

type StringSplit string

func (ss StringSplit) Unmarshal(sep string) []string {
	strs := make([]string, 0)
	for _, s := range strings.Split(string(ss), sep) {
		s = strings.TrimSpace(s)
		if s != "" {
			strs = append(strs, s)
		}
	}
	return strs
}

func (ss StringSplit) Marshal(val []string, sep string) StringSplit {
	strs := make([]string, 0)
	for _, s := range val {
		s = strings.TrimSpace(s)
		if s != "" {
			strs = append(strs, s)
		}
	}
	return StringSplit(strings.Join(strs, sep))
}

func (ss StringSplit) Has(v string, sep string) bool {
	strs := ss.Unmarshal(sep)
	for _, s := range strs {
		if s == v {
			return true
		}
	}
	return false
}

type IntSplit string

func (is IntSplit) Unmarshal(sep string) ([]int, error) {
	ints := make([]int, 0)
	for _, s := range strings.Split(string(is), sep) {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		ints = append(ints, i)
	}
	return ints, nil
}

func TrimStringToInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}

func TrimStringToFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

// StrRangeN 取字符串 前后部分，总计N个字符
// 这样写主要是处理中文乱码问题
// 注意，返回非严格N字符
func StrRangeN(str string, n int) string {
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

// UniqueSlice 切片去重，是否去掉空字符串，切片顺序不变
func UniqueSlice(s []string, isDropEmpty bool) []string {
	if len(s) == 0 {
		return s
	}
	// 去重
	tmp := make([]string, 0, len(s))
	for _, v := range s {
		if isDropEmpty {
			if v == "" {
				continue
			}
		}
		if !InSlice(v, tmp) {
			tmp = append(tmp, v)
		}
	}
	return tmp

}

func InSlice(str string, s []string) bool {
	isHit := false
	for _, v := range s {
		if str == v {
			isHit = true
			break
		}
	}
	return isHit
}

// IsLetterOrDigit 是否字母或数字
func IsLetterOrDigit(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return true
}

// ContainsChinese 是否存在中文字符
func ContainsChinese(text string) bool {
	re := regexp.MustCompile(`\p{Han}`)
	return re.MatchString(text)
}
