package main

//
//import (
//	_ "embed" // 必须导入embed包
//	"fmt"
//)
//
//// 使用//go:embed指令嵌入文件内容到字符串变量
////go:embed static.txt
//var staticContent string
//
//// 也可以嵌入为字节切片
////go:embed static.txt
//var staticBytes []byte
//
//func main() {
//	fmt.Println("=== embed包基础使用示例 ===")
//
//	// 1. 读取嵌入的字符串内容
//	fmt.Println("嵌入的文件内容（字符串形式）：")
//	fmt.Println(staticContent)
//
//	// 2. 读取嵌入的字节切片内容
//	fmt.Printf("\n嵌入的文件内容（字节切片形式，长度：%d）：\n", len(staticBytes))
//	fmt.Println(string(staticBytes))
//
//	// 3. 验证两种方式内容一致
//	fmt.Printf("\n字符串和字节切片内容是否一致：%t\n", staticContent == string(staticBytes))
//}
