package reporter

import (
	"io"
	"time"

	e "github.com/jaegertracing/jaeger/plugin/storage/es/spanstore/dbmodel"
	json "github.com/json-iterator/go"
	"github.com/pubgo/x/syncutil"
	_ "github.com/pubgo/x/syncutil"
	"github.com/pubgo/xerror"
	"github.com/uber/jaeger-client-go"
	j "github.com/uber/jaeger-client-go/thrift-gen/jaeger"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/pubgo/lava/internal/logz"
)

var _ jaeger.Reporter = (*ioReporter)(nil)

func NewIoReporter(writer io.Writer, batch int32) jaeger.Reporter {
	var reporter = &ioReporter{
		writer:    writer,
		batchSize: 1,
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
	var tick = time.NewTicker(time.Millisecond * 100)
	defer tick.Stop()

	for {
		select {
		case span, ok := <-t.unbounded.Get():
			if !ok {
				return
			}

			xerror.Panic(t.saveSpan(span))
			t.unbounded.Load()
		case <-tick.C:
			t.unbounded.Load()
		}
	}
}

var logs = logz.New("jaeger.reporter")

func (t *ioReporter) Report(span *jaeger.Span) {
	if t.count.Load() > t.batchSize {
		logs.With(
			zap.Int32("batch", t.batchSize),
			zap.Int32("count", t.count.Load()),
		).Error("The maximum number of spans has been exceeded")
	}

	if t.process == nil {
		t.process = jaeger.BuildJaegerProcessThrift(span)
	}

	jSpan := jaeger.BuildJaegerThrift(span)
	sp := t.domain.FromDomainEmbedProcess(toDomainSpan(jSpan, t.process))
	for _, ref := range sp.References {
		if ref.RefType == e.ChildOf {
			// *The reason of this field be deprecated is unknown*.
			sp.ParentSpanID = ref.SpanID
		}
	}

	t.count.Inc()
	t.unbounded.Put(sp)
}

func (t *ioReporter) Close() {}
func (t *ioReporter) saveSpan(span interface{}) (gErr error) {
	defer xerror.RespErr(&gErr)
	defer t.count.Dec()

	if span == nil || t.process == nil {
		return nil
	}

	s, err := json.Marshal(span)
	if err != nil {
		return xerror.Wrap(err)
	}

	_, err = t.writer.Write(append(s, '\n'))
	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}
