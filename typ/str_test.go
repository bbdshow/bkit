package typ

import (
	"fmt"
	"testing"
)

func TestRandNumCode(t *testing.T) {
	fmt.Println(RandNumCode(6))
}

func TestRandAlphaNumString(t *testing.T) {
	fmt.Println(RandAlphaNumString(16, true))
}

func TestPasswordSlat(t *testing.T) {
	p1 := PasswordSlatMD5("1", "123456")
	p2 := PasswordSlatMD5("112", "3456")
	if p1 == p2 {
		t.Fatal(p1, p2)
	}

	p3 := PasswordSlatMD5("112", "3456")
	p4 := PasswordSlatMD5("112", "3456")
	if p3 != p4 {
		t.Fatal(p3, p4)
	}
}

func TestSubstring(t *testing.T) {
	str := "hello word - 你好世界"
	v1 := Substring(str, 3, 8)
	if v1 != "lo wo" {
		t.Fatal(v1)
	}
	v2 := Substring(str, -1, 2)
	if v2 != "世界" {
		t.Fatal(v2)
	}
	v3 := Substring(str, -1, 17)
	if v3 != "hello word - 你好世界" {
		t.Fatal(v3)
	}
}

func TestSubstr(t *testing.T) {
	str := "hello word - 你好世界"
	v1 := Substr(str, 3, 8)
	if v1 != "lo word " {
		t.Fatal(v1)
	}
	v2 := Substr(str, -1, 2)
	if v2 != "世界" {
		t.Fatal(v2)
	}
	v3 := Substr(str, 16, 17)
	if v3 != "界" {
		t.Fatal(v3)
	}
}
