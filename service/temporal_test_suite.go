package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/cadence-workflow/starlark-worker/internal/backend"
	"github.com/cadence-workflow/starlark-worker/internal/encoded"
	"github.com/cadence-workflow/starlark-worker/internal/temporal"
	"github.com/cadence-workflow/starlark-worker/internal/worker"
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"github.com/cadence-workflow/starlark-worker/star"
	"github.com/cadence-workflow/starlark-worker/test/types"
	"github.com/stretchr/testify/require"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	tmpactivity "go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	tmpworker "go.temporal.io/sdk/worker"
	tmpworkflow "go.temporal.io/sdk/workflow"
	"strings"
	"testing"
)

type tempRegistry struct {
	env *testsuite.TestWorkflowEnvironment
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
	r.env.RegisterWorkflow(temporal.UpdateWorkflowFunctionContextArgument(w))
}
func (r tempRegistry) RegisterWorkflowWithOptions(w interface{}, options worker.RegisterWorkflowOptions) {
	r.env.RegisterWorkflowWithOptions(temporal.UpdateWorkflowFunctionContextArgument(w), tmpworkflow.RegisterOptions{
		Name:                          options.Name,
		DisableAlreadyRegisteredCheck: options.DisableAlreadyRegisteredCheck,
	})
}

type StarTempTestEnvironment struct {
	env     *testsuite.TestWorkflowEnvironment
	service *Service
	tar     []byte
	fs      star.FS
}

func (r *StarTempTestEnvironment) GetTestWorkflowEnvironment() *testsuite.TestWorkflowEnvironment {
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
	r.env.ExecuteWorkflow(temporal.UpdateWorkflowFunctionContextArgument(r.service.Run), r.tar, filePath, fn, args, kw, environ)
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
	testsuite.WorkflowTestSuite
	tarCache map[string][]byte
	fsCache  map[string]star.FS
}

type StarTempTestEnvironmentParams struct {
	RootDirectory  string
	Plugins        map[string]IPlugin
	DataConvertor  encoded.DataConvertor
	ServiceBackend backend.Backend
}

func (r *StarTempTestSuite) NewTempEnvironment(t *testing.T, p *StarTempTestEnvironmentParams) *StarTempTestEnvironment {
	env := r.NewTestWorkflowEnvironment()
	ctx := context.WithValue(context.Background(), workflow.BackendContextKey, temporal.GetBackend().RegisterWorkflow())
	env.SetWorkerOptions(tmpworker.Options{
		BackgroundActivityContext: ctx,
	})
	env.SetDataConverter(temporal.DataConverter{})
	service := &Service{
		Plugins:        p.Plugins,
		ClientTaskList: "test",
		Backend:        p.ServiceBackend,
	}

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
	testsuite.WorkflowTestSuite
	env *testsuite.TestActivityEnvironment
}

func NewTempTestActivitySuite() types.StarTestActivitySuite {
	s := starTempTestActivitySuite{}
	s.env = s.NewTestActivityEnvironment()
	ctx := context.WithValue(context.Background(), workflow.BackendContextKey, temporal.GetBackend().RegisterWorkflow())
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
