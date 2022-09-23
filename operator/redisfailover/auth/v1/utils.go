package authv1

import (
	"fmt"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/service/k8s"
)

func GetAuthSpecAsRedisConf(rf *redisfailoverv1.RedisFailover, k8sService k8s.Services) (string, error) {
	defaultUserPassword, err := k8s.GetRedisPassword(k8sService, rf)
	if nil != err {
		return "", err
	}
	defaultUSerSpec := ""
	if defaultUserPassword != "" {
		defaultUSerSpec = fmt.Sprintf("requirepass  %v\n", defaultUserPassword)
	}
	pingerUserSpec := fmt.Sprintf("user %v on >%v %v %v %v\n", PingerUserName, DefaultPingerUserPassword, PingerUserPermissions, DefaultPermittedKeys, DefaultPermittedChannels)
	return defaultUSerSpec + pingerUserSpec, nil
}
