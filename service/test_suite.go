package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cadence-workflow/starlark-worker/cadence"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/cadence-workflow/starlark-worker/star"
	"github.com/cadence-workflow/starlark-worker/temporal"
	"github.com/cadence-workflow/starlark-worker/test/types"
	"github.com/cadence-workflow/starlark-worker/worker"
	"github.com/cadence-workflow/starlark-worker/workflow"
	"github.com/stretchr/testify/require"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	tmpactivity "go.temporal.io/sdk/activity"
	temptestsuite "go.temporal.io/sdk/testsuite"
	tmpworker "go.temporal.io/sdk/worker"
	tmpworkflow "go.temporal.io/sdk/workflow"
	cadactivity "go.uber.org/cadence/activity"
	"go.uber.org/cadence/testsuite"
	cadworker "go.uber.org/cadence/worker"
	cadworkflow "go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"strings"
	"testing"
)

type TestSuite struct {
	Cadence  StarCadTestSuite
	Temporal StarTempTestSuite
}

type TestEnvironment struct {
	Cadence  *StarCadTestEnvironment
	Temporal *StarTempTestEnvironment
}

type TestEnvironmentParams struct {
	RootDirectory string
	Plugins       map[string]IPlugin
}

func (s *TestSuite) NewTestEnvironment(t *testing.T, p *TestEnvironmentParams) *TestEnvironment {
	cadEnv := s.Cadence.NewCadEnvironment(t, &StarCadTestEnvironmentParams{
		RootDirectory: p.RootDirectory,
		Plugins:       p.Plugins,
	})
	tempEnv := s.Temporal.NewTempEnvironment(t, &StarTempTestEnvironmentParams{
		RootDirectory: p.RootDirectory,
		Plugins:       p.Plugins,
	})
	return &TestEnvironment{
		Cadence:  cadEnv,
		Temporal: tempEnv,
	}
}

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
	wf, _ := cadence.UpdateWorkflowFunctionContextArgument(w)
	r.env.RegisterWorkflow(wf)
}
func (r cadRegistry) RegisterWorkflowWithOptions(w interface{}, options worker.RegisterWorkflowOptions) {
	wf, _ := cadence.UpdateWorkflowFunctionContextArgument(w)
	r.env.RegisterWorkflowWithOptions(wf, cadworkflow.RegisterOptions{
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
	wf, _ := cadence.UpdateWorkflowFunctionContextArgument(r.service.Run)
	env.ExecuteWorkflow(wf, r.tar, filePath, fn, args, kw, environ)
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
	RootDirectory string
	Plugins       map[string]IPlugin
}

// NewCadEnvironment creates a new StarTestEnvironment - test environment for the Starlark functions.
func (r *StarCadTestSuite) NewCadEnvironment(t *testing.T, p *StarCadTestEnvironmentParams) *StarCadTestEnvironment {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))

	ctx := context.WithValue(context.Background(), workflow.BackendContextKey, cadence.NewWorkflow())
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
	ctx := context.WithValue(context.Background(), workflow.BackendContextKey, cadence.NewWorkflow())
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

type tempRegistry struct {
	env *temptestsuite.TestWorkflowEnvironment
}

func (r tempRegistry) RegisterActivity(a interface{}) {
	r.env.RegisterActivity(a)
}
func (r tempRegistry) RegisterActivityWithOptions(a interface{}, opt worker.RegisterActivityOptions) {
	r.env.RegisterActivityWithOptions(a, tmpactivity.RegisterOptions{
		Name:                          opt.Name,
		DisableAlreadyRegisteredCheck: opt.DisableAlreadyRegisteredCheck,
		SkipInvalidStructFunctions:    opt.SkipInvalidStructFunctions,
	})
}
func (r tempRegistry) RegisterWorkflow(w interface{}) {
	wf, _ := temporal.UpdateWorkflowFunctionContextArgument(w)
	r.env.RegisterWorkflow(wf)
}
func (r tempRegistry) RegisterWorkflowWithOptions(w interface{}, options worker.RegisterWorkflowOptions) {
	wf, _ := temporal.UpdateWorkflowFunctionContextArgument(w)
	r.env.RegisterWorkflowWithOptions(wf, tmpworkflow.RegisterOptions{
		Name:                          options.Name,
		DisableAlreadyRegisteredCheck: options.DisableAlreadyRegisteredCheck,
	})
}

type StarTempTestEnvironment struct {
	env     *temptestsuite.TestWorkflowEnvironment
	service *Service
	tar     []byte
	fs      star.FS
}

func (r *StarTempTestEnvironment) GetTestWorkflowEnvironment() *temptestsuite.TestWorkflowEnvironment {
	return r.env
}

func (r *StarTempTestEnvironment) AssertExpectations(t *testing.T) {
	r.env.AssertExpectations(t)
}

func (r *StarTempTestEnvironment) ExecuteFunction(
	filePath string,
	fn string,
	args starlark.Tuple,
	kw []starlark.Tuple,
	environ *starlark.Dict,
) {
	wf, _ := temporal.UpdateWorkflowFunctionContextArgument(r.service.Run)
	r.env.ExecuteWorkflow(wf, r.tar, filePath, fn, args, kw, environ)
}

func (r *StarTempTestEnvironment) GetResult(valuePtr any) error {
	if !r.env.IsWorkflowCompleted() {
		return fmt.Errorf("workflow is not completed")
	}
	if err := r.env.GetWorkflowError(); err != nil {
		return err
	}
	return r.env.GetWorkflowResult(valuePtr)
}

func (r *StarTempTestEnvironment) GetTestFunctions(filePath string) ([]string, error) {
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

func (r *StarTempTestEnvironment) getGlobals(filePath string) ([]*resolve.Binding, error) {
	src, err := r.fs.Read(filePath)
	if err != nil {
		return nil, err
	}
	code, err := star.FileOptions.Parse(filePath, src, 0)
	if err != nil {
		return nil, err
	}
	if err := resolve.File(code, func(string) bool { return true }, starlark.Universe.Has); err != nil {
		return nil, err
	}
	return code.Module.(*resolve.Module).Globals, nil
}

type StarTempTestSuite struct {
	temptestsuite.WorkflowTestSuite
	tarCache map[string][]byte
	fsCache  map[string]star.FS
}

type StarTempTestEnvironmentParams struct {
	RootDirectory string
	Plugins       map[string]IPlugin
}

func (r *StarTempTestSuite) NewTempEnvironment(t *testing.T, p *StarTempTestEnvironmentParams) *StarTempTestEnvironment {
	env := r.NewTestWorkflowEnvironment()
	ctx := context.WithValue(context.Background(), workflow.BackendContextKey, temporal.NewWorkflow())
	env.SetWorkerOptions(tmpworker.Options{
		BackgroundActivityContext: ctx,
	})
	env.SetContextPropagators([]tmpworkflow.ContextPropagator{&temporal.HeadersContextPropagator{}})
	env.SetDataConverter(temporal.DataConverter{})
	service, serviceErr := NewService(p.Plugins, "test", TemporalBackend)
	require.NoError(t, serviceErr)

	service.Register(tempRegistry{env: env})

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

	return &StarTempTestEnvironment{
		env:     env,
		service: service,
		tar:     tar,
		fs:      fs,
	}
}

func (r *StarTempTestSuite) buildTar(t *testing.T, dir string) ([]byte, bool) {
	if r.tarCache == nil {
		r.tarCache = map[string][]byte{}
	}
	if tar, ok := r.tarCache[dir]; ok {
		fmt.Println("tarCache hit:", dir)
		return tar, true
	}
	var buf bytes.Buffer
	require.NoError(t, ext.DirToTar(dir, &buf))
	tar := buf.Bytes()
	r.tarCache[dir] = tar
	return tar, false
}

type starTempTestActivitySuite struct {
	temptestsuite.WorkflowTestSuite
	env *temptestsuite.TestActivityEnvironment
}

func NewTempTestActivitySuite() types.StarTestActivitySuite {
	s := starTempTestActivitySuite{}
	s.env = s.NewTestActivityEnvironment()
	ctx := context.WithValue(context.Background(), workflow.BackendContextKey, temporal.NewWorkflow())
	s.env.SetWorkerOptions(tmpworker.Options{
		BackgroundActivityContext: ctx,
	})
	return s
}

func (r starTempTestActivitySuite) RegisterActivity(a interface{}) {
	r.env.RegisterActivity(a)
}

func (r starTempTestActivitySuite) ExecuteActivity(a interface{}, opts interface{}) (types.EncodedValue, error) {
	return r.env.ExecuteActivity(a, opts)
}
