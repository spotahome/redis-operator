package k8s

import (
	"fmt"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
)

// GetRedisPassword retreives password from kubernetes secret or, if
// unspecified, returns a blank string
func GetRedisPassword(s Services, rf *redisfailoverv1.RedisFailover) (string, error) {

	if rf.Spec.Auth.SecretPath == "" {
		// no auth settings specified, return blank password
		return "", nil
	}

	secret, err := s.GetSecret(rf.ObjectMeta.Namespace, rf.Spec.Auth.SecretPath)
	if err != nil {
		return "", err
	}

	if password, ok := secret.Data["password"]; ok {
		return string(password), nil
	}

	return "", fmt.Errorf("secret \"%s\" does not have a password field", rf.Spec.Auth.SecretPath)
}
