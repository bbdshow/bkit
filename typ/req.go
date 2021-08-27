package typ

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"xorm.io/builder"
)

type LoginReq struct {
	// 4-24
	Username string `json:"username" binding:"required,gte=4,lte=24"`
	// 必需MD5
	Password string `json:"password" binding:"required,len=32"`
}

type TimeStringReq struct {
	// 格式 2006-01-02 15:04:05
	StartTime string `json:"startTime" form:"startTime"`
	EndTime   string `json:"endTime" form:"endTime"`
}

func (req TimeStringReq) FieldSQLCond(field string) []builder.Cond {
	cond := make([]builder.Cond, 0)
	if field == "" {
		return cond
	}
	if req.StartTime != "" {
		s, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
		if err == nil {
			cond = append(cond, builder.Gte{field: s.String()})
		}
	}
	if req.EndTime != "" {
		e, err := time.Parse("2006-01-02 15:04:05", req.EndTime)
		if err == nil {
			cond = append(cond, builder.Lt{field: e.String()})
		}
	}
	return cond
}

func (req TimeStringReq) FieldMongoBson() (bson.M, bool) {
	filter := bson.M{}
	if req.StartTime != "" {
		s, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
		if err == nil {
			filter["$gte"] = s
		}
	}
	if req.EndTime != "" {
		e, err := time.Parse("2006-01-02 15:04:05", req.EndTime)
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
	// Unix 时间戳
	StartTimestamp int64 `json:"startTimestamp" form:"startTimestamp,default=0" binding:"omitempty,min=0"`
	EndTimestamp   int64 `json:"endTimestamp" form:"endTimestamp,default=0" binding:"omitempty,gtefield=StartTimestamp"`
}

func (req TimestampReq) FieldSQLCond(field string) []builder.Cond {
	cond := make([]builder.Cond, 0)
	if field == "" {
		return cond
	}
	if req.StartTimestamp > 0 {
		s := time.Unix(req.StartTimestamp, 0)
		cond = append(cond, builder.Gte{field: s.String()})
	}
	if req.EndTimestamp > 0 {
		e := time.Unix(req.EndTimestamp, 0)
		cond = append(cond, builder.Gte{field: e.String()})
	}
	return cond
}

func (req TimestampReq) FieldMongoBson() (bson.M, bool) {
	filter := bson.M{}
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

func (req PageReq) Limit() (limit, start int) {
	return req.Size, (req.Page - 1) * req.Size
}

func (req PageReq) FindPage(opt *options.FindOptions) *options.FindOptions {
	return opt.SetLimit(int64(req.Size)).SetSkip(int64((req.Page - 1) * req.Size))
}

type IdReq struct {
	Id int64 `json:"id" form:"id" binding:"required,gt=0"`
}

type IdOmitReq struct {
	IdReq `json:"id" form:"id" binding:"omitempty,gt=0"`
}

type ObjectIdReq struct {
	Id string `json:"id" form:"id" binding:"required,len=24"`
}

func (req ObjectIdReq) ObjectId() (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(req.Id)
}

type ObjectIdOmitReq struct {
	ObjectIdReq `json:"id" form:"id" binding:"omitempty,len=24"`
}

type UidReq struct {
	// 用户ID
	Uid int64 `json:"uid" form:"uid" binding:"required,min=1"`
}

type UidOmitReq struct {
	UidReq `json:"uid" form:"uid" binding:"omitempty,min=1"`
}
