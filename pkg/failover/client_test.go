package failover_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	k8stesting "k8s.io/client-go/testing"

	"github.com/spotahome/redis-operator/mocks"
	"github.com/spotahome/redis-operator/pkg/clock"
	"github.com/spotahome/redis-operator/pkg/config"
	"github.com/spotahome/redis-operator/pkg/failover"
	"github.com/spotahome/redis-operator/pkg/log"
)

const (
	name      = "test"
	namespace = "testns"

	bootstrapName = "rfb-test"
	sentinelName  = "rfs-test"
	redisName     = "rfr-test"
)

type testResources struct {
	Input    failover.RedisFailoverResources
	Expected v1.ResourceRequirements
}

func generateResources(cpu string, memory string) []testResources {
	var result []testResources
	cpuQuantity, _ := resource.ParseQuantity(cpu)
	memoryQuantity, _ := resource.ParseQuantity(memory)
	// testCase -> request-cpu, request-memory, limit-cpu, limit-memory
	testCases := [][]bool{
		{false, false, false, false},
		{false, false, false, true},
		{false, false, true, false},
		{false, false, true, true},
		{false, true, false, false},
		{false, true, false, true},
		{false, true, true, false},
		{false, true, true, true},
		{true, false, false, false},
		{true, false, false, true},
		{true, false, true, false},
		{true, false, true, true},
		{true, true, false, false},
		{true, true, false, true},
		{true, true, true, false},
		{true, true, true, true},
	}
	for _, testCase := range testCases {
		testResource := testResources{}
		testResource.Expected.Requests = v1.ResourceList{}
		testResource.Expected.Limits = v1.ResourceList{}
		if testCase[0] {
			testResource.Input.Requests.CPU = cpu
			testResource.Expected.Requests[v1.ResourceCPU] = cpuQuantity
		}
		if testCase[1] {
			testResource.Input.Requests.Memory = memory
			testResource.Expected.Requests[v1.ResourceMemory] = memoryQuantity
		}
		if testCase[2] {
			testResource.Input.Limits.CPU = cpu
			testResource.Expected.Limits[v1.ResourceCPU] = cpuQuantity
		}
		if testCase[3] {
			testResource.Input.Limits.Memory = memory
			testResource.Expected.Limits[v1.ResourceMemory] = memoryQuantity
		}
		result = append(result, testResource)
	}
	return result
}

func TestGetBootstrapPodError(t *testing.T) {
	assert := assert.New(t)
	client := fake.NewSimpleClientset()

	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	_, err := r.GetBootstrapPod(redisFailover)
	assert.Error(err)
}

func TestGetBootstrapPod(t *testing.T) {
	assert := assert.New(t)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	existingPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bootstrapName,
			Namespace: namespace,
		},
	}

	client := fake.NewSimpleClientset(existingPod)
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	getPod, err := r.GetBootstrapPod(redisFailover)
	assert.NoError(err)
	assert.Equal(existingPod, getPod, "")
}

func TestGetSentinelServiceError(t *testing.T) {
	assert := assert.New(t)

	client := fake.NewSimpleClientset()

	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	_, err := r.GetSentinelService(redisFailover)
	assert.Error(err)
}

func TestGetSentinelService(t *testing.T) {
	assert := assert.New(t)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	existingService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sentinelName,
			Namespace: namespace,
		},
	}

	client := fake.NewSimpleClientset(existingService)
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	getSvc, err := r.GetSentinelService(redisFailover)
	assert.NoError(err)
	assert.Equal(existingService, getSvc, "")
}

func TestGetSentinelDeploymentError(t *testing.T) {
	assert := assert.New(t)

	client := fake.NewSimpleClientset()

	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	_, err := r.GetSentinelDeployment(redisFailover)
	assert.Error(err)
}

func TestGetSentinelDeployment(t *testing.T) {
	assert := assert.New(t)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	existingDeployment := &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sentinelName,
			Namespace: namespace,
		},
	}

	client := fake.NewSimpleClientset(existingDeployment)
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	getDepl, err := r.GetSentinelDeployment(redisFailover)
	assert.NoError(err)
	assert.Equal(existingDeployment, getDepl, "")
}

func TestGetRedisServiceError(t *testing.T) {
	assert := assert.New(t)

	client := fake.NewSimpleClientset()

	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	_, err := r.GetRedisService(redisFailover)
	assert.Error(err)
}

func TestGetRedisService(t *testing.T) {
	assert := assert.New(t)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	existingService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redisName,
			Namespace: namespace,
		},
	}

	client := fake.NewSimpleClientset(existingService)
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	getSvc, err := r.GetRedisService(redisFailover)
	assert.NoError(err)
	assert.Equal(existingService, getSvc, "")
}

