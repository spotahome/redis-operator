package failover

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	policy "k8s.io/client-go/pkg/apis/policy/v1beta1"

	"github.com/spotahome/redis-operator/pkg/clock"
	"github.com/spotahome/redis-operator/pkg/config"
	"github.com/spotahome/redis-operator/pkg/log"
)

// variables refering to the redis exporter port
const (
	exporterPort     = 9121
	exporterPortName = "http-metrics"
)

const (
	description      = "Manage a Redis Failover deployment"
	baseName         = "rf"
	bootstrapName    = "b"
	sentinelName     = "s"
	sentinelRoleName = "sentinel"
	redisName        = "r"
	redisRoleName    = "redis"
)

const (
	loopInterval     = 5 * time.Second
	redisfailoverAPI = "/apis/%s/%s/namespaces/%s/%s/%s"
)

var (
	redisToolkitImage = fmt.Sprintf("%s:%s", config.RedisToolkitImage, config.RedisToolkitImageVersion)
	exporterImage     = fmt.Sprintf("%s:%s", config.ExporterImage, config.ExporterImageVersion)
)

// RedisFailoverClient has the minimumm methods that a Redis failover controller needs to satisfy
// in order to talk with K8s
type RedisFailoverClient interface {
	UpdateStatus(rFailover *RedisFailover) (*RedisFailover, error)
	GetBootstrapName(rFailover *RedisFailover) string
	GetRedisName(rFailover *RedisFailover) string
	GetSentinelName(rFailover *RedisFailover) string
	GetAllRedisfailovers() (*RedisFailoverList, error)
	GetBootstrapPod(rFailover *RedisFailover) (*v1.Pod, error)
	GetSentinelService(rFailover *RedisFailover) (*v1.Service, error)
	GetSentinelDeployment(rFailover *RedisFailover) (*v1beta1.Deployment, error)
	GetRedisStatefulset(rFailover *RedisFailover) (*v1beta1.StatefulSet, error)
	GetRedisService(rFailover *RedisFailover) (*v1.Service, error)
	GetSentinelPodsIPs(rFailover *RedisFailover) ([]string, error)
	GetRedisPodsIPs(rFailover *RedisFailover) ([]string, error)
	CreateBootstrapPod(rFailover *RedisFailover) error
	CreateSentinelService(rFailover *RedisFailover) error
	CreateSentinelDeployment(rFailover *RedisFailover) error
	CreateRedisStatefulset(rFailover *RedisFailover) error
	CreateRedisService(rFailover *RedisFailover) error
	UpdateSentinelDeployment(rFailover *RedisFailover) error
	UpdateRedisStatefulset(rFailover *RedisFailover) error
	DeleteBootstrapPod(rFailover *RedisFailover) error
	DeleteRedisStatefulset(rFailover *RedisFailover) error
	DeleteSentinelDeployment(rFailover *RedisFailover) error
	DeleteSentinelService(rFailover *RedisFailover) error
	DeleteRedisService(rFailover *RedisFailover) error
}

// RedisFailoverKubeClient implements the required methods to talk with kubernetes
type RedisFailoverKubeClient struct {
	Client kubernetes.Interface
	clock  clock.Clock
	logger log.Logger
}

// NewRedisFailoverKubeClient creates a new RedisFailoverKubeClient
func NewRedisFailoverKubeClient(client kubernetes.Interface, clock clock.Clock, logger log.Logger) *RedisFailoverKubeClient {
	return &RedisFailoverKubeClient{
		Client: client,
		clock:  clock,
		logger: logger,
	}
}

// GetBootstrapName returns the name for bootstrap resources
func (r *RedisFailoverKubeClient) GetBootstrapName(rf *RedisFailover) string {
	return generateName(bootstrapName, rf.Metadata.Name)
}

// GetRedisName returns the name for redis resources
func (r *RedisFailoverKubeClient) GetRedisName(rf *RedisFailover) string {
	return generateName(redisName, rf.Metadata.Name)
}

// GetSentinelName returns the name for sentinel resources
func (r *RedisFailoverKubeClient) GetSentinelName(rf *RedisFailover) string {
	return generateName(sentinelName, rf.Metadata.Name)
}

func generateName(typeName, metaName string) string {
	return fmt.Sprintf("%s%s-%s", baseName, typeName, metaName)
}

