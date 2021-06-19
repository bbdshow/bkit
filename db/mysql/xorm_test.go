package mysql

import (
	"github.com/bbdshow/bkit/tests"
	"testing"
	"time"
)

type KvTable struct {
	Id        int64     `xorm:"not null pk autoincr INT"`
	ValueKey  string    `xorm:"NOT NULL VARCHAR(128) unique comment('ValueKey长度128字符')"`
	ValueData string    `xorm:"NOT NULL VARCHAR(1024) comment('ValueData长度1024字符')"`
	UpdatedAt time.Time `xorm:"updated"`
}

func TestXorm(t *testing.T) {
	cfg := Config{
		ShowSQL: true,
		Master: Conn{
			Username: "root",
			Password: "111111",
			HostPort: "mysql1.dev.crycx.com:3306",
			Database: "crycx_wallet",
		},
		Slaves:          nil,
		MaxOpenConn:     5,
		MaxIdleConn:     10,
		ConnMaxLifetime: 120,
	}

	x := NewXorm(cfg)
	kv := &KvTable{}
	if _, err := x.Where("1 = 1").Get(kv); err != nil {
		t.Fatal(err)
	}
	tests.PrintBeautifyJSON(kv)
}
