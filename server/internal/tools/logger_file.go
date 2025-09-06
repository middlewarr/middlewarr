package tools

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var loggers sync.Map

func newFileLogger(proxyID string) *zerolog.Logger {
	lj := &lumberjack.Logger{
		Filename:   filepath.Join(GetLogsPath(), fmt.Sprintf("%s.log", proxyID)),
		MaxSize:    50,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   true,
	}

	l := zerolog.New(lj).With().Timestamp().Str("proxy_id", proxyID).Logger()

	return &l
}

func GetLoggerFile(proxyID string) *zerolog.Logger {
	if l, ok := loggers.Load(proxyID); ok {
		return l.(*zerolog.Logger)
	}

	logger := newFileLogger(proxyID)
	actual, _ := loggers.LoadOrStore(proxyID, logger)

	return actual.(*zerolog.Logger)
}