func getRedisImage(rf *RedisFailover) string {
	return fmt.Sprintf("%s:%s", config.RedisImage, rf.Spec.Redis.Version)
}

// GetAllRedisfailovers connects to k8s and returns all RF deployed on cluster
func (r *RedisFailoverKubeClient) GetAllRedisfailovers() (*RedisFailoverList, error) {
	uri := fmt.Sprintf("/apis/%s/%s/%s/", config.Domain, config.Version, config.APIName)
	rfs := &RedisFailoverList{}
	d, err := r.Client.Apps().RESTClient().Get().RequestURI(uri).DoRaw()
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(d, rfs); err != nil {
		return nil, fmt.Errorf("read spec from json data failed: %v", err)
	}
	return rfs, nil
}

// UpdateStatus saves the actual status of the RF on the K8S api
func (r *RedisFailoverKubeClient) UpdateStatus(rf *RedisFailover) (*RedisFailover, error) {
	uri := rf.GetObjectMeta().GetSelfLink()
	newRF := &RedisFailover{}
	oldRF, err := r.Client.Apps().RESTClient().Get().RequestURI(uri).DoRaw()
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(oldRF, newRF); err != nil {
		return nil, fmt.Errorf("read spec from json data failed: %v", err)
	}
	newRF.Status = rf.Status
	_, err = r.Client.Apps().RESTClient().Put().RequestURI(uri).Body(newRF).DoRaw()
	if err != nil {
		return nil, err
	}
	b, err := r.Client.Apps().RESTClient().Get().RequestURI(uri).DoRaw()
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(b, newRF); err != nil {
		return nil, fmt.Errorf("read spec from json data failed: %v", err)
	}
	return newRF, nil
}

// GetBootstrapPod connects to k8s and return the pod if it exists
func (r *RedisFailoverKubeClient) GetBootstrapPod(rf *RedisFailover) (*v1.Pod, error) {
	name := r.GetBootstrapName(rf)
	namespace := rf.Metadata.Namespace
	pod, err := r.Client.Core().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.New("Could not get pod")
	}
	return pod, nil
}

// GetSentinelService connects to k8s and returns the sentinel service on it
func (r *RedisFailoverKubeClient) GetSentinelService(rf *RedisFailover) (*v1.Service, error) {
	name := r.GetSentinelName(rf)
	namespace := rf.Metadata.Namespace
	service, err := r.Client.Core().Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.New("Could not get service")
	}
	return service, nil
}

// GetSentinelDeployment connects to k8s and returns the sentinel deployment on it
func (r *RedisFailoverKubeClient) GetSentinelDeployment(rf *RedisFailover) (*v1beta1.Deployment, error) {
	name := r.GetSentinelName(rf)
	namespace := rf.Metadata.Namespace
	deployment, err := r.Client.AppsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.New("Could not get deployment")
	}
	return deployment, nil
}

// GetRedisService connects to k8s and returns the redis service on it
func (r *RedisFailoverKubeClient) GetRedisService(rf *RedisFailover) (*v1.Service, error) {
	name := r.GetRedisName(rf)
	namespace := rf.Metadata.Namespace
	service, err := r.Client.Core().Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.New("Could not get service")
	}
	return service, nil
}

// GetRedisStatefulset connects to k8s and returns the redis statefulset on it
func (r *RedisFailoverKubeClient) GetRedisStatefulset(rf *RedisFailover) (*v1beta1.StatefulSet, error) {
	name := r.GetRedisName(rf)
	namespace := rf.Metadata.Namespace
	statefulset, err := r.Client.AppsV1beta1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.New("Could not get statefulset")
	}
	return statefulset, nil
}

