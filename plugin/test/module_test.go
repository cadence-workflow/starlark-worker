package test

import (
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	tsuite "go.temporal.io/sdk/testsuite"
	temp "go.temporal.io/sdk/workflow"
	csuite "go.uber.org/cadence/testsuite"
	cad "go.uber.org/cadence/workflow"
	"strings"
	"testing"
)

// Define a type for workflow execution context
type workflowExecutor func(t *testing.T)

func TestTest_TableDriven(t *testing.T) {
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
					d := &starlark.Dict{}
					_, err := _true(thread, nil, starlark.Tuple{d}, nil)

					require.Error(t, err)

					require.True(t, strings.HasPrefix(err.Error(), "assert"))

					err = d.SetKey(starlark.String("foo"), starlark.String("bar"))
					require.NoError(t, err)

					_, err = _true(thread, nil, starlark.Tuple{d}, nil)
					require.NoError(t, err)
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
					d := &starlark.Dict{}
					_, err := _true(thread, nil, starlark.Tuple{d}, nil)

					require.Error(t, err)

					require.True(t, strings.HasPrefix(err.Error(), "assert"))

					err = d.SetKey(starlark.String("foo"), starlark.String("bar"))
					require.NoError(t, err)

					_, err = _true(thread, nil, starlark.Tuple{d}, nil)
					require.NoError(t, err)
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
