package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

// Command is a wrapper for running commands on a database.
type Command struct {
	db *Database
}

// NewCommand create a new Command
func NewCommand(db *Database) *Command {
	return &Command{db: db}
}

type CollStatsResp struct {
	Ok             int              `json:"ok"`
	Ns             string           `json:"ns"`
	Size           int64            `json:"size"`
	Count          int64            `json:"count"`
	AvgObjSize     int64            `json:"avgObjSize"`
	StorageSize    int64            `json:"storageSize"`
	Capped         bool             `json:"capped"`
	TotalIndexSize int64            `json:"totalIndexSize"`
	IndexSizes     map[string]int64 `json:"indexSizes"`
}

// CollStats 获取集合统计信息
func (c *Command) CollStats(ctx context.Context, collection string) (CollStatsResp, error) {
	out := CollStatsResp{}
	err := c.db.RunCommand(ctx, bson.D{{Key: "collStats", Value: collection}}).Decode(&out)
	return out, err
}

type DBStatsResp struct {
	Ok          int32  `json:"ok"`
	DB          string `json:"db"`
	Collections int32  `json:"collections"`
	Objects     int64  `json:"objects"`
	DataSize    int64  `json:"dataSize"`
	StorageSize int64  `json:"storageSize"`
	Indexes     int64  `json:"indexes"`
	IndexSize   int64  `json:"indexSize"`
}

// DBStats 获取数据库统计信息
func (c *Command) DBStats(ctx context.Context) (DBStatsResp, error) {
	out := DBStatsResp{}
	err := c.db.RunCommand(ctx, bson.D{{Key: "dbStats", Value: 1}}).Decode(&out)
	return out, err
}

type HostInfoResp struct {
	System map[string]interface{} `json:"system"`
	Os     map[string]interface{} `json:"os"`
}

// HostInfo 获取主机信息
func (c *Command) HostInfo(ctx context.Context) (HostInfoResp, error) {
	out := HostInfoResp{}
	err := c.db.RunCommand(ctx, bson.D{{Key: "hostInfo", Value: 1}}).Decode(&out)
	return out, err
}
