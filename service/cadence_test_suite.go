package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cadence-workflow/starlark-worker/backend"
	"github.com/cadence-workflow/starlark-worker/cadence"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/cadence-workflow/starlark-worker/star"
	"github.com/cadence-workflow/starlark-worker/test/types"
	"github.com/cadence-workflow/starlark-worker/worker"
	"github.com/cadence-workflow/starlark-worker/workflow"
	"github.com/stretchr/testify/require"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	cadactivity "go.uber.org/cadence/activity"
	"go.uber.org/cadence/testsuite"
	cadworker "go.uber.org/cadence/worker"
	cadworkflow "go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"strings"
	"testing"
)

type cadRegistry struct {
	env *testsuite.TestWorkflowEnvironment
}

func (r cadRegistry) RegisterActivity(a interface{}) {
	r.env.RegisterActivity(a)
}
func (r cadRegistry) RegisterActivityWithOptions(a interface{}, opt worker.RegisterActivityOptions) {
	r.env.RegisterActivityWithOptions(a, cadactivity.RegisterOptions{
		Name:                          opt.Name,
		EnableShortName:               opt.EnableShortName,
		DisableAlreadyRegisteredCheck: opt.DisableAlreadyRegisteredCheck,
		EnableAutoHeartbeat:           opt.EnableAutoHeartbeat,
	})
}
func (r cadRegistry) RegisterWorkflow(w interface{}) {
	r.env.RegisterWorkflow(cadence.UpdateWorkflowFunctionContextArgument(w))
}
func (r cadRegistry) RegisterWorkflowWithOptions(w interface{}, options worker.RegisterWorkflowOptions) {
	r.env.RegisterWorkflowWithOptions(cadence.UpdateWorkflowFunctionContextArgument(w), cadworkflow.RegisterOptions{
		Name:                          options.Name,
		EnableShortName:               options.EnableShortName,
		DisableAlreadyRegisteredCheck: options.DisableAlreadyRegisteredCheck,
	})
}

// StarCadTestEnvironment is a test environment for the Starlark functions.
type StarCadTestEnvironment struct {
	env     *testsuite.TestWorkflowEnvironment
	service *Service
	tar     []byte
	fs      star.FS
}

// GetTestWorkflowEnvironment returns the underlying testsuite.TestWorkflowEnvironment instance.
func (r *StarCadTestEnvironment) GetTestWorkflowEnvironment() *testsuite.TestWorkflowEnvironment {
	return r.env
}

func (r *StarCadTestEnvironment) AssertExpectations(t *testing.T) {
	r.env.AssertExpectations(t)
}

func (r *StarCadTestEnvironment) ExecuteFunction(
	filePath string,
	fn string,
	args starlark.Tuple,
	kw []starlark.Tuple,
	environ *starlark.Dict,
) {
	env := r.env
	env.ExecuteWorkflow(cadence.UpdateWorkflowFunctionContextArgument(r.service.Run), r.tar, filePath, fn, args, kw, environ)
}

func (r *StarCadTestEnvironment) GetResult(valuePtr any) error {
	env := r.env
	if !env.IsWorkflowCompleted() {
		return fmt.Errorf("workflow is not completed")
	}
	err := env.GetWorkflowError()
	if err != nil {
		return err
	}
	return env.GetWorkflowResult(valuePtr)
}

// GetTestFunctions returns the list of test functions in the given file. Test functions are the functions that start with "test_".
func (r *StarCadTestEnvironment) GetTestFunctions(filePath string) ([]string, error) {
	globals, err := r.getGlobals(filePath)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, binding := range globals {
		bn := binding.First.Name
		if strings.HasPrefix(bn, "test_") {
			res = append(res, bn)
		}
	}
	return res, nil
}

// getGlobals is a helper function that returns the global bindings in the given Starlark file.
func (r *StarCadTestEnvironment) getGlobals(filePath string) ([]*resolve.Binding, error) {
	src, err := r.fs.Read(filePath)
	if err != nil {
		return nil, err
	}
	code, err := star.FileOptions.Parse(filePath, src, 0)
	if err != nil {
		return nil, err
	}
	if err := resolve.File(code, func(s string) bool { return true }, starlark.Universe.Has); err != nil {
		return nil, err
	}
	return code.Module.(*resolve.Module).Globals, nil
}

type StarCadTestSuite struct {
	testsuite.WorkflowTestSuite
	tarCache map[string][]byte
	fsCache  map[string]star.FS
}

type StarCadTestEnvironmentParams struct {
	RootDirectory  string
	Plugins        map[string]IPlugin
	ServiceBackend backend.Backend
}

// NewCadEnvironment creates a new StarTestEnvironment - test environment for the Starlark functions.
func (r *StarCadTestSuite) NewCadEnvironment(t *testing.T, p *StarCadTestEnvironmentParams) *StarCadTestEnvironment {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))

	ctx := context.WithValue(context.Background(), workflow.BackendContextKey, cadence.GetBackend().RegisterWorkflow())
	env := r.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(cadworker.Options{
		Logger:                    logger,
		DataConverter:             &cadence.DataConverter{},
		BackgroundActivityContext: ctx,
	})

	service, serviceErr := NewService(p.Plugins, "test", CadenceBackend)
	require.NoError(t, serviceErr)
	service.Register(cadRegistry{env: env})

	if r.tarCache == nil {
		r.tarCache = map[string][]byte{}
		r.fsCache = map[string]star.FS{}
	}

	var tar []byte
	var fs star.FS
	var err error
	if cached, found := r.tarCache[p.RootDirectory]; found {
		fmt.Println("tarCache hit:", p.RootDirectory)
		tar = cached
		fs = r.fsCache[p.RootDirectory]
	} else {
		var bb bytes.Buffer
		require.NoError(t, ext.DirToTar(p.RootDirectory, &bb))
		tar = bb.Bytes()
		fs, err = star.NewTarFS(tar)
		require.NoError(t, err)
		r.tarCache[p.RootDirectory] = tar
		r.fsCache[p.RootDirectory] = fs
	}

	return &StarCadTestEnvironment{env: env, service: service, tar: tar, fs: fs}
}

func (r *StarCadTestSuite) buildTar(t *testing.T, p string) ([]byte, bool) {
	if r.tarCache == nil {
		r.tarCache = make(map[string][]byte)
	}
	if cached, found := r.tarCache[p]; found {
		fmt.Println("tarCache hit:", p)
		return cached, true
	}
	var bb bytes.Buffer
	require.NoError(t, ext.DirToTar(p, &bb))
	tar := bb.Bytes()
	r.tarCache[p] = tar
	return tar, false
}

type starCadTestActivitySuite struct {
	testsuite.WorkflowTestSuite
	env *testsuite.TestActivityEnvironment
}

func NewCadTestActivitySuite() types.StarTestActivitySuite {
	s := starCadTestActivitySuite{}
	s.env = s.NewTestActivityEnvironment()
	ctx := context.WithValue(context.Background(), workflow.BackendContextKey, cadence.GetBackend().RegisterWorkflow())
	s.env.SetWorkerOptions(cadworker.Options{
		BackgroundActivityContext: ctx,
	})
	return s
}

func (r starCadTestActivitySuite) RegisterActivity(a interface{}) {
	if r.env == nil {
		r.env = r.NewTestActivityEnvironment()
	}
	r.env.RegisterActivity(a)
}

func (r starCadTestActivitySuite) ExecuteActivity(a interface{}, opts interface{}) (types.EncodedValue, error) {
	return r.env.ExecuteActivity(a, opts)
}
