package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/DataDog/gostackparse"
	"github.com/goccy/go-json"
)

func main() {
	stack := debug.Stack()
	fmt.Println(string(stack))
	goroutines, _ := gostackparse.Parse(bytes.NewReader(stack))
	json.NewEncoder(os.Stdout).Encode(goroutines)
}
