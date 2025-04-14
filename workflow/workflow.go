package workflow

import (
	"github.com/cadence-workflow/starlark-worker/encoded"
	"github.com/cadence-workflow/starlark-worker/internal"
	"go.uber.org/zap"
	"time"
)

var BackendContextKey = "BackendContextKey"

type (
	// Workflow represents the core interface for a workflow engine backend (e.g., Temporal or Cadence).
	// It defines methods to:
	// - Execute activities and child workflows
	// - Manage context and metadata propagation
	// - Access logging, metrics, and custom behavior
	//
	// This interface is implemented differently for each backend (e.g., `internal_temporal.go`, `internal_cadence.go`).
	// It is attached to a workflow.Context using `WithBackend(ctx, workflowImpl)` and later retrieved using `GetBackend(ctx)`.
	//
	// Example:
	//   backend, ok := workflow.GetBackend(ctx)
	//   if ok {
	//       backend.ExecuteActivity(ctx, SomeActivity, args...)
	//   }
	Workflow = internal.Workflow

	// Context is the core execution context used across all workflow and activity logic.
	// It allows passing metadata, cancellation, and timeout control across workflow and activity boundaries.
	// It abstracts over Temporal/Cadence's context propagation and supports attaching backend-specific values.
	//
	// Key capabilities:
	// - Provides scoped values via `WithValue`.
	// - Propagates cancellation (e.g., parent-child workflow relationships).
	// - Used to access execution-specific APIs (logger, metrics, backend, etc.).
	//
	// This is the primary handle passed into most framework functions such as:
	//   ExecuteActivity(ctx, activityFn, args...)
	//   GetInfo(ctx)
	//   WithActivityOptions(ctx, options)
	//   WithBackend(ctx, w)
	//   GetLogger(ctx)
	//
	// Example Usage:
	//
	//	func MyWorkflow(ctx workflow.Context) error {
	//	    logger := workflow.GetLogger(ctx)
	//	    logger.Info("starting my workflow")
	//
	//	    options := workflow.ActivityOptions{
	//	        StartToCloseTimeout: time.Minute,
	//	    }
	//	    ctx = workflow.WithActivityOptions(ctx, options)
	//
	//	    return workflow.ExecuteActivity(ctx, MyActivity, "input").Get(ctx, nil)
	//	}
	//
	// Note:
	// Context is interface-based and backend-aware — calling GetBackend(ctx) lets you retrieve the registered
	// workflow runtime implementation (Temporal/Cadence), enabling pluggable backends under the same abstraction.
	Context = internal.Context

	// IInfo provides execution metadata about the current workflow instance.
	// Includes fields such as:
	// - WorkflowType
	// - WorkflowID
	// - RunID
	// - TaskQueue
	// - Namespace
	//
	// Accessed via:
	//   info := workflow.GetInfo(ctx)
	//
	// Useful for diagnostics, routing logic, or emitting contextual logs and metrics.
	IInfo = internal.IInfo

	// ActivityOptions specifies how activities are scheduled and executed.
	// Fields include:
	// - StartToCloseTimeout
	// - ScheduleToStartTimeout
	// - RetryPolicy
	// - TaskQueue (optional)
	//
	// Attached to context via:
	//   ctx = workflow.WithActivityOptions(ctx, ActivityOptions{...})
	//
	// Enables customizing timeouts and retries on a per-call basis.
	ActivityOptions = internal.ActivityOptions

	// Future is an abstraction over asynchronous results in workflows.
	// Returned from:
	// - ExecuteActivity
	// - ExecuteChildWorkflow
	//
	// Provides methods:
	// - Get(ctx, &result) – blocks until complete
	// - IsReady()         – non-blocking completion check
	//
	// Enables concurrency patterns and chaining:
	//
	//   future := workflow.ExecuteActivity(ctx, MyActivity)
	//   var result string
	//   err := future.Get(ctx, &result)
	Future = internal.Future

	// ChildWorkflowOptions configures execution of a child workflow.
	// Includes options like:
	// - TaskQueue
	// - RetryPolicy
	// - ExecutionTimeout
	// - WorkflowIDReusePolicy
	//
	// Set with:
	//   ctx = workflow.WithChildOptions(ctx, options)
	//
	// These mirror ActivityOptions but are tailored for workflow-to-workflow calls.
	ChildWorkflowOptions = internal.ChildWorkflowOptions

	// Settable represents a writable Future — allows manual fulfillment of a result or error.
	// Primarily used for internal mocking, stubbing, or combining multiple futures.
	//
	// Example (mocked implementation):
	//   f, s := workflow.NewFuture()
	//   s.Set("value", nil)
	//
	// Used internally by plugins or the future helper package.
	Settable = internal.Settable

	// ChildWorkflowFuture is a specialization of Future with additional context for child workflows.
	// Might provide:
	// - Execution metadata (WorkflowID, RunID)
	// - Cancellation or signal APIs
	//
	// Useful when tracking the lifecycle of child workflows or reacting to failures.
	//
	// Returned from ExecuteChildWorkflow() in backends.
	ChildWorkflowFuture = internal.ChildWorkflowFuture

	// RetryPolicy defines retry behavior for workflows or activities.
	// Fields include:
	// - InitialInterval
	// - BackoffCoefficient
	// - MaximumAttempts
	// - NonRetryableErrorTypes
	//
	// Example:
	//   workflow.WithActivityOptions(ctx, ActivityOptions{
	//       RetryPolicy: RetryPolicy{
	//           InitialInterval:    1 * time.Second,
	//           MaximumAttempts:    5,
	//           BackoffCoefficient: 2.0,
	//       },
	//   })
	//
	// Helps enforce robustness in distributed calls.
	RetryPolicy = internal.RetryPolicy

	// CustomError represents a domain-specific error that can be passed across activity/workflow boundaries.
	// Encodes a string "reason" and optional structured "details".
	//
	// Example:
	//   return workflow.NewCustomError("validation_failed", map[string]interface{}{"field": "name"})
	//
	// Use workflow errors.As() to unwrap and inspect the details in the caller.
	// Useful for safe, structured cross-boundary error handling.
	CustomError = internal.CustomError

	// CanceledError indicates that a workflow or activity was canceled, usually due to context cancellation.
	//
	// Should be checked in activities to exit early:
	//
	//   select {
	//   case <-ctx.Done():
	//       return workflow.CanceledError{}
	//   case val := <-input:
	//       return val
	//   }
	//
	// Used extensively in graceful shutdowns, parent-child propagation, and timed operations.
	CanceledError = internal.CanceledError
)

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

func NewCustomError(ctx Context, reason string, details ...interface{}) CustomError {
	if backend, ok := GetBackend(ctx); ok {
		return backend.NewCustomError(reason, details...)
	}
	return nil
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

func WithRetryPolicy(ctx Context, retryPolicy RetryPolicy) Context {
	if backend, ok := GetBackend(ctx); ok {
		return backend.WithRetryPolicy(ctx, retryPolicy)
	}
	return ctx
}
