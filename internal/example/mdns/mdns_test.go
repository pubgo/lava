package mdns

import (
	"context"
	"testing"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
)

func TestServer(t *testing.T) {
	server, err := zeroconf.RegisterProxy("t1", "t", "lava.local.", 1234, "kkk", []string{"127.0.0.1"}, []string{"hello1"}, nil)
	xerror.Panic(err)
	_ = server
	q.Q(server)

	select {}
}

func TestServer1(t *testing.T) {
	server, err := zeroconf.Register("t2", "t", "local.", 1234, []string{"hello"}, nil)
	xerror.Panic(err)
	q.Q(server)

	server, err = zeroconf.Register("t3", "t", "local.", 1234, []string{"hello"}, nil)
	xerror.Panic(err)
	q.Q(server)
	select {}
}

func TestClient(t *testing.T) {
	resolver, err := zeroconf.NewResolver()
	xerror.Panic(err)

	entries := make(chan *zeroconf.ServiceEntry)
	_ = fx.Go(func(ctx context.Context) {
		go func(results <-chan *zeroconf.ServiceEntry) {
			for s := range results {
				q.Q(s)
			}
		}(entries)

		//time.Sleep(time.Second * 5)
	})

	var ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	//xerror.Panic(resolver.Lookup(ctx, "t2", "t", "local", entries), "Failed to Lookup")
	xerror.Panic(resolver.Browse(ctx, "t", "local.", entries), "Failed to Lookup")
	<-ctx.Done()
}
