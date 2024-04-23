package ordernum

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var num int64

type OrderId string

var pid = os.Getpid() % 1000

func NewOrderId() OrderId {
	return NewOrderIdWithTime(time.Now())
}

//gen 24-bit order num
//17-bit mean time precision ms ，3-bit mean process id，last 4-bit mean incr num
func NewOrderIdWithTime(t time.Time) OrderId {
	sec := t.Format("20060102150405")
	mill := t.UnixNano()/1e6 - t.UnixNano()/1e9*1e3
	i := atomic.AddInt64(&num, 1)
	r := i % 10000
	//rs := sup(r, 4)
	id := fmt.Sprintf("%s%03d%03d%04d", sec, mill, pid, r)
	return OrderId(id)
}

func (id OrderId) String() string {
	return string(id)
}

func (id OrderId) Len() int {
	return len(id)
}

func (id OrderId) Time() time.Time {
	idStr := id.String()
	i := strings.LastIndex(idStr, ":")
	if id.Len() > i {
		s := idStr[i+1:]
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

func (id OrderId) Tags() []string {
	idStr := id.String()
	i := strings.LastIndex(idStr, ":")
	if id.Len() > i {
		str := idStr[:i]
		return strings.Split(str, ":")
	}
	return []string{}
}

func (id OrderId) ExistsTag(tag string) bool {
	tags := id.Tags()
	for _, v := range tags {
		if v == tag {
			return true
		}
	}
	return false
}

func (id OrderId) WithTag(tag string) OrderId {
	return OrderId(fmt.Sprintf("%s:%s", tag, id))
}
