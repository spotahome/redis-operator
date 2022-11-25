package k8s

import (
	"fmt"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/metrics"
	"k8s.io/apimachinery/pkg/api/errors"
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

func recordMetrics(namespace string, kind string, object string, operation string, err error, metricsRecorder metrics.Recorder) {
	if nil == err {
		metricsRecorder.RecordK8sOperation(namespace, kind, object, operation, metrics.SUCCESS, metrics.NOT_APPLICABLE)
	} else if errors.IsForbidden(err) {
		metricsRecorder.RecordK8sOperation(namespace, kind, object, operation, metrics.FAIL, metrics.K8S_FORBIDDEN_ERR)
	} else if errors.IsUnauthorized(err) {
		metricsRecorder.RecordK8sOperation(namespace, kind, object, operation, metrics.FAIL, metrics.K8S_UNAUTH)
	} else if errors.IsNotFound(err) {
		metricsRecorder.RecordK8sOperation(namespace, kind, object, operation, metrics.FAIL, metrics.K8S_NOT_FOUND)
	} else {
		metricsRecorder.RecordK8sOperation(namespace, kind, object, operation, metrics.FAIL, metrics.K8S_MISC)
	}
}
