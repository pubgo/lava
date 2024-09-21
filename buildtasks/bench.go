package buildtasks

import "fmt"

type BenchMode int

const (
	BenchModeWazero BenchMode = iota
	BenchModeCGO
	BenchModeDefault
)

func BenchArgs(pkg string, count int, mode BenchMode, libName string) []string {
	args := []string{"test", "-bench=.", "-run=^$", "-v", "-timeout=60m"}
	if count > 0 {
		args = append(args, fmt.Sprintf("-count=%d", count))
	}
	switch mode {
	case BenchModeCGO:
		args = append(args, fmt.Sprintf("-tags=%s_cgo", libName))
	case BenchModeDefault:
		args = append(args, fmt.Sprintf("-tags=%s_bench_default", libName))
	}
	args = append(args, pkg)

	return args
}
