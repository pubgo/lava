package resolver

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/pubgo/xlog"
	"google.golang.org/grpc/resolver"
)

const (
	//direct connect to service, it is can be used in k8s or other systems without service discovery
	DirectScheme    = "direct"
	DiscovScheme    = "discov"
	EndpointSepChar = ','
	subsetSize      = 32
)

var (
	EndpointSep = fmt.Sprintf("%c", EndpointSepChar)
)

func init() {
	//register resolver to global map
	resolver.Register(&directBuilder{})
	resolver.Register(&discovBuilder{})
}

type baseResolver struct {
	cc resolver.ClientConn
}

func (r *baseResolver) Close() {
}

func (r *baseResolver) ResolveNow(options resolver.ResolveNowOptions) {
	xlog.Fatal("do nothing here")
}

func BuildDirectTarget(endpoints []string) string {
	return fmt.Sprintf("%s:///%s", DirectScheme, strings.Join(endpoints, EndpointSep))
}

func BuildDiscovTarget(endpoints []string, key string) string {
	return fmt.Sprintf("%s://%s/%s", DiscovScheme, strings.Join(endpoints, EndpointSep), key)
}

//为了同一个地址多个连接的存在先使用y一个mock的sericeName
func getServiceName(address string, index int) string {
	return fmt.Sprintf("%s-%s", address, strconv.Itoa(index))
}

//对targets打散
func reshuffle(targets []string, limit int) []string {
	rand.Shuffle(len(targets), func(i, j int) {
		targets[i], targets[j] = targets[j], targets[i]
	})
	if len(targets) <= limit {
		return targets
	} else {
		return targets[:limit]
	}
}
