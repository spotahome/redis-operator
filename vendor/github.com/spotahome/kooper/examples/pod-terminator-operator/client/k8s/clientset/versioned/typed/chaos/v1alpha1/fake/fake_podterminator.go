package fake

import (
	v1alpha1 "github.com/spotahome/kooper/examples/pod-terminator-operator/apis/chaos/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePodTerminators implements PodTerminatorInterface
type FakePodTerminators struct {
	Fake *FakeChaosV1alpha1
}

var podterminatorsResource = schema.GroupVersionResource{Group: "chaos.spotahome.com", Version: "v1alpha1", Resource: "podterminators"}

var podterminatorsKind = schema.GroupVersionKind{Group: "chaos.spotahome.com", Version: "v1alpha1", Kind: "PodTerminator"}

// Get takes name of the podTerminator, and returns the corresponding podTerminator object, and an error if there is any.
func (c *FakePodTerminators) Get(name string, options v1.GetOptions) (result *v1alpha1.PodTerminator, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(podterminatorsResource, name), &v1alpha1.PodTerminator{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodTerminator), err
}

// List takes label and field selectors, and returns the list of PodTerminators that match those selectors.
func (c *FakePodTerminators) List(opts v1.ListOptions) (result *v1alpha1.PodTerminatorList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(podterminatorsResource, podterminatorsKind, opts), &v1alpha1.PodTerminatorList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PodTerminatorList{}
	for _, item := range obj.(*v1alpha1.PodTerminatorList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested podTerminators.
func (c *FakePodTerminators) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(podterminatorsResource, opts))
}

// Create takes the representation of a podTerminator and creates it.  Returns the server's representation of the podTerminator, and an error, if there is any.
func (c *FakePodTerminators) Create(podTerminator *v1alpha1.PodTerminator) (result *v1alpha1.PodTerminator, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(podterminatorsResource, podTerminator), &v1alpha1.PodTerminator{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodTerminator), err
}

// Update takes the representation of a podTerminator and updates it. Returns the server's representation of the podTerminator, and an error, if there is any.
func (c *FakePodTerminators) Update(podTerminator *v1alpha1.PodTerminator) (result *v1alpha1.PodTerminator, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(podterminatorsResource, podTerminator), &v1alpha1.PodTerminator{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodTerminator), err
}

// Delete takes name of the podTerminator and deletes it. Returns an error if one occurs.
func (c *FakePodTerminators) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(podterminatorsResource, name), &v1alpha1.PodTerminator{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePodTerminators) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(podterminatorsResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.PodTerminatorList{})
	return err
}

// Patch applies the patch and returns the patched podTerminator.
func (c *FakePodTerminators) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.PodTerminator, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(podterminatorsResource, name, data, subresources...), &v1alpha1.PodTerminator{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodTerminator), err
}
