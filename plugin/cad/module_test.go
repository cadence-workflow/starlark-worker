package cad

import (
	"github.com/cadence-workflow/starlark-worker/cadstar"
	"github.com/stretchr/testify/suite"
	"testing"
)

type Test struct {
	suite.Suite
	cadstar.StarTestSuite
	env *cadstar.StarTestEnvironment
}

func TestSuite(t *testing.T) { suite.Run(t, new(Test)) }

func (r *Test) SetupTest() {
	r.env = r.NewEnvironment(r.T(), &cadstar.StarTestEnvironmentParams{
		RootDirectory: "testdata",
		Plugins:       []cadstar.IPlugin{Plugin},
	})
}

func (r *Test) TearDownTest() { r.env.AssertExpectations(r.T()) }

func (r *Test) TestExecutionID() {
	r.env.ExecuteFunction("/test.star", "test_execution_id", nil, nil, nil)
	require := r.Require()
	var res string
	require.NoError(r.env.GetResult(&res))
	require.Equal("default-test-workflow-id", res)
}

func (r *Test) TestExecutionRunID() {
	r.env.ExecuteFunction("/test.star", "test_execution_run_id", nil, nil, nil)
	require := r.Require()
	var res string
	require.NoError(r.env.GetResult(&res))
	require.Equal("default-test-run-id", res)
}
