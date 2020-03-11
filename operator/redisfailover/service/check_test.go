package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
	mK8SService "github.com/spotahome/redis-operator/mocks/service/k8s"
	mRedisService "github.com/spotahome/redis-operator/mocks/service/redis"
	rfservice "github.com/spotahome/redis-operator/operator/redisfailover/service"
)

func generateRF() *redisfailoverv1.RedisFailover {
	return &redisfailoverv1.RedisFailover{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: redisfailoverv1.RedisFailoverSpec{
			Redis: redisfailoverv1.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: redisfailoverv1.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}
}

func TestCheckRedisNumberError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSet", namespace, rfservice.GetRedisName(rf)).Once().Return(nil, errors.New(""))
	mr := &mRedisService.Client{}

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckRedisNumber(rf)
	assert.Error(err)
}

func TestCheckRedisNumberFalse(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	wrongNumber := int32(4)
	ss := &appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Replicas: &wrongNumber,
		},
	}
	ms := &mK8SService.Services{}
	ms.On("GetStatefulSet", namespace, rfservice.GetRedisName(rf)).Once().Return(ss, nil)
	mr := &mRedisService.Client{}

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckRedisNumber(rf)
	assert.Error(err)
}

func TestCheckRedisNumberTrue(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	goodNumber := int32(3)
	ss := &appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Replicas: &goodNumber,
		},
	}
	ms := &mK8SService.Services{}
	ms.On("GetStatefulSet", namespace, rfservice.GetRedisName(rf)).Once().Return(ss, nil)
	mr := &mRedisService.Client{}

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckRedisNumber(rf)
	assert.NoError(err)
}

func TestCheckSentinelNumberError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	ms := &mK8SService.Services{}
	ms.On("GetDeployment", namespace, rfservice.GetSentinelName(rf)).Once().Return(nil, errors.New(""))
	mr := &mRedisService.Client{}

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelNumber(rf)
	assert.Error(err)
}

func TestCheckSentinelNumberFalse(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	wrongNumber := int32(4)
	ss := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &wrongNumber,
		},
	}
	ms := &mK8SService.Services{}
	ms.On("GetDeployment", namespace, rfservice.GetSentinelName(rf)).Once().Return(ss, nil)
	mr := &mRedisService.Client{}

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelNumber(rf)
	assert.Error(err)
}

func TestCheckSentinelNumberTrue(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	goodNumber := int32(3)
	ss := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &goodNumber,
		},
	}
	ms := &mK8SService.Services{}
	ms.On("GetDeployment", namespace, rfservice.GetSentinelName(rf)).Once().Return(ss, nil)
	mr := &mRedisService.Client{}

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelNumber(rf)
	assert.NoError(err)
}

func TestCheckAllSlavesFromMasterGetStatefulSetError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(nil, errors.New(""))
	mr := &mRedisService.Client{}

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckAllSlavesFromMaster("", rf)
	assert.Error(err)
}

func TestCheckAllSlavesFromMasterGetSlaveOfError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "",
					Phase: corev1.PodRunning,
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("GetSlaveOf", "", "").Once().Return("", errors.New(""))

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckAllSlavesFromMaster("", rf)
	assert.Error(err)
}

func TestCheckAllSlavesFromMasterDifferentMaster(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
					Phase: corev1.PodRunning,
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("GetSlaveOf", "0.0.0.0", "").Once().Return("1.1.1.1", nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckAllSlavesFromMaster("0.0.0.0", rf)
	assert.Error(err)
}

func TestCheckAllSlavesFromMaster(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
					Phase: corev1.PodRunning,
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("GetSlaveOf", "0.0.0.0", "").Once().Return("1.1.1.1", nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckAllSlavesFromMaster("1.1.1.1", rf)
	assert.NoError(err)
}

func TestCheckSentinelNumberInMemoryGetDeploymentPodsError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	ms := &mK8SService.Services{}
	mr := &mRedisService.Client{}
	mr.On("GetNumberSentinelsInMemory", "1.1.1.1").Once().Return(int32(0), errors.New("expected error"))

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelNumberInMemory("1.1.1.1", rf)
	assert.Error(err)
}

func TestCheckSentinelNumberInMemoryGetNumberSentinelInMemoryError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	ms := &mK8SService.Services{}
	mr := &mRedisService.Client{}
	mr.On("GetNumberSentinelsInMemory", "1.1.1.1").Once().Return(int32(0), errors.New(""))

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelNumberInMemory("1.1.1.1", rf)
	assert.Error(err)
}

