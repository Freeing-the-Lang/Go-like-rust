package runtime

import "fmt"

// ---------- Struct Model ----------

type StructDef struct {
	Name            string
	Fields          map[string]*HeapObject
	Impls           map[string]func(*HeapObject, ...*HeapObject) *HeapObject
	Implementations map[string]*TraitDef
}

var StructRegistry = map[string]*StructDef{}

func NewStruct(name string) *StructDef {
	s := &StructDef{
		Name:            name,
		Fields:          map[string]*HeapObject{},
		Impls:           map[string]func(*HeapObject, ...*HeapObject) *HeapObject{},
		Implementations: map[string]*TraitDef{},
	}
	StructRegistry[name] = s
	return s
}

func (s *StructDef) AddMethod(name string, f func(*HeapObject, ...*HeapObject) *HeapObject) {
	s.Impls[name] = f
}

func (s *StructDef) CallMethod(name string, self *HeapObject, args ...*HeapObject) *HeapObject {
	if fn, ok := s.Impls[name]; ok {
		return fn(self, args...)
	}
	fmt.Printf("⚠️ Method not found: %s on %s\n", name, s.Name)
	return nil
}

// ---------- Trait Model ----------

type TraitDef struct {
	Name  string
	Funcs map[string]func(*HeapObject, ...*HeapObject) *HeapObject
}

var TraitRegistry = map[string]*TraitDef{}

func NewTrait(name string) *TraitDef {
	t := &TraitDef{Name: name, Funcs: map[string]func(*HeapObject, ...*HeapObject) *HeapObject{}}
	TraitRegistry[name] = t
	return t
}

func (t *TraitDef) AddFunc(name string, f func(*HeapObject, ...*HeapObject) *HeapObject) {
	t.Funcs[name] = f
}

func ImplementTrait(structName, traitName string) {
	s := StructRegistry[structName]
	t := TraitRegistry[traitName]
	if s != nil && t != nil {
		s.Implementations[traitName] = t
		fmt.Printf("[IMPL] %s implements %s\n", structName, traitName)
	}
}

func CallTraitMethod(obj *HeapObject, traitName, fnName string, args ...*HeapObject) *HeapObject {
	if def, ok := StructRegistry[obj.Type]; ok {
		if t, ok2 := def.Implementations[traitName]; ok2 {
			if fn, ok3 := t.Funcs[fnName]; ok3 {
				return fn(obj, args...)
			}
		}
	}
	fmt.Printf("⚠️ Trait call failed: %s::%s\n", traitName, fnName)
	return nil
}
