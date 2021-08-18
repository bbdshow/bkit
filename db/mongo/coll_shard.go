package mongo

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// CollShardByTime 集合名按时间
type CollShardByDay struct {
	Prefix string
	// 天范围
	daySpan map[int]int
}

// NewCollShardByTime
// prefix 集合名前缀
// day>=31 | day<=0 则按月分片,  day = x 则每 x 天为一个分片区间
func NewCollShardByDay(prefix string, day int) CollShardByDay {
	coll := CollShardByDay{Prefix: prefix, daySpan: make(map[int]int)}
	coll.daySpan = coll.calcDaySpan(day)
	return coll
}

func (coll CollShardByDay) calcDaySpan(day int) map[int]int {
	if day <= 0 || day > 31 {
		day = 31
	}
	size := 0
	span := 1
	daySpan := make(map[int]int, 0)
	for i := 1; i <= 31; i++ {
		if size >= day {
			size = 0
			span++
		}
		size++
		daySpan[i] = span
	}
	return daySpan
}
func (coll CollShardByDay) collName(bucket string, year, month, span int) string {
	return fmt.Sprintf("%s_%s_%d%02d_%02d", coll.Prefix, bucket, year, month, span)
}

func (coll CollShardByDay) EncodeCollName(bucket string, timestamp int64) string {
	y, m, d := time.Unix(timestamp, 0).Date()
	s := coll.daySpan[d]
	return coll.collName(bucket, y, int(m), s)
}

func (coll CollShardByDay) DecodeCollName(collName string) (prefix string, index, year int, month time.Month, span int, err error) {
	str := strings.Split(collName, "_")
	if len(str) != 4 {
		err = fmt.Errorf("invalid collection name %s", collName)
		return
	}
	prefix = str[0]
	di, _ := strconv.ParseInt(str[1], 10, 64)
	index = int(di)
	y, _ := strconv.ParseInt(str[2][:4], 10, 64)
	year = int(y)
	m, _ := strconv.ParseInt(str[2][4:], 10, 64)
	month = time.Month(m)
	s, _ := strconv.ParseInt(str[3], 10, 64)
	span = int(s)
	return
}

func (coll CollShardByDay) DaySpan() map[int]int {
	v := make(map[int]int)
	for i, span := range coll.daySpan {
		v[i] = span
	}
	return v
}

func (coll CollShardByDay) CollNameDate(collName string) (time.Time, error) {
	_, _, y, m, _, err := coll.DecodeCollName(collName)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(y, m, 0, 0, 0, 0, 0, time.Local), nil
}

// SpanLastTime 当前集合时间分区内，最后一天
func (coll CollShardByDay) SpanLastTime(collName string) (t time.Time, err error) {
	_, _, y1, m1, n1Span, err := coll.DecodeCollName(collName)
	if err != nil {
		return t, err
	}
	minDay := math.MaxInt32
	for d, span := range coll.daySpan {
		if span > n1Span {
			if d < minDay {
				minDay = d
			}
		}
	}
	t = time.Date(y1, m1, minDay, 0, 0, 0, 0, time.Local)
	return t, nil
}

// CollNameByStartEnd 根据开始时间和结束时间，查询出所有生成的 name
func (coll CollShardByDay) CollNameByStartEnd(bucket string, start, end int64) []string {
	startTime := time.Unix(start, 0)
	endTime := time.Unix(end, 0)

	midTime := startTime
	date := []time.Time{startTime}
	for {
		midTime = midTime.AddDate(0, 0, 1)
		if midTime.Before(endTime) {
			date = append(date, midTime)
			continue
		}
		break
	}
	nameMap := make(map[string]struct{})
	names := make([]string, 0, len(date))
	for _, v := range date {
		name := coll.EncodeCollName(bucket, v.Unix())
		_, ok := nameMap[name]
		if !ok {
			nameMap[name] = struct{}{}
			names = append(names, name)
		}
	}
	return names
}
