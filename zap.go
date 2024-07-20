package bkit

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"
)

// Zap is a global zap logger
var Zap *zap.Logger

func init() {
	if err := InitZapGlobalLogger(-1, "console"); err != nil {
		panic(err)
	}
}

// InitZapGlobalLogger 先临时这样写，后面再优化
func InitZapGlobalLogger(level int, encoding string, outputPaths ...string) error {
	encoder := zap.NewDevelopmentEncoderConfig()
	encoder.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    encoder,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if encoding != "" {
		cfg.Encoding = encoding
	}
	if len(outputPaths) > 0 {
		if err := mkdirPath(outputPaths...); err != nil {
			return fmt.Errorf("mkdir path %v", err)
		}
		cfg.OutputPaths = outputPaths
		cfg.ErrorOutputPaths = outputPaths
	}
	cfg.Level = zap.NewAtomicLevelAt(zapcore.Level(level))
	lg, err := cfg.Build()
	if err != nil {
		return err
	}
	Zap = lg

	return nil
}

func mkdirPath(paths ...string) error {
	for _, p := range paths {
		if !strings.Contains(p, "/") {
			continue
		}
		if err := os.MkdirAll(path.Dir(p), os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// CloseZapGlobalLogger -
func CloseZapGlobalLogger() {
	if Zap != nil {
		if err := Zap.Sync(); err != nil {
			log.Println("close zap sync", err)
		}
	}
}

// GetLevelLogWriter 这是一个特殊的 writer 仅用于某些组件需要 写入日志的情况
func GetLevelLogWriter(level zapcore.Level, msg string) io.Writer {
	if Zap != nil {
		return newZapWriter(Zap, level, msg)
	}
	return log.New(os.Stdout, fmt.Sprintf("%s %s", level.String(), msg), log.LstdFlags).Writer()
}

type wrapZapWriter struct {
	level zapcore.Level
	msg   string
	log   *zap.Logger
}

func newZapWriter(l *zap.Logger, level zapcore.Level, msg string) *wrapZapWriter {
	return &wrapZapWriter{
		level: level,
		msg:   msg,
		log:   l,
	}
}

// Write -
func (w *wrapZapWriter) Write(b []byte) (n int, err error) {
	str := string(b)
	if ce := w.log.Check(w.level, w.msg); ce != nil {
		ce.Write(zap.String("DATA", str))
		return len(b), nil
	}
	return 0, errors.New("not found Write, check level setting")
}
