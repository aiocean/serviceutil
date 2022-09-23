package logger

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ignoreHealthCheckCore struct {
	c zapcore.Core
}

func (ig *ignoreHealthCheckCore) Enabled(lv zapcore.Level) bool {
	return ig.c.Enabled(lv)
}

func (ig *ignoreHealthCheckCore) With(fs []zapcore.Field) zapcore.Core {
	return ig.c.With(fs)
}

func (ig *ignoreHealthCheckCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	log.Println("Check log entry", e.Level, e.Message)
	if e.Level == zap.InfoLevel && strings.Contains(e.Message, "finished unary call with code OK") {
		return nil
	}

	return ig.c.Check(e, ce)
}

func (ig *ignoreHealthCheckCore) Write(e zapcore.Entry, fs []zapcore.Field) error {

	return ig.c.Write(e, fs)
}

func (ig *ignoreHealthCheckCore) Sync() error {
	return ig.c.Sync()
}

func NewLogger(ctx context.Context) (*zap.Logger, error) {

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:       true,
		EncoderConfig:     encoderConfig,
		DisableStacktrace: true,
		DisableCaller:     true,
		Encoding:          "json",
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	instanceID := uuid.New().String()
	logger = logger.With(zap.String("K_REVISION", os.Getenv("K_REVISION")), zap.String("instance_id", instanceID))
	// grpc_zap.ReplaceGrpcLoggerV2(logger)
	logger = logger.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return &ignoreHealthCheckCore{c}
	}))

	return logger, nil

}
