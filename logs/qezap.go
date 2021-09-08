package logs

import (
	"github.com/bbdshow/qelog/qezap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Qezap
var Qezap *qezap.Logger

func init() {
	// init local Log，not support remote write
	Qezap = qezap.New(qezap.NewConfig(nil, ""), zap.DebugLevel)
}

type Config struct {
	Addr     []string `defval:"127.0.0.1:31082"`
	Module   string   `defval:"example"`
	Filename string   `defval:"./log/logger.log"`
	Level    int      `defval:"-1"` // -1=debug 0=info ...
}

func InitQezap(cfg *Config) {
	_ = Qezap.Close()
	Qezap = qezap.New(qezap.NewConfig(cfg.Addr, cfg.Module).SetFilename(cfg.Filename), zapcore.Level(cfg.Level))
}
