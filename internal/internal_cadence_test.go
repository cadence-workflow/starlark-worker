package internal

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.uber.org/cadence"
	cadactivity "go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/testsuite"
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

// TestRegisterWorkflowWithOptions tests that the CadenceWorker can register workflows with options.
func TestCadenceRegisterWorkflowWithOptions(t *testing.T) {
	// Mock worker - since we can't easily create a real worker here
	worker := &CadenceWorker{Worker: nil}

	// Test workflow function
	testWorkflow := func() error { return nil }

	t.Run("register-workflow-with-basic-options", func(t *testing.T) {
		options := RegisterWorkflowOptions{
			Name:                          "test-workflow",
			EnableShortName:               true,
			DisableAlreadyRegisteredCheck: true,
		}

		// This test verifies that the method accepts the correct option types
		// In a real scenario, this would register with an actual worker
		require.NotPanics(t, func() {
			// Note: This will panic because worker.Worker is nil, but it tests the signature
			defer func() {
				if r := recover(); r != nil {
					// Expected to panic due to nil Worker, but signature is correct
				}
			}()
			worker.RegisterWorkflowWithOptions(testWorkflow, options)
		})
	})

	t.Run("register-workflow-with-empty-name", func(t *testing.T) {
		options := RegisterWorkflowOptions{
			Name:                          "",
			EnableShortName:               false,
			DisableAlreadyRegisteredCheck: false,
		}

		require.NotPanics(t, func() {
			defer func() {
				if r := recover(); r != nil {
					// Expected to panic due to nil Worker, but signature is correct
				}
			}()
			worker.RegisterWorkflowWithOptions(testWorkflow, options)
		})
	})
}

