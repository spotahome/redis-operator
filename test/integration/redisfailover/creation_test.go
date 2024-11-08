//go:build integration
// +build integration

package redisfailover_test

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	rediscli "github.com/go-redis/redis/v8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/util/homedir"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	redisfailoverclientset "github.com/spotahome/redis-operator/client/k8s/clientset/versioned"
	"github.com/spotahome/redis-operator/cmd/utils"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
	"github.com/spotahome/redis-operator/operator/redisfailover"
	"github.com/spotahome/redis-operator/service/k8s"
	"github.com/spotahome/redis-operator/service/redis"
)

const (
	name           = "testing"
	namespace      = "testns"
	redisSize      = int32(3)
	sentinelSize   = int32(3)
	authSecretPath = "redis-auth"
	testPass       = "test-pass"
	redisAddr      = "redis://127.0.0.1:6379"
)

type clients struct {
	k8sClient   kubernetes.Interface
	rfClient    redisfailoverclientset.Interface
	aeClient    apiextensionsclientset.Interface
	redisClient redis.Client
}

func (c *clients) prepareNS(currentNamespace string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: currentNamespace,
		},
	}
	_, err := c.k8sClient.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	return err
}

func (c *clients) cleanup(stopC chan struct{}, currentNamespace string) {
	c.k8sClient.CoreV1().Namespaces().Delete(context.Background(), currentNamespace, metav1.DeleteOptions{})
	close(stopC)
}

func TestRedisFailover(t *testing.T) {
	require := require.New(t)
	disableMyMaster := true
	currentNamespace := namespace

	// Create signal channels.
	stopC := make(chan struct{})
	errC := make(chan error)

	flags := &utils.CMDFlags{
		KubeConfig:  filepath.Join(homedir.HomeDir(), ".kube", "config"),
		Development: true,
	}

	// Kubernetes clients.
	k8sClient, customClient, aeClientset, err := utils.CreateKubernetesClients(flags)
	require.NoError(err)

	// Create the redis clients
	redisClient := redis.New(metrics.Dummy)

	clients := clients{
		k8sClient:   k8sClient,
		rfClient:    customClient,
		aeClient:    aeClientset,
		redisClient: redisClient,
	}

	// Create kubernetes service.
	k8sservice := k8s.New(k8sClient, customClient, aeClientset, log.Dummy, metrics.Dummy)

	// Prepare namespace
	prepErr := clients.prepareNS(currentNamespace)
	require.NoError(prepErr)

	// Give time to the namespace to be ready
	time.Sleep(15 * time.Second)

	// Create operator and run.
	redisfailoverOperator, err := redisfailover.New(redisfailover.Config{}, k8sservice, k8sClient, namespace, redisClient, metrics.Dummy, log.Dummy)
	require.NoError(err)

	go func() {
		errC <- redisfailoverOperator.Run(context.Background())
	}()

	// Prepare cleanup for when the test ends
	defer clients.cleanup(stopC, currentNamespace)

	// Give time to the operator to start
	time.Sleep(15 * time.Second)

	// Create secret
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      authSecretPath,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"password": []byte(testPass),
		},
	}
	_, err = k8sClient.CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	require.NoError(err)

	// Check that if we create a RedisFailover, it is certainly created and we can get it
	ok := t.Run("Check Custom Resource Creation", func(t *testing.T) {
		clients.testCRCreation(t, currentNamespace, disableMyMaster)
	})
	require.True(ok, "the custom resource has to be created to continue")

	// Giving time to the operator to create the resources
	time.Sleep(3 * time.Minute)

	// Verify that auth is set and actually working
	t.Run("Check that auth is set in sentinel and redis configs", func(t *testing.T) {
		clients.testAuth(t, currentNamespace)
	})

	// Check custom config is set
	t.Run("Check that custom config is behave expected", func(t *testing.T) {
		clients.testCustomConfig(t, currentNamespace)
	})

	// Check that a Redis Statefulset is created and the size of it is the one defined by the
	// Redis Failover definition created before.
	t.Run("Check Redis Statefulset existing and size", func(t *testing.T) {
		clients.testRedisStatefulSet(t, currentNamespace)
	})

	// Check that a Sentinel Deployment is created and the size of it is the one defined by the
	// Redis Failover definition created before.
	t.Run("Check Sentinel Deployment existing and size", func(t *testing.T) {
		clients.testSentinelDeployment(t, currentNamespace)
	})

	// Connect to all the Redis pods and, asking to the Redis running inside them, check
	// that only one of them is the master of the failover.
	t.Run("Check Only One Redis Master", func(t *testing.T) {
		clients.testRedisMaster(t, currentNamespace)
	})

	// Connect to all the Sentinel pods and, asking to the Sentinel running inside them,
	// check that all of them are connected to the same Redis node, and also that that node
	// is the master.
	t.Run("Check Sentinels Checking the Redis Master", func(t *testing.T) {
		clients.testSentinelMonitoring(t, currentNamespace, disableMyMaster)
	})
}

