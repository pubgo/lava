package transport

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/apache/thrift/lib/go/thrift"
	jConv "github.com/jaegertracing/jaeger/model/converter/thrift/jaeger"
	e "github.com/jaegertracing/jaeger/plugin/storage/es/spanstore/dbmodel"
	tJaeger "github.com/jaegertracing/jaeger/thrift-gen/jaeger"
	"github.com/uber/jaeger-client-go"
	jThrift "github.com/uber/jaeger-client-go/thrift"
	j "github.com/uber/jaeger-client-go/thrift-gen/jaeger"
)

var _ jaeger.Transport = (*ioTransport)(nil)

type ioTransport struct {
	batchSize int
	writer    io.Writer

	spans   []*j.Span
	process *j.Process
}

// NewIOTransport return a transport for jaeger
// the transport will write messages by the io.Writer which the user set.
func NewIOTransport(writer io.Writer, batchSize int) jaeger.Transport {
	return &ioTransport{
		batchSize: batchSize,
		writer:    writer,
		spans:     []*j.Span{},
	}
}

// Append converts the spans to the wire representation and adds it
// to sender's internal buffer.  If the buffer exceeds its designated
// size, the transport should call Flush() and return the number of spans
// flushed, otherwise return 0. If error is returned, the returned number
// of spans is treated as failed spans, and reported to metrics accordingly.
func (t *ioTransport) Append(span *jaeger.Span) (int, error) {
	if t.process == nil {
		t.process = jaeger.BuildJaegerProcessThrift(span)
	}
	jSpan := jaeger.BuildJaegerThrift(span)
	t.spans = append(t.spans, jSpan)
	if len(t.spans) >= t.batchSize {
		return t.Flush()
	}
	return 0, nil
}

// Flush submits the internal buffer to the remote server. It returns the
// number of spans flushed. If error is returned, the returned number of
// spans is treated as failed spans, and reported to metrics accordingly.
func (t *ioTransport) Flush() (int, error) {
	count := len(t.spans)
	if count == 0 {
		return 0, nil
	}
	err := t.save(t.spans)
	t.spans = t.spans[:0]
	return count, err
}

func (t *ioTransport) Close() error {
	return nil
}

// save 内部方法，不可以暴露，并且，如果重写这个类时，也要确保
// 1. spans 有内容
// 2. t.process 已经被初始化
func (t *ioTransport) save(spans []*j.Span) error {
	if len(spans) == 0 || t.process == nil {
		return nil
	}

	batch := &j.Batch{
		Spans:   spans,
		Process: t.process,
	}
	body, err := t.serializeThrift(batch)
	if err != nil {
		return err
	}

	if err := t.deserializeAsES(body.Bytes()); err != nil {
		return err
	}

	return nil
}

func (t *ioTransport) serializeThrift(obj jThrift.TStruct) (*bytes.Buffer, error) {
	buf := jThrift.NewTMemoryBuffer()
	p := jThrift.NewTBinaryProtocolTransport(buf)
	if err := obj.Write(p); err != nil {
		return nil, err
	}

	return buf.Buffer, nil
}

func (t *ioTransport) deserializeAsES(b []byte) error {
	tdes := thrift.NewTDeserializer()
	// (NB): We decided to use this struct instead of straight batches to be as consistent with tchannel intake as possible.
	batch := tJaeger.Batch{}
	if err := tdes.Read(&batch, b); err != nil {
		return err
	}

	// FromDomain can convert domain-spans to embed-spans
	do := e.NewFromDomain(false, nil, "")

	for _, span := range batch.Spans {
		mspan := jConv.ToDomainSpan(span, batch.Process)
		espan := do.FromDomainEmbedProcess(mspan)

		// Add ParentSpanID to spans.
		for _, ref := range espan.References {
			if ref.RefType == e.ChildOf {
				// *The reason of this field be deprecated is unknown*.
				espan.ParentSpanID = ref.SpanID
			}
		}

		s, err := json.Marshal(espan)
		if err != nil {
			return err
		}

		// write messages
		if _, err := t.writer.Write(s); err != nil {
			return err
		}
	}

	return nil
}
