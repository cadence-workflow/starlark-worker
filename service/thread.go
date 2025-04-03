package service

import (
	"strconv"

	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"go.starlark.net/starlark"
	"go.uber.org/zap"
)

const (
	threadLocalContextKey = "context"
	envLogLen             = starlark.String("STAR_CORE_LOG_LEN")
	defaultLogLen         = 1000
)

func CreateThread(ctx workflow.Context) *starlark.Thread {
	logger := workflow.GetLogger(ctx)
	globals := getGlobals(ctx)

	ll := defaultLogLen
	if v, found := globals.getEnviron(envLogLen); found {
		var err error
		if ll, err = strconv.Atoi(v.GoString()); err != nil {
			logger.Error("invalid environ", zap.String("environ_name", envLogLen.GoString()), zap.String("environ_value", v.GoString()))
			ll = defaultLogLen
		}
	}

	logs := globals.logs
	t := &starlark.Thread{
		Print: func(t *starlark.Thread, msg string) {
			logger.Info(msg)
			logs.PushBack(msg)
			if logs.Len() > ll {
				logs.Remove(logs.Front())
			}
		},
	}
	t.SetLocal(threadLocalContextKey, ctx)
	return t
}

func GetContext(t *starlark.Thread) workflow.Context {
	ctx := t.Local(threadLocalContextKey).(workflow.Context)
	if getGlobals(ctx).isCanceled {
		ctx, _ = workflow.NewDisconnectedContext(ctx)
	}
	return ctx
}
