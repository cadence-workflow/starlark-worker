package internal

import (
	"reflect"
)

// GetRemainingInTypes returns a slice of reflect.Type containing all input types
// except the first one.
func GetRemainingInTypes(fnType reflect.Type) []reflect.Type {
	var types []reflect.Type
	for i := 1; i < fnType.NumIn(); i++ {
		types = append(types, fnType.In(i))
	}
	return types
}

// GetOutTypes returns a slice of reflect.Type containing all output types
// of the provided function type.
func GetOutTypes(fnType reflect.Type) []reflect.Type {
	var types []reflect.Type
	for i := 0; i < fnType.NumOut(); i++ {
		types = append(types, fnType.Out(i))
	}
	return types
}

// UpdateWorkflowFunctionContextArgument takes a workflow function and a new context type,
// and returns a new function with the same signature but with the first argument
// replaced by the provided context type. The rest of the arguments remain unchanged.
// It is useful for adapting workflow functions to different context types.
// The original function must have at least one argument (the context).
// If the original function does not have a context argument, it panics.
func UpdateWorkflowFunctionContextArgument(wf interface{}, contextType reflect.Type) interface{} {
	originalFunc := reflect.ValueOf(wf)
	originalType := originalFunc.Type()

	if originalType.Kind() != reflect.Func || originalType.NumIn() == 0 {
		panic("workflow function must be a function and have at least one argument (context)")
	}

	// Build a new function with the same signature but the provided context type
	wrappedFuncType := reflect.FuncOf(
		append([]reflect.Type{contextType}, GetRemainingInTypes(originalType)...),
		GetOutTypes(originalType),
		false,
	)

	return reflect.MakeFunc(wrappedFuncType, func(args []reflect.Value) []reflect.Value {
		newArgs := make([]reflect.Value, len(args))
		newArgs[0] = args[0].Convert(contextType)
		for i := 1; i < len(args); i++ {
			newArgs[i] = args[i]
		}
		return originalFunc.Call(newArgs)
	}).Interface()
}
