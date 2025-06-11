package internal

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.uber.org/cadence/encoded"
	cadactivity "go.uber.org/cadence/activity"
	cad "go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// CadenceTestStruct test struct.
type CadenceTestStruct struct {
	ID int `json:"id"`
}

// TestToData tests that the converter can encode Go and Starlark values into bytes.
func TestCadenceToData(t *testing.T) {
	converter := newCadenceTestConverter(t)

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
		data, err := converter.ToData(CadenceTestStruct{ID: 101})
		require.NoError(t, err)
		require.Equal(t, []byte("{\"id\":101}\n"), data)
	})

	t.Run("encode-several-values", func(t *testing.T) {
		// Assert the encoder can encode several values using the delimiter.
		data, err := converter.ToData(
			starlark.Tuple{starlark.String("pi"), starlark.Float(3.14)},
			true,
		)
		require.NoError(t, err)
		require.Equal(t, []byte("[\"pi\",3.14]\ntrue\n"), data)
	})
}

// TestFromData tests that the converter can decode bytes into Go and Starlark values.
func TestCadenceFromData(t *testing.T) {
	converter := newCadenceTestConverter(t)

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
		var out CadenceTestStruct
		require.NoError(t, converter.FromData([]byte("{\"id\":101}\n"), &out))
		require.Equal(t, CadenceTestStruct{ID: 101}, out)
	})

	t.Run("decode-no-delimiter", func(t *testing.T) {
		// Assert the encoder can handle the case when the delimiter is missing in the end.
		var out CadenceTestStruct
		require.NoError(t, converter.FromData([]byte("{\"id\":101}"), &out))
		require.Equal(t, CadenceTestStruct{ID: 101}, out)
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

func newCadenceTestConverter(t *testing.T) encoded.DataConverter {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	return &CadenceDataConverter{Logger: logger}
}

// Mock workflow function for testing
func testWorkflow(ctx Context, input string) (string, error) {
	return "result", nil
}

// Mock cadence worker for testing
type mockCadenceWorker struct {
	registeredWorkflows []mockRegisteredWorkflow
}

type mockRegisteredWorkflow struct {
	workflow interface{}
	options  cad.RegisterOptions
}

func (m *mockCadenceWorker) RegisterWorkflowWithOptions(wf interface{}, options cad.RegisterOptions) {
	m.registeredWorkflows = append(m.registeredWorkflows, mockRegisteredWorkflow{
		workflow: wf,
		options:  options,
	})
}

func (m *mockCadenceWorker) RegisterWorkflow(wf interface{}) {}
func (m *mockCadenceWorker) RegisterActivity(a interface{}) {}
func (m *mockCadenceWorker) RegisterActivityWithOptions(runFunc interface{}, options cadactivity.RegisterOptions) {}
func (m *mockCadenceWorker) Start() error { return nil }
func (m *mockCadenceWorker) Run() error { return nil }
func (m *mockCadenceWorker) Stop() {}

// TestRegisterWorkflowWithOptions tests that RegisterWorkflowWithOptions properly transforms the workflow function
func TestRegisterWorkflowWithOptions(t *testing.T) {
	t.Run("workflow-function-transformation", func(t *testing.T) {
		// Create a mock cadence worker
		mockWorker := &mockCadenceWorker{}
		
		// Create CadenceWorker with mock
		cadenceWorker := &CadenceWorker{Worker: mockWorker}
		
		// Test options
		options := RegisterWorkflowOptions{
			Name:                          "test-workflow",
			EnableShortName:               true,
			DisableAlreadyRegisteredCheck: false,
		}
		
		// Register workflow with options
		cadenceWorker.RegisterWorkflowWithOptions(testWorkflow, options)
		
		// Verify that the workflow was registered
		require.Len(t, mockWorker.registeredWorkflows, 1)
		
		registered := mockWorker.registeredWorkflows[0]
		
		// Verify options were passed correctly
		require.Equal(t, "test-workflow", registered.options.Name)
		require.True(t, registered.options.EnableShortName)
		require.False(t, registered.options.DisableAlreadyRegisteredCheck)
		
		// Verify the workflow function was transformed
		registeredFunc := reflect.ValueOf(registered.workflow)
		require.Equal(t, reflect.Func, registeredFunc.Kind())
		
		// Verify the function signature - first parameter should be cad.Context
		funcType := registeredFunc.Type()
		require.True(t, funcType.NumIn() >= 1)
		require.Equal(t, reflect.TypeOf((*cad.Context)(nil)).Elem(), funcType.In(0))
	})
	
	t.Run("compare-with-register-workflow", func(t *testing.T) {
		// Test that RegisterWorkflowWithOptions behaves the same as RegisterWorkflow
		// in terms of function transformation
		mockWorker1 := &mockCadenceWorker{}
		mockWorker2 := &mockCadenceWorker{}
		
		cadenceWorker1 := &CadenceWorker{Worker: mockWorker1}
		cadenceWorker2 := &CadenceWorker{Worker: mockWorker2}
		
		// Register with RegisterWorkflow
		cadenceWorker1.RegisterWorkflow(testWorkflow, "test-workflow")
		
		// Register with RegisterWorkflowWithOptions
		options := RegisterWorkflowOptions{Name: "test-workflow"}
		cadenceWorker2.RegisterWorkflowWithOptions(testWorkflow, options)
		
		// Verify both registered workflows have the same transformed function signature
		require.Len(t, mockWorker1.registeredWorkflows, 1)
		require.Len(t, mockWorker2.registeredWorkflows, 1)
		
		func1Type := reflect.ValueOf(mockWorker1.registeredWorkflows[0].workflow).Type()
		func2Type := reflect.ValueOf(mockWorker2.registeredWorkflows[0].workflow).Type()
		
		// Both should have the same signature after transformation
		require.Equal(t, func1Type, func2Type)
		require.Equal(t, reflect.TypeOf((*cad.Context)(nil)).Elem(), func1Type.In(0))
		require.Equal(t, reflect.TypeOf((*cad.Context)(nil)).Elem(), func2Type.In(0))
	})
}

// Mock cadence context for testing - implements both internal.Context and cad.Context
type mockCadenceContext struct{}

// Implement internal.Context interface
func (m *mockCadenceContext) Value(key interface{}) interface{} { return nil }

// Implement cad.Context interface methods (extends context.Context)
func (m *mockCadenceContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (m *mockCadenceContext) Done() <-chan struct{} { return nil }
func (m *mockCadenceContext) Err() error { return nil }
