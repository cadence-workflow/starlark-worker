package internal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/cadence-workflow/starlark-worker/encoded"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/cadence-workflow/starlark-worker/star"
	jsoniter "github.com/json-iterator/go"
	"go.starlark.net/starlark"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	tempactivity "go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	tmpworker "go.temporal.io/sdk/worker"
	temp "go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"io"
	"reflect"
	"time"
)

// Workflow checks that TemporalWorkflow implements the Workflow interface.
var _ Workflow = (*TemporalWorkflow)(nil)

// TemporalWorkflow is a wrapper around the Temporal SDK workflow interface.
type TemporalWorkflow struct{}

// WithRetryPolicy sets the retry policy for the workflow execution.
func (w TemporalWorkflow) WithRetryPolicy(ctx Context, retryPolicy RetryPolicy) Context {
	tempRetryPolicy := temporal.RetryPolicy{
		InitialInterval:        retryPolicy.InitialInterval,
		BackoffCoefficient:     retryPolicy.BackoffCoefficient,
		MaximumInterval:        retryPolicy.MaximumInterval,
		MaximumAttempts:        retryPolicy.MaximumAttempts,
		NonRetryableErrorTypes: retryPolicy.NonRetriableErrorReasons,
	}
	return temp.WithRetryPolicy(ctx.(temp.Context), tempRetryPolicy)
}

// IsCanceledError checks if the error is a CanceledError.
func (w TemporalWorkflow) IsCanceledError(ctx Context, err error) bool {
	var canceledError *temporal.CanceledError
	return errors.As(err, &canceledError)
}

// temporalWorkflowInfo is a wrapper around the Temporal SDK workflow context.
type tempWorkflowInfo struct {
	context temp.Context
}

// temporalFuture is a wrapper around the Temporal SDK future interface.
type temporalFuture struct {
	f temp.Future
}

// temporalChildWorkflowFuture is a wrapper around the Temporal SDK child workflow future interface.
type temporalChildWorkflowFuture struct {
	cf temp.ChildWorkflowFuture
}

// temporalSettable is a wrapper around the Temporal SDK settable interface.
type temporalSettable struct {
	s temp.Settable
}

// temporalSelector implements Selector interface
type temporalSelector struct {
	s temp.Selector
}

// TemporalWorker is a wrapper around the Temporal SDK worker interface.
type TemporalWorker struct {
	Worker tmpworker.Worker
}

// Get returns the result of the future.
func (f *temporalFuture) Get(ctx Context, valPtr interface{}) error {
	return f.f.Get(ctx.(temp.Context), valPtr)
}

// IsReady checks if the future is ready.
func (f *temporalFuture) IsReady() bool {
	return f.f.IsReady()
}

// RegisterWorkflow registers a workflow with the Temporal worker.
func (tw *TemporalWorker) RegisterWorkflow(wf interface{}, funcName string) {
	tw.Worker.RegisterWorkflowWithOptions(UpdateWorkflowFunctionContextArgument(wf, reflect.TypeOf((*temp.Context)(nil)).Elem()), temp.RegisterOptions{
		Name: funcName,
	})
}

// RegisterActivity registers an activity with the Temporal worker.
func (tw *TemporalWorker) RegisterActivity(a interface{}) {
	tw.Worker.RegisterActivity(a)
}

// Start starts the Temporal worker.
func (tw *TemporalWorker) Start() error {
	return tw.Worker.Start()
}

// Stop stops the Temporal worker.
func (tw *TemporalWorker) Stop() {
	tw.Worker.Stop()
}

// Run runs the Temporal worker and blocks until it is interrupted.
func (tw *TemporalWorker) Run(interruptCh <-chan interface{}) error {
	return tw.Worker.Run(interruptCh)
}

// RegisterWorkflowWithOptions registers a workflow with the Temporal worker using options.
func (tw *TemporalWorker) RegisterWorkflowWithOptions(wf interface{}, options RegisterWorkflowOptions) {
	tw.Worker.RegisterWorkflowWithOptions(UpdateWorkflowFunctionContextArgument(wf, reflect.TypeOf((*temp.Context)(nil)).Elem()), temp.RegisterOptions{
		Name: options.Name,
		// Optional: Provides a Versioning Behavior to workflows of this type. It is required
		// when WorkerOptions does not specify [DeploymentOptions.DefaultVersioningBehavior],
		// [DeploymentOptions.DeploymentSeriesName] is set, and [UseBuildIDForVersioning] is true.
		// NOTE: Experimental
		VersioningBehavior:            temp.VersioningBehavior(options.VersioningBehavior),
		DisableAlreadyRegisteredCheck: options.DisableAlreadyRegisteredCheck,
	})
}

