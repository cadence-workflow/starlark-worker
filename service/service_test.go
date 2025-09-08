package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.starlark.net/starlark"
	"go.uber.org/zap"

	"github.com/cadence-workflow/starlark-worker/activity"
	"github.com/cadence-workflow/starlark-worker/star"
	"github.com/cadence-workflow/starlark-worker/worker"
	"github.com/cadence-workflow/starlark-worker/workflow"
)

type TestPlugin struct{}

var _ IPlugin = (*TestPlugin)(nil)

const testPluginID = "testplugin"

func (t *TestPlugin) ID() string {
	return testPluginID
}

func (t *TestPlugin) Create(_ RunInfo) starlark.Value {
	return &TestModule{}
}

func (t *TestPlugin) Register(registry worker.Registry) {
	registry.RegisterActivityWithOptions(stringifyActivity, worker.RegisterActivityOptions{Name: "stringify_activity"})
}

func stringifyActivityWrapper(t *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	ctx := GetContext(t)
	var res starlark.String
	if err := workflow.ExecuteActivity(ctx, "stringify_activity", args).Get(ctx, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func stringifyActivity(ctx context.Context, args starlark.Tuple) (starlark.String, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("stringify_activity", zap.String("args", args.String()))
	return starlark.String(args.String()), nil
}

type TestModule struct{}

var _ starlark.HasAttrs = &TestModule{}

func (f *TestModule) String() string        { return testPluginID }
func (f *TestModule) Type() string          { return testPluginID }
func (f *TestModule) Freeze()               {}
func (f *TestModule) Truth() starlark.Bool  { return true }
func (f *TestModule) Hash() (uint32, error) { return 0, fmt.Errorf("no-hash") }
func (f *TestModule) Attr(n string) (starlark.Value, error) {
	return star.Attr(f, n, testBuiltins, properties)
}
func (f *TestModule) AttrNames() []string { return star.AttrNames(testBuiltins, properties) }

var properties = map[string]star.PropertyFactory{}

var testBuiltins = map[string]*starlark.Builtin{
	"stringify_activity": starlark.NewBuiltin("stringify_activity", stringifyActivityWrapper),
}

type CadTest struct {
	suite.Suite
	StarCadTestSuite
	env *StarCadTestEnvironment
}

func TestCadSuite(t *testing.T) { suite.Run(t, new(CadTest)) }

func (r *CadTest) SetupSuite() {}

func (r *CadTest) SetupTest() {
	tp := &TestPlugin{}
	r.env = r.NewCadEnvironment(r.T(), &StarCadTestEnvironmentParams{
		RootDirectory: "testdata",
		Plugins:       map[string]IPlugin{tp.ID(): tp},
	})
}

func (r *CadTest) TearDownTest() {
	r.env.AssertExpectations(r.T())
}

func (r *CadTest) TestCad1() {
	r.env.ExecuteFunction("/app.star", "plus", starlark.Tuple{
		starlark.MakeInt(2),
		starlark.MakeInt(3),
	}, nil, nil)

	require := r.Require()
	var res starlark.Int
	require.NoError(r.env.GetResult(&res))
	require.Equal(starlark.MakeInt(5), res)
}

func (r *CadTest) TestCadPluginFunction() {
	r.env.ExecuteFunction("/app.star", "stringify", starlark.Tuple{
		starlark.String("foo"),
		starlark.MakeInt(100),
	}, nil, nil)

	require := r.Require()
	var res starlark.String
	require.NoError(r.env.GetResult(&res))
	require.Equal(starlark.String(`("foo", 100)`), res)
}

type TempTest struct {
	suite.Suite
	StarTempTestSuite
	env *StarTempTestEnvironment
}

func TestTempSuite(t *testing.T) { suite.Run(t, new(TempTest)) }

func (r *TempTest) SetupSuite() {}

func (r *TempTest) SetupTest() {
	tp := &TestPlugin{}
	r.env = r.NewTempEnvironment(r.T(), &StarTempTestEnvironmentParams{
		RootDirectory: "testdata",
		Plugins:       map[string]IPlugin{tp.ID(): tp},
	})
}

func (r *TempTest) TearDownTest() {
	r.env.AssertExpectations(r.T())
}

func (r *TempTest) TestTemp1() {
	r.env.ExecuteFunction("/app.star", "plus", starlark.Tuple{
		starlark.MakeInt(2),
		starlark.MakeInt(3),
	}, nil, nil)

	require := r.Require()
	var res starlark.Int
	require.NoError(r.env.GetResult(&res))
	require.Equal(starlark.MakeInt(5), res)
}

func (r *TempTest) TestTempPluginFunction() {
	r.env.ExecuteFunction("/app.star", "stringify", starlark.Tuple{
		starlark.String("foo"),
		starlark.MakeInt(100),
	}, nil, nil)

	require := r.Require()
	var res starlark.String
	require.NoError(r.env.GetResult(&res))
	require.Equal(starlark.String(`("foo", 100)`), res)
}

// TestServiceActivityOptionsConfiguration tests Service creation with various ActivityOptions configurations
func TestServiceActivityOptionsConfiguration(t *testing.T) {
	plugins := map[string]IPlugin{}
	
	t.Run("NewService", func(t *testing.T) {
		service, err := NewService(plugins, "test-tasklist", TemporalBackend)
		assert.NoError(t, err)
		assert.NotNil(t, service)
		
		// Should use default options with tasklist set
		expected := DefaultActivityOptions
		expected.TaskList = "test-tasklist"
		assert.Equal(t, expected, service.ActivityOptions)
	})
	
	t.Run("NewServiceWithDefaults", func(t *testing.T) {
		service, err := NewServiceWithDefaults(plugins, "test-tasklist", TemporalBackend)
		assert.NoError(t, err)
		assert.NotNil(t, service)
		
		// Should use default options with tasklist set
		expected := DefaultActivityOptions
		expected.TaskList = "test-tasklist"
		assert.Equal(t, expected, service.ActivityOptions)
	})
	
	t.Run("NewServiceWithCustomActivityOptions", func(t *testing.T) {
		customOptions := &workflow.ActivityOptions{
			TaskList:               "custom-tasklist",
			ScheduleToStartTimeout: time.Minute * 2,
			StartToCloseTimeout:    time.Minute * 5,
			HeartbeatTimeout:       time.Second * 30,
			WaitForCancellation:    true,
			ActivityID:             "custom-activity-id",
			DisableEagerExecution:  true,
			VersioningIntent:       1,
			Summary:                "Custom activity summary",
		}
		
		service, err := NewServiceWithOptions(plugins, "test-tasklist", TemporalBackend, customOptions)
		assert.NoError(t, err)
		assert.NotNil(t, service)
		
		// Should use custom options exactly as provided
		assert.Equal(t, *customOptions, service.ActivityOptions)
	})
	
	t.Run("NewServiceWithOptionsNil", func(t *testing.T) {
		service, err := NewServiceWithOptions(plugins, "test-tasklist", TemporalBackend, nil)
		assert.NoError(t, err)
		assert.NotNil(t, service)
		
		// Should use defaults with tasklist set
		expected := DefaultActivityOptions
		expected.TaskList = "test-tasklist"
		assert.Equal(t, expected, service.ActivityOptions)
	})
	
	t.Run("NewServiceWithOptionsPartial", func(t *testing.T) {
		partialOptions := &workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute * 10,
			HeartbeatTimeout:    time.Second * 45,
		}
		
		service, err := NewServiceWithOptions(plugins, "test-tasklist", TemporalBackend, partialOptions)
		assert.NoError(t, err)
		assert.NotNil(t, service)
		
		// Should use partial options with tasklist filled in
		expected := *partialOptions
		expected.TaskList = "test-tasklist"
		assert.Equal(t, expected, service.ActivityOptions)
	})
	
	t.Run("NewServiceWithOptionsTaskListOverride", func(t *testing.T) {
		customOptions := &workflow.ActivityOptions{
			TaskList:            "options-tasklist", // This should be preserved
			StartToCloseTimeout: time.Minute * 3,
		}
		
		service, err := NewServiceWithOptions(plugins, "param-tasklist", TemporalBackend, customOptions)
		assert.NoError(t, err)
		assert.NotNil(t, service)
		
		// TaskList from options should be preserved, not overridden
		assert.Equal(t, "options-tasklist", service.ActivityOptions.TaskList)
		assert.Equal(t, time.Minute*3, service.ActivityOptions.StartToCloseTimeout)
	})
	
	t.Run("NewServiceWithOptionsEmptyTaskList", func(t *testing.T) {
		customOptions := &workflow.ActivityOptions{
			TaskList:            "", // Empty tasklist in options
			StartToCloseTimeout: time.Minute * 3,
		}
		
		service, err := NewServiceWithOptions(plugins, "param-tasklist", TemporalBackend, customOptions)
		assert.NoError(t, err)
		assert.NotNil(t, service)
		
		// Empty TaskList in options should be filled from parameter
		assert.Equal(t, "param-tasklist", service.ActivityOptions.TaskList)
		assert.Equal(t, time.Minute*3, service.ActivityOptions.StartToCloseTimeout)
	})
}
