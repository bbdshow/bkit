package str

import (
	"encoding/hex"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// RandNumCode 随机生成数字验证证码, 最大20长度
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
		// 补零
		bit--
		s += "0"
	}
	return s + val
}

// RandAlphaNumString 随机生成字母数字字符串
func RandAlphaNumString(strLen int, lower ...bool) string {
	if strLen <= 0 {
		return ""
	}
	str := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := make([]byte, strLen)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < strLen; i++ {
		result[i] = bytes[r.Intn(len(bytes))]
	}
	// 防止连续调用，使用到同一随机因子，如果并发调用使用到同一随机因子，可能生成出来一样的字符
	time.Sleep(time.Nanosecond)
	if len(lower) > 0 && lower[0] {
		return strings.ToLower(string(result))
	}
	return string(result)
}

func RandHexString(strLen int, upper ...bool) string {
	if strLen <= 0 {
		return ""
	}
	str := ""
	buff := make([]byte, strLen)
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(buff)
	str = hex.EncodeToString(buff)
	if len(upper) > 0 && upper[0] {
		str = strings.ToUpper(str)
	}
	if len(str) > strLen {
		str = str[len(str)-strLen:]
	}
	return str
}
