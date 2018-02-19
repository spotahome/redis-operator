package chaos

import (
	"sync"

	"k8s.io/client-go/kubernetes"

	chaosv1alpha1 "github.com/spotahome/kooper/examples/pod-terminator-operator/apis/chaos/v1alpha1"
	"github.com/spotahome/kooper/examples/pod-terminator-operator/log"
)

// Syncer is the interface that every chaos service implementation
// needs to implement.
type Syncer interface {
	// EnsurePodTerminator will ensure that the pod terminator is running and working.
	EnsurePodTerminator(pt *chaosv1alpha1.PodTerminator) error
	// DeletePodTerminator will stop and delete the pod terminator.
	DeletePodTerminator(name string) error
}

// Chaos is the service that will ensure that the desired pod terminator CRDs are met.
// Chaos will have running instances of PodDestroyers.
type Chaos struct {
	k8sCli kubernetes.Interface
	reg    sync.Map
	logger log.Logger
}

// NewChaos returns a new Chaos service.
func NewChaos(k8sCli kubernetes.Interface, logger log.Logger) *Chaos {
	return &Chaos{
		k8sCli: k8sCli,
		reg:    sync.Map{},
		logger: logger,
	}
}

// EnsurePodTerminator satisfies ChaosSyncer interface.
func (c *Chaos) EnsurePodTerminator(pt *chaosv1alpha1.PodTerminator) error {
	pkt, ok := c.reg.Load(pt.Name)
	var pk *PodKiller

	// We are already running.
	if ok {
		pk = pkt.(*PodKiller)
		// If not the same spec means options have changed, so we don't longer need this pod killer.
		if !pk.SameSpec(pt) {
			c.logger.Infof("spec of %s changed, recreating pod killer", pt.Name)
			if err := c.DeletePodTerminator(pt.Name); err != nil {
				return err
			}
		} else { // We are ok, nothing changed.
			return nil
		}
	}

	// Create a pod killer.
	ptCopy := pt.DeepCopy()
	pk = NewPodKiller(ptCopy, c.k8sCli, c.logger)
	c.reg.Store(pt.Name, pk)
	return pk.Start()
	// TODO: garbage collection.
}

// DeletePodTerminator satisfies ChaosSyncer interface.
func (c *Chaos) DeletePodTerminator(name string) error {
	pkt, ok := c.reg.Load(name)
	if !ok {
		return nil
	}

	pk := pkt.(*PodKiller)
	if err := pk.Stop(); err != nil {
		return err
	}

	c.reg.Delete(name)
	return nil
}
