package k8s

import (
	"context"
	"fmt"
	"time"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/metrics"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
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

// TODO: Update *CacheStoreFromKubeClient be  implemented via generics

func PodCacheStoreFromKubeClient(rc *rest.RESTClient) (*cache.Store, error) {
	if rc == nil {
		// this case usually happens during testing where dummy / fake clientsets are used
		return nil, fmt.Errorf("rest client not initialized")
	}
	s := runtime.NewScheme()
	err := corev1.AddToScheme(s)
	if nil != err {
		return nil, err
	}
	watchFunc := func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.Watch = true
		parameterCodec := runtime.NewParameterCodec(s)
		return rc.Get().
			Resource("pods").
			VersionedParams(&opts, parameterCodec).
			Watch(context.Background())
	}
	listFunc := func(opts metav1.ListOptions) (*corev1.PodList, error) {
		result := corev1.PodList{}
		err := rc.Get().Resource("pods").Do(context.Background()).Into(&result)
		return &result, err
	}
	podCacheStore, podCacheController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return listFunc(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return watchFunc(lo)
			},
		},
		&corev1.Pod{},
		0*time.Second,
		cache.ResourceEventHandlerFuncs{},
	)

	go podCacheController.Run(wait.NeverStop)

	return &podCacheStore, nil

}

func ServiceCacheStoreFromKubeClient(rc *rest.RESTClient) (*cache.Store, error) {
	if rc == nil {
		// this case usually happens during testing where dummy / fake clientsets are used
		return nil, fmt.Errorf("rest client not initialized")
	}
	s := runtime.NewScheme()
	err := corev1.AddToScheme(s)
	if nil != err {
		return nil, err
	}
	watchFunc := func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.Watch = true
		parameterCodec := runtime.NewParameterCodec(s)
		return rc.Get().
			Resource("services").
			VersionedParams(&opts, parameterCodec).
			Watch(context.Background())
	}
	listFunc := func(opts metav1.ListOptions) (*corev1.ServiceList, error) {
		result := corev1.ServiceList{}
		err := rc.Get().Resource("services").Do(context.Background()).Into(&result)
		return &result, err
	}
	cacheStore, cacheController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return listFunc(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return watchFunc(lo)
			},
		},
		&corev1.Service{},
		0*time.Second,
		cache.ResourceEventHandlerFuncs{},
	)

	go cacheController.Run(wait.NeverStop)

	return &cacheStore, nil
}

func ConfigMapCacheStoreFromKubeClient(rc *rest.RESTClient) (*cache.Store, error) {
	if rc == nil {
		// this case usually happens during testing where dummy / fake clientsets are used
		return nil, fmt.Errorf("rest client not initialized")
	}
	s := runtime.NewScheme()
	err := corev1.AddToScheme(s)
	if err != nil {
		return nil, err
	}
	watchFunc := func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.Watch = true
		parameterCodec := runtime.NewParameterCodec(s)
		return rc.Get().
			Resource("configmap").
			VersionedParams(&opts, parameterCodec).
			Watch(context.Background())
	}
	listFunc := func(opts metav1.ListOptions) (*corev1.ConfigMapList, error) {
		fmt.Printf("cm lister calling...")
		fmt.Printf("resr client: %v...", rc)
		result := corev1.ConfigMapList{}
		err := rc.Get().Resource("configmap").Do(context.Background()).Into(&result)
		fmt.Printf("cm lister called; error found: %v\n", err)
		return &result, err
	}
	cacheStore, cacheController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return listFunc(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return watchFunc(lo)
			},
		},
		&corev1.ConfigMap{},
		0*time.Second,
		cache.ResourceEventHandlerFuncs{},
	)

	go cacheController.Run(wait.NeverStop)

	return &cacheStore, nil
}