// GetSentinelPodsIPs connects to k8s and returns sentinel pods ip
func (r *RedisFailoverKubeClient) GetSentinelPodsIPs(rf *RedisFailover) ([]string, error) {
	name := r.GetSentinelName(rf)
	namespace := rf.Metadata.Namespace
	endpoints, err := r.Client.Core().Endpoints(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if len(endpoints.Subsets) != 1 {
		return nil, errors.New("The Sentinel Service has different endpoints than expected")
	}
	pods := []string{}
	for _, e := range endpoints.Subsets[0].Addresses {
		pods = append(pods, e.IP)
	}
	return pods, nil
}

// GetRedisPodsIPs connects to k8s and returns redis pods ip
func (r *RedisFailoverKubeClient) GetRedisPodsIPs(rf *RedisFailover) ([]string, error) {
	name := r.GetRedisName(rf)
	namespace := rf.Metadata.Namespace
	pods := []string{}
	for i := 0; i < int(rf.Spec.Redis.Replicas); i++ {
		podName := fmt.Sprintf("%s-%d", name, i)
		pod, err := r.Client.Core().Pods(namespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		pods = append(pods, pod.Status.PodIP)
	}
	return pods, nil
}

// CreateBootstrapPod create the initial pod
func (r *RedisFailoverKubeClient) CreateBootstrapPod(rf *RedisFailover) error {
	name := r.GetBootstrapName(rf)
	namespace := rf.Metadata.Namespace
	quorum := rf.GetQuorum()

	logger := r.logger.WithField(logNameField, rf.Metadata.Name).WithField(logNamespaceField, rf.Metadata.Namespace)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":       "redis-failover",
				"component": "sentinel",
				"sentinel":  rf.Metadata.Name,
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Name:            "redis",
					Image:           redisToolkitImage,
					ImagePullPolicy: "Always",
					Command: []string{
						"redis-server",
					},
					Ports: []v1.ContainerPort{
						v1.ContainerPort{
							Name:          "redis",
							ContainerPort: 6379,
							Protocol:      v1.ProtocolTCP,
						},
					},
					ReadinessProbe: &v1.Probe{
						InitialDelaySeconds: 15,
						TimeoutSeconds:      5,
						Handler: v1.Handler{
							Exec: &v1.ExecAction{
								Command: []string{
									"sh",
									"-c",
									"redis-cli -h $(hostname) ping",
								},
							},
						},
					},
				},
				v1.Container{
					Name:            "sentinel",
					Image:           redisToolkitImage,
					ImagePullPolicy: "Always",
					Command: []string{
						"bootstrap-sentinel",
					},
					Env: []v1.EnvVar{
						v1.EnvVar{
							Name:  "SENTINEL_QUORUM",
							Value: fmt.Sprintf("%d", quorum),
						},
					},
					Ports: []v1.ContainerPort{
						v1.ContainerPort{
							Name:          "sentinel",
							ContainerPort: 26379,
							Protocol:      v1.ProtocolTCP,
						},
					},
					ReadinessProbe: &v1.Probe{
						InitialDelaySeconds: 15,
						TimeoutSeconds:      5,
						Handler: v1.Handler{
							Exec: &v1.ExecAction{
								Command: []string{
									"sh",
									"-c",
									"redis-cli -h $(hostname) -p 26379 ping",
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := r.Client.CoreV1().Pods(namespace).Create(pod)

	if err != nil {
		return err
	}

	t := r.clock.NewTicker(loopInterval)
	for range t.C {
		logger.Debug("Waiting for pod to be ready")
		ready := false
		pod, _ = r.Client.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
		for _, condition := range pod.Status.Conditions {

			if condition.Type == "Ready" && condition.Status == v1.ConditionTrue {
				ready = true
				break
			}
		}
		if ready {
			t.Stop()
			break
		}
	}

	return nil
}

// CreateSentinelService creates the Sentinel service
func (r *RedisFailoverKubeClient) CreateSentinelService(rf *RedisFailover) error {
	name := r.GetSentinelName(rf)
	namespace := rf.Metadata.Namespace

	logger := r.logger.WithField(logNameField, rf.Metadata.Name).WithField(logNamespaceField, rf.Metadata.Namespace)

	sentinelTargetPort := intstr.FromInt(26379)

	sentinelSvc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":       "redis-failover",
				"component": "sentinel",
				"sentinel":  rf.Metadata.Name,
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app":       "redis-failover",
				"component": "sentinel",
				"sentinel":  rf.Metadata.Name,
			},
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "sentinel",
					Port:       26379,
					TargetPort: sentinelTargetPort,
					Protocol:   "TCP",
				},
			},
		},
	}

	if _, err := r.Client.CoreV1().Services(namespace).Create(sentinelSvc); err != nil {
		return err
	}

	t := r.clock.NewTicker(loopInterval)
	for range t.C {
		logger.Debug("Waiting for service to find bootstrap pod")
		endpoints, _ := r.Client.CoreV1().Endpoints(namespace).Get(name, metav1.GetOptions{})
		addresses := 0
		for _, subset := range endpoints.Subsets {
			addresses += len(subset.Addresses)
		}
		if addresses > 0 {
			t.Stop()
			break
		}
	}

	return nil
}

// CreateSentinelDeployment Creates the sentine deployment
func (r *RedisFailoverKubeClient) CreateSentinelDeployment(rf *RedisFailover) error {
	name := r.GetSentinelName(rf)
	namespace := rf.Metadata.Namespace
	spec := rf.Spec
	quorum := rf.GetQuorum()

	logger := r.logger.WithField(logNameField, rf.Metadata.Name).WithField(logNamespaceField, rf.Metadata.Namespace)

	redisImage := getRedisImage(rf)

	resources := getSentinelResources(spec)

	sentinelDeployment := &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &spec.Sentinel.Replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":       "redis-failover",
						"component": "sentinel",
						"sentinel":  rf.Metadata.Name,
					},
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						v1.Container{
							Name:            "sentinel-config",
							Image:           redisToolkitImage,
							ImagePullPolicy: "Always",
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "sentinel-config",
									MountPath: "/redis",
								},
							},
							Command: []string{
								"generate-sentinel-config",
								"/redis/sentinel.conf",
							},
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name:  "REDIS_SENTINEL_HOST",
									Value: r.GetSentinelName(rf),
								},
								v1.EnvVar{
									Name:  "SENTINEL_QUORUM",
									Value: fmt.Sprintf("%d", quorum),
								},
							},
						},
					},
					Containers: []v1.Container{
						v1.Container{
							Name:            "sentinel",
							Image:           redisImage,
							ImagePullPolicy: "Always",
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									Name:          "sentinel",
									ContainerPort: 26379,
									Protocol:      v1.ProtocolTCP,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "sentinel-config",
									MountPath: "/redis",
								},
							},
							Command: []string{
								"redis-server",
								"/redis/sentinel.conf",
								"--sentinel",
							},
							ReadinessProbe: &v1.Probe{
								InitialDelaySeconds: 15,
								TimeoutSeconds:      5,
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											"redis-cli -h $(hostname) -p 26379 ping",
										},
									},
								},
							},
							LivenessProbe: &v1.Probe{
								InitialDelaySeconds: 5,
								TimeoutSeconds:      5,
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											"redis-cli -h $(hostname) -p 26379 ping",
										},
									},
								},
							},
							Resources: resources,
						},
					},
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "sentinel-config",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	if _, err := r.Client.AppsV1beta1().Deployments(namespace).Create(sentinelDeployment); err != nil {
		return err
	}

	t := r.clock.NewTicker(loopInterval)
	for range t.C {
		logger.Debug("Waiting for Sentinel deployment to be fully operative")
		deployment, _ := r.Client.AppsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
		if deployment.Status.ReadyReplicas == spec.Sentinel.Replicas {
			t.Stop()
			break
		}
	}

	logger.Debug("Creating Sentinel PodDisruptionBudget...")
	if err := r.createPodDisruptionBudget(rf, sentinelName, sentinelRoleName); err != nil {
		return err
	}
	logger.Debug("Sentinel PodDisruptionBudget created!")

	return nil
}

