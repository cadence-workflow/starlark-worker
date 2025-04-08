package temporal

import (
	"context"
	"fmt"

	"github.com/cadence-workflow/starlark-worker/temporal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	temp "go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"testing"
	"time"
)

// TemporalTestSuite is a collection of unit tests that verify various guarantees provided by the Temporal workflow engine.
// These tests cover graceful shutdowns, workflow executions, and retry policies, ensuring Temporal's robustness and reliability.
type TemporalTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func TestTempTestSuite(t *testing.T) {
	suite.Run(t, new(TemporalTestSuite))
}

func (s *TemporalTestSuite) BeforeTest(sn, tn string) {

	fmt.Printf(`
-------------------------------- %s.%s --------------------------------

`, sn, tn)

	s.env = s.NewTestWorkflowEnvironment()

	s.env.RegisterWorkflow(Noop)
	s.env.RegisterActivity(NoopActivity)
}

func (s *TemporalTestSuite) AfterTest(_, _ string) {
	s.env.AssertExpectations(s.T())
}

// Test_Graceful_Cancellation verifies that the workflow can gracefully shutdown and clean up resources when a
// cancellation event is triggered.
// This ensures that even long-running or stuck jobs are properly handled without leaving lingering resources.
func (s *TemporalTestSuite) Test_Graceful_Cancellation() {

	// Set up the simulated job processing flow to test graceful shutdown:
	//  1. "create" simulates creating a test job and expects to be called once.
	//  2. "is-complete" simulates a long-running job that remains in a "running" state, to test cancellation handling.
	//  3. "delete" ensures that the job is deleted even if it hasn't completed, triggered by the workflow cancellation.
	s.env.OnActivity(NoopActivity, mock.Anything, "create").Times(1).Return("test-job", nil)
	s.env.OnActivity(NoopActivity, mock.Anything, "is-complete").Return(nil, fmt.Errorf("status: running"))
	s.env.OnActivity(NoopActivity, mock.Anything, "delete").Times(1).Return("ok", nil)

	// Schedule a workflow cancellation 5 seconds into the test to simulate an external cancellation event.
	s.env.RegisterDelayedCallback(func() { s.env.CancelWorkflow() }, time.Second*5)

	// Execute the workflow, encapsulating the job creation, monitoring, and cleanup within a single transactional context.
	s.env.ExecuteWorkflow(func(ctx workflow.Context) (ret any, err error) {

		logger := workflow.GetLogger(ctx)

		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Second,
			StartToCloseTimeout:    time.Second,
		})

		// Step 1: Create the job and defer its deletion to ensure cleanup occurs regardless of the job completion state.
		var res string
		if err := workflow.ExecuteActivity(ctx, NoopActivity, "create").Get(ctx, &res); err != nil {
			return nil, err
		}
		defer func() {
			if temp.IsCanceledError(err) {
				// Important! Run the clean-up logic using the disconnected context.
				ctx, _ = workflow.NewDisconnectedContext(ctx)
			}
			_ = workflow.ExecuteActivity(ctx, NoopActivity, "delete").Get(ctx, nil)
			logger.Info("deleted....")
		}()

		// Step 2: Monitor the job for completion, with retries to handle transient failures or delays in job processing.
		sensorCtx := workflow.WithRetryPolicy(ctx, temp.RetryPolicy{
			InitialInterval:    time.Second * 2,
			BackoffCoefficient: 1,
			MaximumAttempts:    10,
		})
		if err := workflow.ExecuteActivity(sensorCtx, NoopActivity, "is-complete").Get(sensorCtx, &res); err != nil {
			return nil, err
		}
		return res, nil
	})
	require := s.Require()
	require.True(s.env.IsWorkflowCompleted())
	require.EqualError(s.env.GetWorkflowError(), "workflow execution error (type: func2, workflowID: default-test-workflow-id, runID: default-test-run-id): canceled")
}

// Test_ExecuteWorkflow_No_RetryPolicy verifies the behavior of a child workflow execution without a retry policy.
// It ensures that the Noop workflow is executed exactly once, as no retries should be attempted in the absence of a retry policy.
func (s *TemporalTestSuite) Test_ExecuteWorkflow_No_RetryPolicy() {
	// Mock the Noop workflow to fail and report its attempt number.
	// Expect it to be called exactly once, as there's no retry policy to trigger any retries.
	s.env.OnWorkflow(Noop, mock.Anything, mock.Anything).Times(1).Return(func(ctx workflow.Context, request any) (any, error) {
		info := workflow.GetInfo(ctx)
		return nil, fmt.Errorf("error, attempt=%d", info.Attempt)
	})

	// Execute the workflow which includes invoking the Noop child workflow without specifying a retry policy.
	s.env.ExecuteWorkflow(func(ctx workflow.Context) (any, error) {
		var res any
		err := workflow.ExecuteChildWorkflow(ctx, Noop, "test").Get(ctx, &res)
		return res, err
	})
	require := s.Require()
	require.True(s.env.IsWorkflowCompleted())
	// Check that the error message indicates a single attempt (attempt 0).
	require.EqualError(s.env.GetWorkflowError(), "workflow execution error (type: func2, workflowID: default-test-workflow-id, runID: default-test-run-id): child workflow execution error (type: Noop, workflowID: default-test-run-id_1, runID: default-test-run-id_1_RunID, initiatedEventID: 0, startedEventID: 0): error, attempt=1")
}

// Test_ExecuteWorkflow_RetryPolicy_MaxAttempts_1 verifies the behavior of a child workflow execution with a retry
// policy configured to allow only one retry.
// It tests that the Noop workflow is executed twice - once for the initial attempt and once for the retry, given the
// retry policy's maximum attempts is set to 1.
func (s *TemporalTestSuite) Test_ExecuteWorkflow_RetryPolicy_MaxAttempts_1() {
	// Mock the Noop workflow to fail and report its attempt number.
	// Expect it to be called twice due to the retry policy with a maximum of 1 attempt.
	s.env.OnWorkflow(Noop, mock.Anything, mock.Anything).Times(1).Return(func(ctx workflow.Context, request any) (any, error) {
		info := workflow.GetInfo(ctx)
		return nil, fmt.Errorf("error, attempt=%d", info.Attempt)
	})

	// Execute the workflow which includes invoking the Noop child workflow with a retry policy of 1 attempt.
	s.env.ExecuteWorkflow(func(ctx workflow.Context) (any, error) {
		ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{})
		var res any
		err := workflow.ExecuteChildWorkflow(ctx, Noop, "test").Get(ctx, &res)
		return res, err
	})
	s.env.SetDataConverter(temporal.DataConverter{})
	require := s.Require()
	require.True(s.env.IsWorkflowCompleted())
	// Verify the error message indicates two attempts (initial attempt is 0, followed by a retry attempt 1).
	require.EqualError(s.env.GetWorkflowError(), "workflow execution error (type: func2, workflowID: default-test-workflow-id, runID: default-test-run-id): child workflow execution error (type: Noop, workflowID: default-test-run-id_1, runID: default-test-run-id_1_RunID, initiatedEventID: 0, startedEventID: 0): error, attempt=1")
}

func Noop(_ workflow.Context, request any) (any, error) {
	return request, nil
}

func NoopActivity(_ context.Context, request any) (any, error) {
	return request, nil
}
