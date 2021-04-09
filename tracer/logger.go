package tracer

import jaegerLog "github.com/uber/jaeger-client-go/log"

var _ jaegerLog.Logger = (*logger)(nil)

type logger struct {

}

func (l logger) Error(msg string) {
	panic("implement me")
}

func (l logger) Infof(msg string, args ...interface{}) {
	panic("implement me")
}


