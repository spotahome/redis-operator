package chaos

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	chaosv1alpha1 "github.com/spotahome/kooper/examples/pod-terminator-operator/apis/chaos/v1alpha1"
	"github.com/spotahome/kooper/examples/pod-terminator-operator/log"
)

// TimeWrapper is a wrapper around time so it can be mocked
type TimeWrapper interface {
	// After is the same as time.After
	After(d time.Duration) <-chan time.Time
	// Now is the same as Now.
	Now() time.Time
}

type timeStd struct{}

func (t *timeStd) After(d time.Duration) <-chan time.Time { return time.After(d) }
func (t *timeStd) Now() time.Time                         { return time.Now() }

// PodKiller will kill pods at regular intervals.
type PodKiller struct {
	pt     *chaosv1alpha1.PodTerminator
	k8sCli kubernetes.Interface
	logger log.Logger
	time   TimeWrapper

	running bool
	mutex   sync.Mutex
	stopC   chan struct{}
}

// NewPodKiller returns a new pod killer.
func NewPodKiller(pt *chaosv1alpha1.PodTerminator, k8sCli kubernetes.Interface, logger log.Logger) *PodKiller {
	return &PodKiller{
		pt:     pt,
		k8sCli: k8sCli,
		logger: logger,
		time:   &timeStd{},
	}
}

// NewCustomPodKiller is a constructor that lets you customize everything on the object construction.
func NewCustomPodKiller(pt *chaosv1alpha1.PodTerminator, k8sCli kubernetes.Interface, time TimeWrapper, logger log.Logger) *PodKiller {
	return &PodKiller{
		pt:     pt,
		k8sCli: k8sCli,
		logger: logger,
		time:   time,
	}
}

// SameSpec checks if the pod killer has the same spec.
func (p *PodKiller) SameSpec(pt *chaosv1alpha1.PodTerminator) bool {
	return reflect.DeepEqual(p.pt.Spec, pt.Spec)
}

// Start will run the pod killer at regular intervals.
func (p *PodKiller) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.running {
		return fmt.Errorf("already running")
	}

	p.stopC = make(chan struct{})
	p.running = true

	go func() {
		p.logger.Infof("started %s pod killer", p.pt.Name)
		if err := p.run(); err != nil {
			p.logger.Errorf("error executing pod killer: %s", err)
		}
	}()

	return nil
}

// Stop stops the pod killer.
func (p *PodKiller) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.running {
		close(p.stopC)
		p.logger.Infof("stopped %s pod killer", p.pt.Name)
	}

	p.running = false
	return nil
}

// run will run the loop that will kill eventually the required pods.
func (p *PodKiller) run() error {
	for {
		select {
		case <-p.time.After(time.Duration(p.pt.Spec.PeriodSeconds) * time.Second):
			if err := p.kill(); err != nil {
				p.logger.Errorf("error terminating pods: %s", err)
			}
		case <-p.stopC:
			return nil
		}
	}
}

// kill will terminate the required pods.
func (p *PodKiller) kill() error {
	// Get all probable targets.
	pods, err := p.getProbableTargets()
	if err != nil {
		return err
	}

	// Select the number to delete and check is safe to terminate.
	total := len(pods.Items)
	if total == 0 {
		p.logger.Warningf("0 pods probable targets")
	}

	totalTargets := int(p.pt.Spec.TerminationPercent) * total / 100

	// Check if we met the minimum after killing and adjust the targets to met the minimum instance requirement.
	if (total - totalTargets) < int(p.pt.Spec.MinimumInstances) {
		totalTargets = total - int(p.pt.Spec.MinimumInstances)
		if totalTargets < 0 {
			totalTargets = 0
		}
		p.logger.Infof("%d minimum will not be met after killing, only killing %d targets", p.pt.Spec.MinimumInstances, totalTargets)
	}

	// Get random pods.
	targets := p.getRandomTargets(pods, totalTargets)
	p.logger.Infof("%s pod killer will kill %d targets", p.pt.Name, len(targets))
	// Terminate.
	for _, target := range targets {
		if p.pt.Spec.DryRun {
			p.logger.Infof("pod %s not killed (dry run)", target.Name)
		} else {
			// if any error the abort deletion.
			if err := p.k8sCli.CoreV1().Pods(target.Namespace).Delete(target.Name, &metav1.DeleteOptions{}); err != nil {
				return err
			}
			p.logger.Infof("pod %s killed", target.Name)
		}
	}

	return nil
}

// Gets all the pods filtered that can be a target of termination.
func (p *PodKiller) getProbableTargets() (*corev1.PodList, error) {
	set := labels.Set(p.pt.Spec.Selector)
	slc := set.AsSelector()
	opts := metav1.ListOptions{
		LabelSelector: slc.String(),
	}
	return p.k8sCli.CoreV1().Pods("").List(opts)
}

// getRandomTargets will get the targets randomly.
func (p *PodKiller) getRandomTargets(pods *corev1.PodList, q int) []corev1.Pod {
	if len(pods.Items) == q {
		return pods.Items
	}

	// TODO: Optimize to pick randomly without duplicates and remove the way of sorting
	// a whole slice.
	// Sort targets randomly.
	randomPods := pods.DeepCopy().Items
	sort.Slice(randomPods, func(_, _ int) bool {
		// Create a random number and check if is even, if its true then a < b.
		r := rand.New(rand.NewSource(p.time.Now().UnixNano()))
		return r.Intn(1000)%2 == 0
	})

	// Get the desired quantity.
	return randomPods[:q]
}
