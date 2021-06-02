package pprofplus

import (
	"github.com/shirou/gopsutil/v3/process"

	"os"
	"runtime"
	"time"
)

type capture struct {
	captureIntervalSec int
}

func NewCapture(captureIntervalSec int) *capture {
	return &capture{
		captureIntervalSec: captureIntervalSec,
	}
}

func (c *capture) doAsync() chan Info {
	ret := make(chan Info)
	go func() {
		p := process.Process{
			Pid: int32(os.Getpid()),
		}

		ret <- c.do(p)

		t := time.Tick(time.Second * time.Duration(c.captureIntervalSec))
		for {
			select {
			case <-t:
				ret <- c.do(p)
			}
		}
	}()
	return ret
}

func (c *capture) do(p process.Process) Info {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	mis, _ := p.MemoryInfo()

	info := Info{
		Timestamp:    time.Now().Unix(),
		Sys:          ms.Sys,
		HeapSys:      ms.HeapSys,
		HeapAlloc:    ms.HeapAlloc,
		HeapInuse:    ms.HeapInuse,
		HeapReleased: ms.HeapReleased,
		HeapIdle:     ms.HeapIdle,
		VMS:          mis.VMS,
		RSS:          mis.RSS,
	}
	return info
}

type Info struct {
	Timestamp int64

	Sys          uint64 `json:"sys"`
	HeapSys      uint64 `json:"heapsys"`
	HeapAlloc    uint64 `json:"heapalloc"`
	HeapInuse    uint64 `json:"heapinuse"`
	HeapReleased uint64 `json:"heapreleased"`
	HeapIdle     uint64 `json:"heapidle"`

	VMS uint64 `json:"vms"`
	RSS uint64 `json:"rss"`
}
