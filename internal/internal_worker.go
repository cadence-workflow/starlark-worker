package internal

// RegisterWorkflowOptions provides unified configuration options for registering workflow functions
// across different workflow engines (Cadence and Temporal). This structure abstracts engine-specific
// registration parameters into a common interface.
type RegisterWorkflowOptions struct {
	// Name specifies an explicit name for the workflow type. If empty, the function name will be used.
	// This name is used by the workflow engine to identify and route workflow executions.
	// For better portability, it's recommended to always specify an explicit name.
	Name string

	// EnableShortName controls whether to use the short function name instead of the fully qualified name
	// (including package path) when Name is not explicitly provided.
	// This option has no effect when an explicit Name is provided.
	//
	// Example:
	//   - With EnableShortName=false: "github.com/myorg/myapp/workflows.MyWorkflow"
	//   - With EnableShortName=true: "MyWorkflow"
	EnableShortName bool

	// DisableAlreadyRegisteredCheck allows re-registration of workflows with the same name without panicking.
	// By default, attempting to register a workflow with a name that's already registered will cause a panic.
	// Setting this to true bypasses that check, which is useful for testing or hot-reloading scenarios.
	// Use with caution in production as it can lead to unexpected behavior.
	DisableAlreadyRegisteredCheck bool

	// VersioningBehavior specifies how the workflow should handle versioning in Temporal.
	// This is a Temporal-specific feature that controls workflow evolution and deployment strategies.
	// Ignored by Cadence workers.
	//
	// Values:
	//   - 0: Unspecified (uses worker's default versioning behavior)
	//   - 1: Pinned - workflow execution is tied to a specific version
	//   - 2: Compatible - workflow can run on compatible versions
	//
	// This field is required when using Temporal's Build ID versioning feature with deployment options.
	// NOTE: This is an experimental Temporal feature and may change in future versions.
	VersioningBehavior int
}

// RegisterActivityOptions provides unified configuration options for registering activity functions
// across different workflow engines (Cadence and Temporal). This structure abstracts engine-specific
// registration parameters into a common interface, supporting both single function and struct-based registration.
type RegisterActivityOptions struct {
	// Name specifies the activity type name used by the workflow engine for routing and identification.
	//
	// For single activity functions:
	//   - This becomes the exact activity type name
	//   - If empty, the function name will be used
	//
	// For struct registration (registering multiple activities at once):
	//   - This becomes a prefix for all exported methods of the struct
	//   - Each method becomes an activity with name: "{Name}{MethodName}"
	//   - If empty, the struct type name will be used as prefix
	//
	// Example:
	//   RegisterActivityWithOptions(&MyActivities{}, RegisterActivityOptions{Name: "MyApp_"})
	//   // Results in activities: "MyApp_ProcessData", "MyApp_SendEmail", etc.
	Name string

	// EnableShortName controls whether to use the short function/struct name instead of the fully qualified name
	// (including package path) when Name is not explicitly provided.
	// This option has no effect when an explicit Name is provided.
	//
	// Example:
	//   - With EnableShortName=false: "github.com/myorg/myapp/activities.ProcessData"
	//   - With EnableShortName=true: "ProcessData"
	EnableShortName bool

	// DisableAlreadyRegisteredCheck allows re-registration of activities with the same name without panicking.
	// By default, attempting to register an activity with a name that's already registered will cause a panic.
	// Setting this to true bypasses that check, which is useful for testing or hot-reloading scenarios.
	// Use with caution in production as it can lead to unexpected behavior.
	DisableAlreadyRegisteredCheck bool

	// EnableAutoHeartbeat automatically sends heartbeats for long-running activities at regular intervals.
	// The heartbeat interval is calculated to be less than the activity's HeartbeatTimeout setting.
	// This option has no effect if the activity is executed with HeartbeatTimeout set to 0.
	//
	// Benefits:
	//   - Prevents activity timeouts for long-running operations
	//   - Provides worker liveness indication to the workflow engine
	//   - Enables activity progress reporting
	//
	// Note: Currently ignored by both Cadence and Temporal implementations, but reserved for future use.
	// Default: false
	EnableAutoHeartbeat bool

	// SkipInvalidStructFunctions controls how struct registration handles methods that don't conform
	// to valid activity signatures. This is a Temporal-specific option, ignored by Cadence.
	//
	// When registering a struct as a collection of activities, the workflow engine examines each
	// exported method to determine if it's a valid activity function. Valid activities must have
	// signatures like: func(context.Context, ...args) (...results, error)
	//
	// Behavior:
	//   - true: Invalid methods are silently skipped, only valid activities are registered
	//   - false: Registration panics if any method is not a valid activity signature
	//
	// This is useful when you have structs that contain both activity methods and utility methods,
	// allowing you to register only the valid activities without restructuring your code.
	//
	// Temporal-specific: This option is only used by Temporal workers. Cadence ignores this field.
	SkipInvalidStructFunctions bool
}