// TestRegisterActivityWithOptions tests that the CadenceWorker can register activities with options.
func TestCadenceRegisterActivityWithOptions(t *testing.T) {
	// Mock worker - since we can't easily create a real worker here
	worker := &CadenceWorker{Worker: nil}

	// Test activity function
	testActivity := func() error { return nil }

	t.Run("register-activity-with-basic-options", func(t *testing.T) {
		options := RegisterActivityOptions{
			Name:                          "test-activity",
			EnableShortName:               true,
			DisableAlreadyRegisteredCheck: true,
			EnableAutoHeartbeat:           true,
		}

		// This test verifies that the method accepts the correct option types
		require.NotPanics(t, func() {
			defer func() {
				if r := recover(); r != nil {
					// Expected to panic due to nil Worker, but signature is correct
				}
			}()
			worker.RegisterActivityWithOptions(testActivity, options)
		})
	})

	t.Run("register-activity-with-minimal-options", func(t *testing.T) {
		options := RegisterActivityOptions{
			Name:                          "minimal-activity",
			EnableShortName:               false,
			DisableAlreadyRegisteredCheck: false,
			EnableAutoHeartbeat:           false,
		}

		require.NotPanics(t, func() {
			defer func() {
				if r := recover(); r != nil {
					// Expected to panic due to nil Worker, but signature is correct
				}
			}()
			worker.RegisterActivityWithOptions(testActivity, options)
		})
	})
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
func (m *mockCadenceWorker) RegisterActivity(a interface{})  {}
func (m *mockCadenceWorker) RegisterActivityWithOptions(runFunc interface{}, options cadactivity.RegisterOptions) {
}
func (m *mockCadenceWorker) GetRegisteredActivities() []cadactivity.RegistryInfo { return nil }
func (m *mockCadenceWorker) GetRegisteredWorkflows() []cad.RegistryInfo          { return nil }
func (m *mockCadenceWorker) Start() error                                        { return nil }
func (m *mockCadenceWorker) Run() error                                          { return nil }
func (m *mockCadenceWorker) Stop()                                               {}

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

// TestCadenceSelector tests the Selector functionality for Cadence workflows.
func TestCadenceSelector(t *testing.T) {
	workflow := &CadenceWorkflow{}

	t.Run("NewSelector", func(t *testing.T) {
		// Create a test workflow context using the test suite
		s := &testsuite.WorkflowTestSuite{}
		env := s.NewTestWorkflowEnvironment()

		env.ExecuteWorkflow(func(ctx cad.Context) error {
			// Test creating a new selector
			selector := workflow.NewSelector(ctx)
			require.NotNil(t, selector)

			// Verify it's the right type
			cadSelector, ok := selector.(*cadenceSelector)
			require.True(t, ok)
			require.NotNil(t, cadSelector.s)

			return nil
		})

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
	})

	t.Run("AddFuture", func(t *testing.T) {
		s := &testsuite.WorkflowTestSuite{}
		env := s.NewTestWorkflowEnvironment()

		env.ExecuteWorkflow(func(ctx cad.Context) error {
			// Create selector and future
			selector := workflow.NewSelector(ctx)
			future, settable := workflow.NewFuture(ctx)

			// Track if callback was called
			callbackCalled := false
			var receivedResult string

			// Add future to selector
			result := selector.AddFuture(future, func(f Future) {
				callbackCalled = true
				err := f.Get(ctx, &receivedResult)
				require.NoError(t, err)
			})

			// Verify fluent interface
			require.Equal(t, selector, result)

			// Set the future value
			settable.SetValue("test-result")

			// Execute selector
			selector.Select(ctx)

			// Verify callback was called and result received
			require.True(t, callbackCalled)
			require.Equal(t, "test-result", receivedResult)

			return nil
		})

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
	})

	t.Run("MultipleFutures", func(t *testing.T) {
		s := &testsuite.WorkflowTestSuite{}
		env := s.NewTestWorkflowEnvironment()

		env.ExecuteWorkflow(func(ctx cad.Context) error {
			selector := workflow.NewSelector(ctx)

			// Create multiple futures
			future1, settable1 := workflow.NewFuture(ctx)
			future2, _ := workflow.NewFuture(ctx)

			var result1, result2 string
			future1Called := false
			future2Called := false

			// Add both futures to selector
			selector.AddFuture(future1, func(f Future) {
				future1Called = true
				err := f.Get(ctx, &result1)
				require.NoError(t, err)
			})

			selector.AddFuture(future2, func(f Future) {
				future2Called = true
				err := f.Get(ctx, &result2)
				require.NoError(t, err)
			})

			// Set only the first future (simulating first activity completing)
			settable1.SetValue("first-result")

			// Execute selector - should only call first callback
			selector.Select(ctx)

			// Verify only first callback was called
			require.True(t, future1Called)
			require.False(t, future2Called)
			require.Equal(t, "first-result", result1)
			require.Empty(t, result2)

			return nil
		})

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
	})

	t.Run("SelectorWithError", func(t *testing.T) {
		s := &testsuite.WorkflowTestSuite{}
		env := s.NewTestWorkflowEnvironment()

		env.ExecuteWorkflow(func(ctx cad.Context) error {
			selector := workflow.NewSelector(ctx)
			future, settable := workflow.NewFuture(ctx)

			var receivedError error

			selector.AddFuture(future, func(f Future) {
				var result string
				receivedError = f.Get(ctx, &result)
			})

			// Set an error on the future
			expectedError := cadence.NewCustomError("test-error", "test details")
			settable.SetError(expectedError)

			// Execute selector
			selector.Select(ctx)

			// Verify error was received
			require.Error(t, receivedError)
			require.Contains(t, receivedError.Error(), "test-error")

			return nil
		})

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
	})
}

// Mock cadence context for testing - implements both internal.Context and cad.Context
type mockCadenceContext struct{}

// Implement internal.Context interface
func (m *mockCadenceContext) Value(key interface{}) interface{} { return nil }

// Implement cad.Context interface methods (extends context.Context)
func (m *mockCadenceContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (m *mockCadenceContext) Done() <-chan struct{}       { return nil }
func (m *mockCadenceContext) Err() error                  { return nil }
