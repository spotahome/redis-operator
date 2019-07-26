package service

import (
	"fmt"

	redisfailoverv1alpha2 "github.com/spotahome/redis-operator/api/redisfailover/v1alpha2"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/operator/redisfailover/util"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// HaproxyBackends HaproxyBackends
type HaproxyBackends interface {
	GetHAProxyBackendsList(name string, list corev1.EndpointsList) ([]string, error)
}

// EnpointsService is the service account service implementation using API calls to kubernetes.
type EnpointsService struct {
	logger log.Logger
}

// NewEndpointsService NewEndpointsService
func NewEndpointsService(logger log.Logger) *EnpointsService {
	logger = logger.With("service", "k8s.endpoints")
	return &EnpointsService{
		logger: logger,
	}
}

func (s *EnpointsService) getObjInfo(obj runtime.Object) (string, error) {
	objMeta, ok := obj.(metav1.Object)
	if !ok {
		return "", fmt.Errorf("could not print object information")
	}
	return fmt.Sprintf("%s", objMeta.GetName()), nil
}

// GetHAProxyBackendsList GetHAProxyBackendsList
func (s *EnpointsService) GetHAProxyBackendsList(name string, list corev1.EndpointsList) ([]string, error) {
	// search current instance EP and return array of IPs
	return []string{}, nil
}

func getHaproxyImage(rf *redisfailoverv1alpha2.RedisFailover) string {
	return fmt.Sprintf("%s:%s", rf.Spec.HAProxy.Image, rf.Spec.HAProxy.Version)
}

func generateHAProxyDeployment(rf *redisfailoverv1alpha2.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) *appsv1beta2.Deployment {
	name := GetSentinelName(rf) + "-haproxy"
	// make possible to set CM by hands in CRD
	configMapName := GetSentinelName(rf) + "-haproxy"
	namespace := rf.Namespace

	spec := rf.Spec
	haproxyImage := getHaproxyImage(rf)
	resources := getSentinelResources(spec)
	labels = util.MergeLabels(labels, generateLabels("haproxy", rf.Name))

	return &appsv1beta2.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Spec: appsv1beta2.DeploymentSpec{
			Replicas: &rf.Spec.HAProxy.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity:    rf.Spec.NodeAffinity,
						PodAntiAffinity: createPodAntiAffinity(rf.Spec.HardAntiAffinity, labels),
					},
					Tolerations:     rf.Spec.Tolerations,
					SecurityContext: rf.Spec.SecurityContext,
					Containers: []corev1.Container{
						{
							Name:            "haproxy",
							Image:           haproxyImage,
							ImagePullPolicy: "Always",
							Ports: []corev1.ContainerPort{
								{
									Name:          "haproxy",
									ContainerPort: 26379,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "haproxy-config",
									MountPath: "/etc/haproxy/haproxy.cfg",
								},
							},
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: graceTime,
								TimeoutSeconds:      5,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											"redis-cli -h $(hostname) -p 26379 ping",
										},
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								InitialDelaySeconds: graceTime,
								TimeoutSeconds:      5,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
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
					Volumes: []corev1.Volume{
						{
							Name: "haproxy-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMapName,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func generateHAProxyService(rf *redisfailoverv1alpha2.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) *corev1.Service {
	name := GetRedisName(rf)
	namespace := rf.Namespace

	labels = util.MergeLabels(labels, generateLabels(redisRoleName, rf.Name))

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/port":   "http",
				"prometheus.io/path":   "/metrics",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
			Ports: []corev1.ServicePort{
				{
					Port:     exporterPort,
					Protocol: corev1.ProtocolTCP,
					Name:     exporterPortName,
				},
			},
			Selector: labels,
		},
	}
}

func generateHAProxyConfigMap(rf *redisfailoverv1alpha2.RedisFailover, labels map[string]string, ownerRefs []metav1.OwnerReference) *corev1.ConfigMap {
	name := GetSentinelName(rf) + "-haproxy"
	namespace := rf.Namespace

	labels = util.MergeLabels(labels, generateLabels(sentinelRoleName, rf.Name))

	sentinelConfigFileContent := `
defaults
  mode tcp
  timeout connect 3s
  timeout server 6s
  timeout client 6s
listen stats
  mode http
  bind :9000
  stats enable
  stats hide-version
  stats realm Haproxy\ Statistics
  stats uri /haproxy_stats
frontend ft_redis
  mode tcp
  bind *:80
  default_backend bk_redis
backend bk_redis
  mode tcp
  option tcp-check
  tcp-check send PING\r\n
  tcp-check expect string +PONG
  tcp-check send info\ replication\r\n
  tcp-check expect string role:master
  tcp-check send QUIT\r\n
  tcp-check expect string +OK
# autogenerate at enpoints watch
#  server redis_backend_01 redis01:6379 maxconn 1024 check inter 1s
#  server redis_backend_02 redis02:6379 maxconn 1024 check inter 1s
#  server redis_backend_03 redis03:6379 maxconn 1024 check inter 1s
`

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Data: map[string]string{
			sentinelConfigFileName: sentinelConfigFileContent,
		},
	}
}
