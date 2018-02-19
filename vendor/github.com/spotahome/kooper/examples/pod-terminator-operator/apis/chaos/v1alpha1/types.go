package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodTerminator represents a pod terminator.
type PodTerminator struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the ddesired behaviour of the pod terminator.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec PodTerminatorSpec `json:"spec,omitempty"`
}

// PodTerminatorSpec is the spec for a PodTerminator resource.
type PodTerminatorSpec struct {
	// Selector is how the target will be selected.
	Selector map[string]string `json:"selector,omitempty"`
	// PeriodSeconds is how often (in seconds) to perform the attack.
	PeriodSeconds int32 `json:"periodSeconds,omitempty"`
	// TerminationPercent is the percent of pods that will be killed randomly.
	TerminationPercent int32 `json:"terminationPercent,omitempty"`
	// MinimumInstances is the number of minimum instances that need to be alive.
	// +optional
	MinimumInstances int32 `json:"minimumInstances,omitempty"`
	// DryRun will set the killing in dryrun mode or not.
	// +optional
	DryRun bool `json:"dryRun,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodTerminatorList is a list of PodTerminator resources
type PodTerminatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PodTerminator `json:"items"`
}
