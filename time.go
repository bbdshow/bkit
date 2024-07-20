package bkit

import (
	"context"
	"fmt"
	"math"
	"time"
)

const (
	DateTime = time.DateTime
	Date     = "2006-01-02"
)

var (
	Time = TimeUtil{}
)

type TimeUtil struct{}

func (TimeUtil) DateString(t time.Time, layout ...string) string {
	var f = Date
	if len(layout) > 0 && layout[0] != "" {
		f = layout[0]
	}
	return t.Format(f)
}

func (TimeUtil) ToMill(t time.Time) int64 {
	return t.UnixNano() / time.Millisecond.Nanoseconds()
}

func (TimeUtil) YesterdayDate() time.Time {
	y := time.Now().AddDate(0, 0, -1)
	return time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, time.Local)
}

func (TimeUtil) BeforeDayDate(day int) time.Time {
	y := time.Now().AddDate(0, 0, -int(math.Abs(float64(day))))
	return time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, time.Local)
}

func (TimeUtil) UnixSecToDate(sec int64) time.Time {
	y, m, d := time.Unix(sec, 0).Date()
	date := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	return date
}

// TimeToDate parse time string to date
func (tu TimeUtil) TimeToDate(dst ...string) time.Time {
	if len(dst) > 0 {
		t, err := time.ParseInLocation(DateTime, dst[0], time.Local)
		if err == nil {
			return tu.UnixSecToDate(t.Unix())
		}
	}
	return tu.UnixSecToDate(time.Now().Unix())
}

// CtxAfterSecDeadline if not deadline, return defSec, if defSec <= 0, return int32 max sec duration
func (TimeUtil) CtxAfterSecDeadline(ctx context.Context, defSec int32) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		if defSec <= 0 {
			defSec = math.MaxInt32
		}
		return time.Duration(defSec) * time.Second
	}
	sec := int32(time.Until(deadline).Seconds())
	if sec <= 0 {
		sec = defSec
	}
	return time.Duration(sec) * time.Second
}

// StringToTime 解析主流的时间字符串 time.Time 类型,精确到秒，如果解析失败返回当前时间
// 支持的格式: time.DateTime , time.RFC3339
func (TimeUtil) StringToTime(v string) time.Time {
	if v == "" {
		return time.Now()
	}
	t, err := time.ParseInLocation(DateTime, v, time.Local)
	if err == nil {
		return t
	} else {
		fmt.Println(err)
	}
	t, err = time.ParseInLocation(time.RFC3339, v, time.Local)
	if err == nil {
		return t
	} else {
		fmt.Println(err)
	}
	return time.Now()
}
