package k8s

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/metrics"
)

// RBAC is the service that knows how to interact with k8s to manage RBAC related resources.
type RBAC interface {
	GetClusterRole(name string) (*rbacv1.ClusterRole, error)
	GetRole(namespace, name string) (*rbacv1.Role, error)
	GetRoleBinding(namespace, name string) (*rbacv1.RoleBinding, error)
	CreateRole(namespace string, role *rbacv1.Role) error
	UpdateRole(namespace string, role *rbacv1.Role) error
	CreateOrUpdateRole(namespace string, binding *rbacv1.Role) error
	CreateRoleBinding(namespace string, binding *rbacv1.RoleBinding) error
	UpdateRoleBinding(namespace string, binding *rbacv1.RoleBinding) error
	CreateOrUpdateRoleBinding(namespace string, binding *rbacv1.RoleBinding) error
}

// NamespaceService is the Namespace service implementation using API calls to kubernetes.
type RBACService struct {
	kubeClient      kubernetes.Interface
	logger          log.Logger
	metricsRecorder metrics.Recorder
}

// NewRBACService returns a new RBAC KubeService.
func NewRBACService(kubeClient kubernetes.Interface, logger log.Logger, metricsRecorder metrics.Recorder) *RBACService {
	logger = logger.With("service", "k8s.rbac")
	return &RBACService{
		kubeClient:      kubeClient,
		logger:          logger,
		metricsRecorder: metricsRecorder,
	}
}

func (r *RBACService) GetClusterRole(name string) (*rbacv1.ClusterRole, error) {
	clusterRole, err := r.kubeClient.RbacV1().ClusterRoles().Get(context.TODO(), name, metav1.GetOptions{})
	recordMetrics(metrics.NOT_APPLICABLE, "ClusterRole", name, "GET", err, r.metricsRecorder)
	return clusterRole, err
}

func (r *RBACService) GetRole(namespace, name string) (*rbacv1.Role, error) {
	role, err := r.kubeClient.RbacV1().Roles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	recordMetrics(namespace, "Role", name, "GET", err, r.metricsRecorder)
	return role, err
}

func (r *RBACService) GetRoleBinding(namespace, name string) (*rbacv1.RoleBinding, error) {
	rolbinding, err := r.kubeClient.RbacV1().RoleBindings(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	recordMetrics(namespace, "RoleBinding", name, "GET", err, r.metricsRecorder)
	return rolbinding, err
}

func (r *RBACService) DeleteRole(namespace, name string) error {
	err := r.kubeClient.RbacV1().Roles(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	recordMetrics(namespace, "Role", name, "DELETE", err, r.metricsRecorder)
	if err != nil {
		return err
	}
	r.logger.WithField("namespace", namespace).WithField("role", name).Debugf("role deleted")
	return nil
}

func (r *RBACService) CreateRole(namespace string, role *rbacv1.Role) error {
	_, err := r.kubeClient.RbacV1().Roles(namespace).Create(context.TODO(), role, metav1.CreateOptions{})
	recordMetrics(namespace, "Role", role.GetName(), "CREATE", err, r.metricsRecorder)
	if err != nil {
		return err
	}
	r.logger.WithField("namespace", namespace).WithField("role", role.Name).Debugf("role created")
	return nil
}

func (s *RBACService) UpdateRole(namespace string, role *rbacv1.Role) error {
	_, err := s.kubeClient.RbacV1().Roles(namespace).Update(context.TODO(), role, metav1.UpdateOptions{})
	recordMetrics(namespace, "Role", role.GetName(), "UPDATE", err, s.metricsRecorder)
	if err != nil {
		return err
	}
	s.logger.WithField("namespace", namespace).WithField("role", role.ObjectMeta.Name).Debugf("role updated")
	return err
}

func (r *RBACService) CreateOrUpdateRole(namespace string, role *rbacv1.Role) error {
	storedRole, err := r.GetRole(namespace, role.Name)
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return r.CreateRole(namespace, role)
		}
		return err
	}

	// Already exists, need to Update.
	// Set the correct resource version to ensure we are on the latest version. This way the only valid
	// namespace is our spec(https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#concurrency-control-and-consistency),
	// we will replace the current namespace state.
	role.ResourceVersion = storedRole.ResourceVersion
	return r.UpdateRole(namespace, role)
}

func (r *RBACService) DeleteRoleBinding(namespace, name string) error {
	err := r.kubeClient.RbacV1().RoleBindings(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	recordMetrics(namespace, "RoleBinding", name, "DELETE", err, r.metricsRecorder)
	if err != nil {
		return err
	}
	r.logger.WithField("namespace", namespace).WithField("binding", name).Debugf("role binding deleted")
	return nil
}

func (r *RBACService) CreateRoleBinding(namespace string, binding *rbacv1.RoleBinding) error {
	_, err := r.kubeClient.RbacV1().RoleBindings(namespace).Create(context.TODO(), binding, metav1.CreateOptions{})
	recordMetrics(namespace, "RoleBinding", binding.GetName(), "CREATE", err, r.metricsRecorder)
	if err != nil {
		return err
	}
	r.logger.WithField("namespace", namespace).WithField("binding", binding.Name).Debugf("role binding created")
	return nil
}

func (r *RBACService) UpdateRoleBinding(namespace string, binding *rbacv1.RoleBinding) error {
	_, err := r.kubeClient.RbacV1().RoleBindings(namespace).Update(context.TODO(), binding, metav1.UpdateOptions{})
	recordMetrics(namespace, "Role", binding.GetName(), "UPDATE", err, r.metricsRecorder)
	if err != nil {
		return err
	}
	r.logger.WithField("namespace", namespace).WithField("binding", binding.Name).Debugf("role binding updated")
	return nil
}

func (r *RBACService) CreateOrUpdateRoleBinding(namespace string, binding *rbacv1.RoleBinding) error {
	storedBinding, err := r.GetRoleBinding(namespace, binding.Name)
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return r.CreateRoleBinding(namespace, binding)
		}
		return err
	}

	// Check if the role ref has changed, roleref updates are not allowed, if changed then delete and create again the role binding.
	// https://github.com/kubernetes/kubernetes/blob/0f0a5223dfc75337d03c9b80ae552ae8ef138eeb/pkg/apis/rbac/validation/validation.go#L157-L159
	if storedBinding.RoleRef != binding.RoleRef {
		r.logger.WithField("namespace", namespace).WithField("binding", binding.Name).Infof("roleref changed, need to recreate role binding resource")
		if err := r.DeleteRoleBinding(namespace, binding.Name); err != nil {
			return err
		}
		return r.CreateRoleBinding(namespace, binding)
	}

	// Already exists, need to Update.
	// Set the correct resource version to ensure we are on the latest version. This way the only valid
	// namespace is our spec(https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#concurrency-control-and-consistency),
	// we will replace the current namespace state.
	binding.ResourceVersion = storedBinding.ResourceVersion
	return r.UpdateRoleBinding(namespace, binding)
}
