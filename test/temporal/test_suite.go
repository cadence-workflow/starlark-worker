package temporal

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/cadence-workflow/starlark-worker/internal/backend"
	"github.com/cadence-workflow/starlark-worker/internal/encoded"
	"github.com/cadence-workflow/starlark-worker/internal/temporal"
	"github.com/cadence-workflow/starlark-worker/internal/worker"
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/star"
	"github.com/cadence-workflow/starlark-worker/test"
	"github.com/stretchr/testify/require"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	tmpworker "go.temporal.io/sdk/worker"
	tmp "go.temporal.io/sdk/workflow"
	"strings"
	"testing"
)

type testEnvironment struct {
	env *testsuite.TestWorkflowEnvironment
}

func (t testEnvironment) RegisterWorkflowWithOptions(w interface{}, options worker.RegisterWorkflowOptions) {
	t.env.RegisterWorkflowWithOptions(temporal.UpdateWorkflowFunctionContextArgument(w), tmp.RegisterOptions{})
}
func (t testEnvironment) RegisterWorkflow(w interface{}) {
	t.env.RegisterWorkflow(temporal.UpdateWorkflowFunctionContextArgument(w))
}
func (t testEnvironment) RegisterActivity(a interface{}) {
	t.env.RegisterActivity(a)
}
func (t testEnvironment) RegisterActivityWithOptions(a interface{}, options worker.RegisterActivityOptions) {
	t.env.RegisterActivityWithOptions(a, activity.RegisterOptions{})
}

type StarTestEnvironment struct {
	env     *testsuite.TestWorkflowEnvironment
	service *service.Service
	tar     []byte
	fs      star.FS
}

func (r *StarTestEnvironment) GetTestWorkflowEnvironment() *testsuite.TestWorkflowEnvironment {
	return r.env
}

func (r *StarTestEnvironment) AssertExpectations(t *testing.T) {
	r.env.AssertExpectations(t)
}

func (r *StarTestEnvironment) ExecuteFunction(
	filePath string,
	fn string,
	args starlark.Tuple,
	kw []starlark.Tuple,
	environ *starlark.Dict,
) {
	r.env.ExecuteWorkflow(temporal.UpdateWorkflowFunctionContextArgument(r.service.Run), r.tar, filePath, fn, args, kw, environ)
}

func (r *StarTestEnvironment) GetResult(valuePtr any) error {
	if !r.env.IsWorkflowCompleted() {
		return fmt.Errorf("workflow is not completed")
	}
	if err := r.env.GetWorkflowError(); err != nil {
		return err
	}
	return r.env.GetWorkflowResult(valuePtr)
}

func (r *StarTestEnvironment) GetTestFunctions(filePath string) ([]string, error) {
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

func (r *StarTestEnvironment) getGlobals(filePath string) ([]*resolve.Binding, error) {
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

type StarTestSuite struct {
	testsuite.WorkflowTestSuite
	tarCache map[string][]byte
	fsCache  map[string]star.FS
}

type StarTestEnvironmentParams struct {
	RootDirectory  string
	Plugins        map[string]service.IPlugin
	DataConvertor  encoded.DataConvertor
	ServiceBackend backend.Backend
}

func (r *StarTestSuite) NewEnvironment(t *testing.T, p *StarTestEnvironmentParams) *StarTestEnvironment {
	env := r.NewTestWorkflowEnvironment()
	service := &service.Service{
		Plugins:        p.Plugins,
		ClientTaskList: "test",
		Workflow:       p.ServiceBackend.RegisterWorkflow(),
	}
	testEnv := testEnvironment{env: env}

	for _, plugin := range p.Plugins {
		plugin.Register(testEnv)
	}

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

	return &StarTestEnvironment{
		env:     env,
		service: service,
		tar:     tar,
		fs:      fs,
	}
}

func (r *StarTestSuite) buildTar(t *testing.T, dir string) ([]byte, bool) {
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

type starTestActivitySuite struct {
	testsuite.WorkflowTestSuite
	env *testsuite.TestActivityEnvironment
}

func NewTestActivitySuite() test.StarTestActivitySuite {
	s := starTestActivitySuite{}
	s.env = s.NewTestActivityEnvironment()
	ctx := context.WithValue(context.Background(), "backendContextKey", temporal.GetBackend().RegisterWorkflow())
	s.env.SetWorkerOptions(tmpworker.Options{
		BackgroundActivityContext: ctx,
	})
	return s
}

func (r starTestActivitySuite) RegisterActivity(a interface{}) {
	r.env.RegisterActivity(a)
}

func (r starTestActivitySuite) ExecuteActivity(a interface{}, opts interface{}) (test.EncodedValue, error) {
	return r.env.ExecuteActivity(a, opts)
}
