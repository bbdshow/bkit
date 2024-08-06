package mgo

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ShardCollection 分片集合名称
type ShardName struct {
	// 前缀
	prefix string
	// 分隔符
	sep string
	// 天数跨度最大31
	daySpan map[int]int
}

// NewShardCollection 创建分片集合 如果 day >= 31 | day <= 0 则=31,  day = x, span = x
func NewShardName(prefix string, day int) ShardName {
	sn := ShardName{prefix: prefix, daySpan: make(map[int]int), sep: "_"}
	sn.daySpan = sn.calcDaySpan(day)
	return sn
}

func (sn ShardName) calcDaySpan(day int) map[int]int {
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
func (sn ShardName) name(bucket string, year, month, span int) string {
	return fmt.Sprintf("%s%s%s%s%d%02d%s%02d", sn.prefix, sn.sep, bucket, sn.sep, year, month, sn.sep, span)
}

func (sn ShardName) EncodeName(bucket string, timestamp int64) string {
	y, m, d := time.Unix(timestamp, 0).Date()
	span := sn.daySpan[d]
	return sn.name(bucket, y, int(m), span)
}

func (sn ShardName) DecodeName(name string) (prefix, bucket string, year int, month time.Month, span int, err error) {
	str := strings.Split(name, sn.sep)
	if len(str) != 4 {
		err = fmt.Errorf("invalid shard name: %s", name)
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

func (sn ShardName) DaySpan() map[int]int {
	v := make(map[int]int)
	for i, span := range sn.daySpan {
		v[i] = span
	}
	return v
}

func (sn ShardName) NameToDate(name string) (time.Time, error) {
	_, _, y, m, _, err := sn.DecodeName(name)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(y, m, 0, 0, 0, 0, 0, time.Local), nil
}

// CalcSpanDate 计算分隔日期，超过这个日期，则代表到下一个时间段
func (sn ShardName) CalcSpanDate(name string) (t time.Time, err error) {
	_, _, y, m, n, err := sn.DecodeName(name)
	if err != nil {
		return t, err
	}
	spanM := map[int]int{}
	for d, span := range sn.daySpan {
		v, ok := spanM[span]
		if ok {
			// 如果存在则取最大的天数
			if d > v {
				spanM[span] = d
			}
		} else {
			spanM[span] = d
		}
	}
	// 取出下一个时间段的天数
	day, ok := spanM[n]
	if !ok {
		return t, fmt.Errorf("calc span Day: %d not exists", n)
	}
	t = time.Date(y, m, day, 0, 0, 0, 0, time.Local)
	return t, nil
}

// EncodeNameWithRangeTime 根据时间范围生成分片名称, 解决时间，跨片问题
func (sn ShardName) EncodeNameWithRangeTime(bucket string, start, end int64) []string {
	startTime := time.Unix(start, 0)
	endTime := time.Unix(end, 0)

	midTime := startTime
	date := []time.Time{startTime}
	// 计算中间时间
	for {
		midTime = midTime.AddDate(0, 0, 1)
		if midTime.Before(endTime) {
			date = append(date, midTime)
			continue
		}
		break
	}

	names := make([]string, 0, len(date))
	for _, v := range date {
		name := sn.EncodeName(bucket, v.Unix())
		// 名称可能重复
		hit := false
		for _, n := range names {
			if n == name {
				hit = true
				break
			}
		}
		if !hit {
			names = append(names, name)
		}
	}
	return names
}
