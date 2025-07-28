package test

import (
	"errors"
	"fmt"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/cadence-workflow/starlark-worker/plugin"
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/stretchr/testify/suite"
	"go.starlark.net/starlark"
	temporalsdk "go.temporal.io/sdk/temporal"
	"io/fs"
	"log"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

type TempSuite struct {
	suite.Suite
	service.StarTempTestSuite
	httpHandler ext.HTTPTestHandler
	server      *httptest.Server
	env         *service.StarTempTestEnvironment
}

func TestTemp(t *testing.T) { suite.Run(t, new(TempSuite)) }

func (r *TempSuite) SetupSuite() {
	r.httpHandler = ext.NewHTTPTestHandler(r.T())
	r.server = httptest.NewServer(r.httpHandler)
}

func (r *TempSuite) SetupTest() {
	r.env = r.NewTempEnvironment(r.T(), &service.StarTempTestEnvironmentParams{
		RootDirectory: ".",
		Plugins:       plugin.Registry,
	})
}

func (r *TempSuite) TearDownTest() {
	r.env.AssertExpectations(r.T())
}

func (r *TempSuite) TearDownSuite() {
	r.server.Close()
}

func (r *TempSuite) TestAll() {
	var testFiles []string
	err := filepath.WalkDir("testdata", func(entryPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.star") {
			testFiles = append(testFiles, entryPath)
		}
		return nil
	})
	require := r.Require()
	require.NoError(err)
	require.True(len(testFiles) > 0, "no test files found")

	for _, file := range testFiles {
		r.runTestFile(file)
	}
}

func (r *TempSuite) TestAtExit() {
	// clean up test server resources if any
	resources := r.httpHandler.GetResources()
	for k := range resources {
		delete(resources, k)
	}

	// run the test
	r.runTestFunction("./testdata/atexit_test.star", "injected_error_test", nil, func() {
		err := r.env.GetResult(nil)
		require := r.Require()
		require.Error(err)

		var appError *temporalsdk.ApplicationError
		require.True(errors.As(err, &appError))
		require.Equal("fail: injected error", appError.Message())
		require.Equal("TemporalCustomError", appError.Type())
	})

	// make sure the test run did not leak any resources on the test server
	r.Require().Equal(0, len(resources), "Test server contains leaked resources:\n%v", resources)
}

func (r *TempSuite) runTestFile(filePath string) {

	testFunctions, err := r.env.GetTestFunctions(filePath)
	require := r.Require()
	require.NoError(err)
	require.True(len(testFunctions) > 0, "no test functions found in %s", filePath)

	for _, fn := range testFunctions {
		r.runTestFunction(filePath, fn, nil, func() {
			var res any
			if err := r.env.GetResult(&res); err != nil {
				details := err.Error()
				var customErr *temporalsdk.ApplicationError
				if errors.As(err, &customErr) && customErr.HasDetails() {
					var d []byte
					r.Require().NoError(customErr.Details(&d))
					details = fmt.Sprintf("%s: %s", customErr.Message(), d)
				}
				r.Require().Fail(details)
			}
		})
	}
}

func (r *TempSuite) runTestFunction(filePath string, fn string, environ map[string]string, assert func()) {

	r.Run(fmt.Sprintf("%s//%s", filePath, fn), func() {

		r.SetupTest()
		defer r.TearDownTest()

		defaultEnviron := map[string]string{
			"TEST_SERVER_URL": r.server.URL,
		}
		environDict := starlark.NewDict(len(defaultEnviron) + len(environ))
		for _, e := range []map[string]string{defaultEnviron, environ} {
			for k, v := range e {
				log.Printf("[t] environ: %s=%s", k, v)
				r.Require().NoError(environDict.SetKey(starlark.String(k), starlark.String(v)))
			}
		}
		r.env.ExecuteFunction(filePath, fn, nil, nil, environDict)
		assert()
	})
}
