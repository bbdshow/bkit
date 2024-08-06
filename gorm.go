package bkit

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MysqlConf struct {
	Database    string `yaml:"database"`
	Host        string `yaml:"host"`
	Port        string `yaml:"port"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	MaxOpenConn int    `yaml:"max_open_conn"`
	MaxIdleConn int    `yaml:"max_idle_conn"`
	Level       int    `yaml:"level"` // Silent = 1, Error = 2, Warn = 3, Info = 4
}

// Validate -
func (mc *MysqlConf) Validate() error {
	if mc.Database == "" {
		return fmt.Errorf("database required")
	}
	if mc.Host == "" {
		return fmt.Errorf("host required")
	}
	if mc.Port == "" {
		return fmt.Errorf("port required")
	}
	if mc.User == "" {
		return fmt.Errorf("user required")
	}
	if mc.Password == "" {
		return fmt.Errorf("password required")
	}
	return nil
}

// GenDSN -
func (mc *MysqlConf) DataSourceName() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		mc.User, mc.Password, mc.Host, mc.Port, mc.Database)
}

// NewGormWithMysql 初始化 Mysql GORM DB
func NewGormWithMysql(mc MysqlConf) (*gorm.DB, error) {
	sqlDB, err := sql.Open("mysql", mc.DataSourceName())
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(mc.MaxOpenConn)
	sqlDB.SetMaxIdleConns(mc.MaxIdleConn)
	sqlDB.SetConnMaxLifetime(time.Hour)
	conf := &gorm.Config{}
	if mc.Level > 0 {
		conf.Logger = logger.Default.LogMode(logger.LogLevel(mc.Level))
	}
	db, err := gorm.Open(
		mysql.New(mysql.Config{
			Conn: sqlDB,
		}), conf,
	)
	return db, err
}

// CloseGormDB -
func CloseGormDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		log.Println("close gorm db error: ", err)
		return err
	}
	return sqlDB.Close()
}