// RegisterActivityWithOptions registers an activity with the Temporal worker using options.
func (tw *TemporalWorker) RegisterActivityWithOptions(w interface{}, options RegisterActivityOptions) {
	tw.Worker.RegisterActivityWithOptions(w, tempactivity.RegisterOptions{
		Name:                          options.Name,
		SkipInvalidStructFunctions:    options.SkipInvalidStructFunctions,
		DisableAlreadyRegisteredCheck: options.DisableAlreadyRegisteredCheck,
	})
}

func (s *temporalSettable) SetValue(value interface{}) {
	s.s.SetValue(value)
}

func (s *temporalSettable) SetError(err error) {
	s.s.SetError(err)
}

func (s *temporalSettable) Set(value interface{}, err error) {
	s.s.Set(value, err)
}

func (s *temporalSettable) Chain(future Future) {
	s.s.Chain(future.(*temporalFuture).f)
}

func (f *temporalChildWorkflowFuture) Get(ctx Context, valPtr interface{}) error {
	return f.cf.Get(ctx.(temp.Context), valPtr)
}

func (f *temporalChildWorkflowFuture) IsReady() bool {
	return f.cf.IsReady()
}
func (f *temporalChildWorkflowFuture) GetChildWorkflowExecution() Future {
	future := f.cf.GetChildWorkflowExecution()
	return &temporalFuture{f: future}
}

func (f *temporalChildWorkflowFuture) SignalChildWorkflow(ctx Context, signalName string, data interface{}) Future {
	future := f.cf.SignalChildWorkflow(ctx.(temp.Context), signalName, data)
	return &temporalFuture{f: future}
}

func (w *tempWorkflowInfo) ExecutionID() string {
	return temp.GetInfo(w.context).WorkflowExecution.ID
}
func (w *tempWorkflowInfo) RunID() string {
	return temp.GetInfo(w.context).WorkflowExecution.RunID
}

var _ Workflow = (*TemporalWorkflow)(nil)

func (w TemporalWorkflow) GetLogger(ctx Context) *zap.Logger {
	logger := temp.GetLogger(ctx.(temp.Context))

	if zl, ok := logger.(*ZapLoggerAdapter); ok {
		zap := zl.Zap()
		return zap
	}
	return zap.NewNop()
}

func (w TemporalWorkflow) GetActivityLogger(ctx context.Context) *zap.Logger {
	logger := tempactivity.GetLogger(ctx)
	if zl, ok := logger.(*ZapLoggerAdapter); ok {
		zap := zl.Zap()
		return zap
	}
	return zap.NewNop()
}

func (w TemporalWorkflow) GetInfo(ctx Context) IInfo {
	return &tempWorkflowInfo{
		context: ctx.(temp.Context),
	}
}

func (w TemporalWorkflow) ExecuteActivity(ctx Context, activity interface{}, args ...interface{}) Future {
	f := temp.ExecuteActivity(ctx.(temp.Context), activity, args...)
	return &temporalFuture{f: f}
}

func (w TemporalWorkflow) ExecuteChildWorkflow(ctx Context, name interface{}, args ...interface{}) ChildWorkflowFuture {
	f := temp.ExecuteChildWorkflow(ctx.(temp.Context), name, args...)
	return &temporalChildWorkflowFuture{cf: f}
}

func (w TemporalWorkflow) WithValue(parent Context, key interface{}, val interface{}) Context {
	return temp.WithValue(parent.(temp.Context), key, val)
}

func (w TemporalWorkflow) NewDisconnectedContext(parent Context) (ctx Context, cancel func()) {
	return temp.NewDisconnectedContext(parent.(temp.Context))
}

func (w TemporalWorkflow) GetMetricsScope(ctx Context) interface{} {
	return temp.GetMetricsHandler(ctx.(temp.Context))

}

func (w *TemporalWorkflow) WithTaskList(ctx Context, name string) Context {
	return temp.WithTaskQueue(ctx.(temp.Context), name)
}

