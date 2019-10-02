// +build integration

package redisfailover_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	kmetrics "github.com/spotahome/kooper/monitoring/metrics"
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
	namespace      = "rf-integration-tests"
	redisSize      = int32(3)
	sentinelSize   = int32(3)
	authSecretPath = "redis-auth"
	testPass       = "test-pass"
)

type clients struct {
	k8sClient   kubernetes.Interface
	rfClient    redisfailoverclientset.Interface
	aeClient    apiextensionsclientset.Interface
	redisClient redis.Client
}

func (c *clients) prepareNS() error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := c.k8sClient.CoreV1().Namespaces().Create(ns)
	return err
}

func (c *clients) cleanup(stopC chan struct{}) {
	c.k8sClient.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
	close(stopC)
}

func TestRedisFailover(t *testing.T) {
	require := require.New(t)

	// Create signal channels.
	stopC := make(chan struct{})
	errC := make(chan error)

	flags := &utils.CMDFlags{
		KubeConfig:  filepath.Join(homedir.HomeDir(), ".kube", "config"),
		Development: true,
	}

	// Kubernetes clients.
	stdclient, customclient, aeClientset, err := utils.CreateKubernetesClients(flags)
	require.NoError(err)

	// Create the redis clients
	redisClient := redis.New()

	clients := clients{
		k8sClient:   stdclient,
		rfClient:    customclient,
		aeClient:    aeClientset,
		redisClient: redisClient,
	}

	// Create kubernetes service.
	k8sservice := k8s.New(stdclient, customclient, aeClientset, log.Dummy)

	// Prepare namespace
	prepErr := clients.prepareNS()
	require.NoError(prepErr)

	// Give time to the namespace to be ready
	time.Sleep(15 * time.Second)

	// Create operator and run.
	redisfailoverOperator := redisfailover.New(redisfailover.Config{}, k8sservice, redisClient, metrics.Dummy, kmetrics.Dummy, log.Dummy)
	go func() {
		errC <- redisfailoverOperator.Run(stopC)
	}()

	// Prepare cleanup for when the test ends
	defer clients.cleanup(stopC)

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
	_, err = stdclient.CoreV1().Secrets(namespace).Create(secret)
	require.NoError(err)

	// Check that if we create a RedisFailover, it is certainly created and we can get it
	ok := t.Run("Check Custom Resource Creation", clients.testCRCreation)
	require.True(ok, "the custom resource has to be created to continue")

	// Giving time to the operator to create the resources
	time.Sleep(3 * time.Minute)

	// Verify that auth is set and actually working
	t.Run("Check that auth is set and able to connect to redis", clients.testAuth)

	// Check that a Redis Statefulset is created and the size of it is the one defined by the
	// Redis Failover definition created before.
	t.Run("Check Redis Statefulset existing and size", clients.testRedisStatefulSet)

	// Check that a Sentinel Deployment is created and the size of it is the one defined by the
	// Redis Failover definition created before.
	t.Run("Check Sentinel Deployment existing and size", clients.testSentinelDeployment)

	// Connect to all the Redis pods and, asking to the Redis running inside them, check
	// that only one of them is the master of the failover.
	t.Run("Check Only One Redis Master", clients.testRedisMaster)

	// Connect to all the Sentinel pods and, asking to the Sentinel running inside them,
	// check that all of them are connected to the same Redis node, and also that that node
	// is the master.
	t.Run("Check Sentinels Checking the Redis Master", clients.testSentinelMonitoring)

}

func (c *clients) testCRCreation(t *testing.T) {
	assert := assert.New(t)
	toCreate := &redisfailoverv1.RedisFailover{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: redisfailoverv1.RedisFailoverSpec{
			Redis: redisfailoverv1.RedisSettings{
				Replicas: redisSize,
			},
			Sentinel: redisfailoverv1.SentinelSettings{
				Replicas: sentinelSize,
			},
			Auth: redisfailoverv1.AuthSettings{
				SecretPath: authSecretPath,
			},
		},
	}

	c.rfClient.DatabasesV1().RedisFailovers(namespace).Create(toCreate)
	gotRF, err := c.rfClient.DatabasesV1().RedisFailovers(namespace).Get(name, metav1.GetOptions{})

	assert.NoError(err)
	assert.Equal(toCreate.Spec, gotRF.Spec)
}

func (c *clients) testRedisStatefulSet(t *testing.T) {
	assert := assert.New(t)
	redisSS, err := c.k8sClient.AppsV1().StatefulSets(namespace).Get(fmt.Sprintf("rfr-%s", name), metav1.GetOptions{})
	assert.NoError(err)
	assert.Equal(redisSize, int32(redisSS.Status.Replicas))
}

func (c *clients) testSentinelDeployment(t *testing.T) {
	assert := assert.New(t)
	sentinelD, err := c.k8sClient.AppsV1().Deployments(namespace).Get(fmt.Sprintf("rfs-%s", name), metav1.GetOptions{})
	assert.NoError(err)
	assert.Equal(3, int(sentinelD.Status.Replicas))
}

func (c *clients) testRedisMaster(t *testing.T) {
	assert := assert.New(t)
	masters := []string{}

	redisSS, err := c.k8sClient.AppsV1().StatefulSets(namespace).Get(fmt.Sprintf("rfr-%s", name), metav1.GetOptions{})
	assert.NoError(err)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.FormatLabels(redisSS.Spec.Selector.MatchLabels),
	}
	redisPodList, err := c.k8sClient.CoreV1().Pods(namespace).List(listOptions)
	assert.NoError(err)

	for _, pod := range redisPodList.Items {
		ip := pod.Status.PodIP
		if ok, _ := c.redisClient.IsMaster(ip, testPass); ok {
			masters = append(masters, ip)
		}
	}

	assert.Equal(1, len(masters), "only one master expected")
}

func (c *clients) testSentinelMonitoring(t *testing.T) {
	assert := assert.New(t)
	masters := []string{}

	sentinelD, err := c.k8sClient.AppsV1().Deployments(namespace).Get(fmt.Sprintf("rfs-%s", name), metav1.GetOptions{})
	assert.NoError(err)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.FormatLabels(sentinelD.Spec.Selector.MatchLabels),
	}
	sentinelPodList, err := c.k8sClient.CoreV1().Pods(namespace).List(listOptions)
	assert.NoError(err)

	for _, pod := range sentinelPodList.Items {
		ip := pod.Status.PodIP
		master, _ := c.redisClient.GetSentinelMonitor(ip)
		masters = append(masters, master)
	}

	for _, masterIP := range masters {
		assert.Equal(masters[0], masterIP, "all master ip monitoring should equal")
	}

	isMaster, err := c.redisClient.IsMaster(masters[0], testPass)
	assert.NoError(err)
	assert.True(isMaster, "Sentinel should monitor the Redis master")
}
