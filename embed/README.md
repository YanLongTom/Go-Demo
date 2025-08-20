# Go embed 包使用示例

这个目录展示了 Go 1.16 引入的 `embed` 包的基础和高级使用方法。

## 什么是 embed 包？

`embed` 包允许你在编译时将文件内容嵌入到 Go 程序的二进制文件中。这意味着：

- **无需外部文件依赖**：所有资源都打包在可执行文件中
- **简化部署**：只需分发一个二进制文件
- **避免路径问题**：不用担心相对路径或文件丢失
- **提高性能**：避免运行时的文件I/O操作

## 基础用法示例

### 1. 基础示例 (main.go)

```bash
go run main.go
```

演示了最简单的embed用法：
- 将文件嵌入为字符串
- 将文件嵌入为字节切片

### 2. 高级示例 (advanced_example.go)

```bash
go run advanced_example.go
```

展示了更多高级特性：
- 使用 `embed.FS` 文件系统
- 通配符模式匹配
- 遍历嵌入的文件
- JSON配置文件解析

## embed 指令语法

### 基本语法
```go
//go:embed 文件名
var content string  // 嵌入为字符串

//go:embed 文件名
var data []byte     // 嵌入为字节切片

//go:embed 文件名
var fs embed.FS     // 嵌入为文件系统
```

### 支持的模式
```go
//go:embed file.txt              // 单个文件
//go:embed *.txt                 // 通配符匹配
//go:embed dir                   // 整个目录
//go:embed dir/*.html            // 目录下特定文件
//go:embed file1.txt file2.txt   // 多个文件
```

## 使用场景

1. **Web应用静态资源**：HTML、CSS、JS文件
2. **配置文件**：JSON、YAML、TOML配置
3. **模板文件**：Go template、其他模板
4. **数据文件**：CSV、JSON数据
5. **文档文件**：README、帮助文档

## 注意事项

1. **编译时嵌入**：文件内容在编译时确定，运行时无法修改
2. **文件大小**：嵌入的文件会增加二进制文件大小
3. **路径限制**：只能嵌入当前模块内的文件
4. **版本要求**：需要 Go 1.16 或更高版本

## 优势

- ✅ 简化部署流程
- ✅ 避免文件丢失问题  
- ✅ 提高程序启动速度
- ✅ 更好的安全性（文件内容不可外部访问）

## 局限性

- ❌ 增加二进制文件大小
- ❌ 运行时无法修改嵌入内容
- ❌ 只能嵌入模块内文件
- ❌ 大文件会影响编译速度

## 文件说明

- `static.txt` - 用于嵌入的静态文本文件
- `config.json` - JSON配置文件示例
- `template.html` - HTML模板文件示例
- `main.go` - 基础embed使用示例
- `advanced_example.go` - 高级embed功能示例
