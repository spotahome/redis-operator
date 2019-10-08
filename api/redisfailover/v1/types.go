package v1

import (
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RedisFailover represents a Redis failover
type RedisFailover struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RedisFailoverSpec `json:"spec"`
}

// RedisFailoverSpec represents a Redis failover spec
type RedisFailoverSpec struct {
	Redis    RedisSettings    `json:"redis,omitempty"`
	Sentinel SentinelSettings `json:"sentinel,omitempty"`
}

// RedisSettings defines the specification of the redis cluster
type RedisSettings struct {
	Image             string                      `json:"image,omitempty"`
	ImagePullPolicy   ImagePullPolicy             `json:"imagePullPolicy,omitempty"`
	Replicas          int32                       `json:"replicas,omitempty"`
	Resources         corev1.ResourceRequirements `json:"resources,omitempty"`
	CustomConfig      []string                    `json:"customConfig,omitempty"`
	Command           []string                    `json:"command,omitempty"`
	ShutdownConfigMap string                      `json:"shutdownConfigMap,omitempty"`
	Storage           RedisStorage                `json:"storage,omitempty"`
	Exporter          RedisExporter               `json:"exporter,omitempty"`
	Affinity          *corev1.Affinity            `json:"affinity,omitempty"`
	SecurityContext   *corev1.PodSecurityContext  `json:"securityContext,omitempty"`
	Tolerations       []corev1.Toleration         `json:"tolerations,omitempty"`
	PodAnnotations    map[string]string           `json:"podAnnotations,omitempty"`
}

// SentinelSettings defines the specification of the sentinel cluster
type SentinelSettings struct {
	Image           string                      `json:"image,omitempty"`
	ImagePullPolicy ImagePullPolicy             `json:"imagePullPolicy,omitempty"`
	Replicas        int32                       `json:"replicas,omitempty"`
	Resources       corev1.ResourceRequirements `json:"resources,omitempty"`
	CustomConfig    []string                    `json:"customConfig,omitempty"`
	Command         []string                    `json:"command,omitempty"`
	Affinity        *corev1.Affinity            `json:"affinity,omitempty"`
	SecurityContext *corev1.PodSecurityContext  `json:"securityContext,omitempty"`
	Tolerations     []corev1.Toleration         `json:"tolerations,omitempty"`
	PodAnnotations  map[string]string           `json:"podAnnotations,omitempty"`
}

// RedisExporter defines the specification for the redis exporter
type RedisExporter struct {
	Enabled         bool            `json:"enabled,omitempty"`
	Image           string          `json:"image,omitempty"`
	ImagePullPolicy ImagePullPolicy `json:"imagePullPolicy,omitempty"`
}

// RedisStorage defines the structure used to store the Redis Data
type RedisStorage struct {
	KeepAfterDeletion     bool                          `json:"keepAfterDeletion,omitempty"`
	EmptyDir              *corev1.EmptyDirVolumeSource  `json:"emptyDir,omitempty"`
	PersistentVolumeClaim *corev1.PersistentVolumeClaim `json:"persistentVolumeClaim,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RedisFailoverList represents a Redis failover list
type RedisFailoverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []RedisFailover `json:"items"`
}

// ImagePullPolicy defines the pull policy with a default of Always
type ImagePullPolicy corev1.PullPolicy

// UnmarshalJSON sets the default value to Always on JSON decode
func (e *ImagePullPolicy) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s == "" {
		*e = ImagePullPolicy("Always")
	} else {
		*e = ImagePullPolicy(s)
	}
	return nil
}
