package main

import (
	bloomfilter "Demo/bloomfilter/bloom"
	"fmt"
	"strconv"
	"time"
)

func main() {
	// 预计插入100万个元素，误判率0.01
	bf := bloomfilter.NewBloomFilter(1000000, 0.01)

	// 记录开始时间
	start := time.Now()

	// 插入100万个元素
	for i := 0; i < 1000000; i++ {
		bf.Add(strconv.Itoa(i))
	}

	// 计算插入耗时
	insertTime := time.Since(start)
	fmt.Printf("插入100万个元素耗时: %v\n", insertTime)

	// 测试误判率
	falsePositiveCount := 0
	testCount := 100000

	start = time.Now()
	for i := 1000000; i < 1000000+testCount; i++ {
		if bf.MightContain(strconv.Itoa(i)) {
			falsePositiveCount++
		}
	}

	// 计算查询耗时
	queryTime := time.Since(start)
	fmt.Printf("查询%d个元素耗时: %v\n", testCount, queryTime)
	fmt.Printf("误判数: %d, 误判率: %.4f%%\n",
		falsePositiveCount, float64(falsePositiveCount)/float64(testCount)*100)
}
