package jwt

import (
	"encoding/json"
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
	byt, _ := json.Marshal(data)
	claims := NewCustomClaims(string(byt), time.Second)

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

	ctxData, err := GetCustomData(str)
	if err != nil {
		t.Fatal(err)
	}
	ctxD := &Data{}
	if err := json.Unmarshal([]byte(ctxData), ctxD); err != nil {
		t.Fatal(err)
	}

	if ctxD.User != data.User {
		t.Fatal("user not equal")
	}

	if ctxD.Password != data.Password {
		t.Fatal("password not equal")
	}
	time.Sleep(2 * time.Second)

	ok, err = VerifyJWTToken(str)
	if err == nil || ok {
		t.Fatal("should expired")
	}
}
