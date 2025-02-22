package main

import "sync"

var mockFuncs = sync.Map{}

// 注册
func RegisterMockFunc(funcName string, fun interface{}) {
	mockFuncs.Store(funcName, fun)
}

// InterceptMock 拦截，拦截原函数 Greet 调用。类似蹦床函数
// 需要重写 Greet 函数，使得可以跳转到 InterceptMock 函数。
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
