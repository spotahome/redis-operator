package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generateRedisFailoverMasterName(name string, disableMyMaster bool) *RedisFailover {
	return &RedisFailover{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "namespace",
		},
		Spec: RedisFailoverSpec{
			Sentinel: SentinelSettings{
				DisableMyMaster: disableMyMaster,
			},
		},
	}
}

func TestMyMaster(t *testing.T) {
	tests := []struct {
		name            string
		expectation     string
		disableMyMaster bool
	}{
		{
			name:        "use default mymaster",
			expectation: "mymaster",
		},
		{
			name:            "passing-false",
			expectation:     "mymaster",
			disableMyMaster: false,
		},
		{
			name:            "passing-true",
			expectation:     "passing-true",
			disableMyMaster: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rf := generateRedisFailoverMasterName(test.name, test.disableMyMaster)
			assert.Equal(t, test.expectation, rf.MasterName())
		})
	}
}
