package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=spidermans

// Spiderman is a superhero
type Spiderman struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec is the Spiderman spec.
	Spec SpidermanSpec `json:"spec,omitempty"`
}

// SpidermanSpec contains the specification for Spiderman.
type SpidermanSpec struct {
	Name string `json:"name"`
	Year int    `json:"year,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpidermanList is a collection of spidermans.
type SpidermanList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Spiderman `json:"items"`
}