func (w *TemporalWorkflow) WithActivityOptions(ctx Context, options ActivityOptions) Context {
	cadOptions := temp.ActivityOptions{
		TaskQueue:              options.TaskList,
		ScheduleToCloseTimeout: options.ScheduleToCloseTimeout,
		ScheduleToStartTimeout: options.ScheduleToStartTimeout,
		StartToCloseTimeout:    options.StartToCloseTimeout,
		HeartbeatTimeout:       options.HeartbeatTimeout,
		WaitForCancellation:    options.WaitForCancellation,
		ActivityID:             options.ActivityID,
		DisableEagerExecution:  options.DisableEagerExecution,
		Summary:                options.Summary,
	}
	if options.RetryPolicy != nil {
		cadOptions.RetryPolicy = &temporal.RetryPolicy{
			NonRetryableErrorTypes: options.RetryPolicy.NonRetriableErrorReasons,
			InitialInterval:        options.RetryPolicy.InitialInterval,
			BackoffCoefficient:     options.RetryPolicy.BackoffCoefficient,
			MaximumInterval:        options.RetryPolicy.MaximumInterval,
			MaximumAttempts:        options.RetryPolicy.MaximumAttempts,
		}
	}
	// Handle VersioningIntent mapping if needed
	if options.VersioningIntent != 0 {
		cadOptions.VersioningIntent = temporal.VersioningIntent(options.VersioningIntent)
	}
	return temp.WithActivityOptions(ctx.(temp.Context), cadOptions)
}

// WithChildOptions sets the child workflow options for the workflow execution.
func (w TemporalWorkflow) WithChildOptions(ctx Context, cwo ChildWorkflowOptions) Context {
	var retryPolicy *temporal.RetryPolicy
	if cwo.RetryPolicy != nil {
		retryPolicy = &temporal.RetryPolicy{
			InitialInterval:        cwo.RetryPolicy.InitialInterval,
			BackoffCoefficient:     cwo.RetryPolicy.BackoffCoefficient,
			MaximumInterval:        cwo.RetryPolicy.MaximumInterval,
			MaximumAttempts:        cwo.RetryPolicy.MaximumAttempts,
			NonRetryableErrorTypes: cwo.RetryPolicy.NonRetriableErrorReasons,
		}
	}
	// Construct temporal.SearchAttributes
	if cwo.SearchAttributes != nil {
		for k, v := range cwo.SearchAttributes {
			typed := temporal.NewSearchAttributeKeyString(k)
			updates := typed.ValueSet(v.(string))
			// TODO: Check if this is the correct way to set search attributes
			_ = temp.UpsertTypedSearchAttributes(ctx.(temp.Context), updates)
		}
	}
	opt := temp.ChildWorkflowOptions{
		Namespace:                cwo.Domain,
		WorkflowID:               cwo.WorkflowID,
		TaskQueue:                cwo.TaskList,
		WorkflowExecutionTimeout: cwo.ExecutionStartToCloseTimeout,
		WorkflowTaskTimeout:      cwo.TaskStartToCloseTimeout,
		WaitForCancellation:      cwo.WaitForCancellation,
		RetryPolicy:              retryPolicy,
		CronSchedule:             cwo.CronSchedule,
		Memo:                     cwo.Memo,
		TypedSearchAttributes:    temp.GetTypedSearchAttributes(ctx.(temp.Context)),
		ParentClosePolicy:        enumspb.PARENT_CLOSE_POLICY_UNSPECIFIED,
		VersioningIntent:         0,
	}

	if _, ok := enumspb.WorkflowIdReusePolicy_name[int32(cwo.WorkflowIDReusePolicy)]; ok {
		opt.WorkflowIDReusePolicy = convertCadenceToTemporalReusePolicy(cwo.WorkflowIDReusePolicy)
	}
	return temp.WithChildOptions(ctx.(temp.Context), opt)
}

func convertCadenceToTemporalReusePolicy(cadenceVal int) enumspb.WorkflowIdReusePolicy {
	switch cadenceVal {
	case 0:
		return enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY
	case 1:
		return enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE
	case 2:
		return enumspb.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE
	case 3:
		return enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING
	default:
		return enumspb.WORKFLOW_ID_REUSE_POLICY_UNSPECIFIED
	}
}

func (w TemporalWorkflow) SetQueryHandler(ctx Context, queryType string, handler interface{}) error {
	return temp.SetQueryHandler(ctx.(temp.Context), queryType, handler)
}

func (w TemporalWorkflow) WithWorkflowDomain(ctx Context, name string) Context {
	opts := ChildWorkflowOptions{
		Domain: name,
	}
	return w.WithChildOptions(ctx.(temp.Context), opts)
}

