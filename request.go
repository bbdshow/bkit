package bkit

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

type TimeStringReq struct {
	// format 2006-01-02 15:04:05
	StartTime string `json:"start_time" form:"start_time"`
	EndTime   string `json:"end_time" form:"end_time"`
}

func (req TimeStringReq) FieldSQLCond(field string) []string {
	cond := make([]string, 0)
	if field == "" {
		return cond
	}
	if req.StartTime != "" {
		_, err := time.Parse(time.DateTime, req.StartTime)
		if err == nil {
			cond = append(cond, fmt.Sprintf("%s >= '%s'", field, req.StartTime))
		}
	}
	if req.EndTime != "" {
		_, err := time.Parse(time.DateTime, req.EndTime)
		if err == nil {
			cond = append(cond, fmt.Sprintf("%s < '%s'", field, req.EndTime))
		}
	}
	return cond
}

func (req TimeStringReq) FieldMongo() (map[string]interface{}, bool) {
	filter := map[string]interface{}{}
	if req.StartTime != "" {
		s, err := time.Parse(time.DateTime, req.StartTime)
		if err == nil {
			filter["$gte"] = s
		}
	}
	if req.EndTime != "" {
		e, err := time.Parse(time.DateTime, req.EndTime)
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
	StartTimestamp int64 `json:"start_timestamp" form:"start_timestamp,default=0" binding:"omitempty,min=0"`
	EndTimestamp   int64 `json:"end_timestamp" form:"end_timestamp,default=0" binding:"omitempty,gtefield=StartTimestamp"`
}

func (req TimestampReq) FieldSQLCond(field string) []string {
	cond := make([]string, 0)
	if field == "" {
		return cond
	}
	if req.StartTimestamp > 0 {
		s := time.Unix(req.StartTimestamp, 0)
		cond = append(cond, fmt.Sprintf("%s >= '%s'", field, s.Format(time.DateTime)))
	}
	if req.EndTimestamp > 0 {
		e := time.Unix(req.EndTimestamp, 0)
		cond = append(cond, fmt.Sprintf("%s < '%s'", field, e.Format(time.DateTime)))
	}
	return cond
}

func (req TimestampReq) FieldMongo() (map[string]interface{}, bool) {
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
	Page int `json:"page" form:"page,default=1" binding:"omitempty,gte=1"`            // page index, start index=1
	Size int `json:"size" form:"size,default=20" binding:"omitempty,gte=1,lte=10000"` // page size
}

func (req *PageReq) SetDefaultVal() {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
}

func (req *PageReq) LimitStart() (limit, start int) {
	return req.Size, (req.Page - 1) * req.Size
}

func (req *PageReq) Skip() int64 {
	return int64((req.Page - 1) * req.Size)
}

type IDReq struct {
	ID int64 `json:"id" form:"id" binding:"required,gt=0"` // 当前资源表ID
}

type IDOmitReq struct {
	ID int64 `json:"id" form:"id" binding:"omitempty,min=1"`
}
