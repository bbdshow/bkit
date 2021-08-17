package mysql

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"xorm.io/xorm"
)

type Xorm struct {
	*xorm.EngineGroup
}

// NewXorm  EngineGroup
func NewXorm(cfg *Config) *Xorm {
	master := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&loc=%v", cfg.Master.Username,
		cfg.Master.Password, cfg.Master.HostPort, cfg.Master.Database, "Asia%2fShanghai")
	conns := []string{master}
	for _, slave := range cfg.Slaves {
		conns = append(conns, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&loc=%v", slave.Username,
			slave.Password, slave.HostPort, slave.Database, "Asia%2fShanghai"))
	}
	engine, err := xorm.NewEngineGroup("mysql", conns)
	if err != nil {
		panic(err)
	}

	engine.SetMaxOpenConns(cfg.MaxOpenConn)
	engine.SetMaxIdleConns(cfg.MaxIdleConn)
	engine.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	engine.ShowSQL(cfg.ShowSQL)

	if err := engine.Ping(); err != nil {
		panic(err)
	}
	x := &Xorm{
		EngineGroup: engine,
	}

	return x
}

// TransactionWithSession 支持事务嵌套，如果事务已经开始，则此事务失效事务的步骤
// 传入的 session 可以是Master/Slave/上层传入
func (x *Xorm) TransactionWithSession(sess *xorm.Session, tx func(sess *xorm.Session) error) error {
	if x.isStartTx(sess) {
		// 只执行，此时相当上个事务的一个步骤
		return tx(sess)
	}

	defer func() {
		_ = sess.Close()
	}()

	if err := sess.Begin(); err != nil {
		return err
	}

	if err := tx(sess); err != nil {
		_ = sess.Rollback()
		return err
	}
	return sess.Commit()
}

func (x *Xorm) isStartTx(sess *xorm.Session) bool {
	lastSql, _ := sess.LastSQL()
	return lastSql == "BEGIN TRANSACTION" || lastSql == "ROLL BACK" || lastSql == "COMMIT"
}

// Transaction Master 事务
func (x *Xorm) Transaction(tx func(sess *xorm.Session) error) error {
	_, err := x.Engine.Transaction(func(session *xorm.Session) (interface{}, error) {
		return nil, tx(session)
	})
	return err
}
