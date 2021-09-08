package typ

import (
	"fmt"
	"time"
)

type LoginReq struct {
	// 4-24
	Username string `json:"username" binding:"required,gte=4,lte=24"`
	// must md5
	Password string `json:"password" binding:"required,len=32"`
}

const dateTime = "2006-01-02 15:04:05"

type TimeStringReq struct {
	// format 2006-01-02 15:04:05
	StartTime string `json:"startTime" form:"startTime"`
	EndTime   string `json:"endTime" form:"endTime"`
}

func (req TimeStringReq) FieldSQLCond(field string) []string {
	cond := make([]string, 0)
	if field == "" {
		return cond
	}
	if req.StartTime != "" {
		_, err := time.Parse(dateTime, req.StartTime)
		if err == nil {
			cond = append(cond, fmt.Sprintf("%s >= %s", field, req.StartTime))
		}
	}
	if req.EndTime != "" {
		_, err := time.Parse(dateTime, req.EndTime)
		if err == nil {
			cond = append(cond, fmt.Sprintf("%s < %s", field, req.EndTime))
		}
	}
	return cond
}

func (req TimeStringReq) FieldMgoBson() (map[string]interface{}, bool) {
	filter := map[string]interface{}{}
	if req.StartTime != "" {
		s, err := time.Parse(dateTime, req.StartTime)
		if err == nil {
			filter["$gte"] = s
		}
	}
	if req.EndTime != "" {
		e, err := time.Parse(dateTime, req.EndTime)
		if err == nil {
			filter["$lt"] = e
		}
	}
	if len(filter) == 0 {
		return filter, false
	}
	return filter, true
}

type TimestampReq struct {
	// Unix timestamp Sec
	StartTimestamp int64 `json:"startTimestamp" form:"startTimestamp,default=0" binding:"omitempty,min=0"`
	EndTimestamp   int64 `json:"endTimestamp" form:"endTimestamp,default=0" binding:"omitempty,gtefield=StartTimestamp"`
}

func (req TimestampReq) FieldSQLCond(field string) []string {
	cond := make([]string, 0)
	if field == "" {
		return cond
	}
	if req.StartTimestamp > 0 {
		s := time.Unix(req.StartTimestamp, 0)
		cond = append(cond, fmt.Sprintf("%s >= %s", field, s.Format(dateTime)))
	}
	if req.EndTimestamp > 0 {
		e := time.Unix(req.EndTimestamp, 0)
		cond = append(cond, fmt.Sprintf("%s < %s", field, e.Format(dateTime)))
	}
	return cond
}

func (req TimestampReq) FieldMgoBson() (map[string]interface{}, bool) {
	filter := map[string]interface{}{}
	if req.StartTimestamp > 0 {
		s := time.Unix(req.StartTimestamp, 0)
		filter["$gte"] = s
	}
	if req.EndTimestamp > 0 {
		e := time.Unix(req.EndTimestamp, 0)
		filter["$lt"] = e
	}
	if len(filter) == 0 {
		return filter, false
	}
	return filter, true
}

type PageReq struct {
	Page int `json:"page" form:"page,default=1" binding:"required,gte=1"`
	Size int `json:"size" form:"size,default=20" binding:"required,gte=1,lte=1000"`
}

func (req PageReq) LimitStart() (limit, start int) {
	return req.Size, (req.Page - 1) * req.Size
}

func (req PageReq) Skip() int64 {
	return int64((req.Page - 1) * req.Size)
}

type IdReq struct {
	Id int64 `json:"id" form:"id" binding:"required,gt=0"`
}

type IdOmitReq struct {
	IdReq `json:"id" form:"id" binding:"omitempty,gt=0"`
}

type UidReq struct {
	Uid int64 `json:"uid" form:"uid" binding:"required,min=1"`
}

type UidOmitReq struct {
	UidReq `json:"uid" form:"uid" binding:"omitempty,min=1"`
}
