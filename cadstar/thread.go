package cadstar

import (
	"strconv"

	"go.starlark.net/starlark"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	threadLocalContextKey = "context"
	envLogLen             = starlark.String("STAR_CORE_LOG_LEN")
	defaultLogLen         = 1000
)

// Thread encapsulates starlark thread and its parent
type Thread struct {
	// Native is the starlark thread
	Native *starlark.Thread

	parent *starlark.Thread
	ctx    workflow.Context
}

func (t *Thread) Finish() {
	if t.parent == nil {
		return
	}

	// Copy shared variables from the child to parent. Those existing on parent will be overwritten.
	for _, k := range GetSharedThreadStorageKeys(t.ctx) {
		v := t.Native.Local(k)
		if v != nil {
			t.parent.SetLocal(k, v)
		}
	}
}

func CreateThread(ctx workflow.Context, parent *starlark.Thread) *Thread {
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
	t := &Thread{
		ctx:    ctx,
		parent: parent,
		Native: &starlark.Thread{
			Print: func(t *starlark.Thread, msg string) {
				logger.Info(msg)
				logs.PushBack(msg)
				if logs.Len() > ll {
					logs.Remove(logs.Front())
				}
			},
		},
	}
	t.Native.SetLocal(threadLocalContextKey, ctx)
	// copy shared thread storage from parent
	if parent != nil {
		for _, k := range GetSharedThreadStorageKeys(ctx) {
			v := parent.Local(k)
			if v != nil {
				t.Native.SetLocal(k, v)
			}
		}
	}

	return t
}

func GetContext(t *starlark.Thread) workflow.Context {
	ctx := t.Local(threadLocalContextKey).(workflow.Context)
	if getGlobals(ctx).isCanceled {
		ctx, _ = workflow.NewDisconnectedContext(ctx)
	}
	return ctx
}

func GetSharedThreadStorageKeys(ctx workflow.Context) []string {
	var res []string
	for _, p := range getGlobals(ctx).plugins {
		res = append(res, p.SharedLocalStorageKeys()...)
	}
	return res
}
