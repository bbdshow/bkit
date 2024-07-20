package bkit

import (
	"fmt"
	"testing"
)

func TestInviteCode_DumpKey(t *testing.T) {
	ic := NewInviteCode()
	keys := make(map[string]struct{}, 100000000)
	for i := 1; i < 1000000000; i++ {
		code := ic.Encode(uint64(i))
		_, ok := keys[code]
		if ok {
			t.Fatal("dump key")
		}
		keys[code] = struct{}{}
		if ic.Decode(code) != uint64(i) {
			t.Fatal("decode exception")
		}
	}
}

func TestGenerate_OpenID(t *testing.T) {
	type arg struct {
		Uid       int64
		ChannelNo string
	}
	args := []arg{{Uid: 1, ChannelNo: "JHJSFD"}, {Uid: 1231234, ChannelNo: "JHJSFD"}, {Uid: 12323, ChannelNo: "9HJSFD"}, {Uid: 9981, ChannelNo: "SFWES1"}}

	for _, v := range args {
		openId := Generate.GenOpenID(v.Uid, v.ChannelNo)
		fmt.Println(openId)

		uid, err := openId.Uid()
		if err != nil {
			t.Fatal(err)
		}
		if uid != v.Uid {
			t.Fatal("uid not equal")
		}

		channelNo, err := openId.ChannelNo()
		if err != nil {
			t.Fatal(err)
		}
		if channelNo != v.ChannelNo {
			t.Fatal("ChannelNo not equal")
		}
	}
}

func TestGenerate_OrderID(t *testing.T) {
	orderID := Generate.NewOrderID().WithTag("m").WithTag("m")
	fmt.Println(orderID, orderID.Time(), orderID.Tags(), orderID.HasTag("m"))
}
