package reporter

import (
	"io"
	"time"

	"github.com/pubgo/funk/log"
	"github.com/uber/jaeger-client-go"

	e "github.com/jaegertracing/jaeger/plugin/storage/es/spanstore/dbmodel"
	json "github.com/json-iterator/go"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/syncutil"
	j "github.com/uber/jaeger-client-go/thrift-gen/jaeger"
	"go.uber.org/atomic"
)

var (
	logger                 = log.GetLogger("jaeger.reporter")
	_      jaeger.Reporter = (*ioReporter)(nil)
)

func NewIoReporter(writer io.Writer, batch int32) jaeger.Reporter {
	reporter := &ioReporter{
		writer:    writer,
		batchSize: batch,
		unbounded: syncutil.NewUnbounded(),
		domain:    e.NewFromDomain(false, nil, ""),
	}

	go reporter.loop()
	return reporter
}

type ioReporter struct {
	batchSize int32
	writer    io.Writer

	count     atomic.Int32
	process   *j.Process
	unbounded *syncutil.Unbounded
	domain    e.FromDomain
}

func (t *ioReporter) loop() {
	tick := time.NewTicker(time.Millisecond * 100)
	defer tick.Stop()

	for {
		select {
		case span, ok := <-t.unbounded.Get():
			if !ok {
				return
			}

			t.saveSpan(span)
			t.unbounded.Load()
		case <-tick.C:
			t.unbounded.Load()
		}
	}
}

func (t *ioReporter) Report(span *jaeger.Span) {
	if t.count.Load() > t.batchSize {
		logger.Error().
			Int32("batch", t.batchSize).
			Int32("count", t.count.Load()).
			Msg("The maximum number of spans has been exceeded")
	}

	if t.process == nil {
		t.process = jaeger.BuildJaegerProcessThrift(span)
	}

	jSpan := jaeger.BuildJaegerThrift(span)
	sp := t.domain.FromDomainEmbedProcess(toDomainSpan(jSpan, t.process))
	for _, ref := range sp.References {
		if ref.RefType == e.ChildOf {
			sp.ParentSpanID = ref.SpanID
		}
	}

	t.count.Inc()
	t.unbounded.Put(sp)
}

func (t *ioReporter) Close() {}
func (t *ioReporter) saveSpan(span interface{}) {
	defer recovery.Recovery(func(err error) {
		logger.Err(err).
			Int32("batch", t.batchSize).
			Int32("count", t.count.Load()).
			Msg("failed to saveSpan")
	})

	defer t.count.Dec()

	if span == nil || t.process == nil {
		return
	}

	s := assert.Must1(json.Marshal(span))
	assert.Must1(t.writer.Write(append(s, '\n')))
}
