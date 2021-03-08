package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pubgo/golug/metric"
	"github.com/pubgo/golug/metric/prometheus"
)

func main() {
	reporter, err := prometheus.NewReporter(
		metric.Path("/metrics"),
		metric.Address(":8089"),
		metric.Name("test"),
	)
	checkErr(err)

	defer func() {
		checkErr(reporter.Stop())
	}()
	_ = reporter.Start()

	metric.SetDefaultReporter(reporter)

	c := metric.NewCounter("count_req")
	g := metric.NewGauge("gauge")
	h := metric.NewHistogram("histogram")
	s := metric.NewSummary("summary")

	var tag = metric.Tags{"test": "123456"}

	for {
		checkErr(c.With(tag).Add(1))
		checkErr(g.Set(1))
		checkErr(h.With(tag).Observe(1))
		checkErr(s.With(tag).Observe(1))
		checkErr(metric.Count("count1", 2, metric.Tags{"t1": "1", "t2": "2"}))
		checkErr(metric.Summary("summary", 1, tag))

		time.Sleep(time.Second)

		resp, err := http.Get("http://localhost:8089/metrics")
		checkErr(err)
		dt, err := ioutil.ReadAll(resp.Body)
		checkErr(err)
		fmt.Println(string(dt))
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
