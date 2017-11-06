package failover_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"

	"github.com/spotahome/redis-operator/mocks"
	"github.com/spotahome/redis-operator/pkg/failover"
	"github.com/spotahome/redis-operator/pkg/log"
	"github.com/spotahome/redis-operator/pkg/metrics"
)

func TestOnStatusGetRedisFailoversError(t *testing.T) {
	// mock our client
	mc := &mocks.RedisFailoverClient{}

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	mc.On("GetAllRedisfailovers").Once().Return(nil, errors.New(""))

	ctrl.OnStatus()

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnStatusSkipNotRunning(t *testing.T) {
	// mock our client
	mc := &mocks.RedisFailoverClient{}

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	rfs := &failover.RedisFailoverList{
		Items: []failover.RedisFailover{
			failover.RedisFailover{
				Status: failover.RedisFailoverStatus{
					Phase: failover.PhaseCreating,
				},
			},
		},
	}

	mc.On("GetAllRedisfailovers").Once().Return(rfs, nil)

	ctrl.OnStatus()

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnStatusSkipNotReady(t *testing.T) {
	// mock our client
	mc := &mocks.RedisFailoverClient{}

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	rfs := &failover.RedisFailoverList{
		Items: []failover.RedisFailover{
			failover.RedisFailover{
				Status: failover.RedisFailoverStatus{
					Phase: failover.PhaseRunning,
					Conditions: []failover.Condition{
						failover.Condition{
							Type: failover.ConditionNotReady,
						},
					},
				},
			},
		},
	}

	mc.On("GetAllRedisfailovers").Once().Return(rfs, nil)

	ctrl.OnStatus()

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnStatusReadyCheckError(t *testing.T) {
	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mck := &mocks.RedisFailoverCheck{}

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, mck)

	rf := failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Version:  "3.2-alpine",
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
		Status: failover.RedisFailoverStatus{
			Phase: failover.PhaseRunning,
			Conditions: []failover.Condition{
				failover.Condition{
					Type: failover.ConditionReady,
				},
			},
		},
	}

	rfs := &failover.RedisFailoverList{
		Items: []failover.RedisFailover{
			rf,
		},
	}

	mc.On("GetAllRedisfailovers").Once().Return(rfs, nil)
	mck.On("Check", &rf).Once().Return(errors.New(""))

	ctrl.OnStatus()

	// assert calls
	mc.AssertExpectations(t)
	mck.AssertExpectations(t)
}

func TestOnStatusReadyCheckGetMasterError(t *testing.T) {
	//TODO
	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mck := &mocks.RedisFailoverCheck{}

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, mck)

	rf := failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Version:  "3.2-alpine",
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
		Status: failover.RedisFailoverStatus{
			Phase: failover.PhaseRunning,
			Conditions: []failover.Condition{
				failover.Condition{
					Type: failover.ConditionReady,
				},
			},
		},
	}

	rfs := &failover.RedisFailoverList{
		Items: []failover.RedisFailover{
			rf,
		},
	}

	mc.On("GetAllRedisfailovers").Once().Return(rfs, nil)
	mck.On("Check", &rf).Once().Return(nil)
	mck.On("GetMaster", &rf).Once().Return("", errors.New(""))

	ctrl.OnStatus()

	// assert calls
	mc.AssertExpectations(t)
	mck.AssertExpectations(t)
}

func TestOnStatusReadyCheckGetMasterSameMaster(t *testing.T) {
	//TODO
	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mck := &mocks.RedisFailoverCheck{}

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, mck)

	master := "1.1.1.1"
	rf := failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Version:  "3.2-alpine",
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
		Status: failover.RedisFailoverStatus{
			Master: master,
			Phase:  failover.PhaseRunning,
			Conditions: []failover.Condition{
				failover.Condition{
					Type: failover.ConditionReady,
				},
			},
		},
	}

	rfs := &failover.RedisFailoverList{
		Items: []failover.RedisFailover{
			rf,
		},
	}

	mc.On("GetAllRedisfailovers").Once().Return(rfs, nil)
	mck.On("Check", &rf).Once().Return(nil)
	mck.On("GetMaster", &rf).Once().Return(master, nil)

	ctrl.OnStatus()

	// assert calls
	mc.AssertExpectations(t)
	mck.AssertExpectations(t)
}

