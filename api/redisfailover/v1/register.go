package v1

import (
	"github.com/spotahome/redis-operator/api/redisfailover"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	version = "v1"
)

// Team constants
const (
	RFKind       = "RedisFailover"
	RFName       = "redisfailover"
	RFNamePlural = "redisfailovers"
	RFScope      = apiextensionsv1.NamespaceScoped
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: redisfailover.GroupName, Version: version}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return VersionKind(kind).GroupKind()
}

// VersionKind takes an unqualified kind and returns back a Group qualified GroupVersionKind
func VersionKind(kind string) schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind(kind)
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&RedisFailover{},
		&RedisFailoverList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
