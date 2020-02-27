package redisfailover_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
	mRFService "github.com/spotahome/redis-operator/mocks/operator/redisfailover/service"
	mK8SService "github.com/spotahome/redis-operator/mocks/service/k8s"
	rfOperator "github.com/spotahome/redis-operator/operator/redisfailover"
)

const (
	name      = "test"
	namespace = "testns"

	bootstrapName = "rfb-test"
	sentinelName  = "rfs-test"
	redisName     = "rfr-test"
)

func generateConfig() rfOperator.Config {
	return rfOperator.Config{
		ListenAddress: "1234",
		MetricsPath:   "/awesome",
	}
}

func generateRF(enableExporter bool, bootstrapping bool) *redisfailoverv1.RedisFailover {
	return &redisfailoverv1.RedisFailover{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: redisfailoverv1.RedisFailoverSpec{
			Redis: redisfailoverv1.RedisSettings{
				Replicas: int32(3),
				Exporter: redisfailoverv1.RedisExporter{
					Enabled: enableExporter,
				},
			},
			Sentinel: redisfailoverv1.SentinelSettings{
				Replicas: int32(3),
			},
			BootstrapNode: generateRFBootstrappingNode(bootstrapping),
		},
	}
}

func generateRFBootstrappingNode(bootstrapping bool) *redisfailoverv1.BootstrapSettings {
	if bootstrapping {
		return &redisfailoverv1.BootstrapSettings{
			Host: "127.0.0.1",
			Port: "6379",
		}
	}
	return nil
}

func TestEnsure(t *testing.T) {
	tests := []struct {
		name                        string
		exporter                    bool
		bootstrapping               bool
		bootstrappingAllowSentinels bool
	}{
		{
			name:                        "Call everything, use exporter",
			exporter:                    true,
			bootstrapping:               false,
			bootstrappingAllowSentinels: false,
		},
		{
			name:                        "Call everything, don't use exporter",
			exporter:                    false,
			bootstrapping:               false,
			bootstrappingAllowSentinels: false,
		},
		{
			name:                        "Only ensure Redis when bootstrapping",
			exporter:                    false,
			bootstrapping:               true,
			bootstrappingAllowSentinels: false,
		},
		{
			name:                        "call everything when bootstrapping allows sentinels",
			exporter:                    false,
			bootstrapping:               true,
			bootstrappingAllowSentinels: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			rf := generateRF(test.exporter, test.bootstrapping)
			if test.bootstrapping {
				rf.Spec.BootstrapNode.AllowSentinels = test.bootstrappingAllowSentinels
			}

			config := generateConfig()
			mk := &mK8SService.Services{}
			mrfc := &mRFService.RedisFailoverCheck{}
			mrfh := &mRFService.RedisFailoverHeal{}
			mrfs := &mRFService.RedisFailoverClient{}
			if test.exporter {
				mrfs.On("EnsureRedisService", rf, mock.Anything, mock.Anything).Once().Return(nil)
			} else {
				mrfs.On("EnsureNotPresentRedisService", rf).Once().Return(nil)
			}

			if !test.bootstrapping || test.bootstrappingAllowSentinels {
				mrfs.On("EnsureSentinelService", rf, mock.Anything, mock.Anything).Once().Return(nil)
				mrfs.On("EnsureSentinelConfigMap", rf, mock.Anything, mock.Anything).Once().Return(nil)
				mrfs.On("EnsureSentinelDeployment", rf, mock.Anything, mock.Anything).Once().Return(nil)
			}

			mrfs.On("EnsureRedisConfigMap", rf, mock.Anything, mock.Anything).Once().Return(nil)
			mrfs.On("EnsureRedisShutdownConfigMap", rf, mock.Anything, mock.Anything).Once().Return(nil)
			mrfs.On("EnsureRedisReadinessConfigMap", rf, mock.Anything, mock.Anything).Once().Return(nil)
			mrfs.On("EnsureRedisStatefulset", rf, mock.Anything, mock.Anything).Once().Return(nil)

			// Create the Kops client and call the valid logic.
			handler := rfOperator.NewRedisFailoverHandler(config, mrfs, mrfc, mrfh, mk, metrics.Dummy, log.Dummy)
			err := handler.Ensure(rf, map[string]string{}, []metav1.OwnerReference{})

			assert.NoError(err)
			mrfs.AssertExpectations(t)
		})
	}
}
