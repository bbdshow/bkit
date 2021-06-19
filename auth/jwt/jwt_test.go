package jwt

import (
	"testing"
	"time"
)

type Data struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Value    int64
}

func TestJWTToken(t *testing.T) {
	data := Data{
		User:     "test",
		Password: "123456",
	}
	claims := NewCustomClaims(data, time.Second)

	str, err := GenerateJWTToken(claims)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := VerifyJWTToken(str)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("verify not ok")
	}
	ctxData := &Data{}
	if err := GetCustomData(str, ctxData); err != nil {
		t.Fatal(err)
	}
	if ctxData.User != data.User {
		t.Fatal("user not equal")
	}

	if ctxData.Password != data.Password {
		t.Fatal("password not equal")
	}
	time.Sleep(2 * time.Second)

	ok, err = VerifyJWTToken(str)
	if err == nil || ok {
		t.Fatal("should expired")
	}
}