func TestGetRedisStatefulsetError(t *testing.T) {
	assert := assert.New(t)

	client := fake.NewSimpleClientset()

	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	_, err := r.GetRedisStatefulset(redisFailover)
	assert.Error(err)
}

func TestGetRedisStatefulset(t *testing.T) {
	assert := assert.New(t)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	existingStatefulset := &v1beta1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redisName,
			Namespace: namespace,
		},
	}

	client := fake.NewSimpleClientset(existingStatefulset)
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	getSS, err := r.GetRedisStatefulset(redisFailover)
	assert.NoError(err)
	assert.Equal(existingStatefulset, getSS, "")
}

func TestGetSentinelPodsIPsEndpointsError(t *testing.T) {
	assert := assert.New(t)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	client := fake.NewSimpleClientset()
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	_, err := r.GetSentinelPodsIPs(redisFailover)
	assert.Error(err)
}

func TestGetSentinelPodsIPsEndpointsNumberError(t *testing.T) {
	assert := assert.New(t)
	existingEndpoints := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sentinelName,
			Namespace: namespace,
		},
	}

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	client := fake.NewSimpleClientset(existingEndpoints)
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	_, err := r.GetSentinelPodsIPs(redisFailover)
	assert.Error(err)
}

func TestGetSentinelPodsIPs(t *testing.T) {
	assert := assert.New(t)
	ip := "0.0.0.0"
	existingEndpoints := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sentinelName,
			Namespace: namespace,
		},
		Subsets: []v1.EndpointSubset{
			v1.EndpointSubset{
				Addresses: []v1.EndpointAddress{
					v1.EndpointAddress{
						IP: ip,
					},
				},
			},
		},
	}

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	client := fake.NewSimpleClientset(existingEndpoints)
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	ips, err := r.GetSentinelPodsIPs(redisFailover)
	assert.NoError(err)
	assert.Len(ips, 1, "")
	assert.Equal(ip, ips[0], "")
}

func TestGetRedisPodsIPsGetPodError(t *testing.T) {
	assert := assert.New(t)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	client := fake.NewSimpleClientset()
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	_, err := r.GetRedisPodsIPs(redisFailover)
	assert.Error(err)
}

func TestGetRedisPodsIPs(t *testing.T) {
	assert := assert.New(t)
	ip := "0.0.0.0"

	existingPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", redisName, 0),
			Namespace: namespace,
		},
		Status: v1.PodStatus{
			PodIP: ip,
		},
	}

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(1),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	client := fake.NewSimpleClientset(existingPod)
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	pods, err := r.GetRedisPodsIPs(redisFailover)
	assert.NoError(err)
	assert.Len(pods, 1, "")
	assert.Equal(ip, pods[0], "")
}

func TestCreateBootstrapPodError(t *testing.T) {
	assert := assert.New(t)
	existingPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bootstrapName,
			Namespace: namespace,
		},
	}
	redisFailover := failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}
	client := fake.NewSimpleClientset(existingPod)
	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	err := r.CreateBootstrapPod(&redisFailover)
	assert.Error(err)
}

func TestCreateBootstrapPod(t *testing.T) {
	assert := assert.New(t)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	// Create a faked K8S client
	client := &fake.Clientset{}
	// Add a reactor when calling pods
	client.Fake.AddReactor("get", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		// Create the pod to be returned with the status Ready = True
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bootstrapName,
				Namespace: namespace,
			},
			Status: v1.PodStatus{
				Conditions: []v1.PodCondition{
					v1.PodCondition{
						Type:   "Ready",
						Status: v1.ConditionTrue,
					},
				},
			},
		}

		// Return the pod as if we where the API responding to GET pods
		return true, pod, nil
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	err := r.CreateBootstrapPod(redisFailover)
	assert.NoError(err)
}

func TestCreateSentinelServiceError(t *testing.T) {
	assert := assert.New(t)
	existingService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sentinelName,
			Namespace: namespace,
		},
	}
	client := fake.NewSimpleClientset(existingService)
	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.CreateSentinelService(redisFailover)
	assert.Error(err)
}

func TestCreateSentinelService(t *testing.T) {
	assert := assert.New(t)

	// Create a faked K8S client
	client := &fake.Clientset{}
	// Add a reactor when calling pods
	client.Fake.AddReactor("get", "endpoints", func(action k8stesting.Action) (bool, runtime.Object, error) {
		// Create the endpoint to be returned with one ready address
		endpoint := &v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sentinelName,
				Namespace: namespace,
			},
			Subsets: []v1.EndpointSubset{
				v1.EndpointSubset{
					Addresses: []v1.EndpointAddress{
						v1.EndpointAddress{
							Hostname: "test",
							IP:       "1.1.1.1",
						},
					},
					NotReadyAddresses: []v1.EndpointAddress{},
					Ports:             []v1.EndpointPort{},
				},
			},
		}

		// Return the endpoint as if we where the API responding to GET endpoints
		return true, endpoint, nil
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.CreateSentinelService(redisFailover)
	assert.NoError(err)
}

