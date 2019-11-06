package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spotahome/redis-operator/log"
	mK8SService "github.com/spotahome/redis-operator/mocks/service/k8s"
	mRedisService "github.com/spotahome/redis-operator/mocks/service/redis"
	rfservice "github.com/spotahome/redis-operator/operator/redisfailover/service"
)

func TestSetOldestAsMasterNewMasterError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("MakeMaster", "0.0.0.0", "").Once().Return(errors.New(""))

	healer := rfservice.NewRedisFailoverHealer(ms, mr, log.DummyLogger{})

	err := healer.SetOldestAsMaster(rf)
	assert.Error(err)
}

func TestSetOldestAsMaster(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("MakeMaster", "0.0.0.0", "").Once().Return(nil)

	healer := rfservice.NewRedisFailoverHealer(ms, mr, log.DummyLogger{})

	err := healer.SetOldestAsMaster(rf)
	assert.NoError(err)
}

func TestSetOldestAsMasterMultiplePodsMakeSlaveOfError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
				},
			},
			{
				Status: corev1.PodStatus{
					PodIP: "1.1.1.1",
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("MakeMaster", "0.0.0.0", "").Once().Return(nil)
	mr.On("MakeSlaveOf", "1.1.1.1", "0.0.0.0", "").Once().Return(errors.New(""))

	healer := rfservice.NewRedisFailoverHealer(ms, mr, log.DummyLogger{})

	err := healer.SetOldestAsMaster(rf)
	assert.Error(err)
}

func TestSetOldestAsMasterMultiplePods(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
				},
			},
			{
				Status: corev1.PodStatus{
					PodIP: "1.1.1.1",
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("MakeMaster", "0.0.0.0", "").Once().Return(nil)
	mr.On("MakeSlaveOf", "1.1.1.1", "0.0.0.0", "").Once().Return(nil)

	healer := rfservice.NewRedisFailoverHealer(ms, mr, log.DummyLogger{})

	err := healer.SetOldestAsMaster(rf)
	assert.NoError(err)
}

func TestSetOldestAsMasterOrdering(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Time{
						Time: time.Now(),
					},
				},
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Time{
						Time: time.Now().Add(-1 * time.Hour), // This is older by 1 hour
					},
				},
				Status: corev1.PodStatus{
					PodIP: "1.1.1.1",
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("MakeMaster", "1.1.1.1", "").Once().Return(nil)
	mr.On("MakeSlaveOf", "0.0.0.0", "1.1.1.1", "").Once().Return(nil)

	healer := rfservice.NewRedisFailoverHealer(ms, mr, log.DummyLogger{})

	err := healer.SetOldestAsMaster(rf)
	assert.NoError(err)
}

func TestSetMasterOnAllMakeMasterError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
				},
			},
			{
				Status: corev1.PodStatus{
					PodIP: "1.1.1.1",
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("MakeMaster", "0.0.0.0", "").Once().Return(errors.New(""))

	healer := rfservice.NewRedisFailoverHealer(ms, mr, log.DummyLogger{})

	err := healer.SetMasterOnAll("0.0.0.0", rf)
	assert.Error(err)
}

func TestSetMasterOnAllMakeSlaveOfError(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
				},
			},
			{
				Status: corev1.PodStatus{
					PodIP: "1.1.1.1",
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("MakeMaster", "0.0.0.0", "").Once().Return(nil)
	mr.On("MakeSlaveOf", "1.1.1.1", "0.0.0.0", "").Once().Return(errors.New(""))

	healer := rfservice.NewRedisFailoverHealer(ms, mr, log.DummyLogger{})

	err := healer.SetMasterOnAll("0.0.0.0", rf)
	assert.Error(err)
}

func TestSetMasterOnAll(t *testing.T) {
	assert := assert.New(t)

	rf := generateRF()

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				Status: corev1.PodStatus{
					PodIP: "0.0.0.0",
				},
			},
			{
				Status: corev1.PodStatus{
					PodIP: "1.1.1.1",
				},
			},
		},
	}

	ms := &mK8SService.Services{}
	ms.On("GetStatefulSetPods", namespace, rfservice.GetRedisName(rf)).Once().Return(pods, nil)
	mr := &mRedisService.Client{}
	mr.On("MakeMaster", "0.0.0.0", "").Once().Return(nil)
	mr.On("MakeSlaveOf", "1.1.1.1", "0.0.0.0", "").Once().Return(nil)

	healer := rfservice.NewRedisFailoverHealer(ms, mr, log.DummyLogger{})

	err := healer.SetMasterOnAll("0.0.0.0", rf)
	assert.NoError(err)
}