func TestRedisFailoverMyMaster(t *testing.T) {
	require := require.New(t)
	disableMyMaster := false
	currentNamespace := "mymaster-" + namespace

	// Create signal channels.
	stopC := make(chan struct{})
	errC := make(chan error)

	flags := &utils.CMDFlags{
		KubeConfig:  filepath.Join(homedir.HomeDir(), ".kube", "config"),
		Development: true,
	}

	// Kubernetes clients.
	k8sClient, customClient, aeClientset, err := utils.CreateKubernetesClients(flags)
	require.NoError(err)

	// Create the redis clients
	redisClient := redis.New(metrics.Dummy)

	clients := clients{
		k8sClient:   k8sClient,
		rfClient:    customClient,
		aeClient:    aeClientset,
		redisClient: redisClient,
	}

	// Create kubernetes service.
	k8sservice := k8s.New(k8sClient, customClient, aeClientset, log.Dummy, metrics.Dummy)

	// Prepare namespace
	prepErr := clients.prepareNS(currentNamespace)
	require.NoError(prepErr)

	// Give time to the namespace to be ready
	time.Sleep(15 * time.Second)

	// Create operator and run.
	redisfailoverOperator, err := redisfailover.New(redisfailover.Config{}, k8sservice, k8sClient, currentNamespace, redisClient, metrics.Dummy, log.Dummy)
	require.NoError(err)

	go func() {
		errC <- redisfailoverOperator.Run(context.Background())
	}()

	// Prepare cleanup for when the test ends
	defer clients.cleanup(stopC, currentNamespace)

	// Give time to the operator to start
	time.Sleep(15 * time.Second)

	// Create secret
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      authSecretPath,
			Namespace: currentNamespace,
		},
		Data: map[string][]byte{
			"password": []byte(testPass),
		},
	}
	_, err = k8sClient.CoreV1().Secrets(currentNamespace).Create(context.Background(), secret, metav1.CreateOptions{})
	require.NoError(err)

	// Check that if we create a RedisFailover, it is certainly created and we can get it
	ok := t.Run("Check Custom Resource Creation", func(t *testing.T) {
		clients.testCRCreation(t, currentNamespace, disableMyMaster)
	})
	require.True(ok, "the custom resource has to be created to continue")

	// Giving time to the operator to create the resources
	time.Sleep(3 * time.Minute)

	// Verify that auth is set and actually working
	t.Run("Check that auth is set in sentinel and redis configs", func(t *testing.T) {
		clients.testAuth(t, currentNamespace)
	})

	// Check custom config is set
	t.Run("Check that custom config is behave expected", func(t *testing.T) {
		clients.testCustomConfig(t, currentNamespace)
	})

	// Check that a Redis Statefulset is created and the size of it is the one defined by the
	// Redis Failover definition created before.
	t.Run("Check Redis Statefulset existing and size", func(t *testing.T) {
		clients.testRedisStatefulSet(t, currentNamespace)
	})

	// Check that a Sentinel Deployment is created and the size of it is the one defined by the
	// Redis Failover definition created before.
	t.Run("Check Sentinel Deployment existing and size", func(t *testing.T) {
		clients.testSentinelDeployment(t, currentNamespace)
	})

	// Connect to all the Redis pods and, asking to the Redis running inside them, check
	// that only one of them is the master of the failover.
	t.Run("Check Only One Redis Master", func(t *testing.T) {
		clients.testRedisMaster(t, currentNamespace)
	})

	// Connect to all the Sentinel pods and, asking to the Sentinel running inside them,
	// check that all of them are connected to the same Redis node, and also that that node
	// is the master.
	t.Run("Check Sentinels Checking the Redis Master", func(t *testing.T) {
		clients.testSentinelMonitoring(t, currentNamespace, disableMyMaster)
	})
}