func TestCreateSentinelDeploymentError(t *testing.T) {
	assert := assert.New(t)
	existingDeployment := &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sentinelName,
			Namespace: namespace,
		},
	}

	client := fake.NewSimpleClientset(existingDeployment)
	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.CreateSentinelDeployment(redisFailover)
	assert.Error(err)
}

func TestCreateSentinelDeploymentPDBError(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}

	deploymentSize := int32(3)

	client.Fake.AddReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		deployment := &v1beta1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sentinelName,
				Namespace: namespace,
			},
			Status: v1beta1.DeploymentStatus{
				ReadyReplicas: deploymentSize,
			},
		}
		return true, deployment, nil
	})

	client.Fake.AddReactor("get", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})
	client.Fake.AddReactor("create", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.CreateSentinelDeployment(redisFailover)
	assert.Error(err)
}

func TestCreateSentinelDeploymentReplicas(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}

	deploymentSize := int32(5) // Different from default

	replicasRequested := int32(0)

	client.Fake.AddReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		deployment := &v1beta1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sentinelName,
				Namespace: namespace,
			},
			Status: v1beta1.DeploymentStatus{
				ReadyReplicas: deploymentSize,
			},
		}
		return true, deployment, nil
	})

	client.Fake.AddReactor("get", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})
	client.Fake.AddReactor("create", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})

	client.Fake.AddReactor("create", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		createAction := action.(k8stesting.CreateAction)
		deployment := createAction.GetObject().(*v1beta1.Deployment)
		replicasRequested = *deployment.Spec.Replicas
		return true, nil, nil
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: deploymentSize,
			},
		},
	}

	err := r.CreateSentinelDeployment(redisFailover)
	assert.Error(err)
	assert.Equal(deploymentSize, replicasRequested, "Replicas set on deployment spec differs from object created")
}

func TestCreateSentinelDeploymentRequests(t *testing.T) {
	assert := assert.New(t)

	cpu := "100m"
	memory := "100Mi"

	tests := generateResources(cpu, memory)

	for _, test := range tests {
		client := &fake.Clientset{}

		var createdRequests v1.ResourceRequirements

		client.Fake.AddReactor("create", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
			createAction := action.(k8stesting.CreateAction)
			deployment := createAction.GetObject().(*v1beta1.Deployment)
			createdRequests = deployment.Spec.Template.Spec.Containers[0].Resources
			return true, nil, errors.New("")
		})

		mc := &mocks.Clock{}
		mc.On("NewTicker", mock.Anything).
			Once().Return(time.NewTicker(1))
		r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

		redisFailover := &failover.RedisFailover{
			Metadata: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: failover.RedisFailoverSpec{
				Sentinel: failover.SentinelSettings{
					Resources: failover.RedisFailoverResources{
						Requests: failover.CPUAndMem{
							CPU:    test.Input.Requests.CPU,
							Memory: test.Input.Requests.Memory,
						},
						Limits: failover.CPUAndMem{
							CPU:    test.Input.Limits.CPU,
							Memory: test.Input.Limits.Memory,
						},
					},
				},
			},
		}

		r.CreateSentinelDeployment(redisFailover)
		assert.Equal(test.Expected, createdRequests, "Requests are not equal as required")
	}
}

func TestCreateSentinelDeployment(t *testing.T) {
	assert := assert.New(t)

	// Create a faked K8S client
	client := &fake.Clientset{}

	deploymentSize := int32(3)
	// Add a reactor when calling pods
	client.Fake.AddReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		// Create the deployment to be returned with ReadyReplicas = 3
		deployment := &v1beta1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sentinelName,
				Namespace: namespace,
			},
			Status: v1beta1.DeploymentStatus{
				ReadyReplicas: deploymentSize,
			},
		}

		// Return the deployment as if we where the API responding to GET deployments
		return true, deployment, nil
	})
	client.Fake.AddReactor("get", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})
	called := false
	client.Fake.AddReactor("create", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		called = true
		return true, nil, nil
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: deploymentSize,
			},
		},
	}

	err := r.CreateSentinelDeployment(redisFailover)
	assert.NoError(err)
	assert.True(called, "PDB creation is not called")
}

