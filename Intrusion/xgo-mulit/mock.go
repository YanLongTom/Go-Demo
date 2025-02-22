package main

import (
	"reflect"
	"sync"
)

var mockFuncs = sync.Map{}

// 注册
func RegisterMockFunc(funcName string, fun interface{}) {
	mockFuncs.Store(funcName, fun)
}

func InterceptMock(funcName string, args []interface{}, results []interface{}) bool {
	mockFn, ok := mockFuncs.Load(funcName)
	if !ok {
		return false
	}

	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	mockFnValue := reflect.ValueOf(mockFn)
	out := mockFnValue.Call(in)
	if len(out) != len(results) {
		panic("mock function return value number is not equal to results number")
	}

	for i, result := range results {
		reflect.ValueOf(result).Elem().Set(out[i])
	}
	return true
}
