package crd

import (
	"fmt"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

// CRD is a generic implementation agnostic crd
type CRD struct {
	clientset  apiextensionsclient.Interface
	config     rest.Config
	signalChan chan int
	kind       string
	domain     string
	version    string
	apiName    string

	object             runtime.Object
	objectList         runtime.Object
	schemeBuilder      runtime.SchemeBuilder
	schemaGroupVersion schema.GroupVersion
	watcher            Watcher
}

// NewCRD returns a new CRD
func NewCRD(config rest.Config, signalChan chan int, kind string, domain string, version string,
	apiName string, object, objectList runtime.Object, eventHandler EventHandler) (*CRD, error) {
	clientset, err := apiextensionsclient.NewForConfig(&config)
	if err != nil {
		return nil, err
	}
	sgv := schema.GroupVersion{
		Group:   domain,
		Version: version,
	}
	sb := runtime.NewSchemeBuilder(func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(sgv, object, objectList)
		metav1.AddToGroupVersion(scheme, sgv)
		return nil
	})
	crd := &CRD{
		clientset:  clientset,
		config:     config,
		signalChan: signalChan,
		kind:       kind,
		domain:     domain,
		version:    version,
		apiName:    apiName,

		object:             object,
		objectList:         objectList,
		schemeBuilder:      sb,
		schemaGroupVersion: sgv,
	}
	client, err := crd.createClient()
	if err != nil {
		return nil, err
	}
	watcher := NewWatcher(client, signalChan, apiName, object, eventHandler, clientset)
	crd.watcher = *watcher
	return crd, nil
}

// Create creates the crd on k8s
func (t *CRD) Create() error {
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s.%s", t.apiName, t.domain),
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   t.domain,
			Version: t.version,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural: t.apiName,
				Kind:   t.kind,
			},
		},
	}

	_, err := t.clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	return err
}

func (t *CRD) createClient() (*rest.RESTClient, error) {
	scheme := runtime.NewScheme()
	if err := t.schemeBuilder.AddToScheme(scheme); err != nil {
		return nil, err
	}

	t.config.GroupVersion = &t.schemaGroupVersion
	t.config.APIPath = "/apis"
	t.config.ContentType = runtime.ContentTypeJSON
	t.config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	return rest.RESTClientFor(&t.config)
}

// Watch will listen to crd k8s events
func (t *CRD) Watch() {
	t.watcher.Watch()
}
