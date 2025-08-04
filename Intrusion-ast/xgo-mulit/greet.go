package main

import "fmt"

func Greet(s string) (res string) {
	return "hello " + s
}

func Greet2(s2 string) (res2 string) {
	return "hello 2 " + s2
}

func Greet3(s3 string) string {
	return "hello 3 " + s3
}

func Pair1(s1, s2 string) (res string) {
	return "pair 1 " + s1 + " " + s2
}

func Pair2(s1, s2 string) (res1, res2 string) {
	return "pair 1 " + s1, "pair 2 " + s2
}

func Other(i int, s string, f float64) string {
	return fmt.Sprintf("int: %d, string: %s, float: %f", i, s, f)
}
