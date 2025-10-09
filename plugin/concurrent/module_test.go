package concurrent

import (
	"testing"
	"time"

	"go.starlark.net/starlark"

	ptime "github.com/cadence-workflow/starlark-worker/plugin/time"
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/stretchr/testify/require"
)

type env interface {
	ExecuteFunction(filePath, function string, args starlark.Tuple, kw []starlark.Tuple, env *starlark.Dict)
	GetResult(ptr any) error
	AssertExpectations(t *testing.T)
}

type envBuilder func(t *testing.T) env

func testBatchRunSimple(t *testing.T, envBuilder envBuilder) {
	testEnv := envBuilder(t)
	t.Cleanup(func() {
		testEnv.AssertExpectations(t)
	})
	testEnv.ExecuteFunction("/test.star", "test_batch_run_simple", nil, nil, nil)
	var res *starlark.List
	require.NoError(t, testEnv.GetResult(&res))
	require.Equal(t, starlark.NewList([]starlark.Value{starlark.MakeInt(0), starlark.MakeInt(5), starlark.MakeInt(10)}), res, "expected 3 results")
}

func testBatchRunWithConcurrencyWithConcurrencyLimitLessThanTaskListLength(t *testing.T, envBuilder envBuilder) {
	testEnv := envBuilder(t)
	t.Cleanup(func() {
		testEnv.AssertExpectations(t)
	})
	testEnv.ExecuteFunction("/test.star", "test_batch_run_with_max_concurrency_less_than_task_list_length", nil, nil, nil)
	var res int64
	require.NoError(t, testEnv.GetResult(&res))
	require.LessOrEqual(t, res, int64(20*time.Second), "expected entire workflow to take less than 20 seconds")
}

func testBatchRunWithConcurrencyWithConcurrencyLimitGreaterThanTaskListLength(t *testing.T, envBuilder envBuilder) {
	testEnv := envBuilder(t)
	t.Cleanup(func() {
		testEnv.AssertExpectations(t)
	})
	testEnv.ExecuteFunction("/test.star", "test_batch_run_with_max_concurrency_greater_than_task_list_length", nil, nil, nil)
	var res int64
	require.NoError(t, testEnv.GetResult(&res))
	require.LessOrEqual(t, res, int64(10*time.Second), "expected entire workflow to take less than 20 seconds")
}

func testBatchRunWithConcurrencyWithConcurrencyLimitEqualToTaskListLength(t *testing.T, envBuilder envBuilder) {
	testEnv := envBuilder(t)
	t.Cleanup(func() {
		testEnv.AssertExpectations(t)
	})
	testEnv.ExecuteFunction("/test.star", "test_batch_run_with_max_concurrency_equal_to_task_list_length", nil, nil, nil)
	var res int64
	require.NoError(t, testEnv.GetResult(&res))
	require.LessOrEqual(t, res, int64(10*time.Second), "expected entire workflow to take less than 20 seconds")
}

func testBatchFutureIsReadyHappyPath(t *testing.T, envBuilder envBuilder) {
	testEnv := envBuilder(t)
	t.Cleanup(func() {
		testEnv.AssertExpectations(t)
	})
	testEnv.ExecuteFunction("/test.star", "test_batch_future_is_ready_happy_path", nil, nil, nil)
	var res starlark.Bool
	require.NoError(t, testEnv.GetResult(&res))
	require.Equal(t, starlark.Bool(true), res, "expected is ready to be true")
}

func testBatchFutureIsReadyBeforeGet(t *testing.T, envBuilder envBuilder) {
	testEnv := envBuilder(t)
	t.Cleanup(func() {
		testEnv.AssertExpectations(t)
	})
	testEnv.ExecuteFunction("/test.star", "test_batch_future_is_ready_before_get", nil, nil, nil)
	var res starlark.Bool
	require.NoError(t, testEnv.GetResult(&res))
	require.Equal(t, starlark.Bool(false), res, "expected is ready to be false")
}

func testBatchFutureGetFutures(t *testing.T, envBuilder envBuilder) {
	testEnv := envBuilder(t)
	t.Cleanup(func() {
		testEnv.AssertExpectations(t)
	})
	testEnv.ExecuteFunction("/test.star", "test_batch_future_get_futures", nil, nil, nil)
	var res *starlark.List
	require.NoError(t, testEnv.GetResult(&res))
	require.Equal(t, starlark.NewList([]starlark.Value{starlark.MakeInt(0), starlark.MakeInt(10)}), res, "expected 2 results")
}

func runTestSuite(t *testing.T, label string, build envBuilder) {
	for _, tc := range []struct {
		name     string
		function string
		testFunc func(t *testing.T, envBuilder envBuilder)
	}{
		{
			name:     "BatchRunSimple",
			testFunc: testBatchRunSimple,
		},
		{
			name:     "BatchRunWithConcurrencyWithConcurrencyLimitLessThanTaskListLength",
			testFunc: testBatchRunWithConcurrencyWithConcurrencyLimitLessThanTaskListLength,
		},
		{
			name:     "BatchRunWithConcurrencyWithConcurrencyLimitGreaterThanTaskListLength",
			testFunc: testBatchRunWithConcurrencyWithConcurrencyLimitGreaterThanTaskListLength,
		},
		{
			name:     "BatchRunWithConcurrencyWithConcurrencyLimitEqualToTaskListLength",
			testFunc: testBatchRunWithConcurrencyWithConcurrencyLimitEqualToTaskListLength,
		},
		{
			name:     "BatchFutureIsReadyHappyPath",
			testFunc: testBatchFutureIsReadyHappyPath,
		},
		{
			name:     "BatchFutureIsReadyBeforeGet",
			testFunc: testBatchFutureIsReadyBeforeGet,
		},
		{
			name:     "BatchFutureGetFutures",
			testFunc: testBatchFutureGetFutures,
		},
	} {
		t.Run(label+"_"+tc.name, func(t *testing.T) {
			tc.testFunc(t, build)
		})
	}
}

func TestCadenceRunner(t *testing.T) {
	runTestSuite(t, "Cadence", func(t *testing.T) env {
		suite := &service.StarCadTestSuite{}
		return suite.NewCadEnvironment(t, &service.StarCadTestEnvironmentParams{
			RootDirectory: "testdata",
			Plugins: map[string]service.IPlugin{
				Plugin.ID():       Plugin,
				ptime.Plugin.ID(): ptime.Plugin,
			},
		})
	})
}

func TestTemporalRunner(t *testing.T) {
	runTestSuite(t, "Temporal", func(t *testing.T) env {
		suite := &service.StarTempTestSuite{}
		return suite.NewTempEnvironment(t, &service.StarTempTestEnvironmentParams{
			RootDirectory: "testdata",
			Plugins: map[string]service.IPlugin{
				Plugin.ID():       Plugin,
				ptime.Plugin.ID(): ptime.Plugin,
			},
		})
	})
}
