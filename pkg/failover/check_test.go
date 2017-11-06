package failover_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"

	"github.com/spotahome/redis-operator/mocks"
	"github.com/spotahome/redis-operator/pkg/clock"
	"github.com/spotahome/redis-operator/pkg/failover"
	"github.com/spotahome/redis-operator/pkg/log"
	"github.com/spotahome/redis-operator/pkg/metrics"
	"github.com/spotahome/redis-operator/pkg/redis"
)

func TestCheckRedisNumberError(t *testing.T) {
	assert := assert.New(t)
	rf := &failover.RedisFailover{
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
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", rf).Once().Return(nil, errors.New(""))

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, redis.New(), clock.Base(), log.Nil)

	err := checker.Check(rf)
	assert.Error(err)
}

func TestCheckRedisNumberFalse(t *testing.T) {
	assert := assert.New(t)
	rf := &failover.RedisFailover{
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
	wrongNumber := int32(4)
	ss := &v1beta1.StatefulSet{
		Spec: v1beta1.StatefulSetSpec{
			Replicas: &wrongNumber,
		},
	}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", rf).Once().Return(ss, nil)

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, redis.New(), clock.Base(), log.Nil)

	err := checker.Check(rf)
	assert.Error(err)
}

func TestCheckSentinelNumberError(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	ss := &v1beta1.StatefulSet{
		Spec: v1beta1.StatefulSetSpec{
			Replicas: &correctNumber,
		},
	}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", rf).Once().Return(ss, nil)
	mc.On("GetSentinelDeployment", rf).Once().Return(nil, errors.New(""))

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, redis.New(), clock.Base(), log.Nil)

	err := checker.Check(rf)
	assert.Error(err)
}

func TestCheckSentinelNumberFalse(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	wrongNumber := int32(4)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	ss := &v1beta1.StatefulSet{
		Spec: v1beta1.StatefulSetSpec{
			Replicas: &correctNumber,
		},
	}
	d := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Replicas: &wrongNumber,
		},
	}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", rf).Once().Return(ss, nil)
	mc.On("GetSentinelDeployment", rf).Once().Return(d, nil)

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, redis.New(), clock.Base(), log.Nil)

	err := checker.Check(rf)
	assert.Error(err)
}

func TestCheckSentinelGetMasterError(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	ss := &v1beta1.StatefulSet{
		Spec: v1beta1.StatefulSetSpec{
			Replicas: &correctNumber,
		},
	}
	d := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Replicas: &correctNumber,
		},
	}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", rf).Once().Return(ss, nil)
	mc.On("GetSentinelDeployment", rf).Once().Return(d, nil)
	mc.On("GetRedisPodsIPs", rf).Once().Return(nil, errors.New(""))

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, redis.New(), clock.Base(), log.Nil)

	err := checker.Check(rf)
	assert.Error(err)
}

func TestCheckSentinelGetRedisPodsIPsError(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	ss := &v1beta1.StatefulSet{
		Spec: v1beta1.StatefulSetSpec{
			Replicas: &correctNumber,
		},
	}
	d := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Replicas: &correctNumber,
		},
	}
	redisPods := []string{"0.0.0.0", "1.1.1.1"}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", rf).Once().Return(ss, nil)
	mc.On("GetSentinelDeployment", rf).Once().Return(d, nil)
	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr := &mocks.Client{}
	mr.On("IsMaster", redisPods[0]).Once().Return(true, nil)
	mr.On("IsMaster", redisPods[1]).Once().Return(false, nil)

	mc.On("GetRedisPodsIPs", rf).Once().Return(nil, errors.New(""))

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, mr, clock.Base(), log.Nil)

	err := checker.Check(rf)
	assert.Error(err)
}

func TestCheckSentinelSlaveOfError(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	ss := &v1beta1.StatefulSet{
		Spec: v1beta1.StatefulSetSpec{
			Replicas: &correctNumber,
		},
	}
	d := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Replicas: &correctNumber,
		},
	}
	redisPods := []string{"0.0.0.0", "1.1.1.1"}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", rf).Once().Return(ss, nil)
	mc.On("GetSentinelDeployment", rf).Once().Return(d, nil)
	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr := &mocks.Client{}
	mr.On("IsMaster", redisPods[0]).Once().Return(true, nil)
	mr.On("IsMaster", redisPods[1]).Once().Return(false, nil)

	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr.On("GetSlaveOf", redisPods[0]).Once().Return("", errors.New(""))

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, mr, clock.Base(), log.Nil)

	err := checker.Check(rf)
	assert.Error(err)
}

