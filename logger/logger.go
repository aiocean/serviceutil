package logger

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ignoreHealthCheckCore struct {
	c             zapcore.Core
	isHealthCheck bool
}

func (ig ignoreHealthCheckCore) Enabled(lv zapcore.Level) bool {
	return ig.c.Enabled(lv)
}

func (ig ignoreHealthCheckCore) With(fs []zapcore.Field) zapcore.Core {
	for _, f := range fs {
		// GRPC health check
		if f.Key == "grpc.service" && f.String == "grpc.health.v1.Health" {
			ig.isHealthCheck = true
			break
		}

		// HTTP health check
		if f.Key == "url" && strings.HasSuffix(f.String, "/healthz") {
			ig.isHealthCheck = true
			break
		}
	}
	return ignoreHealthCheckCore{
		c:             ig.c.With(fs),
		isHealthCheck: ig.isHealthCheck,
	}
}

func (ig ignoreHealthCheckCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if ig.isHealthCheck {
		return nil
	}

	return ig.c.Check(e, ce)
}

func (ig ignoreHealthCheckCore) Write(e zapcore.Entry, fs []zapcore.Field) error {
	return ig.c.Write(e, fs)
}

func (ig ignoreHealthCheckCore) Sync() error {
	return ig.c.Sync()
}

func NewLogger(ctx context.Context) (*zap.Logger, error) {
	loc := time.FixedZone("UTC+7", 7*60*60)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
			zapcore.RFC3339TimeEncoder(t.In(loc), pae)
		},
		EncodeDuration: zapcore.MillisDurationEncoder,
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

	logger, err := zapConfig.Build(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return ignoreHealthCheckCore{c: c}
	}))
	if err != nil {
		return nil, err
	}

	instanceID := uuid.New().String()
	logger = logger.With(zap.String("K_REVISION", os.Getenv("K_REVISION")), zap.String("instance_id", instanceID))
	// grpc_zap.ReplaceGrpcLoggerV2(logger)

	return logger, nil

}
