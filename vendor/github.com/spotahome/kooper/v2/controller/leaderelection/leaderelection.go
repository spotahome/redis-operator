package leaderelection

import (
	"context"
	"fmt"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"

	"github.com/spotahome/kooper/v2/log"
)

const (
	defLeaseDuration = 15 * time.Second
	defRenewDeadline = 10 * time.Second
	defRetryPeriod   = 2 * time.Second
)

// LockConfig is the configuration for the lock (timing, leases...).
type LockConfig struct {
	// LeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership. This is measured against time of
	// last observed ack.
	LeaseDuration time.Duration
	// RenewDeadline is the duration that the acting master will retry
	// refreshing leadership before giving up.
	RenewDeadline time.Duration
	// RetryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions.
	RetryPeriod time.Duration
}

// Runner knows how to run using the leader election.
type Runner interface {
	// Run will run if the instance takes the lead. It's a blocking action.
	Run(func() error) error
}

// runner is the leader election default implementation.
type runner struct {
	key          string
	namespace    string
	k8scli       kubernetes.Interface
	lockCfg      *LockConfig
	resourceLock resourcelock.Interface
	logger       log.Logger
}

// NewDefault returns a new leader election service with a safe lock configuration.
func NewDefault(key, namespace string, k8scli kubernetes.Interface, logger log.Logger) (Runner, error) {
	return New(key, namespace, nil, k8scli, logger)
}

// New returns a new leader election service.
func New(key, namespace string, lockCfg *LockConfig, k8scli kubernetes.Interface, logger log.Logger) (Runner, error) {
	// If lock configuration is nil then fallback to defaults.
	if lockCfg == nil {
		lockCfg = &LockConfig{
			LeaseDuration: defLeaseDuration,
			RenewDeadline: defRenewDeadline,
			RetryPeriod:   defRetryPeriod,
		}
	}

	r := &runner{
		lockCfg:   lockCfg,
		key:       key,
		namespace: namespace,
		k8scli:    k8scli,
		logger: logger.WithKV(log.KV{
			"source-service":     "kooper/leader-election",
			"leader-election-id": fmt.Sprintf("%s/%s", namespace, key),
		}),
	}

	if err := r.validate(); err != nil {
		return nil, err
	}

	if err := r.initResourceLock(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *runner) validate() error {
	// Error if no namespace set.
	if r.namespace == "" {
		return fmt.Errorf("running in leader election mode requires the namespace running")
	}
	// Key required
	if r.key == "" {
		return fmt.Errorf("running in leader election mode requires a key for identification the different instances")
	}
	return nil
}

func (r *runner) initResourceLock() error {
	// Create the lock resource for the leader election.
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	id := hostname + "_" + string(uuid.NewUUID())

	eventBroadcaster := record.NewBroadcaster()
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: r.key, Host: id})

	rl, err := resourcelock.New(
		resourcelock.ConfigMapsResourceLock,
		r.namespace,
		r.key,
		r.k8scli.CoreV1(),
		r.k8scli.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: recorder,
		},
	)
	if err != nil {
		return fmt.Errorf("error creating lock: %v", err)
	}

	r.resourceLock = rl
	return nil

}

func (r *runner) Run(f func() error) error {
	errC := make(chan error, 1) // Channel to get the function returning error.

	// The function to execute when leader acquired.
	lef := func(ctx context.Context) {
		r.logger.Infof("lead acquire, starting...")
		// Wait until f finishes or leader elector runner stops.
		select {
		case <-ctx.Done():
			errC <- nil
		case errC <- f():
		}
		r.logger.Infof("lead execution stopped")
	}

	// Create the leader election configuration
	lec := leaderelection.LeaderElectionConfig{
		Lock:          r.resourceLock,
		LeaseDuration: r.lockCfg.LeaseDuration,
		RenewDeadline: r.lockCfg.RenewDeadline,
		RetryPeriod:   r.lockCfg.RetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: lef,
			OnStoppedLeading: func() {
				errC <- fmt.Errorf("leadership lost")
			},
		},
	}

	// Create the leader elector.
	le, err := leaderelection.NewLeaderElector(lec)
	if err != nil {
		return fmt.Errorf("error creating leader election: %s", err)
	}

	// Execute!
	r.logger.Infof("running in leader election mode, waiting to acquire leadership...")
	go le.Run(context.TODO())

	// Wait until stopping the execution returns the result.
	err = <-errC
	return err
}
