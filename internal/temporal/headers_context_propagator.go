package temporal

import (
	"context"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"

	"go.temporal.io/sdk/workflow"
)

type contextKey int

const (
	contextKeyHeaders contextKey = iota
)

func GetContextHeaders(ctx interface{ Value(key any) any }) map[string][]byte {
	return ctx.Value(contextKeyHeaders).(map[string][]byte)
}

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
	headers := GetContextHeaders(ctx)
	if headers == nil {
		return nil
	}
	for k, v := range headers {
		payload, err := converter.GetDefaultDataConverter().ToPayload(v)
		if err != nil {
			return err
		}
		writer.Set(k, payload)
	}
	return nil
}

func readHeaders(reader workflow.HeaderReader, headers map[string][]byte) error {
	return reader.ForEachKey(func(key string, payload *commonpb.Payload) error {
		var value []byte
		err := converter.GetDefaultDataConverter().FromPayload(payload, &value)
		if err != nil {
			return err
		}
		headers[key] = value
		return nil
	})
}
