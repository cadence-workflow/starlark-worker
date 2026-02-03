package internal

import (
	"context"
	"time"

	"github.com/cadence-workflow/starlark-worker/encoded"
	"go.uber.org/zap"
)

type Workflow interface {
	GetLogger(ctx Context) *zap.Logger
	GetActivityLogger(ctx context.Context) *zap.Logger
	GetActivityInfo(ctx context.Context) ActivityInfo
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
	NewCustomError(reason string, details ...interface{}) CustomError
	NewFuture(ctx Context) (Future, Settable)
	NewBatchFuture(ctx Context, batchSize int, factories []func(ctx Context) Future) (BatchFuture, error)
	Go(ctx Context, f func(ctx Context))
	SideEffect(ctx Context, f func(ctx Context) interface{}) encoded.Value
	Now(ctx Context) time.Time
	Sleep(ctx Context, d time.Duration) (err error)
	IsCanceledError(ctx Context, err error) bool
	WithRetryPolicy(ctx Context, retryPolicy RetryPolicy) Context
	NewSelector(ctx Context) Selector
}
