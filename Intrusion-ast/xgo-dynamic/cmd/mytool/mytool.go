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

// 解析源文件 greet.go，并插入自定义代码。 tmp 文件需要重命名，因为最后有remove，防止删除掉上层tmp文件
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

		f, err := os.Create("tmp1.go")
		if err != nil {
			log.Fatalf("create tmp.go error: %v\n", err)
		}
		_, _ = f.WriteString(newCode(fun))
		f.Close()

		tmpFset := token.NewFileSet()
		tmpF, err := parser.ParseFile(tmpFset, "tmp1.go", nil, parser.AllErrors)
		if err != nil {
			log.Fatalf("parse tmp.go error: %v\n", err)
		}
		fun.Body.List = append(tmpF.Decls[0].(*ast.FuncDecl).Body.List, fun.Body.List...)
		os.Remove("tmp1.go")
	}

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, fast)

	fmt.Println(buf.String())

	return buf.String()
}

// 拼接新代码
func newCode(fun *ast.FuncDecl) string {
	/*
		&{Doc:<nil> Names:[s] Type:string Tag:<nil> Comment:<nil>}
		&{Doc:<nil> Names:[res] Type:string Tag:<nil> Comment:<nil>}
		&{Doc:<nil> Names:[s2] Type:string Tag:<nil> Comment:<nil>}
		&{Doc:<nil> Names:[res2] Type:string Tag:<nil> Comment:<nil>}
		&{Doc:<nil> Names:[s3] Type:string Tag:<nil> Comment:<nil>}
		&{Doc:<nil> Names:[] Type:string Tag:<nil> Comment:<nil>}
	*/
	// 函数名称
	funcName := fun.Name.Name

	// 参数列表
	argName := fun.Type.Params.List[0].Names[0].Name

	// 返回值列表
	resNames := fun.Type.Results.List[0].Names
	if len(resNames) == 0 {
		resNames = append(resNames, &ast.Ident{Name: "_xgo_res_1"})
		fun.Type.Results.List[0].Names = resNames
	}
	resName := resNames[0].Name
	return fmt.Sprintf(newCodeFormat, funcName, argName, resName, resName)
}

// 字符串拼接准备，拼接为自定义的代码块
var newCodeFormat = `
package main

func TmpFunc() {
	if InterceptMock("%s", %s, &%s) {
	 	return %s
    }
}
`

//// 重写代码
//func insertCode(filename string) string {
//	fset := token.NewFileSet()
//	fast, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
//	if err != nil {
//		log.Fatalf("parse file error: %v\n", err)
//	}
//
//	for _, decl := range fast.Decls {
//		fun, ok := decl.(*ast.FuncDecl)
//		if !ok {
//			continue
//		}
//
//		f, err := os.Create("tmp2.go")
//		if err != nil {
//			log.Fatalf("create tmp2.go error: %v\n", err)
//		}
//		_, _ = f.WriteString(fmt.Sprintf(newCodeFormat, fun.Name.Name))
//		f.Close()
//
//		tmpFset := token.NewFileSet()
//		tmpF, err := parser.ParseFile(tmpFset, "tmp2.go", nil, parser.AllErrors)
//		if err != nil {
//			log.Fatalf("parse tmp2.go error: %v\n", err)
//		}
//		// 重写源函数
//		fun.Body.List = append(tmpF.Decls[0].(*ast.FuncDecl).Body.List, fun.Body.List...)
//		os.Remove("tmp2.go")
//	}
//
//	var buf bytes.Buffer
//	printer.Fprint(&buf, fset, fast)
//
//	fmt.Println(buf.String())
//
//	return buf.String()
//}
//
//var newCodeFormat = `
//package main
//
//func TmpFunc() {
//	if InterceptMock("%s", s, &res) {
//	 	return res
//    }
//}
//`
