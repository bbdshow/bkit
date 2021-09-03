package str

import (
	"fmt"
	"testing"
)

func TestRandNumCode(t *testing.T) {
	fmt.Println(RandNumCode(6))
}

func TestRandAlphaNumString(t *testing.T) {
	fmt.Println(RandAlphaNumString(32, true))
}

func TestRandHexString(t *testing.T) {
	str := RandHexString(5)
	fmt.Println(str, len(str))
}
