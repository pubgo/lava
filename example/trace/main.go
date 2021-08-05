package main

import (
	"fmt"
	"runtime"
	"time"

	"golang.org/x/net/trace"
	"google.golang.org/grpc/grpclog"
)

func startTrace() {
	_, file, line, _ := runtime.Caller(1) //  runtime库的Caller函数，可以返回运行时正在执行的文件名和行号
	events := trace.NewEventLog("grpc.Srv", fmt.Sprintf("%s:%d", file, line))

	for i := 0; ; i++ {
		tr := trace.New("grpc.trace11", fmt.Sprintf("%s:%d", file, line))
		tr.SetMaxEvents(100)
		tr.SetError()
		tr.LazyLog(stringer(fmt.Sprintf("heh  %d", i)), true)
		tr.LazyPrintf("test %s", i)
		tr.SetTraceInfo(uint64(i), uint64(i+5))
		events.Printf("ssss %d", i)
		grpclog.Info("Trace listen on 50051")
		time.Sleep(time.Second)
		tr.Finish()
	}
}

type stringer string

func (s stringer) String() string { return string(s) }
