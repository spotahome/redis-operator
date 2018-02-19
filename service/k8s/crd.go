package k8s

import (
	"github.com/spotahome/kooper/client/crd"
	apiextensionscli "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	"github.com/spotahome/redis-operator/log"
)

type CRDConf = crd.Conf

// CRD is the CRD service that knows how to interact with k8s to manage them.
type CRD interface {
	// CreateWorkspaceCRD will create the custom resource and wait to be ready.
	EnsureCRD(conf CRDConf) error
}

// CRDService is the CRD service implementation using API calls to kubernetes.
type CRDService struct {
	crdCli crd.Interface
	logger log.Logger
}

// NewCRDService returns a new CRD KubeService.
func NewCRDService(aeClient apiextensionscli.Interface, logger log.Logger) *CRDService {
	logger = logger.With("service", "k8s.crd")
	crdCli := crd.NewClient(aeClient, logger)

	return &CRDService{
		crdCli: crdCli,
		logger: logger,
	}
}

// EnsureCRD satisfies RedisFailover.Service interface.
func (c *CRDService) EnsureCRD(conf CRDConf) error {
	return c.crdCli.EnsurePresent(conf)
}
