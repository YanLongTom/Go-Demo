package main

// 重写后：
//{
//	if InterceptMock("Greet", s, &res) {
//		return res
//	}
//	return "hello " + s
//}

func Greet(s string) (res string) {
	return "hello " + s
}

func Greet2(s2 string) (res2 string) {
	return "hello 2 " + s2
}

func Greet3(s3 string) string {
	return "hello 3 " + s3
}
