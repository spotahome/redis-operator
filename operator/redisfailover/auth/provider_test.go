package redisauth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	mK8SService "github.com/spotahome/redis-operator/mocks/service/k8s"
	redisauth "github.com/spotahome/redis-operator/operator/redisfailover/auth"
)

var (
	name      = "redisfailover-ut"
	namespace = "redisfailover-ut"
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

func generateRFWithAuthV2(enabled bool) *redisfailoverv1.RedisFailover {
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
			AuthV2: redisfailoverv1.AuthV2Settings{
				Enabled: enabled,
				Users: map[string]redisfailoverv1.UserSpec{
					"testUser": {
						Passwords: []redisfailoverv1.Password{
							{
								Value: "testPassword",
							},
						},
						ACL: redisfailoverv1.ACL{
							Value: "+@all",
						},
					},
				}, // ToDo
			},
		},
	}
}

// AuthV2 must be selected only when AuthV2.Enabled := true
func TestAuthV2VersionSelection(t *testing.T) {
	tests := []struct {
		authV2Enabled bool
	}{
		{
			authV2Enabled: true,
		},
		{
			authV2Enabled: false,
		},
	}
	for _, test := range tests {
		assert := assert.New(t)
		rf := generateRFWithAuthV2(test.authV2Enabled)

		authProvider := redisauth.GetAuthProvider(rf, &mK8SService.Services{})

		if test.authV2Enabled {
			assert.Equal(authProvider.Version(), "V2")
		} else {
			assert.Equal(authProvider.Version(), "V1")
		}
	}
}

// When authV2 is selected, admin user must be selected for GetAdminCredentials
// When authV2 is not selected, default user must be selected for GetAdminCredentials
func TestAuthV2GetAdminCredentials(t *testing.T) {
	tests := []struct {
		authV2Enabled bool
	}{
		{
			authV2Enabled: true,
		},
		{
			authV2Enabled: false,
		},
	}
	for _, test := range tests {
		assert := assert.New(t)
		rf := generateRFWithAuthV2(test.authV2Enabled)

		authProvider := redisauth.GetAuthProvider(rf, &mK8SService.Services{})
		username, _, _ := authProvider.GetAdminCredentials()
		if test.authV2Enabled {

			assert.Equal(username, "admin")
		} else {
			assert.Equal(username, "default")
		}
	}
}

// authV1 must be selected when authV2 spec is not present.
func TestAuthV2WithLegacySpec(t *testing.T) {

	assert := assert.New(t)
	rf := generateRF() // Generate RF spec without authV2 spec

	authProvider := redisauth.GetAuthProvider(rf, &mK8SService.Services{})
	assert.Equal(authProvider.Version(), "V1")
	username, _, _ := authProvider.GetAdminCredentials()
	assert.Equal(username, "default")
}

func TestPasswordHashCorrectness(t *testing.T) {
	assert := assert.New(t)
	rf := generateRFWithAuthV2(true)

	authProvider := redisauth.GetAuthProvider(rf, &mK8SService.Services{})
	password := authProvider.GetHashedPasswords(rf.Spec.AuthV2.Users["testUser"])[0]

	// https://codebeautify.org/sha256-hash-generator/y22ecbf06
	assert.Contains(password, "fd5cb51bafd60f6fdbedde6e62c473da6f247db271633e15919bab78a02ee9eb", "Password sha256 hash is is not computed properly.")

}