func TestOnStatusReadyCheckOk(t *testing.T) {
	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mck := &mocks.RedisFailoverCheck{}
	ml := &mocks.Logger{}

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, ml, &failover.RedisFailoverTransformer{}, mck)

	master := "1.1.1.1"

	rf := failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Status: failover.RedisFailoverStatus{
			Master: "0.0.0.0",
			Phase:  failover.PhaseRunning,
			Conditions: []failover.Condition{
				failover.Condition{
					Type: failover.ConditionReady,
				},
			},
		},
	}

	rfs := &failover.RedisFailoverList{
		Items: []failover.RedisFailover{
			rf,
		},
	}

	mc.On("GetAllRedisfailovers").Once().Return(rfs, nil)
	mck.On("Check", mock.Anything).Once().Return(nil)
	ml.On("Debugf", "RedisFailover %s in namespace %s is ok", "test", "test").Once()
	ml.On("WithField", mock.Anything, mock.Anything).Return(ml)
	mck.On("GetMaster", mock.Anything).Once().Return(master, nil)
	mc.On("UpdateStatus", mock.Anything).Once().Return(nil, nil)

	ctrl.OnStatus()

	// assert calls
	mc.AssertExpectations(t)
	mck.AssertExpectations(t)
	ml.AssertExpectations(t)
}

