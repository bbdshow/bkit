package openid

import (
	"fmt"
	"github.com/bbdshow/bkit/gen/invitecode"
	"github.com/bbdshow/bkit/gen/str"
)

type OpenId string

// Init openId 编码
func (id OpenId) Init(uid int64, chanNo string) OpenId {
	u := invitecode.Encode(uint64(uid))
	// 填充字符串放入基数索引位
	fullStr := str.RandAlphaNumString(28)
	uc := u + chanNo
	index := 0
	openId := ""
	for i := 0; i < 40; i++ {
		if i%2 == 0 && index < len(uc) {
			openId += string(uc[index])
			index++
		} else {
			openId += string(fullStr[i-index])
		}
	}
	return OpenId(openId)
}

func (id OpenId) Uid() (int64, error) {
	if len(id) != 40 {
		return 0, fmt.Errorf("invalid openId")
	}
	indexes := []int{0, 2, 4, 6, 8, 10}
	u := ""
	for _, i := range indexes {
		u += string(id[i])
	}
	uid := invitecode.Decode(u)
	return int64(uid), nil
}

func (id OpenId) ChanNo() (string, error) {
	if len(id) != 40 {
		return "", fmt.Errorf("invalid openId")
	}
	indexes := []int{12, 14, 16, 18, 20, 22}
	c := ""
	for _, i := range indexes {
		c += string(id[i])
	}
	return c, nil
}

func (id OpenId) UidAndChanNo() (int64, string, error) {
	uid, err := id.Uid()
	if err != nil {
		return 0, "", err
	}
	chanNo, err := id.ChanNo()
	if err != nil {
		return 0, "", err
	}

	return uid, chanNo, nil
}

func (id OpenId) String() string {
	return string(id)
}
