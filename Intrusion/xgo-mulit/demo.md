### 跳转
通过注册到 mockFuncs map中，在源执行函数中。添加从map获取函数流程，实现跳转。

- 需要有注册函数 RegisterMockFunc 存储新函数(eg: 主函数调用前每次都执行注册，然后再调用)
- 跳转函数 InterceptMock 的插入与获取。( InterceptMock 预定义，以及在源函数定义中插入调用 InterceptMock 函数逻辑)

### 调用过程
1. 主函数注册
```go
	RegisterMockFunc("Other", func(i int, s string, f float64) string {
		return fmt.Sprintf("mock %d %s %.2f", i, s, f)
	})
```
2. 注册到 map 中，通过 InterceptMock 调用
```go
var mockFuncs = sync.Map{}

// 注册
func RegisterMockFunc(funcName string, fun interface{}) {
	mockFuncs.Store(funcName, fun)
}

func InterceptMock(funcName string, args []interface{}, results []interface{}) bool {
	mockFn, ok := mockFuncs.Load(funcName)
	...
```
3.重写

原函数：
```go
func Other(i int, s string, f float64) string {
	return fmt.Sprintf("int: %d, string: %s, float: %f", i, s, f)
}
```
重写后：
```go
func Other(i int, s string, f float64) (_xgo_res_1 string) {
	// 调用 InterceptMock
	if InterceptMock("Other", []interface{}{i, s, f}, []interface{}{&_xgo_res_1}) {
		return _xgo_res_1
	}

	return fmt.Sprintf("int: %d, string: %s, float: %f", i, s, f)
}
```