func TestOnAddRedisReplicasError(t *testing.T) {
	// mock our client
	mc := &mocks.RedisFailoverClient{}

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

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
	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddSentinelReplicasError(t *testing.T) {
	// mock our client
	mc := &mocks.RedisFailoverClient{}

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

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
				Replicas: int32(1),
			},
		},
	}
	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddRunningGetRedisError(t *testing.T) {
	sSettings := failover.SentinelSettings{
		Replicas: int32(3),
	}
	rSettings := failover.RedisSettings{
		Replicas: int32(3),
	}
	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis:    rSettings,
			Sentinel: sSettings,
		},
		Status: failover.RedisFailoverStatus{
			Phase: failover.PhaseRunning,
		},
	}

	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", redisFailover).
		Once().Return(nil, errors.New(""))

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddRunningGetSentinelError(t *testing.T) {
	sSettings := failover.SentinelSettings{
		Replicas: int32(3),
	}
	rSettings := failover.RedisSettings{
		Replicas: int32(3),
	}
	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis:    rSettings,
			Sentinel: sSettings,
		},
		Status: failover.RedisFailoverStatus{
			Phase: failover.PhaseRunning,
		},
	}

	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", redisFailover).
		Once().Return(&v1beta1.StatefulSet{}, nil)
	mc.On("GetSentinelDeployment", redisFailover).
		Once().Return(nil, errors.New(""))

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddRunningTransformStatefulsetToRedisSettings(t *testing.T) {
	sSettings := failover.SentinelSettings{
		Replicas: int32(3),
	}
	rSettings := failover.RedisSettings{
		Replicas: int32(3),
	}
	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis:    rSettings,
			Sentinel: sSettings,
		},
		Status: failover.RedisFailoverStatus{
			Phase: failover.PhaseRunning,
		},
	}

	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", redisFailover).
		Once().Return(&v1beta1.StatefulSet{}, nil)
	mc.On("GetSentinelDeployment", redisFailover).
		Once().Return(&v1beta1.Deployment{}, nil)
	mt := &mocks.Transformer{}
	mt.On("StatefulsetToRedisSettings", mock.Anything).Return(nil, errors.New(""))

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, mt, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddRunningTransformDeploymentToSentinelSettings(t *testing.T) {
	sSettings := failover.SentinelSettings{
		Replicas: int32(3),
	}
	rSettings := failover.RedisSettings{
		Replicas: int32(3),
	}
	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis:    rSettings,
			Sentinel: sSettings,
		},
		Status: failover.RedisFailoverStatus{
			Phase: failover.PhaseRunning,
		},
	}

	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", redisFailover).
		Once().Return(&v1beta1.StatefulSet{}, nil)
	mc.On("GetSentinelDeployment", redisFailover).
		Once().Return(&v1beta1.Deployment{}, nil)
	mt := &mocks.Transformer{}
	mt.On("StatefulsetToRedisSettings", mock.Anything).Return(&failover.RedisSettings{}, nil)
	mt.On("DeploymentToSentinelSettings", mock.Anything).Return(nil, errors.New(""))

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, mt, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddRunningCallOnUpdate(t *testing.T) {
	sSettings := failover.SentinelSettings{
		Replicas: int32(3),
	}
	rSettings := failover.RedisSettings{
		Replicas: int32(3),
	}
	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis:    rSettings,
			Sentinel: sSettings,
		},
		Status: failover.RedisFailoverStatus{
			Phase: failover.PhaseRunning,
		},
	}

	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", redisFailover).
		Once().Return(&v1beta1.StatefulSet{}, nil)
	mc.On("GetSentinelDeployment", redisFailover).
		Once().Return(&v1beta1.Deployment{}, nil)
	mt := &mocks.Transformer{}
	mt.On("StatefulsetToRedisSettings", mock.Anything).Return(&failover.RedisSettings{}, nil)
	mt.On("DeploymentToSentinelSettings", mock.Anything).Return(&failover.RedisSettings{}, nil)

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddAlreadyPresent(t *testing.T) {
	sSettings := failover.SentinelSettings{
		Replicas: int32(3),
	}
	rSettings := failover.RedisSettings{
		Replicas: int32(3),
	}
	redisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis:    rSettings,
			Sentinel: sSettings,
		},
		Status: failover.RedisFailoverStatus{
			Phase: failover.PhaseRunning,
		},
	}

	mt := &mocks.Transformer{}
	mt.On("StatefulsetToRedisSettings", mock.Anything).Return(&rSettings, nil)
	mt.On("DeploymentToSentinelSettings", mock.Anything).Return(&sSettings, nil)

	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", redisFailover).
		Once().Return(&v1beta1.StatefulSet{}, nil)
	mc.On("GetSentinelDeployment", redisFailover).
		Once().Return(&v1beta1.Deployment{}, nil)

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, mt, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddUpdateStatusError(t *testing.T) {
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

	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, errors.New(""))

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddBootstrapError(t *testing.T) {
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

	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)
	mc.On("GetBootstrapPod", redisFailover).
		Once().Return(nil, errors.New(""))
	mc.On("GetSentinelDeployment", redisFailover).
		Once().Return(nil, errors.New(""))
	mc.On("GetRedisStatefulset", redisFailover).
		Once().Return(nil, errors.New(""))
	mc.On("CreateBootstrapPod", redisFailover).
		Once().Return(errors.New(""))
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddSentinelServiceError(t *testing.T) {
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

	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)
	mc.On("GetBootstrapPod", redisFailover).
		Once().Return(&v1.Pod{}, nil)
	mc.On("GetSentinelService", redisFailover).
		Once().Return(nil, errors.New(""))
	mc.On("CreateSentinelService", redisFailover).
		Once().Return(errors.New(""))
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddSentinelDeploymentError(t *testing.T) {
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

	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)
	mc.On("GetBootstrapPod", redisFailover).
		Once().Return(&v1.Pod{}, nil)
	mc.On("GetSentinelService", redisFailover).
		Once().Return(&v1.Service{}, nil)
	mc.On("GetSentinelDeployment", redisFailover).
		Once().Return(nil, errors.New(""))
	mc.On("CreateSentinelDeployment", redisFailover).
		Once().Return(errors.New(""))
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddRedisServiceError(t *testing.T) {
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

	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)
	mc.On("GetBootstrapPod", redisFailover).
		Once().Return(&v1.Pod{}, nil)
	mc.On("GetSentinelService", redisFailover).
		Once().Return(&v1.Service{}, nil)
	mc.On("GetSentinelDeployment", redisFailover).
		Once().Return(&v1beta1.Deployment{}, nil)
	mc.On("GetRedisService", redisFailover).
		Once().Return(nil, errors.New(""))
	mc.On("CreateRedisService", redisFailover).
		Once().Return(errors.New(""))
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddRedisStatefulsetError(t *testing.T) {
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

	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)
	mc.On("GetBootstrapPod", redisFailover).
		Once().Return(&v1.Pod{}, nil)
	mc.On("GetSentinelService", redisFailover).
		Once().Return(&v1.Service{}, nil)
	mc.On("GetSentinelDeployment", redisFailover).
		Once().Return(&v1beta1.Deployment{}, nil)
	mc.On("GetRedisStatefulset", redisFailover).
		Once().Return(nil, errors.New(""))
	mc.On("CreateRedisStatefulset", redisFailover).
		Once().Return(errors.New(""))
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddDeleteBootstrapError(t *testing.T) {
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

	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)
	mc.On("GetBootstrapPod", redisFailover).
		Once().Return(&v1.Pod{}, nil)
	mc.On("GetSentinelService", redisFailover).
		Once().Return(&v1.Service{}, nil)
	mc.On("GetSentinelDeployment", redisFailover).
		Once().Return(&v1beta1.Deployment{}, nil)
	mc.On("GetRedisStatefulset", redisFailover).
		Once().Return(&v1beta1.StatefulSet{}, nil)
	mc.On("GetBootstrapPod", redisFailover).
		Once().Return(&v1.Pod{}, nil)
	mc.On("DeleteBootstrapPod", redisFailover).
		Once().Return(errors.New(""))
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddUpdateStatusRunningError(t *testing.T) {
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

	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)
	mc.On("GetBootstrapPod", redisFailover).
		Once().Return(&v1.Pod{}, nil)
	mc.On("GetSentinelService", redisFailover).
		Once().Return(&v1.Service{}, nil)
	mc.On("GetSentinelDeployment", redisFailover).
		Once().Return(&v1beta1.Deployment{}, nil)
	mc.On("GetRedisStatefulset", redisFailover).
		Once().Return(&v1beta1.StatefulSet{}, nil)
	mc.On("GetBootstrapPod", redisFailover).
		Once().Return(nil, errors.New(""))
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, errors.New(""))

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnAddCreateOK(t *testing.T) {
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

	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)
	mc.On("GetBootstrapPod", redisFailover).
		Once().Return(&v1.Pod{}, nil)
	mc.On("GetSentinelService", redisFailover).
		Once().Return(&v1.Service{}, nil)
	mc.On("GetSentinelDeployment", redisFailover).
		Once().Return(&v1beta1.Deployment{}, nil)
	mc.On("GetRedisStatefulset", redisFailover).
		Once().Return(&v1beta1.StatefulSet{}, nil)
	mc.On("GetBootstrapPod", redisFailover).
		Once().Return(nil, errors.New(""))
	mc.On("UpdateStatus", redisFailover).
		Once().Return(nil, nil)

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnAdd(redisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnUpdateSpecsEqual(t *testing.T) {
	oldRedisFailover := &failover.RedisFailover{
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

	newRedisFailover := &failover.RedisFailover{
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

	// mock our client
	mc := &mocks.RedisFailoverClient{}

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnUpdate(oldRedisFailover, newRedisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnUpdateSentinelError(t *testing.T) {
	oldRedisFailover := &failover.RedisFailover{
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

	newRedisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(3),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(4),
			},
		},
	}

	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mc.On("UpdateStatus", newRedisFailover).
		Once().Return(nil, nil)
	mc.On("UpdateSentinelDeployment", newRedisFailover).
		Once().Return(errors.New(""))

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnUpdate(oldRedisFailover, newRedisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnUpdateRedisError(t *testing.T) {
	oldRedisFailover := &failover.RedisFailover{
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

	newRedisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(4),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(3),
			},
		},
	}

	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mc.On("UpdateStatus", newRedisFailover).
		Once().Return(nil, nil)
	mc.On("UpdateRedisStatefulset", newRedisFailover).
		Once().Return(errors.New(""))

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnUpdate(oldRedisFailover, newRedisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnUpdateOK(t *testing.T) {
	oldRedisFailover := &failover.RedisFailover{
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

	newRedisFailover := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: int32(4),
			},
			Sentinel: failover.SentinelSettings{
				Replicas: int32(4),
			},
		},
	}

	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mc.On("UpdateStatus", newRedisFailover).
		Once().Return(nil, nil)
	mc.On("UpdateSentinelDeployment", newRedisFailover).
		Once().Return(nil)
	mc.On("UpdateStatus", newRedisFailover).
		Once().Return(nil, nil)
	mc.On("UpdateRedisStatefulset", newRedisFailover).
		Once().Return(nil)
	mc.On("UpdateStatus", newRedisFailover).
		Once().Return(nil, nil)

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	ctrl.OnUpdate(oldRedisFailover, newRedisFailover)

	// assert calls
	mc.AssertExpectations(t)
}

func TestOnDeleteOK(t *testing.T) {
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

	// mock our client
	mc := &mocks.RedisFailoverClient{}
	mc.On("DeleteRedisService", redisFailover).Once().Return(nil)
	mc.On("DeleteRedisStatefulset", redisFailover).Once().Return(nil)
	mc.On("DeleteSentinelService", redisFailover).Once().Return(nil)
	mc.On("DeleteSentinelDeployment", redisFailover).Once().Return(nil)

	// Call our controller as if we where a k8s watcher informer
	ctrl := failover.NewRedisFailoverController(metrics.Dummy, mc, log.Nil, &failover.RedisFailoverTransformer{}, &failover.RedisFailoverChecker{})

	obj := &failover.RedisFailover{
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
	ctrl.OnDelete(obj)

	// assert calls
	mc.AssertExpectations(t)
}
