package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
)

// 嵌入单个文件为字符串
//go:embed static.txt
var staticContent string

// 嵌入单个文件为字节切片
//go:embed config.json
var configData []byte

// 嵌入多个文件到embed.FS文件系统
//go:embed static.txt template.html config.json
var embeddedFiles embed.FS

// 使用通配符嵌入所有.txt文件
//go:embed *.txt
var textFiles embed.FS

type Config struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Settings struct {
		Debug   bool `json:"debug"`
		Timeout int  `json:"timeout"`
	} `json:"settings"`
}

func main() {
	fmt.Println("=== Go embed包完整使用示例 ===\n")

	// 1. 基础用法：直接嵌入为字符串
	fmt.Println("1. 嵌入为字符串：")
	fmt.Println(staticContent)
	fmt.Println()

	// 2. 嵌入为字节切片并解析JSON
	fmt.Println("2. 嵌入JSON配置文件：")
	var config Config
	err := json.Unmarshal(configData, &config)
	if err != nil {
		fmt.Printf("解析JSON失败: %v\n", err)
	} else {
		fmt.Printf("配置名称: %s\n", config.Name)
		fmt.Printf("版本: %s\n", config.Version)
		fmt.Printf("调试模式: %t\n", config.Settings.Debug)
	}
	fmt.Println()

	// 3. 使用embed.FS读取文件
	fmt.Println("3. 使用embed.FS读取文件：")

	// 读取HTML文件
	htmlContent, err := embeddedFiles.ReadFile("template.html")
	if err != nil {
		fmt.Printf("读取HTML文件失败: %v\n", err)
	} else {
		fmt.Printf("HTML文件大小: %d 字节\n", len(htmlContent))
		fmt.Printf("HTML文件前100个字符: %s...\n", string(htmlContent[:100]))
	}
	fmt.Println()

	// 4. 遍历嵌入的文件系统
	fmt.Println("4. 遍历嵌入的文件：")
	err = fs.WalkDir(embeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, _ := d.Info()
			fmt.Printf("文件: %s, 大小: %d 字节\n", path, info.Size())
		}
		return nil
	})
	if err != nil {
		fmt.Printf("遍历文件系统失败: %v\n", err)
	}
	fmt.Println()

	// 5. 使用通配符嵌入的文件
	fmt.Println("5. 通配符嵌入的.txt文件：")
	err = fs.WalkDir(textFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			content, _ := textFiles.ReadFile(path)
			fmt.Printf("文件 %s 的内容：\n%s\n", path, string(content))
		}
		return nil
	})
	if err != nil {
		fmt.Printf("读取.txt文件失败: %v\n", err)
	}

	fmt.Println("\n=== embed包的主要优势 ===")
	fmt.Println("✓ 编译时嵌入，运行时无需外部文件")
	fmt.Println("✓ 部署简单，只需一个可执行文件")
	fmt.Println("✓ 避免文件路径问题")
	fmt.Println("✓ 提高程序启动速度")
}
