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

	log.Println("run successfully")
}
