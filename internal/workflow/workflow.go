package workflow

import (
	"github.com/cadence-workflow/starlark-worker/internal/encoded"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
	"time"
)

type Workflow interface {
	GetLogger(ctx Context) *zap.Logger
	WithValue(parent Context, key interface{}, val interface{}) Context
	NewDisconnectedContext(parent Context) (ctx Context, cancel func())
	GetMetricsScope(ctx Context) tally.Scope
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
}

var backendWorkflow Workflow

func RegisterBackend(b Workflow) {
	backendWorkflow = b
}

//
//func GetLogger(ctx Context) *zap.Logger {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.GetLogger(ctx)
//}
//
//func NewDisconnectedContext(parent Context) (ctx Context, cancel func()) {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.NewDisconnectedContext(parent)
//}
//
//func GetInfo(ctx Context) IInfo {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.GetInfo(ctx)
//}
//
//func WithValue(ctx Context, key interface{}, value interface{}) Context {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.WithValue(ctx, key, value)
//}
//
//func GetMetricsScope(ctx Context) tally.Scope {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.GetMetricsScope(ctx)
//}
//
//func ExecuteActivity(ctx Context, activity interface{}, args ...interface{}) Future {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.ExecuteActivity(ctx, activity, args...)
//}
//
//func WithTaskList(ctx Context, name string) Context {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.WithTaskList(ctx, name)
//}
//
//func WithActivityOptions(ctx Context, options ActivityOptions) Context {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.WithActivityOptions(ctx, options)
//}
//
//func WithChildOptions(ctx Context, cwo ChildWorkflowOptions) Context {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.WithChildOptions(ctx, cwo)
//}
//
//func SetQueryHandler(ctx Context, queryType string, handler interface{}) error {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.SetQueryHandler(ctx, queryType, handler)
//}
//
//func WithWorkflowDomain(ctx Context, name string) Context {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.WithWorkflowDomain(ctx, name)
//}
//
//func WithWorkflowTaskList(ctx Context, name string) Context {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.WithWorkflowTaskList(ctx, name)
//}
//
//func ExecuteChildWorkflow(ctx Context, childWorkflow interface{}, args ...interface{}) ChildWorkflowFuture {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.ExecuteChildWorkflow(ctx, childWorkflow, args...)
//}
//
//func NewCustomError(reason string, details ...interface{}) error {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.NewCustomError(reason, details...)
//}
//
//func NewFuture(ctx Context) (Future, Settable) {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.NewFuture(ctx)
//}
//
//func Go(ctx Context, f func(ctx Context)) {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	backendWorkflow.Go(ctx, f)
//}
//
//func SideEffect(ctx Context, f func(ctx Context) interface{}) encoded.Value {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.SideEffect(ctx, f)
//}
//
//func Now(ctx Context) time.Time {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.Now(ctx)
//}
//
//func Sleep(ctx Context, d time.Duration) (err error) {
//	if backendWorkflow == nil {
//		panic("workflow backendWorkflow not registered")
//	}
//	return backendWorkflow.Sleep(ctx, d)
//}
