package utils

import (
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	redisfailoverclientset "github.com/spotahome/redis-operator/client/k8s/clientset/versioned"
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

	cfg.QPS = float32(flags.K8sQueriesPerSecond)
	cfg.Burst = flags.K8sQueriesBurstable

	return cfg, nil
}

// CreateKubernetesClients create the clients to connect to kubernetes
func CreateKubernetesClients(flags *CMDFlags) (kubernetes.Interface, redisfailoverclientset.Interface, apiextensionsclientset.Interface, monitoringv1.MonitoringV1Interface, error) {
	config, err := LoadKubernetesConfig(flags)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	customClientset, err := redisfailoverclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	aeClientset, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	monitoringV1Client, err := monitoringv1.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return clientset, customClientset, aeClientset, monitoringV1Client, nil
}