func (w TemporalWorkflow) WithWorkflowTaskList(ctx Context, name string) Context {
	return temp.WithTaskQueue(ctx.(temp.Context), name)
}

func (w TemporalWorkflow) NewCustomError(reason string, details ...interface{}) CustomError {
	var message = reason

	// Check if the first detail is a map[string]interface{} and has key "error"
	if len(details) > 0 {
		if detailMap, ok := details[0].(map[string]interface{}); ok {
			if errVal, exists := detailMap["error"]; exists {
				if str, ok := errVal.(string); ok {
					message = str
				}
			}
		}
	}

	err := temporal.NewApplicationError(message, reason, details...)
	return &TemporalCustomError{
		ApplicationError: *err.(*temporal.ApplicationError),
	}
}

func (w TemporalWorkflow) NewFuture(ctx Context) (Future, Settable) {
	f, s := temp.NewFuture(ctx.(temp.Context))
	return &temporalFuture{f: f}, &temporalSettable{s: s}
}

func (w TemporalWorkflow) SideEffect(ctx Context, f func(ctx Context) interface{}) encoded.Value {
	return temp.SideEffect(ctx.(temp.Context), func(c temp.Context) interface{} {
		return f(c)
	})
}

func (w TemporalWorkflow) Now(ctx Context) time.Time {
	return temp.Now(ctx.(temp.Context))
}

func (w TemporalWorkflow) Sleep(ctx Context, d time.Duration) (err error) {
	return temp.Sleep(ctx.(temp.Context), d)
}

func (w TemporalWorkflow) Go(ctx Context, f func(ctx Context)) {
	temp.Go(ctx.(temp.Context), func(c temp.Context) {
		f(c)
	})
}

// NewSelector creates a new selector for deterministic workflow operations.
func (w TemporalWorkflow) NewSelector(ctx Context) Selector {
	s := temp.NewSelector(ctx.(temp.Context))
	return &temporalSelector{s: s}
}

// AddFuture adds a future case to the selector.
func (s *temporalSelector) AddFuture(future Future, f func(f Future)) Selector {
	s.s.AddFuture(future.(*temporalFuture).f, func(fut temp.Future) {
		f(&temporalFuture{f: fut})
	})
	return s
}

// Select executes the selector and waits for one of the cases to be ready.
func (s *temporalSelector) Select(ctx Context) {
	s.s.Select(ctx.(temp.Context))
}

type ZapLoggerAdapter struct {
	zapLogger *zap.Logger
}

func NewZapLoggerAdapter(z *zap.Logger) *ZapLoggerAdapter {
	return &ZapLoggerAdapter{zapLogger: z}
}

// Implement go.temporal.io/sdk/log.Logger
func (l *ZapLoggerAdapter) Debug(msg string, keyvals ...interface{}) {
	l.zapLogger.Sugar().Debugw(msg, keyvals...)
}

func (l *ZapLoggerAdapter) Info(msg string, keyvals ...interface{}) {
	l.zapLogger.Sugar().Infow(msg, keyvals...)
}

func (l *ZapLoggerAdapter) Warn(msg string, keyvals ...interface{}) {
	l.zapLogger.Sugar().Warnw(msg, keyvals...)
}

func (l *ZapLoggerAdapter) Error(msg string, keyvals ...interface{}) {
	l.zapLogger.Sugar().Errorw(msg, keyvals...)
}

func (l *ZapLoggerAdapter) With(keyvals ...interface{}) log.Logger {
	newLogger := l.zapLogger.Sugar().With(keyvals...)
	return &ZapLoggerAdapter{zapLogger: newLogger.Desugar()}
}

// Expose the underlying *zap.Logger
func (l *ZapLoggerAdapter) Zap() *zap.Logger {
	return l.zapLogger
}

const temporalDelimiter byte = '\n'

// TemporalDataConverter is a Temporal TemporalDataConverter that supports Starlark types.
// Enables passing Starlark values between Temporal workflows and activities.
type TemporalDataConverter struct {
	Logger *zap.Logger
}

var _ converter.DataConverter = (*TemporalDataConverter)(nil)

// ToStrings converts a *commonpb.Payloads object into a slice of human-readable strings.
func (s TemporalDataConverter) ToStrings(payloads *commonpb.Payloads) []string {
	var result []string
	for _, payload := range payloads.Payloads {
		result = append(result, s.ToString(payload))
	}
	return result
}

