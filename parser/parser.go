package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gcrust/runtime"
)

type Function struct {
	Name string
	Args []string
	Body []string
}

var (
	functions       = map[string]*Function{}
	currentStruct   *runtime.StructDef
	currentTrait    *runtime.TraitDef
	implTraitTarget = struct {
		Trait  string
		Struct string
	}{}
)

// -------------------- Entry --------------------

func RunScript(path string) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	vars := map[string]*runtime.HeapObject{}
	scanner := bufio.NewScanner(file)
	var fn *Function

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// trait 정의
		if strings.HasPrefix(line, "trait ") {
			name := strings.TrimSpace(strings.TrimPrefix(line, "trait "))
			currentTrait = runtime.NewTrait(name)
			continue
		}
		if currentTrait != nil {
			if line == "}" {
				currentTrait = nil
				continue
			}
			if strings.HasPrefix(line, "fn ") {
				fnName := strings.TrimSpace(strings.Split(strings.TrimPrefix(line, "fn "), "(")[0])
				currentTrait.AddFunc(fnName, func(self *runtime.HeapObject, args ...*runtime.HeapObject) *runtime.HeapObject {
					fmt.Printf("[TRAIT] %s::%s called on %s (no impl)\n", currentTrait.Name, fnName, self.Type)
					return runtime.Alloc("nil", nil)
				})
			}
			continue
		}

		// struct 정의
		if strings.HasPrefix(line, "struct ") {
			name := strings.TrimSpace(strings.TrimPrefix(line, "struct "))
			currentStruct = runtime.NewStruct(name)
			continue
		}
		if currentStruct != nil && line == "}" {
			currentStruct = nil
			continue
		}

		// impl Trait for Struct
		if strings.HasPrefix(line, "impl ") && strings.Contains(line, " for ") {
			parts := strings.Split(line, "for")
			trait := strings.TrimSpace(strings.TrimPrefix(parts[0], "impl "))
			strct := strings.TrimSpace(strings.TrimSuffix(parts[1], "{"))
			implTraitTarget = struct {
				Trait  string
				Struct string
			}{trait, strct}
			runtime.ImplementTrait(strct, trait)
			continue
		}

		// impl Struct (일반 impl)
		if strings.HasPrefix(line, "impl ") && !strings.Contains(line, "for ") {
			structName := strings.TrimSpace(strings.TrimPrefix(line, "impl "))
			currentStruct = runtime.StructRegistry[structName]
			continue
		}
		if (currentStruct != nil || implTraitTarget.Trait != "") && line == "}" {
			currentStruct = nil
			implTraitTarget = struct{ Trait, Struct string }{}
			continue
		}

		// impl 내부의 메서드
		if strings.HasPrefix(line, "fn ") {
			fnName := strings.TrimSpace(strings.Split(strings.TrimPrefix(line, "fn "), "(")[0])
			handler := func(self *runtime.HeapObject, args ...*runtime.HeapObject) *runtime.HeapObject {
				fmt.Printf("[METHOD] %s.%s() called on %s\n", currentStruct.Name, fnName, self.Type)
				return runtime.Alloc("nil", nil)
			}

			if currentStruct != nil {
				currentStruct.AddMethod(fnName, handler)
			} else if implTraitTarget.Trait != "" {
				if t := runtime.TraitRegistry[implTraitTarget.Trait]; t != nil {
					t.AddFunc(fnName, handler)
				}
			}
			continue
		}

		// fn 정의
		if strings.HasPrefix(line, "fn ") {
			fn = parseFnHeader(line)
			continue
		}
		if fn != nil {
			if line == "}" {
				functions[fn.Name] = fn
				fn = nil
			} else {
				fn.Body = append(fn.Body, line)
			}
			continue
		}

		execLine(line, vars)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

// -------------------- Statement Execution --------------------

