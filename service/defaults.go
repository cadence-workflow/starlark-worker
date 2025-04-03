package service

import (
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"go.uber.org/yarpc/yarpcerrors"
	"time"
)

const noTimeout = time.Hour * 24 * 365 * 10 // 10 years, practically - no timeout

var DefaultNonRetriableErrorReasons = []string{
	"cadenceInternal:Panic",                  // panics
	"cadenceInternal:Generic",                // cadence converter errors (similar to invalid-argument)
	"400",                                    // bad-request https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400
	"401",                                    // unauthorized
	"403",                                    // forbidden
	"404",                                    // not-found
	"405",                                    // method-not-allowed
	"502",                                    // bad-gateway
	yarpcerrors.CodeCancelled.String(),       // client error
	yarpcerrors.CodeNotFound.String(),        // client error
	yarpcerrors.CodeAlreadyExists.String(),   // client error
	yarpcerrors.CodeInvalidArgument.String(), // client error
	yarpcerrors.CodeUnauthenticated.String(), // client error
	yarpcerrors.CodePermissionDenied.String(), // client error
	yarpcerrors.CodeUnimplemented.String(),    // client error
	yarpcerrors.CodeDataLoss.String(),         // server error; unrecoverable data corruption
	yarpcerrors.CodeInternal.String(),         // server error; serious error, like panic
}

var DefaultRetryPolicy = workflow.RetryPolicy{
	InitialInterval:          time.Second * 15,
	BackoffCoefficient:       1,
	ExpirationInterval:       time.Minute * 5,
	NonRetriableErrorReasons: DefaultNonRetriableErrorReasons,
}

var DefaultSensorRetryPolicy = workflow.RetryPolicy{
	InitialInterval:          time.Second * 15,
	BackoffCoefficient:       1,
	ExpirationInterval:       noTimeout,
	NonRetriableErrorReasons: DefaultNonRetriableErrorReasons,
}

var DefaultActivityOptions = workflow.ActivityOptions{
	ScheduleToStartTimeout: time.Second * 15,
	StartToCloseTimeout:    time.Second * 15,
	RetryPolicy:            &DefaultRetryPolicy,
}

var DefaultChildWorkflowOptions = workflow.ChildWorkflowOptions{
	ExecutionStartToCloseTimeout: noTimeout,
	RetryPolicy:                  nil,
}
