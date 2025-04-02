package worker

import "reflect"

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