func DeploymentCacheStoreFromKubeClient(rc *rest.RESTClient) (*cache.Store, error) {
	if rc == nil {
		// this case usually happens during testing where dummy / fake clientsets are used
		return nil, fmt.Errorf("rest client not initialized")
	}
	s := runtime.NewScheme()
	err := appsv1.AddToScheme(s)
	if nil != err {
		return nil, err
	}
	watchFunc := func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.Watch = true
		parameterCodec := runtime.NewParameterCodec(s)
		return rc.Get().
			Resource("deployments").
			VersionedParams(&opts, parameterCodec).
			Watch(context.Background())
	}
	listFunc := func(opts metav1.ListOptions) (*appsv1.DeploymentList, error) {
		result := appsv1.DeploymentList{}
		err := rc.Get().Resource("deployments").Do(context.Background()).Into(&result)
		return &result, err
	}
	cacheStore, cacheController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return listFunc(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return watchFunc(lo)
			},
		},
		&appsv1.Deployment{},
		0*time.Second,
		cache.ResourceEventHandlerFuncs{},
	)

	go cacheController.Run(wait.NeverStop)

	return &cacheStore, nil
}

func PodDisruptionBudgetCacheStoreFromKubeClient(rc *rest.RESTClient) (*cache.Store, error) {
	if rc == nil {
		// this case usually happens during testing where dummy / fake clientsets are used
		return nil, fmt.Errorf("rest client not initialized")
	}
	s := runtime.NewScheme()
	err := policyv1.AddToScheme(s)
	if nil != err {
		return nil, err
	}
	watchFunc := func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.Watch = true
		parameterCodec := runtime.NewParameterCodec(s)
		return rc.Get().
			Resource("poddisruptionbudgets").
			VersionedParams(&opts, parameterCodec).
			Watch(context.Background())
	}
	listFunc := func(opts metav1.ListOptions) (*policyv1.PodDisruptionBudgetList, error) {
		result := policyv1.PodDisruptionBudgetList{}
		err := rc.Get().Resource("poddisruptionbudgets").Do(context.Background()).Into(&result)
		return &result, err
	}
	cacheStore, cacheController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return listFunc(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return watchFunc(lo)
			},
		},
		&policyv1.PodDisruptionBudget{},
		0*time.Second,
		cache.ResourceEventHandlerFuncs{},
	)

	go cacheController.Run(wait.NeverStop)

	return &cacheStore, nil
}

func RoleCacheStoreFromKubeClient(rc *rest.RESTClient) (*cache.Store, error) {
	if rc == nil {
		// this case usually happens during testing where dummy / fake clientsets are used
		return nil, fmt.Errorf("rest client not initialized")
	}
	s := runtime.NewScheme()
	err := rbacv1.AddToScheme(s)
	if nil != err {
		return nil, err
	}
	watchFunc := func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.Watch = true
		parameterCodec := runtime.NewParameterCodec(s)
		return rc.Get().
			Resource("roles").
			VersionedParams(&opts, parameterCodec).
			Watch(context.Background())
	}
	listFunc := func(opts metav1.ListOptions) (*rbacv1.RoleList, error) {
		result := rbacv1.RoleList{}
		err := rc.Get().Resource("roles").Do(context.Background()).Into(&result)
		return &result, err
	}
	cacheStore, cacheController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return listFunc(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return watchFunc(lo)
			},
		},
		&rbacv1.Role{},
		0*time.Second,
		cache.ResourceEventHandlerFuncs{},
	)

	go cacheController.Run(wait.NeverStop)

	return &cacheStore, nil
}

func ClusterRoleCacheStoreFromKubeClient(rc *rest.RESTClient) (*cache.Store, error) {
	if rc == nil {
		// this case usually happens during testing where dummy / fake clientsets are used
		return nil, fmt.Errorf("rest client not initialized")
	}
	s := runtime.NewScheme()
	err := rbacv1.AddToScheme(s)
	if nil != err {
		return nil, err
	}
	watchFunc := func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.Watch = true
		parameterCodec := runtime.NewParameterCodec(s)
		return rc.Get().
			Resource("clusterroles").
			VersionedParams(&opts, parameterCodec).
			Watch(context.Background())
	}
	listFunc := func(opts metav1.ListOptions) (*rbacv1.ClusterRoleList, error) {
		result := rbacv1.ClusterRoleList{}
		err := rc.Get().Resource("clusterroles").Do(context.Background()).Into(&result)
		return &result, err
	}
	cacheStore, cacheController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return listFunc(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return watchFunc(lo)
			},
		},
		&rbacv1.ClusterRole{},
		0*time.Second,
		cache.ResourceEventHandlerFuncs{},
	)

	go cacheController.Run(wait.NeverStop)

	return &cacheStore, err
}

