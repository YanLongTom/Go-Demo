package main

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func main() {
	tool, args := os.Args[1], os.Args[2:]

	if len(args) > 0 && args[0] == "-V=full" {
		// 跳过版本信息处理
	} else if len(args) > 0 {
		if filepath.Base(tool) == "compile" {
			if index := findTargetFile(args, "greet.go"); index > -1 {
				modifiedCode, err := generateInstrumentedCode(args[index])
				if err != nil {
					log.Fatal(err)
				}

				tmpFile, err := createTempFile(modifiedCode)
				if err != nil {
					log.Fatal(err)
				}
				defer os.Remove(tmpFile)

				args[index] = tmpFile
			}
		}
		log.Printf("Executing: %s %v\n", tool, args)
	}

	cmd := exec.Command(tool, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Execution failed: %v", err)
	}
}

// 生成包含插桩代码的AST
func buildInstrumentedAST(funcName string, params []string) dst.Node {
	return &dst.FuncDecl{
		Name: dst.NewIdent(funcName),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{
						Names: []*dst.Ident{dst.NewIdent("s")},
						Type:  dst.NewIdent("string"),
					},
				},
			},
			Results: &dst.FieldList{
				List: []*dst.Field{
					{
						Names: []*dst.Ident{dst.NewIdent("res")},
						Type:  dst.NewIdent("string"),
					},
				},
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				// 插入拦截逻辑
				&dst.IfStmt{
					Cond: &dst.CallExpr{
						Fun: dst.NewIdent("InterceptMock"),
						Args: []dst.Expr{
							dst.NewIdent(`"Greet"`),
							dst.NewIdent("s"),
							&dst.UnaryExpr{
								Op: token.AND,
								X:  dst.NewIdent("res"),
							},
						},
					},
					Body: &dst.BlockStmt{
						List: []dst.Stmt{
							&dst.ReturnStmt{
								Results: []dst.Expr{
									dst.NewIdent("res"),
								},
							},
						},
					},
				},
				// 保留原始逻辑
				&dst.ReturnStmt{
					Results: []dst.Expr{
						&dst.BinaryExpr{
							X:  dst.NewIdent(`"hello "`),
							Op: token.ADD,
							Y:  dst.NewIdent("s"),
						},
					},
				},
			},
		},
	}
}

// 生成插桩后的代码
func generateInstrumentedCode(filename string) (string, error) {
	// 解析原始文件
	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("解析失败: %w", err)
	}

	// 遍历AST查找目标函数
	var found bool
	dst.Inspect(f, func(n dst.Node) bool {
		if fn, ok := n.(*dst.FuncDecl); ok && fn.Name.Name == "Greet" {
			// 替换函数体
			fn.Body = buildInstrumentedAST("Greet", []string{"s"}).(*dst.FuncDecl).Body
			found = true
			return false
		}
		return true
	})

	if !found {
		return "", fmt.Errorf("未找到目标函数 Greet")
	}

	// 生成代码
	var buf bytes.Buffer
	if err := decorator.Fprint(&buf, f); err != nil {
		return "", fmt.Errorf("代码生成失败: %w", err)
	}

	return buf.String(), nil
}

// 创建临时文件（带随机后缀）
func createTempFile(content string) (string, error) {
	tmpFile, err := os.CreateTemp("", "instrumented_*.go")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}
	return tmpFile.Name(), nil
}

// 查找目标文件索引
func findTargetFile(args []string, pattern string) int {
	for i, arg := range args {
		if strings.Contains(arg, pattern) {
			return i
		}
	}
	return -1
}