// CreateRedisStatefulset Creates the redis server statefulset
func (r *RedisFailoverKubeClient) CreateRedisStatefulset(rf *RedisFailover) error {
	name := r.GetRedisName(rf)
	namespace := rf.Metadata.Namespace
	spec := rf.Spec
	logger := r.logger.WithField(logNameField, rf.Metadata.Name).WithField(logNamespaceField, rf.Metadata.Namespace)

	redisImage := getRedisImage(rf)

	resources := getRedisResources(spec)

	redisStatefulset := &v1beta1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1beta1.StatefulSetSpec{
			ServiceName: name,
			Replicas:    &spec.Redis.Replicas,
			UpdateStrategy: v1beta1.StatefulSetUpdateStrategy{
				Type: "RollingUpdate",
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":       "redis-failover",
						"component": "redis",
						"redis":     rf.Metadata.Name,
					},
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						v1.Container{
							Name:            "redis-config",
							Image:           redisToolkitImage,
							ImagePullPolicy: "Always",
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "server-config",
									MountPath: "/redis",
								},
							},
							Command: []string{
								"generate-server-config",
								"/redis/server.conf",
							},
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name:  "REDIS_SENTINEL_HOST",
									Value: r.GetSentinelName(rf),
								},
							},
						},
					},
					Containers: []v1.Container{
						v1.Container{
							Name:            "redis",
							Image:           redisImage,
							ImagePullPolicy: "Always",
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									Name:          "redis",
									ContainerPort: 6379,
									Protocol:      v1.ProtocolTCP,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "server-config",
									MountPath: "/redis",
								},
							},
							Command: []string{
								"redis-server",
								"/redis/server.conf",
							},
							ReadinessProbe: &v1.Probe{
								InitialDelaySeconds: 15,
								TimeoutSeconds:      5,
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											"redis-cli -h $(hostname) ping",
										},
									},
								},
							},
							LivenessProbe: &v1.Probe{
								InitialDelaySeconds: 5,
								TimeoutSeconds:      5,
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											"redis-cli -h $(hostname) ping",
										},
									},
								},
							},
							Resources: resources,
						},
					},
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "server-config",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	if rf.Spec.Redis.Exporter {
		exporter := v1.Container{
			Name:            "redis-exporter",
			Image:           exporterImage,
			ImagePullPolicy: "Always",
			Ports: []v1.ContainerPort{
				v1.ContainerPort{
					Name:          "metrics",
					ContainerPort: exporterPort,
					Protocol:      v1.ProtocolTCP,
				},
			},
			ReadinessProbe: &v1.Probe{
				InitialDelaySeconds: 10,
				TimeoutSeconds:      3,
				Handler: v1.Handler{
					HTTPGet: &v1.HTTPGetAction{
						Path: "/",
						Port: intstr.FromString("metrics"),
					},
				},
			},
			LivenessProbe: &v1.Probe{
				TimeoutSeconds: 3,
				Handler: v1.Handler{
					HTTPGet: &v1.HTTPGetAction{
						Path: "/",
						Port: intstr.FromString("metrics"),
					},
				},
			},
			Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("300m"),
					v1.ResourceMemory: resource.MustParse("300Mi"),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("200m"),
					v1.ResourceMemory: resource.MustParse("150Mi"),
				},
			},
		}
		redisStatefulset.Spec.Template.Spec.Containers = append(redisStatefulset.Spec.Template.Spec.Containers, exporter)
	}

	if _, err := r.Client.AppsV1beta1().StatefulSets(namespace).Create(redisStatefulset); err != nil {
		return err
	}

	t := r.clock.NewTicker(loopInterval)
	for range t.C {
		logger.Debug("Waiting for Redis statefulset to be fully operative")
		statefulset, _ := r.Client.AppsV1beta1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
		if statefulset.Status.Replicas == spec.Redis.Replicas {
			t.Stop()
			break
		}
	}

	logger.Debug("Creating Redis PodDisruptionBudget...")
	if err := r.createPodDisruptionBudget(rf, redisName, redisRoleName); err != nil {
		return err
	}
	logger.Debug("Redis PodDisruptionBudget created!")

	return nil
}