func TestCreateRedisServiceError(t *testing.T) {
	assert := assert.New(t)
	existingService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redisName,
			Namespace: namespace,
		},
	}
	client := fake.NewSimpleClientset(existingService)
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.CreateRedisService(redisFailover)
	assert.Error(err)
}

func TestCreateRedisService(t *testing.T) {
	assert := assert.New(t)

	// Create a faked K8S client
	client := &fake.Clientset{}

	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.CreateRedisService(redisFailover)
	assert.NoError(err)
}

func TestCreateRedisStatefulsetError(t *testing.T) {
	assert := assert.New(t)
	existingStatefulset := &v1beta1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redisName,
			Namespace: namespace,
		},
	}
	client := fake.NewSimpleClientset(existingStatefulset)
	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.CreateRedisStatefulset(redisFailover)
	assert.Error(err)
}

func TestCreateRedisStatefulsetWithExporter(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
				Exporter: true,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	var nContainers int

	client.Fake.AddReactor("create", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		createAction := action.(k8stesting.CreateAction)
		statefulset := createAction.GetObject().(*v1beta1.StatefulSet)
		nContainers = len(statefulset.Spec.Template.Spec.Containers)
		return true, nil, errors.New("")
	})

	err := r.CreateRedisStatefulset(redisFailover)
	assert.Error(err)
	assert.Equal(nContainers, 2, "")
}

func TestCreateRedisStatefulsetWithoutExporter(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	var nContainers int

	client.Fake.AddReactor("create", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		createAction := action.(k8stesting.CreateAction)
		statefulset := createAction.GetObject().(*v1beta1.StatefulSet)
		nContainers = len(statefulset.Spec.Template.Spec.Containers)
		return true, nil, errors.New("")
	})

	err := r.CreateRedisStatefulset(redisFailover)
	assert.Error(err)
	assert.Equal(nContainers, 1, "")
}

func TestCreateRedisStatefulsetPDBError(t *testing.T) {
	assert := assert.New(t)

	// Create a faked K8S client
	client := &fake.Clientset{}

	statefulsetSize := int32(3)
	// Add a reactor when calling pods
	client.Fake.AddReactor("get", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		// Create the statefulset to be returned with Replicas = 3
		statefulset := &v1beta1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      redisName,
				Namespace: namespace,
			},
			Status: v1beta1.StatefulSetStatus{
				ReadyReplicas: statefulsetSize,
			},
		}

		// Return the statefulset as if we where the API responding to GET statefulsets
		return true, statefulset, nil
	})

	client.Fake.AddReactor("get", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})
	client.Fake.AddReactor("create", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: statefulsetSize,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.CreateRedisStatefulset(redisFailover)
	assert.Error(err)
}

func TestCreateRedisStatefulsetRequests(t *testing.T) {
	assert := assert.New(t)

	cpu := "100m"
	memory := "100Mi"

	tests := generateResources(cpu, memory)

	for _, test := range tests {
		client := &fake.Clientset{}

		var createdRequests v1.ResourceRequirements

		client.Fake.AddReactor("create", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
			createAction := action.(k8stesting.CreateAction)
			statefulset := createAction.GetObject().(*v1beta1.StatefulSet)
			createdRequests = statefulset.Spec.Template.Spec.Containers[0].Resources
			return true, nil, errors.New("")
		})

		mc := &mocks.Clock{}
		mc.On("NewTicker", mock.Anything).
			Once().Return(time.NewTicker(1))
		r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

		redisFailover := &failover.RedisFailover{
			Metadata: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: failover.RedisFailoverSpec{
				Redis: failover.RedisSettings{
					Resources: failover.RedisFailoverResources{
						Requests: failover.CPUAndMem{
							CPU:    test.Input.Requests.CPU,
							Memory: test.Input.Requests.Memory,
						},
						Limits: failover.CPUAndMem{
							CPU:    test.Input.Limits.CPU,
							Memory: test.Input.Limits.Memory,
						},
					},
				},
			},
		}

		r.CreateRedisStatefulset(redisFailover)
		assert.Equal(test.Expected, createdRequests, "Requests are not equal as required")
	}
}

func TestCreateRedisStatefulsetReplicas(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}

	statefulsetSize := int32(4) // Different from default

	replicasRequested := int32(0)

	client.Fake.AddReactor("get", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		statefulset := &v1beta1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      redisName,
				Namespace: namespace,
			},
			Status: v1beta1.StatefulSetStatus{
				ReadyReplicas: statefulsetSize,
			},
		}
		return true, statefulset, nil
	})

	client.Fake.AddReactor("get", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})
	client.Fake.AddReactor("create", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})

	client.Fake.AddReactor("create", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		createAction := action.(k8stesting.CreateAction)
		statefulset := createAction.GetObject().(*v1beta1.StatefulSet)
		replicasRequested = *statefulset.Spec.Replicas
		return true, nil, nil
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: statefulsetSize,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.CreateRedisStatefulset(redisFailover)
	assert.Error(err)
	assert.Equal(statefulsetSize, replicasRequested, "Replicas set on deployment spec differs from object created")
}

