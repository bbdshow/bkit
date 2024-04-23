package openid

import (
	"fmt"
	"testing"
)

func TestOpenId(t *testing.T) {
	type arg struct {
		Uid    int64
		ChanNo string
	}
	args := []arg{{Uid: 1, ChanNo: "JHJSFD"}, {Uid: 1231234, ChanNo: "JHJSFD"}, {Uid: 12323, ChanNo: "9HJSFD"}, {Uid: 9981, ChanNo: "SFWES1"}}

	for _, v := range args {
		openId := new(OpenId).Init(v.Uid, v.ChanNo)
		fmt.Println(openId)

		uid, err := openId.Uid()
		if err != nil {
			t.Fatal(err)
		}
		if uid != v.Uid {
			t.Fatal("uid not equal")
		}

		channNo, err := openId.ChanNo()
		if err != nil {
			t.Fatal(err)
		}
		if channNo != v.ChanNo {
			t.Fatal("channNo not equal")
		}
	}
}