func TestCheckSentinelSlaveOfSlaveEmptyError(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	ss := &v1beta1.StatefulSet{
		Spec: v1beta1.StatefulSetSpec{
			Replicas: &correctNumber,
		},
	}
	d := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Replicas: &correctNumber,
		},
	}
	redisPods := []string{"0.0.0.0", "1.1.1.1"}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", rf).Once().Return(ss, nil)
	mc.On("GetSentinelDeployment", rf).Once().Return(d, nil)
	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr := &mocks.Client{}
	mr.On("IsMaster", redisPods[0]).Once().Return(true, nil)
	mr.On("IsMaster", redisPods[1]).Once().Return(false, nil)

	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr.On("GetSlaveOf", redisPods[0]).Once().Return("2.2.2.2", nil)

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, mr, clock.Base(), log.Nil)

	err := checker.Check(rf)
	assert.Error(err)
}

func TestCheckSentinelNumberInMemoryGetSentinelPodsIPsError(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	ss := &v1beta1.StatefulSet{
		Spec: v1beta1.StatefulSetSpec{
			Replicas: &correctNumber,
		},
	}
	d := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Replicas: &correctNumber,
		},
	}
	redisPods := []string{"0.0.0.0", "1.1.1.1"}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", rf).Once().Return(ss, nil)
	mc.On("GetSentinelDeployment", rf).Once().Return(d, nil)
	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr := &mocks.Client{}
	mr.On("IsMaster", redisPods[0]).Once().Return(true, nil)
	mr.On("IsMaster", redisPods[1]).Once().Return(false, nil)

	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr.On("GetSlaveOf", redisPods[0]).Once().Return("", nil)
	mr.On("GetSlaveOf", redisPods[1]).Once().Return(redisPods[0], nil)

	mc.On("GetSentinelPodsIPs", rf).Once().Return(nil, errors.New(""))

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, mr, clock.Base(), log.Nil)

	err := checker.Check(rf)
	assert.Error(err)
}

func TestCheckSentinelNumberInMemoryResetSentinelErrorNoReset(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	ss := &v1beta1.StatefulSet{
		Spec: v1beta1.StatefulSetSpec{
			Replicas: &correctNumber,
		},
	}
	d := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Replicas: &correctNumber,
		},
	}
	redisPods := []string{"0.0.0.0", "1.1.1.1"}
	sentinelPods := []string{"9.9.9.9"}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", rf).Once().Return(ss, nil)
	mc.On("GetSentinelDeployment", rf).Once().Return(d, nil)
	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr := &mocks.Client{}
	mr.On("IsMaster", redisPods[0]).Once().Return(true, nil)
	mr.On("IsMaster", redisPods[1]).Once().Return(false, nil)

	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr.On("GetSlaveOf", redisPods[0]).Once().Return("", nil)
	mr.On("GetSlaveOf", redisPods[1]).Once().Return(redisPods[0], nil)

	mc.On("GetSentinelPodsIPs", rf).Once().Return(sentinelPods, nil)

	mr.On("GetNumberSentinelsInMemory", sentinelPods[0]).Once().Return(int32(0), errors.New(""))

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, mr, clock.Base(), log.Nil)

	err := checker.Check(rf)
	assert.NoError(err)
}

func TestCheckSentinelNumberInMemoryResetSentinelNumberOk(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	ss := &v1beta1.StatefulSet{
		Spec: v1beta1.StatefulSetSpec{
			Replicas: &correctNumber,
		},
	}
	d := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Replicas: &correctNumber,
		},
	}
	redisPods := []string{"0.0.0.0", "1.1.1.1"}
	sentinelPods := []string{"9.9.9.9"}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", rf).Once().Return(ss, nil)
	mc.On("GetSentinelDeployment", rf).Once().Return(d, nil)
	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr := &mocks.Client{}
	mr.On("IsMaster", redisPods[0]).Once().Return(true, nil)
	mr.On("IsMaster", redisPods[1]).Once().Return(false, nil)

	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr.On("GetSlaveOf", redisPods[0]).Once().Return("", nil)
	mr.On("GetSlaveOf", redisPods[1]).Once().Return(redisPods[0], nil)

	mc.On("GetSentinelPodsIPs", rf).Once().Return(sentinelPods, nil)

	mr.On("GetNumberSentinelsInMemory", sentinelPods[0]).Once().Return(correctNumber, nil)

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, mr, clock.Base(), log.Nil)

	err := checker.Check(rf)
	assert.NoError(err)
}

