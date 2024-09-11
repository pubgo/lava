// Code generated by protoc-gen-cloud-job. DO NOT EDIT.
// versions:
// - protoc-gen-cloud-job v0.0.1
// - protoc                 v5.27.1
// source: gid/id.proto

package gidpb

import (
	"context"
	cloudjobs "github.com/pubgo/lava/component/cloudjobs"
)

// IdProxyExecEventKey Id/ProxyExecEvent
const IdProxyExecEventKey = "gid.proxy.exec"

var _ = cloudjobs.RegisterSubject(IdProxyExecEventKey, new(DoProxyEventReq))

func RegisterIdProxyExecEventCloudJob(jobCli *cloudjobs.Client, handler func(ctx *cloudjobs.Context, req *DoProxyEventReq) error) {
	cloudjobs.RegisterJobHandler(jobCli, "gid", IdProxyExecEventKey, handler)
}

func PushIdProxyExecEventCloudJob(jobCli *cloudjobs.Client, ctx context.Context, req *DoProxyEventReq) error {
	return jobCli.Publish(ctx, IdProxyExecEventKey, req)
}