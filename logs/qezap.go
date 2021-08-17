package logs

import (
	"github.com/bbdshow/qelog/qezap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Qezap 支持远端的 Zap, 是Zap的超集
var Qezap *qezap.Logger

func init() {
	// 注册本地Log，不具有远程写入能力
	Qezap = qezap.New(qezap.NewConfig(nil, ""), zap.DebugLevel)
}

type Config struct {
	Addr     []string `defval:""`
	Module   string   `defval:""`
	Filename string   `defval:"./log/logger.log"`
	Level    int      `defval:"-1"` // -1=debug 0=info ...
}

func InitQezap(cfg *Config) {
	_ = Qezap.Close()
	Qezap = qezap.New(qezap.NewConfig(cfg.Addr, cfg.Module).SetFilename(cfg.Filename), zapcore.Level(cfg.Level))
}
