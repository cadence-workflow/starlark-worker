package temporal

import (
	"go.temporal.io/sdk/log"
	"go.uber.org/zap"
)

type ZapLoggerAdapter struct {
	zapLogger *zap.Logger
}

// NewZapLoggerAdapter returns both a Temporal-compatible logger and keeps zap.Logger
func NewZapLoggerAdapter(z *zap.Logger) *ZapLoggerAdapter {
	return &ZapLoggerAdapter{zapLogger: z}
}

// Implement go.temporal.io/sdk/log.Logger
func (l *ZapLoggerAdapter) Debug(msg string, keyvals ...interface{}) {
	l.zapLogger.Sugar().Debugw(msg, keyvals...)
}

func (l *ZapLoggerAdapter) Info(msg string, keyvals ...interface{}) {
	l.zapLogger.Sugar().Infow(msg, keyvals...)
}

func (l *ZapLoggerAdapter) Warn(msg string, keyvals ...interface{}) {
	l.zapLogger.Sugar().Warnw(msg, keyvals...)
}

func (l *ZapLoggerAdapter) Error(msg string, keyvals ...interface{}) {
	l.zapLogger.Sugar().Errorw(msg, keyvals...)
}

func (l *ZapLoggerAdapter) With(keyvals ...interface{}) log.Logger {
	newLogger := l.zapLogger.Sugar().With(keyvals...)
	return &ZapLoggerAdapter{zapLogger: newLogger.Desugar()}
}

// Expose the underlying *zap.Logger
func (l *ZapLoggerAdapter) Zap() *zap.Logger {
	return l.zapLogger
}
