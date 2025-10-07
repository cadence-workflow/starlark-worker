package temporalbatch

import (
	"github.com/cadence-workflow/starlark-worker/internal/temporalbatch/batchfuture"
	"go.temporal.io/sdk/workflow"
)

// This package provides a Temporal-compatible batch future implementation.
// Unlike the cadence-go SDK, the Temporal SDK does not currently expose a batch future API.
// To support the BatchFuture API in starlark-worker, we include this implementation here
// until the Temporal SDK offers native support.

var _ workflow.Future = (BatchFuture)(nil) // to ensure it's compatible

// BatchFuture wraps a collection of futures, and provides some convenience methods for dealing with them in bulk.
type BatchFuture interface {
	// IsReady returns true when all wrapped futures return true from their IsReady
	IsReady() bool
	// Get acts like workflow.Future.Get, but it reads out all wrapped futures into the provided slice pointer.
	// You MUST either
	//	1. provide a pointer to a slice as the value-pointer here, but the slice itself can be nil - it will be allocated and/or resized to fit if needed.
	//	2. provide a nil to indicate that you don't want to collect the results.
	//
	// This call will wait for all futures to resolve, and will then write all results to the output slice in the same order as the input.
	//
	// Any errors encountered are merged with go.uber.org/multierr, so single errors are
	// exposed normally, but multiple ones are bundled in the same way as errors.Join.
	// For consistency when checking individual errors, consider using `multierr.Errors(err)` in all cases,
	// or `GetFutures()[i].Get(ctx, nil)` to get the original errors at each index.
	Get(ctx workflow.Context, valuePtr interface{}) error
	// GetFutures returns a slice of all the wrapped futures.
	// This slice MUST NOT be modified, but the individual futures can be used normally.
	GetFutures() []workflow.Future
}

// NewBatchFuture creates a bounded-concurrency helper for doing bulk work in your workflow.
// It does not reduce the amount of history your workflow stores, so any event-count
// or history-size limits are unaffected - you must still be cautious about the total
// amount of work you do in any workflow.
//
// When NewBatchFuture is called, futures created by the factories will be started concurrently until the concurrency limit (batchSize) is reached.
// The remaining factories will be queued and started as previous futures complete, maintaining the specified concurrency level.
func NewBatchFuture(ctx workflow.Context, batchSize int, factories []func(ctx workflow.Context) workflow.Future) (BatchFuture, error) {
	return batchfuture.NewBatchFuture(ctx, batchSize, factories)
}
