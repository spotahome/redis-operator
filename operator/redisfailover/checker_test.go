package redisfailover_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
	mRFService "github.com/spotahome/redis-operator/mocks/operator/redisfailover/service"
	mK8SService "github.com/spotahome/redis-operator/mocks/service/k8s"
	rfOperator "github.com/spotahome/redis-operator/operator/redisfailover"
)

func TestCheckAndHeal(t *testing.T) {
	tests := []struct {
		name                           string
		nMasters                       int
		nRedis                         int
		forceNewMaster                 bool
		slavesOK                       bool
		sentinelMonitorOK              bool
		sentinelNumberInMemoryOK       bool
		sentinelSlavesNumberInMemoryOK bool
		redisCheckNumberOK             bool
		redisSetMasterOnAllOK          bool
		bootstrapping                  bool
		allowSentinels                 bool
	}{
		{
			name:                           "Everything ok, no need to heal",
			nMasters:                       1,
			nRedis:                         3,
			forceNewMaster:                 false,
			slavesOK:                       true,
			sentinelMonitorOK:              true,
			sentinelNumberInMemoryOK:       true,
			sentinelSlavesNumberInMemoryOK: true,
			redisCheckNumberOK:             true,
			redisSetMasterOnAllOK:          true,
			bootstrapping:                  false,
			allowSentinels:                 false,
		},
		{
			name:                           "Multiple masters",
			nMasters:                       2,
			nRedis:                         3,
			forceNewMaster:                 false,
			slavesOK:                       true,
			sentinelMonitorOK:              true,
			sentinelNumberInMemoryOK:       true,
			sentinelSlavesNumberInMemoryOK: true,
			redisCheckNumberOK:             true,
			redisSetMasterOnAllOK:          true,
			bootstrapping:                  false,
			allowSentinels:                 false,
		},
		{
			name:                           "No masters but wait",
			nMasters:                       0,
			nRedis:                         3,
			forceNewMaster:                 false,
			slavesOK:                       true,
			sentinelMonitorOK:              true,
			sentinelNumberInMemoryOK:       true,
			sentinelSlavesNumberInMemoryOK: true,
			redisCheckNumberOK:             true,
			redisSetMasterOnAllOK:          true,
			bootstrapping:                  false,
			allowSentinels:                 false,
		},
		{
			name:                           "No masters, only one redis available, make master",
			nMasters:                       0,
			nRedis:                         1,
			forceNewMaster:                 false,
			slavesOK:                       true,
			sentinelMonitorOK:              true,
			sentinelNumberInMemoryOK:       true,
			sentinelSlavesNumberInMemoryOK: true,
			redisCheckNumberOK:             true,
			redisSetMasterOnAllOK:          true,
			bootstrapping:                  false,
			allowSentinels:                 false,
		},
		{
			name:                           "No masters, set random",
			nMasters:                       0,
			nRedis:                         3,
			forceNewMaster:                 true,
			slavesOK:                       true,
			sentinelMonitorOK:              true,
			sentinelNumberInMemoryOK:       true,
			redisCheckNumberOK:             true,
			redisSetMasterOnAllOK:          true,
			sentinelSlavesNumberInMemoryOK: true,
			allowSentinels:                 false,
		},
		{
			name:                           "Slaves from master wrong",
			nMasters:                       1,
			nRedis:                         3,
			forceNewMaster:                 false,
			slavesOK:                       false,
			sentinelMonitorOK:              true,
			sentinelNumberInMemoryOK:       true,
			sentinelSlavesNumberInMemoryOK: true,
			redisCheckNumberOK:             true,
			redisSetMasterOnAllOK:          true,
			bootstrapping:                  false,
			allowSentinels:                 false,
		},
		{
			name:                           "Sentinels not pointing correct monitor",
			nMasters:                       1,
			nRedis:                         3,
			forceNewMaster:                 false,
			slavesOK:                       true,
			sentinelMonitorOK:              false,
			sentinelNumberInMemoryOK:       true,
			sentinelSlavesNumberInMemoryOK: true,
			redisCheckNumberOK:             true,
			redisSetMasterOnAllOK:          true,
			bootstrapping:                  false,
			allowSentinels:                 false,
		},
		{
			name:                           "Sentinels with wrong number of sentinels",
			nMasters:                       1,
			nRedis:                         3,
			forceNewMaster:                 false,
			slavesOK:                       true,
			sentinelMonitorOK:              true,
			sentinelNumberInMemoryOK:       false,
			sentinelSlavesNumberInMemoryOK: true,
			redisCheckNumberOK:             true,
			redisSetMasterOnAllOK:          true,
			bootstrapping:                  false,
			allowSentinels:                 false,
		},
		{
			name:                           "Sentinels with wrong number of slaves",
			nMasters:                       1,
			nRedis:                         3,
			forceNewMaster:                 false,
			slavesOK:                       true,
			sentinelMonitorOK:              true,
			sentinelNumberInMemoryOK:       true,
			sentinelSlavesNumberInMemoryOK: false,
			redisCheckNumberOK:             true,
			redisSetMasterOnAllOK:          true,
			bootstrapping:                  false,
			allowSentinels:                 false,
		},
		{
			name:                  "Bootstrapping Mode",
			nMasters:              1,
			nRedis:                3,
			redisCheckNumberOK:    true,
			redisSetMasterOnAllOK: true,
			bootstrapping:         true,
			allowSentinels:        false,
		},
		{
			name:                  "Bootstrapping Mode with failure to check redis number",
			nMasters:              1,
			nRedis:                3,
			redisCheckNumberOK:    false,
			redisSetMasterOnAllOK: true,
			bootstrapping:         true,
			allowSentinels:        false,
		},
		{
			name:                  "Bootstrapping Mode with failure to set master on all",
			nMasters:              1,
			nRedis:                3,
			redisCheckNumberOK:    true,
			redisSetMasterOnAllOK: false,
			bootstrapping:         true,
			allowSentinels:        false,
		},
		{
			name:                           "Bootstrapping Mode that allows sentinels",
			nMasters:                       1,
			nRedis:                         3,
			redisCheckNumberOK:             true,
			redisSetMasterOnAllOK:          true,
			sentinelMonitorOK:              true,
			sentinelNumberInMemoryOK:       true,
			sentinelSlavesNumberInMemoryOK: true,
			bootstrapping:                  true,
			allowSentinels:                 true,
		},
		{
			name:                           "Bootstrapping Mode that allows sentinels sentinel monitor fails",
			nMasters:                       1,
			nRedis:                         3,
			redisCheckNumberOK:             true,
			redisSetMasterOnAllOK:          true,
			sentinelMonitorOK:              false,
			sentinelNumberInMemoryOK:       true,
			sentinelSlavesNumberInMemoryOK: true,
			bootstrapping:                  true,
			allowSentinels:                 true,
		},
		{
			name:                           "Bootstrapping Mode that allows sentinels sentinel with wrong number of sentinels",
			nMasters:                       1,
			nRedis:                         3,
			redisCheckNumberOK:             true,
			redisSetMasterOnAllOK:          true,
			sentinelMonitorOK:              true,
			sentinelNumberInMemoryOK:       false,
			sentinelSlavesNumberInMemoryOK: true,
			bootstrapping:                  true,
			allowSentinels:                 true,
		},
		{
			name:                           "Bootstrapping Mode that allows sentinels sentinel with wrong number of slaves",
			nMasters:                       1,
			nRedis:                         3,
			redisCheckNumberOK:             true,
			redisSetMasterOnAllOK:          true,
			sentinelMonitorOK:              true,
			sentinelNumberInMemoryOK:       true,
			sentinelSlavesNumberInMemoryOK: false,
			bootstrapping:                  true,
			allowSentinels:                 true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			allowSentinels := true
			bootstrappingTests := test.bootstrapping
			bootstrapMaster := "127.0.0.1"
			bootstrapMasterPort := "6379"

			rf := generateRF(false, bootstrappingTests)
			if bootstrappingTests {
				allowSentinels = test.allowSentinels
				rf.Spec.BootstrapNode.AllowSentinels = allowSentinels
			}

			expErr := false
			continueTests := true

			master := "0.0.0.0"
			sentinel := "1.1.1.1"

			config := generateConfig()
			mk := &mK8SService.Services{}
			mrfs := &mRFService.RedisFailoverClient{}
			mrfc := &mRFService.RedisFailoverCheck{}
			mrfh := &mRFService.RedisFailoverHeal{}

			if test.redisCheckNumberOK {
				mrfc.On("CheckRedisNumber", rf).Once().Return(nil)
			} else {
				continueTests = false
				mrfc.On("CheckRedisNumber", rf).Once().Return(errors.New(""))
			}

			if allowSentinels {
				mrfc.On("CheckSentinelNumber", rf).Once().Return(nil)
			}

			if bootstrappingTests && continueTests {
				// once to get ips for config update, once for the UpdateRedisesPods go right
				mrfc.On("GetRedisesIPs", rf).Twice().Return([]string{"0.0.0.1", "0.0.0.2", "0.0.0.3"}, nil)
				mrfh.On("SetRedisCustomConfig", "0.0.0.1", rf).Once().Return(nil)
				mrfh.On("SetRedisCustomConfig", "0.0.0.2", rf).Once().Return(nil)
				mrfh.On("SetRedisCustomConfig", "0.0.0.3", rf).Once().Return(nil)
				mrfc.On("CheckRedisSlavesReady", "0.0.0.1", rf).Once().Return(true, nil)
				mrfc.On("CheckRedisSlavesReady", "0.0.0.2", rf).Once().Return(true, nil)
				mrfc.On("CheckRedisSlavesReady", "0.0.0.3", rf).Once().Return(true, nil)
				mrfc.On("GetStatefulSetUpdateRevision", rf).Once().Return("1", nil)
				mrfc.On("GetRedisesSlavesPods", rf).Once().Return([]string{}, nil)

				if test.redisSetMasterOnAllOK {
					mrfh.On("SetExternalMasterOnAll", bootstrapMaster, bootstrapMasterPort, rf).Once().Return(nil)
				} else {
					expErr = true
					mrfh.On("SetExternalMasterOnAll", bootstrapMaster, bootstrapMasterPort, rf).Once().Return(errors.New(""))
				}
			} else if continueTests {
				mrfc.On("GetNumberMasters", rf).Once().Return(test.nMasters, nil)
				switch test.nMasters {
				case 0:
					mrfc.On("GetRedisesIPs", rf).Once().Return(make([]string, test.nRedis), nil)
					if test.nRedis == 1 {
						mrfh.On("MakeMaster", mock.Anything, rf).Once().Return(nil)
						break
					}
					if test.forceNewMaster {
						mrfc.On("GetMinimumRedisPodTime", rf).Once().Return(1*time.Hour, nil)
						mrfh.On("SetOldestAsMaster", rf).Once().Return(nil)
					} else {
						mrfc.On("GetMinimumRedisPodTime", rf).Once().Return(1*time.Second, nil)
						continueTests = false
					}
				case 1:
					break
				default:
					// always expect error
					expErr = true
				}
				if !expErr && continueTests {
					mrfc.On("GetMasterIP", rf).Twice().Return(master, nil)
					if test.slavesOK {
						mrfc.On("CheckAllSlavesFromMaster", master, rf).Once().Return(nil)
					} else {
						mrfc.On("CheckAllSlavesFromMaster", master, rf).Once().Return(errors.New(""))
						if test.redisSetMasterOnAllOK {
							mrfh.On("SetMasterOnAll", master, rf).Once().Return(nil)
						} else {
							expErr = true
							mrfh.On("SetMasterOnAll", master, rf).Once().Return(errors.New(""))
						}

					}
					mrfc.On("GetRedisesIPs", rf).Twice().Return([]string{master}, nil)
					mrfc.On("GetStatefulSetUpdateRevision", rf).Once().Return("1", nil)
					mrfc.On("GetRedisesSlavesPods", rf).Once().Return([]string{}, nil)
					mrfc.On("GetRedisesMasterPod", rf).Once().Return(master, nil)
					mrfc.On("GetRedisRevisionHash", master, rf).Once().Return("1", nil)
					mrfh.On("SetRedisCustomConfig", master, rf).Once().Return(nil)
				}
			}

			if allowSentinels && !expErr && continueTests {
				mrfc.On("GetSentinelsIPs", rf).Once().Return([]string{sentinel}, nil)
				if test.sentinelMonitorOK {
					if test.bootstrapping {
						mrfc.On("CheckSentinelMonitor", sentinel, bootstrapMaster, bootstrapMasterPort).Once().Return(nil)
					} else {
						mrfc.On("CheckSentinelMonitor", sentinel, master).Once().Return(nil)
					}
				} else {
					if test.bootstrapping {
						mrfc.On("CheckSentinelMonitor", sentinel, bootstrapMaster, bootstrapMasterPort).Once().Return(errors.New(""))
						mrfh.On("NewSentinelMonitorWithPort", sentinel, bootstrapMaster, bootstrapMasterPort, rf).Once().Return(nil)
					} else {
						mrfc.On("CheckSentinelMonitor", sentinel, master).Once().Return(errors.New(""))
						mrfh.On("NewSentinelMonitor", sentinel, master, rf).Once().Return(nil)
					}
				}
				if test.sentinelNumberInMemoryOK {
					mrfc.On("CheckSentinelNumberInMemory", sentinel, rf).Once().Return(nil)
				} else {
					mrfc.On("CheckSentinelNumberInMemory", sentinel, rf).Once().Return(errors.New(""))
					mrfh.On("RestoreSentinel", sentinel).Once().Return(nil)
				}
				if test.sentinelSlavesNumberInMemoryOK {
					mrfc.On("CheckSentinelSlavesNumberInMemory", sentinel, rf).Once().Return(nil)
				} else {
					mrfc.On("CheckSentinelSlavesNumberInMemory", sentinel, rf).Once().Return(errors.New(""))
					mrfh.On("RestoreSentinel", sentinel).Once().Return(nil)
				}
				mrfh.On("SetSentinelCustomConfig", sentinel, rf).Once().Return(nil)
			}

			handler := rfOperator.NewRedisFailoverHandler(config, mrfs, mrfc, mrfh, mk, metrics.Dummy, log.Dummy)
			err := handler.CheckAndHeal(rf)

			if expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			mrfc.AssertExpectations(t)
			mrfh.AssertExpectations(t)
		})
	}
}

