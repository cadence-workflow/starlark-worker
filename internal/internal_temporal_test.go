package internal

import (
	"reflect"
	"testing"
	"time"

	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.temporal.io/sdk/activity"
	temp "go.temporal.io/sdk/workflow"
	"go.uber.org/cadence/encoded"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// TemporalTestStruct test struct.
type TemporalTestStruct struct {
	ID int `json:"id"`
}

// TestToData tests that the converter can encode Go and Starlark values into bytes.
func TestTemporalToData(t *testing.T) {
	converter := newTemporalTestConverter(t)

	t.Run("encode-go-bytes", func(t *testing.T) {
		// Assert the encoder can encode Go bytes as-is.
		data, err := converter.ToData([]byte("abc"))
		require.NoError(t, err)
		require.Equal(t, []byte("abc"), data)
	})

	t.Run("encode-star-bytes", func(t *testing.T) {
		// Assert the encoder can encode Starlark bytes as-is.
		data, err := converter.ToData(starlark.Bytes("abc"))
		require.NoError(t, err)
		require.Equal(t, []byte("abc"), data)
	})

	t.Run("encode-go-map", func(t *testing.T) {
		// Assert the encoder can encode Go struct as JSON object.
		data, err := converter.ToData(TemporalTestStruct{ID: 101})
		require.NoError(t, err)
		require.Equal(t, []byte("{\"id\":101}\n"), data)
	})

	t.Run("encode-several-values", func(t *testing.T) {
		// Assert the encoder can encode several values using the cadenceDelimiter.
		data, err := converter.ToData(
			starlark.Tuple{starlark.String("pi"), starlark.Float(3.14)},
			true,
		)
		require.NoError(t, err)
		require.Equal(t, []byte("[\"pi\",3.14]\ntrue\n"), data)
	})
}

// TestFromData tests that the converter can decode bytes into Go and Starlark values.
func TestTemporalFromData(t *testing.T) {
	converter := newTemporalTestConverter(t)

	t.Run("decode-go-bytes", func(t *testing.T) {
		// Assert the encoder can decode into Go bytes as-is.
		var out []byte
		require.NoError(t, converter.FromData([]byte("abc"), &out))
		require.Equal(t, []byte("abc"), out)
	})

	t.Run("decode-star-bytes", func(t *testing.T) {
		// Assert the encoder can decode into Star bytes as-is.
		var out starlark.Bytes
		require.NoError(t, converter.FromData([]byte("abc"), &out))
		require.Equal(t, starlark.Bytes("abc"), out)
	})

	t.Run("decode-go-struct", func(t *testing.T) {
		// Assert the encoder can decode a JSON object into a Go struct.
		var out TemporalTestStruct
		require.NoError(t, converter.FromData([]byte("{\"id\":101}\n"), &out))
		require.Equal(t, TemporalTestStruct{ID: 101}, out)
	})

	t.Run("decode-no-cadenceDelimiter", func(t *testing.T) {
		// Assert the encoder can handle the case when the cadenceDelimiter is missing in the end.
		var out TemporalTestStruct
		require.NoError(t, converter.FromData([]byte("{\"id\":101}"), &out))
		require.Equal(t, TemporalTestStruct{ID: 101}, out)
	})

	t.Run("decode-several", func(t *testing.T) {
		// Assert the encoder can decode several delimited values.

		var out1 starlark.Tuple
		var out2 any
		var out3 starlark.Value
		require.NoError(t, converter.FromData([]byte("[\"pi\",3.14]\ntrue\ntrue\n"), &out1, &out2, &out3))

		require.Equal(t, starlark.Tuple{starlark.String("pi"), starlark.Float(3.14)}, out1)
		require.Equal(t, true, out2)
		require.Equal(t, starlark.True, out3)
	})

	t.Run("decode-incompatible-types", func(t *testing.T) {
		// Assert the encoder returns an error when trying to decode data into an incompatible type.
		var out starlark.String
		require.Error(t, converter.FromData([]byte("true\n"), &out))
	})
}

func newTemporalTestConverter(t *testing.T) encoded.DataConverter {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	return &CadenceDataConverter{Logger: logger}
}

// Mock workflow function for testing
func testTemporalWorkflow(ctx Context, input string) (string, error) {
	return "result", nil
}

// Mock temporal worker for testing
type mockTemporalWorker struct {
	registeredWorkflows []mockTemporalRegisteredWorkflow
}

type mockTemporalRegisteredWorkflow struct {
	workflow interface{}
	options  temp.RegisterOptions
}

func (m *mockTemporalWorker) RegisterWorkflowWithOptions(wf interface{}, options temp.RegisterOptions) {
	m.registeredWorkflows = append(m.registeredWorkflows, mockTemporalRegisteredWorkflow{
		workflow: wf,
		options:  options,
	})
}

func (m *mockTemporalWorker) RegisterWorkflow(wf interface{}) {}
func (m *mockTemporalWorker) RegisterActivity(a interface{}) {}
func (m *mockTemporalWorker) RegisterActivityWithOptions(runFunc interface{}, options activity.RegisterOptions) {}
func (m *mockTemporalWorker) RegisterNexusService(service *nexus.Service) {}
func (m *mockTemporalWorker) Start() error { return nil }
func (m *mockTemporalWorker) Run(interruptCh <-chan interface{}) error { return nil }
func (m *mockTemporalWorker) Stop() {}

// TestTemporalRegisterWorkflowWithOptions tests that RegisterWorkflowWithOptions properly transforms the workflow function
func TestTemporalRegisterWorkflowWithOptions(t *testing.T) {
	t.Run("workflow-function-transformation", func(t *testing.T) {
		// Create a mock temporal worker
		mockWorker := &mockTemporalWorker{}
		
		// Create TemporalWorker with mock
		temporalWorker := &TemporalWorker{Worker: mockWorker}
		
		// Test options
		options := RegisterWorkflowOptions{
			Name:                          "test-workflow",
			VersioningBehavior:            1,
			DisableAlreadyRegisteredCheck: false,
		}
		
		// Register workflow with options
		temporalWorker.RegisterWorkflowWithOptions(testTemporalWorkflow, options)
		
		// Verify that the workflow was registered
		require.Len(t, mockWorker.registeredWorkflows, 1)
		
		registered := mockWorker.registeredWorkflows[0]
		
		// Verify options were passed correctly
		require.Equal(t, "test-workflow", registered.options.Name)
		require.Equal(t, temp.VersioningBehavior(1), registered.options.VersioningBehavior)
		require.False(t, registered.options.DisableAlreadyRegisteredCheck)
		
		// Verify the workflow function was transformed
		registeredFunc := reflect.ValueOf(registered.workflow)
		require.Equal(t, reflect.Func, registeredFunc.Kind())
		
		// Verify the function signature - first parameter should be temp.Context
		funcType := registeredFunc.Type()
		require.True(t, funcType.NumIn() >= 1)
		require.Equal(t, reflect.TypeOf((*temp.Context)(nil)).Elem(), funcType.In(0))
	})
	
	t.Run("compare-with-register-workflow", func(t *testing.T) {
		// Test that RegisterWorkflowWithOptions behaves the same as RegisterWorkflow
		// in terms of function transformation
		mockWorker1 := &mockTemporalWorker{}
		mockWorker2 := &mockTemporalWorker{}
		
		temporalWorker1 := &TemporalWorker{Worker: mockWorker1}
		temporalWorker2 := &TemporalWorker{Worker: mockWorker2}
		
		// Register with RegisterWorkflow
		temporalWorker1.RegisterWorkflow(testTemporalWorkflow, "test-workflow")
		
		// Register with RegisterWorkflowWithOptions
		options := RegisterWorkflowOptions{Name: "test-workflow"}
		temporalWorker2.RegisterWorkflowWithOptions(testTemporalWorkflow, options)
		
		// Verify both registered workflows have the same transformed function signature
		require.Len(t, mockWorker1.registeredWorkflows, 1)
		require.Len(t, mockWorker2.registeredWorkflows, 1)
		
		func1Type := reflect.ValueOf(mockWorker1.registeredWorkflows[0].workflow).Type()
		func2Type := reflect.ValueOf(mockWorker2.registeredWorkflows[0].workflow).Type()
		
		// Both should have the same signature after transformation
		require.Equal(t, func1Type, func2Type)
		require.Equal(t, reflect.TypeOf((*temp.Context)(nil)).Elem(), func1Type.In(0))
		require.Equal(t, reflect.TypeOf((*temp.Context)(nil)).Elem(), func2Type.In(0))
	})
}

// Mock temporal context for testing - implements both internal.Context and temp.Context
type mockTemporalContext struct{}

// Implement internal.Context interface
func (m *mockTemporalContext) Value(key interface{}) interface{} { return nil }

// Implement temp.Context interface methods (extends context.Context)
func (m *mockTemporalContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (m *mockTemporalContext) Done() <-chan struct{} { return nil }
func (m *mockTemporalContext) Err() error { return nil }
