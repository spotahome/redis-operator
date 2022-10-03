// adapts `builder` pattern to abstract away auth version that is used, and implementation details in each.
package redisauth

import (
	"strings"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
	authv1 "github.com/spotahome/redis-operator/operator/redisfailover/auth/v1"
	authv2 "github.com/spotahome/redis-operator/operator/redisfailover/auth/v2"
	redisauthv2 "github.com/spotahome/redis-operator/operator/redisfailover/auth/v2"
	"github.com/spotahome/redis-operator/service/k8s"
)

const (
	DefaultUserName = "default"
)

var (
	DefaultUsers = map[string]bool{
		"admin":   true,
		"pinger":  true,
		"default": true,
	}

	DefaultPermissionSpace = []string{
		"~*",
		"&*",
	}
)

type provider interface {
	Version() string
	GetAdminCredentials() (string /* username */, string /* password */, error)
	GetAuthSpecAsRedisCliCommands() (string, error)
	GetAuthSpecAsRedisConf() (string, error)
	GetUserPassword(string /* username */) (string /* password*/, error)
	//Load latest data from secrets and return a copy of users
	InterceptUsers(crUsers /*cr = users from CR */ map[string]redisfailoverv1.UserSpec, namespace string, k8sServices k8s.Services) (map[string]redisfailoverv1.UserSpec, error)
	// given a user spec, returns hashed passwords
	GetHashedPasswords(crUser redisfailoverv1.UserSpec) []string
	// given a user spec, returns list of ACL setting.
	GetACLs(crUser redisfailoverv1.UserSpec) []string
	// labels to attach to statefulset reflecting authmode
	GetAuthModeLabels() map[string]string
}

type providerV1 struct {
	version    string
	rf         *redisfailoverv1.RedisFailover
	k8sService k8s.Services
}

type providerV2 struct {
	version    string
	rf         *redisfailoverv1.RedisFailover
	k8sService k8s.Services
}

func GetAuthProvider(rf *redisfailoverv1.RedisFailover, k8sService k8s.Services) provider {
	if authv2.IsEnabled(*rf) {
		return providerV2{
			version:    "V2",
			rf:         rf,
			k8sService: k8sService,
		}
	}
	return providerV1{
		version:    "V1",
		rf:         rf,
		k8sService: k8sService,
	}
}

// --------------------------------------- AuthV1 Impl --------------------------------------- #
func (p providerV1) GetAdminCredentials() (string /* username */, string /* password */, error) {
	username := DefaultUserName // in authv1, default user is the admin user
	password, err := k8s.GetRedisPassword(p.k8sService, p.rf)
	return username, password, err
}

func (p providerV1) GetAuthSpecAsRedisCliCommands() (string, error) {
	log.WithField("namespace", p.rf.Namespace).WithField("resource", p.rf.Name).Infof("authv1 does not support `GetAuthSpecAsRedisCliCommands`", p.rf.Name, p.rf.Namespace)
	return "", nil
}

func (p providerV1) GetAuthSpecAsRedisConf() (string, error) {
	log.WithField("namespace", p.rf.Namespace).WithField("resource", p.rf.Name).Infof("authv1 does not support `GetAuthSpecAsRedisConf`", p.rf.Name, p.rf.Namespace)
	return authv1.GetAuthSpecAsRedisConf(p.rf, p.k8sService)
}

func (p providerV1) GetUserPassword(username string) (string, error) {
	if username == DefaultUserName {
		_, password, err := p.GetAdminCredentials()
		if nil != err {
			return "", err
		}
		return password, err
	}
	log.WithField("namespace", p.rf.Namespace).WithField("resource", p.rf.Name).Warnf("unable to fetch password of %v user;", username, p.rf.Name, p.rf.Namespace)
	return "", nil
}

func (p providerV1) Version() string {
	return p.version
}

func (p providerV1) InterceptUsers(crUsers /*cr = users from CR */ map[string]redisfailoverv1.UserSpec, namespace string, k8sServices k8s.Services) (map[string]redisfailoverv1.UserSpec, error) {
	// make deep copy of users - *must* not update spec directly - which will lead to perpetual reconciliation cycle
	log.WithField("namespace", namespace).Warnf("cannot intecept users for auth version %v")
	return nil, nil
}

func (p providerV1) GetHashedPasswords(crUser redisfailoverv1.UserSpec) []string {
	passwords := []string{}
	log.Warnf("Get passwords called when auth %v is selected", p.version)
	return passwords
}

func (p providerV1) GetACLs(crUser redisfailoverv1.UserSpec) []string {
	acls := []string{}
	log.Warnf("Get ACLs called when auth %v is selected", p.version)
	return acls
}

func (p providerV1) GetDefaultPermissionSpace() []string {
	return []string{authv2.DefaultPermittedKeys, authv2.DefaultPermittedChannels}
}

func (p providerV1) GetAuthModeLabels() map[string]string {
	return map[string]string{
		"auth-mode": p.Version(),
	}
}

// --------------------------------------- AuthV2 Impl --------------------------------------- #
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
	userCreationConfig, err := redisauthv2.GetAuthSpecAsRedisCliCommands(users)
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

//func (p providerV2) GetUserSpecAsRedisCliCommand() ([]string /* passwords */, []string /* permissions */, error)

func (p providerV2) GetUserPassword(username string) (string, error) {
	users, err := redisauthv2.InterceptUsers(p.rf.Spec.AuthV2.Users, p.rf.Namespace, p.k8sService)
	if err != nil {
		return "", err
	}
	password, err := authv2.GetUserPassword(username, users)
	if nil != err {
		return "", err
	}
	return password, nil
}

func (p providerV2) Version() string {
	return p.version
}

func (p providerV2) InterceptUsers(crUsers /*cr = users from CR */ map[string]redisfailoverv1.UserSpec, namespace string, k8sServices k8s.Services) (map[string]redisfailoverv1.UserSpec, error) {
	// make deep copy of users - *must* not update spec directly - which will lead to perpetual reconciliation cycle
	return authv2.InterceptUsers(crUsers, namespace, k8sServices)
}

func (p providerV2) GetHashedPasswords(crUser redisfailoverv1.UserSpec) []string {
	passwords := []string{}
	for _, passwordSpec := range crUser.Passwords {
		passwords = append(passwords, "#"+authv2.GetHashedPassword(passwordSpec.Value))
	}
	return passwords
}

func (p providerV2) GetACLs(crUser redisfailoverv1.UserSpec) []string {
	acls := []string{}
	for _, acl := range strings.Split(crUser.ACL.Value, " ") {
		acls = append(acls, acl)
	}
	return acls
}

func (p providerV2) GetDefaultPermissionSpace() []string {
	return []string{authv2.DefaultPermittedKeys, authv2.DefaultPermittedChannels}
}

func (p providerV2) GetAuthModeLabels() map[string]string {
	return map[string]string{
		"auth-mode": p.Version(),
	}
}
