package mgo

import (
	"context"
	"fmt"
	"time"
)

type DBConn struct {
	Database string
	URI      string
}

type DBGroupConfig struct {
	Conns []DBConn
}

type DBGroup struct {
	cfg DBGroupConfig

	databases map[string]*Database
}

func NewDBGroup(cfg DBGroupConfig) (*DBGroup, error) {
	dg := &DBGroup{
		cfg:       cfg,
		databases: make(map[string]*Database),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, conn := range cfg.Conns {
		db, err := NewDatabase(ctx, conn.URI, conn.Database)
		if err != nil {
			return nil, err
		}
		dg.databases[conn.Database] = db
	}
	return dg, nil
}

func (dg *DBGroup) GetInstance(database string) (*Database, error) {
	db, ok := dg.databases[database]
	if !ok {
		return nil, ErrDatabaseNotFound
	}
	return db, nil
}

func (dg *DBGroup) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var errs []error
	for _, i := range dg.databases {
		if i != nil {
			if err := i.Client().Disconnect(ctx); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}
