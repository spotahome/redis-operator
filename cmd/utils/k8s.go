package utils

import (
	"fmt"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	redisfailoverclientset "github.com/spotahome/redis-operator/client/k8s/clientset/versioned"
)

const (
	defCliQPS   = 100
	defCliBurst = 100
)

// LoadKubernetesConfig loads kubernetes configuration based on flags.
func LoadKubernetesConfig(flags *CMDFlags) (*rest.Config, error) {
	var cfg *rest.Config
	// If devel mode then use configuration flag path.
	if flags.Development {
		config, err := clientcmd.BuildConfigFromFlags("", flags.KubeConfig)
		if err != nil {
			return nil, fmt.Errorf("could not load configuration: %s", err)
		}
		cfg = config
	} else {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("error loading kubernetes configuration inside cluster, check app is running outside kubernetes cluster or run in development mode: %s", err)
		}
		cfg = config
	}

	cfg.QPS = defCliQPS
	cfg.Burst = defCliBurst

	return cfg, nil
}

// CreateKubernetesClients create the clients to connect to kubernetes
func CreateKubernetesClients(flags *CMDFlags) (kubernetes.Interface, redisfailoverclientset.Interface, apiextensionsclientset.Interface, error) {
	config, err := LoadKubernetesConfig(flags)
	if err != nil {
		return nil, nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}
	customClientset, err := redisfailoverclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	aeClientset, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	return clientset, customClientset, aeClientset, nil
}