func (c *clients) testCRCreation(t *testing.T, currentNamespace string, args ...bool) {
	disableMyMaster := false
	if len(args) > 0 && args[0] {
		disableMyMaster = args[0]
	}
	assert := assert.New(t)
	toCreate := &redisfailoverv1.RedisFailover{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: currentNamespace,
		},
		Spec: redisfailoverv1.RedisFailoverSpec{
			Redis: redisfailoverv1.RedisSettings{
				Replicas: redisSize,
				Exporter: redisfailoverv1.Exporter{
					Enabled: true,
				},
				CustomConfig: []string{`save ""`},
			},
			Sentinel: redisfailoverv1.SentinelSettings{
				Replicas:        sentinelSize,
				DisableMyMaster: disableMyMaster,
			},
			Auth: redisfailoverv1.AuthSettings{
				SecretPath: authSecretPath,
			},
		},
	}

	c.rfClient.DatabasesV1().RedisFailovers(currentNamespace).Create(context.Background(), toCreate, metav1.CreateOptions{})
	gotRF, err := c.rfClient.DatabasesV1().RedisFailovers(currentNamespace).Get(context.Background(), name, metav1.GetOptions{})

	assert.NoError(err)
	assert.Equal(toCreate.Spec, gotRF.Spec)
}

func (c *clients) testRedisStatefulSet(t *testing.T, currentNamespace string) {
	assert := assert.New(t)
	redisSS, err := c.k8sClient.AppsV1().StatefulSets(currentNamespace).Get(context.Background(), fmt.Sprintf("rfr-%s", name), metav1.GetOptions{})
	assert.NoError(err)
	assert.Equal(redisSize, int32(redisSS.Status.Replicas))
}

func (c *clients) testSentinelDeployment(t *testing.T, currentNamespace string) {
	assert := assert.New(t)
	sentinelD, err := c.k8sClient.AppsV1().Deployments(currentNamespace).Get(context.Background(), fmt.Sprintf("rfs-%s", name), metav1.GetOptions{})
	assert.NoError(err)
	assert.Equal(3, int(sentinelD.Status.Replicas))
}

func (c *clients) testRedisMaster(t *testing.T, currentNamespace string) {
	assert := assert.New(t)
	masters := []string{}

	redisSS, err := c.k8sClient.AppsV1().StatefulSets(currentNamespace).Get(context.Background(), fmt.Sprintf("rfr-%s", name), metav1.GetOptions{})
	assert.NoError(err)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.FormatLabels(redisSS.Spec.Selector.MatchLabels),
	}

	redisPodList, err := c.k8sClient.CoreV1().Pods(currentNamespace).List(context.Background(), listOptions)

	assert.NoError(err)

	for _, pod := range redisPodList.Items {
		ip := pod.Status.PodIP
		if ok, _ := c.redisClient.IsMaster(ip, "6379", testPass); ok {
			masters = append(masters, ip)
		}
	}

	assert.Equal(1, len(masters), "only one master expected")
}

