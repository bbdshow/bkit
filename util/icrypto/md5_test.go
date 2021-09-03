package icrypto

import "testing"

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
