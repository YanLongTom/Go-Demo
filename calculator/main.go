package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Calculator struct {
	op  []rune
	num []float64
}

// 数字
func (c *Calculator) pushNum(value float64) {
	c.num = append(c.num, value)
}

func (c *Calculator) popNum() float64 {
	ret := c.num[len(c.num)-1]
	c.num = c.num[:len(c.num)-1]
	return ret
}

// 操作
func (c *Calculator) pushOP(value rune) {
	c.op = append(c.op, value)
}

func (c *Calculator) popOP() rune {
	ret := c.op[len(c.op)-1]
	c.op = c.op[:len(c.op)-1]
	return ret
}

func newCalculator() *Calculator {
	return &Calculator{
		op:  make([]rune, 0),
		num: make([]float64, 0),
	}
}

// 计算结果压栈：后进先出
func (c *Calculator) calculator() {
	a := c.popNum()
	b := c.popNum()
	op := c.popOP()
	var ret float64
	switch op {
	case '+':
		ret = a + b
	case '-':
		ret = a - b
	case '*':
		ret = a * b
	case '/':
		ret = a / b
	}
	c.pushNum(ret)
}
func (c *Calculator) Evaluate(value string) {
	i := 0
	expr := strings.ReplaceAll(value, " ", "")
	for i < len(expr) {
		char := rune(expr[i])
		if char < '9' && char > '0' {
			ret := string(expr[i])
			for i < len(expr) && (expr[i] < '9' && expr[i] > '0') {
				ret += string(expr[i])
				i++
			}
			v, _ := strconv.ParseFloat(ret, 64)
			c.pushNum(v)
			continue
		} else {
			c.pushOP(char)
			i++
		}
	}
	for i := 0; i < len(c.op); i++ {
		c.calculator()
	}
}

func main() {
	a := "12+54*123/43"
	c := newCalculator()
	c.Evaluate(a)
	fmt.Println(c.popNum())
}