// CreateRedisService creates a service for redis.
func (r *RedisFailoverKubeClient) CreateRedisService(rf *RedisFailover) error {
	name := r.GetRedisName(rf)
	namespace := rf.Metadata.Namespace
	srv := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: v1.ClusterIPNone,
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Port:     exporterPort,
					Protocol: v1.ProtocolTCP,
					Name:     exporterPortName,
				},
			},
			Selector: map[string]string{
				"app":       "redis-failover",
				"component": "redis",
				"redis":     rf.Metadata.Name,
			},
		},
	}

	_, err := r.Client.CoreV1().Services(namespace).Create(srv)
	return err
}

// createPodDisruptionBudget creates a PodDisruptionBudget for redis or sentinel
func (r *RedisFailoverKubeClient) createPodDisruptionBudget(rf *RedisFailover, name string, role string) error {
	name = generateName(name, rf.Metadata.Name)
	namespace := rf.Metadata.Namespace
	if _, err := r.Client.PolicyV1beta1().PodDisruptionBudgets(namespace).Get(name, metav1.GetOptions{}); err != nil {
		minAvailable := intstr.FromInt(2)
		pdb := &policy.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: policy.PodDisruptionBudgetSpec{
				MinAvailable: &minAvailable,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":       "redis-failover",
						"component": role,
						role:        rf.Metadata.Name,
					},
				},
			},
		}
		_, err := r.Client.PolicyV1beta1().PodDisruptionBudgets(namespace).Create(pdb)
		return err
	}
	return nil
}

