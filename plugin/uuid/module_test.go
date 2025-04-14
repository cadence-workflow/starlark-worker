package uuid

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	tsuite "go.temporal.io/sdk/testsuite"
	temp "go.temporal.io/sdk/workflow"
	csuite "go.uber.org/cadence/testsuite"
	cad "go.uber.org/cadence/workflow"
)

// Define a type for workflow execution context
type workflowExecutor func(t *testing.T)

func TestUUID4_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		executor workflowExecutor
	}{
		{
			name: "Cadence",
			executor: func(t *testing.T) {
				suite := &csuite.WorkflowTestSuite{}
				env := suite.NewTestWorkflowEnvironment()

				env.ExecuteWorkflow(func(ctx cad.Context) error {
					thread := &starlark.Thread{Name: "test-thread"}
					result, err := uuid4(thread, nil, nil, nil)

					require.NoError(t, err, "Expected no error from uuid4")
					uuidObj, ok := result.(*UUID)
					require.True(t, ok, "Expected result to be of type *UUID")
					require.NotEmpty(t, string(uuidObj.StringUUID), "UUID should not be empty")

					return nil
				})

				env.AssertExpectations(t)
			},
		},
		{
			name: "Temporal",
			executor: func(t *testing.T) {
				suite := &tsuite.WorkflowTestSuite{}
				env := suite.NewTestWorkflowEnvironment()

				env.ExecuteWorkflow(func(ctx temp.Context) error {
					thread := &starlark.Thread{Name: "test-thread"}
					result, err := uuid4(thread, nil, nil, nil)

					require.NoError(t, err, "Expected no error from uuid4")
					uuidObj, ok := result.(*UUID)
					require.True(t, ok, "Expected result to be of type *UUID")
					require.NotEmpty(t, string(uuidObj.StringUUID), "UUID should not be empty")

					return nil
				})

				env.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.executor)
	}
}
