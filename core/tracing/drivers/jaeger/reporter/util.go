package reporter

import (
	"fmt"
	
	"github.com/jaegertracing/jaeger/model"
	j "github.com/uber/jaeger-client-go/thrift-gen/jaeger"
)

func toDomainSpan(jSpan *j.Span, jProcess *j.Process) *model.Span {
	return transformSpan(jSpan, getProcess(jProcess))
}

func getProcess(jProcess *j.Process) *model.Process {
	if jProcess == nil {
		return nil
	}

	tags := getTags(jProcess.Tags, 0)
	return &model.Process{Tags: tags, ServiceName: jProcess.ServiceName}
}

func getTags(tags []*j.Tag, extraSpace int) model.KeyValues {
	if len(tags) == 0 {
		return nil
	}

	retKv := make(model.KeyValues, len(tags), len(tags)+extraSpace)
	for i, tag := range tags {
		retKv[i] = getTag(tag)
	}
	return retKv
}

func getTag(tag *j.Tag) model.KeyValue {
	switch tag.VType {
	case j.TagType_BOOL:
		return model.Bool(tag.Key, tag.GetVBool())
	case j.TagType_BINARY:
		return model.Binary(tag.Key, tag.GetVBinary())
	case j.TagType_DOUBLE:
		return model.Float64(tag.Key, tag.GetVDouble())
	case j.TagType_LONG:
		return model.Int64(tag.Key, tag.GetVLong())
	case j.TagType_STRING:
		return model.String(tag.Key, tag.GetVStr())
	default:
		return model.String(tag.Key, fmt.Sprintf("Unknown VType: %+v", tag))
	}
}

func transformSpan(jSpan *j.Span, mProcess *model.Process) *model.Span {
	traceID := model.NewTraceID(uint64(jSpan.TraceIdHigh), uint64(jSpan.TraceIdLow))
	//allocate extra space for future append operation
	tags := getTags(jSpan.Tags, 1)
	refs := getReferences(jSpan.References)

	// We no longer store ParentSpanID in the domain model, but the data in Thrift model
	// might still have these IDs without representing them in the References, so we
	// convert it back into child-of reference.
	if jSpan.ParentSpanId != 0 {
		parentSpanID := model.NewSpanID(uint64(jSpan.ParentSpanId))
		refs = model.MaybeAddParentSpanID(traceID, parentSpanID, refs)
	}

	return &model.Span{
		TraceID:       traceID,
		SpanID:        model.NewSpanID(uint64(jSpan.SpanId)),
		OperationName: jSpan.OperationName,
		References:    refs,
		Flags:         model.Flags(jSpan.Flags),
		StartTime:     model.EpochMicrosecondsAsTime(uint64(jSpan.StartTime)),
		Duration:      model.MicrosecondsAsDuration(uint64(jSpan.Duration)),
		Tags:          tags,
		Logs:          getLogs(jSpan.Logs),
		Process:       mProcess,
	}
}

func getReferences(jRefs []*j.SpanRef) []model.SpanRef {
	if len(jRefs) == 0 {
		return nil
	}

	mRefs := make([]model.SpanRef, len(jRefs))
	for idx, jRef := range jRefs {
		mRefs[idx].RefType = model.SpanRefType(int(jRef.RefType))
		mRefs[idx].TraceID = model.NewTraceID(uint64(jRef.TraceIdHigh), uint64(jRef.TraceIdLow))
		mRefs[idx].SpanID = model.NewSpanID(uint64(jRef.SpanId))
	}
	return mRefs
}

func getLogs(logs []*j.Log) []model.Log {
	if len(logs) == 0 {
		return nil
	}

	retLog := make([]model.Log, len(logs))
	for i, log := range logs {
		retLog[i].Fields = getTags(log.Fields, 0)
		retLog[i].Timestamp = model.EpochMicrosecondsAsTime(uint64(log.Timestamp))
	}
	return retLog
}
