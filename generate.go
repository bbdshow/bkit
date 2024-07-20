package bkit

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var Generate = NewGenerateUtil()

var _inviteCode = NewInviteCode()

type GenerateUtil struct {
	InviteCode *InviteCode
}

func NewGenerateUtil() *GenerateUtil {
	return &GenerateUtil{
		InviteCode: NewInviteCode(),
	}
}

type InviteCode struct {
	base    string
	decimal int
	pad     string
	length  int
}

func NewInviteCode() *InviteCode {
	return &InviteCode{
		base:    "E8S2DZX9WYLTN6BGQF7P5IK3MJUAR4HV",
		decimal: 32,
		pad:     "C",
		length:  6,
	}
}

func (ic *InviteCode) SetBase(b string) {
	b = strings.ToUpper(strings.TrimSpace(b))
	if len(b) <= 0 {
		return
	}
	ic.base = b
	ic.decimal = len(ic.base)
}

func (ic *InviteCode) SetPad(p string) error {
	p = strings.ToUpper(p)
	if strings.Contains(ic.base, p) {
		return errors.New("pad should not exists in base")
	}
	ic.pad = p
	return nil
}

func (ic *InviteCode) SetLength(n int) {
	ic.length = n
}

func (ic *InviteCode) Encode(uid uint64) string {
	id := uid
	mod := uint64(0)
	res := ""
	for id != 0 {
		mod = id % uint64(ic.decimal)
		id = id / uint64(ic.decimal)
		res += string(ic.base[mod])
	}
	resLen := len(res)
	if resLen < ic.length {
		res += ic.pad
		for i := 0; i < ic.length-resLen-1; i++ {
			res += string(ic.base[(int(uid)+i)%ic.decimal])
		}
	}
	return res
}

func (ic *InviteCode) Decode(code string) uint64 {
	res := uint64(0)
	lenCode := len(code)
	baseArr := []byte(ic.base)    // string decimal to byte array
	baseRev := make(map[byte]int) // decimal data key to map
	for k, v := range baseArr {
		baseRev[v] = k
	}
	// find cover char addr
	isPad := strings.Index(code, ic.pad)
	if isPad != -1 {
		lenCode = isPad
	}
	r := 0
	for i := 0; i < lenCode; i++ {
		// if cover char , continue
		if string(code[i]) == ic.pad {
			continue
		}
		index, ok := baseRev[code[i]]
		if !ok {
			return 0
		}
		b := uint64(1)
		for j := 0; j < r; j++ {
			b *= uint64(ic.decimal)
		}
		res += uint64(index) * b
		r++
	}
	return res
}

// OpenID - 编码规则 长度40位
type OpenID string

// Uid - 获取用户ID
func (id OpenID) Uid() (int64, error) {
	if len(id) != 40 {
		return 0, errors.New("invalid openID")
	}
	indexes := []int{0, 2, 4, 6, 8, 10}
	u := ""
	for _, i := range indexes {
		u += string(id[i])
	}
	uid := _inviteCode.Decode(u)
	return int64(uid), nil
}

// ChannelNo - 获取渠道号
func (id OpenID) ChannelNo() (string, error) {
	if len(id) != 40 {
		return "", fmt.Errorf("invalid openID")
	}
	indexes := []int{12, 14, 16, 18, 20, 22}
	c := ""
	for _, i := range indexes {
		c += string(id[i])
	}
	return c, nil
}

func (id OpenID) String() string {
	return string(id)
}

// GenOpenID 通过用户ID和渠道号生成OpenID
func (gu *GenerateUtil) GenOpenID(uid int64, channelNo string) OpenID {
	u := _inviteCode.Encode(uint64(uid))
	// 填充字符串放入基数索引位
	fullStr := Str.RandAlphaNumString(28)
	uc := u + channelNo
	index := 0
	openID := ""
	for i := 0; i < 40; i++ {
		if i%2 == 0 && index < len(uc) {
			openID += string(uc[index])
			index++
		} else {
			openID += string(fullStr[i-index])
		}
	}
	return OpenID(openID)
}

var num int64

type OrderID string

var pid = os.Getpid() % 1000

func (gu *GenerateUtil) NewOrderID() OrderID {
	return NewOrderIdWithTime(time.Now())
}

// gen 24-bit order num
// 17-bit mean time precision ms ，3-bit mean process id，last 4-bit mean incr num
func NewOrderIdWithTime(t time.Time) OrderID {
	sec := t.Format("20060102150405")
	mill := t.UnixNano()/1e6 - t.UnixNano()/1e9*1e3
	i := atomic.AddInt64(&num, 1)
	r := i % 10000
	//rs := sup(r, 4)
	id := fmt.Sprintf("%s%03d%03d%04d", sec, mill, pid, r)
	return OrderID(id)
}

func (id OrderID) Time() time.Time {
	str := string(id)
	i := strings.LastIndex(str, ":")
	if len(str) > i {
		s := str[i+1:]
		if len(s) == 24 {
			sec := s[:14]
			mill := s[14:17]
			t, err := time.Parse("20060102150405", sec)
			if err == nil {
				var m time.Duration
				if millInt, err := strconv.ParseInt(mill, 10, 64); err == nil {
					m = time.Duration(millInt) * time.Millisecond
				}
				return t.Add(m)
			}
		}
	}
	return time.Time{}
}

func (id OrderID) Tags() []string {
	str := string(id)
	i := strings.LastIndex(str, ":")
	if len(str) > i {
		s := str[:i]
		return strings.Split(s, ":")
	}
	return []string{}
}

func (id OrderID) HasTag(tag string) bool {
	tags := id.Tags()
	for _, v := range tags {
		if v == tag {
			return true
		}
	}
	return false
}

func (id OrderID) WithTag(tag string) OrderID {
	return OrderID(fmt.Sprintf("%s:%s", tag, id))
}
