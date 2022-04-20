// Package etcdv3 provides an etcd version 3 registry
package etcdv3

import (
	"context"
	"encoding/json"
	"errors"
	"path"
	"strings"
	"sync"

	hash "github.com/mitchellh/hashstructure"
	"github.com/pubgo/xerror"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	"go.etcd.io/etcd/client/v3"

	"github.com/pubgo/lava/clients/etcdv3"
	"github.com/pubgo/lava/config"
	registry2 "github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/pkg/merge"
)

func init() {
	registry2.Register(Name, func(m config.CfgMap) (registry2.Registry, error) {
		var cfg Cfg
		merge.MapStruct(&cfg, m)

		return &Registry{
			Cfg:      cfg,
			register: make(map[string]uint64),
			leases:   make(map[string]clientv3.LeaseID),
		}, nil
	})
}

type Registry struct {
	Cfg Cfg
	sync.Mutex
	Client   *etcdv3.Client `inject-expr:"$.Cfg.Name"`
	register map[string]uint64
	leases   map[string]clientv3.LeaseID
}

func (e *Registry) Close() {
}

func (e *Registry) Init() {
	inject.Inject(e)
}

func (e *Registry) RegLoop(f func() *registry2.Service, opt ...registry2.RegOpt) error {
	return e.Register(f(), opt...)
}

func encode(s *registry2.Service) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func decode(ds []byte) *registry2.Service {
	var s *registry2.Service
	xerror.Panic(json.Unmarshal(ds, &s))
	return s
}

func nodePath(prefix, s, id string) string {
	service := strings.Replace(s, "/", "-", -1)
	node := strings.Replace(id, "/", "-", -1)
	return path.Join(prefix, service, node)
}

func servicePath(prefix, s string) string {
	return path.Join(prefix, strings.Replace(s, "/", "-", -1))
}

func (e *Registry) Deregister(s *registry2.Service, opts ...registry2.DeregOpt) error {
	if len(s.Nodes) == 0 {
		return errors.New("Require at least one node")
	}

	e.Lock()
	// delete our hash of the service
	delete(e.register, s.Name)
	// delete our lease of the service
	delete(e.leases, s.Name)
	e.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, node := range s.Nodes {
		_, err := e.Client.Get().Delete(ctx, nodePath(e.Cfg.Prefix, s.Name, node.Id))
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Registry) Register(s *registry2.Service, opts ...registry2.RegOpt) error {
	if len(s.Nodes) == 0 {
		return errors.New("Require at least one node")
	}

	var leaseNotFound bool
	//refreshing lease if existing
	leaseID, ok := e.leases[s.Name]
	if ok {
		if _, err := e.Client.Get().KeepAliveOnce(context.TODO(), leaseID); err != nil {
			if err != rpctypes.ErrLeaseNotFound {
				return err
			}

			// lease not found do register
			leaseNotFound = true
		}
	}

	// create hash of service; uint64
	h, err := hash.Hash(s, nil)
	if err != nil {
		return err
	}

	// get existing hash
	e.Lock()
	v, ok := e.register[s.Name]
	e.Unlock()

	// the service is unchanged, skip registering
	if ok && v == h && !leaseNotFound {
		return nil
	}

	service := &registry2.Service{
		Name:      s.Name,
		Metadata:  s.Metadata,
		Endpoints: s.Endpoints,
	}

	var options registry2.RegOpts
	for _, o := range opts {
		o(&options)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var lgr *clientv3.LeaseGrantResponse
	if options.TTL.Seconds() > 0 {
		lgr, err = e.Client.Get().Grant(ctx, int64(options.TTL.Seconds()))
		if err != nil {
			return err
		}
	}

	for _, node := range s.Nodes {
		service.Nodes = []*registry2.Node{node}
		if lgr != nil {
			_, err = e.Client.Get().Put(ctx, nodePath(e.Cfg.Prefix, service.Name, node.Id), encode(service), clientv3.WithLease(lgr.ID))
		} else {
			_, err = e.Client.Get().Put(ctx, nodePath(e.Cfg.Prefix, service.Name, node.Id), encode(service))
		}
		if err != nil {
			return err
		}
	}

	e.Lock()
	// save our hash of the service
	e.register[s.Name] = h
	// save our leaseID of the service
	if lgr != nil {
		e.leases[s.Name] = lgr.ID
	}
	e.Unlock()

	return nil
}

func (e *Registry) GetService(name string, opts ...registry2.GetOpt) ([]*registry2.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	rsp, err := e.Client.Get().Get(ctx, servicePath(e.Cfg.Prefix, name)+"/", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		return nil, err
	}

	if len(rsp.Kvs) == 0 {
		return nil, registry2.ErrNotFound
	}

	var services []*registry2.Service
	for _, n := range rsp.Kvs {
		if sn := decode(n.Value); sn != nil {
			s := &registry2.Service{
				Name:      sn.Name,
				Metadata:  sn.Metadata,
				Endpoints: sn.Endpoints,
			}

			for _, node := range sn.Nodes {
				s.Nodes = append(s.Nodes, node)
			}
			services = append(services, s)
		}
	}

	return services, nil
}

func (e *Registry) ListService(opts ...registry2.ListOpt) ([]*registry2.Service, error) {
	var services []*registry2.Service
	nameSet := make(map[string]struct{})

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	rsp, err := e.Client.Get().Get(ctx, e.Cfg.Prefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		return nil, err
	}

	if len(rsp.Kvs) == 0 {
		return []*registry2.Service{}, nil
	}

	for _, n := range rsp.Kvs {
		if sn := decode(n.Value); sn != nil {
			nameSet[sn.Name] = struct{}{}
		}
	}
	for k := range nameSet {
		service := &registry2.Service{}
		service.Name = k
		services = append(services, service)
	}

	return services, nil
}

func (e *Registry) Watch(service string, opts ...registry2.WatchOpt) (registry2.Watcher, error) {
	return newWatcher(e, timeout, opts...)
}

func (e *Registry) String() string { return Name }
