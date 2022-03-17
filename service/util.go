package service

import (
	"context"
	"github.com/pubgo/x/byteutil"
	"net"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/pubgo/lava/types"
)

// getPeerName 获取对端应用名称
func getPeerName(md metadata.MD) string {
	return types.HeaderGet(md, "app")
}

// getPeerIP 获取对端ip
func getPeerIP(md metadata.MD, ctx context.Context) string {
	clientIP := types.HeaderGet(md, "client-ip")
	if clientIP != "" {
		return clientIP
	}

	// 从grpc里取对端ip
	pr, ok2 := peer.FromContext(ctx)
	if !ok2 {
		return ""
	}

	if pr.Addr == net.Addr(nil) {
		return ""
	}

	addSlice := strings.Split(pr.Addr.String(), ":")
	if len(addSlice) > 1 {
		return addSlice[0]
	}
	return ""
}

func ignoreMuxError(err error) bool {
	if err == nil {
		return true
	}
	return strings.Contains(err.Error(), "use of closed network connection") ||
		strings.Contains(err.Error(), "mux: server closed")
}

func convertHeader(request interface{ VisitAll(func(key, value []byte)) }) types.Header {
	var h = metadata.MD{}
	request.VisitAll(func(key, value []byte) {
		h.Set(byteutil.ToStr(key), byteutil.ToStr(value))
	})
	return h
}

func getPort(addr string) string {
	var addrList = strings.Split(addr, ":")
	return addrList[len(addrList)-1]
}
