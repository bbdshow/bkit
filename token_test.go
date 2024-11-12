package bkit

import (
	"fmt"
	"testing"
)

func TestToken_GenRandomToken(t *testing.T) {
	token := Token.GenRandomToken()
	fmt.Println(token)
	fmt.Println(Token.VerifyRandomToken(token))

	fmt.Println("pre token", Token.VerifyRandomToken("f5e304d401af3ce60f1aafd179421fd9"))
}
