package main

import (
	"fmt"
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
				f, err := os.Create("tmp.go")
				if err != nil {
					log.Fatalf("create tmp.go error: %v\n", err)
				}
				defer f.Close()
				defer os.Remove("tmp.go")
				_, _ = f.WriteString(newCode)
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

var newCode = `
package main

func Greet(s string) (res string) {
	if InterceptMock("Greet", s, &res) {
	 	return res
    }
	return "hello " + s
}
`