func (c *clients) testSentinelMonitoring(t *testing.T, currentNamespace string, args ...bool) {
	disableMyMaster := false
	if len(args) > 0 {
		disableMyMaster = args[0]
	}

	masterName := "mymaster"
	if disableMyMaster {
		masterName = name
	}

	assert := assert.New(t)
	masters := []string{}

	sentinelD, err := c.k8sClient.AppsV1().Deployments(currentNamespace).Get(context.Background(), fmt.Sprintf("rfs-%s", name), metav1.GetOptions{})
	assert.NoError(err)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.FormatLabels(sentinelD.Spec.Selector.MatchLabels),
	}
	sentinelPodList, err := c.k8sClient.CoreV1().Pods(currentNamespace).List(context.Background(), listOptions)
	assert.NoError(err)

	for _, pod := range sentinelPodList.Items {
		ip := pod.Status.PodIP
		master, _, _ := c.redisClient.GetSentinelMonitor(ip, masterName)
		masters = append(masters, master)
	}

	for _, masterIP := range masters {
		assert.Equal(masters[0], masterIP, "all master ip monitoring should equal")
	}

	isMaster, err := c.redisClient.IsMaster(masters[0], "6379", testPass)
	assert.NoError(err)
	assert.True(isMaster, "Sentinel should monitor the Redis master")
}

func (c *clients) testAuth(t *testing.T, currentNamespace string) {
	assert := assert.New(t)

	redisCfg, err := c.k8sClient.CoreV1().ConfigMaps(currentNamespace).Get(context.Background(), fmt.Sprintf("rfr-%s", name), metav1.GetOptions{})
	assert.NoError(err)
	assert.Contains(redisCfg.Data["redis.conf"], "requirepass "+testPass)
	assert.Contains(redisCfg.Data["redis.conf"], "masterauth "+testPass)

	redisSS, err := c.k8sClient.AppsV1().StatefulSets(currentNamespace).Get(context.Background(), fmt.Sprintf("rfr-%s", name), metav1.GetOptions{})
	assert.NoError(err)

	assert.Len(redisSS.Spec.Template.Spec.Containers, 2)
	assert.Equal(redisSS.Spec.Template.Spec.Containers[1].Env[1].Name, "REDIS_ADDR")
	assert.Equal(redisSS.Spec.Template.Spec.Containers[1].Env[1].Value, redisAddr)
	assert.Equal(redisSS.Spec.Template.Spec.Containers[1].Env[2].Name, "REDIS_PORT")
	assert.Equal(redisSS.Spec.Template.Spec.Containers[1].Env[2].Value, "6379")
	assert.Equal(redisSS.Spec.Template.Spec.Containers[1].Env[3].Name, "REDIS_USER")
	assert.Equal(redisSS.Spec.Template.Spec.Containers[1].Env[3].Value, "default")
	assert.Equal(redisSS.Spec.Template.Spec.Containers[1].Env[4].Name, "REDIS_PASSWORD")
	assert.Equal(redisSS.Spec.Template.Spec.Containers[1].Env[4].ValueFrom.SecretKeyRef.Key, "password")
	assert.Equal(redisSS.Spec.Template.Spec.Containers[1].Env[4].ValueFrom.SecretKeyRef.LocalObjectReference.Name, authSecretPath)
}

func (c *clients) testCustomConfig(t *testing.T, currentNamespace string) {
	assert := assert.New(t)

	redisSS, err := c.k8sClient.AppsV1().StatefulSets(currentNamespace).Get(context.Background(), fmt.Sprintf("rfr-%s", name), metav1.GetOptions{})
	assert.NoError(err)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.FormatLabels(redisSS.Spec.Selector.MatchLabels),
	}
	redisPodList, err := c.k8sClient.CoreV1().Pods(currentNamespace).List(context.Background(), listOptions)
	assert.NoError(err)

	rClient := rediscli.NewClient(&rediscli.Options{
		Addr:     net.JoinHostPort(redisPodList.Items[0].Status.PodIP, "6379"),
		Password: testPass,
		DB:       0,
	})
	defer rClient.Close()

	result := rClient.ConfigGet(context.TODO(), "save")
	assert.NoError(result.Err())

	values, err := result.Result()
	assert.NoError(err)

	assert.Len(values, 2)
	assert.Empty(values[1])
}
