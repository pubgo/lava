package lua

import (
	"context"
	"hash/fnv"
	"testing"

	"github.com/kelindar/lua"
	"github.com/pubgo/xerror"
)

func TestName(t *testing.T) {
	s, err := lua.FromString("test.lua", `
    function main(n)
        if n < 2 then return 1 end
        return main(n - 2) + main(n - 1)
    end
`)
	xerror.Panic(err)

	// Run the main() function with 10 as argument
	result, err := s.Run(context.Background(), 10)
	println(result.String()) // Output: 89
}

func TestName1(t *testing.T) {
	type Person struct {
		Name string
	}

	// Load the script
	s, err := lua.FromString("test.lua", `
    function main(input)
        input.Name = "Updated"
        return input.Name
    end
`)
	xerror.Panic(err)

	input := &Person{Name: "Roman"}
	out, err := s.Run(context.Background(), input)
	println(out.String()) // Outputs: "Updated"
	println(input.Name)   // Outputs: "Updated"
}

func hash(s lua.String) (lua.Number, error) {
	h := fnv.New32a()
	h.Write([]byte(s))

	return lua.Number(h.Sum32()), nil
}

func TestNameFunc(t *testing.T) {
	// Create a test module which provides hash function
	module := &lua.NativeModule{
		Name:    "test",
		Version: "1.0.0",
	}
	module.Register("hash", hash)

	// Load the script
	s, err := lua.FromString("test.lua", `
    local api = require("test")

    function main(input)
        return api.hash(input)
    end
`, module) // <- attach the module
	xerror.Panic(err)

	out, err := s.Run(context.Background(), "abcdef")
	println(out.String()) // Output: 4282878506
}

func TestModules(t *testing.T) {
	moduleCode, err := lua.FromString("module.lua", `
    local demo_mod = {} -- The main table

    function demo_mod.Mult(a, b)
        return a * b
    end

    return demo_mod
`)
	xerror.Panic(err)

	// Create a test module which provides hash function
	module := &lua.ScriptModule{
		Script:  moduleCode,
		Name:    "demo_mod",
		Version: "1.0.0",
	}

	// Load the script
	s, err := lua.FromString("test.lua", `
    local demo = require("demo_mod")

    function main(input)
        return demo.Mult(5, 5)
    end
`, module) // <- attach the module

	out, err := s.Run(context.Background())
	println(out.String()) // Output: 25
}
