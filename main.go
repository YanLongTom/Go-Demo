package main

import (
	"fmt"
	"time"
)

// test push
func main() {
	s := []int{1, 2, 3}
	for _, v := range s {
		go func() {
			fmt.Println(v)
		}()
	}
	time.Sleep(time.Second)
}
