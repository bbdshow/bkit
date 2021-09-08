package typ

import (
	"bytes"
	"regexp"
	"strings"
)

// IsPhone
func IsPhone(phone string) bool {
	reg := regexp.MustCompile(`^\d{1,4}-\d{6,11}$`)
	flag := reg.MatchString(phone)
	return flag
}

// IsChinesePhone
func IsChinesePhone(phone string) bool {
	if IsPhone(phone) {
		return strings.HasPrefix(phone, "86-") || !strings.Contains(phone, "-")
	} else {
		return false
	}
}

// HidePhoneMiddle eg：86-187****0987
func HidePhoneMiddle(phone string) string {
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
