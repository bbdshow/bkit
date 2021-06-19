package mysql

import (
	"xorm.io/xorm"
)

type Config struct {
	ShowSQL         bool `defval:"true"`
	Master          Conn
	Slaves          []Conn
	MaxOpenConn     int  `defval:"50"`
	MaxIdleConn     int  `defval:"200"`
	ConnMaxLifetime uint `defval:"7200"` // sec 空闲连接最大生命周期
}

type Conn struct {
	Username string `defval:"root"`
	Password string `defval:"111111" null:""`
	HostPort string `defval:"localhost:3306"`
	Database string `defval:"test"`
}

type ORM interface {
	xorm.EngineInterface
}

var _ ORM = &Xorm{}
