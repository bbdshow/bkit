package mongo

import (
	"context"
	"fmt"
	"time"
)

type Conn struct {
	Database string `defval:"test"`
	URI      string `defval:"mongodb://127.0.0.1:27017/admin" null:""`
}

type SlotsConn struct {
	Slots []int // 连接对应位置
	Conn
}

type Config struct {
	Master Conn
	Slaves []SlotsConn
}

// Cluster Mongodb 实例集群
type Cluster struct {
	Config Config
	master *Database
	slaves map[int]*Database
}

// NewCluster 简单的多实例管理
func NewCluster(config Config) (*Cluster, error) {
	cluster := &Cluster{
		Config: config,
		slaves: make(map[int]*Database),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	master, err := NewDatabase(ctx, config.Master.URI, config.Master.Database)
	if err != nil {
		return nil, err
	}
	cluster.master = master

	for _, conn := range config.Slaves {
		slave, err := NewDatabase(ctx, conn.URI, conn.Database)
		if err != nil {
			return nil, err
		}
		for _, slot := range conn.Slots {
			cluster.slaves[slot] = slave
		}
	}

	return cluster, nil
}

// 主实例
func (c *Cluster) Master() (*Database, error) {
	if c.master != nil {
		return c.master, nil
	}
	return nil, ErrMasterNotFound
}

// 从实例
func (c *Cluster) Slave(slot int) (*Database, error) {
	store, ok := c.slaves[slot]
	if !ok {
		return nil, ErrSlaveNotFound
	}
	return store, nil
}

func (c *Cluster) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var errs []error
	if c.master != nil {
		if err := c.master.Client().Disconnect(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	for _, slave := range c.slaves {
		if slave != nil {
			if err := slave.Client().Disconnect(ctx); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}
