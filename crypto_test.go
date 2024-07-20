package bkit

import (
	"fmt"
	"os"
	"testing"
)

func TestAES_DecryptSecretKey(t *testing.T) {
	InitAES([]byte(os.Getenv("CSM_DEFAULT_AES_KEY")))
	v, err := DecryptSecretKey("")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(v)
}
