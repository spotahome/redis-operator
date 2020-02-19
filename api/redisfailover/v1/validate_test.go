package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name            string
		rfName          string
		rfBootstrapNode *BootstrapSettings
		expectedError   string
	}{
		{
			name:   "populates default values",
			rfName: "test",
		},
		{
			name:          "errors on too long of name",
			rfName:        "some-super-absurdely-unnecessarily-long-name-that-will-most-definitely-fail",
			expectedError: "name length can't be higher than 48",
		},
		{
			name:            "BootstrapNode provided without a host",
			rfName:          "test",
			rfBootstrapNode: &BootstrapSettings{},
			expectedError:   "BootstrapNode must include a host when provided",
		},
		{
			name:            "Populates default bootstrap port when valid",
			rfName:          "test",
			rfBootstrapNode: &BootstrapSettings{Host: "127.0.0.1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			rf := generateRedisFailover(test.rfName, test.rfBootstrapNode)

			err := rf.Validate()

			if test.expectedError == "" {
				assert.NoError(err)

				var expectedBootstrapNode *BootstrapSettings
				if test.rfBootstrapNode != nil {
					expectedBootstrapNode = &BootstrapSettings{
						Host: test.rfBootstrapNode.Host,
						Port: defaultRedisPort,
					}
				}

				expectedRF := &RedisFailover{
					ObjectMeta: metav1.ObjectMeta{
						Name:      test.rfName,
						Namespace: "namespace",
					},
					Spec: RedisFailoverSpec{
						Redis: RedisSettings{
							Image:    defaultImage,
							Replicas: defaultRedisNumber,
							Exporter: RedisExporter{
								Image: defaultExporterImage,
							},
						},
						Sentinel: SentinelSettings{
							Image:        defaultImage,
							Replicas:     defaultSentinelNumber,
							CustomConfig: defaultSentinelCustomConfig,
							Exporter: SentinelExporter{
								Image: defaultSentinelExporterImage,
							},
						},
						BootstrapNode: expectedBootstrapNode,
					},
				}
				assert.Equal(expectedRF, rf)
			} else {
				if assert.Error(err) {
					assert.Contains(test.expectedError, err.Error())
				}
			}
		})
	}
}
