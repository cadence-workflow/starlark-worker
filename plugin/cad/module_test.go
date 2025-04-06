package cad

import (
	"go.starlark.net/starlark"
	"testing"

	cad "github.com/cadence-workflow/starlark-worker/internal/cadence"
	tem "github.com/cadence-workflow/starlark-worker/internal/temporal"
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name       string
	function   string
	wantResult string
}

var testCases = []testCase{
	{
		name:       "ExecutionID",
		function:   "test_execution_id",
		wantResult: "default-test-workflow-id",
	},
	{
		name:       "ExecutionRunID",
		function:   "test_execution_run_id",
		wantResult: "default-test-run-id",
	},
}

type env interface {
	ExecuteFunction(filePath, function string, args starlark.Tuple, kw []starlark.Tuple, env *starlark.Dict)
	GetResult(ptr any) error
	AssertExpectations(t *testing.T)
}

type envBuilder func(t *testing.T) env

func runTestSuite(t *testing.T, label string, build envBuilder) {

	for _, tc := range testCases {
		t.Run(label+"_"+tc.name, func(t *testing.T) {
			testEnv := build(t)
			t.Cleanup(func() {
				testEnv.AssertExpectations(t)
			})
			testEnv.ExecuteFunction("/test.star", tc.function, nil, nil, nil)

			var res string
			require.NoError(t, testEnv.GetResult(&res))
			require.Equal(t, tc.wantResult, res)
		})
	}
}
func TestCadenceRunner(t *testing.T) {
	runTestSuite(t, "Cadence", func(t *testing.T) env {
		suite := &service.StarCadTestSuite{}
		return suite.NewCadEnvironment(t, &service.StarCadTestEnvironmentParams{
			RootDirectory:  "testdata",
			Plugins:        map[string]service.IPlugin{Plugin.ID(): Plugin},
			ServiceBackend: cad.GetBackend(),
		})
	})
}

func TestTemporalRunner(t *testing.T) {
	runTestSuite(t, "Temporal", func(t *testing.T) env {
		suite := &service.StarTempTestSuite{}
		return suite.NewTempEnvironment(t, &service.StarTempTestEnvironmentParams{
			RootDirectory:  "testdata",
			Plugins:        map[string]service.IPlugin{Plugin.ID(): Plugin},
			ServiceBackend: tem.GetBackend(),
		})
	})
}