func RoleBindingCacheStoreFromKubeClient(rc *rest.RESTClient) (*cache.Store, error) {
	if rc == nil {
		// this case usually happens during testing where dummy / fake clientsets are used
		return nil, fmt.Errorf("rest client not initialized")
	}
	s := runtime.NewScheme()
	err := rbacv1.AddToScheme(s)
	if nil != err {
		return nil, err
	}
	watchFunc := func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.Watch = true
		parameterCodec := runtime.NewParameterCodec(s)
		return rc.Get().
			Resource("rolebindings").
			VersionedParams(&opts, parameterCodec).
			Watch(context.Background())
	}
	listFunc := func(opts metav1.ListOptions) (*rbacv1.RoleBindingList, error) {
		result := rbacv1.RoleBindingList{}
		err := rc.Get().Resource("rolebindings").Do(context.Background()).Into(&result)
		return &result, err
	}
	cacheStore, cacheController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return listFunc(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return watchFunc(lo)
			},
		},
		&rbacv1.RoleBinding{},
		0*time.Second,
		cache.ResourceEventHandlerFuncs{},
	)

	go cacheController.Run(wait.NeverStop)

	return &cacheStore, err
}
func SecretCacheStoreFromKubeClient(rc *rest.RESTClient) (*cache.Store, error) {
	if rc == nil {
		// this case usually happens during testing where dummy / fake clientsets are used
		return nil, fmt.Errorf("rest client not initialized")
	}
	s := runtime.NewScheme()
	err := corev1.AddToScheme(s)
	if nil != err {
		return nil, err
	}
	watchFunc := func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.Watch = true
		parameterCodec := runtime.NewParameterCodec(s)
		return rc.Get().
			Resource("secrets").
			VersionedParams(&opts, parameterCodec).
			Watch(context.Background())
	}
	listFunc := func(opts metav1.ListOptions) (*corev1.SecretList, error) {
		result := corev1.SecretList{}
		err := rc.Get().Resource("secrets").Do(context.Background()).Into(&result)
		return &result, err
	}
	cacheStore, cacheController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return listFunc(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return watchFunc(lo)
			},
		},
		&corev1.Secret{},
		0*time.Second,
		cache.ResourceEventHandlerFuncs{},
	)

	go cacheController.Run(wait.NeverStop)

	return &cacheStore, err
}

func StatefulSetCacheStoreFromKubeClient(rc *rest.RESTClient) (*cache.Store, error) {
	if rc == nil {
		// this case usually happens during testing where dummy / fake clientsets are used
		return nil, fmt.Errorf("rest client not initialized")
	}
	s := runtime.NewScheme()
	err := appsv1.AddToScheme(s)
	if nil != err {
		return nil, err
	}
	watchFunc := func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.Watch = true
		parameterCodec := runtime.NewParameterCodec(s)
		return rc.Get().
			Resource("statefulsets").
			VersionedParams(&opts, parameterCodec).
			Watch(context.Background())
	}
	listFunc := func(opts metav1.ListOptions) (*appsv1.StatefulSetList, error) {
		result := appsv1.StatefulSetList{}
		err := rc.Get().Resource("statefulsets").Do(context.Background()).Into(&result)
		return &result, err
	}
	cacheStore, cacheController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return listFunc(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return watchFunc(lo)
			},
		},
		&appsv1.StatefulSet{},
		0*time.Second,
		cache.ResourceEventHandlerFuncs{},
	)

	go cacheController.Run(wait.NeverStop)

	return &cacheStore, err
}
