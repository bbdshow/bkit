package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Conn struct {
	Database string `defval:"test"`
	URI      string `defval:"mongodb://127.0.0.1:27017/admin" null:""`
}

type Config struct {
	Conns []Conn
}

// Groups Mongodb instance groups
type Groups struct {
	Config Config

	lock      sync.RWMutex
	instances map[string]*Database
}

// NewGroups  mongo instance groups
func NewGroups(cfg Config) (*Groups, error) {
	g := &Groups{
		Config:    cfg,
		instances: make(map[string]*Database),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	g.lock.Lock()
	defer g.lock.Unlock()

	for _, conn := range cfg.Conns {
		instance, err := NewDatabase(ctx, conn.URI, conn.Database)
		if err != nil {
			return nil, err
		}
		g.instances[conn.Database] = instance
	}
	return g, nil
}

// GetInstance get instance by database
func (g *Groups) GetInstance(database string) (*Database, error) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	instance, ok := g.instances[database]
	if !ok {
		return nil, ErrInstanceNotFound
	}
	return instance, nil
}

func (g *Groups) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var errs []error
	for _, i := range g.instances {
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
