package main

import (
	"fmt"
	"time"
)
import "github.com/evanw/esbuild/pkg/api"

func main() {
	var now = time.Now()
	defer func() { fmt.Println(time.Since(now)) }()

	result := api.Transform("let x: number = 1", api.TransformOptions{
		Loader: api.LoaderTS,
	})

	if len(result.Errors) == 0 {
		fmt.Printf("%s", result.Code)
	}
}
