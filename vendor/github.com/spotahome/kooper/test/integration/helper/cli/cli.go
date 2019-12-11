package cli

import (
	"fmt"
	"os"
	"path/filepath"

	integrationtestk8scli "github.com/spotahome/kooper/test/integration/operator/client/k8s/clientset/versioned"
	apiextensionscli "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // Load oidc authentication when creating the kubernetes client.
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// GetK8sClients returns a all k8s clients.
// * Kubernetes core resources client.
// * Kubernetes api extensions client.
// * Custom test integration CR client.
func GetK8sClients(kubehome string) (kubernetes.Interface, apiextensionscli.Interface, integrationtestk8scli.Interface, error) {
	// Try fallbacks.
	if kubehome == "" {
		if kubehome = os.Getenv("KUBECONFIG"); kubehome == "" {
			kubehome = filepath.Join(homedir.HomeDir(), ".kube", "config")
		}
	}

	// Load kubernetes local connection.
	config, err := clientcmd.BuildConfigFromFlags("", kubehome)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not load configuration: %s", err)
	}

	// Get the client.
	k8sCli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	// App CRD k8s types client.
	itCli, err := integrationtestk8scli.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	// api extensions cli.
	aexCli, err := apiextensionscli.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	return k8sCli, aexCli, itCli, nil
}