func TestCreateRedisStatefulsetImageVersion(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}

	imageVersion := "testing"
	var imageRequested string

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
				Version:  imageVersion,
				Exporter: false,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	client.Fake.AddReactor("create", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		createAction := action.(k8stesting.CreateAction)
		statefulset := createAction.GetObject().(*v1beta1.StatefulSet)
		imageRequested = statefulset.Spec.Template.Spec.Containers[0].Image
		return true, nil, errors.New("")
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	err := r.CreateRedisStatefulset(redisFailover)
	assert.Error(err)
	assert.Equal(fmt.Sprintf("%s:%s", config.RedisImage, imageVersion), imageRequested, "Redis image is not well formed. It's different from the asked one.")
}

func TestCreateRedisStatefulset(t *testing.T) {
	assert := assert.New(t)

	// Create a faked K8S client
	client := &fake.Clientset{}

	statefulsetSize := int32(3)
	// Add a reactor when calling pods
	client.Fake.AddReactor("get", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		// Create the statefulset to be returned with Replicas = 3
		statefulset := &v1beta1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      redisName,
				Namespace: namespace,
			},
			Status: v1beta1.StatefulSetStatus{
				ReadyReplicas: statefulsetSize,
			},
		}

		// Return the statefulset as if we where the API responding to GET statefulsets
		return true, statefulset, nil
	})

	client.Fake.AddReactor("get", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})
	pdbCalled := false
	client.Fake.AddReactor("create", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		pdbCalled = true
		return true, nil, nil
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.CreateRedisStatefulset(redisFailover)
	assert.NoError(err)
	assert.True(pdbCalled, "Create PodDisruptionBudget should have been called")
}

func TestUpdateSentinelDeploymentGetError(t *testing.T) {
	assert := assert.New(t)

	// Create a faked K8S client
	client := &fake.Clientset{}
	client.Fake.AddReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.UpdateSentinelDeployment(redisFailover)
	assert.Error(err)
}

func TestUpdateSentinelDeploymentError(t *testing.T) {
	assert := assert.New(t)

	replicas := int32(3) // Different from default

	// Create a faked K8S client
	client := &fake.Clientset{}
	client.Fake.AddReactor("update", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})
	client.Fake.AddReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		deployment := &v1beta1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sentinelName,
				Namespace: namespace,
			},
			Status: v1beta1.DeploymentStatus{
				ReadyReplicas:   replicas,
				UpdatedReplicas: replicas,
			},
			Spec: v1beta1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						InitContainers: []v1.Container{
							v1.Container{},
						},
						Containers: []v1.Container{
							v1.Container{},
						},
					},
				},
			},
		}
		return true, deployment, nil
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.UpdateSentinelDeployment(redisFailover)
	assert.Error(err)
}

func TestUpdateSentinelDeployment(t *testing.T) {
	assert := assert.New(t)

	replicas := int32(4) // Different from default
	cpu := "200m"
	memory := "200Mi"
	cpuQuantity, _ := resource.ParseQuantity(cpu)
	memoryQuantity, _ := resource.ParseQuantity(memory)
	var updatedRequests v1.ResourceRequirements

	requiredRequests := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    cpuQuantity,
			v1.ResourceMemory: memoryQuantity,
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    cpuQuantity,
			v1.ResourceMemory: memoryQuantity,
		},
	}

	// Create a faked K8S client
	client := &fake.Clientset{}
	client.Fake.AddReactor("update", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		updateAction := action.(k8stesting.UpdateAction)
		deployment := updateAction.GetObject().(*v1beta1.Deployment)
		updatedRequests = deployment.Spec.Template.Spec.Containers[0].Resources
		return true, nil, nil
	})

	client.Fake.AddReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		deployment := &v1beta1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sentinelName,
				Namespace: namespace,
			},
			Status: v1beta1.DeploymentStatus{
				ReadyReplicas:   replicas,
				UpdatedReplicas: replicas,
			},
			Spec: v1beta1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						InitContainers: []v1.Container{
							v1.Container{},
						},
						Containers: []v1.Container{
							v1.Container{},
						},
					},
				},
			},
		}
		return true, deployment, nil
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: replicas,
				Resources: failover.RedisFailoverResources{
					Limits: failover.CPUAndMem{
						CPU:    cpu,
						Memory: memory,
					},
					Requests: failover.CPUAndMem{
						CPU:    cpu,
						Memory: memory,
					},
				},
			},
		},
	}

	err := r.UpdateSentinelDeployment(redisFailover)
	assert.NoError(err)
	assert.Equal(requiredRequests, updatedRequests, "Requests are not equal as updated")
}

