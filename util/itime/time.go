package itime

import "time"

func DateString(t time.Time, layout ...string) string {
	var f = "2006-01-02"
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
