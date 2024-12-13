package cadstar

import (
	"context"
	"github.com/stretchr/testify/suite"
	"go.starlark.net/starlark"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"testing"
)

type TestPlugin struct{}

func (t *TestPlugin) Create(_ RunInfo) starlark.StringDict {
	return starlark.StringDict{
		"stringify_activity": starlark.NewBuiltin("stringify_activity", func(th *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
			ctx := GetContext(th)
			var res starlark.String
			if err := workflow.ExecuteActivity(ctx, t.stringifyActivity, args).Get(ctx, &res); err != nil {
				return nil, err
			}
			return res, nil
		}),
	}
}

func (t *TestPlugin) Register(registry worker.Registry) {
	registry.RegisterActivity(t.stringifyActivity)
}

func (t *TestPlugin) SharedLocalStorageKeys() []string { return nil }

func (t *TestPlugin) stringifyActivity(ctx context.Context, args starlark.Tuple) (starlark.String, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("stringify_activity", zap.String("args", args.String()))
	return starlark.String(args.String()), nil
}

var _ IPlugin = (*TestPlugin)(nil)

type Test struct {
	suite.Suite
	StarTestSuite
	env *StarTestEnvironment
}

func TestSuite(t *testing.T) { suite.Run(t, new(Test)) }

func (r *Test) SetupSuite() {}

func (r *Test) SetupTest() {
	r.env = r.NewEnvironment(r.T(), &StarTestEnvironmentParams{
		RootDirectory: "testdata",
		Plugins:       []IPlugin{&TestPlugin{}},
	})
}

func (r *Test) TearDownTest() { r.env.AssertExpectations(r.T()) }

func (r *Test) Test1() {
	r.env.ExecuteFunction("/app.star", "plus", starlark.Tuple{
		starlark.MakeInt(2),
		starlark.MakeInt(3),
	}, nil, nil)

	require := r.Require()
	var res starlark.Int
	require.NoError(r.env.GetResult(&res))
	require.Equal(starlark.MakeInt(5), res)
}

func (r *Test) TestPluginFunction() {
	r.env.ExecuteFunction("/app.star", "stringify", starlark.Tuple{
		starlark.String("foo"),
		starlark.MakeInt(100),
	}, nil, nil)

	require := r.Require()
	var res starlark.String
	require.NoError(r.env.GetResult(&res))
	require.Equal(starlark.String(`("foo", 100)`), res)
}
