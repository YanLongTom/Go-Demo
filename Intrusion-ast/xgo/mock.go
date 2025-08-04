package main

import "sync"

var mockFuncs = sync.Map{}

// 注册
func RegisterMockFunc(funcName string, fun interface{}) {
	mockFuncs.Store(funcName, fun)
}

// 拦截器
func InterceptMock(funcName string, arg string, result *string) bool {
	fn, ok := mockFuncs.Load(funcName)
	if ok {
		f, ok := fn.(func(s string) string)
		if ok {
			*result = f(arg)
			return true
		}
	}
	return false
}
