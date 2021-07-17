package reporter

import (
	"encoding/json"
	"io"

	e "github.com/jaegertracing/jaeger/plugin/storage/es/spanstore/dbmodel"
	"github.com/pubgo/xerror"
	"github.com/uber/jaeger-client-go"
	j "github.com/uber/jaeger-client-go/thrift-gen/jaeger"
)

var _ jaeger.Reporter = (*ioReporter)(nil)

func NewIoReporter(writer io.Writer) jaeger.Reporter {
	return &ioReporter{writer: writer, batchSize: 1}
}

type ioReporter struct {
	batchSize int
	writer    io.Writer

	spans   []*j.Span
	process *j.Process
}

func (t *ioReporter) Report(span *jaeger.Span) {
	if t.process == nil {
		t.process = jaeger.BuildJaegerProcessThrift(span)
	}

	jSpan := jaeger.BuildJaegerThrift(span)
	t.spans = append(t.spans, jSpan)
	if len(t.spans) >= t.batchSize {
		t.Flush()
	}
}

func (t *ioReporter) Close() {}

// Flush submits the internal buffer to the remote server. It returns the
// number of spans flushed. If error is returned, the returned number of
// spans is treated as failed spans, and reported to metrics accordingly.
func (t *ioReporter) Flush() {
	count := len(t.spans)
	if count == 0 {
		return
	}
	xerror.Panic(t.save(t.spans))
	t.spans = t.spans[:0]
	return
}

// save 内部方法，不可以暴露，并且，如果重写这个类时，也要确保
// 1. spans 有内容
// 2. t.process 已经被初始化
func (t *ioReporter) save(spans []*j.Span) error {
	if len(spans) == 0 || t.process == nil {
		return nil
	}

	do := e.NewFromDomain(false, nil, "")

	for i := range spans {
		sp := do.FromDomainEmbedProcess(ToDomainSpan(spans[i], t.process))

		// Add ParentSpanID to spans.
		for _, ref := range sp.References {
			if ref.RefType == e.ChildOf {
				// *The reason of this field be deprecated is unknown*.
				sp.ParentSpanID = ref.SpanID
			}
		}

		s, err := json.Marshal(sp)
		if err != nil {
			return err
		}

		t.writer.Write(s)
	}

	return nil
}
