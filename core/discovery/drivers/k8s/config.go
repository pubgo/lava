package k8s

import (
	"github.com/pubgo/funk/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const name = "k8s"

type Cfg struct {
	// kube namespace
	Namespace string
	// kube labelSelector example `app=test`
	LabelSelector string
	// kube fieldSelector example `app=test`
	FieldSelector string
	// set KubeConfig out-of-cluster Use outside cluster
	KubeConfig string
	// set master url
	Master   string
	Filename string
}

func (t *Cfg) Build() *kubernetes.Clientset {
	var config *rest.Config
	if t.KubeConfig != "" {
		config = assert.Must1(clientcmd.BuildConfigFromFlags(t.Master, t.KubeConfig))
	} else {
		config = assert.Must1(rest.InClusterConfig())
	}

	client := assert.Must1(kubernetes.NewForConfig(config))
	return client
}
