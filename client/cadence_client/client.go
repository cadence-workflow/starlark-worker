package cadence_client

import (
	"context"
	"fmt"
	"github.com/cadence-workflow/starlark-worker/cadence"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/star"
	"go.starlark.net/starlark"
	cadenceshared "go.uber.org/cadence/.gen/go/shared"
	cadenceclient "go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var WorkflowFunc any = (*service.Service)(nil).Run

func Run(
	tar []byte,
	file string,
	function string,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
	env *starlark.Dict,
	cadenceClient cadenceclient.Client,
	cadenceTaskList string,
) error {
	opt := cadenceclient.StartWorkflowOptions{
		TaskList:                     cadenceTaskList,
		ExecutionStartToCloseTimeout: time.Hour * 24 * 365 * 10,
	}

	ctx := context.Background()
	var exec *workflow.Execution
	var err error
	wf, _ := cadence.UpdateWorkflowFunctionContextArgument(WorkflowFunc)
	if exec, err = cadenceClient.StartWorkflow(
		ctx,
		opt,
		wf,
		tar,
		file,
		function,
		args,
		kwargs,
		env,
	); err != nil {
		return err
	}

	log.Printf("%-15s :: %s", "Execution.ID", exec.ID)
	log.Printf("%-15s :: %s", "Execution.RunID", exec.RunID)

	iter := cadenceClient.GetWorkflowHistory(ctx, exec.ID, exec.RunID, true, cadenceshared.HistoryEventFilterTypeCloseEvent)
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return err
		}
		if event.WorkflowExecutionCompletedEventAttributes != nil {
			result := string(event.WorkflowExecutionCompletedEventAttributes.GetResult())
			log.Printf("%-15s ::", "Result")
			log.Println(result)
		}
		if event.WorkflowExecutionFailedEventAttributes != nil {
			log.Printf("%-15s :: %s", "Fail.Reason", event.WorkflowExecutionFailedEventAttributes.GetReason())
			log.Printf("%-15s ::", "Fail.Details")
			log.Println(string(event.WorkflowExecutionFailedEventAttributes.GetDetails()))
		}
	}
	return nil
}

func Package(root, path string, out io.Writer) error {

	var err error
	if root, err = filepath.Abs(root); err != nil {
		return err
	}

	files := map[string][]byte{}

	onLoad := func(p string) error {

		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}

		log.Printf("[+] %s", rel)

		if content, err := os.ReadFile(p); err != nil {
			return err
		} else {
			files[rel] = content
		}
		return nil
	}

	if err = load(root, path, map[string]bool{}, onLoad); err != nil {
		return err
	}
	return ext.WriteTar(files, out)
}

func load(root string, path string, cache map[string]bool, callback func(string) error) error {

	if strings.HasPrefix(path, star.PluginPrefix) { // skip modules
		return nil
	}

	var err error

	if root, err = filepath.Abs(root); err != nil {
		return err
	}

	if strings.HasPrefix(path, "//") { // root relative path
		path = filepath.Join(root, path)
	}

	if path, err = filepath.Abs(path); err != nil {
		return err
	}

	if !strings.HasPrefix(path, root) { // file must be within the root
		return fmt.Errorf("out-of-root: root: %s, path: %s", root, path)
	}

	if _, found := cache[path]; found {
		return nil
	}

	if !strings.HasSuffix(path, ".star") {
		cache[path] = true
		if err := callback(path); err != nil {
			return err
		}
		return nil
	}

	yes := func(s string) bool { return true }
	if _, p, err := starlark.SourceProgramOptions(star.FileOptions, path, nil, yes); err != nil {
		return err
	} else {
		cache[path] = true
		if err := callback(path); err != nil {
			return err
		}
		for i := 0; i < p.NumLoads(); i++ {
			l, _ := p.Load(i)
			if err := load(root, l, cache, callback); err != nil {
				return err
			}
		}
	}
	return nil
}
