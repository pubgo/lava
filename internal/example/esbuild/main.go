package main

import (
	"fmt"
	"time"
)
import "github.com/evanw/esbuild/pkg/api"

func main() {
	var now = time.Now()
	defer func() { fmt.Println(time.Since(now)) }()

	result := api.Transform(`
import { Trend } from 'k6/metrics';

const myTrend = new Trend('my_trend');

export default function () {
  myTrend.add(1);
  myTrend.add(2);
}
`, api.TransformOptions{
		Loader: api.LoaderJS,
		Format: api.FormatCommonJS,
	})

	if len(result.Errors) == 0 {
		fmt.Printf("%s", result.Code)
	}
}
