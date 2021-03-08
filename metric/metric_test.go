package metric

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pubgo/golug/metric/prometheus"
	"github.com/pubgo/x/fx"
)

func TestMetrics(t *testing.T) {
	report, err := prometheus.NewReporter(
		Path("/metrics"),
		Address(":8089"),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = report.Stop()
	}()
	_ = report.Start()

	SetDefaultReporter(report)

	done := make(chan error)
	_ = fx.Go(func(ctx context.Context) {
		time.Sleep(1 * time.Second)

		r, err := http.NewRequest(http.MethodGet, "http://localhost:8089/metrics", nil)
		if err != nil {
			done <- err
			return
		}

		c := http.Client{}
		rsp, err := c.Do(r)
		if err != nil {
			done <- err
			return
		}

		if rsp.Body != nil {
			defer rsp.Body.Close()
		}

		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			done <- err
			return
		}

		// 检测一个进程指标 跟 检测一个内存指标
		if !strings.Contains(string(body), "go_goroutines") || !strings.Contains(string(body), "go_info") {
			done <- errors.New("指标信息不存在")
		}

		done <- nil
	})

	if err := <-done; err != nil {
		fmt.Println(err.Error())
		t.Error(err)
	}
}