func TestUpdateRedisStatefulsetGetError(t *testing.T) {
	assert := assert.New(t)

	// Create a faked K8S client
	client := &fake.Clientset{}
	client.Fake.AddReactor("get", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.UpdateRedisStatefulset(redisFailover)
	assert.Error(err)
}

func TestUpdateRedisStatefulsetError(t *testing.T) {
	assert := assert.New(t)

	// Create a faked K8S client
	client := &fake.Clientset{}
	client.Fake.AddReactor("get", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		statefulset := &v1beta1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      redisName,
				Namespace: namespace,
			},
			Status: v1beta1.StatefulSetStatus{
				ReadyReplicas: int32(3),
			},
			Spec: v1beta1.StatefulSetSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							v1.Container{},
						},
					},
				},
			},
		}
		return true, statefulset, nil
	})
	client.Fake.AddReactor("update", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.UpdateRedisStatefulset(redisFailover)
	assert.Error(err)
}

func TestUpdateRedisStatefulsetWithUpdate(t *testing.T) {
	assert := assert.New(t)

	replicas := int32(3)
	replicasUpdated := int32(4)
	called := false
	cpu := "200m"
	memory := "200Mi"
	cpuQuantityOriginal, _ := resource.ParseQuantity("100m")
	memoryQuantityOriginal, _ := resource.ParseQuantity("100Mi")
	cpuQuantityRequired, _ := resource.ParseQuantity(cpu)
	memoryQuantityRequired, _ := resource.ParseQuantity(memory)
	var updatedRequests v1.ResourceRequirements

	requiredRequests := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    cpuQuantityRequired,
			v1.ResourceMemory: memoryQuantityRequired,
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    cpuQuantityRequired,
			v1.ResourceMemory: memoryQuantityRequired,
		},
	}

	exporterExists := false

	// Create a faked K8S client
	client := &fake.Clientset{}
	client.Fake.AddReactor("get", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		r := replicas
		if called {
			r = replicasUpdated
		}
		statefulset := &v1beta1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      redisName,
				Namespace: namespace,
			},
			Status: v1beta1.StatefulSetStatus{
				ReadyReplicas:   r,
				UpdatedReplicas: r,
			},
			Spec: v1beta1.StatefulSetSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							v1.Container{
								Name: redisName,
								Resources: v1.ResourceRequirements{
									Limits: v1.ResourceList{
										v1.ResourceCPU:    cpuQuantityOriginal,
										v1.ResourceMemory: memoryQuantityOriginal,
									},
									Requests: v1.ResourceList{
										v1.ResourceCPU:    cpuQuantityOriginal,
										v1.ResourceMemory: memoryQuantityOriginal,
									},
								},
							},
						},
					},
				},
			},
		}
		called = true
		return true, statefulset, nil
	})
	client.Fake.AddReactor("update", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		updateAction := action.(k8stesting.UpdateAction)
		statefulset := updateAction.GetObject().(*v1beta1.StatefulSet)
		for _, container := range statefulset.Spec.Template.Spec.Containers {
			if container.Name == redisName {
				updatedRequests = container.Resources
			}
			if container.Name == "redis-exporter" {
				exporterExists = true
			}
		}
		return true, nil, nil
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: replicasUpdated,
				Resources: failover.RedisFailoverResources{
					Limits: failover.CPUAndMem{
						CPU:    cpu,
						Memory: memory,
					},
					Requests: failover.CPUAndMem{
						CPU:    cpu,
						Memory: memory,
					},
				},
				Exporter: true,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.UpdateRedisStatefulset(redisFailover)
	assert.NoError(err)
	assert.Equal(requiredRequests, updatedRequests, "Requests are not equal as updated")
	assert.True(exporterExists, "Redis-exporter should exist")
}