func execLine(line string, vars map[string]*runtime.HeapObject) {
	switch {
	case strings.HasPrefix(line, "let "):
		parseLet(line, vars)
	case strings.HasPrefix(line, "println!"):
		parsePrint(line, vars)
	case strings.HasPrefix(line, "if "):
		parseIf(line, vars)
	case strings.HasPrefix(line, "loop"):
		parseLoop(line, vars)
	case strings.Contains(line, "=") && strings.Contains(line, "("):
		parseCall(line, vars)
	default:
		fmt.Printf("⚠️ Unrecognized line: %s\n", line)
	}
}

// -------------------- Core Parsers --------------------

func parseLet(line string, vars map[string]*runtime.HeapObject) {
	parts := strings.Split(line, "=")
	left := strings.TrimSpace(strings.TrimPrefix(parts[0], "let"))
	right := strings.TrimSpace(strings.TrimSuffix(parts[1], ";"))
	val := 0
	fmt.Sscanf(right, "%d", &val)
	vars[left] = runtime.Alloc("int", val)
}

func parsePrint(line string, vars map[string]*runtime.HeapObject) {
	start := strings.Index(line, "(")
	end := strings.Index(line, ")")
	name := strings.TrimSpace(line[start+1 : end])
	obj, ok := vars[name]
	if !ok {
		fmt.Printf("⚠️ Undefined variable: %s\n", name)
		return
	}
	runtime.Print(obj)
}

func parseIf(line string, vars map[string]*runtime.HeapObject) {
	cond := strings.TrimPrefix(line, "if ")
	cond = strings.TrimSuffix(cond, "{")
	varName := strings.Split(cond, ">")[0]
	threshold := 0
	fmt.Sscanf(strings.Split(cond, ">")[1], "%d", &threshold)
	obj := vars[strings.TrimSpace(varName)]
	if obj.Data.(int) > threshold {
		fmt.Printf("[IF TRUE] %v > %d\n", obj.Data, threshold)
	}
}

func parseLoop(line string, vars map[string]*runtime.HeapObject) {
	line = strings.TrimPrefix(line, "loop")
	count := 0
	fmt.Sscanf(line, "%d", &count)
	for i := 0; i < count; i++ {
		fmt.Printf("[LOOP] Iteration %d\n", i+1)
	}
}

func parseCall(line string, vars map[string]*runtime.HeapObject) {
	parts := strings.Split(line, "=")
	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(strings.TrimSuffix(parts[1], ";"))
	fnName := right[:strings.Index(right, "(")]
	argsRaw := right[strings.Index(right, "(")+1 : strings.Index(right, ")")]
	argNames := strings.Split(argsRaw, ",")

	var args []*runtime.HeapObject
	for _, name := range argNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		args = append(args, vars[name])
	}

	if builtin := runtime.GetBuiltin(fnName); builtin != nil {
		vars[left] = runtime.Call(builtin, args...)
		return
	}

	if userFn, ok := functions[fnName]; ok {
		fmt.Printf("[CALL] User fn %s\n", userFn.Name)
		localVars := map[string]*runtime.HeapObject{}
		for i, argName := range userFn.Args {
			if i < len(args) {
				localVars[argName] = args[i]
			}
		}
		for _, bodyLine := range userFn.Body {
			execLine(bodyLine, localVars)
		}
		return
	}

	fmt.Printf("⚠️ Unknown function: %s\n", fnName)
}

func parseFnHeader(line string) *Function {
	parts := strings.Split(line, "(")
	name := strings.TrimSpace(strings.TrimPrefix(parts[0], "fn "))
	argsStr := strings.TrimSuffix(strings.Split(parts[1], ")")[0], "{")
	args := []string{}
	for _, arg := range strings.Split(argsStr, ",") {
		arg = strings.TrimSpace(arg)
		if arg != "" {
			args = append(args, arg)
		}
	}
	return &Function{Name: name, Args: args, Body: []string{}}
}
