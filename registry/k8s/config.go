package k8s

import (
	"github.com/pubgo/xerror"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const name = "k8s"

type config struct {
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

func (t *config) Build() *kubernetes.Clientset {
	var config *rest.Config
	if t.Filename != "" {
		config = xerror.PanicErr(clientcmd.LoadFromFile(t.Filename)).(*rest.Config)
	}

	if t.KubeConfig != "" {
		config = xerror.PanicErr(clientcmd.BuildConfigFromFlags(t.Master, t.KubeConfig)).(*rest.Config)
	} else {
		config = xerror.PanicErr(rest.InClusterConfig()).(*rest.Config)
	}


	client := xerror.PanicErr(kubernetes.NewForConfig(config)).(*kubernetes.Clientset)
	return client
}
