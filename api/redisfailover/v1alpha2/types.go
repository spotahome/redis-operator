package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RedisFailover represents a Redis failover
type RedisFailover struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RedisFailoverSpec   `json:"spec"`
	Status            RedisFailoverStatus `json:"status,omitempty"`
}

// RedisFailoverSpec represents a Redis failover spec
type RedisFailoverSpec struct {
	// Redis defines its failover settings
	Redis RedisSettings `json:"redis,omitempty"`

	// Sentinel defines its failover settings
	Sentinel SentinelSettings `json:"sentinel,omitempty"`

	// HardAntiAffinity defines if the PodAntiAffinity on the deployments and
	// statefulsets has to be hard (it's soft by default)
	HardAntiAffinity bool `json:"hardAntiAffinity,omitempty"`

	// NodeAffinity defines the rules for scheduling the Redis and Sentinel
	// nodes
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty"`
}

// RedisSettings defines the specification of the redis cluster
type RedisSettings struct {
	Replicas          int32                  `json:"replicas,omitempty"`
	Resources         RedisFailoverResources `json:"resources,omitempty"`
	Exporter          bool                   `json:"exporter,omitempty"`
	ExporterImage     string                 `json:"exporterImage,omitempty"`
	ExporterVersion   string                 `json:"exporterVersion,omitempty"`
	Image             string                 `json:"image,omitempty"`
	Version           string                 `json:"version,omitempty"`
	CustomConfig      []string               `json:"customConfig,omitempty"`
	ShutdownConfigMap string                 `json:"shutdownConfigMap,omitempty"`
	Storage           RedisStorage           `json:"storage,omitempty"`
}

// SentinelSettings defines the specification of the sentinel cluster
type SentinelSettings struct {
	Replicas     int32                  `json:"replicas,omitempty"`
	Resources    RedisFailoverResources `json:"resources,omitempty"`
	CustomConfig []string               `json:"customConfig,omitempty"`
}

// RedisFailoverResources sets the limits and requests for a container
type RedisFailoverResources struct {
	Requests CPUAndMem `json:"requests,omitempty"`
	Limits   CPUAndMem `json:"limits,omitempty"`
}

// CPUAndMem defines how many cpu and ram the container will request/limit
type CPUAndMem struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// RedisStorage defines the structure used to store the Redis Data
type RedisStorage struct {
	KeepAfterDeletion     bool                          `json:"keepAfterDeletion,omitempty"`
	EmptyDir              *corev1.EmptyDirVolumeSource  `json:"emptyDir,omitempty"`
	PersistentVolumeClaim *corev1.PersistentVolumeClaim `json:"persistentVolumeClaim,omitempty"`
}

// RedisFailoverStatus has the status of the cluster
type RedisFailoverStatus struct {
	Phase      Phase       `json:"phase"`
	Conditions []Condition `json:"conditions"`
	Master     string      `json:"master"`
}

// Phase of the RF status
type Phase string

// Condition saves the state information of the redisfailover
type Condition struct {
	Type           ConditionType `json:"type"`
	Reason         string        `json:"reason"`
	TransitionTime string        `json:"transitionTime"`
}

// ConditionType defines the condition that the RF can have
type ConditionType string

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RedisFailoverList represents a Redis failover list
type RedisFailoverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []RedisFailover `json:"items"`
}
