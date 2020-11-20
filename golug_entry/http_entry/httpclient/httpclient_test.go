package httpclient_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/atomic"

	"github.com/pubgo/golug/golug_entry/http_entry/httpclient"
	"github.com/smartystreets/goconvey/convey"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	const URL = "http://nexus.wpt.la/repository/wpt-raw-local/micro-gen/.version"

	convey.Convey("Test HTTP Client", t, func() {
		convey.Convey("Case: Normal", func() {
			client := httpclient.New()
			rsp, err := client.Get(URL, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(rsp, convey.ShouldNotBeNil)
			convey.So(rsp.Body, convey.ShouldNotBeNil)
			convey.So(rsp.StatusCode, convey.ShouldEqual, http.StatusOK)
			b, err := ioutil.ReadAll(rsp.Body)
			m := make(map[string]string)
			err = json.Unmarshal(b, &m)
			convey.So(err, convey.ShouldBeNil)
			convey.So(m["version"], convey.ShouldNotBeNil)
		})

		convey.Convey("Case: Not Found", func() {
			client := httpclient.New()
			rsp, err := client.Get(URL+"1", nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(rsp, convey.ShouldNotBeNil)
			convey.So(rsp.Body, convey.ShouldNotBeNil)
			convey.So(rsp.StatusCode, convey.ShouldEqual, http.StatusNotFound)
		})

		convey.Convey("Case: Timeout", func() {
			client := httpclient.New(httpclient.WithHTTPTimeout(time.Microsecond))
			rsp, err := client.Get(URL, nil)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(rsp, convey.ShouldBeNil)
			convey.ShouldContain(err.Error(), "Timeout")
		})

		convey.Convey("Case: Retry Count 3", func() {
			var c atomic.Int32
			client := httpclient.New(
				httpclient.WithHTTPTimeout(time.Microsecond),
				httpclient.WithRetryCount(3),
				httpclient.WithMiddleware(func(doFunc httpclient.DoFunc) httpclient.DoFunc {
					return func(request *http.Request, f func(*http.Response) error) error {
						c.Inc()
						return doFunc(request, f)
					}
				}),
			)
			rsp, err := client.Get(URL, nil)
			convey.So(c.Load(), convey.ShouldEqual, 3)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(rsp, convey.ShouldBeNil)
			convey.So(strings.Contains(err.Error(), "Timeout"), convey.ShouldBeTrue)
		})

		convey.Convey("Case: Retrier", func() {
			var c atomic.Int32
			client := httpclient.New(
				httpclient.WithHTTPTimeout(time.Microsecond),
				httpclient.WithRetrier(httpclient.NewRetrier(httpclient.NewConstantBackoff(10*time.Millisecond, 50*time.Millisecond))),
				httpclient.WithRetryCount(3),
				httpclient.WithMiddleware(func(doFunc httpclient.DoFunc) httpclient.DoFunc {
					return func(request *http.Request, f func(*http.Response) error) error {
						c.Inc()
						return doFunc(request, f)
					}
				}),
			)
			rsp, err := client.Get(URL, nil)
			convey.So(c.Load(), convey.ShouldEqual, 3)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(rsp, convey.ShouldBeNil)
			convey.ShouldContain(err.Error(), "Timeout")
		})
	})
}
