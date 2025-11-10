package runtime

import (
	"fmt"
	"sync"
)

type HeapObject struct {
	Type string
	Data any
}

var heap = sync.Pool{
	New: func() any { return &HeapObject{} },
}

func Alloc(t string, v any) *HeapObject {
	obj := heap.Get().(*HeapObject)
	obj.Type = t
	obj.Data = v
	return obj
}

func Release(o *HeapObject) {
	// Go GC가 자동 처리하므로 명시 해제 불필요
	heap.Put(o)
}
