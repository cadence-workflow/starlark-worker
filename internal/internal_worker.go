package internal

type RegisterWorkflowOptions struct {
	Name string
	// Workflow type name is equal to function name instead of fully qualified name including function package.
	// This option has no effect when explicit Name is provided.
	EnableShortName               bool
	DisableAlreadyRegisteredCheck bool
	// This is for temporal RegisterOptions
	// Optional: Provides a Versioning Behavior to workflows of this type. It is required
	// when WorkerOptions does not specify [DeploymentOptions.DefaultVersioningBehavior],
	// [DeploymentOptions.DeploymentSeriesName] is set, and [UseBuildIDForVersioning] is true.
	// NOTE: Experimental
	VersioningBehavior int
}

type RegisterActivityOptions struct {
	// When an activity is a function the name is an actual activity type name.
	// When an activity is part of a structure then each member of the structure becomes an activity with
	// this Name as a prefix + activity function name.
	Name string
	// Activity type name is equal to function name instead of fully qualified
	// name including function package (and struct type if used).
	// This option has no effect when explicit Name is provided.
	EnableShortName               bool
	DisableAlreadyRegisteredCheck bool
	// Automatically send heartbeats for this activity at an interval that is less than the HeartbeatTimeout.
	// This option has no effect if the activity is executed with a HeartbeatTimeout of 0.
	// Default: false
	EnableAutoHeartbeat bool
	// This is for temporal activity RegisterOptions
	// When registering a struct with activities, skip functions that are not valid activities. If false,
	// registration panics.
	SkipInvalidStructFunctions bool
}
