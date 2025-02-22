package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	tool, args := os.Args[1], os.Args[2:]

	if len(args) > 0 && args[0] == "-V=full" {
		// don't do anything to infuence the version full output.
	} else if len(args) > 0 {
		if filepath.Base(tool) == "compile" {
			index := findGreetFile(args)
			if index > -1 {
				filename := args[index]
				f, err := os.Create("tmp.go")
				defer f.Close()
				defer os.Remove("tmp.go")
				if err != nil {
					log.Fatalf("create tmp.go error: %v\n", err)
				}
				_, _ = f.WriteString(insertCode(filename))
				args[index] = "tmp.go"
			}
		}
		fmt.Printf("tool: %s\n", tool)
		fmt.Printf("args: %v\n", args)
	}
	// 继续执行之前的命令
	cmd := exec.Command(tool, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("run command error: %v\n", err)
	}
}

func findGreetFile(args []string) int {
	for i, arg := range args {
		if strings.Contains(arg, "greet.go") {
			return i
		}
	}
	return -1
}

func insertCode(filename string) string {
	fset := token.NewFileSet()
	fast, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		log.Fatalf("parse file error: %v\n", err)
	}

	for _, decl := range fast.Decls {
		fun, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		f, err := os.Create("tmp2.go")
		if err != nil {
			log.Fatalf("create tmp2.go error: %v\n", err)
		}
		_, _ = f.WriteString(newCode(fun))
		f.Close()

		tmpFset := token.NewFileSet()
		tmpF, err := parser.ParseFile(tmpFset, "tmp2.go", nil, parser.AllErrors)
		if err != nil {
			log.Fatalf("parse tmp2.go error: %v\n", err)
		}
		fun.Body.List = append(tmpF.Decls[0].(*ast.FuncDecl).Body.List, fun.Body.List...)
		os.Remove("tmp2.go")
	}

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, fast)

	fmt.Println(buf.String())

	return buf.String()
}

func newCode(fun *ast.FuncDecl) string {
	// 函数名称
	funcName := fun.Name.Name

	// 参数列表
	args := make([]string, 0)
	for _, arg := range fun.Type.Params.List {
		for _, name := range arg.Names {
			args = append(args, name.Name)
		}
	}
	// 返回值列表
	returns := make([]string, 0)
	returnRefs := make([]string, 0)
	returnNames := fun.Type.Results.List[0].Names
	if len(returnNames) == 0 {
		for i := 0; i < fun.Type.Results.NumFields(); i++ {
			fun.Type.Results.List[0].Names = append(fun.Type.Results.List[0].Names,
				&ast.Ident{Name: fmt.Sprintf("_xgo_res_%d", i+1)})
		}
	}
	for _, re := range fun.Type.Results.List[0].Names {
		returns = append(returns, re.Name)
		returnRefs = append(returnRefs, "&"+re.Name)
	}
	return fmt.Sprintf(newCodeFormat,
		funcName,
		strings.Join(args, ","),
		strings.Join(returnRefs, ","),
		strings.Join(returns, ","))
}

var newCodeFormat = `
package main

func TmpFunc() {
	if InterceptMock("%s", []interface{}{%s}, []interface{}{%s}) {
		return %s
	}
}
`