// UpdateSentinelDeployment updates the spec of the existing sentinel deployment
func (r *RedisFailoverKubeClient) UpdateSentinelDeployment(rf *RedisFailover) error {
	logger := r.logger.WithField(logNameField, rf.Metadata.Name).WithField(logNamespaceField, rf.Metadata.Namespace)

	quorum := rf.GetQuorum()
	replicas := rf.Spec.Sentinel.Replicas
	initEnv := []v1.EnvVar{
		v1.EnvVar{
			Name:  "REDIS_SENTINEL_HOST",
			Value: r.GetSentinelName(rf),
		},
		v1.EnvVar{
			Name:  "SENTINEL_QUORUM",
			Value: fmt.Sprintf("%d", quorum),
		},
	}

	oldSD, err := r.GetSentinelDeployment(rf)
	if err != nil {
		return err
	}

	oldSD.Spec.Replicas = &replicas
	oldSD.Spec.Template.Spec.InitContainers[0].Env = initEnv
	oldSD.Spec.Template.Spec.Containers[0].Image = getRedisImage(rf)
	oldSD.Spec.Template.Spec.Containers[0].Resources = getSentinelResources(rf.Spec)

	if _, err := r.Client.AppsV1beta1().Deployments(rf.Metadata.Namespace).Update(oldSD); err != nil {
		return err
	}

	t := r.clock.NewTicker(loopInterval)
	for range t.C {
		logger.Debug("Waiting for Sentinel deployment to be updated")
		deployment, _ := r.GetSentinelDeployment(rf)
		if deployment.Status.ReadyReplicas == replicas && deployment.Status.UpdatedReplicas == replicas {
			t.Stop()
			break
		}
	}

	return nil
}

// UpdateRedisStatefulset updates the spec of the existing redis statefulset
func (r *RedisFailoverKubeClient) UpdateRedisStatefulset(rf *RedisFailover) error {
	logger := r.logger.WithField(logNameField, rf.Metadata.Name).WithField(logNamespaceField, rf.Metadata.Namespace)

	replicas := rf.Spec.Redis.Replicas

	oldSS, err := r.GetRedisStatefulset(rf)
	if err != nil {
		return err
	}

	oldSS.Spec.Replicas = &replicas
	oldSS.Spec.Template.Spec.Containers[0].Resources = getRedisResources(rf.Spec)
	oldSS.Spec.Template.Spec.Containers[0].Image = getRedisImage(rf)

	if _, err := r.Client.AppsV1beta1().StatefulSets(rf.Metadata.Namespace).Update(oldSS); err != nil {
		return err
	}

	t := r.clock.NewTicker(loopInterval)
	for range t.C {
		logger.Debug("Waiting for Redis statefulset to be updated")
		statefulset, _ := r.GetRedisStatefulset(rf)
		if statefulset.Status.Replicas == replicas && statefulset.Status.UpdatedReplicas == replicas {
			t.Stop()
			break
		}
	}

	return nil
}

