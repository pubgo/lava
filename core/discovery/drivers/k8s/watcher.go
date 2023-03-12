package k8s

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/result"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/tools/cache"

	"github.com/pubgo/lava/core/discovery"
	"github.com/pubgo/lava/core/service"
	"github.com/pubgo/lava/pkg/k8sutil"
	"github.com/pubgo/lava/pkg/proto/event/v1"
)

var _ discovery.Watcher = (*Watcher)(nil)

// Watcher performs the conversion from channel to iterator
// It reads the latest changes from the `chan []*discovery.ServiceInstance`
// And the outside can sense the closure of Watcher through stopCh
type Watcher struct {
	service string
	watcher watch.Interface
	client  *kubernetes.Clientset
}

// newWatcher is used to initialize Watcher
func newWatcher(client *kubernetes.Clientset, service string) (r result.Result[discovery.Watcher]) {
	watcher := assert.Must1(client.CoreV1().Endpoints(k8sutil.Namespace()).
		Watch(context.Background(),
			metav1.ListOptions{FieldSelector: fmt.Sprintf("%s=%s", "metadata.name", service)}))
	return r.WithVal(&Watcher{watcher: watcher, client: client, service: service})
}

func (t *Watcher) Watch(ctx context.Context, host string) (<-chan watch.Event, chan struct{}, error) {
	ev := make(chan watch.Event)

	watchList := cache.NewListWatchFromClient(new(kubernetes.Clientset).CoreV1().RESTClient(), "endpoints", t.namespace, fields.OneTermEqualSelector("metadata.name", host))
	_, controller := cache.NewInformer(watchList, &v1.Endpoints{}, time.Second*5, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ev <- watch.Event{Type: watch.Added, Object: obj.(runtime.Object)}
		},
		DeleteFunc: func(obj interface{}) {
			ev <- watch.Event{Type: watch.Deleted, Object: obj.(runtime.Object)}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			ev <- watch.Event{Type: watch.Modified, Object: newObj.(runtime.Object)}
		},
	})

	stop := make(chan struct{})

	go controller.Run(stop)

	return ev, stop, nil
}

// Next will block until ServiceInstance changes
func (t *Watcher) Next() result.Result[*discovery.Result] {
	select {
	case _, ok := <-t.watcher.ResultChan():
		if ok {
			endpoints := assert.Must1(t.client.
				CoreV1().
				Endpoints(k8sutil.Namespace()).
				Get(context.Background(), "name", metav1.GetOptions{}))

			var resp = &discovery.Result{
				Action: eventpbv1.EventType_UPDATE,
				Service: &service.Service{
					Name: t.service,
				},
			}

			for _, endpoint := range endpoints.Items {
				for _, subset := range endpoint.Subsets {
					realPort := ""
					for _, p := range subset.Ports {
						realPort = fmt.Sprint(p.Port)
						break
					}

					for _, addr := range subset.Addresses {
						resp.Service.Nodes = append(resp.Service.Nodes, &service.Node{
							Id:      string(addr.TargetRef.UID),
							Address: fmt.Sprintf("%s:%s", addr.IP, realPort),
						})
					}
				}
			}

			return result.OK(resp)
		}

		return result.Wrap(new(discovery.Result), discovery.ErrWatcherStopped)
	}
}

// Stop is used to close the iterator
func (t *Watcher) Stop() error {
	t.watcher.Stop()
	return nil
}
