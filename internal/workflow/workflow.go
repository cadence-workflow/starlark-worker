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
