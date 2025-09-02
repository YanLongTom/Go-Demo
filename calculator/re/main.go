package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Calculator 经典双栈计算器
type Calculator struct {
	numStack []float64 // 数字栈
	opStack  []rune    // 操作符栈
}

// NewCalculator 创建计算器
func NewCalculator() *Calculator {
	return &Calculator{
		numStack: make([]float64, 0),
		opStack:  make([]rune, 0),
	}
}

// 获取操作符优先级
func priority(op rune) int {
	switch op {
	case '+', '-':
		return 1
	case '*', '/':
		return 2
	default:
		return 0
	}
}

// 压入数字
func (c *Calculator) pushNum(num float64) {
	c.numStack = append(c.numStack, num)
}

// 弹出数字
func (c *Calculator) popNum() float64 {
	if len(c.numStack) == 0 {
		return 0
	}
	num := c.numStack[len(c.numStack)-1]
	c.numStack = c.numStack[:len(c.numStack)-1]
	return num
}

// 压入操作符
func (c *Calculator) pushOp(op rune) {
	c.opStack = append(c.opStack, op)
}

// 弹出操作符
func (c *Calculator) popOp() rune {
	if len(c.opStack) == 0 {
		return 0
	}
	op := c.opStack[len(c.opStack)-1]
	c.opStack = c.opStack[:len(c.opStack)-1]
	return op
}

// 查看栈顶操作符
func (c *Calculator) topOp() rune {
	if len(c.opStack) == 0 {
		return 0
	}
	return c.opStack[len(c.opStack)-1]
}

// 执行运算
func (c *Calculator) calculate() {
	if len(c.numStack) < 2 || len(c.opStack) == 0 {
		return
	}

	b := c.popNum()
	a := c.popNum()
	op := c.popOp()

	var result float64
	switch op {
	case '+':
		result = a + b
	case '-':
		result = a - b
	case '*':
		result = a * b
	case '/':
		if b != 0 {
			result = a / b
		}
	}

	c.pushNum(result)
}

// 主计算函数
func (c *Calculator) Evaluate(expression string) float64 {
	// 清空栈
	c.numStack = c.numStack[:0]
	c.opStack = c.opStack[:0]

	// 去掉空格
	expr := strings.ReplaceAll(expression, " ", "")

	i := 0
	for i < len(expr) {
		char := rune(expr[i])

		// 处理数字（包括小数）
		if (char >= '0' && char <= '9') || char == '.' {
			numStr := ""
			for i < len(expr) && ((expr[i] >= '0' && expr[i] <= '9') || expr[i] == '.') {
				numStr += string(expr[i])
				i++
			}
			num, _ := strconv.ParseFloat(numStr, 64)
			c.pushNum(num)
			continue
		}

		// 处理负号
		if char == '-' && (i == 0 || expr[i-1] == '(' || expr[i-1] == '+' || expr[i-1] == '-' || expr[i-1] == '*' || expr[i-1] == '/') {
			i++
			numStr := "-"
			for i < len(expr) && ((expr[i] >= '0' && expr[i] <= '9') || expr[i] == '.') {
				numStr += string(expr[i])
				i++
			}
			num, _ := strconv.ParseFloat(numStr, 64)
			c.pushNum(num)
			continue
		}

		// 处理左括号
		if char == '(' {
			c.pushOp(char)
		} else if char == ')' {
			// 处理右括号：计算到左括号为止
			for len(c.opStack) > 0 && c.topOp() != '(' {
				c.calculate()
			}
			c.popOp() // 弹出左括号
		} else if char == '+' || char == '-' || char == '*' || char == '/' {
			// 处理操作符：如果栈顶操作符优先级 >= 当前操作符，先计算
			for len(c.opStack) > 0 && c.topOp() != '(' && priority(c.topOp()) >= priority(char) {
				c.calculate()
			}
			c.pushOp(char)
		}

		i++
	}

	// 计算剩余的操作符
	for len(c.opStack) > 0 {
		c.calculate()
	}

	// 返回结果
	if len(c.numStack) > 0 {
		return c.numStack[0]
	}
	return 0
}

func main() {
	calc := NewCalculator()

	fmt.Println("经典双栈计算器")
	fmt.Println("===============")

	// 测试用例
	tests := []string{
		"3+4*2",
		"(3+4)*-2",
		"10-2*3",
		"(10-2)*3",
		"2+3*(4-1)",
		"(2+3)*(4-1)",
		"10/2+3",
		"10/(2+3)",
		"-5+3",
		"(-5+3)*2",
		"3.5+2.5*2",
		"1+2*3+4",
		"(9-(1+2)*8)-(3+4)",
	}

	for _, test := range tests {
		result := calc.Evaluate(test)
		fmt.Printf("%s = %.2f\n", test, result)
	}

	// 交互模式
	fmt.Println("\n交互式计算器 (输入q退出):")
	for {
		fmt.Print("输入表达式: ")
		var input string
		fmt.Scanln(&input)

		if input == "q" || input == "quit" {
			break
		}

		if input != "" {
			result := calc.Evaluate(input)
			fmt.Printf("结果: %.6g\n\n", result)
		}
	}
}
