package main

import (
	"fmt"
	"github.com/bigwhite/functrace"
	"runtime"

	"sync"
)

func A1() {
	defer functrace.Trace()()



	B1()
}

func B1() {
	defer functrace.Trace()()
	C1()
}

func C1() {
	defer functrace.Trace()()
	D()
}

func A2() {
	//defer functrace.Trace()()
	B2()
}

func B2() {
	//defer functrace.Trace()()
	C2()
}

func C2() {
	defer functrace.Trace()()
	D()
}

func D() {
	pc, _, _, ok := runtime.Caller(0)
	if !ok {
		panic("not found caller")
	}

	fn := runtime.FuncForPC(pc)
	fmt.Println(fn.Name())

	defer functrace.Trace()()
}

func main() {
	defer functrace.Trace()()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		A1()
		wg.Done()
	}()

	A2()
	wg.Wait()
}