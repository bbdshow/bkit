package itime

import (
	"math"
	"time"
)

const (
	DateTime = "2006-01-02 15:04:05"
	Date     = "2006-01-02"
)

func DateString(t time.Time, layout ...string) string {
	var f = Date
	if len(layout) > 0 && layout[0] != "" {
		f = layout[0]
	}
	return t.Format(f)
}

func ToMill(t time.Time) int64 {
	return t.UnixNano() / time.Millisecond.Nanoseconds()
}

func YesterdayDate() time.Time {
	y := time.Now().AddDate(0, 0, -1)
	return time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, time.Local)
}

func BeforeDayDate(day int) time.Time {
	y := time.Now().AddDate(0, 0, -int(math.Abs(float64(day))))
	return time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, time.Local)
}

func UnixSecToDate(sec int64) time.Time {
	y, m, d := time.Unix(sec, 0).Date()
	date := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	return date
}
