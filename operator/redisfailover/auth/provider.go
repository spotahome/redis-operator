// adapts `builder` pattern to abstract away auth version that is used, and implementation details in each.
package redisauth

import (
	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
	authv2 "github.com/spotahome/redis-operator/operator/redisfailover/auth/v2"
	redisauthv2 "github.com/spotahome/redis-operator/operator/redisfailover/auth/v2"
	"github.com/spotahome/redis-operator/service/k8s"
)

type provider interface {
	GetAdminCredentials() (string /* username */, string /* password */, error)
	GetAuthSpecAsRedisCliCommands() (string, error)
	GetAuthSpecAsRedisConf() (string, error)
}

type providerV1 struct {
	rf         *redisfailoverv1.RedisFailover
	k8sService k8s.Services
}

type providerV2 struct {
	rf         *redisfailoverv1.RedisFailover
	k8sService k8s.Services
}

func GetAuthProvider(rf *redisfailoverv1.RedisFailover, k8sService k8s.Services) provider {
	if authv2.IsEnabled(*rf) {
		return providerV2{
			rf:         rf,
			k8sService: k8sService,
		}
	}
	return providerV1{
		rf:         rf,
		k8sService: k8sService,
	}
}

// AuthV1 impl
func (p providerV1) GetAdminCredentials() (string /* username */, string /* password */, error) {
	username := "default"
	password, err := k8s.GetRedisPassword(p.k8sService, p.rf)
	return username, password, err
}

func (p providerV1) GetAuthSpecAsRedisCliCommands() (string, error) {
	log.Warnf("authv1 does not support `GetAuthSpecAsRedisCliCommands`; resource %v in %v namespace has authv1 selected.", p.rf.Name, p.rf.Namespace)
	return "", nil
}

func (p providerV1) GetAuthSpecAsRedisConf() (string, error) {
	log.Warnf("authv1 does not support `GetAuthSpecAsRedisConf`; resource %v in %v namespace has authv1 selected.", p.rf.Name, p.rf.Namespace)
	return "", nil
}

// AuthV2 Impl
func (p providerV2) GetAdminCredentials() (string /* username */, string /* password */, error) {
	users, err := redisauthv2.InterceptUsers(p.rf.Spec.AuthV2.Users, p.rf.Namespace, p.k8sService)
	if err != nil {
		return "", "", err
	}
	password, err := redisauthv2.GetUserPassword(redisauthv2.AdminUserName, users)
	return redisauthv2.AdminUserName, password, nil

}
func (p providerV2) GetAuthSpecAsRedisCliCommands() (string, error) {

	users, err := redisauthv2.InterceptUsers(p.rf.Spec.AuthV2.Users, p.rf.Namespace, p.k8sService)
	if err != nil {
		return "", err
	}
	userCreationConfig, err := redisauthv2.GetAuthSpecAsRedisConf(users)
	return userCreationConfig, err

}

func (p providerV2) GetAuthSpecAsRedisConf() (string, error) {

	users, err := redisauthv2.InterceptUsers(p.rf.Spec.AuthV2.Users, p.rf.Namespace, p.k8sService)
	if err != nil {
		return "", err
	}
	userCreationConfig, err := redisauthv2.GetAuthSpecAsRedisConf(users)
	return userCreationConfig, err
}
