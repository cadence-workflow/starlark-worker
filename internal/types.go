package internal

import (
	"time"
)

type CustomError interface {
	Error() string
	Reason() string
	HasDetails() bool
	Details(d ...interface{}) error
}

type CanceledError interface {
	Error() string
	HasDetails() bool
	Details(d ...interface{}) error
}

type ChildWorkflowOptions struct {
	// Domain of the child workflow.
	// Optional: the current workflow (parent)'s domain will be used if this is not provided.
	Domain string

	// WorkflowID of the child workflow to be scheduled.
	// Optional: an auto generated workflowID will be used if this is not provided.
	WorkflowID string

	// TaskList that the child workflow needs to be scheduled on.
	// Optional: the parent workflow task list will be used if this is not provided.
	TaskList string

	// ExecutionStartToCloseTimeout - The end to end timeout for the child workflow execution.
	// Mandatory: no default
	ExecutionStartToCloseTimeout time.Duration

	// TaskStartToCloseTimeout - The decision task timeout for the child workflow.
	// Optional: default is 10s if this is not provided (or if 0 is provided).
	TaskStartToCloseTimeout time.Duration

	// WaitForCancellation - Whether to wait for cancelled child workflow to be ended (child workflow can be ended
	// as: completed/failed/timedout/terminated/canceled)
	// Optional: default false
	WaitForCancellation bool

	// WorkflowIDReusePolicy - Whether server allow reuse of workflow ID, can be useful
	// for dedup logic if set to WorkflowIdReusePolicyRejectDuplicate
	// Here the value we use Cadence constant, and we map to Temporal equivalent based on the table below.
	//               Cadence Constant	             Value	Meaning	Temporal Equivalent	Value
	//----------------------------------------------------------------------------------------------------
	//WorkflowIDReusePolicyAllowDuplicateFailedOnly	   0	Allow reuse if last run was terminated/cancelled/timeouted/failed	WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY	2
	//WorkflowIDReusePolicyAllowDuplicate	           1	Allow reuse as long as workflow is not running	WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE	1
	//WorkflowIDReusePolicyRejectDuplicate	           2	Never allow reuse	WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE	3
	//WorkflowIDReusePolicyTerminateIfRunning	       3	Terminate if running, then start new	WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING
	//----------------------------------------------------------------------------------------------------
	WorkflowIDReusePolicy int

	// RetryPolicy specify how to retry child workflow if error happens.
	// Optional: default is no retry
	RetryPolicy *RetryPolicy

	// CronSchedule - Optional cron schedule for workflow. If a cron schedule is specified, the workflow will run
	// as a cron based on the schedule. The scheduling will be based on UTC time. Schedule for next run only happen
	// after the current run is completed/failed/timeout. If a RetryPolicy is also supplied, and the workflow failed
	// or timeout, the workflow will be retried based on the retry policy. While the workflow is retrying, it won't
	// schedule its next run. If next schedule is due while workflow is running (or retrying), then it will skip that
	// schedule. Cron workflow will not stop until it is terminated or cancelled (by returning cadence.CanceledError).
	// The cron spec is as following:
	// ┌───────────── minute (0 - 59)
	// │ ┌───────────── hour (0 - 23)
	// │ │ ┌───────────── day of the month (1 - 31)
	// │ │ │ ┌───────────── month (1 - 12)
	// │ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday)
	// │ │ │ │ │
	// │ │ │ │ │
	// * * * * *
	CronSchedule string

	// Memo - Optional non-indexed info that will be shown in list workflow.
	Memo map[string]interface{}

	// SearchAttributes - Optional indexed info that can be used in query of List/Scan/Count workflow APIs (only
	// supported when Cadence server is using ElasticSearch). The key and value type must be registered on Cadence server side.
	// Use GetSearchAttributes API to get valid key and corresponding value type.
	SearchAttributes map[string]interface{}

	// ParentClosePolicy - Optional policy to decide what to do for the child.
	// Default is Terminate (if onboarded to this feature)
	ParentClosePolicy int
}

