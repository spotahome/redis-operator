package operator

import (
	"github.com/spotahome/kooper/client/crd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	chaosv1alpha1 "github.com/spotahome/kooper/examples/pod-terminator-operator/apis/chaos/v1alpha1"
	podtermk8scli "github.com/spotahome/kooper/examples/pod-terminator-operator/client/k8s/clientset/versioned"
)

// podTerminatorCRD is the crd pod terminator.
type podTerminatorCRD struct {
	crdCli     crd.Interface
	kubecCli   kubernetes.Interface
	podTermCli podtermk8scli.Interface
}

func newPodTermiantorCRD(podTermCli podtermk8scli.Interface, crdCli crd.Interface, kubeCli kubernetes.Interface) *podTerminatorCRD {
	return &podTerminatorCRD{
		crdCli:     crdCli,
		podTermCli: podTermCli,
		kubecCli:   kubeCli,
	}
}

// podTerminatorCRD satisfies resource.crd interface.
func (p *podTerminatorCRD) Initialize() error {
	crd := crd.Conf{
		Kind:       chaosv1alpha1.PodTerminatorKind,
		NamePlural: chaosv1alpha1.PodTerminatorNamePlural,
		ShortNames: chaosv1alpha1.PodTerminatorShortNames,
		Group:      chaosv1alpha1.SchemeGroupVersion.Group,
		Version:    chaosv1alpha1.SchemeGroupVersion.Version,
		Scope:      chaosv1alpha1.PodTerminatorScope,
		Categories: []string{"chaos", "podterm"},
	}

	return p.crdCli.EnsurePresent(crd)
}

// GetListerWatcher satisfies resource.crd interface (and retrieve.Retriever).
func (p *podTerminatorCRD) GetListerWatcher() cache.ListerWatcher {
	return &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return p.podTermCli.ChaosV1alpha1().PodTerminators().List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return p.podTermCli.ChaosV1alpha1().PodTerminators().Watch(options)
		},
	}
}

// GetObject satisfies resource.crd interface (and retrieve.Retriever).
func (p *podTerminatorCRD) GetObject() runtime.Object {
	return &chaosv1alpha1.PodTerminator{}
}
