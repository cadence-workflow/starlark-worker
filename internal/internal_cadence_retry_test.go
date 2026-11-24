package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/testsuite"
	cad "go.uber.org/cadence/workflow"
)

// Compile-time guards: RetryPolicy must keep the same fields, order and types as the Cadence client's
// RetryPolicy so that direct conversion works in both directions.
var (
	_ = cad.RetryPolicy(RetryPolicy{})
	_ = RetryPolicy(cad.RetryPolicy{})
)

func fullRetryPolicy() RetryPolicy {
	return RetryPolicy{
		InitialInterval:          time.Second,
		BackoffCoefficient:       2.5,
		JitterCoefficient:        0.15,
		MaximumInterval:          time.Minute,
		ExpirationInterval:       time.Hour,
		MaximumAttempts:          7,
		NonRetriableErrorReasons: []string{"bad-input"},
	}
}

func runInCadenceWorkflow(t *testing.T, fn func(ctx cad.Context)) {
	s := &testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(func(ctx cad.Context) error {
		fn(ctx)
		return nil
	})
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}

func TestCadenceRetryPolicyJitterPassthrough(t *testing.T) {
	w := &CadenceWorkflow{}

	t.Run("WithActivityOptions", func(t *testing.T) {
		runInCadenceWorkflow(t, func(ctx cad.Context) {
			p := fullRetryPolicy()
			ctx = w.WithActivityOptions(ctx, ActivityOptions{
				StartToCloseTimeout: time.Minute,
				RetryPolicy:         &p,
			}).(cad.Context)
			ao := cad.GetActivityOptions(ctx)
			require.NotNil(t, ao)
			require.NotNil(t, ao.RetryPolicy)
			require.Equal(t, 0.15, ao.RetryPolicy.JitterCoefficient)
			require.Equal(t, 2.5, ao.RetryPolicy.BackoffCoefficient)
			require.Equal(t, time.Minute, ao.RetryPolicy.MaximumInterval)
		})
	})

	t.Run("WithRetryPolicy", func(t *testing.T) {
		runInCadenceWorkflow(t, func(ctx cad.Context) {
			ctx = w.WithRetryPolicy(ctx, fullRetryPolicy()).(cad.Context)
			ao := cad.GetActivityOptions(ctx)
			require.NotNil(t, ao)
			require.NotNil(t, ao.RetryPolicy)
			require.Equal(t, 0.15, ao.RetryPolicy.JitterCoefficient)
			require.Equal(t, int32(7), ao.RetryPolicy.MaximumAttempts)
		})
	})

	t.Run("WithChildOptions", func(t *testing.T) {
		s := &testsuite.WorkflowTestSuite{}
		env := s.NewTestWorkflowEnvironment()
		child := func(ctx cad.Context) (string, error) { return "ok", nil }
		env.RegisterWorkflowWithOptions(child, cad.RegisterOptions{Name: "child"})
		env.ExecuteWorkflow(func(ctx cad.Context) (string, error) {
			p := fullRetryPolicy()
			ctx = w.WithChildOptions(ctx, ChildWorkflowOptions{
				ExecutionStartToCloseTimeout: time.Minute,
				TaskStartToCloseTimeout:      10 * time.Second,
				RetryPolicy:                  &p,
			}).(cad.Context)
			var res string
			err := cad.ExecuteChildWorkflow(ctx, "child").Get(ctx, &res)
			return res, err
		})
		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		var res string
		require.NoError(t, env.GetWorkflowResult(&res))
		require.Equal(t, "ok", res)
	})
}
