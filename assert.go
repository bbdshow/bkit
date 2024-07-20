package bkit

import (
	"regexp"
	"strings"
)

// 断言工具
var Assert = AssertUtil{}

type AssertUtil struct{}

// IsPhone
func (AssertUtil) IsPhone(phone string) bool {
	reg := regexp.MustCompile(`^\d{1,4}-\d{6,11}$`)
	flag := reg.MatchString(phone)
	return flag
}

// IsChinesePhone
func (au AssertUtil) IsChinesePhone(phone string) bool {
	if au.IsPhone(phone) {
		return strings.HasPrefix(phone, "86-") || !strings.Contains(phone, "-")
	} else {
		return false
	}
}
