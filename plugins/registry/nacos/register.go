package nacos

import (
	"errors"
	"fmt"
	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/plugins/registry/registry_type"
	"net"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/clients/nacos"
	"github.com/pubgo/lava/plugins/registry"
)

func init() {
	registry.Register(Name, func(m config_type.CfgMap) (registry_type.Registry, error) {
		var cfg Cfg
		xerror.Panic(merge.MapStruct(&cfg, m))
		var c = nacos.Get(cfg.Driver)
		xerror.Assert(c == nil, "please init nacos client")
		n := &nacosRegistry{cfg: cfg}
		n.client = c.GetRegistry()
		return n, nil
	})
}

type nacosRegistry struct {
	client naming_client.INamingClient
	cfg    Cfg
}

func (n *nacosRegistry) RegLoop(f func() *registry_type.Service, opt ...registry_type.RegOpt) error {
	return n.Register(f(), opt...)
}

func (n *nacosRegistry) Register(s *registry_type.Service, opts ...registry_type.RegOpt) error {
	var options registry_type.RegOpts
	for _, o := range opts {
		o(&options)
	}
	withContext := false
	param := vo.RegisterInstanceParam{}
	if options.Context != nil {
		if p, ok := options.Context.Value("register_instance_param").(vo.RegisterInstanceParam); ok {
			param = p
			withContext = ok
		}
	}
	if !withContext {
		host, port, err := getNodeIPPort(s)
		if err != nil {
			return err
		}
		s.Nodes[0].Metadata["version"] = s.Version
		param.Ip = host
		param.Port = uint64(port)
		param.Metadata = s.Nodes[0].Metadata
		param.ServiceName = s.Name
		param.Enable = true
		param.Healthy = true
		param.Weight = 1.0
		param.Ephemeral = true
	}
	_, err := n.client.RegisterInstance(param)
	return err
}

func (n *nacosRegistry) Deregister(s *registry_type.Service, opts ...registry_type.DeregOpt) error {
	var options registry_type.DeregOpts
	for _, o := range opts {
		o(&options)
	}
	withContext := false
	param := vo.DeregisterInstanceParam{}
	if options.Context != nil {
		if p, ok := options.Context.Value("deregister_instance_param").(vo.DeregisterInstanceParam); ok {
			param = p
			withContext = ok
		}
	}
	if !withContext {
		host, port, err := getNodeIPPort(s)
		if err != nil {
			return err
		}
		param.Ip = host
		param.Port = uint64(port)
		param.ServiceName = s.Name
	}

	_, err := n.client.DeregisterInstance(param)
	return err
}

func (n *nacosRegistry) Watch(s string, opt ...registry_type.WatchOpt) (registry_type.Watcher, error) {
	return newWatcher(n, opt...)
}

func (n *nacosRegistry) ListService(opts ...registry_type.ListOpt) ([]*registry_type.Service, error) {
	var options registry_type.ListOpts
	for _, o := range opts {
		o(&options)
	}
	withContext := false
	param := vo.GetAllServiceInfoParam{}
	if options.Context != nil {
		if p, ok := options.Context.Value("get_all_service_info_param").(vo.GetAllServiceInfoParam); ok {
			param = p
			withContext = ok
		}
	}
	if !withContext {
		services, err := n.client.GetAllServicesInfo(param)
		if err != nil {
			return nil, err
		}
		param.PageNo = 1
		param.PageSize = uint32(services.Count)
	}
	services, err := n.client.GetAllServicesInfo(param)
	if err != nil {
		return nil, err
	}
	var registryServices []*registry_type.Service
	for _, v := range services.Doms {
		registryServices = append(registryServices, &registry_type.Service{Name: v})
	}
	return registryServices, nil
}

func (n *nacosRegistry) GetService(name string, opts ...registry_type.GetOpt) ([]*registry_type.Service, error) {
	var options registry_type.GetOpts
	for _, o := range opts {
		o(&options)
	}
	withContext := false
	param := vo.GetServiceParam{}
	if options.Context != nil {
		if p, ok := options.Context.Value("select_instances_param").(vo.GetServiceParam); ok {
			param = p
			withContext = ok
		}
	}
	if !withContext {
		param.ServiceName = name
	}
	service, err := n.client.GetService(param)
	if err != nil {
		return nil, err
	}
	services := make([]*registry_type.Service, 0)
	for _, v := range service.Hosts {
		if !v.Healthy || !v.Enable || v.Weight <= 0 {
			continue
		}

		nodes := make([]*registry_type.Node, 0)
		nodes = append(nodes, &registry_type.Node{
			Id:       v.InstanceId,
			Address:  net.JoinHostPort(v.Ip, fmt.Sprintf("%d", v.Port)),
			Metadata: v.Metadata,
		})
		s := registry_type.Service{
			Name:     v.ServiceName,
			Version:  v.Metadata["version"],
			Metadata: v.Metadata,
			Nodes:    nodes,
		}
		services = append(services, &s)
	}

	return services, nil
}

func getNodeIPPort(s *registry_type.Service) (host string, port int, err error) {
	if len(s.Nodes) == 0 {
		return "", 0, errors.New("you must deregister at least one node")
	}
	node := s.Nodes[0]
	host, pt, err := net.SplitHostPort(node.Address)
	if err != nil {
		return "", 0, err
	}
	port, err = strconv.Atoi(pt)
	if err != nil {
		return "", 0, err
	}
	return
}

func (n *nacosRegistry) String() string { return Name }
