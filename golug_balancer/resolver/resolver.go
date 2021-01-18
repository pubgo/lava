package resolver

import (
	"fmt"
	"math/rand"
	"strings"

	registry "github.com/pubgo/golug/golug_registry"
	"google.golang.org/grpc/resolver"
)

const (
	//direct connect to service, it is can be used in k8s or other systems without service discovery
	DirectScheme    = "direct"
	DiscovScheme    = "discov"
	EndpointSepChar = ','
)

var (
	EndpointSep = fmt.Sprintf("%c", EndpointSepChar)
	Replica     = 3
)

func init() {
	resolver.Register(&directBuilder{})
	resolver.Register(&discovBuilder{})
}

type baseResolver struct {
	cc resolver.ClientConn
	r  registry.Watcher
}

func (r *baseResolver) Close() {
	if r.r != nil {
		r.r.Stop()
	}
}

func (r *baseResolver) ResolveNow(options resolver.ResolveNowOptions) {}

func BuildDirectTarget(endpoints []string) string {
	return fmt.Sprintf("%s:///%s", DirectScheme, strings.Join(endpoints, EndpointSep))
}

func BuildDiscovTarget(endpoints []string, key string) string {
	return fmt.Sprintf("%s://%s/%s", DiscovScheme, strings.Join(endpoints, EndpointSep), key)
}

//对targets打散
func reshuffle(targets []string) []string {
	rand.Shuffle(len(targets), func(i, j int) { targets[i], targets[j] = targets[j], targets[i] })
	return targets
}
