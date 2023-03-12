package k8s

import (
	"context"
	"fmt"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/merge"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/pubgo/lava/core/discovery"
	"github.com/pubgo/lava/core/service"
	"github.com/pubgo/lava/internal/consts"
	"github.com/pubgo/lava/pkg/k8sutil"
)

const (
	// LabelsKeyServiceID is used to define the ID of the service
	LabelsKeyServiceID = "lava-service-id"
	// LabelsKeyServiceName is used to define the name of the service
	LabelsKeyServiceName = "lava-service-app"
	// LabelsKeyServiceVersion is used to define the version of the service
	LabelsKeyServiceVersion = "lava-service-version"
	// AnnotationsKeyMetadata is used to define the metadata of the service
	AnnotationsKeyMetadata = "lava-service-metadata"
	// AnnotationsKeyProtocolMap is used to define the protocols of the service
	// Through the value of this field, lava can obtain the application layer protocol corresponding to the port
	// Example value: {"80": "http", "8081": "grpc"}
	AnnotationsKeyProtocolMap = "lava-service-protocols"
)

func New(c *discovery.Config) (_ discovery.Discovery, err error) {
	defer recovery.Err(&err)

	var cfg Cfg
	merge.MapStruct(&cfg, c.DriverCfg).Unwrap()

	var client = cfg.Build()
	return NewDiscovery(client), nil
}

// NewDiscovery is used to initialize the discoveryImpl
func NewDiscovery(clientSet *kubernetes.Clientset) discovery.Discovery {
	return &discoveryImpl{client: clientSet, stopCh: make(chan struct{})}
}

var _ discovery.Discovery = (*discoveryImpl)(nil)

type discoveryImpl struct {
	client *kubernetes.Clientset
	stopCh chan struct{}
}

func (s *discoveryImpl) GetService(ctx1 context.Context, name string, opt ...discovery.GetOpt) result.Result[[]*service.Service] {
	var ctx, cancel = context.WithTimeout(ctx1, consts.DefaultTimeout)
	defer cancel()

	ep := assert.Must1(s.client.CoreV1().Endpoints(k8sutil.Namespace()).Get(ctx, name, metav1.GetOptions{}))

	endpoints := assert.Must1(s.client.
		CoreV1().
		Endpoints(k8sutil.Namespace()).
		List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("%s=%s", "metadata.name", name)}))

	return async.Yield(func(yield func(*service.Service)) error {
		for _, endpoint := range endpoints.Items {
			for _, subset := range endpoint.Subsets {
				realPort := ""
				for _, p := range subset.Ports {
					realPort = fmt.Sprint(p.Port)
					break
				}

				for _, addr := range subset.Addresses {
					yield(&service.Service{
						Name: name,
						Nodes: []*service.Node{
							{
								Id:      string(addr.TargetRef.UID),
								Address: fmt.Sprintf("%s:%s", addr.IP, realPort),
							},
						},
					})
				}
			}
		}
		return nil
	}).ToList()
}

func (s *discoveryImpl) String() string { return name }

// Watch creates a watcher according to the service name.
func (s *discoveryImpl) Watch(ctx context.Context, name string, opt ...discovery.WatchOpt) result.Result[discovery.Watcher] {
	return newWatcher(s.client, name)
}
