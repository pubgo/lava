package k8s

import (
	"context"
	"fmt"
	"github.com/pubgo/dix"

	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/pkg/k8s"
)

// Defines the key name of specific fields
// lava needs to cooperate with the following fields to run properly on Kubernetes:
// lava-service-id: define the ID of the service
// lava-service-app: define the name of the service
// lava-service-version: define the version of the service
// lava-service-metadata: define the metadata of the service
// lava-service-protocols: define the protocols of the service
//
// Example Deployment:
//
// apiVersion: apps/v1
// kind: Deployment
// metadata:
//  name: nginx
//  labels:
//    app: nginx
// spec:
//  replicas: 5
//  selector:
//    matchLabels:
//      app: nginx
//  template:
//    metadata:
//      labels:
//        app: nginx
//        lava-service-id: "56991810-c77f-4a95-8190-393efa9c1a61"
//        lava-service-app: "nginx"
//        lava-service-version: "v3.5.0"
//      annotations:
//        lava-service-protocols: |
//          {"80": "http"}
//        lava-service-metadata: |
//          {"region": "sh", "zone": "sh001", "cluster": "pd"}
//    spec:
//      containers:
//        - name: nginx
//          image: nginx:1.7.9
//          ports:
//           - containerPort: 80

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

func init() {
	dix.Register(func(m config.CfgMap) (_ registry.Registry, err error) {
		defer xerror.RespErr(&err)

		var cfg Cfg
		xerror.Panic(merge.MapStruct(&cfg, m))

		var client = cfg.Build()
		return NewRegistry(client), nil
	})
}

// NewRegistry is used to initialize the Registry
func NewRegistry(clientSet *kubernetes.Clientset) *Registry {
	return &Registry{
		client: clientSet,
		stopCh: make(chan struct{}),
	}
}

var _ registry.Registry = (*Registry)(nil)

// Registry The registry simply implements service discovery based on Kubernetes
// It has not been verified in the production environment and is currently for reference only
type Registry struct {
	client *kubernetes.Clientset
	stopCh chan struct{}
}

func (s *Registry) Init() {
}

func (s *Registry) Close() {
}

func (s *Registry) Deregister(service *registry.Service, opt ...registry.DeregOpt) error {
	return nil
	//return s.Dix(&registry.Service{Metadata: map[string]string{},})
}

func (s *Registry) GetService(name string, opt ...registry.GetOpt) (_ []*registry.Service, err error) {
	defer xerror.RespErr(&err)

	var ctx, cancel = context.WithTimeout(context.Background(), consts.DefaultTimeout)
	defer cancel()

	endpoints, err := s.client.
		CoreV1().
		Endpoints(k8s.Namespace()).
		List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("%s=%s", "metadata.name", name)})
	xerror.Panic(err)

	var resp []*registry.Service
	for _, endpoint := range endpoints.Items {
		for _, subset := range endpoint.Subsets {
			realPort := ""
			for _, p := range subset.Ports {
				realPort = fmt.Sprint(p.Port)
				break
			}

			for _, addr := range subset.Addresses {
				resp = append(resp, &registry.Service{
					Name: name,
					Nodes: []*registry.Node{
						{
							Id:      string(addr.TargetRef.UID),
							Address: fmt.Sprintf("%s:%s", addr.IP, realPort),
						},
					},
				})
			}
		}
	}

	return resp, nil
}

func (s *Registry) ListService(opt ...registry.ListOpt) ([]*registry.Service, error) {
	return nil, nil
}

func (s *Registry) String() string { return name }

// Register is used to register services
// Note that on Kubernetes, it can only be used to update the id/name/version/metadata/protocols of the current service,
// but it cannot be used to update node.
func (s *Registry) Register(service *registry.Service, opt ...registry.RegOpt) error {

	//patchBytes, err := jsoniter.Marshal(map[string]interface{}{
	//	"metadata": metav1.ObjectMeta{
	//		Labels: map[string]string{
	//			LabelsKeyServiceID:      service.ID,
	//			LabelsKeyServiceName:    service.Name,
	//			LabelsKeyServiceVersion: service.Version,
	//		},
	//		Annotations: map[string]string{
	//			AnnotationsKeyMetadata:    metadataVal,
	//			AnnotationsKeyProtocolMap: protocolMapVal,
	//		},
	//	},
	//})

	//if _, err = s.client.
	//	CoreV1().
	//	Pods(k8s.Namespace()).
	//	Patch(context.TODO(), k8s.GetPodName(), types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}); err != nil {
	//	return err
	//}
	return nil
}

// Watch creates a watcher according to the service name.
func (s *Registry) Watch(name string, opt ...registry.WatchOpt) (registry.Watcher, error) {
	return newWatcher(s, name), nil
}
