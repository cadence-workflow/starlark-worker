package workflow

import (
	"fmt"

	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/star"
	"github.com/cadence-workflow/starlark-worker/workflow"
	"go.starlark.net/starlark"
)

type BatchFuture struct {
	BatchFuture workflow.BatchFuture
}

var (
	_ starlark.HasAttrs = (*BatchFuture)(nil)
)

func (bf *BatchFuture) String() string        { return "workflow.batch_future" }
func (bf *BatchFuture) Type() string          { return "workflow.batch_future" }
func (bf *BatchFuture) Freeze()               {}
func (bf *BatchFuture) Truth() starlark.Bool  { return true }
func (bf *BatchFuture) Hash() (uint32, error) { return 0, fmt.Errorf("no-hash") }
func (bf *BatchFuture) AttrNames() []string {
	return star.AttrNames(batchFutureBuiltins, batchFutureProperties)
}
func (bf *BatchFuture) Attr(n string) (starlark.Value, error) {
	return star.Attr(bf, n, batchFutureBuiltins, batchFutureProperties)
}

func (bf *BatchFuture) Get(t *starlark.Thread) (starlark.Value, error) {
	ctx := service.GetContext(t)
	var res []starlark.Value // this needs to be pointer to a slice to retrieve all values for BatchFutures.Get(...)
	if err := bf.BatchFuture.Get(ctx, &res); err != nil {
		return nil, err
	}
	return starlark.NewList(res), nil
}

func (bf *BatchFuture) GetFutures(t *starlark.Thread) (starlark.Value, error) {
	var futures []starlark.Value
	for _, gotfuture := range bf.BatchFuture.GetFutures() {
		futures = append(futures, &Future{Future: gotfuture})
	}
	return starlark.Tuple(futures), nil
}

func (bf *BatchFuture) IsReady(t *starlark.Thread) (starlark.Value, error) {
	return starlark.Bool(bf.BatchFuture.IsReady()), nil
}

var batchFutureBuiltins = map[string]*starlark.Builtin{
	"get":         starlark.NewBuiltin("get", bfGet),
	"get_futures": starlark.NewBuiltin("get_futures", bfGetFutures),
	"is_ready":    starlark.NewBuiltin("is_ready", bfIsReady),
}

var batchFutureProperties = map[string]star.PropertyFactory{}

func bfGet(t *starlark.Thread, fn *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	recv := fn.Receiver().(*BatchFuture)
	return recv.Get(t)
}

func bfGetFutures(t *starlark.Thread, fn *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	recv := fn.Receiver().(*BatchFuture)
	return recv.GetFutures(t)
}
func bfIsReady(t *starlark.Thread, fn *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	recv := fn.Receiver().(*BatchFuture)
	return recv.IsReady(t)
}
