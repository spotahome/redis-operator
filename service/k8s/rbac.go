package k8s

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/spotahome/redis-operator/log"
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
	kubeClient kubernetes.Interface
	logger     log.Logger
}

// NewRBACService returns a new RBAC KubeService.
func NewRBACService(kubeClient kubernetes.Interface, logger log.Logger) *RBACService {
	logger = logger.With("service", "k8s.rbac")
	return &RBACService{
		kubeClient: kubeClient,
		logger:     logger,
	}
}

func (r *RBACService) GetClusterRole(name string) (*rbacv1.ClusterRole, error) {
	return r.kubeClient.RbacV1().ClusterRoles().Get(name, metav1.GetOptions{})
}

func (r *RBACService) GetRole(namespace, name string) (*rbacv1.Role, error) {
	return r.kubeClient.RbacV1().Roles(namespace).Get(name, metav1.GetOptions{})
}

func (r *RBACService) GetRoleBinding(namespace, name string) (*rbacv1.RoleBinding, error) {
	return r.kubeClient.RbacV1().RoleBindings(namespace).Get(name, metav1.GetOptions{})
}

func (r *RBACService) DeleteRole(namespace, name string) error {
	err := r.kubeClient.RbacV1().Roles(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	r.logger.WithField("namespace", namespace).WithField("role", name).Infof("role deleted")
	return nil
}

func (r *RBACService) CreateRole(namespace string, role *rbacv1.Role) error {
	_, err := r.kubeClient.RbacV1().Roles(namespace).Create(role)
	if err != nil {
		return err
	}
	r.logger.WithField("namespace", namespace).WithField("role", role.Name).Infof("role created")
	return nil
}

func (s *RBACService) UpdateRole(namespace string, role *rbacv1.Role) error {
	_, err := s.kubeClient.RbacV1().Roles(namespace).Update(role)
	if err != nil {
		return err
	}
	s.logger.WithField("namespace", namespace).WithField("role", role.ObjectMeta.Name).Infof("role updated")
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
	err := r.kubeClient.RbacV1().RoleBindings(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	r.logger.WithField("namespace", namespace).WithField("binding", name).Infof("role binding deleted")
	return nil
}

func (r *RBACService) CreateRoleBinding(namespace string, binding *rbacv1.RoleBinding) error {
	_, err := r.kubeClient.RbacV1().RoleBindings(namespace).Create(binding)
	if err != nil {
		return err
	}
	r.logger.WithField("namespace", namespace).WithField("binding", binding.Name).Infof("role binding created")
	return nil
}

func (r *RBACService) UpdateRoleBinding(namespace string, binding *rbacv1.RoleBinding) error {
	_, err := r.kubeClient.RbacV1().RoleBindings(namespace).Update(binding)
	if err != nil {
		return err
	}
	r.logger.WithField("namespace", namespace).WithField("binding", binding.Name).Infof("role binding updated")
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
