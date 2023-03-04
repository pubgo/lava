package process

import (
	"bufio"
	"debug/buildinfo"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"runtime"
	rd "runtime/debug"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/gops/signal"
	ps "github.com/keybase/go-ps"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/result"

	"github.com/pubgo/lava/core/debug"
)

func init() {
	debug.Get("/process", func(ctx *fiber.Ctx) error {
		processes := assert.Must1(ps.Processes())
		return ctx.JSON(generic.Map(processes, func(i int) map[string]any {
			var p = processes[i]
			return map[string]any{
				"pid":        p.Pid(),
				"ppid":       p.PPid(),
				"exec":       p.Executable(),
				"path":       result.Wrap(p.Path()),
				"go_version": goVersion(result.Wrap(p.Path())),
			}
		}))
	})
}

func goVersion(path result.Result[string]) result.Result[string] {
	if path.IsErr() {
		return path
	}

	info, err := buildinfo.ReadFile(path.Unwrap())
	if err != nil {
		return result.Wrap("", err)
	}
	return result.OK(info.GoVersion)
}

func handle(conn io.ReadWriter, msg []byte) error {
	switch msg[0] {
	case signal.StackTrace:
		return pprof.Lookup("goroutine").WriteTo(conn, 2)
	case signal.GC:
		runtime.GC()
		_, err := conn.Write([]byte("ok"))
		return err
	case signal.MemStats:
		var s runtime.MemStats
		runtime.ReadMemStats(&s)
		fmt.Fprintf(conn, "alloc: %v\n", formatBytes(s.Alloc))
		fmt.Fprintf(conn, "total-alloc: %v\n", formatBytes(s.TotalAlloc))
		fmt.Fprintf(conn, "sys: %v\n", formatBytes(s.Sys))
		fmt.Fprintf(conn, "lookups: %v\n", s.Lookups)
		fmt.Fprintf(conn, "mallocs: %v\n", s.Mallocs)
		fmt.Fprintf(conn, "frees: %v\n", s.Frees)
		fmt.Fprintf(conn, "heap-alloc: %v\n", formatBytes(s.HeapAlloc))
		fmt.Fprintf(conn, "heap-sys: %v\n", formatBytes(s.HeapSys))
		fmt.Fprintf(conn, "heap-idle: %v\n", formatBytes(s.HeapIdle))
		fmt.Fprintf(conn, "heap-in-use: %v\n", formatBytes(s.HeapInuse))
		fmt.Fprintf(conn, "heap-released: %v\n", formatBytes(s.HeapReleased))
		fmt.Fprintf(conn, "heap-objects: %v\n", s.HeapObjects)
		fmt.Fprintf(conn, "stack-in-use: %v\n", formatBytes(s.StackInuse))
		fmt.Fprintf(conn, "stack-sys: %v\n", formatBytes(s.StackSys))
		fmt.Fprintf(conn, "stack-mspan-inuse: %v\n", formatBytes(s.MSpanInuse))
		fmt.Fprintf(conn, "stack-mspan-sys: %v\n", formatBytes(s.MSpanSys))
		fmt.Fprintf(conn, "stack-mcache-inuse: %v\n", formatBytes(s.MCacheInuse))
		fmt.Fprintf(conn, "stack-mcache-sys: %v\n", formatBytes(s.MCacheSys))
		fmt.Fprintf(conn, "other-sys: %v\n", formatBytes(s.OtherSys))
		fmt.Fprintf(conn, "gc-sys: %v\n", formatBytes(s.GCSys))
		fmt.Fprintf(conn, "next-gc: when heap-alloc >= %v\n", formatBytes(s.NextGC))
		lastGC := "-"
		if s.LastGC != 0 {
			lastGC = fmt.Sprint(time.Unix(0, int64(s.LastGC)))
		}
		fmt.Fprintf(conn, "last-gc: %v\n", lastGC)
		fmt.Fprintf(conn, "gc-pause-total: %v\n", time.Duration(s.PauseTotalNs))
		fmt.Fprintf(conn, "gc-pause: %v\n", s.PauseNs[(s.NumGC+255)%256])
		fmt.Fprintf(conn, "gc-pause-end: %v\n", s.PauseEnd[(s.NumGC+255)%256])
		fmt.Fprintf(conn, "num-gc: %v\n", s.NumGC)
		fmt.Fprintf(conn, "num-forced-gc: %v\n", s.NumForcedGC)
		fmt.Fprintf(conn, "gc-cpu-fraction: %v\n", s.GCCPUFraction)
		fmt.Fprintf(conn, "enable-gc: %v\n", s.EnableGC)
		fmt.Fprintf(conn, "debug-gc: %v\n", s.DebugGC)
	case signal.Version:
		fmt.Fprintf(conn, "%v\n", runtime.Version())
	case signal.HeapProfile:
		return pprof.WriteHeapProfile(conn)
	case signal.CPUProfile:
		if err := pprof.StartCPUProfile(conn); err != nil {
			return err
		}
		time.Sleep(30 * time.Second)
		pprof.StopCPUProfile()
	case signal.Stats:
		fmt.Fprintf(conn, "goroutines: %v\n", runtime.NumGoroutine())
		fmt.Fprintf(conn, "OS threads: %v\n", pprof.Lookup("threadcreate").Count())
		fmt.Fprintf(conn, "GOMAXPROCS: %v\n", runtime.GOMAXPROCS(0))
		fmt.Fprintf(conn, "num CPU: %v\n", runtime.NumCPU())
	case signal.BinaryDump:
		path, err := os.Executable()
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = bufio.NewReader(f).WriteTo(conn)
		return err
	case signal.Trace:
		if err := trace.Start(conn); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
		trace.Stop()
	case signal.SetGCPercent:
		perc, err := binary.ReadVarint(bufio.NewReader(conn))
		if err != nil {
			return err
		}
		fmt.Fprintf(conn, "New GC percent set to %v. Previous value was %v.\n", perc, rd.SetGCPercent(int(perc)))
	}
	return nil
}

var units = []string{" bytes", "KB", "MB", "GB", "TB", "PB"}

func formatBytes(val uint64) string {
	var i int
	var target uint64
	for i = range units {
		target = 1 << uint(10*(i+1))
		if val < target {
			break
		}
	}
	if i > 0 {
		return fmt.Sprintf("%0.2f%s (%d bytes)", float64(val)/(float64(target)/1024), units[i], val)
	}
	return fmt.Sprintf("%d bytes", val)
}