func TestCheckSentinelNumberInMemoryResetSentinelResetError(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	ss := &v1beta1.StatefulSet{
		Spec: v1beta1.StatefulSetSpec{
			Replicas: &correctNumber,
		},
	}
	d := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Replicas: &correctNumber,
		},
	}
	redisPods := []string{"0.0.0.0", "1.1.1.1"}
	sentinelPods := []string{"9.9.9.9"}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", rf).Once().Return(ss, nil)
	mc.On("GetSentinelDeployment", rf).Once().Return(d, nil)
	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr := &mocks.Client{}
	mr.On("IsMaster", redisPods[0]).Once().Return(true, nil)
	mr.On("IsMaster", redisPods[1]).Once().Return(false, nil)

	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr.On("GetSlaveOf", redisPods[0]).Once().Return("", nil)
	mr.On("GetSlaveOf", redisPods[1]).Once().Return(redisPods[0], nil)

	mc.On("GetSentinelPodsIPs", rf).Once().Return(sentinelPods, nil)

	mr.On("GetNumberSentinelsInMemory", sentinelPods[0]).Once().Return(int32(0), nil)
	mr.On("ResetSentinel", sentinelPods[0]).Once().Return(errors.New(""))

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, mr, clock.Base(), log.Nil)

	err := checker.Check(rf)
	assert.Error(err)
}

func TestCheck(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	ss := &v1beta1.StatefulSet{
		Spec: v1beta1.StatefulSetSpec{
			Replicas: &correctNumber,
		},
	}
	d := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Replicas: &correctNumber,
		},
	}
	redisPods := []string{"0.0.0.0", "1.1.1.1"}
	sentinelPods := []string{"9.9.9.9"}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisStatefulset", rf).Once().Return(ss, nil)
	mc.On("GetSentinelDeployment", rf).Once().Return(d, nil)
	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr := &mocks.Client{}
	mr.On("IsMaster", redisPods[0]).Once().Return(true, nil)
	mr.On("IsMaster", redisPods[1]).Once().Return(false, nil)

	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr.On("GetSlaveOf", redisPods[0]).Once().Return("", nil)
	mr.On("GetSlaveOf", redisPods[1]).Once().Return(redisPods[0], nil)

	mc.On("GetSentinelPodsIPs", rf).Once().Return(sentinelPods, nil)

	mr.On("GetNumberSentinelsInMemory", sentinelPods[0]).Once().Return(int32(0), nil)
	mr.On("ResetSentinel", sentinelPods[0]).Once().Return(nil)

	mck := &mocks.Clock{}
	mck.On("Sleep", mock.Anything).Once().Return()

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, mr, mck, log.Nil)

	err := checker.Check(rf)
	assert.NoError(err)
}

func TestGetMasterGetRedisPodsIPsError(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisPodsIPs", rf).Once().Return(nil, errors.New(""))

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, redis.New(), clock.Base(), log.Nil)

	_, err := checker.GetMaster(rf)
	assert.Error(err)
}

func TestGetMasterIsMasterError(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	redisPods := []string{"0.0.0.0"}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr := &mocks.Client{}
	mr.On("IsMaster", redisPods[0]).Once().Return(false, errors.New(""))

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, mr, clock.Base(), log.Nil)

	_, err := checker.GetMaster(rf)
	assert.Error(err)
}

func TestGetMasteIsMasterNumberError(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	redisPods := []string{"0.0.0.0", "1.1.1.1"}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr := &mocks.Client{}
	mr.On("IsMaster", redisPods[0]).Once().Return(true, nil)
	mr.On("IsMaster", redisPods[1]).Once().Return(true, nil)

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, mr, clock.Base(), log.Nil)

	_, err := checker.GetMaster(rf)
	assert.Error(err)
}

func TestGetMasteIsMasterNumber(t *testing.T) {
	assert := assert.New(t)
	correctNumber := int32(3)
	rf := &failover.RedisFailover{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: failover.RedisFailoverSpec{
			Redis: failover.RedisSettings{
				Replicas: correctNumber,
			},
			Sentinel: failover.SentinelSettings{
				Replicas: correctNumber,
			},
		},
	}
	redisPods := []string{"0.0.0.0", "1.1.1.1"}
	mc := &mocks.RedisFailoverClient{}
	mc.On("GetRedisPodsIPs", rf).Once().Return(redisPods, nil)

	mr := &mocks.Client{}
	mr.On("IsMaster", redisPods[0]).Once().Return(true, nil)
	mr.On("IsMaster", redisPods[1]).Once().Return(false, nil)

	checker := failover.NewRedisFailoverChecker(metrics.Dummy, mc, mr, clock.Base(), log.Nil)

	master, err := checker.GetMaster(rf)
	assert.NoError(err)
	assert.Equal(redisPods[0], master, "")
}
