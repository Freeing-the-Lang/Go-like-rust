package runtime

import "fmt"

// 함수 호출 규약
func Call(fn func(...*HeapObject) *HeapObject, args ...*HeapObject) *HeapObject {
	return fn(args...)
}

// 내장 Add
func Add(args ...*HeapObject) *HeapObject {
	a := args[0].Data.(int)
	b := args[1].Data.(int)
	return Alloc("int", a+b)
}

// Print 함수 (Rust의 println! 매핑)
func Print(o *HeapObject) {
	fmt.Printf("[PRINT] %v (%s)\n", o.Data, o.Type)
}
