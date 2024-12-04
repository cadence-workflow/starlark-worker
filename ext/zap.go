package ext

import (
	"fmt"
	"go.uber.org/zap"
)

func ZapError(err error) []zap.Field {
	return []zap.Field{
		zap.Error(err),
		zap.String("error_type", fmt.Sprintf("%T", err)),
	}
}
