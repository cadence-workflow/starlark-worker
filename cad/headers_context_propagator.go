package cad

import (
	"context"
	"go.uber.org/cadence/workflow"
)

type HeadersContextPropagator struct{}

var _ workflow.ContextPropagator = (*HeadersContextPropagator)(nil)

func (r *HeadersContextPropagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	return inject(ctx, writer)
}

func (r *HeadersContextPropagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	headers := map[string][]byte{}
	if err := readHeaders(reader, headers); err != nil {
		return nil, err
	}
	return context.WithValue(ctx, contextKeyHeaders, headers), nil
}

func (r *HeadersContextPropagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	return inject(ctx, writer)
}

func (r *HeadersContextPropagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	headers := map[string][]byte{}
	if err := readHeaders(reader, headers); err != nil {
		return nil, err
	}
	return workflow.WithValue(ctx, contextKeyHeaders, headers), nil
}

func inject(ctx interface{ Value(key any) any }, writer workflow.HeaderWriter) error {
	for k, v := range GetContextHeaders(ctx) {
		writer.Set(k, v)
	}
	return nil
}

func readHeaders(reader workflow.HeaderReader, headers map[string][]byte) error {
	return reader.ForEachKey(func(key string, value []byte) error {
		headers[key] = value
		return nil
	})
}
