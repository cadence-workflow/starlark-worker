// Package activity provides cross-platform utilities for activity execution within
// workflow systems. This package offers a unified interface for accessing activity
// metadata and logging capabilities that works consistently across both Cadence and
// Temporal workflow engines.
//
// The package abstracts engine-specific differences and provides a common API for:
//   - Retrieving activity execution metadata (GetInfo)
//   - Accessing activity-scoped logging (GetLogger)
//   - Working with unified data structures (Info type alias)
//
// Key benefits:
//   - Write once, run anywhere: Activity code works with both Cadence and Temporal
//   - Unified data structures: Consistent field names and types across engines
//   - Engine abstraction: No need to import engine-specific packages in activity code
//   - Graceful handling: Automatic mapping of engine differences and missing fields
//
// This package is designed for use within activity functions to access execution
// context and provide portable activity implementations.
package activity

import (
	"context"
	"github.com/cadence-workflow/starlark-worker/internal"
	"github.com/cadence-workflow/starlark-worker/workflow"
	"go.uber.org/zap"
)

type (
	// Info is a type alias for internal.ActivityInfo that provides a unified interface
	// for accessing activity execution metadata across both Cadence and Temporal workflow engines.
	// This alias allows activity code to access comprehensive execution details without
	// importing internal packages directly.
	//
	// The Info structure contains essential metadata about the current activity execution including:
	//   - TaskToken: Unique identifier for task completion and heartbeat operations
	//   - WorkflowExecution: Parent workflow ID and run ID that scheduled this activity
	//   - ActivityID: Business-level identifier for this activity execution
	//   - ActivityType: Name and path information about the activity function
	//   - TaskList/TaskQueue: Worker pool where the activity was scheduled
	//   - Timing information: Scheduled, started timestamps, and deadline
	//   - Retry information: Current attempt number and heartbeat timeout
	//   - Domain/Namespace: Execution environment details
	//
	// This type alias enables portable activity code that works identically across
	// both Cadence and Temporal implementations.
	Info = internal.ActivityInfo
)

// GetLogger retrieves a logger instance configured for the current activity execution.
// This function provides cross-platform access to activity-scoped logging that works
// identically with both Cadence and Temporal workflow engines.
//
// Parameters:
//   - ctx: The activity context passed to the activity function. Must be a valid activity
//          execution context obtained from the workflow engine.
//
// Returns:
//   - *zap.Logger: A logger instance configured with activity-specific context and metadata.
//                  Returns nil if the context is not a valid activity context or no backend
//                  is available.
//
// The returned logger is automatically configured with activity execution context,
// enabling structured logging that includes activity and workflow identifiers for
// better traceability and debugging.
//
// Cross-Platform Behavior:
//   - Works identically with both Cadence and Temporal workflow engines
//   - Automatically detects the workflow backend and uses appropriate logging mechanisms
//   - Provides consistent logging behavior regardless of underlying engine
//
// Usage Examples:
//
//   Basic structured logging:
//     func MyActivity(ctx context.Context, data interface{}) error {
//       logger := activity.GetLogger(ctx)
//       if logger != nil {
//         logger.Info("Processing activity", zap.Any("input", data))
//       }
//       // ... activity logic
//       return nil
//     }
//
//   Error logging with context:
//     func ProcessDataActivity(ctx context.Context, id string) error {
//       logger := activity.GetLogger(ctx)
//       if err := processData(id); err != nil {
//         if logger != nil {
//           logger.Error("Failed to process data",
//             zap.String("dataID", id),
//             zap.Error(err))
//         }
//         return err
//       }
//       return nil
//     }
//
// The logger automatically includes workflow and activity context in log entries,
// making it easier to trace activity executions across distributed workflow systems.
func GetLogger(ctx context.Context) *zap.Logger {
	if b, ok := workflow.GetBackend(ctx); ok {
		return b.GetActivityLogger(ctx)
	}
	return nil
}

// GetInfo retrieves comprehensive metadata about the currently executing activity.
// This is the primary function for activity code to access execution context information
// in a cross-platform manner that works identically with both Cadence and Temporal.
//
// Parameters:
//   - ctx: The activity context passed to the activity function. Must be a valid activity
//          execution context obtained from the workflow engine.
//
// Returns:
//   - Info: A comprehensive structure containing activity execution metadata including:
//     * Unique identifiers (TaskToken, ActivityID, WorkflowExecution details)
//     * Timing information (scheduled time, start time, deadline)
//     * Execution context (domain/namespace, task list/queue, workflow type)
//     * Retry details (attempt number, heartbeat timeout)
//     * Activity type information (name and optional path)
//
// If the context is not a valid activity context or no workflow backend is available,
// returns an empty Info structure. Activity code should check for valid data before
// using the returned information.
//
// Cross-Platform Behavior:
//   - Works identically with both Cadence and Temporal workflow engines
//   - Automatically maps engine-specific field names (e.g., namespace↔domain, taskQueue↔taskList)
//   - Provides unified data structure regardless of underlying engine
//   - Gracefully handles engine differences (e.g., missing path information in Temporal)
//
// Usage Examples:
//
//   Basic usage within an activity:
//     func MyActivity(ctx context.Context, input string) (string, error) {
//       info := activity.GetInfo(ctx)
//       logger.Info("Activity starting",
//         "activityID", info.ActivityID,
//         "attempt", info.Attempt,
//         "workflowID", info.WorkflowExecution.ID)
//       // ... activity logic
//     }
//
//   Conditional logic based on retry attempts:
//     func RetriableActivity(ctx context.Context, data interface{}) error {
//       info := activity.GetInfo(ctx)
//       if info.Attempt > 2 {
//         // Use more conservative approach on later retries
//         return performSlowButReliableOperation(data)
//       }
//       return performFastOperation(data)
//     }
//
//   Progress reporting with activity ID:
//     func LongRunningActivity(ctx context.Context) error {
//       info := activity.GetInfo(ctx)
//       logger.Info("Starting long operation", "activityID", info.ActivityID)
//       for i := 0; i < 100; i++ {
//         // Send heartbeat with progress
//         activity.RecordHeartbeat(ctx, fmt.Sprintf("Progress: %d%%", i))
//         doWork()
//       }
//       return nil
//     }
//
// The function automatically detects the workflow backend (Cadence or Temporal) and
// retrieves the appropriate activity information using the engine's native APIs,
// then converts it to the unified Info structure for consistent access patterns.
func GetInfo(ctx context.Context) Info {
	if b, ok := workflow.GetBackend(ctx); ok {
		return b.GetActivityInfo(ctx)
	}
	return Info{}
}