func TestUpdateRedisStatefulsetWithoutUpdate(t *testing.T) {
	assert := assert.New(t)

	replicas := int32(3)
	replicasUpdated := int32(4)
	called := false
	cpu := "200m"
	memory := "200Mi"
	cpuQuantityOriginal, _ := resource.ParseQuantity("100m")
	memoryQuantityOriginal, _ := resource.ParseQuantity("100Mi")
	cpuQuantityRequired, _ := resource.ParseQuantity(cpu)
	memoryQuantityRequired, _ := resource.ParseQuantity(memory)
	var updatedRequests v1.ResourceRequirements

	requiredRequests := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    cpuQuantityRequired,
			v1.ResourceMemory: memoryQuantityRequired,
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    cpuQuantityRequired,
			v1.ResourceMemory: memoryQuantityRequired,
		},
	}

	exporterExists := false

	// Create a faked K8S client
	client := &fake.Clientset{}
	client.Fake.AddReactor("get", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		r := replicas
		if called {
			r = replicasUpdated
		}
		statefulset := &v1beta1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      redisName,
				Namespace: namespace,
			},
			Status: v1beta1.StatefulSetStatus{
				ReadyReplicas:   r,
				UpdatedReplicas: r,
			},
			Spec: v1beta1.StatefulSetSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							v1.Container{
								Name: redisName,
								Resources: v1.ResourceRequirements{
									Limits: v1.ResourceList{
										v1.ResourceCPU:    cpuQuantityOriginal,
										v1.ResourceMemory: memoryQuantityOriginal,
									},
									Requests: v1.ResourceList{
										v1.ResourceCPU:    cpuQuantityOriginal,
										v1.ResourceMemory: memoryQuantityOriginal,
									},
								},
							},
							v1.Container{
								Name: "redis-exporter",
							},
						},
					},
				},
			},
		}
		called = true
		return true, statefulset, nil
	})
	client.Fake.AddReactor("update", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		updateAction := action.(k8stesting.UpdateAction)
		statefulset := updateAction.GetObject().(*v1beta1.StatefulSet)
		for _, container := range statefulset.Spec.Template.Spec.Containers {
			if container.Name == redisName {
				updatedRequests = container.Resources
			}
			if container.Name == "redis-exporter" {
				exporterExists = true
			}
		}
		return true, nil, nil
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: replicasUpdated,
				Resources: failover.RedisFailoverResources{
					Limits: failover.CPUAndMem{
						CPU:    cpu,
						Memory: memory,
					},
					Requests: failover.CPUAndMem{
						CPU:    cpu,
						Memory: memory,
					},
				},
				Exporter: false,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.UpdateRedisStatefulset(redisFailover)
	assert.NoError(err)
	assert.Equal(requiredRequests, updatedRequests, "Requests are not equal as updated")
	assert.False(exporterExists, "Redis-exporter should not exist")
}

func TestUpdateRedisStatefulset(t *testing.T) {
	assert := assert.New(t)

	replicas := int32(3)
	replicasUpdated := int32(4)
	called := false
	cpu := "200m"
	memory := "200Mi"
	cpuQuantityOriginal, _ := resource.ParseQuantity("100m")
	memoryQuantityOriginal, _ := resource.ParseQuantity("100Mi")
	cpuQuantityRequired, _ := resource.ParseQuantity(cpu)
	memoryQuantityRequired, _ := resource.ParseQuantity(memory)
	var updatedRequests v1.ResourceRequirements

	requiredRequests := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    cpuQuantityRequired,
			v1.ResourceMemory: memoryQuantityRequired,
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    cpuQuantityRequired,
			v1.ResourceMemory: memoryQuantityRequired,
		},
	}

	// Create a faked K8S client
	client := &fake.Clientset{}
	client.Fake.AddReactor("get", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		r := replicas
		if called {
			r = replicasUpdated
		}
		statefulset := &v1beta1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      redisName,
				Namespace: namespace,
			},
			Status: v1beta1.StatefulSetStatus{
				ReadyReplicas:   r,
				UpdatedReplicas: r,
			},
			Spec: v1beta1.StatefulSetSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							v1.Container{
								Resources: v1.ResourceRequirements{
									Limits: v1.ResourceList{
										v1.ResourceCPU:    cpuQuantityOriginal,
										v1.ResourceMemory: memoryQuantityOriginal,
									},
									Requests: v1.ResourceList{
										v1.ResourceCPU:    cpuQuantityOriginal,
										v1.ResourceMemory: memoryQuantityOriginal,
									},
								},
							},
						},
					},
				},
			},
		}
		called = true
		return true, statefulset, nil
	})
	client.Fake.AddReactor("update", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		updateAction := action.(k8stesting.UpdateAction)
		statefulset := updateAction.GetObject().(*v1beta1.StatefulSet)
		updatedRequests = statefulset.Spec.Template.Spec.Containers[0].Resources
		return true, nil, nil
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: replicasUpdated,
				Resources: failover.RedisFailoverResources{
					Limits: failover.CPUAndMem{
						CPU:    cpu,
						Memory: memory,
					},
					Requests: failover.CPUAndMem{
						CPU:    cpu,
						Memory: memory,
					},
				},
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.UpdateRedisStatefulset(redisFailover)
	assert.NoError(err)
	assert.Equal(requiredRequests, updatedRequests, "Requests are not equal as updated")
}

