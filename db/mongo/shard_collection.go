package mongo

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ShardCollection 分片集合名
type ShardCollection struct {
	Prefix string
	Sep    string // 分隔符
	// 天范围
	daySpan map[int]int
}

// prefix 集合名前缀
// day>=31 | day<=0 则按月分片,  day = x 则每 x 天为一个分片区间
func NewShardCollection(prefix string, day int) ShardCollection {
	sc := ShardCollection{Prefix: prefix, daySpan: make(map[int]int), Sep: "_"}
	sc.daySpan = sc.calcDaySpan(day)
	return sc
}

func (sc ShardCollection) calcDaySpan(day int) map[int]int {
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
func (sc ShardCollection) collName(bucket string, year, month, span int) string {
	return fmt.Sprintf("%s%s%s%s%d%02d%s%02d", sc.Prefix, sc.Sep, bucket, sc.Sep, year, month, sc.Sep, span)
}

func (sc ShardCollection) EncodeCollName(bucket string, timestamp int64) string {
	y, m, d := time.Unix(timestamp, 0).Date()
	s := sc.daySpan[d]
	return sc.collName(bucket, y, int(m), s)
}

func (sc ShardCollection) DecodeCollName(collName string) (prefix, bucket string, year int, month time.Month, span int, err error) {
	str := strings.Split(collName, sc.Sep)
	if len(str) != 4 {
		err = fmt.Errorf("invalid collection name %s", collName)
		return
	}
	prefix = str[0]
	bucket = str[1]
	y, _ := strconv.ParseInt(str[2][:4], 10, 64)
	year = int(y)
	m, _ := strconv.ParseInt(str[2][4:], 10, 64)
	month = time.Month(m)
	s, _ := strconv.ParseInt(str[3], 10, 64)
	span = int(s)
	return
}

func (sc ShardCollection) DaySpan() map[int]int {
	v := make(map[int]int)
	for i, span := range sc.daySpan {
		v[i] = span
	}
	return v
}

func (sc ShardCollection) CollNameDate(collName string) (time.Time, error) {
	_, _, y, m, _, err := sc.DecodeCollName(collName)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(y, m, 0, 0, 0, 0, 0, time.Local), nil
}

// SepTime 当前集合时间分区内下一个分隔时间
func (sc ShardCollection) SepTime(collName string) (t time.Time, err error) {
	_, _, y, m, n, err := sc.DecodeCollName(collName)
	if err != nil {
		return t, err
	}
	seps := map[int]int{}
	for d, span := range sc.daySpan {
		v, ok := seps[span]
		if ok {
			if d > v {
				seps[span] = d
			}
		} else {
			seps[span] = d
		}
	}
	day, ok := seps[n]
	if !ok {
		return t, fmt.Errorf("calc sep %d not exists", n)
	}
	t = time.Date(y, m, day, 0, 0, 0, 0, time.Local)
	return t, nil
}

// CollNameByStartEnd 根据开始时间和结束时间，查询出所有生成的 name
func (sc ShardCollection) CollNameByStartEnd(bucket string, start, end int64) []string {
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
		name := sc.EncodeCollName(bucket, v.Unix())
		_, ok := nameMap[name]
		if !ok {
			nameMap[name] = struct{}{}
			names = append(names, name)
		}
	}
	return names
}
