package main

import (
	"fmt"
	"log"
)

func main() {

	RegisterMockFunc("Other", func(i int, s string, f float64) string {
		return fmt.Sprintf("mock %d %s %.2f", i, s, f)
	})
	res := Other(1, "hello", 3.14)
	if res != "mock 1 hello 3.14" {
		log.Fatalf("Other() = %q; want %q", res, "mock 1 hello 3.14")
	}
	log.Println("run other successfully")

	RegisterMockFunc("Pair1", func(s1, s2 string) string {
		return "mock 1 " + s1 + " " + s2
	})
	res = Pair1("hello", "world")
	if res != "mock 1 hello world" {
		log.Fatalf("Pair1() = %q; want %q", res, "mock 1 hello world")
	}
	log.Println("run pair1 successfully")

	RegisterMockFunc("Pair2", func(s1, s2 string) (string, string) {
		return "mock 2 " + s1, "mock 2 " + s2
	})
	res1, res2 := Pair2("hello", "world")
	if res1 != "mock 2 hello" || res2 != "mock 2 world" {
		log.Fatalf("Pair2() = %q, %q; want %q, %q", res1, res2, "mock 2 hello", "mock 2 world")
	}
	log.Println("run pair2 successfully")

	res = Greet("world")
	if res != "hello world" {
		log.Fatalf("Greet() = %q; want %q", res, "hello world")
	}

	RegisterMockFunc("Greet", func(s string) string {
		return "mock " + s
	})
	res = Greet("world")
	if res != "mock world" {
		log.Fatalf("Greet() = %q; want %q", res, "mock world")
	}

	log.Println("run greet 1 successfully")

	RegisterMockFunc("Greet2", func(s string) string {
		return "mock 2 " + s
	})
	res = Greet2("world")
	if res != "mock 2 world" {
		log.Fatalf("Greet2() = %q; want %q", res, "mock 2 world")
	}

	log.Println("run greet 2 successfully")

	RegisterMockFunc("Greet3", func(s string) string {
		return "mock 3 " + s
	})
	res = Greet3("world")
	if res != "mock 3 world" {
		log.Fatalf("Greet3() = %q; want %q", res, "mock 3 world")
	}

	log.Println("run greet 3 successfully")
}