func TestDeleteBootstrapPodError(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}
	client.Fake.AddReactor("delete", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("Pod does not exist")
	})
	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.DeleteBootstrapPod(redisFailover)
	assert.Error(err)
}

func TestDeleteBootstrapPod(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}
	client.Fake.AddReactor("delete", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {

		podEmpty := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "",
				Namespace: namespace,
			},
		}

		// Return a pod with no name
		return true, podEmpty, nil
	})
	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.DeleteBootstrapPod(redisFailover)
	assert.NoError(err)
}

func TestDeleteRedisStatefulsetError(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}
	client.Fake.AddReactor("delete", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("Statefulset does not exist")
	})
	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.DeleteRedisStatefulset(redisFailover)
	assert.Error(err)
}

func TestDeleteRedisStatefulsetDeletePDBError(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}
	client.Fake.AddReactor("delete", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})
	client.Fake.AddReactor("delete", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})
	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.DeleteRedisStatefulset(redisFailover)
	assert.Error(err)
}

func TestDeleteRedisStatefulset(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}
	client.Fake.AddReactor("delete", "statefulsets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})
	pdbCalled := false
	client.Fake.AddReactor("delete", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		pdbCalled = true
		return true, nil, nil
	})
	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.DeleteRedisStatefulset(redisFailover)
	assert.NoError(err)
	assert.True(pdbCalled, "PDB Delete was not called")
}

func TestDeleteSentinelDeploymentError(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}
	client.Fake.AddReactor("delete", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("Deployment does not exist")
	})
	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.DeleteSentinelDeployment(redisFailover)
	assert.Error(err)
}

func TestDeleteSentinelDeploymentDeletePDBError(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}
	client.Fake.AddReactor("delete", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})
	client.Fake.AddReactor("delete", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})
	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.DeleteSentinelDeployment(redisFailover)
	assert.Error(err)
}

func TestDeleteSentinelDeployment(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}
	client.Fake.AddReactor("delete", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})
	pdbCalled := false
	client.Fake.AddReactor("delete", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		pdbCalled = true
		return true, nil, nil
	})
	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.DeleteSentinelDeployment(redisFailover)
	assert.NoError(err)
	assert.True(pdbCalled, "PDB Delete was not called")
}

func TestDeleteSentinelServiceError(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}
	client.Fake.AddReactor("delete", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})
	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.DeleteSentinelService(redisFailover)
	assert.Error(err)
}

func TestDeleteSentinelService(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}
	var deletedName string
	client.Fake.AddReactor("delete", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		deleteAction := action.(k8stesting.DeleteAction)
		deletedName = deleteAction.GetName()
		return true, nil, nil
	})
	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.DeleteSentinelService(redisFailover)
	assert.NoError(err)
	assert.Equal(r.GetSentinelName(redisFailover), deletedName, "")
}

func TestDeleteRedisServiceError(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}
	client.Fake.AddReactor("delete", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("")
	})
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.DeleteRedisService(redisFailover)
	assert.Error(err)
}

func TestDeleteRedisService(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}
	var deletedName string
	client.Fake.AddReactor("delete", "services", func(action k8stesting.Action) (bool, runtime.Object, error) {
		deleteAction := action.(k8stesting.DeleteAction)
		deletedName = deleteAction.GetName()
		return true, nil, nil
	})
	r := failover.NewRedisFailoverKubeClient(client, clock.Base(), log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.DeleteRedisService(redisFailover)
	assert.NoError(err)
	assert.Equal(r.GetRedisName(redisFailover), deletedName, "")
}

func TestCreatePDBAlreadyExists(t *testing.T) {
	assert := assert.New(t)
	client := &fake.Clientset{}

	deploymentSize := int32(3)

	client.Fake.AddReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		deployment := &v1beta1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sentinelName,
				Namespace: namespace,
			},
			Status: v1beta1.DeploymentStatus{
				ReadyReplicas: deploymentSize,
			},
		}
		return true, deployment, nil
	})

	client.Fake.AddReactor("get", "poddisruptionbudgets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})

	mc := &mocks.Clock{}
	mc.On("NewTicker", mock.Anything).
		Once().Return(time.NewTicker(1))
	r := failover.NewRedisFailoverKubeClient(client, mc, log.Nil)

	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	err := r.CreateSentinelDeployment(redisFailover)
	assert.NoError(err)
}