type RetryPolicy struct {
	// Backoff interval for the first retry. If coefficient is 1.0 then it is used for all retries.
	// Required, no default value.
	InitialInterval time.Duration

	// Coefficient used to calculate the next retry backoff interval.
	// The next retry interval is previous interval multiplied by this coefficient.
	// Must be 1 or larger. Default is 2.0.
	BackoffCoefficient float64

	// Maximum backoff interval between retries. Exponential backoff leads to interval increase.
	// This value is the cap of the interval. Default is 100x of initial interval.
	MaximumInterval time.Duration

	// Maximum time to retry. Either ExpirationInterval or MaximumAttempts is required.
	// When exceeded the retries stop even if maximum retries is not reached yet.
	ExpirationInterval time.Duration

	// Maximum number of attempts. When exceeded the retries stop even if not expired yet.
	// If not set or set to 0, it means unlimited, and rely on ExpirationInterval to stop.
	// Either MaximumAttempts or ExpirationInterval is required.
	MaximumAttempts int32

	// Non-Retriable errors. This is optional. Cadence server will stop retry if error reason matches this list.
	// Error reason for custom error is specified when your activity/workflow return cadence.NewCustomError(reason).
	// Error reason for panic error is "cadenceInternal:Panic".
	// Error reason for any other error is "cadenceInternal:Generic".
	// Error reason for timeouts is: "cadenceInternal:Timeout TIMEOUT_TYPE". TIMEOUT_TYPE could be START_TO_CLOSE or HEARTBEAT.
	// Note, cancellation is not a failure, so it won't be retried.
	NonRetriableErrorReasons []string
}

type ActivityOptions struct {

	// TaskList that the activity needs to be scheduled on.
	// optional: The default task list with the same name as the workflow task list.
	TaskList string

	// ScheduleToCloseTimeout - The end to end timeout for the activity needed.
	// The zero value of this uses default value.
	// Optional: The default value is the sum of ScheduleToStartTimeout and StartToCloseTimeout
	ScheduleToCloseTimeout time.Duration

	// ScheduleToStartTimeout - The queue timeout before the activity starts executed.
	// Mandatory: No default.
	ScheduleToStartTimeout time.Duration

	// StartToCloseTimeout - The timeout from the start of execution to end of it.
	// Mandatory: No default.
	StartToCloseTimeout time.Duration

	// HeartbeatTimeout - The periodic timeout while the activity is in execution. This is
	// the max interval the server needs to hear at-least one ping from the activity.
	// Optional: Default zero, means no heart beating is needed.
	HeartbeatTimeout time.Duration

	// WaitForCancellation - Whether to wait for cancelled activity to be completed(
	// activity can be failed, completed, cancel accepted)
	// Optional: default false
	WaitForCancellation bool

	// ActivityID - Business level activity ID, this is not needed for most of the cases if you have
	// to specify this then talk to cadence team. This is something will be done in future.
	// Optional: default empty string
	ActivityID string

	// RetryPolicy specify how to retry activity if error happens. When RetryPolicy.ExpirationInterval is specified
	// and it is larger than the activity's ScheduleToStartTimeout, then the ExpirationInterval will override activity's
	// ScheduleToStartTimeout. This is to avoid retrying on ScheduleToStartTimeout error which only happen when worker
	// is not picking up the task within the timeout. Retrying ScheduleToStartTimeout does not make sense as it just
	// mark the task as failed and create a new task and put back in the queue waiting worker to pick again. Cadence
	// server also make sure the ScheduleToStartTimeout will not be larger than the workflow's timeout.
	// Same apply to ScheduleToCloseTimeout. See more details about RetryPolicy on the doc for RetryPolicy.
	// Optional: default is no retry
	RetryPolicy *RetryPolicy

	// DisableEagerExecution - If true, eager execution will not be requested, regardless of worker settings.
	// If false, eager execution may still be disabled at the worker level or
	// may not be requested due to lack of available slots.
	// Optional: default false
	DisableEagerExecution bool

	// VersioningIntent - Specifies whether this activity should run on a worker with a compatible
	// build ID or not. See temporal.VersioningIntent.
	// WARNING: Worker versioning is currently experimental
	// Optional: default 0
	VersioningIntent int

	// Summary - A single-line summary for this activity that will appear in UI/CLI. This can be
	// in single-line Temporal Markdown format.
	// Optional: defaults to none/empty.
	// NOTE: Experimental
	Summary string
}

type ChildWorkflowFuture interface {
	Future
	// GetChildWorkflowExecution returns a future that will be ready when child workflow execution started. You can
	// get the WorkflowExecution of the child workflow from the future. Then you can use Workflow ID and RunID of
	// child workflow to cancel or send signal to child workflow.
	//  childWorkflowFuture := workflow.ExecuteChildWorkflow(ctx, child, ...)
	//  var childWE WorkflowExecution
	//  if err := childWorkflowFuture.GetChildWorkflowExecution().Get(ctx, &childWE); err == nil {
	//      // child workflow started, you can use childWE to get the WorkflowID and RunID of child workflow
	//  }
	GetChildWorkflowExecution() Future

	// SignalWorkflowByID sends a signal to the child workflow. This call will block until child workflow is started.
	SignalChildWorkflow(ctx Context, signalName string, data interface{}) Future
}

type Settable interface {
	Set(value interface{}, err error)
	SetValue(value interface{})
	SetError(err error)
	Chain(future Future) // Value (or error) of the future become the same of the chained one.
}