// ToString converts a single Payload to a human-readable string.
func (s TemporalDataConverter) ToString(payload *commonpb.Payload) string {
	// Attempt to deserialize the payload data into a generic interface
	var data interface{}
	if err := json.Unmarshal(payload.GetData(), &data); err != nil {
		// If deserialization fails, return the raw data as a string
		return string(payload.GetData())
	}
	// Convert the deserialized data to a pretty-printed JSON string
	readableStr, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		// If pretty-printing fails, return the raw data as a string
		return string(payload.GetData())
	}
	return string(readableStr)
}

// ToPayloads converts input values to Temporal's Payloads format
func (s TemporalDataConverter) ToPayloads(values ...interface{}) (*commonpb.Payloads, error) {
	if len(values) == 1 {
		switch v := values[0].(type) {
		case *[]byte:
			return &commonpb.Payloads{
				Payloads: []*commonpb.Payload{
					{Data: *v},
				},
			}, nil
		case *starlark.Bytes:
			return &commonpb.Payloads{
				Payloads: []*commonpb.Payload{
					{Data: []byte(*v)},
				},
			}, nil
		}
	}
	payloads := &commonpb.Payloads{}
	for _, v := range values {
		payload, err := s.ToPayload(v)
		if err != nil {
			return nil, err
		}
		payloads.Payloads = append(payloads.Payloads, payload)
	}
	return payloads, nil
}

// FromPayloads converts Temporal Payloads back into Go types
func (s TemporalDataConverter) FromPayloads(payloads *commonpb.Payloads, to ...interface{}) error {
	for i := 0; i < len(to); i++ {
		if i >= len(payloads.Payloads) {
			return io.EOF
		}
		err := s.FromPayload(payloads.Payloads[i], to[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// ToPayload converts a single Go value to a Temporal Payload
func (s TemporalDataConverter) ToPayload(value interface{}) (*commonpb.Payload, error) {
	var buf bytes.Buffer
	b, err := star.Encode(value) // Try star encoder first
	if _, ok := err.(star.UnsupportedTypeError); ok {
		b, err = jsoniter.Marshal(value) // Fallback to Go JSON encoder
	}
	if err != nil {
		s.Logger.Error("encode-error", ext.ZapError(err)...)
		return nil, err
	}
	buf.Write(b)
	buf.WriteByte(temporalDelimiter)
	return &commonpb.Payload{Data: buf.Bytes()}, nil
}

// FromPayload converts a single Temporal Payload back to a Go value
func (s TemporalDataConverter) FromPayload(payload *commonpb.Payload, to interface{}) error {
	r := bufio.NewReader(bytes.NewReader(payload.Data))
	line, err := r.ReadBytes(temporalDelimiter)
	if err != nil && err != io.EOF {
		s.Logger.Error("decode-error", ext.ZapError(err)...)
		return err
	}
	if len(line) > 0 && line[len(line)-1] == temporalDelimiter {
		line = line[:len(line)-1]
	}

	if err != nil && err != io.EOF {
		s.Logger.Error("decode-error", ext.ZapError(err)...)
		return err
	}

	err = star.Decode(line, to)
	if _, ok := err.(star.UnsupportedTypeError); ok {
		err = jsoniter.Unmarshal(line, to)
	}
	if err != nil {
		s.Logger.Error("decode-error", ext.ZapError(err)...)
		return err
	}
	return nil
}

type TemporalCustomError struct {
	temporal.ApplicationError
}

func (e TemporalCustomError) Error() string {
	return e.ApplicationError.Error()
}

// Reason gets the reason of this custom error
func (e TemporalCustomError) Reason() string {
	return e.ApplicationError.Error()
}

// HasDetails return if this error has strong typed detail data.
func (e TemporalCustomError) HasDetails() bool {
	return e.ApplicationError.HasDetails()
}

// Details extracts strong typed detail data of this custom error. If there is no details, it will return ErrNoData.
func (e TemporalCustomError) Details(d ...interface{}) error {
	return e.ApplicationError.Details(d...)
}

type TemporalCanceledError struct {
	temporal.CanceledError
}

// Error from error interface
func (e TemporalCanceledError) Error() string {
	return e.Error()
}

// HasDetails return if this error has strong typed detail data.
func (e TemporalCanceledError) HasDetails() bool {
	return e.HasDetails()
}

// Details extracts strong typed detail data of this error.
func (e TemporalCanceledError) Details(d ...interface{}) error {
	return e.Details(d...)
}
