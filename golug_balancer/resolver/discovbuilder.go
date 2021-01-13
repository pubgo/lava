package resolver

import (
	"fmt"
	registry "github.com/pubgo/golug/golug_registry"

	"google.golang.org/grpc/resolver"
)

func updateState(cc resolver.ClientConn, services []string) {
	var addrs []resolver.Address
	for _, val := range reshuffle(services, subsetSize) {
		addrs = append(addrs, resolver.Address{Addr: val})
	}
	cc.UpdateState(resolver.State{Addresses: addrs})
}

type discovBuilder struct{}

func (d *discovBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var r = registry.Get(target.Authority)
	if r == nil {
		return nil, fmt.Errorf("registry %s not exists", target.Authority)
	}

	if services, err := r.GetService(target.Endpoint); err != nil {
		panic(err)
	} else {
		var addrs []string
		for _, service := range services {
			for _, node := range service.Nodes {
				addr := node.Address
				if node.Port > 0 {
					addr = fmt.Sprintf("%s:%d", node.Address, node.Port)
				}
				addrs = append(addrs, addr)
			}
		}

		if len(addrs) == 0 {
			return nil, fmt.Errorf("service none available")
		}

		updateState(cc, addrs)
	}

	w, err := r.Watch(registry.WatchService(target.Endpoint))
	if err != nil {
		return nil, err
	}

	go func() {
		defer w.Stop()
		for {
			res, err := w.Next()
			if err == registry.ErrWatcherStopped {
				break
			}

			if err != nil {
				log.Error(err)
				continue
			}

			var addrs []string
			for _, node := range res.Service.Nodes {
				addr := node.Address
				if node.Port > 0 {
					addr = fmt.Sprintf("%s:%d", node.Address, node.Port)
				}
				addrs = append(addrs, addr)
			}
			updateState(cc, addrs)
		}
	}()

	return &baseResolver{cc: cc}, nil
}

func (d *discovBuilder) Scheme() string {
	return DiscovScheme
}