func TestCheckSentinelNumberInMemoryNumberMismatch(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	ms := &mK8SService.Services{}
	mr := &mRedisService.Client{}
	mr.On("GetNumberSentinelsInMemory", "1.1.1.1").Once().Return(int32(4), nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelNumberInMemory("1.1.1.1", rf)
	assert.Error(err)
}

func TestCheckSentinelNumberInMemory(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	ms := &mK8SService.Services{}
	mr := &mRedisService.Client{}
	mr.On("GetNumberSentinelsInMemory", "1.1.1.1").Once().Return(int32(3), nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelNumberInMemory("1.1.1.1", rf)
	assert.NoError(err)
}

func TestCheckSentinelSlavesNumberInMemoryGetNumberSentinelSlavesInMemoryError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	ms := &mK8SService.Services{}
	mr := &mRedisService.Client{}
	mr.On("GetNumberSentinelSlavesInMemory", "1.1.1.1").Once().Return(int32(0), errors.New(""))

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelSlavesNumberInMemory("1.1.1.1", rf)
	assert.Error(err)
}

func TestCheckSentinelSlavesNumberInMemoryReplicasMismatch(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	ms := &mK8SService.Services{}
	mr := &mRedisService.Client{}
	mr.On("GetNumberSentinelSlavesInMemory", "1.1.1.1").Once().Return(int32(3), nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelSlavesNumberInMemory("1.1.1.1", rf)
	assert.Error(err)
}

func TestCheckSentinelSlavesNumberInMemory(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()
	rf.Spec.Redis.Replicas = 5

	ms := &mK8SService.Services{}
	mr := &mRedisService.Client{}
	mr.On("GetNumberSentinelSlavesInMemory", "1.1.1.1").Once().Return(int32(4), nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelSlavesNumberInMemory("1.1.1.1", rf)
	assert.NoError(err)
}

func TestCheckSentinelMonitorGetSentinelMonitorError(t *testing.T) {
	assert := assert.New(t)

	ms := &mK8SService.Services{}
	mr := &mRedisService.Client{}
	mr.On("GetSentinelMonitor", "0.0.0.0").Once().Return("", "", errors.New(""))

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelMonitor("0.0.0.0", "1.1.1.1")
	assert.Error(err)
}

func TestCheckSentinelMonitorMismatch(t *testing.T) {
	assert := assert.New(t)

	ms := &mK8SService.Services{}
	mr := &mRedisService.Client{}
	mr.On("GetSentinelMonitor", "0.0.0.0").Once().Return("2.2.2.2", "6379", nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelMonitor("0.0.0.0", "1.1.1.1")
	assert.Error(err)
}

func TestCheckSentinelMonitor(t *testing.T) {
	assert := assert.New(t)

	ms := &mK8SService.Services{}
	mr := &mRedisService.Client{}
	mr.On("GetSentinelMonitor", "0.0.0.0").Once().Return("1.1.1.1", "6379", nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelMonitor("0.0.0.0", "1.1.1.1")
	assert.NoError(err)
}

func TestCheckSentinelMonitorWithPort(t *testing.T) {
	assert := assert.New(t)

	ms := &mK8SService.Services{}
	mr := &mRedisService.Client{}
	mr.On("GetSentinelMonitor", "0.0.0.0").Once().Return("1.1.1.1", "6379", nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelMonitor("0.0.0.0", "1.1.1.1", "6379")
	assert.NoError(err)
}

func TestCheckSentinelMonitorWithPortMismatch(t *testing.T) {
	assert := assert.New(t)

	ms := &mK8SService.Services{}
	mr := &mRedisService.Client{}
	mr.On("GetSentinelMonitor", "0.0.0.0").Once().Return("1.1.1.1", "6379", nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelMonitor("0.0.0.0", "0.0.0.0", "6379")
	assert.Error(err)
}

func TestCheckSentinelMonitorWithPortIPMismatch(t *testing.T) {
	assert := assert.New(t)

	ms := &mK8SService.Services{}
	mr := &mRedisService.Client{}
	mr.On("GetSentinelMonitor", "0.0.0.0").Once().Return("1.1.1.1", "6379", nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	err := checker.CheckSentinelMonitor("0.0.0.0", "1.1.1.1", "6380")
	assert.Error(err)
}

func TestGetMasterIPGetStatefulSetPodsError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(nil, errors.New(""))
	mr := &mRedisService.Client{}

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	_, err := checker.GetMasterIP(rf)
	assert.Error(err)
}

func TestGetMasterIPIsMasterError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
					Phase: corev1.PodRunning,
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("IsMaster", "0.0.0.0", "").Once().Return(false, errors.New(""))

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	_, err := checker.GetMasterIP(rf)
	assert.Error(err)
}

func TestGetMasterIPMultipleMastersError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
					Phase: corev1.PodRunning,
				},
			},
			{
				Status: corev1.PodStatus{
					PodIP: "1.1.1.1",
					Phase: corev1.PodRunning,
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("IsMaster", "0.0.0.0", "").Once().Return(true, nil)
	mr.On("IsMaster", "1.1.1.1", "").Once().Return(true, nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	_, err := checker.GetMasterIP(rf)
	assert.Error(err)
}

func TestGetMasterIP(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
					Phase: corev1.PodRunning,
				},
			},
			{
				Status: corev1.PodStatus{
					PodIP: "1.1.1.1",
					Phase: corev1.PodRunning,
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("IsMaster", "0.0.0.0", "").Once().Return(true, nil)
	mr.On("IsMaster", "1.1.1.1", "").Once().Return(false, nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	master, err := checker.GetMasterIP(rf)
	assert.NoError(err)
	assert.Equal("0.0.0.0", master, "the master should be the expected")
}

func TestGetNumberMastersGetStatefulSetPodsError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(nil, errors.New(""))
	mr := &mRedisService.Client{}

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	_, err := checker.GetNumberMasters(rf)
	assert.Error(err)
}

func TestGetNumberMastersIsMasterError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
					Phase: corev1.PodRunning,
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("IsMaster", "0.0.0.0", "").Once().Return(true, errors.New(""))

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	_, err := checker.GetNumberMasters(rf)
	assert.Error(err)
}

func TestGetNumberMasters(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
					Phase: corev1.PodRunning,
				},
			},
			{
				Status: corev1.PodStatus{
					PodIP: "1.1.1.1",
					Phase: corev1.PodRunning,
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("IsMaster", "0.0.0.0", "").Once().Return(true, nil)
	mr.On("IsMaster", "1.1.1.1", "").Once().Return(false, nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	masterNumber, err := checker.GetNumberMasters(rf)
	assert.NoError(err)
	assert.Equal(1, masterNumber, "the master number should be ok")
}

func TestGetNumberMastersTwo(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
					Phase: corev1.PodRunning,
				},
			},
			{
				Status: corev1.PodStatus{
					PodIP: "1.1.1.1",
					Phase: corev1.PodRunning,
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("IsMaster", "0.0.0.0", "").Once().Return(true, nil)
	mr.On("IsMaster", "1.1.1.1", "").Once().Return(true, nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	masterNumber, err := checker.GetNumberMasters(rf)
	assert.NoError(err)
	assert.Equal(2, masterNumber, "the master number should be ok")
}

func TestGetMinimumRedisPodTimeGetStatefulSetPodsError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(nil, errors.New(""))
	mr := &mRedisService.Client{}

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	_, err := checker.GetMinimumRedisPodTime(rf)
	assert.Error(err)
}

func TestGetMinimumRedisPodTime(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	now := time.Now()
	oneHour := now.Add(-1 * time.Hour)
	oneMinute := now.Add(-1 * time.Minute)

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					StartTime: &metav1.Time{
						Time: oneHour,
					},
				},
			},
			{
				Status: corev1.PodStatus{
					StartTime: &metav1.Time{
						Time: oneMinute,
					},
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})

	minTime, err := checker.GetMinimumRedisPodTime(rf)
	assert.NoError(err)

	expected := now.Sub(oneMinute).Round(time.Second)
	assert.Equal(expected, minTime.Round(time.Second), "the closest time should be given")
}

func TestGetRedisPodsNames(t *testing.T) {
	assert := assert.New(t)
	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "slave1",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					PodIP: "0.0.0.0",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "master",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					PodIP: "1.1.1.1",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "slave2",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					PodIP: "0.0.0.0",
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("IsMaster", "0.0.0.0", "").Twice().Return(false, nil)
	mr.On("IsMaster", "1.1.1.1", "").Once().Return(true, nil)

	checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})
	master, err := checker.GetRedisesMasterPod(rf)

	assert.NoError(err)

	assert.Equal(master, "master")

	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr.On("IsMaster", "0.0.0.0", "").Twice().Return(false, nil)
	mr.On("IsMaster", "1.1.1.1", "").Once().Return(true, nil)

	namePods, err := checker.GetRedisesSlavesPods(rf)

	assert.NoError(err)

	assert.Equal(namePods, []string{"slave1", "slave2"})
}

func TestGetStatefulSetUpdateRevision(t *testing.T) {
	tests := []struct {
		name             string
		ss               *appsv1.StatefulSet
		expectedUVersion string
		expectedError    error
	}{
		{
			name: "revision ok",
			ss: &appsv1.StatefulSet{
				Status: appsv1.StatefulSetStatus{
					UpdateRevision: "10",
				},
			},
			expectedUVersion: "10",
			expectedError:    nil,
		},
		{
			name:             "no stateful set",
			ss:               nil,
			expectedUVersion: "",
			expectedError:    errors.New("not found"),
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		rf := generateRF()
		ms := &mK8SService.Services{}
		ms.On("GetStatefulSet", namespace, rfservice.GetRedisName(rf)).Once().Return(test.ss, nil)
		mr := &mRedisService.Client{}

		checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})
		version, err := checker.GetStatefulSetUpdateRevision(rf)

		if test.expectedError == nil {
			assert.NoError(err)
		} else {
			assert.Error(err)
		}

		assert.Equal(version, test.expectedUVersion)
	}

}

func TestGetRedisRevisionHash(t *testing.T) {
	tests := []struct {
		name          string
		pod           *corev1.Pod
		expectedHash  string
		expectedError error
	}{
		{
			name: "has ok",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						appsv1.ControllerRevisionHashLabelKey: "10",
					},
				},
			},
			expectedHash:  "10",
			expectedError: nil,
		},
		{
			name:          "no pod",
			pod:           nil,
			expectedHash:  "",
			expectedError: errors.New("not found"),
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		rf := generateRF()
		ms := &mK8SService.Services{}
		ms.On("GetPod", namespace, "namepod").Once().Return(test.pod, nil)
		mr := &mRedisService.Client{}

		checker := rfservice.NewRedisFailoverChecker(ms, mr, log.DummyLogger{})
		hash, err := checker.GetRedisRevisionHash("namepod", rf)

		if test.expectedError == nil {
			assert.NoError(err)
		} else {
			assert.Error(err)
		}

		assert.Equal(hash, test.expectedHash)
	}

}
