package failover

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
)

// CPUAndMem defines how many cpu and ram the container will request/limit
type CPUAndMem struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// RedisFailoverResources sets the limits and requests for a container
type RedisFailoverResources struct {
	Requests CPUAndMem `json:"requests,omitempty"`
	Limits   CPUAndMem `json:"limits,omitempty"`
}

// RedisSettings defines the specification of the redis cluster
type RedisSettings struct {
	Replicas  int32                  `json:"replicas,omitempty"`
	Resources RedisFailoverResources `json:"resources,omitempty"`
	Exporter  bool                   `json:"exporter,omitempty"`
	Version   string                 `json:"version,omitempty"`
}

// SentinelSettings defines the specification of the sentinel cluster
type SentinelSettings struct {
	Replicas  int32 `json:"replicas,omitempty"`
	quorum    int32
	Resources RedisFailoverResources `json:"resources,omitempty"`
}

// RedisFailoverSpec represents a Redis failover spec
type RedisFailoverSpec struct {
	// Redis defines its failover settings
	Redis RedisSettings `json:"redis,omitempty"`

	// Sentinel defines its failover settings
	Sentinel SentinelSettings `json:"sentinel,omitempty"`
}

// RedisFailover represents a Redis failover
type RedisFailover struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ObjectMeta   `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Kind            string              `json:"kind"`
	APIVersion      string              `json:"apiVersion"`
	Spec            RedisFailoverSpec   `json:"spec"`
	Status          RedisFailoverStatus `json:"status,omitempty"`
}

// RedisFailoverList represents a Redis failover list
type RedisFailoverList struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Kind            string          `json:"kind"`
	APIVersion      string          `json:"apiVersion"`
	Items           []RedisFailover `json:"items"`
}

// GetQuorum returns the quorum according to the sentinel instances
func (f *RedisFailover) GetQuorum() int32 {
	if f.Spec.Sentinel.quorum == 0 {
		f.Spec.Sentinel.quorum = f.Spec.Sentinel.Replicas/2 + 1
	}
	return f.Spec.Sentinel.quorum
}

// GetObjectKind satisfies Object interface
func (f *RedisFailover) GetObjectKind() schema.ObjectKind {
	return &f.TypeMeta
}

// GetObjectMeta satisfies ObjectMetaAccessor interface
func (f *RedisFailover) GetObjectMeta() metav1.Object {
	return &f.Metadata
}

// GetObjectKind satisfies Object interface
func (rl *RedisFailoverList) GetObjectKind() schema.ObjectKind {
	return &rl.TypeMeta
}

// GetListMeta satisfies ListMetaAccessor interface
func (rl *RedisFailoverList) GetListMeta() metav1.List {
	return &rl.Metadata
}

// Transformer implements the conversion between statefulsets and deployments
// into redis/sentinel settings
type Transformer interface {
	StatefulsetToRedisSettings(*v1beta1.StatefulSet) (*RedisSettings, error)
	DeploymentToSentinelSettings(*v1beta1.Deployment) (*SentinelSettings, error)
}

// RedisFailoverTransformer defines the data structure
type RedisFailoverTransformer struct{}

// StatefulsetToRedisSettings transforms from statefulset to redis settings
func (r *RedisFailoverTransformer) StatefulsetToRedisSettings(ss *v1beta1.StatefulSet) (*RedisSettings, error) {
	if lc := len(ss.Spec.Template.Spec.Containers); lc != 1 {
		return nil, fmt.Errorf("The number of containers is not 1, have: %d", lc)
	}
	statefulset := &RedisSettings{
		Replicas: *ss.Spec.Replicas,
		Resources: RedisFailoverResources{
			Limits:   getCPUAndMem(ss.Spec.Template.Spec.Containers[0].Resources.Limits),
			Requests: getCPUAndMem(ss.Spec.Template.Spec.Containers[0].Resources.Requests),
		},
	}
	return statefulset, nil
}

// DeploymentToSentinelSettings transforms from deployment to sentinel settings
func (r *RedisFailoverTransformer) DeploymentToSentinelSettings(d *v1beta1.Deployment) (*SentinelSettings, error) {
	if lc := len(d.Spec.Template.Spec.Containers); lc != 1 {
		return nil, fmt.Errorf("The number of containers is not 1, have: %d", lc)
	}
	deployment := &SentinelSettings{
		Replicas: *d.Spec.Replicas,
		Resources: RedisFailoverResources{
			Limits:   getCPUAndMem(d.Spec.Template.Spec.Containers[0].Resources.Limits),
			Requests: getCPUAndMem(d.Spec.Template.Spec.Containers[0].Resources.Requests),
		},
	}
	return deployment, nil
}

func getCPUAndMem(resources v1.ResourceList) CPUAndMem {
	result := CPUAndMem{}
	if cpu := resources.Cpu().String(); cpu != "0" {
		result.CPU = cpu
	}
	if memory := resources.Memory().String(); memory != "0" {
		result.Memory = memory
	}
	return result
}
