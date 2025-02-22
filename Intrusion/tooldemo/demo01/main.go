package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// 显示当前时间
func showTime() {
	fmt.Println("当前时间：", time.Now().Format(time.RFC1123))
}

// 执行 go build命令来编译Go代码”
func buildGoFile(filePath string) {
	cmd := exec.Command("go", "build", filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("编译失败:", err)
		return
	}
	fmt.Println("编译成功:", filePath)
}

// 显示帮助信息
func showHelp() {
	fmt.Println("使用方法：")
	fmt.Println("  toolexec time    - 显示当前时间")
	fmt.Println("  toolexec build <file.go> - 编译Go文件")
	fmt.Println("  toolexec help    - 显示帮助信息")
}

// 执行用户请求的命令
func executeCommand(command string, args []string) {
	switch command {
	case "time":
		showTime()
	case "build":
		if len(args) < 1 {
			fmt.Println("错误: 需要指定Go文件路径")
			showHelp()
			return
		}
		buildGoFile(args[0])
	case "help":
		showHelp()
	default:
		fmt.Printf("未知命令: %s\n", command)
		showHelp()
	}
}

func main() {
	// 检查是否传入了命令参数
	if len(os.Args) < 2 {
		// 如果没有传入命令参数，打印使用方法并退出
		fmt.Println("Usage: toolexec <command> [args...]")
		os.Exit(1)
	}

	// 获取用户传入的第一个命令参数
	command := os.Args[1]
	args := os.Args[2:]

	// 执行相应的命令
	executeCommand(command, args)
}
