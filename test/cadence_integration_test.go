package test

import (
	"errors"
	"fmt"
	"github.com/cadence-workflow/starlark-worker/cadence"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/cadence-workflow/starlark-worker/plugin"
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/stretchr/testify/suite"
	"go.starlark.net/starlark"
	cad "go.uber.org/cadence"
	"io/fs"
	"log"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

type CadSuite struct {
	suite.Suite
	service.StarCadTestSuite
	httpHandler ext.HTTPTestHandler
	server      *httptest.Server
	env         *service.StarCadTestEnvironment
}

func TestCad(t *testing.T) { suite.Run(t, new(CadSuite)) }

func (r *CadSuite) SetupSuite() {
	r.httpHandler = ext.NewHTTPTestHandler(r.T())
	r.server = httptest.NewServer(r.httpHandler)
}

func (r *CadSuite) SetupTest() {
	r.env = r.NewCadEnvironment(r.T(), &service.StarCadTestEnvironmentParams{
		RootDirectory:  ".",
		Plugins:        plugin.Registry,
		ServiceBackend: cadence.GetBackend(),
	})
}

func (r *CadSuite) TearDownTest() {
	r.env.AssertExpectations(r.T())
}

func (r *CadSuite) TearDownSuite() {
	r.server.Close()
}

func (r *CadSuite) TestAll() {
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

func (r *CadSuite) TestAtExit() {
	// clean up test server resources if any
	resources := r.httpHandler.GetResources()
	for k := range resources {
		delete(resources, k)
	}

	// run the test
	r.runTestFunction("testdata/atexit_test.star", "injected_error_test", func() {
		err := r.env.GetResult(nil)
		require := r.Require()
		require.Error(err)

		var cadenceErr *cad.CustomError
		require.True(errors.As(err, &cadenceErr))

		var details map[string]any
		require.NoError(cadenceErr.Details(&details))
		require.NotNil(details["error"])
		require.IsType("", details["error"])
		require.True(strings.Contains(details["error"].(string), "injected error"), "Unexpected error details:\n%s", details)
	})

	// make sure the test run did not leak any resources on the test server
	r.Require().Equal(0, len(resources), "Test server contains leaked resources:\n%v", resources)
}

func (r *CadSuite) runTestFile(filePath string) {

	testFunctions, err := r.env.GetTestFunctions(filePath)
	require := r.Require()
	require.NoError(err)
	require.True(len(testFunctions) > 0, "no test functions found in %s", filePath)

	for _, fn := range testFunctions {
		r.runTestFunction(filePath, fn, func() {
			var res any
			if err := r.env.GetResult(&res); err != nil {
				details := err.Error()
				var customErr *cad.CustomError
				if errors.As(err, &customErr) && customErr.HasDetails() {
					var d []byte
					r.Require().NoError(customErr.Details(&d))
					details = fmt.Sprintf("%s: %s", customErr.Reason(), d)
				}
				r.Require().Fail(details)
			}
		})
	}
}

func (r *CadSuite) runTestFunction(filePath string, fn string, assert func()) {

	r.Run(fmt.Sprintf("%s//%s", filePath, fn), func() {

		r.SetupTest()
		defer r.TearDownTest()

		environ := starlark.NewDict(1)
		r.Require().NoError(environ.SetKey(starlark.String("TEST_SERVER_URL"), starlark.String(r.server.URL)))
		log.Printf("[t] environ: %s", environ.String())

		r.env.ExecuteFunction(filePath, fn, nil, nil, environ)
		assert()
	})
}