// DeleteBootstrapPod deletes the bootstrapped pod
func (r *RedisFailoverKubeClient) DeleteBootstrapPod(rf *RedisFailover) error {
	name := r.GetBootstrapName(rf)
	namespace := rf.Metadata.Namespace

	logger := r.logger.WithField(logNameField, rf.Metadata.Name).WithField(logNamespaceField, rf.Metadata.Namespace)
	err := r.Client.CoreV1().Pods(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	t := r.clock.NewTicker(loopInterval)
	for range t.C {
		logger.Debug("Waiting for pod to terminate")
		pod, _ := r.Client.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
		if len(pod.Name) == 0 {
			t.Stop()
			break
		}
	}
	return nil
}

// DeleteRedisStatefulset deletes a redis statefulset
func (r *RedisFailoverKubeClient) DeleteRedisStatefulset(rf *RedisFailover) error {
	name := r.GetRedisName(rf)
	namespace := rf.Metadata.Namespace
	propagation := metav1.DeletePropagationForeground
	if err := r.Client.AppsV1beta1().StatefulSets(namespace).Delete(name, &metav1.DeleteOptions{PropagationPolicy: &propagation}); err != nil {
		return err
	}
	logger := r.logger.WithField(logNameField, rf.Metadata.Name).WithField(logNamespaceField, rf.Metadata.Namespace)
	logger.Debug("Deleting Redis PodDisruptionBudget...")
	if err := r.deletePodDisruptionBudget(rf, redisName); err != nil {
		return err
	}
	logger.Debug("Redis PodDisruptionBudget deleted!")
	// TODO: Wait for statefulset to really delete
	return nil
}

// DeleteSentinelDeployment deletes a sentinel deployment
func (r *RedisFailoverKubeClient) DeleteSentinelDeployment(rf *RedisFailover) error {
	name := r.GetSentinelName(rf)
	namespace := rf.Metadata.Namespace
	propagation := metav1.DeletePropagationForeground
	if err := r.Client.AppsV1beta1().Deployments(namespace).Delete(name, &metav1.DeleteOptions{PropagationPolicy: &propagation}); err != nil {
		return err
	}
	logger := r.logger.WithField(logNameField, rf.Metadata.Name).WithField(logNamespaceField, rf.Metadata.Namespace)
	logger.Debug("Deleting Sentinel PodDisruptionBudget...")
	if err := r.deletePodDisruptionBudget(rf, sentinelName); err != nil {
		return err
	}
	logger.Debug("Sentinel PodDisruptionBudget deleted!")
	// TODO: Wait for deployment to really delete
	return nil
}

// DeleteSentinelService deletes a sentinel service
func (r *RedisFailoverKubeClient) DeleteSentinelService(rf *RedisFailover) error {
	name := r.GetSentinelName(rf)
	namespace := rf.Metadata.Namespace
	propagation := metav1.DeletePropagationForeground
	if err := r.Client.CoreV1().Services(namespace).Delete(name, &metav1.DeleteOptions{PropagationPolicy: &propagation}); err != nil {
		return err
	}
	// TODO: Wait for service to really delete
	return nil
}

// DeleteRedisService deletes redis service
func (r *RedisFailoverKubeClient) DeleteRedisService(rf *RedisFailover) error {
	name := r.GetRedisName(rf)
	namespace := rf.Metadata.Namespace
	propagation := metav1.DeletePropagationForeground
	if err := r.Client.CoreV1().Services(namespace).Delete(name, &metav1.DeleteOptions{PropagationPolicy: &propagation}); err != nil {
		return err
	}
	return nil
}

// deletePodDisruptionBudget deletes a PodDisruptionBudget for redis or sentinel
func (r *RedisFailoverKubeClient) deletePodDisruptionBudget(rf *RedisFailover, role string) error {
	name := generateName(role, rf.Metadata.Name)
	namespace := rf.Metadata.Namespace
	if _, err := r.Client.PolicyV1beta1().PodDisruptionBudgets(namespace).Get(name, metav1.GetOptions{}); err == nil {
		return r.Client.PolicyV1beta1().PodDisruptionBudgets(namespace).Delete(name, &metav1.DeleteOptions{})
	}
	return nil
}

func getSentinelResources(spec RedisFailoverSpec) v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: getRequests(spec.Sentinel.Resources),
		Limits:   getLimits(spec.Sentinel.Resources),
	}
}

func getRedisResources(spec RedisFailoverSpec) v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: getRequests(spec.Redis.Resources),
		Limits:   getLimits(spec.Redis.Resources),
	}
}

func getLimits(resources RedisFailoverResources) v1.ResourceList {
	return generateResourceList(resources.Limits.CPU, resources.Limits.Memory)
}

func getRequests(resources RedisFailoverResources) v1.ResourceList {
	return generateResourceList(resources.Requests.CPU, resources.Requests.Memory)
}

func generateResourceList(cpu string, memory string) v1.ResourceList {
	resources := v1.ResourceList{}
	if cpu != "" {
		resources[v1.ResourceCPU], _ = resource.ParseQuantity(cpu)
	}
	if memory != "" {
		resources[v1.ResourceMemory], _ = resource.ParseQuantity(memory)
	}
	return resources
}
