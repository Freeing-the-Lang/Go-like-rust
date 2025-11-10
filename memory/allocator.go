package memory

import (
	"Go-like-rust/runtime/gc"
)

type Allocator struct {
	gc *gc.GC
}

func NewAllocator(g *gc.GC) *Allocator {
	return &Allocator{gc: g}
}

func (a *Allocator) Box(v interface{}) *gc.Object {
	return a.gc.Allocate(v)
}

func (a *Allocator) Vec(items ...interface{}) []*gc.Object {
	vec := make([]*gc.Object, 0, len(items))
	for _, it := range items {
		vec = append(vec, a.gc.Allocate(it))
	}
	return vec
}
