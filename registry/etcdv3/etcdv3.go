// Package etcdv3 provides an etcd version 3 registry
package etcdv3

import (
	"context"
	"encoding/json"
	"errors"
	etcdv32 "github.com/pubgo/lug/plugins/etcdv3"
	"path"
	"strings"
	"sync"

	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	hash "github.com/mitchellh/hashstructure"
	registry "github.com/pubgo/lug/registry"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"go.etcd.io/etcd/clientv3"
)

func init() {
	registry.Register(Name, func(m map[string]interface{}) (registry.Registry, error) {
		var cfg Cfg
		xerror.Panic(merge.MapStruct(&cfg, m))
		return &Registry{
			cfg:      cfg,
			register: make(map[string]uint64),
			leases:   make(map[string]clientv3.LeaseID),
		}, nil
	})
}

func NewRegistry() registry.Registry {
	return &Registry{
		register: make(map[string]uint64),
		leases:   make(map[string]clientv3.LeaseID),
	}
}

type Registry struct {
	sync.Mutex
	client   *etcdv32.Client
	cfg      Cfg
	register map[string]uint64
	leases   map[string]clientv3.LeaseID
}

func encode(s *registry.Service) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func decode(ds []byte) *registry.Service {
	var s *registry.Service
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

func (e *Registry) DeRegister(s *registry.Service, opts ...registry.DeRegOpt) error {
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
		_, err := e.client.Delete(ctx, nodePath(e.cfg.Prefix, s.Name, node.Id))
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Registry) Register(s *registry.Service, opts ...registry.RegOpt) error {
	if len(s.Nodes) == 0 {
		return errors.New("Require at least one node")
	}

	var leaseNotFound bool
	//refreshing lease if existing
	leaseID, ok := e.leases[s.Name]
	if ok {
		if _, err := e.client.KeepAliveOnce(context.TODO(), leaseID); err != nil {
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

	service := &registry.Service{
		Name:      s.Name,
		Version:   s.Version,
		Metadata:  s.Metadata,
		Endpoints: s.Endpoints,
	}

	var options registry.RegOpts
	for _, o := range opts {
		o(&options)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var lgr *clientv3.LeaseGrantResponse
	if options.TTL.Seconds() > 0 {
		lgr, err = e.client.Grant(ctx, int64(options.TTL.Seconds()))
		if err != nil {
			return err
		}
	}

	for _, node := range s.Nodes {
		service.Nodes = []*registry.Node{node}
		if lgr != nil {
			_, err = e.client.Put(ctx, nodePath(e.cfg.Prefix, service.Name, node.Id), encode(service), clientv3.WithLease(lgr.ID))
		} else {
			_, err = e.client.Put(ctx, nodePath(e.cfg.Prefix, service.Name, node.Id), encode(service))
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

func (e *Registry) GetService(name string, opts ...registry.GetOpt) ([]*registry.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	rsp, err := e.client.Get(ctx, servicePath(e.cfg.Prefix, name)+"/", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		return nil, err
	}

	if len(rsp.Kvs) == 0 {
		return nil, registry.ErrNotFound
	}

	serviceMap := map[string]*registry.Service{}

	for _, n := range rsp.Kvs {
		if sn := decode(n.Value); sn != nil {
			s, ok := serviceMap[sn.Version]
			if !ok {
				s = &registry.Service{
					Name:      sn.Name,
					Version:   sn.Version,
					Metadata:  sn.Metadata,
					Endpoints: sn.Endpoints,
				}
				serviceMap[s.Version] = s
			}

			for _, node := range sn.Nodes {
				s.Nodes = append(s.Nodes, node)
			}
		}
	}

	var services []*registry.Service
	for _, service := range serviceMap {
		services = append(services, service)
	}
	return services, nil
}

func (e *Registry) ListServices(opts ...registry.ListOpt) ([]*registry.Service, error) {
	var services []*registry.Service
	nameSet := make(map[string]struct{})

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	rsp, err := e.client.Get(ctx, e.cfg.Prefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		return nil, err
	}

	if len(rsp.Kvs) == 0 {
		return []*registry.Service{}, nil
	}

	for _, n := range rsp.Kvs {
		if sn := decode(n.Value); sn != nil {
			nameSet[sn.Name] = struct{}{}
		}
	}
	for k := range nameSet {
		service := &registry.Service{}
		service.Name = k
		services = append(services, service)
	}

	return services, nil
}

func (e *Registry) Watch(service string, opts ...registry.WatchOpt) (registry.Watcher, error) {
	return newWatcher(e, timeout, opts...)
}

func (e *Registry) String() string { return Name }