type Future interface {
	Get(ctx Context, valuePtr interface{}) error
	// IsReady will return true Get is guaranteed to not block.
	IsReady() bool
}

type BatchFuture interface {
	// IsReady returns true when all wrapped futures return true from their IsReady
	IsReady() bool
	// Get acts like workflow.Future.Get, but it reads out all wrapped futures into the provided slice pointer.
	Get(ctx Context, valuePtr interface{}) error
	// GetFutures returns a slice of all the wrapped futures.
	// This slice MUST NOT be modified, but the individual futures can be used normally.
	GetFutures() []Future
}

type IInfo interface {
	ExecutionID() string
	RunID() string
}

// Selector provides a deterministic alternative to Go's select statement in workflows.
// It allows waiting on multiple futures in a deterministic way.
type Selector interface {
	AddFuture(future Future, f func(f Future)) Selector
	Select(ctx Context)
}

// ActivityInfo provides comprehensive metadata about an activity execution, unified across
// both Cadence and Temporal workflow engines. This structure contains all the information
// needed to understand the context and lifecycle of an activity execution.
type ActivityInfo struct {
	// TaskToken is an opaque binary token that uniquely identifies this activity task.
	// It's used for activity completion, heartbeat reporting, and task routing within the workflow engine.
	// This token should be included in all activity lifecycle operations (complete, fail, heartbeat).
	TaskToken []byte

	// WorkflowType contains information about the parent workflow that scheduled this activity.
	// May be nil if workflow type information is not available.
	// Includes both the workflow name and path (package information).
	WorkflowType *WorkflowType

	// WorkflowDomain specifies the domain/namespace where the parent workflow is executing.
	// In Cadence: this is the domain name
	// In Temporal: this is the namespace name (mapped to WorkflowDomain for consistency)
	WorkflowDomain string

	// WorkflowExecution contains the unique identifiers of the parent workflow instance
	// that scheduled this activity. Includes both the workflow ID and run ID.
	WorkflowExecution WorkflowExecution

	// ActivityID is a business-level identifier for this specific activity execution.
	// This can be used to identify and correlate activity executions, and is often
	// used for activity deduplication and tracking purposes.
	ActivityID string

	// ActivityType contains information about the activity function being executed,
	// including the activity name and optional path information.
	ActivityType ActivityType

	// TaskList specifies the task list (Cadence) or task queue (Temporal) where this activity was scheduled.
	// This indicates which worker pool is responsible for executing this activity.
	TaskList string

	// HeartbeatTimeout specifies the maximum time allowed between activity heartbeats.
	// If set to 0, no heartbeat is required for this activity.
	// Activities should send heartbeats within this interval to indicate they're still alive and making progress.
	HeartbeatTimeout time.Duration

	// ScheduledTimestamp indicates when the activity was scheduled by the parent workflow.
	// This represents the time the workflow decision was made to execute this activity.
	ScheduledTimestamp time.Time

	// StartedTimestamp indicates when the activity actually began executing on a worker.
	// There may be a delay between ScheduledTimestamp and StartedTimestamp due to task list queuing.
	StartedTimestamp time.Time

	// Deadline specifies the absolute time when this activity execution will timeout.
	// Activities must complete before this deadline or they will be considered failed.
	Deadline time.Time

	// Attempt indicates the current retry attempt number for this activity execution.
	// Starts from 0 for the first attempt, and increments by 1 for each retry attempt
	// when a retry policy is configured for the activity.
	Attempt int32
}

// WorkflowType identifies a workflow type with both human-readable name and technical path information.
// This structure provides metadata about workflow functions across different workflow engines.
type WorkflowType struct {
	// Name is the human-readable identifier for the workflow type.
	// This is the name used when starting workflow executions and in workflow registration.
	// Example: "ProcessOrderWorkflow", "user-onboarding", etc.
	Name string

	// Path contains the fully qualified path or package information for the workflow function.
	// This provides technical details about where the workflow code is located.
	// In Cadence: typically includes the full package path
	// In Temporal: may not always be available (depends on implementation)
	// Example: "github.com/myorg/workflows.ProcessOrderWorkflow"
	Path string
}

// WorkflowExecution Details.
type WorkflowExecution struct {
	ID    string
	RunID string
}

// ActivityType identifies an activity type with both human-readable name and technical path information.
// This structure provides metadata about activity functions across different workflow engines.
type ActivityType struct {
	// Name is the human-readable identifier for the activity type.
	// This is the name used when scheduling activity executions and in activity registration.
	// Example: "ProcessPayment", "send-email", "validate-input", etc.
	Name string

	// Path contains the fully qualified path or package information for the activity function.
	// This provides technical details about where the activity code is located.
	// In Cadence: typically includes the full package path
	// In Temporal: may not always be available (often empty)
	// Example: "github.com/myorg/activities.ProcessPayment"
	Path string
}
