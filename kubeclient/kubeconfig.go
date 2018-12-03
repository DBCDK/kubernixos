package kubeclient

import (
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

func GetKubeConfig(server string) (*rest.Config, error) {
	var kubeconfig string

	// Honor the KUBECONFIG env var while still being able to fallback to ~/.kube/config
	if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	config, err := clientcmd.BuildConfigFromFlags(server, kubeconfig)
	if err != nil {
		return nil, err
	}

	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
	return config, nil
}

func GetKubeClient(config *rest.Config) (*kubernetes.Clientset, error) {
	clients, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clients, nil
}