func TestUpdate(t *testing.T) {
	type podStatus struct {
		pod    corev1.Pod
		ready  bool
		master bool
	}
	tests := []struct {
		name          string
		pods          []podStatus
		ssVersion     string
		errExpected   bool
		bootstrapping bool
		noMaster      bool
	}{
		{
			name: "all ok, no change needed",
			pods: []podStatus{
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave1",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.0",
						},
					},
					master: false,
					ready:  true,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave2",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.1",
						},
					},
					master: false,
					ready:  true,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "master",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "1.1.1.1",
						},
					},
					master: true,
					ready:  true,
				},
			},
			ssVersion:     "10",
			errExpected:   false,
			bootstrapping: false,
		},
		{
			name: "syncing",
			pods: []podStatus{
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave1",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.0",
						},
					},
					master: false,
					ready:  true,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave2",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.1",
						},
					},
					master: false,
					ready:  false,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "master",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "1.1.1.1",
						},
					},
					master: true,
					ready:  true,
				},
			},
			ssVersion:     "10",
			errExpected:   false,
			bootstrapping: false,
		},
		{
			name: "pod version incorrect",
			pods: []podStatus{
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave1",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.0",
						},
					},
					master: false,
					ready:  true,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave2",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.1",
						},
					},
					master: false,
					ready:  true,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "master",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "1.1.1.1",
						},
					},
					master: true,
					ready:  true,
				},
			},
			ssVersion:     "1",
			errExpected:   false,
			bootstrapping: false,
		},
		{
			name: "master version incorrect",
			pods: []podStatus{
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave1",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.0",
						},
					},
					master: false,
					ready:  true,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave2",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.1",
						},
					},
					master: false,
					ready:  true,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "master",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "1",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "1.1.1.1",
						},
					},
					master: true,
					ready:  true,
				},
			},
			ssVersion:     "10",
			errExpected:   false,
			bootstrapping: false,
		},
		{
			name: "all ok, no change needed when in bootstrap mode",
			pods: []podStatus{
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave1",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.0",
						},
					},
					master: false,
					ready:  true,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave2",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.1",
						},
					},
					master: false,
					ready:  true,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave3",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "1.1.1.1",
						},
					},
					master: false,
					ready:  true,
				},
			},
			ssVersion:     "10",
			errExpected:   false,
			bootstrapping: true,
		},
		{
			name: "syncing when in bootstrap mode",
			pods: []podStatus{
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave1",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.0",
						},
					},
					master: false,
					ready:  true,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave2",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.1",
						},
					},
					master: false,
					ready:  false,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave3",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "1.1.1.1",
						},
					},
					master: false,
					ready:  true,
				},
			},
			ssVersion:     "10",
			errExpected:   false,
			bootstrapping: true,
		},
		{
			name: "pod version incorrect when in bootstrap mode",
			pods: []podStatus{
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave1",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.0",
						},
					},
					master: false,
					ready:  true,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave2",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.1",
						},
					},
					master: false,
					ready:  true,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave3",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "1.1.1.1",
						},
					},
					master: false,
					ready:  true,
				},
			},
			ssVersion:     "1",
			errExpected:   false,
			bootstrapping: true,
		},
		{
			name: "when no master exists",
			pods: []podStatus{
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave1",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.0",
						},
					},
					master: false,
					ready:  true,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave2",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "0.0.0.1",
						},
					},
					master: false,
					ready:  true,
				},
				{
					pod: corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "slave3",
							Labels: map[string]string{
								appsv1.ControllerRevisionHashLabelKey: "10",
							},
						},
						Status: corev1.PodStatus{
							PodIP: "1.1.1.1",
						},
					},
					master: false,
					ready:  true,
				},
			},
			ssVersion:     "10",
			errExpected:   true,
			bootstrapping: false,
			noMaster:      true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			rf := generateRF(false, test.bootstrapping)

			config := generateConfig()
			mrfs := &mRFService.RedisFailoverClient{}

			mrfc := &mRFService.RedisFailoverCheck{}
			mrfc.On("GetRedisesIPs", rf).Once().Return([]string{"0.0.0.0", "0.0.0.1", "1.1.1.1"}, nil)

			next := true
			if !test.bootstrapping {
				master := "1.1.1.1"
				if test.noMaster {
					master = ""
				}
				mrfc.On("GetMasterIP", rf).Once().Return(master, nil)
			}

			for _, pod := range test.pods {
				if !pod.master {
					mrfc.On("CheckRedisSlavesReady", pod.pod.Status.PodIP, rf).Once().Return(pod.ready, nil)
				}
				if !pod.ready {
					next = false
					break
				}
			}
			mrfh := &mRFService.RedisFailoverHeal{}

			if next {
				replicas := []string{"slave1", "slave2"}
				if test.bootstrapping || test.noMaster {
					replicas = append(replicas, "slave3")
				}
				mrfc.On("GetStatefulSetUpdateRevision", rf).Once().Return(test.ssVersion, nil)
				mrfc.On("GetRedisesSlavesPods", rf).Once().Return(replicas, nil)

				for _, pod := range test.pods {
					mrfc.On("GetRedisRevisionHash", pod.pod.ObjectMeta.Name, rf).Once().Return(pod.pod.ObjectMeta.Labels[appsv1.ControllerRevisionHashLabelKey], nil)
					if pod.pod.ObjectMeta.Labels[appsv1.ControllerRevisionHashLabelKey] != test.ssVersion {
						mrfh.On("DeletePod", pod.pod.ObjectMeta.Name, rf).Once().Return(nil)
						if pod.master == false {
							next = false
							break
						}
					}
				}
				fmt.Println(fmt.Sprintf("%v - %v", test.name, next))
				if next && !test.bootstrapping {
					if test.noMaster {
						mrfc.On("GetRedisesMasterPod", rf).Once().Return("", errors.New(""))
					} else {
						mrfc.On("GetRedisesMasterPod", rf).Once().Return("master", nil)
					}
				}
			}

			mk := &mK8SService.Services{}

			handler := rfOperator.NewRedisFailoverHandler(config, mrfs, mrfc, mrfh, mk, metrics.Dummy, log.Dummy)
			err := handler.UpdateRedisesPods(rf)

			if test.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			mrfc.AssertExpectations(t)
			mrfh.AssertExpectations(t)

		})
	}
}
