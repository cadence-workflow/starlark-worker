package internal

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

func GetRemainingInTypes(fnType reflect.Type) []reflect.Type {
	var types []reflect.Type
	for i := 1; i < fnType.NumIn(); i++ {
		types = append(types, fnType.In(i))
	}
	return types
}

func GetOutTypes(fnType reflect.Type) []reflect.Type {
	var types []reflect.Type
	for i := 0; i < fnType.NumOut(); i++ {
		types = append(types, fnType.Out(i))
	}
	return types
}

func UpdateWorkflowFunctionContextArgument(wf interface{}, contextType reflect.Type) (interface{}, string) {
	originalFunc := reflect.ValueOf(wf)
	originalType := originalFunc.Type()

	if originalType.Kind() != reflect.Func || originalType.NumIn() == 0 {
		panic("workflow function must be a function and have at least one argument (context)")
	}

	// Get original function name (if available)
	funcName := runtime.FuncForPC(originalFunc.Pointer()).Name()

	// Build a new function with the same signature but the provided context type
	wrappedFuncType := reflect.FuncOf(
		append([]reflect.Type{contextType}, GetRemainingInTypes(originalType)...),
		GetOutTypes(originalType),
		false,
	)

	wrappedFunc := reflect.MakeFunc(wrappedFuncType, func(args []reflect.Value) []reflect.Value {
		newArgs := make([]reflect.Value, len(args))
		newArgs[0] = args[0].Convert(contextType)
		for i := 1; i < len(args); i++ {
			newArgs[i] = args[i]
		}
		return originalFunc.Call(newArgs)
	})
	funcName = strings.TrimSuffix(funcName, "-fm")
	fmt.Printf("Registering workflow: %s", funcName)

	return wrappedFunc.Interface(), funcName
}
