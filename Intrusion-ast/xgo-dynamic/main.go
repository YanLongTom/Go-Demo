package main

import "log"

func main() {
	res := Greet("world")
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
