package chaos_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/spotahome/kooper/log"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"

	chaosv1alpha1 "github.com/spotahome/kooper/examples/pod-terminator-operator/apis/chaos/v1alpha1"
	"github.com/spotahome/kooper/examples/pod-terminator-operator/service/chaos"
)

type timeMock struct {
	after chan time.Time
	now   time.Time
}

func (t *timeMock) After(_ time.Duration) <-chan time.Time { return t.after }
func (t *timeMock) Now() time.Time                         { return t.now }

func TestPodKillerSameSpec(t *testing.T) {
	defaultPT := &chaosv1alpha1.PodTerminator{
		Spec: chaosv1alpha1.PodTerminatorSpec{
			PeriodSeconds:      300,
			MinimumInstances:   2,
			TerminationPercent: 50,
			Selector: map[string]string{
				"label1": "value1",
				"label2": "label2",
			},
			DryRun: true,
		},
	}
	tests := []struct {
		name  string
		pt    *chaosv1alpha1.PodTerminator
		newPT *chaosv1alpha1.PodTerminator

		expRes bool
	}{
		{
			name:   "Giving the same podTerminator should return that are the same.",
			pt:     defaultPT,
			newPT:  defaultPT,
			expRes: true,
		},
		{
			name: "Giving the same podTerminator should return that are the same.",
			pt:   defaultPT,
			newPT: &chaosv1alpha1.PodTerminator{
				Spec: chaosv1alpha1.PodTerminatorSpec{
					PeriodSeconds:      300,
					MinimumInstances:   2,
					TerminationPercent: 50,
					Selector: map[string]string{
						"label1": "value1",
						"label2": "label2",
						"label3": "label3",
					},
					DryRun: true,
				},
			},
			expRes: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mocks.
			mk8s := fake.NewSimpleClientset()

			// Call the logic to test.
			pk := chaos.NewPodKiller(test.pt, mk8s, log.Dummy)
			gotRes := pk.SameSpec(test.newPT)

			// Check.
			assert.Equal(test.expRes, gotRes)
		})
	}
}

func getProbableTargets(n int) *corev1.PodList {
	targets := make([]corev1.Pod, n)

	for i := 0; i < n; i++ {
		targets[i] = corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("podtarget-%d", i),
			},
		}
	}

	return &corev1.PodList{
		Items: targets,
	}
}

func onKubeClientListPods(client *fake.Clientset, pdl *corev1.PodList, callChan chan struct{}) {
	client.AddReactor("list", "pods", func(action kubetesting.Action) (bool, runtime.Object, error) {
		// Notify that was called.
		go func() { callChan <- struct{}{} }()
		return true, pdl, nil
	})
}

func TestPodKillerPodKill(t *testing.T) {
	tests := []struct {
		name            string
		pt              *chaosv1alpha1.PodTerminator
		probableTargets *corev1.PodList

		expDeletions int
	}{
		{
			name: "Having a quantity of probable targets and meeting the minimum instances and not dry run mode it should kill the desired percent targets.",
			pt: &chaosv1alpha1.PodTerminator{
				Spec: chaosv1alpha1.PodTerminatorSpec{
					MinimumInstances:   2,
					TerminationPercent: 50,
					DryRun:             false,
				},
			},
			probableTargets: getProbableTargets(10),
			expDeletions:    5,
		},
		{
			name: "Having a quantity of probable targets and meeting the minimum instances in dry run mode it should not kill any target.",
			pt: &chaosv1alpha1.PodTerminator{
				Spec: chaosv1alpha1.PodTerminatorSpec{
					MinimumInstances:   2,
					TerminationPercent: 50,
					DryRun:             true,
				},
			},
			probableTargets: getProbableTargets(10),
			expDeletions:    0,
		},
		{
			name: "Having a quantity of probable targets and not meeting the minimum instances and not dry run mode it should kill less targets that the ones we wanted.",
			pt: &chaosv1alpha1.PodTerminator{
				Spec: chaosv1alpha1.PodTerminatorSpec{
					MinimumInstances:   7,
					TerminationPercent: 50,
					DryRun:             false,
				},
			},
			probableTargets: getProbableTargets(10),
			expDeletions:    3,
		},
		{
			name: "With no targets should not kill anything.",
			pt: &chaosv1alpha1.PodTerminator{
				Spec: chaosv1alpha1.PodTerminatorSpec{
					MinimumInstances:   2,
					TerminationPercent: 50,
					DryRun:             false,
				},
			},
			probableTargets: getProbableTargets(0),
			expDeletions:    0,
		},
		{
			name: "The 25% of 3 is not an instance, it should not kill anything.",
			pt: &chaosv1alpha1.PodTerminator{
				Spec: chaosv1alpha1.PodTerminatorSpec{
					MinimumInstances:   2,
					TerminationPercent: 25,
					DryRun:             false,
				},
			},
			probableTargets: getProbableTargets(3),
			expDeletions:    0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock time
			mtimeC := make(chan time.Time, 1)
			mtimeC <- time.Now()
			mtime := &timeMock{after: mtimeC}

			// Mocks kubernetes.
			listPodcalled := make(chan struct{})
			mk8s := &fake.Clientset{}
			onKubeClientListPods(mk8s, test.probableTargets, listPodcalled)

			// Call the logic to test.
			pk := chaos.NewCustomPodKiller(test.pt, mk8s, mtime, log.Dummy)
			err := pk.Start()
			defer pk.Stop()
			if assert.NoError(err) {
				select {
				case <-time.After(100 * time.Millisecond):
					assert.Fail("timeout waiting for pod list call")
					return
				case <-listPodcalled:
					// Ready to check.
				}
				// Wait until pod killing
				<-time.After(5 * time.Millisecond)

				// Get deletions.
				var gotDeletions int
				for _, act := range mk8s.Actions() {
					if act.Matches("delete", "pods") {
						gotDeletions++
					}
				}
				assert.Equal(test.expDeletions, gotDeletions)
			}
		})
	}
}
