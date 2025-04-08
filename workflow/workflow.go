package workflow

import (
	"context"
	"fmt"
	"github.com/cadence-workflow/starlark-worker/encoded"
	"go.uber.org/zap"
	"time"
)

var BackendContextKey = "BackendContextKey"

type Workflow interface {
	GetLogger(ctx Context) *zap.Logger
	GetActivityLogger(ctx context.Context) *zap.Logger
	WithValue(parent Context, key interface{}, val interface{}) Context
	NewDisconnectedContext(parent Context) (ctx Context, cancel func())
	GetMetricsScope(ctx Context) interface{}
	ExecuteActivity(ctx Context, activity interface{}, args ...interface{}) Future
	WithTaskList(ctx Context, name string) Context
	GetInfo(ctx Context) IInfo
	WithActivityOptions(ctx Context, options ActivityOptions) Context
	WithChildOptions(ctx Context, cwo ChildWorkflowOptions) Context
	SetQueryHandler(ctx Context, queryType string, handler interface{}) error
	WithWorkflowDomain(ctx Context, name string) Context
	WithWorkflowTaskList(ctx Context, name string) Context
	ExecuteChildWorkflow(ctx Context, childWorkflow interface{}, args ...interface{}) ChildWorkflowFuture
	NewCustomError(reason string, details ...interface{}) error
	NewFuture(ctx Context) (Future, Settable)
	Go(ctx Context, f func(ctx Context))
	SideEffect(ctx Context, f func(ctx Context) interface{}) encoded.Value
	Now(ctx Context) time.Time
	Sleep(ctx Context, d time.Duration) (err error)
	IsCanceledError(ctx Context, err error) bool
	CustomError(ctx Context, err error) (bool, string, string)
}

func GetBackend(ctx Context) (Workflow, bool) {
	backend, ok := ctx.Value(BackendContextKey).(Workflow)
	return backend, ok
}

func WithBackend(parent Context, w Workflow) Context {
	c := w.WithValue(parent, BackendContextKey, w)
	if _, ok := GetBackend(c); !ok {
		panic("failed to set backend in context")
	}
	return c
}

func GetLogger(ctx Context) *zap.Logger {
	if backend, ok := GetBackend(ctx); ok {
		return backend.GetLogger(ctx)
	}
	return nil
}

func WithValue(parent Context, key interface{}, val interface{}) Context {
	if backend, ok := GetBackend(parent); ok {
		return backend.WithValue(parent, key, val)
	}
	return parent
}

func NewDisconnectedContext(parent Context) (Context, func()) {
	if backend, ok := GetBackend(parent); ok {
		return backend.NewDisconnectedContext(parent)
	}
	return nil, func() {}
}

func GetMetricsScope(ctx Context) interface{} {
	if backend, ok := GetBackend(ctx); ok {
		return backend.GetMetricsScope(ctx)
	}
	return nil
}

func ExecuteActivity(ctx Context, activity interface{}, args ...interface{}) Future {
	if backend, ok := GetBackend(ctx); ok {
		return backend.ExecuteActivity(ctx, activity, args...)
	}
	return nil
}

func WithTaskList(ctx Context, name string) Context {
	if backend, ok := GetBackend(ctx); ok {
		return backend.WithTaskList(ctx, name)
	}
	return ctx
}

func GetInfo(ctx Context) IInfo {
	if backend, ok := GetBackend(ctx); ok {
		return backend.GetInfo(ctx)
	}
	return nil
}

func WithActivityOptions(ctx Context, options ActivityOptions) Context {
	if backend, ok := GetBackend(ctx); ok {
		return backend.WithActivityOptions(ctx, options)
	}
	return ctx
}

func WithChildOptions(ctx Context, cwo ChildWorkflowOptions) Context {
	if backend, ok := GetBackend(ctx); ok {
		return backend.WithChildOptions(ctx, cwo)
	}
	return ctx
}

func SetQueryHandler(ctx Context, queryType string, handler interface{}) error {
	if backend, ok := GetBackend(ctx); ok {
		return backend.SetQueryHandler(ctx, queryType, handler)
	}
	return nil
}

func WithWorkflowDomain(ctx Context, name string) Context {
	if backend, ok := GetBackend(ctx); ok {
		return backend.WithWorkflowDomain(ctx, name)
	}
	return ctx
}

func WithWorkflowTaskList(ctx Context, name string) Context {
	if backend, ok := GetBackend(ctx); ok {
		return backend.WithWorkflowTaskList(ctx, name)
	}
	return ctx
}

func ExecuteChildWorkflow(ctx Context, childWorkflow interface{}, args ...interface{}) ChildWorkflowFuture {
	if backend, ok := GetBackend(ctx); ok {
		return backend.ExecuteChildWorkflow(ctx, childWorkflow, args...)
	}
	return nil
}

func NewCustomError(ctx Context, reason string, details ...interface{}) error {
	if backend, ok := GetBackend(ctx); ok {
		return backend.NewCustomError(reason, details...)
	}
	return fmt.Errorf("custom error: %s", reason)
}

func NewFuture(ctx Context) (Future, Settable) {
	if backend, ok := GetBackend(ctx); ok {
		return backend.NewFuture(ctx)
	}
	return nil, nil
}

func Go(ctx Context, f func(ctx Context)) {
	if backend, ok := GetBackend(ctx); ok {
		backend.Go(ctx, f)
	}
}

func SideEffect(ctx Context, f func(ctx Context) interface{}) encoded.Value {
	if backend, ok := GetBackend(ctx); ok {
		return backend.SideEffect(ctx, f)
	}
	return nil
}

func Now(ctx Context) time.Time {
	if backend, ok := GetBackend(ctx); ok {
		return backend.Now(ctx)
	}
	return time.Now()
}

func Sleep(ctx Context, d time.Duration) error {
	if backend, ok := GetBackend(ctx); ok {
		return backend.Sleep(ctx, d)
	}
	time.Sleep(d)
	return nil
}

func IsCanceledError(ctx Context, err error) bool {
	if backend, ok := GetBackend(ctx); ok {
		return backend.IsCanceledError(ctx, err)
	}
	return false
}

func CustomError(ctx Context, err error) (bool, string, string) {
	if backend, ok := GetBackend(ctx); ok {
		return backend.CustomError(ctx, err)
	}
	return false, "", ""
}
