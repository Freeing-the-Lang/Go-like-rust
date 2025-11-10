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

var functions = map[string]*Function{}

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

		// 함수 정의 블록 감지
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
	}
}

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
	obj := vars[name]
	runtime.Print(obj)
}

func parseIf(line string, vars map[string]*runtime.HeapObject) {
	// 예: if a > 5 { println!(a); }
	cond := strings.TrimPrefix(line, "if ")
	cond = strings.TrimSuffix(cond, "{")
	cond = strings.TrimSpace(cond)
	varName := strings.Split(cond, ">")[0]
	threshold := 0
	fmt.Sscanf(strings.Split(cond, ">")[1], "%d", &threshold)
	obj := vars[strings.TrimSpace(varName)]
	if obj.Data.(int) > threshold {
		fmt.Printf("[IF TRUE] %v > %d\n", obj.Data, threshold)
	}
}

func parseLoop(line string, vars map[string]*runtime.HeapObject) {
	// 예: loop 3 { println!(a); }
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
			localVars[argName] = args[i]
		}
		for _, bodyLine := range userFn.Body {
			execLine(bodyLine, localVars)
		}
		return
	}

	fmt.Printf("⚠️ Unknown function: %s\n", fnName)
}

func parseFnHeader(line string) *Function {
	// fn add(a, b) {
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
