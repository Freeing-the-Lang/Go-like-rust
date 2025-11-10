package main

import (
	"fmt"
	"gcrust/parser"
)

func main() {
	fmt.Println("🦀 Go-like-Rust Script Runner")
	parser.RunScript("examples/hello.rs")
	fmt.Println("✅ Execution done")
}
